// Package automsg — шаблонизированные системные сообщения.
//
// Send(userID, key, vars) идемпотентен: если (user, key) уже отправлено —
// no-op. Это важно для welcome-сообщения: повторный вызов (например
// при resend после сбоя) не плодит дубликаты.
//
// Шаблонизация — примитивная: strings.ReplaceAll для каждой пары
// из vars. Синтаксис {{name}}. Нет условных блоков, форматирования
// или i18n-ветвления (legacy AutoMsg.class.php имел 1228 LOC,
// нам такой объём не нужен).
package automsg

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/oxsar/nova/backend/internal/repo"
	"github.com/oxsar/nova/backend/pkg/ids"
)

type Service struct {
	db repo.Exec
}

func NewService(db repo.Exec) *Service { return &Service{db: db} }

var ErrDefNotFound = errors.New("automsg: def not found")

// Send отправляет automsg по key, подставляя vars. Идемпотентен по
// (userID, key): повторный вызов ничего не делает.
//
// tx опционален: если nil, используется пул. В hot-path (регистрация,
// критические триггеры) передаётся существующий tx, чтобы отправка
// была атомарна с основной операцией.
func (s *Service) Send(ctx context.Context, tx pgx.Tx, userID, key string, vars map[string]string) error {
	exec := txOrPool(s, tx)

	// Def.
	var title, body string
	var folder int
	if err := exec.QueryRow(ctx, `
		SELECT title, body_template, folder FROM automsg_defs WHERE key = $1
	`, key).Scan(&title, &body, &folder); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrDefNotFound
		}
		return fmt.Errorf("read def: %w", err)
	}

	// Подстановка vars в title и body.
	for k, v := range vars {
		placeholder := "{{" + k + "}}"
		title = strings.ReplaceAll(title, placeholder, v)
		body = strings.ReplaceAll(body, placeholder, v)
	}

	// Фиксируем отправку: INSERT ON CONFLICT DO NOTHING. Если
	// RowsAffected()==0 — уже отправляли, выходим.
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

	// Message в inbox.
	if _, err := exec.Exec(ctx, `
		INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
		VALUES ($1, $2, NULL, $3, $4, $5)
	`, ids.New(), userID, folder, title, body); err != nil {
		return fmt.Errorf("insert message: %w", err)
	}
	return nil
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

// SendInactivityReminders отправляет INACTIVITY_REMINDER всем игрокам,
// которые не заходили более inactiveDays дней. Использует week-суффикс
// в ключе idempotency (INACTIVITY_REMINDER_<year>W<week>), чтобы при
// ежедневном вызове письмо уходило не чаще раза в неделю.
// Возвращает кол-во отправленных писем.
func (s *Service) SendInactivityReminders(ctx context.Context, inactiveDays int, weekSuffix string) (int, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT id, username FROM users
		WHERE last_seen_at < now() - ($1 * interval '1 day')
	`, inactiveDays)
	if err != nil {
		return 0, fmt.Errorf("inactivity: query: %w", err)
	}
	defer rows.Close()

	sentKey := "INACTIVITY_REMINDER_" + weekSuffix
	var sent int
	for rows.Next() {
		var uid, username string
		if err := rows.Scan(&uid, &username); err != nil {
			return sent, err
		}
		if err := s.sendWithKey(ctx, uid, sentKey, "INACTIVITY_REMINDER",
			map[string]string{"username": username}); err != nil {
			if errors.Is(err, ErrDefNotFound) {
				return 0, nil // шаблон не добавлен в БД — пропускаем всё
			}
			return sent, fmt.Errorf("inactivity send %s: %w", uid, err)
		}
		sent++
	}
	return sent, rows.Err()
}

// sendWithKey — как Send, но idempotency-ключ задаётся явно (не key шаблона).
func (s *Service) sendWithKey(ctx context.Context, userID, sentKey, defKey string, vars map[string]string) error {
	exec := txOrPool(s, nil)

	var title, body string
	var folder int
	if err := exec.QueryRow(ctx, `
		SELECT title, body_template, folder FROM automsg_defs WHERE key = $1
	`, defKey).Scan(&title, &body, &folder); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrDefNotFound
		}
		return fmt.Errorf("read def: %w", err)
	}
	for k, v := range vars {
		placeholder := "{{" + k + "}}"
		title = strings.ReplaceAll(title, placeholder, v)
		body = strings.ReplaceAll(body, placeholder, v)
	}
	tag, err := exec.Exec(ctx, `
		INSERT INTO automsg_sent (user_id, key)
		VALUES ($1, $2)
		ON CONFLICT (user_id, key) DO NOTHING
	`, userID, sentKey)
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

// --- tx/pool switch ---
//
// Повторяется из achievement/service.go (такой же execer). Могли бы
// вынести в shared helper, но это две короткие обёртки — пока
// дублирование дешевле общего пакета.

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
