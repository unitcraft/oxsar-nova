// Package client — HTTP-клиент game-nova к billing-service для списания
// и возврата оксаров (hard-валюта по ADR-0009).
//
// План 77: разблокирует премиум-фичи ремастера (план 65 Ф.6 KindTeleportPlanet,
// план 66 Ф.5 платный выкуп удержания инопланетянами и т.д.).
//
// Образец расширен из projects/portal/backend/internal/portalsvc/credits.go.
package client

import "errors"

// ErrInsufficientOxsar — у юзера не хватает оксаров (HTTP 402 Payment Required
// от billing-service). Доменная ошибка, проксируется клиенту как 402.
var ErrInsufficientOxsar = errors.New("billing: insufficient oxsar")

// ErrBillingUnavailable — billing-service недоступен (timeout, refused
// connection, 5xx после retry). Транзиентная — handler может разрешить retry
// в течение бизнес-окна (например, alien-buyout — 5 мин).
var ErrBillingUnavailable = errors.New("billing: service unavailable")

// ErrIdempotencyConflict — повторный запрос с тем же Idempotency-Key, но
// другим body (HTTP 409 Conflict). Это клиент-баг: разные параметры под одним
// ключом. Игроку показать «технический сбой, повторите попытку».
var ErrIdempotencyConflict = errors.New("billing: idempotency conflict")

// ErrFrozenWallet — кошелёк заморожен (HTTP 423 Locked). Бывает при
// reconcile-расхождениях; решается админом, не пользователем.
var ErrFrozenWallet = errors.New("billing: wallet frozen")

// ErrNotConfigured — BILLING_URL не задан. game-nova-сервер должен fail-fast
// до запуска handler'ов, требующих списания, но клиент защищается на случай
// конфигурационной ошибки.
var ErrNotConfigured = errors.New("billing: client not configured")
