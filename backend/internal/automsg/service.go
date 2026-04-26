// Package automsg — шаблонизированные системные сообщения.
//
// Send(userID, key, vars) идемпотентен: если (user, key) уже отправлено —
// no-op. Это важно для welcome-сообщения: повторный вызов (например
// при resend после сбоя) не плодит дубликаты.
//
// Тексты шаблонов хранятся в configs/i18n/<lang>.yml группа autoMessages.
// Пары ключей: "<name>.title" и "<name>.body".
// Язык получателя читается из users.language при каждом Send.
//
// Папки inbox (folder): 2 — системные сообщения (legacy consts).
package automsg

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/oxsar/nova/backend/internal/i18n"
	"github.com/oxsar/nova/backend/internal/repo"
	"github.com/oxsar/nova/backend/pkg/ids"
)

// KnownKeys — допустимые ключи automsg. Проверяется в Send (defensive).
var KnownKeys = []string{
	"welcome",
	"starterGuide",
	"firstAttackReceived",
	"inactivityReminder",
}

// Folder — inbox-папка для автомесседжей (legacy consts.php).
const Folder = 2

type Service struct {
	db     repo.Exec
	bundle *i18n.Bundle
}

func NewService(db repo.Exec) *Service { return &Service{db: db} }

// WithBundle подключает i18n-бандл для получения текстов шаблонов.
// Если не вызван — Send возвращает ErrNoBundle.
func (s *Service) WithBundle(b *i18n.Bundle) *Service {
	s.bundle = b
	return s
}

var (
	ErrDefNotFound = errors.New("automsg: def not found")
	ErrNoBundle    = errors.New("automsg: i18n bundle not configured")
)

// Send отправляет automsg по key, подставляя vars. Идемпотентен по
// (userID, key): повторный вызов ничего не делает.
//
// tx опционален: если nil, используется пул. В hot-path (регистрация,
// критические триггеры) передаётся существующий tx, чтобы отправка
// была атомарна с основной операцией.
func (s *Service) Send(ctx context.Context, tx pgx.Tx, userID, key string, vars map[string]string) error {
	if s.bundle == nil {
		return ErrNoBundle
	}

	// Язык получателя.
	lang := s.userLang(ctx, userID)

	title := s.bundle.Tr(lang, "autoMessages", key+".title", vars)
	body := s.bundle.Tr(lang, "autoMessages", key+".body", vars)

	if title == "[autoMessages."+key+".title]" {
		return fmt.Errorf("%w: %s", ErrDefNotFound, key)
	}

	return s.insertMsg(ctx, tx, userID, key, Folder, title, body)
}

// SendDirect вставляет сообщение без шаблона и без идемпотентности.
// Используется для многоразовых событий (сканы, транзакции, события
// альянса). folder — номер папки из legacy consts.php.
// tx опционален (при nil используется пул).
func (s *Service) SendDirect(ctx context.Context, tx pgx.Tx, userID string, folder int, title, body string) error {
	exec := txOrPool(s, tx)
	if _, err := exec.Exec(ctx, `
		INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
		VALUES ($1, $2, NULL, $3, $4, $5)
	`, ids.New(), userID, folder, title, body); err != nil {
		return fmt.Errorf("automsg direct: %w", err)
	}
	return nil
}

// SendInactivityReminders отправляет inactivityReminder всем игрокам,
// которые не заходили более inactiveDays дней. Использует week-суффикс
// в ключе idempotency, чтобы при ежедневном вызове письмо уходило
// не чаще раза в неделю. Возвращает кол-во отправленных писем.
func (s *Service) SendInactivityReminders(ctx context.Context, inactiveDays int, weekSuffix string) (int, error) {
	if s.bundle == nil {
		return 0, ErrNoBundle
	}

	rows, err := s.db.Pool().Query(ctx, `
		SELECT id, username, language FROM users
		WHERE last_seen_at < now() - ($1 * interval '1 day')
	`, inactiveDays)
	if err != nil {
		return 0, fmt.Errorf("inactivity: query: %w", err)
	}
	defer rows.Close()

	sentKey := "inactivityReminder_" + weekSuffix
	var sent int
	for rows.Next() {
		var uid, username, lang string
		if err := rows.Scan(&uid, &username, &lang); err != nil {
			return sent, err
		}
		l := i18n.Lang(lang)
		vars := map[string]string{"username": username}
		title := s.bundle.Tr(l, "autoMessages", "inactivityReminder.title", vars)
		body := s.bundle.Tr(l, "autoMessages", "inactivityReminder.body", vars)
		if err := s.insertMsg(ctx, nil, uid, sentKey, Folder, title, body); err != nil {
			return sent, fmt.Errorf("inactivity send %s: %w", uid, err)
		}
		sent++
	}
	return sent, rows.Err()
}

// insertMsg — идемпотентная вставка сообщения через automsg_sent.
func (s *Service) insertMsg(ctx context.Context, tx pgx.Tx, userID, key string, folder int, title, body string) error {
	exec := txOrPool(s, tx)

	tag, err := exec.Exec(ctx, `
		INSERT INTO automsg_sent (user_id, key)
		VALUES ($1, $2)
		ON CONFLICT (user_id, key) DO NOTHING
	`, userID, key)
	if err != nil {
		return fmt.Errorf("mark sent: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil
	}

	if _, err := exec.Exec(ctx, `
		INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
		VALUES ($1, $2, NULL, $3, $4, $5)
	`, ids.New(), userID, folder, title, body); err != nil {
		return fmt.Errorf("insert message: %w", err)
	}
	return nil
}

// userLang читает язык пользователя из БД. При ошибке — fallback ru.
func (s *Service) userLang(ctx context.Context, userID string) i18n.Lang {
	var lang string
	_ = s.db.Pool().QueryRow(ctx,
		`SELECT language FROM users WHERE id = $1`, userID,
	).Scan(&lang)
	if lang == "" {
		return i18n.LangRu
	}
	return i18n.Lang(lang)
}

// --- tx/pool switch ---

type execer interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func txOrPool(s *Service, tx pgx.Tx) execer {
	if tx != nil {
		return tx
	}
	return poolAdapter{pool: s.db.Pool()}
}

type poolAdapter struct {
	pool poolIface
}
type poolIface interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func (p poolAdapter) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return p.pool.Exec(ctx, sql, args...)
}
func (p poolAdapter) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return p.pool.QueryRow(ctx, sql, args...)
}
