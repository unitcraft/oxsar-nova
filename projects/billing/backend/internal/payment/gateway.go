// Package payment — платёжные шлюзы для billing-service.
//
// План 38 Ф.3-4. Поддерживаются провайдеры:
//   - mock (dev/test, без подписей и внешних HTTP)
//   - robokassa (готов к проду, требует MerchantID + SecretKey)
//   - enot.io (резервный)
//
// Выбор шлюза по env PAYMENT_PROVIDER (см. cmd/server/main.go).
package payment

import (
	"context"
	"errors"
	"net/http"
)

// Errors.
var (
	ErrWebhookInvalid     = errors.New("payment: invalid webhook payload")
	ErrSignatureMismatch  = errors.New("payment: signature mismatch")
	ErrTimestampOld       = errors.New("payment: timestamp too old (replay protection)")
)

// Gateway — общий интерфейс платёжного шлюза.
//
// BuildPayURL — сформировать ссылку для редиректа игрока на страницу оплаты.
// VerifyWebhook — распарсить и проверить webhook от шлюза. Возвращает
// (orderID, amountKop, error). signature/timestamp проверяются внутри.
// SuccessResponse — записать в w то, что шлюз ожидает увидеть в ответ
// (для Robokassa это "OK<order_id>", для других — пусто/204).
type Gateway interface {
	Name() string
	BuildPayURL(ctx context.Context, orderID, userID string, amountKop int64, returnURL string) (string, error)
	VerifyWebhook(r *http.Request, body []byte) (orderID string, amountKop int64, err error)
	SuccessResponse(w http.ResponseWriter, orderID string)
}

// Package — пакет кредитов (offer).
type Package struct {
	ID         string `json:"id" yaml:"id"`             // 'pack_500'
	Title      string `json:"title" yaml:"title"`       // '500 кредитов'
	AmountKop  int64  `json:"amount_kop" yaml:"amount_kop"` // 50000 копеек = 500 руб
	Credits    int64  `json:"credits" yaml:"credits"`   // 50000 (внутренние единицы OXC)
	Bonus      int64  `json:"bonus,omitempty" yaml:"bonus"`
	IsBest     bool   `json:"is_best,omitempty" yaml:"is_best"`
}
