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

Добавление вселенной = новый `docker-compose.uniN.yml` + строка в `configs/universes.yaml`.  
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
- Global credits и платёжные webhooks — сквозные, не принадлежат ни одной вселенной.

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

### API Auth Service

```
POST /auth/register           — регистрация (email + username + password)
POST /auth/login              — вход по паролю
POST /auth/refresh            — обновить access token
POST /auth/logout             — отозвать refresh token
GET  /auth/me                 — профиль текущего пользователя
GET  /auth/universes          — публичный список вселенных (из configs/universes.yaml)

GET  /auth/oauth/{provider}           — редирект к провайдеру
GET  /auth/oauth/{provider}/callback  — приём кода, выдача JWT

POST /auth/universe-token     — one-time handoff token для переключения вселенной
POST /auth/token/exchange     — обмен one-time code → JWT (после OAuth redirect)

GET  /auth/credits/balance    — актуальный баланс global credits
POST /auth/credits/spend      — списать кредиты (внутренний, вызывается Portal Backend)
POST /auth/payments/{provider}/webhook — webhook от платёжного шлюза

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
6. **Платёжные webhooks** переезжают в Auth Service (Ф.7).

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

## Global Credits

Баланс живёт **только в Auth Service** (`users.global_credits`).  
Каждое движение — запись в `credit_transactions` (полный аудит).

**Пополнение**: платёжный шлюз → `POST /auth/payments/{provider}/webhook` → зачисление.

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

1. `configs/universes.yaml` — настройки всех вселенных.
2. Auth Service: `GET /auth/universes`.
3. Portal Backend: `GET /api/universes` + агрегация онлайна из Redis вселенных.

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

### Ф.7 — Global Credits + платёжный шлюз (2–3 дня)

1. Auth Service: `spend`, `balance` endpoints.
2. Перенести webhooks (CloudPayments, Enot.io) из игрового сервера → Auth Service.
3. Интеграция голосования за feedback с `credits/spend`.
4. Баланс кредитов в шапке портала и игры.

### Ф.8 — Запуск двух вселенных (2–3 дня)

Запуск сразу с двумя вселенными. Вторая — существенно быстрее по скорости (×N),
остальные настройки (bash-лимиты, экономика) уточняются перед деплоем.

1. `configs/universes.yaml` — финальные настройки обеих вселенных (имена, скорость).
2. `docker-compose.uni01.yml` + `docker-compose.uni02.yml`.
3. Деплой обоих стеков + smoke-test каждого.
4. Переключение между вселенными — проверить handoff-флоу end-to-end.
5. DNS: делегировать зону на Selectel, получить wildcard SSL.

### Ф.9 — Реструктуризация репозитория (½ дня) ✅ 2026-04-26

Привести структуру папок к модели «один домен — одна папка»:

```
portal/
  frontend/     ← было frontend/portal/
  backend/      ← будет (backend/cmd/portal + internal/portalsvc + migrations/portal/)
auth/
  backend/      ← будет (backend/cmd/auth-service + internal/authsvc + migrations/auth/)
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

### Ops: нагрузочный тест (отдельный тикет, ~1–2 дня)

После Ф.8, до публичного анонса. Обновить [ops/vps-sizing.md](../ops/vps-sizing.md)
с учётом новой архитектуры (несколько стеков). Тестировать:
- Каждый game-server независимо (как сейчас).
- Auth Service под пиковой нагрузкой (все вселенные логинятся одновременно).
- Всю связку: портал → auth → две вселенных одновременно.

**Итого**: ~18–25 рабочих дней + ops-тест.

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
- `api/openapi.yaml` → `api/openapi-game.yaml` + `api/openapi-auth.yaml` + `api/openapi-portal.yaml`

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
