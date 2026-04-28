// Package client содержит HTTP-клиент к billing-service для списания (Spend)
// и возврата (Refund) оксаров с пользовательских кошельков.
//
// # Архитектура
//
// game-nova не управляет кошельками (создание/чтение балансов — это домен
// billing). game-nova только списывает или возвращает оксары через
// синхронный HTTP POST к billing-сервису. JWT юзера forward'ится через
// Authorization-header. Idempotency-Key обязателен — без него операция
// небезопасна на retry.
//
// # Sentinel-ошибки
//
//   - ErrInsufficientOxsar (HTTP 402)
//   - ErrIdempotencyConflict (HTTP 409)
//   - ErrFrozenWallet (HTTP 423)
//   - ErrBillingUnavailable (timeout/network/5xx после retry)
//
// # Метрики (R8)
//
//   - oxsar_billing_client_spend_total{status} — counter
//   - oxsar_billing_client_duration_seconds{operation} — histogram
//
// Регистрируются через pkg/metrics.RegisterBilling().
//
// # Retry
//
// Client делает один retry на транзиентных ошибках (timeout, connection refused,
// 5xx). Backoff линейный (200ms). Этого достаточно для покрытия типичных
// кратковременных сбоев billing-service. Полноценный exponential backoff
// — отложен до момента, когда метрики покажут необходимость
// (см. simplifications.md).
package client
