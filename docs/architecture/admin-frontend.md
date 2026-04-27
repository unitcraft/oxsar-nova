# Admin-консоль: архитектура

**Введена**: план 53 (2026-04-27)
**Owner**: identity + admin-frontend.
**Связано**: [52-rbac-unification.md](../plans/52-rbac-unification.md),
[53-admin-frontend.md](../plans/53-admin-frontend.md),
[rbac.md](rbac.md).

---

## Обзор

Отдельная админ-консоль на `admin.oxsar-nova.ru` с production-grade
безопасностью (BFF, IP-allowlist, strict CSP, HSTS) и professional
utilitarian-дизайном (стиль Stripe Dashboard / Linear).

```
┌─ Browser ──────────────────────────────────────────┐
│  cookie: admin_session (HttpOnly Secure SameSite)  │
│  cookie: admin_csrf    (для X-CSRF-Token header)   │
└────────────────────────┬───────────────────────────┘
                         │ HTTPS
                         ▼
┌─ nginx (admin.oxsar-nova.ru) ──────────────────────┐
│  IP-allowlist, TLS, CSP, HSTS, X-Frame-Options     │
│  serve dist/ (admin-frontend SPA)                  │
│  proxy /(auth|api)/ → admin-bff                    │
└────────────────────────┬───────────────────────────┘
                         │
                         ▼
┌─ admin-bff (Go) ───────────────────────────────────┐
│  - Redis sessions (idle timeout 30 мин, sliding)   │
│  - Lazy auto-refresh JWT (за 60s до exp)           │
│  - CSRF double-submit middleware                   │
│  - Reverse proxy с JWT-injection                   │
└──┬─────────────┬───────────────────────┬───────────┘
   │             │                       │
   ▼             ▼                       ▼
identity     billing               game-nova
(/api/admin  (/api/admin/billing/*) (/api/admin/events/*)
 /roles*,
 /audit*,
 /users/*/roles*)
```

## Компоненты

### admin-frontend (`projects/admin-frontend/`)

- React 18 + TypeScript 5 (strict, noUncheckedIndexedAccess,
  exactOptionalPropertyTypes, verbatimModuleSyntax).
- Vite 5 (port 5174 в dev, manualChunks: react-vendor / tanstack /
  recharts).
- Tailwind 3 + custom palette (gray scale + accent blue, dark
  mode first).
- shadcn/ui copy-paste primitives в `src/components/ui/` (Button,
  Input, Card, Badge, Skeleton, Dialog).
- TanStack Query v5 + TanStack Table v8.
- React Router v6 + Zustand (auth-store + ui-store).
- react-hook-form + zod валидация.

### admin-bff (`projects/admin-bff/`)

- Go 1.23, chi router, go-redis v9, slog (JSON logs).
- Структура:
  - `cmd/server/main.go` — wiring chi + middlewares.
  - `internal/config/` — env-based config с валидацией
    (SESSION_SECRET ≥32 байт обязательно).
  - `internal/session/` — Redis-backed Store (CRUD + sliding
    TTL), cookies API.
  - `internal/identityclient/` — HTTP-клиент к identity, парсинг
    claims из access JWT (без проверки подписи — trusted-zone).
  - `internal/handler/` — Login/Logout/Me handler-методы +
    SessionLookup middleware (с lazy refresh) + CSRF middleware
    (constant-time compare).
  - `internal/proxy/` — reverse-proxy (httputil.ReverseProxy +
    Director стрипает Cookie/X-CSRF-Token, добавляет
    Authorization).

## Auth-flow (BFF, не PKCE)

После ревью OAuth 2.0 for Browser-Based Apps BCP (IETF) и
рекомендаций Auth0/Okta/Microsoft (с 2023) для new-build first-party
SPA в монорепо BFF — общепринятая default-практика, а не PKCE.

1. Browser открывает `admin.oxsar-nova.ru/login` → форма
   username/password.
2. POST `/auth/login` → admin-bff → POST identity `/auth/login` →
   получает JWT (access+refresh) + claims (sub/username/roles/
   permissions).
3. admin-bff создаёт сессию в Redis (`admin:sess:<uuid>`), TTL
   sliding 30 мин. Возвращает Set-Cookie `admin_session`
   (HttpOnly Secure SameSite=Strict) + `admin_csrf` (читаемый
   JS) + JSON со сводкой claims для UI guards.
4. Все последующие запросы: browser шлёт cookie автоматически
   (credentials: include), admin-bff читает `admin_session`
   cookie, достаёт JWT из Redis, инжектит
   `Authorization: Bearer <JWT>` и проксирует upstream.
5. State-changing запросы (POST/PUT/PATCH/DELETE) требуют
   header `X-CSRF-Token` (равный `admin_csrf` cookie) — middleware
   проверяет constant-time compare.
6. Auto-refresh: при `time.Until(claims.exp) < 60s` admin-bff
   ходит в identity `/auth/refresh`, обновляет JWT и сессию —
   прозрачно для frontend.
7. Logout: POST `/auth/logout` → admin-bff (best-effort) revoke в
   identity + `Del admin:sess:*` + clear cookies.

## RBAC (план 52)

Permissions из JWT-claims (см. [rbac.md](rbac.md)). UI скрывает
кнопки на основании `useAuth(s => s.hasPermission(perm))`, но это
UX — backend всё равно перепроверяет (defence in depth).

В admin-bff JWT-claims не парсятся для permissions (это job upstream-
сервисов через JWKS-loader). Frontend-side claims summary берётся
из `GET /auth/me` (admin-bff отдаёт из Redis-сессии).

## Security headers (nginx)

- **CSP**: `default-src 'self'`. `connect-src 'self'` — все API
  запросы только на admin-bff на том же домене (BFF замыкает
  блюр-зону). `frame-ancestors 'none'` — нельзя iframe'ить.
- **HSTS**: `max-age=63072000; includeSubDomains; preload` (2
  года, для submission в [hstspreload.org](https://hstspreload.org)).
- **X-Frame-Options DENY** — для legacy-браузеров без CSP
  frame-ancestors.
- **X-Content-Type-Options nosniff**.
- **Referrer-Policy strict-origin-when-cross-origin**.
- **Permissions-Policy**: camera/mic/geo все disabled.
- **X-Robots-Tag noindex,nofollow**.

NB: `add_header` в nginx не наследуется в location'ы со своим
`add_header` — security-headers продублированы в `/index.html`.

## IP-allowlist

nginx geo-block:
```
geo $admin_allowed {
    default 0;
    include /etc/nginx/admin-ips.conf;
}
```

Файл `/etc/nginx/admin-ips.conf` монтируется через docker-compose
volume. Формат: `<ip-or-cidr> 1;`. Если IP не в списке → 403 на
nginx-уровне (быстрее app-уровня).

`/healthz` исключён из проверки (для k8s/docker probes).

## Что отложено

- **Ф.7 Moderation** — заглушки, ждёт плана 48 (UGC moderation).
- **Ф.8 2FA / WebAuthn** — отдельный sub-план 53c (большая
  backend-работа в identity: WebAuthn enrollment + TOTP backup +
  recovery codes).
- **Ф.10 Cmd+K command palette + keyboard shortcuts** — UX polish.
- **Sub-план 53b game-namespace** — миграция остальных game-nova
  admin-routes (planets, fleets, credits, users) под `/api/admin/
  game/` префикс.
- **Sub-план 53d users-list** — endpoint `GET /api/admin/users` в
  identity для полноценного списка с фильтрами и пагинацией.

## Bundle size

Code-splitting по routes (React Router lazy):

| Chunk | Raw | Gzipped |
|---|---|---|
| react-vendor | 157 KB | 52 KB |
| index (main) | 110 KB | 31 KB |
| tanstack | 44 KB | 13 KB |
| UserDetail (с Dialog) | 43 KB | 15 KB |
| GameEvents | 12 KB | 4 KB |
| Audit | 8 KB | 3 KB |
| Roles | 4 KB | 2 KB |
| UsersLookup | 2 KB | 1 KB |

Бюджет 300 KB / main — соблюдён. CI guard в `.github/workflows/
admin-console.yml` падает если main превысит 300 KB.

## CI

`.github/workflows/admin-console.yml` (срабатывает на изменения
admin-bff/admin-frontend/Dockerfile.admin-*):
- admin-bff (Go 1.23): vet + test + build.
- admin-frontend (Node 20): typecheck + test + build + bundle-
  size guard.
- docker-build (depends_on оба job'а): сборка обоих images через
  buildx с GHA cache.
