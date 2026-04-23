package payment

import (
	"context"
	"errors"
	"net/http"
)

// Sentinel errors.
var (
	ErrPackageNotFound = errors.New("payment: unknown package key")
	ErrOrderNotFound   = errors.New("payment: order not found")
	ErrAlreadyPaid     = errors.New("payment: order already paid")
	ErrWebhookInvalid  = errors.New("payment: webhook signature invalid")
	ErrGatewayDisabled = errors.New("payment: no payment provider configured")
)

// Gateway — абстракция платёжного шлюза.
// Реализации: RobokassaGateway, EnotGateway.
type Gateway interface {
	// BuildPayURL формирует ссылку для перехода на страницу оплаты.
	BuildPayURL(ctx context.Context, orderID, description string, amountKop int, returnURL string) (string, error)

	// VerifyWebhook разбирает и верифицирует входящий callback.
	// Возвращает orderID заказа и уплаченную сумму в копейках.
	VerifyWebhook(r *http.Request) (orderID string, amountKop int, err error)

	// SuccessResponse пишет ответ, ожидаемый шлюзом при успешной обработке.
	SuccessResponse(w http.ResponseWriter, orderID string)
}
