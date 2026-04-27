// Package payment — платёжные шлюзы для billing-service.
//
// План 38 Ф.3-4. Поддерживаются провайдеры:
//   - mock — dev/test, HMAC-SHA256 подпись webhook
//   - yookassa — основной (план 42), API + IP allowlist + re-fetch
//
// Robokassa и Enot.io остаются опциональными (через тот же Gateway-интерфейс),
// но не реализованы — добавятся при появлении тестового аккаунта.
//
// Выбор шлюза по env PAYMENT_PROVIDER (см. cmd/server/main.go).
package payment

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// Errors.
var (
	ErrWebhookInvalid     = errors.New("payment: invalid webhook payload")
	ErrSignatureMismatch  = errors.New("payment: signature mismatch")
	ErrTimestampOld       = errors.New("payment: timestamp too old (replay protection)")
	ErrUnknownProvider    = errors.New("payment: unknown provider")
)

// FactoryConfig — параметры для NewGateway-factory.
// Не все поля нужны каждому провайдеру (mock использует BaseURL+Secret,
// YooKassa — ShopID+SecretKey+APIURL и т.д.).
type FactoryConfig struct {
	// общие
	ReturnURL string

	// mock
	MockBaseURL string
	MockSecret  string

	// yookassa
	YooKassaShopID         string
	YooKassaSecretKey      string
	YooKassaAPIURL         string // пусто → prod https://api.yookassa.ru/v3
	// YooKassaDisableIPAllowlist — отключить проверку IP webhook'а.
	// Только для dev/staging с локальным yookassa-mock — webhook прилетит
	// с docker-internal IP, не из YooKassa-подсетей. На проде ОБЯЗАТЕЛЬНО
	// false.
	YooKassaDisableIPAllowlist bool
}

// NewGateway возвращает Gateway по имени провайдера.
// План 42: yookassa добавлен. Robokassa/Enot — отложены до тестового аккаунта.
func NewGateway(provider string, cfg FactoryConfig) (Gateway, error) {
	switch provider {
	case "mock":
		return NewMockGateway(cfg.MockBaseURL, cfg.MockSecret), nil
	case "yookassa":
		if cfg.YooKassaShopID == "" || cfg.YooKassaSecretKey == "" {
			return nil, errors.New("yookassa: YOOKASSA_SHOP_ID and YOOKASSA_SECRET_KEY required")
		}
		gw := NewYooKassaGateway(
			cfg.YooKassaShopID,
			cfg.YooKassaSecretKey,
			cfg.YooKassaAPIURL,
			cfg.ReturnURL,
		)
		if cfg.YooKassaDisableIPAllowlist {
			// dev/staging-режим: webhook прилетит с docker-internal IP.
			gw.SetTrustedNetworks(nil)
		}
		return gw, nil
	default:
		return nil, fmt.Errorf("%w: %q (supported: mock, yookassa)", ErrUnknownProvider, provider)
	}
}

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
