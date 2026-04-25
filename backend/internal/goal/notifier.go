package goal

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/pkg/ids"
)

// Notifier — уведомление пользователя о completion goal.
//
// MVP: запись в inbox (как текущий achievement.UnlockIfNew). Дополнительно
// frontend через `seen_at IS NULL` решит, показать ли toast при
// следующем заходе на /api/goals.
type Notifier interface {
	OnCompleted(ctx context.Context, tx pgx.Tx, userID string, def GoalDef) error
}

// InboxNotifier пишет message в users inbox при completion.
type InboxNotifier struct{}

func NewInboxNotifier() *InboxNotifier { return &InboxNotifier{} }

// OnCompleted — INSERT в messages (folder=2 — system / achievements).
// Идемпотентность гарантирует Engine: вызов делается ровно один раз
// при переходе progress → completed.
func (n *InboxNotifier) OnCompleted(ctx context.Context, tx pgx.Tx, userID string, def GoalDef) error {
	subject := fmt.Sprintf("Цель: %s", def.Title)
	body := def.Description
	if body == "" {
		body = fmt.Sprintf("Вы завершили цель «%s».", def.Title)
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
		VALUES ($1, $2, NULL, 2, $3, $4)
	`, ids.New(), userID, subject, body)
	if err != nil {
		return fmt.Errorf("notifier inbox: %w", err)
	}
	return nil
}

// NoopNotifier — no-op реализация (для тестов и сценариев без inbox).
type NoopNotifier struct{}

func (NoopNotifier) OnCompleted(ctx context.Context, tx pgx.Tx, userID string, def GoalDef) error {
	return nil
}
