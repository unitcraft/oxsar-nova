// Package alliance — создание и управление альянсами (M6).
//
// Потоки вступления:
//   - is_open=true  → Join() сразу добавляет участника.
//   - is_open=false → Join() создаёт заявку; owner вызывает Approve/Reject.
//
// Ограничения:
//   - Один игрок — один альянс (PK alliance_members.user_id).
//   - Только owner может распустить альянс (DELETE).
//   - Kick-механика, ранги, отношения (NAP/WAR/ALLY) — M6+.
package alliance

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/i18n"
	"oxsar/game-nova/internal/moderation"
	"oxsar/game-nova/internal/repo"
)

// AutoMsgSender — узкий интерфейс к automsg.SendDirect.
type AutoMsgSender interface {
	SendDirect(ctx context.Context, tx pgx.Tx, userID string, folder int, title, body string) error
}

type Service struct {
	db        repo.Exec
	automsg   AutoMsgSender
	bundle    *i18n.Bundle
	blacklist *moderation.Blacklist
}

func NewService(db repo.Exec) *Service { return &Service{db: db} }

// WithBlacklist — UGC-модерация для tag/name альянса (план 46).
// Если nil — проверка отключена.
func (s *Service) WithBlacklist(bl *moderation.Blacklist) *Service {
	s.blacklist = bl
	return s
}

// WithAutoMsg подключает сервис системных сообщений (опционально).
func (s *Service) WithAutoMsg(a AutoMsgSender) *Service {
	s.automsg = a
	return s
}

func (s *Service) WithBundle(b *i18n.Bundle) *Service {
	s.bundle = b
	return s
}

func (s *Service) tr(group, key string, vars map[string]string) string {
	if s.bundle == nil {
		return "[" + group + "." + key + "]"
	}
	return s.bundle.Tr(i18n.LangRu, group, key, vars)
}

// notifyAlliance — помощник для рассылки уведомлений в папку
// MSG_FOLDER_ALLIANCE=6 (legacy `config/consts.php:513`, см.
// automsg.FolderAlliance). Ошибки глотаются (не критично).
func (s *Service) notifyAlliance(ctx context.Context, userID, title, body string) {
	if s.automsg == nil {
		return
	}
	_ = s.automsg.SendDirect(ctx, nil, userID, 6, title, body)
}

// BroadcastMail — план 72.1.43 / правило 1:1 для /alliance.
// Legacy `Alliance::globalMail` отправляет message всем участникам
// альянса (с проверкой permission CAN_SEND_GLOBAL_MAIL).
//
// Использует AutoMsg (folder=6 = alliance) для каждого активного
// участника. Не шлёт самому себе и забаненным.
func (s *Service) BroadcastMail(ctx context.Context, requesterID, allianceID, title, body string) error {
	var memberIDs []string
	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		mem, err := LoadMembership(ctx, tx, requesterID, allianceID)
		if err != nil {
			return err
		}
		if mem == nil {
			return ErrNotMember
		}
		can, err := Has(ctx, tx, mem, PermSendGlobalMail)
		if err != nil {
			return err
		}
		if !can {
			return ErrForbidden
		}
		rows, err := tx.Query(ctx, `
			SELECT id FROM users
			WHERE alliance_id = $1 AND id <> $2 AND banned_at IS NULL
		`, allianceID, requesterID)
		if err != nil {
			return fmt.Errorf("broadcast: list members: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var memberID string
			if err := rows.Scan(&memberID); err != nil {
				return err
			}
			memberIDs = append(memberIDs, memberID)
		}
		return rows.Err()
	})
	if err != nil {
		return err
	}

	if s.automsg == nil {
		return nil // тестовое окружение без AutoMsg.
	}
	// Шлём за пределами транзакции — AutoMsg сам начинает свой tx
	// при необходимости (передаём nil).
	for _, mid := range memberIDs {
		_ = s.automsg.SendDirect(ctx, nil, mid, 6, title, body)
	}
	return nil
}

// UpdateTagName — план 72.1.43 / правило 1:1. Legacy
// `Alliance::updateAllyTag` + `updateAllyName`. Только owner может
// менять. Проверка уникальности через ErrTagTaken/ErrNameTaken.
//
// Если tag/name пустой — поле не меняется (PATCH-семантика).
func (s *Service) UpdateTagName(ctx context.Context, ownerID, allianceID, newTag, newName string) error {
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var owner string
		if err := tx.QueryRow(ctx,
			`SELECT owner_user_id FROM alliances WHERE id = $1 FOR UPDATE`, allianceID,
		).Scan(&owner); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("read alliance: %w", err)
		}
		if owner != ownerID {
			return ErrNotOwner
		}
		// tag валидация (если изменяем).
		if newTag != "" {
			if !isValidTag(newTag) {
				return ErrInvalidTag
			}
			// Уникальность tag.
			var taken bool
			if err := tx.QueryRow(ctx,
				`SELECT EXISTS(SELECT 1 FROM alliances WHERE LOWER(tag) = LOWER($1) AND id <> $2)`,
				newTag, allianceID,
			).Scan(&taken); err != nil {
				return fmt.Errorf("tag check: %w", err)
			}
			if taken {
				return ErrTagTaken
			}
			if _, err := tx.Exec(ctx,
				`UPDATE alliances SET tag = $1 WHERE id = $2`, newTag, allianceID,
			); err != nil {
				return fmt.Errorf("update tag: %w", err)
			}
		}
		if newName != "" {
			// Уникальность name.
			var taken bool
			if err := tx.QueryRow(ctx,
				`SELECT EXISTS(SELECT 1 FROM alliances WHERE LOWER(name) = LOWER($1) AND id <> $2)`,
				newName, allianceID,
			).Scan(&taken); err != nil {
				return fmt.Errorf("name check: %w", err)
			}
			if taken {
				return ErrNameTaken
			}
			// План 46: blacklist.
			if s.blacklist != nil {
				if forbidden, _ := s.blacklist.IsForbidden(newName); forbidden {
					return ErrNameForbidden
				}
			}
			if _, err := tx.Exec(ctx,
				`UPDATE alliances SET name = $1 WHERE id = $2`, newName, allianceID,
			); err != nil {
				return fmt.Errorf("update name: %w", err)
			}
		}
		return nil
	})
}

// isValidTag — 3-5 латинских/цифр (legacy regex).
func isValidTag(tag string) bool {
	if len(tag) < 3 || len(tag) > 5 {
		return false
	}
	for _, r := range tag {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
}

var (
	ErrNotFound          = errors.New("alliance: not found")
	ErrAlreadyMember     = errors.New("alliance: already in an alliance")
	ErrNotMember         = errors.New("alliance: not a member")
	ErrNotOwner          = errors.New("alliance: not the owner")
	ErrTagTaken          = errors.New("alliance: tag already taken")
	ErrNameTaken         = errors.New("alliance: name already taken")
	ErrInvalidTag        = errors.New("alliance: tag must be 3–5 latin letters/digits")
	// План 46: tag или name содержит запрещённое слово.
	ErrNameForbidden     = errors.New("alliance: name contains forbidden word")
	ErrCannotLeaveOwn    = errors.New("alliance: owner must transfer or disband before leaving")
	ErrApplicationExists = errors.New("alliance: application already pending")
	ErrApplicationNotFound = errors.New("alliance: application not found")

	// План 67 Ф.2.
	ErrRankNotFound      = errors.New("alliance: rank not found")
	ErrRankNameTaken     = errors.New("alliance: rank name already taken")
	ErrRankNameInvalid   = errors.New("alliance: rank name must be 1–32 characters")
	ErrInvalidPermission = errors.New("alliance: unknown permission key")
	ErrCannotKickOwner   = errors.New("alliance: cannot kick the owner")
	ErrCannotKickSelf    = errors.New("alliance: use leave to remove yourself")
	ErrDescriptionTooLong = errors.New("alliance: description too long")
)

// Alliance — полная запись для UI.
type Alliance struct {
	ID          string    `json:"id"`
	Tag         string    `json:"tag"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsOpen      bool      `json:"is_open"`
	OwnerID     string    `json:"owner_id"`
	OwnerName   string    `json:"owner_name"`
	MemberCount int       `json:"member_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// Application — заявка на вступление в альянс.
type Application struct {
	ID         string    `json:"id"`
	AllianceID string    `json:"alliance_id"`
	UserID     string    `json:"user_id"`
	Username   string    `json:"username"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"created_at"`
	// План 72.1.45 §3: координаты главной планеты + очки кандидата
	// (legacy candidates view — owner видит куда упасть/забрать).
	HomeGalaxy   int     `json:"home_galaxy"`
	HomeSystem   int     `json:"home_system"`
	HomePosition int     `json:"home_position"`
	Points       float64 `json:"points"`
}

// Member — элемент списка участников.
type Member struct {
	UserID   string    `json:"user_id"`
	Username string    `json:"username"`
	Rank     string    `json:"rank"`
	RankName string    `json:"rank_name"` // произвольный ранг от owner'а
	JoinedAt time.Time `json:"joined_at"`
	// План 72.1.45 §3: online-статус для UI индикатора (legacy memberlist
	// использует last_login/online из user-карточки).
	LastSeen time.Time `json:"last_seen"`
	// Очки: legacy memberlist отображает `points` рядом с username.
	// Источник — users.points (поддерживается score-сервисом).
	Points float64 `json:"points"`
}

var ErrMemberNotFound = errors.New("alliance: member not found")

// ListFilters — параметры фильтрации/поиска для GET /api/alliances
// (план 67 Ф.4, U-012).
type ListFilters struct {
	// Q — полнотекстовая строка поиска (по name+tag, prefix-match).
	// Пустая строка → без фильтра.
	Q string
	// IsOpen — если не nil, фильтрует по alliances.is_open.
	IsOpen *bool
	// MinMembers / MaxMembers — диапазон по числу участников.
	// 0 / 0 = без фильтра. MaxMembers == 0 → без верхней границы.
	MinMembers int
	MaxMembers int
	Limit      int
	Offset     int
}

// List возвращает альянсы с опциональными фильтрами/поиском.
//
// Полнотекст: GIN-индекс на to_tsvector('simple', name||' '||tag),
// миграция 0080. Запрос — to_tsquery('simple', $1 || ':*') для
// prefix-match (пользователь видит результаты по мере набора).
func (s *Service) List(ctx context.Context, f ListFilters) ([]Alliance, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 50
	}
	if f.Offset < 0 {
		f.Offset = 0
	}

	q := `
		SELECT a.id, a.tag, a.name, a.description, a.is_open, a.owner_id,
		       COALESCE(u.username,'') AS owner_name,
		       COUNT(m.user_id)        AS member_count,
		       a.created_at
		FROM alliances a
		LEFT JOIN users u ON u.id = a.owner_id
		LEFT JOIN alliance_members m ON m.alliance_id = a.id`
	args := []any{}
	wheres := []string{}

	add := func(cond string, v any) {
		args = append(args, v)
		wheres = append(wheres, strings.Replace(cond, "?", "$"+strconv.Itoa(len(args)), 1))
	}

	if q := strings.TrimSpace(f.Q); q != "" {
		// to_tsquery требует токенов через &; пользовательский ввод
		// проще передать через plainto_tsquery — но он не делает
		// prefix-match. Используем websearch_to_tsquery (доступен с
		// PG 11) — он экранирует спецсимволы и поддерживает фразы.
		// Для prefix-match — отдельная ветка через to_tsquery с :*
		// если пользователь не ввёл пробел.
		if !strings.ContainsAny(q, " \t") {
			add(`to_tsvector('simple', a.name || ' ' || a.tag) @@ to_tsquery('simple', ? || ':*')`, sanitizeTSQuery(q))
		} else {
			add(`to_tsvector('simple', a.name || ' ' || a.tag) @@ websearch_to_tsquery('simple', ?)`, q)
		}
	}
	if f.IsOpen != nil {
		add("a.is_open = ?", *f.IsOpen)
	}

	if len(wheres) > 0 {
		q += " WHERE " + strings.Join(wheres, " AND ")
	}
	q += " GROUP BY a.id, u.username"

	// HAVING для фильтров по member_count.
	having := []string{}
	if f.MinMembers > 0 {
		args = append(args, f.MinMembers)
		having = append(having, "COUNT(m.user_id) >= $"+strconv.Itoa(len(args)))
	}
	if f.MaxMembers > 0 {
		args = append(args, f.MaxMembers)
		having = append(having, "COUNT(m.user_id) <= $"+strconv.Itoa(len(args)))
	}
	if len(having) > 0 {
		q += " HAVING " + strings.Join(having, " AND ")
	}

	q += " ORDER BY member_count DESC, a.created_at ASC"
	args = append(args, f.Limit)
	q += " LIMIT $" + strconv.Itoa(len(args))
	args = append(args, f.Offset)
	q += " OFFSET $" + strconv.Itoa(len(args))

	rows, err := s.db.Pool().Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("alliances list: %w", err)
	}
	defer rows.Close()
	out := []Alliance{}
	for rows.Next() {
		var al Alliance
		if err := rows.Scan(&al.ID, &al.Tag, &al.Name, &al.Description, &al.IsOpen,
			&al.OwnerID, &al.OwnerName, &al.MemberCount, &al.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, al)
	}
	return out, rows.Err()
}

// sanitizeTSQuery убирает символы, которые to_tsquery интерпретирует
// как операторы (& | ! ( ) :), оставляя только буквы/цифры. Для
// websearch_to_tsquery санитайзинг не нужен — он сам справляется.
func sanitizeTSQuery(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r >= 'А' && r <= 'я',
			r == 'ё', r == 'Ё':
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Get возвращает альянс по ID вместе со списком участников.
func (s *Service) Get(ctx context.Context, id string) (Alliance, []Member, error) {
	var al Alliance
	err := s.db.Pool().QueryRow(ctx, `
		SELECT a.id, a.tag, a.name, a.description, a.is_open, a.owner_id,
		       COALESCE(u.username,'') AS owner_name,
		       (SELECT COUNT(*) FROM alliance_members WHERE alliance_id=a.id),
		       a.created_at
		FROM alliances a
		LEFT JOIN users u ON u.id = a.owner_id
		WHERE a.id = $1
	`, id).Scan(&al.ID, &al.Tag, &al.Name, &al.Description, &al.IsOpen,
		&al.OwnerID, &al.OwnerName, &al.MemberCount, &al.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Alliance{}, nil, ErrNotFound
		}
		return Alliance{}, nil, fmt.Errorf("get alliance: %w", err)
	}

	// План 72.1.45 §3: добавили u.last_seen + u.points (агрегированный
	// score из service score.go). legacy memberlist отображает рядом с
	// username.
	rows, err := s.db.Pool().Query(ctx, `
		SELECT m.user_id, COALESCE(u.username,''), m.rank, m.rank_name, m.joined_at,
		       u.last_seen,
		       COALESCE(u.points, 0) AS points
		  FROM alliance_members m
		  JOIN users u ON u.id = m.user_id
		 WHERE m.alliance_id = $1
		 ORDER BY
		    CASE m.rank WHEN 'owner' THEN 0 ELSE 1 END,
		    m.joined_at ASC
	`, id)
	if err != nil {
		return Alliance{}, nil, fmt.Errorf("get members: %w", err)
	}
	defer rows.Close()
	var members []Member
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.UserID, &m.Username, &m.Rank, &m.RankName, &m.JoinedAt, &m.LastSeen, &m.Points); err != nil {
			return Alliance{}, nil, err
		}
		members = append(members, m)
	}
	return al, members, rows.Err()
}

// Create создаёт новый альянс и делает создателя owner'ом.
func (s *Service) Create(ctx context.Context, ownerID, tag, name, description string) (Alliance, error) {
	tag = strings.TrimSpace(tag)
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	if err := validateTag(tag); err != nil {
		return Alliance{}, err
	}
	if n := utf8.RuneCountInString(name); n < 3 || n > 64 {
		return Alliance{}, fmt.Errorf("alliance: name must be 3–64 characters")
	}
	// План 46 (149-ФЗ): проверка tag/name по UGC-blacklist.
	if s.blacklist != nil {
		if forbidden, _ := s.blacklist.IsForbidden(tag); forbidden {
			return Alliance{}, ErrNameForbidden
		}
		if forbidden, _ := s.blacklist.IsForbidden(name); forbidden {
			return Alliance{}, ErrNameForbidden
		}
	}
	if utf8.RuneCountInString(description) > 2000 {
		description = string([]rune(description)[:2000])
	}

	var out Alliance
	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Проверяем что пользователь не в альянсе.
		var existing *string
		if err := tx.QueryRow(ctx,
			`SELECT alliance_id FROM users WHERE id=$1`, ownerID).Scan(&existing); err != nil {
			return fmt.Errorf("check user: %w", err)
		}
		if existing != nil {
			return ErrAlreadyMember
		}

		var id string
		err := tx.QueryRow(ctx, `
			INSERT INTO alliances (tag, name, description, owner_id)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, tag, name, description, ownerID).Scan(&id)
		if err != nil {
			if isDupKey(err) {
				if strings.Contains(err.Error(), "alliances_tag_key") {
					return ErrTagTaken
				}
				return ErrNameTaken
			}
			return fmt.Errorf("insert alliance: %w", err)
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO alliance_members (alliance_id, user_id, rank)
			VALUES ($1, $2, 'owner')
		`, id, ownerID); err != nil {
			return fmt.Errorf("insert member: %w", err)
		}

		if _, err := tx.Exec(ctx,
			`UPDATE users SET alliance_id=$1 WHERE id=$2`, id, ownerID); err != nil {
			return fmt.Errorf("update user: %w", err)
		}

		var ownerName string
		_ = tx.QueryRow(ctx, `SELECT username FROM users WHERE id=$1`, ownerID).Scan(&ownerName)
		out = Alliance{
			ID: id, Tag: tag, Name: name, Description: description,
			OwnerID: ownerID, OwnerName: ownerName, MemberCount: 1,
		}
		writeAuditTx(ctx, tx, id, ownerID, ActionAllianceCreated, TargetKindAlliance, id,
			map[string]any{"tag": tag, "name": name})
		return nil
	})
	return out, err
}

// Description — три варианта описаний альянса (план 67 Ф.2, D-041, U-015).
//
// External — публичное (видят все, в т.ч. не члены).
// Internal — для членов альянса.
// Apply    — для подающих заявку (заполняется при Join).
type Description struct {
	External string `json:"description_external"`
	Internal string `json:"description_internal"`
	Apply    string `json:"description_apply"`
}

// DescriptionView — описание + ссылка на legacy-поле и контекст
// доступа: какие из 3 полей вернутся, зависит от того, кем
// запрашивает пользователь (член/заявитель/посторонний).
type DescriptionView struct {
	// Все три поля присутствуют только для members. Для гостей и
	// заявителей лишние поля будут пустыми (но JSON-форма стабильна).
	External string `json:"description_external"`
	Internal string `json:"description_internal"`
	Apply    string `json:"description_apply"`
	// Legacy: alliances.description, оставлено для обратной совместимости
	// (R0 — старые UI/scripts могут читать). Новые клиенты должны
	// использовать description_external.
	Legacy string `json:"description"`
	// Viewer — кем запрашивающий приходит к этому альянсу:
	//   "member" — состоит в альянсе,
	//   "applicant" — есть pending-заявка,
	//   "outsider" — все остальные.
	Viewer string `json:"viewer"`
}

// GetDescriptions возвращает описания, фильтруя по контексту доступа.
//
// requesterID может быть "" для анонимного доступа (увидит только
// description_external + legacy).
func (s *Service) GetDescriptions(ctx context.Context, requesterID, allianceID string) (DescriptionView, error) {
	var v DescriptionView
	var ext, intDesc, apply, legacy *string
	err := s.db.Pool().QueryRow(ctx, `
		SELECT description_external, description_internal,
		       description_apply, description
		FROM alliances WHERE id=$1
	`, allianceID).Scan(&ext, &intDesc, &apply, &legacy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return v, ErrNotFound
		}
		return v, fmt.Errorf("get descriptions: %w", err)
	}
	v.Legacy = strDeref(legacy)
	v.External = strDeref(ext)

	// Определяем контекст вьюера.
	v.Viewer = "outsider"
	if requesterID != "" {
		var memAlliance *string
		if err := s.db.Pool().QueryRow(ctx,
			`SELECT alliance_id FROM alliance_members WHERE user_id=$1`, requesterID).Scan(&memAlliance); err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return v, fmt.Errorf("get descriptions: check member: %w", err)
			}
		}
		if memAlliance != nil && *memAlliance == allianceID {
			v.Viewer = "member"
			v.Internal = strDeref(intDesc)
			v.Apply = strDeref(apply)
			return v, nil
		}
		var pending bool
		if err := s.db.Pool().QueryRow(ctx, `
			SELECT EXISTS (SELECT 1 FROM alliance_applications
				WHERE alliance_id=$1 AND user_id=$2)
		`, allianceID, requesterID).Scan(&pending); err != nil {
			return v, fmt.Errorf("get descriptions: check pending: %w", err)
		}
		if pending {
			v.Viewer = "applicant"
			v.Apply = strDeref(apply)
			return v, nil
		}
	}
	return v, nil
}

// UpdateDescriptionsInput — частичное обновление: nil = поле не трогаем.
type UpdateDescriptionsInput struct {
	External *string
	Internal *string
	Apply    *string
}

// UpdateDescriptions PATCH /api/alliances/{id}/descriptions
//
// Требует can_change_description (или owner). Каждое из 3 полей
// обновляется независимо (snake_case по R1). Длина каждого поля
// ограничена 4000 символами.
func (s *Service) UpdateDescriptions(ctx context.Context, requesterID, allianceID string, in UpdateDescriptionsInput) error {
	const maxLen = 4000
	if in.External != nil && utf8.RuneCountInString(*in.External) > maxLen {
		return ErrDescriptionTooLong
	}
	if in.Internal != nil && utf8.RuneCountInString(*in.Internal) > maxLen {
		return ErrDescriptionTooLong
	}
	if in.Apply != nil && utf8.RuneCountInString(*in.Apply) > maxLen {
		return ErrDescriptionTooLong
	}

	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		mem, err := LoadMembership(ctx, tx, requesterID, allianceID)
		if err != nil {
			return err
		}
		if mem == nil {
			return ErrNotMember
		}
		ok, err := Has(ctx, tx, mem, PermChangeDescription)
		if err != nil {
			return err
		}
		if !ok {
			return ErrForbidden
		}

		// Собираем UPDATE с COALESCE: nil — не трогаем.
		args := []any{allianceID}
		set := []string{}
		changed := map[string]bool{}
		add := func(col string, val *string) {
			if val == nil {
				return
			}
			args = append(args, *val)
			set = append(set, col+"=$"+strconv.Itoa(len(args)))
			changed[col] = true
		}
		add("description_external", in.External)
		add("description_internal", in.Internal)
		add("description_apply", in.Apply)
		if len(set) == 0 {
			return nil
		}
		q := "UPDATE alliances SET " + strings.Join(set, ", ") + " WHERE id=$1"
		tag, err := tx.Exec(ctx, q, args...)
		if err != nil {
			return fmt.Errorf("update descriptions: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return ErrNotFound
		}

		writeAuditTx(ctx, tx, allianceID, requesterID, ActionDescriptionChanged,
			TargetKindAlliance, allianceID, map[string]any{"fields": keysOf(changed)})
		return nil
	})
}

func strDeref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func keysOf(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// Join добавляет пользователя в альянс. Если альянс закрыт (is_open=false),
// создаётся заявка; owner должен вызвать Approve/Reject.
// Возвращает (true, nil) при прямом вступлении, (false, nil) при заявке.
func (s *Service) Join(ctx context.Context, userID, allianceID, message string) (joined bool, err error) {
	var ownerID, applicantName, allianceTag string
	err = s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var existing *string
		if err := tx.QueryRow(ctx,
			`SELECT alliance_id FROM users WHERE id=$1`, userID).Scan(&existing); err != nil {
			return fmt.Errorf("check user: %w", err)
		}
		if existing != nil {
			return ErrAlreadyMember
		}
		var isOpen bool
		err := tx.QueryRow(ctx,
			`SELECT is_open, owner_id, tag FROM alliances WHERE id=$1`,
			allianceID).Scan(&isOpen, &ownerID, &allianceTag)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("check alliance: %w", err)
		}
		if err := tx.QueryRow(ctx,
			`SELECT username FROM users WHERE id=$1`, userID).Scan(&applicantName); err != nil {
			return fmt.Errorf("read applicant: %w", err)
		}

		if !isOpen {
			// Создаём заявку.
			_, err := tx.Exec(ctx, `
				INSERT INTO alliance_applications (alliance_id, user_id, message)
				VALUES ($1, $2, $3)
			`, allianceID, userID, message)
			if err != nil {
				if isDupKey(err) {
					return ErrApplicationExists
				}
				return fmt.Errorf("insert application: %w", err)
			}
			joined = false
			return nil
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO alliance_members (alliance_id, user_id, rank)
			VALUES ($1, $2, 'member')
		`, allianceID, userID); err != nil {
			return fmt.Errorf("insert member: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`UPDATE users SET alliance_id=$1 WHERE id=$2`, allianceID, userID); err != nil {
			return fmt.Errorf("update user: %w", err)
		}
		joined = true
		return nil
	})
	if err != nil {
		return joined, err
	}
	// Уведомления — вне транзакции (best-effort).
	if joined {
		s.notifyAlliance(ctx, userID,
			s.tr("alliance", "joined.title", nil),
			s.tr("alliance", "joined.body", map[string]string{"allianceName": allianceTag}))
	} else {
		s.notifyAlliance(ctx, ownerID,
			s.tr("alliance", "application.title", nil),
			s.tr("alliance", "application.body", map[string]string{
				"username":     applicantName,
				"allianceName": allianceTag,
			}))
	}
	return joined, nil
}

// SetOpen меняет флаг is_open. Только owner.
func (s *Service) SetOpen(ctx context.Context, userID, allianceID string, isOpen bool) error {
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var ownerID string
		err := tx.QueryRow(ctx, `SELECT owner_id FROM alliances WHERE id=$1`, allianceID).Scan(&ownerID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("check alliance: %w", err)
		}
		if ownerID != userID {
			return ErrNotOwner
		}
		if _, err := tx.Exec(ctx, `UPDATE alliances SET is_open=$1 WHERE id=$2`, isOpen, allianceID); err != nil {
			return err
		}
		writeAuditTx(ctx, tx, allianceID, userID, ActionOpenChanged,
			TargetKindAlliance, allianceID, map[string]any{"is_open": isOpen})
		return nil
	})
}

// Applications возвращает список заявок альянса. Только owner.
func (s *Service) Applications(ctx context.Context, userID, allianceID string) ([]Application, error) {
	var ownerID string
	if err := s.db.Pool().QueryRow(ctx,
		`SELECT owner_id FROM alliances WHERE id=$1`, allianceID).Scan(&ownerID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("check alliance: %w", err)
	}
	if ownerID != userID {
		return nil, ErrNotOwner
	}
	// План 72.1.45 §3: добавили координаты home-планеты (subquery
	// homeworld → coords) + очки. legacy candidates view это показывает.
	rows, err := s.db.Pool().Query(ctx, `
		SELECT a.id, a.alliance_id, a.user_id, COALESCE(u.username,''), a.message, a.created_at,
		       COALESCE(p.galaxy, 0), COALESCE(p.system, 0), COALESCE(p.position, 0),
		       COALESCE(u.points, 0) AS points
		  FROM alliance_applications a
		  JOIN users u ON u.id = a.user_id
		  LEFT JOIN LATERAL (
		      SELECT galaxy, system, position
		        FROM planets
		       WHERE user_id = a.user_id AND is_moon = false
		       ORDER BY created_at ASC
		       LIMIT 1
		  ) p ON true
		 WHERE a.alliance_id = $1
		 ORDER BY a.created_at ASC
	`, allianceID)
	if err != nil {
		return nil, fmt.Errorf("list applications: %w", err)
	}
	defer rows.Close()
	var out []Application
	for rows.Next() {
		var ap Application
		if err := rows.Scan(&ap.ID, &ap.AllianceID, &ap.UserID, &ap.Username,
			&ap.Message, &ap.CreatedAt,
			&ap.HomeGalaxy, &ap.HomeSystem, &ap.HomePosition, &ap.Points); err != nil {
			return nil, err
		}
		out = append(out, ap)
	}
	return out, rows.Err()
}

// Approve принимает заявку, добавляет участника. Только owner.
func (s *Service) Approve(ctx context.Context, ownerID, applicationID string) error {
	var applicantID, allianceTag string
	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var allianceID string
		err := tx.QueryRow(ctx,
			`SELECT alliance_id, user_id FROM alliance_applications WHERE id=$1`,
			applicationID).Scan(&allianceID, &applicantID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrApplicationNotFound
			}
			return fmt.Errorf("read application: %w", err)
		}
		var allianceOwner string
		if err := tx.QueryRow(ctx, `SELECT owner_id, tag FROM alliances WHERE id=$1`,
			allianceID).Scan(&allianceOwner, &allianceTag); err != nil {
			return fmt.Errorf("check alliance: %w", err)
		}
		if allianceOwner != ownerID {
			return ErrNotOwner
		}
		// Check applicant still free.
		var existing *string
		if err := tx.QueryRow(ctx,
			`SELECT alliance_id FROM users WHERE id=$1`, applicantID).Scan(&existing); err != nil {
			return fmt.Errorf("check applicant: %w", err)
		}
		if existing != nil {
			// Applicant joined elsewhere — clean up and return error.
			_, _ = tx.Exec(ctx, `DELETE FROM alliance_applications WHERE id=$1`, applicationID)
			return ErrAlreadyMember
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO alliance_members (alliance_id, user_id, rank)
			VALUES ($1, $2, 'member')
		`, allianceID, applicantID); err != nil {
			return fmt.Errorf("insert member: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`UPDATE users SET alliance_id=$1 WHERE id=$2`, allianceID, applicantID); err != nil {
			return fmt.Errorf("update user: %w", err)
		}
		if _, err = tx.Exec(ctx, `DELETE FROM alliance_applications WHERE id=$1`, applicationID); err != nil {
			return err
		}
		writeAuditTx(ctx, tx, allianceID, ownerID, ActionApplicationApproved,
			TargetKindUser, applicantID, nil)
		writeAuditTx(ctx, tx, allianceID, ownerID, ActionMemberJoined,
			TargetKindUser, applicantID, map[string]any{"via": "application"})
		return nil
	})
	if err != nil {
		return err
	}
	s.notifyAlliance(ctx, applicantID,
		s.tr("alliance", "approved.title", nil),
		s.tr("alliance", "approved.body", map[string]string{"allianceName": allianceTag}))
	return nil
}

// Reject удаляет заявку. Только owner.
func (s *Service) Reject(ctx context.Context, ownerID, applicationID string) error {
	var applicantID, allianceTag string
	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var allianceID string
		err := tx.QueryRow(ctx,
			`SELECT alliance_id, user_id FROM alliance_applications WHERE id=$1`,
			applicationID).Scan(&allianceID, &applicantID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrApplicationNotFound
			}
			return fmt.Errorf("read application: %w", err)
		}
		var allianceOwner string
		if err := tx.QueryRow(ctx, `SELECT owner_id, tag FROM alliances WHERE id=$1`,
			allianceID).Scan(&allianceOwner, &allianceTag); err != nil {
			return fmt.Errorf("check alliance: %w", err)
		}
		if allianceOwner != ownerID {
			return ErrNotOwner
		}
		if _, err = tx.Exec(ctx, `DELETE FROM alliance_applications WHERE id=$1`, applicationID); err != nil {
			return err
		}
		writeAuditTx(ctx, tx, allianceID, ownerID, ActionApplicationRejected,
			TargetKindUser, applicantID, nil)
		return nil
	})
	if err != nil {
		return err
	}
	s.notifyAlliance(ctx, applicantID,
		s.tr("alliance", "rejected.title", nil),
		s.tr("alliance", "rejected.body", map[string]string{"allianceName": allianceTag}))
	return nil
}

// Leave удаляет пользователя из альянса. Owner не может выйти
// (должен сначала передать права или распустить альянс).
func (s *Service) Leave(ctx context.Context, userID string) error {
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var rank string
		var allianceID string
		err := tx.QueryRow(ctx,
			`SELECT alliance_id, rank FROM alliance_members WHERE user_id=$1`,
			userID).Scan(&allianceID, &rank)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotMember
			}
			return fmt.Errorf("check member: %w", err)
		}
		if rank == "owner" {
			return ErrCannotLeaveOwn
		}

		if _, err := tx.Exec(ctx,
			`DELETE FROM alliance_members WHERE user_id=$1`, userID); err != nil {
			return fmt.Errorf("delete member: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`UPDATE users SET alliance_id=NULL WHERE id=$1`, userID); err != nil {
			return fmt.Errorf("update user: %w", err)
		}
		writeAuditTx(ctx, tx, allianceID, userID, ActionMemberLeft,
			TargetKindUser, userID, nil)
		return nil
	})
}

// Kick удаляет участника. Право: PermKick (или owner).
//
// Нельзя кикнуть owner'а и самого себя (для самоудаления — Leave).
func (s *Service) Kick(ctx context.Context, requesterID, allianceID, memberUserID string) error {
	if requesterID == memberUserID {
		return ErrCannotKickSelf
	}
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		mem, err := LoadMembership(ctx, tx, requesterID, allianceID)
		if err != nil {
			return err
		}
		if mem == nil {
			return ErrNotMember
		}
		can, err := Has(ctx, tx, mem, PermKick)
		if err != nil {
			return err
		}
		if !can {
			return ErrForbidden
		}

		var targetRank string
		err = tx.QueryRow(ctx,
			`SELECT rank FROM alliance_members WHERE alliance_id=$1 AND user_id=$2`,
			allianceID, memberUserID).Scan(&targetRank)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrMemberNotFound
			}
			return fmt.Errorf("kick: read member: %w", err)
		}
		if targetRank == "owner" {
			return ErrCannotKickOwner
		}

		if _, err := tx.Exec(ctx,
			`DELETE FROM alliance_members WHERE alliance_id=$1 AND user_id=$2`,
			allianceID, memberUserID); err != nil {
			return fmt.Errorf("kick: delete: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`UPDATE users SET alliance_id=NULL WHERE id=$1`, memberUserID); err != nil {
			return fmt.Errorf("kick: update user: %w", err)
		}
		writeAuditTx(ctx, tx, allianceID, requesterID, ActionMemberKicked,
			TargetKindUser, memberUserID, nil)
		return nil
	})
}

// Disband распускает альянс. Только owner.
func (s *Service) Disband(ctx context.Context, userID, allianceID string) error {
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var ownerID string
		err := tx.QueryRow(ctx,
			`SELECT owner_id FROM alliances WHERE id=$1`, allianceID).Scan(&ownerID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("check alliance: %w", err)
		}
		if ownerID != userID {
			return ErrNotOwner
		}
		// ON DELETE CASCADE уберёт alliance_members.
		// ON DELETE SET NULL уберёт users.alliance_id.
		if _, err := tx.Exec(ctx,
			`DELETE FROM alliances WHERE id=$1`, allianceID); err != nil {
			return fmt.Errorf("disband: %w", err)
		}
		return nil
	})
}

// MyAlliance возвращает альянс текущего пользователя (если есть).
func (s *Service) MyAlliance(ctx context.Context, userID string) (*Alliance, []Member, error) {
	var allianceID *string
	if err := s.db.Pool().QueryRow(ctx,
		`SELECT alliance_id FROM users WHERE id=$1`, userID).Scan(&allianceID); err != nil {
		return nil, nil, fmt.Errorf("my alliance: %w", err)
	}
	if allianceID == nil {
		return nil, nil, nil
	}
	al, members, err := s.Get(ctx, *allianceID)
	if err != nil {
		return nil, nil, err
	}
	return &al, members, nil
}

// SetMemberRank устанавливает отображаемый ранг участника. Только owner.
// rankName — произвольный текст (до 32 символов). Пустая строка сбрасывает ранг.
func (s *Service) SetMemberRank(ctx context.Context, ownerID, allianceID, memberUserID, rankName string) error {
	if utf8.RuneCountInString(rankName) > 32 {
		rankName = string([]rune(rankName)[:32])
	}
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var allianceOwner string
		if err := tx.QueryRow(ctx,
			`SELECT owner_id FROM alliances WHERE id=$1`, allianceID).Scan(&allianceOwner); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("set rank: %w", err)
		}
		if allianceOwner != ownerID {
			return ErrNotOwner
		}
		tag, err := tx.Exec(ctx,
			`UPDATE alliance_members SET rank_name=$1 WHERE alliance_id=$2 AND user_id=$3`,
			rankName, allianceID, memberUserID)
		if err != nil {
			return fmt.Errorf("set rank: update: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return ErrMemberNotFound
		}
		return nil
	})
}

func validateTag(tag string) error {
	n := utf8.RuneCountInString(tag)
	if n < 3 || n > 5 {
		return ErrInvalidTag
	}
	for _, r := range tag {
		if !isAlphanumASCII(r) {
			return ErrInvalidTag
		}
	}
	return nil
}

func isAlphanumASCII(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
}

func isDupKey(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate key")
}

// Relationship — запись об отношениях между альянсами.
type Relationship struct {
	TargetAllianceID string    `json:"target_alliance_id"`
	TargetTag        string    `json:"target_tag"`
	TargetName       string    `json:"target_name"`
	Relation         string    `json:"relation"` // "nap" | "war" | "ally"
	Status           string    `json:"status"`   // "pending" | "active"
	Initiator        bool      `json:"initiator"` // true если мы инициатор
	SetAt            time.Time `json:"set_at"`
}

var (
	ErrInvalidRelation  = errors.New("alliance: relation must be one of: friend, neutral, hostile_neutral, nap, war, none")
	ErrTargetNotFound   = errors.New("alliance: target alliance not found")
	ErrRelationSelf     = errors.New("alliance: cannot set relation with own alliance")
	ErrRelationPending  = errors.New("alliance: relation proposal already pending")
)

// validRelations — расширенный enum (план 67 Ф.2, D-014, B1).
//   - friend         — союз (origin: ally → friend; миграция 0077),
//   - neutral        — нейтрал,
//   - hostile_neutral — враждебный нейтрал (атаковать без объявления войны
//                      разрешено),
//   - nap            — non-aggression pact,
//   - war            — открытая война.
// "ally" принимаем как алиас "friend" для обратной совместимости с
// фикстурами/клиентами, не успевшими обновиться (миграция 0077 уже
// перевела данные).
// "none" — снять отношение (не enum-значение; обрабатывается отдельно).
var validRelations = map[string]string{
	"friend":          "friend",
	"neutral":         "neutral",
	"hostile_neutral": "hostile_neutral",
	"nap":             "nap",
	"war":             "war",
	"ally":            "friend",
	"none":            "none",
}

// normalizeRelation возвращает (canonical, ok). Если ok=false — relation
// не принимается. Для "none" canonical="none".
func normalizeRelation(relation string) (string, bool) {
	v, ok := validRelations[relation]
	return v, ok
}

// relationNeedsAccept — true если статус требует подтверждения target'а.
// "war" и "hostile_neutral" — односторонние; остальные двусторонние.
func relationNeedsAccept(relation string) bool {
	switch relation {
	case "war", "hostile_neutral":
		return false
	default:
		return true
	}
}

// ProposeRelation предлагает отношение от allianceID к targetID.
//
// Enum (план 67 Ф.2): friend / neutral / hostile_neutral / nap / war / none.
// Односторонние (war, hostile_neutral) активны сразу; остальные ждут
// подтверждения от target. relation="none" удаляет записи в обе стороны.
//
// Право: PermProposeRelations (или owner). Если у пользователя нет этого
// права — ErrForbidden.
func (s *Service) ProposeRelation(ctx context.Context, userID, allianceID, targetID, relation string) error {
	if allianceID == targetID {
		return ErrRelationSelf
	}
	canonical, ok := normalizeRelation(relation)
	if !ok {
		return ErrInvalidRelation
	}

	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		mem, err := LoadMembership(ctx, tx, userID, allianceID)
		if err != nil {
			return err
		}
		if mem == nil {
			return ErrNotMember
		}
		// Manage и Propose — два разных уровня (Manage = принимать/отклонять,
		// Propose = предлагать). Для совместимости с owner-only флоу до плана
		// 67: owner всегда имеет оба.
		can, err := Has(ctx, tx, mem, PermProposeRelations)
		if err != nil {
			return err
		}
		if !can {
			return ErrForbidden
		}

		var targetExists bool
		if err := tx.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM alliances WHERE id=$1)`, targetID).Scan(&targetExists); err != nil {
			return fmt.Errorf("propose relation: check target: %w", err)
		}
		if !targetExists {
			return ErrTargetNotFound
		}

		if canonical == "none" {
			if _, err := tx.Exec(ctx, `
				DELETE FROM alliance_relationships
				WHERE (alliance_id=$1 AND target_alliance_id=$2)
				   OR (alliance_id=$2 AND target_alliance_id=$1)
			`, allianceID, targetID); err != nil {
				return fmt.Errorf("propose relation: delete: %w", err)
			}
			writeAuditTx(ctx, tx, allianceID, userID, ActionRelationCleared,
				TargetKindAlliance, targetID, nil)
			return nil
		}

		status := "pending"
		if !relationNeedsAccept(canonical) {
			status = "active"
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO alliance_relationships (alliance_id, target_alliance_id, relation, status)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (alliance_id, target_alliance_id)
			DO UPDATE SET relation=$3, status=$4, set_at=now()
		`, allianceID, targetID, canonical, status); err != nil {
			return fmt.Errorf("propose relation: upsert: %w", err)
		}
		writeAuditTx(ctx, tx, allianceID, userID, ActionRelationProposed,
			TargetKindAlliance, targetID,
			map[string]any{"relation": canonical, "status": status})
		return nil
	})
}

// AcceptRelation подтверждает входящее предложение (двустороннее:
// friend/neutral/nap). После accept — обе записи становятся active.
//
// Право: PermManageDiplomacy (или owner).
func (s *Service) AcceptRelation(ctx context.Context, userID, myAllianceID, initiatorAllianceID string) error {
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		mem, err := LoadMembership(ctx, tx, userID, myAllianceID)
		if err != nil {
			return err
		}
		if mem == nil {
			return ErrNotMember
		}
		can, err := Has(ctx, tx, mem, PermManageDiplomacy)
		if err != nil {
			return err
		}
		if !can {
			return ErrForbidden
		}

		// Читаем pending предложение от initiatorAllianceID → myAllianceID.
		var relation string
		var status string
		err = tx.QueryRow(ctx, `
			SELECT relation::text, status
			FROM alliance_relationships
			WHERE alliance_id=$1 AND target_alliance_id=$2
		`, initiatorAllianceID, myAllianceID).Scan(&relation, &status)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrTargetNotFound
			}
			return fmt.Errorf("accept relation: read proposal: %w", err)
		}
		if status != "pending" {
			return fmt.Errorf("accept relation: proposal is not pending")
		}

		// Активируем инициаторскую запись.
		if _, err := tx.Exec(ctx, `
			UPDATE alliance_relationships SET status='active'
			WHERE alliance_id=$1 AND target_alliance_id=$2
		`, initiatorAllianceID, myAllianceID); err != nil {
			return fmt.Errorf("accept relation: update initiator: %w", err)
		}

		// Создаём зеркальную активную запись для нашего альянса.
		if _, err := tx.Exec(ctx, `
			INSERT INTO alliance_relationships (alliance_id, target_alliance_id, relation, status)
			VALUES ($1, $2, $3, 'active')
			ON CONFLICT (alliance_id, target_alliance_id)
			DO UPDATE SET relation=$3, status='active', set_at=now()
		`, myAllianceID, initiatorAllianceID, relation); err != nil {
			return fmt.Errorf("accept relation: insert mirror: %w", err)
		}
		writeAuditTx(ctx, tx, myAllianceID, userID, ActionRelationAccepted,
			TargetKindAlliance, initiatorAllianceID,
			map[string]any{"relation": relation})
		return nil
	})
}

// RejectRelation отклоняет входящее предложение — удаляет pending запись.
//
// Право: PermManageDiplomacy (или owner).
func (s *Service) RejectRelation(ctx context.Context, userID, myAllianceID, initiatorAllianceID string) error {
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		mem, err := LoadMembership(ctx, tx, userID, myAllianceID)
		if err != nil {
			return err
		}
		if mem == nil {
			return ErrNotMember
		}
		can, err := Has(ctx, tx, mem, PermManageDiplomacy)
		if err != nil {
			return err
		}
		if !can {
			return ErrForbidden
		}

		if _, err := tx.Exec(ctx, `
			DELETE FROM alliance_relationships
			WHERE alliance_id=$1 AND target_alliance_id=$2 AND status='pending'
		`, initiatorAllianceID, myAllianceID); err != nil {
			return fmt.Errorf("reject relation: delete: %w", err)
		}
		writeAuditTx(ctx, tx, myAllianceID, userID, ActionRelationRejected,
			TargetKindAlliance, initiatorAllianceID, nil)
		return nil
	})
}

// GetRelations возвращает все отношения альянса allianceID:
// исходящие (initiator=true) и входящие pending (initiator=false).
func (s *Service) GetRelations(ctx context.Context, allianceID string) ([]Relationship, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT r.target_alliance_id, a.tag, a.name, r.relation::text, r.status, true AS initiator, r.set_at
		FROM alliance_relationships r
		JOIN alliances a ON a.id = r.target_alliance_id
		WHERE r.alliance_id = $1
		UNION ALL
		SELECT r.alliance_id, a.tag, a.name, r.relation::text, r.status, false AS initiator, r.set_at
		FROM alliance_relationships r
		JOIN alliances a ON a.id = r.alliance_id
		WHERE r.target_alliance_id = $1 AND r.status = 'pending'
		ORDER BY set_at DESC
	`, allianceID)
	if err != nil {
		return nil, fmt.Errorf("get relations: %w", err)
	}
	defer rows.Close()
	var out []Relationship
	for rows.Next() {
		var rel Relationship
		if err := rows.Scan(&rel.TargetAllianceID, &rel.TargetTag, &rel.TargetName,
			&rel.Relation, &rel.Status, &rel.Initiator, &rel.SetAt); err != nil {
			return nil, err
		}
		out = append(out, rel)
	}
	return out, rows.Err()
}
