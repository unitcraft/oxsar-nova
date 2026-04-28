# Промпт: выполнить план 63 (identity tokens по RFC 6749)

**Дата создания**: 2026-04-28
**План**: [docs/plans/63-identity-tokens-rfc6749.md](../plans/63-identity-tokens-rfc6749.md)
**Применение**: вставить блок ниже в новую сессию Claude Code в
рабочей директории `d:\Sources\oxsar-nova`. Агент прочитает план 63
самостоятельно и выполнит миграцию формата токенов identity на
стандарт RFC 6749 + переименование у всех потребителей.
**Объём**: 1-2 коммита, ~100-200 строк, 5-15 файлов, 2-4 часа.

---

```
Задача: выполнить план 63 — миграция identity-service на формат
токенов по стандарту RFC 6749, переименование у всех потребителей.

ВАЖНОЕ:
- Это рефакторинг контракта между identity и его потребителями.
  Затрагивает несколько сервисов и фронтов, нужна аккуратность.
- В проде ничего нет, ломать существующие подписки можно свободно.
- Параллельно работают другие сессии (план 62 — research,
  не пересекается по файлам). Будь аккуратен с git.

ПЕРЕД НАЧАЛОМ:

1) git status --short — если есть чужие изменения от параллельных
   сессий, спроси пользователя. Бери только свои файлы.

2) Прочитай ПОЛНОСТЬЮ:
   - docs/plans/63-identity-tokens-rfc6749.md (твоё ТЗ —
     соблюдай каждую фазу Ф.1-Ф.7)
   - CLAUDE.md (правила: коммиты, языки, hooks, ctx, поимённый
     git add)

3) Прочитай выборочно по мере необходимости:
   - docs/plans/51-identity-rename.md (контекст: где появился
     текущий формат)
   - docs/plans/52-rbac-unification.md (RBAC, user-фактуры)
   - docs/plans/53-admin-frontend.md (admin-bff — один из
     потребителей)
   - projects/identity/backend/internal/identitysvc/ (текущие
     handler'ы login + refresh — то, что меняешь)
   - projects/identity/backend/api/openapi.yaml (если есть —
     обновить спеку)

ЧТО НУЖНО СДЕЛАТЬ (по фазам плана 63):

Ф.1. Identity-handler — переход на RFC 6749

Найти текущий LoginResponse / RefreshResponse:
- Было: { user, tokens: { access, refresh } }
- Станет: { access_token, token_type:"Bearer", expires_in,
           refresh_token, user (опционально для UX SPA) }

token_type — всегда "Bearer".
expires_in — TTL access-токена в секундах, берётся из существующей
конфигурации JWT TTL (она точно есть, иначе токены бы не работали).

Файлы (примерно — найди точно):
- projects/identity/backend/internal/identitysvc/login.go (или
  handler.go — где LoginResponse)
- projects/identity/backend/internal/identitysvc/refresh.go (или
  тот же файл)
- projects/identity/backend/api/openapi.yaml (если есть — обновить
  схемы LoginResponse / RefreshResponse)
- Тесты handler'ов (если есть — обновить ожидаемые поля)

Проверка:
- go build ./... в projects/identity/ — компилируется
- go test ./... в projects/identity/ — все зелёные
- go vet ./... — чисто

Ф.2. Грепы по потребителям

Полный grep по всем местам, где парсят wrapper:

  grep -rn '\.tokens\.access\|\.tokens\.refresh\|"tokens"' \
    projects/ --include="*.ts" --include="*.tsx" --include="*.js" \
    --include="*.go" --include="*.php" 2>/dev/null

Зафиксируй полный список затронутых файлов в коммит-сообщении —
это критично для проверки что ничего не пропущено.

Ф.3. Фронт-потребители

- projects/portal/frontend/ — api/auth/* + stores/auth*
- projects/game-nova/frontend/ — api/auth/* + stores/auth*
- В каждом: переименовать data.tokens.access → data.access_token,
  data.tokens.refresh → data.refresh_token. Обновить TS-типы.
- Если используется TanStack Query — проверить, что queryFn
  возвращает плоскую структуру.
- Опционально: использовать expires_in для проактивного refresh
  (за минуту до истечения) — если уже есть refresh-логика,
  улучшить её; если нет — отложить.

Smoke в браузере:
- docker-compose up — портал + game-nova
- Залогиниться через portal AuthPage
- Залогиниться через game-nova LoginScreen
- Проверить, что токены сохраняются в Zustand и используются
  в Authorization: Bearer

Ф.4. Backend-потребители

- projects/admin-bff/internal/identityclient/ — он УЖЕ ждёт плоский
  формат (это его ожидание было правильным и обнаружило проблему).
  Проверить, что:
  · парсит access_token / refresh_token (вероятно уже да)
  · читает expires_in / token_type (добавить, если нет)
  · refresh-flow работает с новым ответом
- projects/admin-bff/internal/proxy/proxy_test.go — обновить mock
  identity-ответа на новый формат.
- projects/portal/backend/, projects/game-nova/backend/ — обычно
  валидируют JWT через JWKS, не парсят wrapper. Но проверить
  middleware'ы login/refresh proxy если они есть.
- projects/billing/, projects/mail/ (если есть login через
  identity) — то же.

Ф.5. PHP-сторона (game-origin)

- projects/game-origin-php/src/core/JwtAuth.php — если парсит
  login-wrapper (handoff из portal с токенами в URL?), обновить.
- Найти все места обращения к identity API в game-origin.
- НЕ ЛОМАТЬ: handoff-механизм должен продолжать работать.

Ф.6. E2E + Smoke

- docker-compose up весь стек (если есть единый scripts/dev-up.sh
  или подобное).
- Залогиниться через каждый фронт: portal, game-nova,
  admin-frontend (через admin-bff).
- Проверить refresh-flow:
  · Дождаться истечения access (или подменить TTL на 1 минуту в
    dev) → фронт автоматически рефрешит, не редиректит на логин
  · Если refresh-flow ломается — это критично, сразу чинить
- game-origin handoff из portal — проверить, что не сломали.

Ф.7. Финализация

1. Обновить шапку плана 63 — статус "Завершён <дата>".
2. Запись в docs/project-creation.txt — итерация 63.
3. Финальный коммит: refactor(identity): tokens по RFC 6749 (план 63).

ВАЖНОЕ: GIT-ИЗОЛЯЦИЯ ОТ ПАРАЛЛЕЛЬНЫХ СЕССИЙ:

Параллельно идёт большая сессия плана 62 в
docs/research/origin-vs-nova/ + docs/plans/62-... +
docs/release-roadmap.md. Не пересекайтесь.

- git status --short ПЕРЕД каждым git add
- git add ТОЛЬКО конкретные пути (свои):
  · projects/identity/...
  · projects/portal/frontend/... + projects/portal/backend/...
  · projects/game-nova/frontend/... + projects/game-nova/backend/...
  · projects/admin-bff/...
  · projects/billing/... (если правил)
  · projects/mail/... (если правил)
  · projects/game-origin-php/src/core/JwtAuth.php (если правил)
  · docs/plans/63-identity-tokens-rfc6749.md
  · docs/project-creation.txt
- git status --short ПЕРЕД каждым git commit (убедись что не
  захватил чужое из docs/research/...)
- Если случайно staged чужой файл: git reset HEAD -- путь
- НИКОГДА не делай git add . или git add -A

КОММИТЫ:

Можно делать в один большой коммит ИЛИ разбить на два:
- Вариант A (один коммит):
  refactor(identity): tokens по RFC 6749 + потребители (план 63)
- Вариант B (два коммита):
  · refactor(identity): tokens по RFC 6749 (план 63 Ф.1)
  · refactor(consumers): синхронизация с RFC 6749 identity (план 63 Ф.2-Ф.5)

Вариант B предпочтительнее — проще ревью и rollback. Коммитить
поочерёдно: сначала identity, проверить локально что фронты
сломались (это ожидаемо), потом коммит с консьюмерами →
проверить что всё снова работает.

Каждый коммит:
- conventional: refactor(...): ... (план 63)
- В теле сообщения упомянуть полный список затронутых файлов
  (для проверки что ничего не пропущено)
- trailer: Generated-with: Claude Code (НЕ Co-Authored-By —
  git hook уберёт автоматически, см. CLAUDE.md Onboarding)

ЧЕГО НЕ ДЕЛАТЬ:

- Не менять JWT-claims внутри токена (issuer, audience, exp, iat,
  scope, sub) — это другой контракт, остаётся как есть.
- Не менять JWKS endpoint (/.well-known/jwks.json) — стандартный.
- Не вводить полный OpenID Connect (отдельный /userinfo, ID Token) —
  пока избыточно, добавим позже если понадобится.
- Не вводить scope в ответ — пока единый scope, не используем.
- Не трогать backend-логику авторизации (RBAC, проверка ролей) —
  только формат wrapper'а login/refresh.
- Не делать "пока я тут, заодно подправлю" — рефакторинг по плану
  63 и всё.
- Не писать "возможно", "вероятно" — каждое решение должно быть
  обосновано планом 63 или RFC 6749.

ОЦЕНКА ОБЪЁМА:

2-4 часа работы агента в активном темпе. Если сильно дольше —
проверь, не уехал ли в "пока я тут". Если сильно быстрее —
проверь полноту грепов (Ф.2): пропущенный потребитель = баг.

ИЗВЕСТНЫЕ РИСКИ (см. план 63, секция "Известные риски"):

- Сломали login во всех фронтах одним коммитом → используй
  Вариант B коммитов (раздельно identity → консьюмеры).
- Пропустили потребителя → полный grep + список в коммит-сообщении.
- expires_in не парсится потребителями → тест refresh-flow в Ф.6
  обязательный.
- game-origin PHP незаметно сломается → Ф.5 явная проверка
  JwtAuth.php.

УСПЕШНЫЙ ИСХОД:

- Identity отдаёт RFC 6749-совместимый ответ login/refresh:
  { access_token, token_type:"Bearer", expires_in, refresh_token, user }
- Все фронты логинятся, токены сохраняются в Zustand.
- admin-frontend через admin-bff логинится — этот сценарий был
  сломан и обнаружен агентом плана 62, после плана 63 работает.
- Refresh-flow работает (авто-refresh за минуту до истечения).
- game-origin handoff из portal продолжает работать.
- 1-2 коммита, шапка плана 63 помечена ✅.
- Готовность к публичному запуску — внешние клиенты (mobile SDK,
  тестовые инструменты) сразу совместимы со стандартом.

Стартуй.
```
