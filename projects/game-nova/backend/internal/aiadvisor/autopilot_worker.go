package aiadvisor

import (
	"context"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/event"
)

// AutopilotWorkerHandler оборачивает AutopilotService.Compute в сигнатуру
// event.Handler, которую регистрирует Worker.Register.
type AutopilotWorkerHandler struct {
	Svc *AutopilotService
}

// Handle — реализация event.Handler.
func (h *AutopilotWorkerHandler) Handle(ctx context.Context, tx pgx.Tx, e event.Event) error {
	return h.Svc.Compute(ctx, tx, e)
}
