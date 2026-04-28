# План 63: identity tokens по стандарту RFC 6749

**Дата**: 2026-04-28
**Статус**: ✅ Завершён 2026-04-28
**Зависимости**: нет блокирующих. Не зависит от плана 62 (исследование
ремастера). Можно делать параллельно с любыми другими планами.
**Связанные документы**:
- [51-identity-rename.md](51-identity-rename.md) — переименование auth-service
  → identity-service (где, вероятно, появился текущий нестандартный формат).
- [52-rbac-unification.md](52-rbac-unification.md) — RBAC unification.
- [53-admin-frontend.md](53-admin-frontend.md) — admin-bff (один из
  потребителей).

---

## Цель

Привести формат ответа identity-service на login/refresh к стандарту
**OAuth 2.0 / RFC 6749** (де-факто стандарт всей индустрии: Google,
GitHub, Auth0, Okta, Keycloak, Яндекс.OAuth, VK ID).

---

## Контекст

### Обнаружение проблемы

При работе агента над планом 62 (попытка залогиниться в admin-frontend
через identity → admin-bff) выяснилось, что **контракты не совпадают**:

- **identity сейчас отдаёт:**
  ```json
  {
    "user": { ... },
    "tokens": {
      "access": "eyJ...",
      "refresh": "..."
    }
  }
  ```

- **admin-bff ждёт:**
  ```json
  {
    "access_token": "eyJ...",
    "refresh_token": "..."
  }
  ```

Это не «admin-bff не дописан», а **самодельный формат identity**,
который не соответствует никакому стандарту. Любая интеграция с
будущими SDK / mobile-клиентами / CI-инструментами будет ломаться
ровно по этой причине.

### Что говорит стандарт

**RFC 6749 §5.1** — обязательный формат ответа на token endpoint:

```json
{
  "access_token": "eyJ...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "...",
  "scope": "..."
}
```

Поля:
- `access_token` (обязательное) — JWT.
- `token_type` (обязательное) — всегда `"Bearer"` для нашего случая.
- `expires_in` (обязательное на практике, рекомендуется RFC) —
  TTL access-токена в секундах.
- `refresh_token` (опциональное в RFC, но фактически нужно нам).
- `scope` (опциональное) — пока не используем, можно опустить.

**Поле `user` в OAuth-ответе НЕ возвращается** по стандарту.
Стандартный паттерн — отдельный endpoint `GET /userinfo` (OpenID
Connect) с access_token. Но для нашего UX-удобства допустимо
вернуть `user` рядом плоско:

```json
{
  "access_token": "...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "...",
  "user": {
    "id": "...",
    "username": "...",
    "email": "..."
  }
}
```

Это не строго RFC, но **широко принятая практика** для login
endpoint'ов SPA (см. ответы Auth0 `/oauth/token` с `id_token`,
ответы Firebase Auth и т.д.). Главное — **плоская структура**
для токенов, без вложенного `tokens: {access, refresh}`.

---

## Что меняем

### 1. Identity-service — handler выдачи токенов

`projects/identity/backend/internal/identitysvc/` (login + refresh):

**Было** (примерно):
```go
type LoginResponse struct {
    User   UserDTO   `json:"user"`
    Tokens TokensDTO `json:"tokens"`
}
type TokensDTO struct {
    Access  string `json:"access"`
    Refresh string `json:"refresh"`
}
```

**Станет:**
```go
type LoginResponse struct {
    AccessToken  string  `json:"access_token"`
    TokenType    string  `json:"token_type"`        // всегда "Bearer"
    ExpiresIn    int     `json:"expires_in"`        // TTL access в секундах
    RefreshToken string  `json:"refresh_token"`
    User         UserDTO `json:"user"`              // удобство SPA
}
```

То же для `RefreshResponse` (на endpoint refresh — `user` опционально,
можно не включать; но для единообразия можно включить).

`expires_in` берётся из существующей конфигурации TTL JWT (которая
точно есть, иначе токены бы не работали). Если переменная
называется как-то иначе — переименовать в `AccessTokenTTL`.

### 2. Потребители — переименование полей

Грепы по всем фронтам и сервисам (поимённо найти все обращения):

```bash
grep -rn '\.tokens\.access\|\.tokens\.refresh\|"tokens"' \
  projects/ --include="*.ts" --include="*.tsx" --include="*.js" \
  --include="*.go" 2>/dev/null
```

**Ожидаемые места:**
- `projects/portal/frontend/src/api/auth/*` — login/refresh клиент.
- `projects/portal/frontend/src/stores/auth*` — Zustand store.
- `projects/game-nova/frontend/src/api/auth/*` — то же.
- `projects/game-nova/frontend/src/stores/auth*` — то же.
- `projects/admin-bff/internal/identityclient/*` — клиент на стороне
  bff. Если ждал `access_token` — оставить, добавить парсинг
  `expires_in`.
- `projects/billing/...`, `projects/portal/backend/...` — если они
  верифицируют JWKS, контракт токена самого не меняется (только
  обёртка), но проверить.
- `projects/game-origin/src/core/JwtAuth.php` — если PHP-сторона
  парсит wrapper'ы login-ответа, тоже править.

### 3. Тесты

- В identity: handler-тесты на новый формат.
- На фронтах: integration-тесты login (если есть). E2E-тесты в
  `tests/e2e/` — переименование полей в моках.

### 4. Смоук

- `docker-compose up identity portal-backend portal-frontend
  game-nova-backend game-nova-frontend admin-bff admin-frontend`.
- Залогиниться через каждый фронт, убедиться что:
  - токены приходят в новом формате;
  - сохраняются в Zustand;
  - access используется в Authorization: Bearer;
  - refresh-flow работает (после истечения access — автоматический
    refresh, не редирект на логин).

---

## Чего НЕ делаем

- Не меняем JWT-claims внутри токена (issuer, audience, exp, iat,
  scope, sub) — это другой контракт, он остаётся.
- Не меняем JWKS endpoint (`/.well-known/jwks.json`) — он стандартный.
- Не вводим OpenID Connect (отдельный `/userinfo`, ID token) — пока
  не нужно, можем добавить позже.
- Не вводим `scope` в ответ — пока единый scope, не используем.
- Не трогаем backend-логику авторизации (RBAC, проверка ролей) —
  только формат wrapper'а.

---

## Этапы

### Ф.1. Identity-handler

- Найти текущий `LoginResponse` / `RefreshResponse` структуры.
- Переименовать поля по RFC + добавить `token_type`, `expires_in`.
- Обновить openapi.yaml identity (если есть — у identity своя
  спецификация в `projects/identity/backend/api/`).
- Тесты handler'ов (минимум: login возвращает все 5 полей,
  TokenType="Bearer", ExpiresIn>0).
- Запустить `go build ./...` + `go test ./...` в identity.

### Ф.2. Грепы по потребителям

- Полный grep по `\.tokens\.access`, `\.tokens\.refresh`, `"tokens"`
  в `projects/`. Зафиксировать список затронутых файлов в
  коммит-сообщении.

### Ф.3. Фронт-потребители

- portal-frontend, game-nova-frontend: переименование полей в
  api-клиентах + auth-store. Если есть TypeScript-типы — обновить.
- Смоук в браузере: логин работает.

### Ф.4. Backend-потребители

- admin-bff: проверить identityclient, добавить парсинг
  `expires_in`/`token_type` если нужно. Обновить mock в proxy_test.go
  если есть.
- portal-backend, game-nova-backend (middleware'ы) — обычно они
  валидируют JWT через JWKS, не парсят wrapper. Но проверить, что
  ничего не сломалось.
- mail-service, billing-* — то же.

### Ф.5. PHP-сторона (game-origin)

- `projects/game-origin/src/core/JwtAuth.php` — если парсит
  login-wrapper, обновить.

### Ф.6. E2E + Smoke

- Запустить весь стек через docker-compose.
- Залогиниться через portal, game-nova, admin-frontend (через bff).
- Проверить refresh-flow.

### Ф.7. Финализация

- Обновить шапку плана 63 — статус ✅.
- Запись в `docs/project-creation.txt` — итерация 63.
- Коммит: `refactor(identity): tokens по стандарту RFC 6749 (план 63)`.

---

## Тестирование

- Unit-тесты identity handler'ов на новый формат.
- E2E-логин в каждом фронте.
- Refresh-flow: дождаться истечения access (или подменить TTL на
  1 минуту в dev), убедиться что фронт авто-рефрешит.

---

## Объём

- ~100-200 строк изменений
- ~5-15 файлов (1 в identity, по 2-3 на фронт, по 1 на backend-сервис)
- 1-2 коммита (можно в одном, если небольшие правки; можно разбить
  на «identity» + «потребители»)

Время выполнения: **~2-4 часа агента** в активном темпе.

---

## Когда запускать

Можно делать **параллельно с планом 62** — они не пересекаются по
файлам.

Не имеет смысла **до** плана 51 (identity-rename) — но 51 уже
закрыт.

Не блокирует ничего. Но желательно **до публичного запуска**, чтобы
не пришлось ломать контракт уже поднятым клиентам.

---

## Известные риски

| Риск | Митигация |
|---|---|
| Сломали login во всех фронтах одним коммитом, на dev-стенде ничего не работает | Один коммит — одна сторона (либо identity + клиенты вместе, либо разнесённые но проверенные оба) |
| Пропустили потребителя в грепе | Полный grep по `.tokens.access`, `.tokens.refresh`, `"tokens"` — фиксировать список затронутых файлов в коммит-сообщении для проверки |
| `expires_in` потребителями не парсится → нет авто-refresh после истечения | Тест refresh-flow в Ф.6 — обязательный |
| game-origin PHP-сторона незаметно сломается | grep по `tokens` в PHP, явная проверка JwtAuth.php в Ф.5 |
| Конфликт с параллельной сессией плана 62 | План 62 не трогает identity/admin-bff/фронты. Если конфликт — разрешать по правилу «свои файлы поимённо» из CLAUDE.md |

---

## Что после плана 63

- Identity отдаёт RFC 6749-совместимые ответы.
- Admin-frontend логин работает (через admin-bff → identity).
- Готовность к публичному запуску — внешние клиенты (mobile SDK,
  тестовые инструменты, OAuth-аналитика) сразу совместимы.
- Не нужны hacky-адаптеры в backend-сервисах.

---

## References

- [RFC 6749 §5.1](https://datatracker.ietf.org/doc/html/rfc6749#section-5.1)
  — Successful Response (точный формат полей).
- [OpenID Connect Core 1.0](https://openid.net/specs/openid-connect-core-1_0.html)
  — для будущего расширения (ID Token, /userinfo).
- Реализации в индустрии: Auth0 `/oauth/token`, Firebase Auth,
  Yandex OAuth, VK ID.
