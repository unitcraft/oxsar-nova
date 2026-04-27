# admin-bff

Backend-for-Frontend для `admin.oxsar-nova.ru` (план 53).

Хранит сессии админов в Redis, проксирует `/api/*` на backend-сервисы
(identity, billing, game-nova), добавляя `Authorization: Bearer <JWT>`.
Браузер видит только opaque `admin_session` cookie + `admin_csrf` для
double-submit CSRF.

## Архитектура

См. [docs/plans/53-admin-frontend.md](../../docs/plans/53-admin-frontend.md)
§Auth-flow (BFF). Кратко:

- Login: frontend → BFF → identity. BFF создаёт session в Redis,
  отдаёт claims summary в JSON + ставит cookies.
- Каждый защищённый запрос: BFF читает `admin_session` cookie, достаёт
  JWT из Redis, инжектит `Authorization`, проксирует upstream.
- Auto-refresh: при `time.Until(exp) < BFF_REFRESH_LEAD_TIME` BFF
  обменивает refresh-token на новую пару (прозрачно).
- Idle timeout: sliding TTL `BFF_IDLE_TIMEOUT` (по умолч. 30 мин).
- Logout: BFF revoke в identity + удаляет session в Redis + чистит
  cookies.

## Конфиг (ENV)

| Var | Default | Description |
|---|---|---|
| `BFF_LISTEN_ADDR` | `:9200` | HTTP listen address |
| `IDENTITY_URL` | `http://localhost:9001` | identity-service base URL |
| `BILLING_URL` | `http://localhost:9100` | billing-service base URL |
| `GAME_NOVA_URL` | `http://localhost:8080` | game-nova base URL |
| `REDIS_ADDR` | `localhost:6379` | Redis для session store |
| `SESSION_SECRET` | — | (required, ≥32 bytes) для HMAC csrf-токенов |
| `BFF_IDLE_TIMEOUT` | `30m` | sliding-TTL сессии |
| `BFF_REFRESH_LEAD_TIME` | `60s` | за сколько до exp обновлять access |
| `BFF_COOKIE_DOMAIN` | (empty) | `.oxsar-nova.ru` в проде |
| `BFF_COOKIE_SECURE` | `true` | `false` в dev |
| `LOG_LEVEL` | `info` | debug/info/warn/error |

## Dev

```bash
go test ./...
go build ./cmd/server
SESSION_SECRET=$(openssl rand -hex 32) BFF_COOKIE_SECURE=false ./server
```

## Endpoints

```
GET  /healthz
POST /auth/login           # body: {username, password} → cookies + claims
POST /auth/logout          # clear session, revoke at identity
GET  /auth/me              # current session claims summary
*    /api/admin/billing/*  # → billing-service
*    /api/admin/game/*     # → game-nova
*    /api/admin/*          # → identity-service (RBAC, audit, users)
```
