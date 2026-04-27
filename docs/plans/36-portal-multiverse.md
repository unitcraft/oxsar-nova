# План 36: Портал мира Oxsar — мультивселенная, единая аутентификация, лента предложений

**Дата**: 2026-04-26  
**Статус**: Концепт (v3 — ADR-решения зафиксированы)  
**Домен**: oxsar-nova.ru  
**Затрагивает**: новые сервисы `auth-service`, `portal-backend`, рефактор игрового монолита,
новый `portal-frontend`, изменения в деплое

---

## Контекст и требования

1. **Несколько вселенных** с разными настройками (скорость, bash-лимиты, экономика).
2. **Единый аккаунт** — один логин для всех вселенных.
3. **OAuth** — вход через ВКонтакте, Mail.ru, Яндекс.
4. **Переключение вселенных из игры** — не выходя на портал.
5. **Главная страница-портал** — список вселенных + новости + лента предложений.
6. **Лента предложений** — игроки пишут идеи, прикладывают скриншоты, голосуют кредитами.
7. **Игра не запущена** — любые архитектурные изменения допустимы.
8. **Нет ограничений на БД** — каждый сервис получает отдельный Postgres-инстанс.

---

## Принятые архитектурные решения

| # | Вопрос | Решение |
|---|---|---|
| 1 | БД | Отдельный Postgres-инстанс на каждый сервис и каждую вселенную |
| 2 | Домен портала | `oxsar-nova.ru` |
| 3 | Домены игровых вселенных | Субдомены: `uni01.oxsar-nova.ru`, `uni02.oxsar-nova.ru` (имена — позже) |
| 3а | Домен Auth Service | Субдомен: `auth.oxsar-nova.ru` |
| 4 | Universe Switcher | npm-пакет в монорепо (`packages/universe-switcher/`), pnpm workspaces |
| 5 | Голоса за feedback | Множественные: 100 кредитов за голос, без лимита числа голосов |
| 6 | Game DB | Отдельная на каждую вселенную |
| 7 | Модерация feedback | Предложения проходят модерацию (admin approve) перед публикацией |
| 8 | Лимит предложений | Не более 5 активных предложений от одного игрока одновременно |
| 9 | Обсуждение предложений | Полноценные треды (комментарии + ответы на комментарии) |
| 10 | Профиль на портале | Баланс кредитов + полная история транзакций |
| 11 | Нагрузочный тест | Отдельный ops-тикет после Ф.8; тестировать каждый стек и всю связку |

---

## Архитектурное решение: «Независимые вселенные под единым порталом»

### Ключевое: каждая вселенная — полностью изолированный стек

```
game-server  +  game-worker  +  postgres (свой инстанс)  +  redis (свой инстанс)
```

Добавление вселенной = новый `docker-compose.uniN.yml` + строка в `projects/game-nova/configs/universes.yaml`.  
Никаких изменений в коде.

---

### Общая схема

```
                      oxsar-nova.ru
                            │
         ┌──────────────────┼──────────────────┐
         ▼                  ▼                  ▼
[ Portal Frontend ]  [ Portal Backend ]  [ Auth Service ]
  oxsar-nova.ru        oxsar-nova.ru/api   auth.oxsar-nova.ru
                                               :9000
         │                  │
         └────────┬──────────┘
                  ▼
     ┌────────────┴────────────┐
     ▼                         ▼
[ uni-standard.              [ uni-speed.
  oxsar-nova.ru ]              oxsar-nova.ru ]
  game-server :8080             game-server :8081
  game-worker                   game-worker
  postgres-uni1                 postgres-uni2
  redis-uni1                    redis-uni2
```

### Nginx / Reverse Proxy

```
oxsar-nova.ru          → portal-frontend (статика Vite)
oxsar-nova.ru/api/     → portal-backend :9001
auth.oxsar-nova.ru     → auth-service :9000
uni-standard.oxsar-nova.ru  → game-server-uni1 :8080
uni-speed.oxsar-nova.ru     → game-server-uni2 :8081
```

Wildcard SSL: `*.oxsar-nova.ru` — один сертификат (Let's Encrypt wildcard через DNS-01).

**DNS и SSL — детали**:
- Домен `oxsar-nova.ru` куплен на nic.ru.
- **Рекомендация**: делегировать DNS-зону на Selectel (поменять NS-записи на nic.ru на серверы Selectel). Домен остаётся на nic.ru, DNS управляется через Selectel. Занимает ~15 минут, стандартная практика.
- Тогда для wildcard-сертификата используется `certbot` + официальный плагин Selectel DNS API — полностью автоматическое продление каждые 90 дней.
- **Альтернатива без делегирования**: `certbot-dns-nic` (неофициальный плагин для nic.ru API) — работает, но менее надёжен. Добавить `--dns-nic-propagation-seconds 300` из-за задержек DNS на nic.ru.
- Настраивается в рамках Ф.8 (деплой) или отдельным ops-тикетом перед запуском.

---

## Сервисы и их инстансы БД

| Сервис | Postgres-инстанс | Redis-инстанс | Назначение |
|---|---|---|---|
| `auth-service` | `postgres-auth` (порт 5432) | `redis-auth` | users, oauth, tokens, global_credits |
| `portal-backend` | `postgres-portal` (порт 5433) | общий с auth | news, feedback, votes |
| `game-server-uni1` + `game-worker-uni1` | `postgres-uni1` (порт 5434) | `redis-uni1` | вселенная «Стандарт» |
| `game-server-uni2` + `game-worker-uni2` | `postgres-uni2` (порт 5435) | `redis-uni2` | вселенная «Ускоренная» |

На одном VPS все инстансы — отдельные Docker-контейнеры postgres с разными портами.  
При росте — каждый инстанс легко переезжает на отдельную машину без изменений кода.

---

## Auth Service (`auth.oxsar-nova.ru`)

### Почему выделенный сервис

- OAuth callback — единый URL, не привязан к вселенной.
- RSA private key не должен быть в каждом игровом инстансе.
- Универсальный membership (`universe_memberships`) — список вселенных юзера.
- Платежи и баланс кошелька — в **отдельном** billing-service (план 38),
  не в auth-service (different bounded context).

### JWT: RSA-256

Auth Service **выпускает** токены приватным ключом.  
Все остальные сервисы **верифицируют** публичным ключом через JWKS:

```
GET auth.oxsar-nova.ru/.well-known/jwks.json
```

### Claims в JWT

```json
{
  "sub":              "uuid-пользователя",
  "username":         "StarLord42",
  "global_credits":   1500,
  "active_universes": ["uni-standard", "uni-speed"],
  "roles":            ["player"],
  "iat": 1714000000,
  "exp": 1714003600
}
```

`active_universes` — вселенные, где у игрока есть аккаунт. Игровой сервер проверяет
своё `UNIVERSE_ID` в этом списке при каждом запросе.

### TTL access / refresh

| Среда | access | refresh | Где |
|---|---|---|---|
| dev | 60m | 30d (720h) | `deploy/docker-compose.yml` |
| prod | 15m | 7d (168h) | `deploy/docker-compose.multiverse.yml` |

**Почему различаются**: dev оптимизирован под удобство (не перелогиниваться часто),
prod — под security (короткий access = revocation действует быстро через TTL даже
без blacklist'а; refresh 7d — индустриальный стандарт OWASP).

При revoke через `/auth/logout` (план Ф.11/Critical-4) refresh попадает в
Redis-blacklist на оставшийся TTL.

### API Auth Service

```
POST /auth/register           — регистрация (email + username + password)
POST /auth/login              — вход по паролю
POST /auth/refresh            — обновить access token
POST /auth/logout             — отозвать refresh token
POST /auth/password           — смена пароля (current + new, требует JWT)
GET  /auth/me                 — профиль текущего пользователя
GET  /auth/me/universes       — мои активные вселенные (из universe_memberships)

GET  /auth/oauth/{provider}           — редирект к провайдеру
GET  /auth/oauth/{provider}/callback  — приём кода, выдача JWT

POST /auth/universe-token     — one-time handoff token для переключения вселенной
POST /auth/token/exchange     — обмен one-time code → JWT (после OAuth redirect)

# Платежи и баланс — НЕ в auth-service, см. план 38 (billing-service):
#   GET  /billing/wallet/balance
#   POST /billing/wallet/spend (внутренний)
#   POST /billing/orders + GET /billing/packages
#   POST /billing/webhooks/{provider}

GET  /.well-known/jwks.json   — публичный ключ RSA
```

### Схема БД `auth_db`

```sql
CREATE TABLE users (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username       TEXT UNIQUE NOT NULL,
    email          TEXT UNIQUE NOT NULL,
    password_hash  TEXT,                  -- NULL если только OAuth
    global_credits BIGINT NOT NULL DEFAULT 0,
    roles          TEXT[] NOT NULL DEFAULT '{player}',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ
);

CREATE TABLE oauth_accounts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id),
    provider    TEXT NOT NULL,            -- 'vk' | 'mailru' | 'yandex'
    provider_id TEXT NOT NULL,
    email       TEXT,
    linked_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (provider, provider_id)
);

CREATE TABLE refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id),
    token_hash  TEXT NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at  TIMESTAMPTZ
);

CREATE TABLE credit_transactions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id),
    delta       BIGINT NOT NULL,          -- >0 пополнение, <0 списание
    reason      TEXT NOT NULL,           -- 'payment' | 'feedback_vote' | 'universe_purchase'
    ref_id      TEXT,                    -- id платежа, feedback_post, вселенной
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

---

## Игровой сервер: изменения

Изменения минимальны:

1. **Убрать** `POST /api/auth/register`, `POST /api/auth/login`, `POST /api/auth/refresh`.
2. **JWT middleware**: HS256 → RS256, публичный ключ загружается из `AUTH_JWKS_URL`.
3. **`UNIVERSE_ID`** в конфиге — сервер знает свой идентификатор.
4. **Lazy join**: при первом запросе нового игрока — создать запись в локальной
   `users` таблице по `global_user_id` из JWT.
5. **Таблица `users` в game DB**: убрать `email`, `password_hash`; добавить
   `global_user_id UUID NOT NULL UNIQUE`.
6. **Платёжные webhooks** переезжают в **billing-service** (план 38).

---

## Переключение вселенных из игры

Игрок в `uni-standard.oxsar-nova.ru` кликает «Ускоренная» в шапке:

```
1. Frontend → POST auth.oxsar-nova.ru/auth/universe-token
             { "universe_id": "uni-speed" }   + Bearer <access_token>

2. Auth Service → генерирует одноразовый handoff_token (UUID, TTL 30с, Redis)
   → возвращает { "handoff_token": "...", "url": "https://uni-speed.oxsar-nova.ru" }

3. Frontend → редирект на uni-speed.oxsar-nova.ru/auth/handoff?token=<handoff_token>

4. uni-speed game-server → POST auth.oxsar-nova.ru/auth/token/exchange { "code": token }
   → получает полноценный JWT для этого игрока
   → lazy join: создать/найти запись в postgres-uni2
   → установить сессию

5. Игрок оказывается в игре вселенной 2 без ввода пароля.
```

Этот же флоу работает при первом входе в вселенную с портала.

---

## Universe Switcher (ADR #4 — решено)

**Решение**: npm-пакет `@oxsar/universe-switcher` в монорепо (pnpm workspaces).

Компонент в шапке игры: список вселенных + баланс global credits.

Структура монорепо:
```
packages/
  universe-switcher/    ← React-компонент, shared
    src/UniverseSwitcher.tsx
    src/useCredits.ts
portal/
  frontend/             ← oxsar-nova.ru
  backend/              ← portal API
auth/
  backend/              ← auth-service
game/
  frontend/             ← uni-*.oxsar-nova.ru (один фронтенд, разный UNIVERSE_ID)
  backend/              ← game-server + worker (единый Go-модуль)
```

Игровой фронтенд — **один для всех вселенных**, параметризуется через `UNIVERSE_ID`
и `UNIVERSE_NAME` в конфиге сборки. Если вселенные в будущем потребуют принципиально
разного UI — выделяется отдельный фронтенд, но сейчас это не нужно.

Обновление Universe Switcher — bump версии пакета, пересборка всех фронтендов при следующем деплое.

---

## Portal Backend (`oxsar-nova.ru/api/`)

Новый бинарник `backend/cmd/portal/`, своя БД `portal_db`.

### Каталог вселенных

Публичный список вселенных (что вообще существует в проекте) живёт на портале,
а не в auth-service. Auth-service отвечает за токены и членство юзера, не за
каталог.

API:
- `GET /api/universes` — публичный список из `configs/universes.yaml`. Дёргает
  главная страница портала и Universe Switcher в шапке игры.

ADR (2026-04-27): `GET /auth/universes` ранее был в auth-service — перенесён
в portal-backend. Reason: каталог это свойство проекта, а не свойство выпуска
JWT. Auth-service оставляет `GET /auth/me/universes` (мои активные вселенные
из `universe_memberships`), для шапки UI.

### Новости

```sql
CREATE TABLE news (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title        TEXT NOT NULL,
    body_md      TEXT NOT NULL,
    author_id    UUID NOT NULL,
    published_at TIMESTAMPTZ,
    pinned       BOOLEAN NOT NULL DEFAULT false,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

API:
- `GET /api/news?limit=20&offset=0` — публичный.
- `POST /api/news` — admin only.
- `PATCH /api/news/{id}` — admin only.
- `DELETE /api/news/{id}` — admin only.

### Лента предложений (Feedback)

```sql
CREATE TABLE feedback_posts (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id     UUID NOT NULL,
    title         TEXT NOT NULL,
    body_md       TEXT NOT NULL,
    moderation    TEXT NOT NULL DEFAULT 'pending', -- pending | approved | rejected
    status        TEXT NOT NULL DEFAULT 'open',    -- open | planned | done | rejected
    vote_count    INT  NOT NULL DEFAULT 0,         -- сумма потраченных кредитов
    comment_count INT  NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- Индекс для быстрой проверки лимита активных предложений игрока
CREATE INDEX ON feedback_posts (author_id) WHERE moderation = 'approved' AND status = 'open';

CREATE TABLE feedback_attachments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id     UUID NOT NULL REFERENCES feedback_posts(id) ON DELETE CASCADE,
    url         TEXT NOT NULL,    -- Selectel Object Storage
    mime_type   TEXT NOT NULL,
    size_bytes  INT  NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Каждый голос = 100 кредитов, одним игроком можно голосовать многократно
CREATE TABLE feedback_votes (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id       UUID NOT NULL REFERENCES feedback_posts(id) ON DELETE CASCADE,
    user_id       UUID NOT NULL,
    credits_spent INT  NOT NULL DEFAULT 100,
    voted_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ON feedback_votes (post_id);
CREATE INDEX ON feedback_votes (user_id);

-- Обсуждения: древовидные треды (parent_id NULL = корневой комментарий)
CREATE TABLE feedback_comments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id     UUID NOT NULL REFERENCES feedback_posts(id) ON DELETE CASCADE,
    parent_id   UUID REFERENCES feedback_comments(id) ON DELETE CASCADE,
    author_id   UUID NOT NULL,
    body_md     TEXT NOT NULL,
    moderation  TEXT NOT NULL DEFAULT 'pending', -- pending | approved | rejected
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ON feedback_comments (post_id);
CREATE INDEX ON feedback_comments (parent_id);
```

API:
- `GET /api/feedback?sort=top|new&status=open&limit=20&offset=0` — публичный (только approved).
- `POST /api/feedback` — создать предложение (JWT required).
  Проверка: не более 5 активных approved+open предложений от автора.
  Новое предложение уходит в `moderation=pending`.
- `GET /api/feedback/{id}` — детали + вложения + суммарные голоса (только approved).
- `POST /api/feedback/{id}/vote` — один голос = 100 global credits (JWT required).
  Атомарно: `credits/spend` → вставить строку в `feedback_votes` → обновить `vote_count`.
- `POST /api/feedback/{id}/attachments` — загрузить скриншот (max 5MB, JPEG/PNG/GIF).
- `GET /api/feedback/{id}/comments` — дерево комментариев (только approved).
- `POST /api/feedback/{id}/comments` — добавить комментарий (JWT required, уходит в pending).
- `POST /api/feedback/{id}/comments/{cid}/reply` — ответить на комментарий.
- `PATCH /api/feedback/{id}/status` — admin only (open→planned→done|rejected).
- `POST /api/admin/feedback/{id}/moderate` — admin: approve/reject предложение.
- `POST /api/admin/feedback/{id}/comments/{cid}/moderate` — admin: approve/reject комментарий.

**Рейтинг**: `ORDER BY vote_count DESC, created_at DESC`.  
**Лимит**: при попытке создать 6-е активное предложение — 422 с объяснением.  
Pinned-новости выводятся отдельным блоком над лентой.

---

## Global Credits → ВЫНЕСЕНО В ПЛАН 38 (billing-service)

Баланс и история транзакций живут в **billing-db** (отдельная Postgres).
Каждое движение — INSERT в `transactions` (immutable, double-entry).

**Пополнение**: платёжный шлюз → `POST /billing/webhooks/{provider}` →
INSERT transaction + UPDATE wallet.balance.

См. [plans/38-billing-service.md](38-billing-service.md). Ниже секция
оставлена для исторического контекста (раньше планировалось в auth).

**Списание**:
- Голос за feedback: Portal Backend → `POST /auth/credits/spend {user_id, amount: 100, reason: "feedback_vote", ref_id}`.
- Покупка в игре: game-server → `POST /auth/credits/spend {user_id, amount, reason: "universe_purchase", universe_id}`.

**В JWT**: `global_credits` — снапшот на момент выдачи токена (TTL 1 час).  
Актуальный баланс: `GET /auth/credits/balance` (отдельный запрос, кешируется 30 сек).

---

## OAuth (Social Login)

**Authorization Code Flow + PKCE** для всех провайдеров.

| Провайдер | Библиотека | scope |
|---|---|---|
| ВКонтакте | ручной OAuth2 (VK ID) | `email` |
| Mail.ru | ручной OAuth2 | `email` |
| Яндекс | ручной OAuth2 (Яндекс OAuth) | `login:email` |

Флоу:
1. Portal/игровой фронтенд: кнопка «Войти через ВК».
2. Редирект → `GET auth.oxsar-nova.ru/auth/oauth/vk?redirect_back=<url>`.
3. Auth Service: генерирует `state` + `code_verifier` (PKCE), сохраняет в Redis (TTL 5 мин).
4. Редирект → `api.vk.com/authorize?...`.
5. Callback → `GET /auth/oauth/vk/callback?code=...&state=...`.
6. Проверка state, обмен code → access_token, получение email + provider_id.
7. Upsert `users` + `oauth_accounts`. Выдача JWT.
8. Генерация one-time code (30 сек, Redis) → редирект на `redirect_back?code=...`.
9. Frontend: `POST /auth/token/exchange {code}` → JWT.

Безопасность: PKCE обязателен, state для CSRF, JWT не передаётся напрямую в URL.

---

## Portal Frontend (`oxsar-nova.ru`)

Отдельное Vite-приложение `frontend/portal/`.

Роуты:
- `/` — главная: карточки вселенных + топ-5 новостей + топ-5 предложений.
- `/universes` — все вселенные с деталями и статусом.
- `/news` — лента новостей.
- `/news/{id}` — статья.
- `/feedback` — лента предложений (top/new, фильтр по статусу).
- `/feedback/new` — форма создания предложения + загрузка скриншотов.
- `/feedback/{id}` — пост + скриншоты + кнопка голосования + история голосов.
- `/login` — email+пароль или OAuth-кнопки (ВК / Mail.ru / Яндекс).
- `/register` — регистрация.
- `/profile` — глобальный профиль: список вселенных, баланс кредитов, история транзакций
  (пополнения, голоса за feedback, покупки в игре) с пагинацией.
- `/profile/feedback` — мои предложения (включая pending/rejected).

Шапка: логотип + список вселенных + баланс кредитов + кнопка «Войти» / аватар.

---

## Деплой: docker-compose per universe

```yaml
# docker-compose.portal.yml
services:
  auth-service:
    build: { context: ./backend, args: { CMD: auth-service } }
    ports: ["9000:9000"]
    environment: { DATABASE_URL: postgres://..., REDIS_URL: redis://redis-auth:6379 }
    depends_on: [postgres-auth, redis-auth]
  portal-backend:
    build: { context: ./backend, args: { CMD: portal } }
    ports: ["9001:9001"]
    environment: { DATABASE_URL: postgres://..., AUTH_JWKS_URL: http://auth-service:9000 }
    depends_on: [postgres-portal]
  portal-frontend:
    build: ./frontend/portal
    ports: ["3000:80"]
  postgres-auth:   { image: postgres:16, ports: ["5432:5432"] }
  postgres-portal: { image: postgres:16, ports: ["5433:5432"] }
  redis-auth:      { image: redis:7 }

# docker-compose.uni-standard.yml
services:
  game-server:
    build: { context: ./backend, args: { CMD: server } }
    ports: ["8080:8080"]
    environment:
      UNIVERSE_ID: uni-standard
      AUTH_JWKS_URL: http://auth.oxsar-nova.ru
      DATABASE_URL: postgres://...
      REDIS_URL: redis://redis-uni:6379
    depends_on: [postgres-uni, redis-uni]
  game-worker:
    build: { context: ./backend, args: { CMD: worker } }
    environment: { UNIVERSE_ID: uni-standard, DATABASE_URL: postgres://... }
    depends_on: [postgres-uni]
  postgres-uni: { image: postgres:16, ports: ["5434:5432"] }
  redis-uni:    { image: redis:7 }

# docker-compose.uni-speed.yml  — аналогично, UNIVERSE_ID=uni-speed, ports 8081/5435
```

---

## Фазы реализации

### Ф.1 — Auth Service + RSA-256 (3–4 дня)

1. `backend/cmd/auth-service/` — новый бинарник.
2. `backend/pkg/auth/` — shared JWT-пакет (перенос из `backend/internal/auth/`).
3. RSA-256: генерация ключевой пары, JWKS endpoint.
4. Миграции `auth_db`: `users`, `oauth_accounts`, `refresh_tokens`, `credit_transactions`.
5. Endpoints: `register`, `login`, `refresh`, `logout`, `me`.
6. Игровой сервер: убрать auth-endpoints, RS256 middleware через JWKS.
7. Скрипт: перенос существующих users из game DB → auth DB.
8. E2E тест: register → JWT → запрос к игровому серверу.

### Ф.2 — Universe Registry (1 день)

1. `projects/game-nova/configs/universes.yaml` — настройки всех вселенных.
2. Portal Backend: `GET /api/universes` (читает yaml) + агрегация онлайна
   из Redis вселенных. Каталог живёт **только** на портале (см. ADR 2026-04-27
   в §«Каталог вселенных»).
3. Auth Service универсами не занимается, кроме `universe_memberships` юзера.

### Ф.3 — Portal Backend MVP (3–4 дня)

1. `backend/cmd/portal/` — новый бинарник.
2. Миграции `portal_db`: `news`, `feedback_posts`, `feedback_votes`, `feedback_attachments`,
   `feedback_comments`.
3. API новостей + API feedback (list, create, get, vote, comments).
4. Модерация: очередь pending для admin, approve/reject.
5. Лимит 5 активных предложений — проверка при создании.
6. Загрузка вложений в Selectel Object Storage.
7. Auth Service: `GET /auth/credits/history?limit=20&offset=0` — история транзакций
   (читается из `credit_transactions`).

### Ф.4 — Portal Frontend (3–4 дня)

1. `frontend/portal/` — новое Vite-приложение.
2. Страницы: главная, вселенные, новости, feedback.
3. Формы: вход, регистрация, создание предложения + загрузка скриншотов.
4. Интеграция с Auth Service + Portal Backend.

### Ф.5 — Переключение вселенных из игры (2 дня)

1. Auth Service: `POST /auth/universe-token` (handoff).
2. Игровой сервер: `GET /auth/handoff?token=...` — принять handoff, lazy join.
3. Universe Switcher в шапке игрового фронтенда (решение по ADR #4).
4. Отображение баланса global credits в шапке.

### Ф.6 — OAuth Social Login (3–4 дня)

1. Auth Service: OAuth флоу ВКонтакте + Mail.ru + Яндекс.
2. PKCE + state + one-time code в Redis.
3. Upsert `oauth_accounts`.
4. OAuth-кнопки на Portal Frontend + игровом фронтенде.

### Ф.7 — Global Credits + платёжный шлюз → ВЫНЕСЕН В ПЛАН 38

См. [plans/38-billing-service.md](38-billing-service.md).

Решение (2026-04-27): платежи и кошельки выделены в **отдельный микросервис**
`billing-service` (а не в auth-service, как планировалось изначально).
Причина: auth-service отвечает за identity, billing — за money. Это разные
bounded context'ы, объединять нельзя (нарушение SRP, общая blast radius
при компрометации, общий compliance-scope).

В рамках плана 38:
- Списание/пополнение, история — в billing-db (отдельный Postgres).
- Платёжные шлюзы (Robokassa, Enot.io, mock) — в billing.
- Голосование за feedback дёргает `POST billing/wallet/spend` с idempotency.
- `users.global_credits` и `credit_transactions` удаляются из auth-db.
- Баланс берётся из `GET billing/wallet/balance` (не из JWT-claims).

### Ф.8 — Запуск двух вселенных (2–3 дня)

Запуск сразу с двумя вселенными. Вторая — существенно быстрее по скорости (×N),
остальные настройки (bash-лимиты, экономика) уточняются перед деплоем.

1. `projects/game-nova/configs/universes.yaml` — финальные настройки обеих вселенных (имена, скорость).
2. `docker-compose.uni01.yml` + `docker-compose.uni02.yml`.
3. Деплой обоих стеков + smoke-test каждого.
4. Переключение между вселенными — проверить handoff-флоу end-to-end.
5. DNS: делегировать зону на Selectel, получить wildcard SSL.

### Ф.9 — Реструктуризация репозитория (½ дня) ✅ 2026-04-26

Привести структуру папок к модели «один домен — одна папка»:

```
portal/
  frontend/     ← было frontend/portal/
  backend/      ← будет (backend/cmd/portal + internal/portalsvc + projects/portal/migrations/)
auth/
  backend/      ← будет (backend/cmd/auth-service + internal/authsvc + projects/auth/migrations/)
game/
  frontend/     ← было frontend/ (src/, e2e/, public/, scripts/, *.config.*)
  backend/      ← будет (весь backend/ — Go-модуль остаётся единым)
```

Что обновляется при переносе:
- `Makefile` — все пути к фронтендам и бэкендам
- `deploy/docker-compose*.yml` — context/dockerfile пути, bind-mount пути
- `backend/Dockerfile` — пути к cmd/*
- `tsconfig.json`, `vite.config.ts`, `package.json` — если есть относительные пути
- Go-импорты — module path остаётся `github.com/oxsar/nova/backend`, только физические пути меняются

### Ф.10 — Разделение Go-модулей auth / portal / game-nova (½–1 день)

После Ф.9 код auth-service и portal-backend физически живёт внутри
`projects/game-nova/backend/` (общий `go.mod`, бинарники собираются вместе).
Это нарушает доменную независимость: изменения в game-nova триггерят пересборку
auth/portal, и наоборот.

**Цель:** разнести по самостоятельным Go-модулям.

#### Module paths

Module path не обязан быть реальным URL — это идентификатор в монорепо.
Принято: префикс `oxsar/`, после него — имя домена. Никакой `github.com`,
никакой `nova` в путях.

```
projects/game-nova/backend/   module oxsar/game-nova   (был github.com/oxsar/nova/backend)
projects/auth/backend/        module oxsar/auth
projects/portal/backend/      module oxsar/portal
```

Переименование game-nova — это `sed`-замена `github.com/oxsar/nova/backend → oxsar/game-nova`
во всех `.go`-файлах + `go.mod`. Делается одной командой, не 600 правок.

#### Целевая структура

```
projects/game-nova/backend/
  cmd/server/                ← бинарь /app/server
  cmd/worker/                ← бинарь /app/worker
  cmd/tools/...              ← CLI-утилиты
  internal/...               (БЕЗ authsvc, portalsvc)
  pkg/...
  go.mod                     module oxsar/game-nova

projects/auth/backend/
  cmd/server/main.go         ← бинарь /app/server (был cmd/auth-service)
  internal/authsvc/          (копия из game-nova)
  internal/auth/             (только password.go — argon2 хеш)
  internal/httpx/            (копия)
  internal/repo/             (только tx.go — InTx обёртка)
  internal/storage/          (копия)
  internal/universe/         (копия)
  pkg/ids/                   (копия)
  pkg/jwtrs/                 (копия)
  pkg/metrics/               (копия)
  go.mod                     module oxsar/auth
  Dockerfile

projects/portal/backend/
  cmd/server/main.go         ← бинарь /app/server (был cmd/portal)
  internal/portalsvc/        (копия из game-nova)
  internal/httpx/            (копия)
  internal/repo/             (копия tx.go)
  internal/storage/          (копия)
  internal/universe/         (копия)
  pkg/ids/                   (копия)
  pkg/jwtrs/                 (копия)
  pkg/metrics/               (копия)
  go.mod                     module oxsar/portal
  Dockerfile
```

Все три бинарника называются одинаково — `/app/server`. В compose
различаются именем сервиса и Dockerfile-ом.

#### Решение по shared-коду — вариант (b): дублировать минимум

Без go workspace, без отдельного shared-модуля. Дубль ~600–800 строк
shared-кода в каждом из двух новых доменов. Каждый файл-копия начинается
с маркера:

```go
// DUPLICATE: этот файл скопирован между Go-модулями.
// При изменении синхронизировать ВСЕ копии:
//   - projects/game-nova/backend/<путь>
//   - projects/auth/backend/<путь>
//   - projects/portal/backend/<путь>
// Причина дубля: каждый домен — отдельный go.mod, без shared-модуля.
```

Тесты (`*_test.go`) копируются вместе с пакетом — `go test ./...` в
каждом модуле сразу проверяет, что копия не сломалась.

**Почему не workspace/shared-модуль:** проект разрабатывается одним
человеком, версионирование общего кода не нужно. Workspace в Docker-сборках
ломает кеширование слоёв, требует `replace`-директив с относительными
путями, ломается на CI. Дубль из ~10 файлов проще ментально.

#### Что обновляется

- 3 новых `go.mod` (auth, portal, game-nova переименовывается)
- `deploy/docker-compose.yml` — `dockerfile: projects/auth/backend/Dockerfile`,
  `projects/portal/backend/Dockerfile`
- 2 новых Dockerfile (auth/backend, portal/backend) — копия
  `game-nova/backend/Dockerfile` с поправленными путями и одним бинарником
- Удалить из game-nova/backend: `cmd/auth-service/`, `cmd/portal/`,
  `internal/authsvc/`, `internal/portalsvc/`
- `Makefile` — entry-points game-nova (оставить как есть, auth/portal
  собираются через `cd projects/auth/backend && go build ./cmd/server`)

#### Acceptance

- `go build ./...` зелёный во всех трёх модулях
- `go test ./...` зелёный во всех трёх модулях
- `docker compose -f deploy/docker-compose.yml up --build` поднимает все
  сервисы, healthchecks проходят
- `grep -rn "DUPLICATE" projects/auth projects/portal` показывает маркеры
  во всех скопированных пакетах

### Ф.11 — Подключить auth-service в dev-стек и frontend (1–2 дня)

После Ф.10 auth-service существует как отдельный модуль, но **не подключён** ни к
dev-окружению, ни к фронтенду:

- В `deploy/docker-compose.yml` сервиса `auth-service` нет (он только в multiverse).
- Игровой backend верифицирует JWT симметричным `JWT_SECRET` (HS256), не RSA через JWKS.
- Игровой frontend ходит на старый `/api/auth/login` в game-nova, а не на auth-service.
- Portal Frontend ProfilePage показывает заглушку, не реальные данные из auth-service.

**Цель:** запустить auth-service локально, подключить к нему оба фронта и игровой бэкенд.
**OAuth (Ф.6) в эту фазу не входит** — только email/пароль.

#### Шаги

1. **auth-service в dev-compose**
   - Добавить сервисы `auth-db` (postgres-инстанс), `auth-migrate`, `auth-service` в
     `deploy/docker-compose.yml`. Порт наружу: 9000.
   - JWT_SECRET убирается отовсюду в dev. RSA-ключ генерируется auth-service-ом сам
     при старте (см. `LoadOrGenerateKey` в `pkg/jwtrs`).

2. **Vite proxy для /auth/***
   - В `projects/game-nova/frontend/vite.config.ts` добавить второй прокси:
     `/auth/* → http://auth-service:9000` (по аналогии с `/api/* → backend:8080`).
   - В `projects/portal/frontend/vite.config.ts` — то же самое.
   - Принцип: фронтенд всегда использует относительные пути. В dev — vite proxy,
     в prod — nginx. Никаких `import.meta.env.VITE_AUTH_URL`, никакого CORS.

3. **Игровой frontend → auth-service**
   - Заменить вызовы `/api/auth/login`, `/api/auth/register` на `/auth/login`,
     `/auth/register`.
   - JWT хранить в Zustand-сторе (как сейчас), подсовывать в `Authorization: Bearer`.
   - Старая логика `/api/auth/*` в game-nova остаётся для обратной совместимости
     до конца Ф.11; в конце фазы — удалить.

4. **Portal Frontend ProfilePage → auth-service**
   - Сейчас профиль с заглушкой. Подключить к `/auth/me`, `/auth/credits/balance`,
     `/auth/credits/history` через TanStack Query.

5. **Игровой backend → JWKS от auth-service**
   - В compose добавлена `AUTH_JWKS_URL=http://auth-service:9000`. `cmd/server/main.go`
     уже умел переключаться: при наличии `AUTH_JWKS_URL` использует RSA через
     `auth.LoadVerifier`, иначе — legacy HS256.
   - `JWT_SECRET` удалён из backend и worker секций dev-compose.
   - Что **остаётся** для следующей фазы: старые ручки `/api/auth/login|register|refresh`
     ещё есть в `internal/auth/handler.go` (никто не вызывает) и lazy-join не реализован.
     Делается в Ф.12 (Universe Switcher + handoff).

#### Acceptance

- ✅ `docker compose up` поднимает auth-service вместе со стеком (healthcheck зелёный).
- ✅ На `http://localhost:5173` (игра) кнопка «Логин» дёргает `/auth/login` (через
  vite proxy → auth-service:9000). Регистрация и логин работают, JWT (RSA) сохраняется
  в Zustand-сторе.
- ✅ На `http://localhost:5174` (портал) ProfilePage дёргает `/auth/me`,
  `/auth/credits/balance` через portalApi.auth — реальные данные из auth-service.
- ✅ Игровой backend стартует в режиме `auth mode: RSA-256 via JWKS`. Запрос с
  RSA-токеном на защищённый эндпоинт (`/api/me`, `/api/galaxy` и т.п.) проходит
  middleware-валидацию (валидная подпись → 200/4xx по бизнес-логике, не 401).
- ✅ В `deploy/docker-compose.yml` нет `JWT_SECRET`.
- ❌ **Не закрыто Ф.11**: запрос на `/api/me` с RSA-токеном падает с 500 «no rows in result set»,
  потому что юзера ещё нет в `users` table game-nova. Это **lazy-join**, делается в Ф.12.
  До Ф.12 фронтенд после логина не сможет работать с игровым API без ручного создания
  записи в `users` (через старый seed или будущий handoff).

#### Что НЕ входит в Ф.11

- OAuth Social Login (ВКонтакте, Mail.ru, Яндекс) — пока не делаем (по решению пользователя).
- Universe Switcher и handoff-flow между вселенными (lazy-join, удаление старых
  `/api/auth/*`) — отдельная фаза Ф.12.
- Перенос платёжных webhook'ов — план 38 (billing-service).

### Ф.12 — Lazy-join, Universe Switcher и handoff между вселенными (1–2 дня)

После Ф.11 RSA-токены принимаются backend, но юзера ещё нет в `users` table game-nova,
поэтому `/api/me` падает 500. Эту фазу закрываем.

**Принцип (паттерн 2: lazy в middleware с ON CONFLICT).** Frontend ничего не знает
про зеркалирование юзеров — после логина в auth-service сразу шлёт обычные API-запросы
к game-nova. Middleware при первом запросе с неизвестным `sub` создаёт строку в `users`,
назначает стартовую планету и шлёт welcome — всё внутри одной транзакции с
`INSERT ... ON CONFLICT (id) DO NOTHING` (защита от гонки одновременных запросов).

Альтернатива (handoff endpoint, который фронт дёргает явно) была отвергнута:
- лишний раунд-трип на каждом первом входе
- frontend начинает знать про распределённую архитектуру
- если завтра появится третий downstream — фронт надо учить дёргать ещё один handoff

В Auth0 / Supabase / Firebase это решается тем же паттерном 2 (lazy + ON CONFLICT).

**Не путать** с handoff `POST /auth/universe-token` для перехода uni01 → uni02 —
это **другой** handoff, нужен только для Universe Switcher (один-time-token,
редирект между вселенными). Он остаётся.

#### Шаги

1. **Email в JWT claims**
   - В `pkg/jwtrs.Claims` и `IssueInput` добавить поле `Email` (во всех 3 модулях,
     синхронно — это DUPLICATE-пакет).
   - Auth-service при `Issue` кладёт email из БД в claims.
   - Зачем: lazy-create в game-nova нужно email для INSERT (NOT NULL UNIQUE), без
     дополнительного запроса /auth/me.

2. **Миграция `password_hash` → nullable**
   - `ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL`.
   - Старые юзеры, созданные через `/api/auth/register`, продолжают работать
     (у них хеш есть). Новые из lazy-create будут с `NULL`.

3. **RSAMiddleware кладёт полные claims в context**
   - Сейчас кладётся только `sub`. Нужны `username` и `email` для lazy-create.
   - Добавить `RSAClaims(ctx) (Claims, bool)` в дополнение к `UserID(ctx)`.

4. **Lazy-create в middleware**
   - Новый middleware `EnsureUser(pool, starterSvc, automsgSvc)` после `RSAMiddleware`.
   - При каждом аутентифицированном запросе делает `INSERT INTO users(id, username, email, password_hash) VALUES (sub, username, email, NULL) ON CONFLICT (id) DO NOTHING`.
   - Если RowsAffected == 1 (юзер был только что создан) — асинхронно (`go func`)
     назначить стартовую планету через `starter.Assign(ctx, uid)` и отправить
     welcome через `automsg.Send`. Если упадёт — не критично, можно ретраить.
   - Регистрация `universe_memberships` в auth-service (`POST /auth/universes/register`) —
     тоже асинхронно.

5. **Удалить старые ручки `/api/auth/login|register|refresh` из game-nova**
   - Из `internal/auth/handler.go` методы `Register`, `Login`, `Refresh`.
   - Из `internal/auth/service.go` — соответствующие методы и `JWTIssuer`.
   - `internal/auth/password.go` оставить (HashPassword не используется,
     но удалим его в финальной чистке).
   - Маршруты в `cmd/server/main.go` удалить. Legacy HS256 режим тоже —
     `JWT_SECRET` env удалить из `config.go`.
   - Тесты обновить.

6. **Universe Switcher (минимально)**
   - Компонент в шапке игры. Получает список из portal-backend.
   - В vite-config игры добавить `/portal-api/* → portal-backend:8090`,
     фронт дёргает `/portal-api/universes`.
   - Выпадашка: список всех вселенных, текущая помечена.
   - При клике на другую: `POST /auth/universe-token` (auth-service выпускает
     one-time handoff token), редирект на `uniXX.oxsar-nova.ru/api/auth/handoff?token=...`.
   - В game-nova endpoint `GET /api/auth/handoff?token=...` — принять
     one-time токен, обменять в auth-service на полноценный JWT, поставить
     в localStorage через query-param и редиректить на `/`.

#### Acceptance

- ✅ Миграция 0067 применяется чисто, тесты `internal/auth` зелёные.
- ✅ `register` в auth-service → запрос `/api/me` в game-nova → 200 с username.
  Юзер создан в game-db с password_hash=NULL, стартовая планета назначена.
- ✅ При повторных запросах юзер уже создан (ON CONFLICT — нет лишних INSERT).
- ✅ Старые `/api/auth/login|register|refresh` возвращают 401 (попадают на
  generic auth middleware, без валидного токена).
- ✅ UniverseSwitcher подключён в шапку игры. Backend-эндпоинты
  `/api/universes`, `/api/universes/switch`, `/auth/handoff` работают.

#### Что НЕ входит в Ф.12

- Полный двух-вселенский dev-стек (uni01 + uni02 одновременно) — это Ф.8.
- Платёжные webhook'и и кошельки — план 38 (billing-service).
- Финальная чистка legacy HS256 кода (`internal/auth/handler.go::Register/Login/Refresh`,
  `service.go::register/login`, `password.go`, `JWT_SECRET` в config) — отдельным
  PR после Ф.12 (мёртвый код, не блокер).

### Ops: нагрузочный тест (отдельный тикет, ~1–2 дня)

После Ф.8, до публичного анонса. Обновить [ops/vps-sizing.md](../ops/vps-sizing.md)
с учётом новой архитектуры (несколько стеков). Тестировать:
- Каждый game-server независимо (как сейчас).
- Auth Service под пиковой нагрузкой (все вселенные логинятся одновременно).
- Всю связку: портал → auth → две вселенных одновременно.

### Что закрыто на 2026-04-27

- ✅ **Ф.9** Реструктуризация репозитория (один домен — одна папка).
- ✅ **Ф.10** Разделение Go-модулей (`oxsar/game-nova`, `oxsar/auth`, `oxsar/portal`)
  с `DUPLICATE`-маркерами в shared-коде.
- ✅ **Ф.11** auth-service подключён к dev-стеку, vite proxy `/auth/*`,
  фронты переключены, backend на JWKS.
- ✅ **Ф.12** Lazy-create в middleware (`EnsureUserMiddleware`), миграция
  `password_hash → NULLABLE`, Universe Switcher в шапке.
- ✅ **Pre-prod sweep** (Critical-1..6, Functional-8, Nice-10, Космет.13):
  RSA fail-fast, удалён HS256, rate-limit, /auth/logout с revoke,
  universe_memberships записывается, /auth/password в auth-service,
  bootstrap retry, email убран из JWT, бинарники из git-tracking.

End-to-end в dev: register в auth-service → JWT (RSA) → `/api/me` в game-nova
возвращает 200, юзер автоматически создан в game-db со стартовой планетой,
запись в `universe_memberships` есть. Logout → refresh потом 401. 6-й login
за минуту → 429.

### Что должно быть до прода (детали в [simplifications.md](../simplifications.md))

**Critical (security / data integrity) — ✅ ВСЁ ЗАКРЫТО 2026-04-27:**

1. ✅ RSA-ключ — fail-fast без autogen в проде (`AUTH_KEY_AUTOGEN=0` default).
2. ✅ Legacy HS256 fallback удалён, AUTH_JWKS_URL обязателен.
3. ✅ Rate-limiter на /auth/login (5/мин), /auth/register (10/мин).
4. ✅ POST /auth/logout + Redis-blacklist на refresh-jti.
5. ✅ EnsureUserMiddleware регистрирует `universe_memberships`.
6. ✅ Смена пароля — `POST /auth/password` в auth-service.

**Functional (не блокирует security, но нужно для multi-universe):**

- ⏳ **План 38** — billing-service: кошельки, платёжные webhooks
  (Robokassa/Enot/mock), голосование за feedback с idempotency.
  Раньше был Ф.7 в auth-service — после ревизии вынесен в отдельный
  микросервис (different bounded context).
- ⏳ **Ф.8** — второй вселенский стек (uni02 Speed) для проверки Universe Switcher
  E2E. DNS на Selectel + wildcard SSL.
- ✅ `bootstrapNewUser` retry-логика — если стартовая планета не назначилась,
  повторная попытка на следующем запросе.

**Nice to have:**

- ⏳ KeySet (current + previous) для ротации RSA без выкидывания живых JWT.
- ✅ Email из JWT claims убран (PII), Email NULLABLE в game-db.
- ⏳ OAuth Social Login (Ф.6) — отложен по решению пользователя.

**Косметика:**

- ✅ Мёртвый код auth handlers (`Register/Login/Refresh`) удалён вместе с
  `service.go`, `jwt.go`, `ratelimit.go`. `password.go` оставлен (используется
  `internal/settings/delete.go` для одноразового кода удаления аккаунта).
- ✅ Бинарники `server`, `worker` убраны из git-tracking, добавлены в `.gitignore`.

**Итого**: ~18–25 рабочих дней + ops-тест. Pre-prod sweep ✅ (выполнено 2026-04-27).

---

## Свобода реструктуризации репозитория

Игра не запущена, поэтому **любые переносы и переименования каталогов приветствуются**
если это делает архитектуру чище. Приоритет — современная и понятная структура,
а не совместимость с текущими путями.

Примеры допустимых изменений:
- `backend/internal/auth/` → `backend/pkg/auth/` (shared между сервисами)
- `frontend/src/` → `frontend/game/src/` (чтобы рядом лежал `frontend/portal/`)
- `backend/cmd/server/` → переименовать в `backend/cmd/game-server/` для ясности
- `migrations/` → `migrations/game/`, `migrations/auth/`, `migrations/portal/` (по сервисам)
- `projects/game-nova/api/openapi.yaml` → `api/openapi-game.yaml` + `api/openapi-auth.yaml` + `api/openapi-portal.yaml`

При реструктуризации обновляются: `Makefile`, `docker-compose*.yml`, импорты Go,
`tsconfig.json`, `vite.config.ts`, CI-конфиги.

---

## Что остаётся без изменений в игровой логике

- Все domain-пакеты (`fleet`, `battle`, `planet`, `economy`, …) — логика не трогается.
- Вся игровая бизнес-логика и миграции (кроме таблицы `users`).
- Worker — без изменений.
- Весь игровой frontend — только добавляется Universe Switcher в шапку.

---

## Ссылки

- [release-roadmap.md](../release-roadmap.md)
- [ops/vps-sizing.md](../ops/vps-sizing.md)
- [plans/07-payments.md](07-payments.md)
- [plans/32-multi-instance.md](32-multi-instance.md)
- [adr/](../adr/)
