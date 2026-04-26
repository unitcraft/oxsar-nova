package goal

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/i18n"
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
type InboxNotifier struct {
	bundle *i18n.Bundle
}

func NewInboxNotifier() *InboxNotifier { return &InboxNotifier{} }

func (n *InboxNotifier) WithBundle(b *i18n.Bundle) *InboxNotifier {
	n.bundle = b
	return n
}

func (n *InboxNotifier) tr(group, key string, vars map[string]string) string {
	if n.bundle == nil {
		return "[" + group + "." + key + "]"
	}
	return n.bundle.Tr(i18n.LangRu, group, key, vars)
}

// OnCompleted — INSERT в messages (folder=2 — system / achievements).
// Идемпотентность гарантирует Engine: вызов делается ровно один раз
// при переходе progress → completed.
func (n *InboxNotifier) OnCompleted(ctx context.Context, tx pgx.Tx, userID string, def GoalDef) error {
	vars := map[string]string{"title": def.Title}
	subject := n.tr("goal", "subject", vars)
	body := def.Description
	if body == "" {
		body = n.tr("goal", "body", vars)
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
