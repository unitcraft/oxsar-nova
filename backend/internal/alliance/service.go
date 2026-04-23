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
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/repo"
)

// AutoMsgSender — узкий интерфейс к automsg.SendDirect.
type AutoMsgSender interface {
	SendDirect(ctx context.Context, tx pgx.Tx, userID string, folder int, title, body string) error
}

type Service struct {
	db      repo.Exec
	automsg AutoMsgSender
}

func NewService(db repo.Exec) *Service { return &Service{db: db} }

// WithAutoMsg подключает сервис системных сообщений (опционально).
func (s *Service) WithAutoMsg(a AutoMsgSender) *Service {
	s.automsg = a
	return s
}

// notifyAlliance (folder=6 MSG_FOLDER_ALLIANCE) — помощник для рассылки
// уведомления конкретному пользователю. Ошибки глотаются (не критично).
func (s *Service) notifyAlliance(ctx context.Context, userID, title, body string) {
	if s.automsg == nil {
		return
	}
	_ = s.automsg.SendDirect(ctx, nil, userID, 6, title, body)
}

var (
	ErrNotFound          = errors.New("alliance: not found")
	ErrAlreadyMember     = errors.New("alliance: already in an alliance")
	ErrNotMember         = errors.New("alliance: not a member")
	ErrNotOwner          = errors.New("alliance: not the owner")
	ErrTagTaken          = errors.New("alliance: tag already taken")
	ErrNameTaken         = errors.New("alliance: name already taken")
	ErrInvalidTag        = errors.New("alliance: tag must be 3–5 latin letters/digits")
	ErrCannotLeaveOwn    = errors.New("alliance: owner must transfer or disband before leaving")
	ErrApplicationExists = errors.New("alliance: application already pending")
	ErrApplicationNotFound = errors.New("alliance: application not found")
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
	ID          string    `json:"id"`
	AllianceID  string    `json:"alliance_id"`
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"created_at"`
}

// Member — элемент списка участников.
type Member struct {
	UserID   string    `json:"user_id"`
	Username string    `json:"username"`
	Rank     string    `json:"rank"`
	RankName string    `json:"rank_name"` // произвольный ранг от owner'а
	JoinedAt time.Time `json:"joined_at"`
}

var ErrMemberNotFound = errors.New("alliance: member not found")

// List возвращает первые N альянсов, сортировка по числу участников.
func (s *Service) List(ctx context.Context, limit int) ([]Alliance, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := s.db.Pool().Query(ctx, `
		SELECT a.id, a.tag, a.name, a.description, a.is_open, a.owner_id,
		       COALESCE(u.username,'') AS owner_name,
		       COUNT(m.user_id)        AS member_count,
		       a.created_at
		FROM alliances a
		LEFT JOIN users u ON u.id = a.owner_id
		LEFT JOIN alliance_members m ON m.alliance_id = a.id
		GROUP BY a.id, u.username
		ORDER BY member_count DESC, a.created_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("alliances list: %w", err)
	}
	defer rows.Close()
	var out []Alliance
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

	rows, err := s.db.Pool().Query(ctx, `
		SELECT m.user_id, COALESCE(u.username,''), m.rank, m.rank_name, m.joined_at
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
		if err := rows.Scan(&m.UserID, &m.Username, &m.Rank, &m.RankName, &m.JoinedAt); err != nil {
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
		return nil
	})
	return out, err
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
			"Вы вступили в альянс",
			fmt.Sprintf("Вы стали членом альянса [%s].", allianceTag))
	} else {
		s.notifyAlliance(ctx, ownerID,
			"Заявка на вступление в альянс",
			fmt.Sprintf("Игрок %s подал заявку на вступление в ваш альянс [%s].", applicantName, allianceTag))
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
		_, err = tx.Exec(ctx, `UPDATE alliances SET is_open=$1 WHERE id=$2`, isOpen, allianceID)
		return err
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
	rows, err := s.db.Pool().Query(ctx, `
		SELECT a.id, a.alliance_id, a.user_id, COALESCE(u.username,''), a.message, a.created_at
		FROM alliance_applications a
		JOIN users u ON u.id = a.user_id
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
			&ap.Message, &ap.CreatedAt); err != nil {
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
		_, err = tx.Exec(ctx, `DELETE FROM alliance_applications WHERE id=$1`, applicationID)
		return err
	})
	if err != nil {
		return err
	}
	s.notifyAlliance(ctx, applicantID,
		"Заявка одобрена",
		fmt.Sprintf("Ваша заявка на вступление в альянс [%s] одобрена.", allianceTag))
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
		_, err = tx.Exec(ctx, `DELETE FROM alliance_applications WHERE id=$1`, applicationID)
		return err
	})
	if err != nil {
		return err
	}
	s.notifyAlliance(ctx, applicantID,
		"Заявка отклонена",
		fmt.Sprintf("Ваша заявка на вступление в альянс [%s] отклонена.", allianceTag))
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
	ErrInvalidRelation  = errors.New("alliance: relation must be 'nap', 'war', or 'ally'")
	ErrTargetNotFound   = errors.New("alliance: target alliance not found")
	ErrRelationSelf     = errors.New("alliance: cannot set relation with own alliance")
	ErrRelationPending  = errors.New("alliance: relation proposal already pending")
)

// ProposeRelation предлагает отношение от allianceID к targetID.
// WAR — активно сразу (односторонне). NAP/ALLY — ждёт подтверждения от target.
// Relation="none" — удаляет любые записи в обе стороны.
func (s *Service) ProposeRelation(ctx context.Context, userID, allianceID, targetID, relation string) error {
	if allianceID == targetID {
		return ErrRelationSelf
	}
	if relation != "nap" && relation != "war" && relation != "ally" && relation != "none" {
		return ErrInvalidRelation
	}

	var ownerID string
	err := s.db.Pool().QueryRow(ctx,
		`SELECT owner_id FROM alliances WHERE id=$1`, allianceID).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("propose relation: read alliance: %w", err)
	}
	if ownerID != userID {
		return ErrNotOwner
	}

	var targetExists bool
	if err := s.db.Pool().QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM alliances WHERE id=$1)`, targetID).Scan(&targetExists); err != nil {
		return fmt.Errorf("propose relation: check target: %w", err)
	}
	if !targetExists {
		return ErrTargetNotFound
	}

	if relation == "none" {
		// Удаляем записи в обе стороны.
		if _, err := s.db.Pool().Exec(ctx, `
			DELETE FROM alliance_relationships
			WHERE (alliance_id=$1 AND target_alliance_id=$2)
			   OR (alliance_id=$2 AND target_alliance_id=$1)
		`, allianceID, targetID); err != nil {
			return fmt.Errorf("propose relation: delete: %w", err)
		}
		return nil
	}

	// WAR немедленно активен; NAP/ALLY — pending.
	status := "pending"
	if relation == "war" {
		status = "active"
	}

	if _, err := s.db.Pool().Exec(ctx, `
		INSERT INTO alliance_relationships (alliance_id, target_alliance_id, relation, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (alliance_id, target_alliance_id)
		DO UPDATE SET relation=$3, status=$4, set_at=now()
	`, allianceID, targetID, relation, status); err != nil {
		return fmt.Errorf("propose relation: upsert: %w", err)
	}
	return nil
}

// AcceptRelation подтверждает входящее предложение NAP/ALLY.
// Вызывается owner'ом targetID. После accept — обе записи становятся active.
func (s *Service) AcceptRelation(ctx context.Context, userID, myAllianceID, initiatorAllianceID string) error {
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Проверяем что userID — owner myAllianceID.
		var ownerID string
		if err := tx.QueryRow(ctx,
			`SELECT owner_id FROM alliances WHERE id=$1`, myAllianceID).Scan(&ownerID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("accept relation: %w", err)
		}
		if ownerID != userID {
			return ErrNotOwner
		}

		// Читаем pending предложение от initiatorAllianceID → myAllianceID.
		var relation string
		var status string
		err := tx.QueryRow(ctx, `
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
		return nil
	})
}

// RejectRelation отклоняет входящее предложение NAP/ALLY — удаляет pending запись.
func (s *Service) RejectRelation(ctx context.Context, userID, myAllianceID, initiatorAllianceID string) error {
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var ownerID string
		if err := tx.QueryRow(ctx,
			`SELECT owner_id FROM alliances WHERE id=$1`, myAllianceID).Scan(&ownerID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("reject relation: %w", err)
		}
		if ownerID != userID {
			return ErrNotOwner
		}

		if _, err := tx.Exec(ctx, `
			DELETE FROM alliance_relationships
			WHERE alliance_id=$1 AND target_alliance_id=$2 AND status='pending'
		`, initiatorAllianceID, myAllianceID); err != nil {
			return fmt.Errorf("reject relation: delete: %w", err)
		}
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
