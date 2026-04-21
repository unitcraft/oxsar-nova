// Package message — личные сообщения (inbox).
//
// Использует таблицы messages (legacy-совместимая, миграция 0005) и
// битовое расширение battle_report_id (миграция 0009).
package message

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/repo"
)

type Service struct {
	db repo.Exec
}

func NewService(db repo.Exec) *Service { return &Service{db: db} }

// Message — строка из messages для UI. body — уже готовая строка,
// battle_report_id/espionage_report_id nullable (nil — обычное
// сообщение).
type Message struct {
	ID                string     `json:"id"`
	FromUserID        *string    `json:"from_user_id,omitempty"`
	FromUsername      string     `json:"from_username,omitempty"`
	Subject           string     `json:"subject"`
	Body              string     `json:"body"`
	Folder            int        `json:"folder"`
	CreatedAt         time.Time  `json:"created_at"`
	ReadAt            *time.Time `json:"read_at,omitempty"`
	BattleReportID     *string    `json:"battle_report_id,omitempty"`
	EspionageReportID  *string    `json:"espionage_report_id,omitempty"`
	ExpeditionReportID *string    `json:"expedition_report_id,omitempty"`
}

// Inbox возвращает последние N сообщений, новые сверху. Без пагинации
// — для M4.4b достаточно (обычно у игрока максимум несколько десятков
// отчётов в первую неделю; реальная пагинация придёт с alliance-chat).
func (s *Service) Inbox(ctx context.Context, userID string, limit int) ([]Message, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	rows, err := s.db.Pool().Query(ctx, `
		SELECT m.id, m.from_user_id, COALESCE(u.username, ''),
		       m.subject, m.body, m.folder, m.created_at, m.read_at,
		       m.battle_report_id, m.espionage_report_id, m.expedition_report_id
		FROM messages m
		LEFT JOIN users u ON u.id = m.from_user_id
		WHERE m.to_user_id = $1
		ORDER BY m.created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("inbox query: %w", err)
	}
	defer rows.Close()

	var out []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.FromUserID, &m.FromUsername,
			&m.Subject, &m.Body, &m.Folder, &m.CreatedAt, &m.ReadAt,
			&m.BattleReportID, &m.EspionageReportID, &m.ExpeditionReportID); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// UnreadCount — сколько у пользователя непрочитанных сообщений.
// Используется для бейджа в header'е UI.
func (s *Service) UnreadCount(ctx context.Context, userID string) (int, error) {
	var n int
	err := s.db.Pool().QueryRow(ctx,
		`SELECT COUNT(*) FROM messages WHERE to_user_id = $1 AND read_at IS NULL`,
		userID).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("unread count: %w", err)
	}
	return n, nil
}

// MarkRead ставит read_at=now() если ещё не прочитано. Идемпотентно.
// Ошибка «чужое сообщение» возвращается как ErrNotOwned.
var (
	ErrMessageNotFound = errors.New("message: not found")
	ErrNotOwned        = errors.New("message: not owned by user")
)

func (s *Service) MarkRead(ctx context.Context, userID, messageID string) error {
	tag, err := s.db.Pool().Exec(ctx, `
		UPDATE messages SET read_at = now()
		WHERE id = $1 AND to_user_id = $2 AND read_at IS NULL
	`, messageID, userID)
	if err != nil {
		return fmt.Errorf("mark read: %w", err)
	}
	if tag.RowsAffected() > 0 {
		return nil
	}
	// 0 rows: либо сообщение не наше, либо уже прочитано, либо не
	// существует. Для UI разницы нет, но чтобы 404/403 отдавать
	// честно — пере-проверим.
	var ownerID string
	err = s.db.Pool().QueryRow(ctx,
		`SELECT to_user_id FROM messages WHERE id = $1`, messageID).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrMessageNotFound
		}
		return err
	}
	if ownerID != userID {
		return ErrNotOwned
	}
	// Уже прочитано — не ошибка.
	return nil
}

// Ошибки compose/delete.
var (
	ErrRecipientNotFound = errors.New("message: recipient not found")
	ErrSelfMessage       = errors.New("message: cannot send to yourself")
)

// FolderInbox — папка личных сообщений (MSG_FOLDER_INBOX из legacy consts.php).
const FolderInbox = 1

// Compose отправляет личное сообщение от fromUserID к получателю по username.
// Сообщение кладётся в папку FolderInbox получателя. Лимит subject 200 символов,
// body 10 000 символов.
func (s *Service) Compose(ctx context.Context, fromUserID, toUsername, subject, body string) error {
	if fromUserID == "" || toUsername == "" {
		return ErrRecipientNotFound
	}
	subject = truncate(subject, 200)
	body = truncate(body, 10000)

	var toUserID string
	err := s.db.Pool().QueryRow(ctx,
		`SELECT id FROM users WHERE username = $1`, toUsername).Scan(&toUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRecipientNotFound
		}
		return fmt.Errorf("compose lookup: %w", err)
	}
	if toUserID == fromUserID {
		return ErrSelfMessage
	}

	_, err = s.db.Pool().Exec(ctx, `
		INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5)
	`, toUserID, fromUserID, FolderInbox, subject, body)
	if err != nil {
		return fmt.Errorf("compose insert: %w", err)
	}
	return nil
}

// Delete удаляет сообщение. Только владелец (to_user_id) может удалить своё.
func (s *Service) Delete(ctx context.Context, userID, messageID string) error {
	tag, err := s.db.Pool().Exec(ctx,
		`DELETE FROM messages WHERE id = $1 AND to_user_id = $2`,
		messageID, userID)
	if err != nil {
		return fmt.Errorf("delete message: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrMessageNotFound
	}
	return nil
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) > max {
		return string(runes[:max])
	}
	return s
}

// ExpeditionReport — полный отчёт экспедиции из expedition_reports.
type ExpeditionReport struct {
	ID      string          `json:"id"`
	UserID  *string         `json:"user_id,omitempty"`
	FleetID *string         `json:"fleet_id,omitempty"`
	Outcome string          `json:"outcome"`
	At      time.Time       `json:"at"`
	Report  json.RawMessage `json:"report"`
}

func (s *Service) GetExpeditionReport(ctx context.Context, userID, reportID string) (ExpeditionReport, error) {
	var r ExpeditionReport
	err := s.db.Pool().QueryRow(ctx, `
		SELECT id, user_id, fleet_id, outcome, at, report
		FROM expedition_reports WHERE id = $1
	`, reportID).Scan(&r.ID, &r.UserID, &r.FleetID, &r.Outcome, &r.At, &r.Report)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ExpeditionReport{}, ErrMessageNotFound
		}
		return ExpeditionReport{}, fmt.Errorf("get expedition report: %w", err)
	}
	if r.UserID == nil || *r.UserID != userID {
		return ExpeditionReport{}, ErrNotOwned
	}
	return r, nil
}

// EspionageReport — полный шпионский отчёт из espionage_reports.
type EspionageReport struct {
	ID           string          `json:"id"`
	SpyUserID    *string         `json:"spy_user_id,omitempty"`
	TargetUserID *string         `json:"target_user_id,omitempty"`
	PlanetID     *string         `json:"planet_id,omitempty"`
	Ratio        int             `json:"ratio"`
	Probes       int             `json:"probes"`
	At           time.Time       `json:"at"`
	Report       json.RawMessage `json:"report"`
}

// GetEspionageReport — доступ открыт шпиону и цели.
func (s *Service) GetEspionageReport(ctx context.Context, userID, reportID string) (EspionageReport, error) {
	var r EspionageReport
	err := s.db.Pool().QueryRow(ctx, `
		SELECT id, spy_user_id, target_user_id, planet_id,
		       ratio, probes, at, report
		FROM espionage_reports
		WHERE id = $1
	`, reportID).Scan(&r.ID, &r.SpyUserID, &r.TargetUserID, &r.PlanetID,
		&r.Ratio, &r.Probes, &r.At, &r.Report)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return EspionageReport{}, ErrMessageNotFound
		}
		return EspionageReport{}, fmt.Errorf("get espionage report: %w", err)
	}
	isOwn := (r.SpyUserID != nil && *r.SpyUserID == userID) ||
		(r.TargetUserID != nil && *r.TargetUserID == userID)
	if !isOwn {
		return EspionageReport{}, ErrNotOwned
	}
	return r, nil
}

// BattleReport — полный отчёт боя + метаданные из battle_reports.
type BattleReport struct {
	ID               string          `json:"id"`
	AttackerUserID   *string         `json:"attacker_user_id,omitempty"`
	AttackerUsername string          `json:"attacker_username,omitempty"`
	DefenderUserID   *string         `json:"defender_user_id,omitempty"`
	DefenderUsername string          `json:"defender_username,omitempty"`
	PlanetID         *string         `json:"planet_id,omitempty"`
	Seed             int64           `json:"seed"`
	Winner           string          `json:"winner"`
	Rounds           int             `json:"rounds"`
	DebrisMetal      int64           `json:"debris_metal"`
	DebrisSilicon    int64           `json:"debris_silicon"`
	LootMetal        int64           `json:"loot_metal"`
	LootSilicon      int64           `json:"loot_silicon"`
	LootHydrogen     int64           `json:"loot_hydrogen"`
	At               time.Time       `json:"at"`
	Report           json.RawMessage `json:"report"`
}

// GetBattleReport читает report по id, доступ открыт только
// attacker/defender (в ТЗ §12.1 предусмотрен «шейринг», но это
// M5-фича; пока — двухсторонний read).
func (s *Service) GetBattleReport(ctx context.Context, userID, reportID string) (BattleReport, error) {
	var r BattleReport
	err := s.db.Pool().QueryRow(ctx, `
		SELECT br.id, br.attacker_user_id, COALESCE(ua.username,''), br.defender_user_id, COALESCE(ud.username,''),
		       br.planet_id,
		       br.seed, br.winner, br.rounds,
		       br.debris_metal, br.debris_silicon,
		       br.loot_metal, br.loot_silicon, br.loot_hydrogen,
		       br.at, br.report
		FROM battle_reports br
		LEFT JOIN users ua ON ua.id = br.attacker_user_id
		LEFT JOIN users ud ON ud.id = br.defender_user_id
		WHERE br.id = $1
	`, reportID).Scan(&r.ID, &r.AttackerUserID, &r.AttackerUsername, &r.DefenderUserID, &r.DefenderUsername,
		&r.PlanetID,
		&r.Seed, &r.Winner, &r.Rounds,
		&r.DebrisMetal, &r.DebrisSilicon,
		&r.LootMetal, &r.LootSilicon, &r.LootHydrogen,
		&r.At, &r.Report)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return BattleReport{}, ErrMessageNotFound
		}
		return BattleReport{}, fmt.Errorf("get report: %w", err)
	}
	// Проверка доступа.
	isOwn := (r.AttackerUserID != nil && *r.AttackerUserID == userID) ||
		(r.DefenderUserID != nil && *r.DefenderUserID == userID)
	if !isOwn {
		return BattleReport{}, ErrNotOwned
	}
	return r, nil
}
