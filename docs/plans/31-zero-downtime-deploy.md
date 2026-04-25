---
title: 31 — Безболезненный deploy + feature flags
date: 2026-04-26
status: in-progress
---

> **Статус 2026-04-26**:
> - Ф.1 (health/ready/draining) — ✅ done.
> - Ф.2 (feature flags) — pending.
> - Ф.3 (API-version header) — pending.
> - Ф.4 (документация) — pending.
> - Ф.5 (blue-green nginx) — отложено до prod-load.

# План 31: Безболезненный deploy + feature flags

**Цель**: дать возможность катить большие рефакторинги (Goal Engine,
будущие планы) **без риска** для игроков. Два кита: feature flags
(включить/выключить новый код одной правкой) + zero-downtime
выкатка (без 502-ошибок при деплое).

**Контекст**: проект на стадии раннего prod, один разработчик. Не
нужен сложный canary/blue-green — нужно **минимум** для того, чтобы
рефакторинги катились без простоя.

---

## 1. Что есть сейчас

| Компонент | Состояние |
|---|---|
| Backend graceful shutdown | ✅ есть (`server/main.go:482` — SIGTERM + 30s timeout) |
| Worker graceful shutdown | ✅ предположительно (наследует pattern) |
| nginx (frontend контейнер) | ✅ проксирует /api → backend:8080 |
| Frontend Vite build | ✅ хеши в filenames, кеш `immutable` |
| Migrations (goose) | ✅ 0001-0064, sequential |
| Health-check endpoint | ❌ нет |
| Feature flags | ❌ нет |
| API-version header | ❌ нет |
| Документация «как катить» | ❌ нет |

## 2. Что плохо при текущем deploy

### Сценарий A: backend-only update
1. Сборка нового образа `backend`.
2. `docker compose up -d backend` — старый контейнер останавливается, новый стартует.
3. **Проблема**: между остановкой старого и health-ready новым — несколько секунд 502 для входящих запросов.
4. Игроки в этот момент видят падение страницы.

### Сценарий B: миграция БД + backend update
1. `goose up` — добавление колонки.
2. **Если `ALTER TABLE` локирует** → запросы зависают.
3. Backend старый ещё работает, не знает про новую колонку → SELECT * падает.

### Сценарий C: breaking-change в API
1. Backend выкатывает новую версию с измененным response-форматом.
2. Frontend (старая версия в кеше браузера) парсит ответ → ошибки.
3. Игрок видит белый экран, F5 не помогает (cached JS).

### Сценарий D: внезапный баг в новой фиче
1. Выкатили рефакторинг, через час замечен баг.
2. Откат = передеплой старого образа = снова 502.
3. Если миграция БД — откат сложнее.

## 3. Что планируется

Минимальный набор инструментов, который покрывает все 4 сценария.

### 3.1 Health-check endpoint

`GET /api/health` — простой liveness check (без auth):
- 200 OK + `{"status": "ok", "version": "...", "uptime_sec": N}` — backend готов принимать запросы.
- 503 + JSON — backend в shutdown state (отказывается от новых соединений).

`GET /api/ready` — readiness:
- 200 — БД соединение живо, миграции применены, кэши прогреты.
- 503 — что-то не готово.

Используется nginx для решения «слать ли запросы на этот upstream».

### 3.2 Graceful shutdown (расширение текущего)

Сейчас:
```go
shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
return srv.Shutdown(shutdownCtx)
```

**Добавить**:
1. **Pre-shutdown delay** — после получения SIGTERM, **до** `srv.Shutdown()`, выставить `health=draining` (новый endpoint возвращает 503), подождать 5-10 секунд, чтобы nginx убрал backend из upstream.
2. **drain notify** — логировать «draining: N requests in-flight» для отладки.
3. **Worker graceful** — то же самое: дать текущему event-handler закончиться, не брать новые из очереди (`SELECT FOR UPDATE SKIP LOCKED`).

```go
// В server/main.go
state := &shutdownState{}

http.HandleFunc("/api/health", func(w, r) {
    if state.Draining() { w.WriteHeader(503); return }
    w.WriteHeader(200)
})

go func() {
    <-ctx.Done()  // SIGTERM
    state.SetDraining()
    log.Info("draining for 10s before shutdown")
    time.Sleep(10 * time.Second)
    srv.Shutdown(shutdownCtx)
}()
```

### 3.3 Feature flags

**Простой YAML + Go-struct**, без внешних сервисов.

`configs/features.yaml`:
```yaml
features:
  goal_engine:
    enabled: false
    description: "Новый движок целей (план 30). false = старый achievement+dailyquest"
  goal_engine_writes:
    enabled: false
    description: "Записывать данные в новые таблицы goals (для shadow-режима)"
  new_battle_formula:
    enabled: false
    description: "..."
```

```go
// backend/internal/features/features.go
type Flags struct {
    GoalEngine        bool
    GoalEngineWrites  bool
    NewBattleFormula  bool
}

func Load(path string) (*Flags, error) { /* parse YAML */ }

func (f *Flags) Enabled(key string) bool {
    switch key {
    case "goal_engine": return f.GoalEngine
    // ...
    }
}
```

Использование в коде:
```go
if features.Enabled("goal_engine") {
    return goalEngine.Handle(ctx, ...)
}
return oldAchievementSvc.Handle(ctx, ...)
```

**Изменение flag = restart backend** (это OK, restart дешёвый при graceful shutdown). Не делаем hot-reload — это сложнее без выгоды.

### 3.4 Online-friendly миграции — правило

Документировать в **CLAUDE.md** или `docs/ops/migrations-style.md`:

```markdown
## Online-friendly migrations

При написании миграции — **не блокировать таблицу** на длительные
операции. Правила:

1. **ADD COLUMN** — всегда `NULL` по умолчанию, без `DEFAULT` для
   больших таблиц. После — backfill отдельной миграцией с `LIMIT N` в loop.

2. **DROP COLUMN** — двухфазная:
   - Migration 1: backend перестаёт читать колонку (deploy).
   - Migration 2: ALTER TABLE DROP COLUMN.

3. **CREATE INDEX** — `CONCURRENTLY` (если не первая миграция в файле).

4. **ALTER TABLE ... ALTER COLUMN TYPE** — крайне опасно на больших
   таблицах. Лучше: добавить новую колонку, backfill, переключить
   код, drop старую.

5. **Foreign keys NOT VALID + VALIDATE** — добавлять FK без проверки,
   потом валидировать отдельно (без блока).

6. **Seed-INSERT** — батч `INSERT ... ON CONFLICT DO NOTHING`.

Для каждой миграции — комментарий, **сколько секунд занимает** на
текущем размере данных (10k users, 100k planets, 1M events).
```

### 3.5 API-version header

Backend в каждом ответе:
```
X-Api-Version: 2026.04.26
X-Min-Client-Version: 2026.04.20
```

Frontend сравнивает `X-Min-Client-Version` со своей встроенной версией
(в `package.json` или `vite.config`). Если frontend старше → показать
toast «Вышла новая версия, обновите страницу (F5)».

Backend **не должен** ломать old-clients жёстко — старые поля
оставлять, новые добавлять. Если поле семантически сломалось — bump
`X-Min-Client-Version` и в коде backend проверять header `X-Client-Version`
запроса; если ниже — возвращать 426 Upgrade Required.

### 3.6 Документация: deploy.md

`docs/ops/deploy.md`:

```markdown
# Deploy procedure

## Малый patch (нет миграций, нет breaking-change)

1. `git push origin main`
2. На сервере: `git pull && docker compose -f deploy/docker-compose.prod.yml up -d --build backend worker`
3. Проверить `curl https://prod/api/health` → 200.
4. Если 502 на 5+ секунд — что-то не так, откат.

## С миграцией БД

1. **На staging**: применить миграцию, проверить что код работает.
2. На prod: `goose up -dir migrations` (отдельным шагом, до backend deploy).
3. Backend deploy.
4. Проверить.

## С feature flag (рекомендуется для рефакторингов)

1. Catйge feature flag = false. Catедер код мёртвый, никаких эффектов.
2. Включить flag для себя: `features.yaml: { goal_engine: true }`.
3. Тестировать в проде на своём аккаунте.
4. Раскатить: `enabled: true`, restart backend.
5. Если баг: `enabled: false`, restart. Откат за 30 секунд.

## Откат

- Code: `git revert + redeploy` или предыдущий image tag.
- БД: миграция down (если возможно). Если down невозможна — backfill
  данных вручную.
```

## 4. План внедрения

### Фаза 1: health/ready endpoints + draining ✅ DONE

**Затронуто**: `backend/internal/health/` (новый пакет),
`backend/cmd/server/main.go`, `backend/cmd/worker/main.go`,
`backend/Dockerfile`, `deploy/docker-compose.yml`.

- [x] Создан пакет `backend/internal/health/` с типом `State`
      (atomic-флаги ready/draining), методами `SetReady`/`SetDraining`,
      handlers `HealthHandler` и `ReadyHandler(Pinger)`.
- [x] Endpoint `GET /api/health` (liveness): 200 OK при normal,
      503 при draining. Не делает БД-вызовов.
- [x] Endpoint `GET /api/ready` (readiness): 200 OK когда `SetReady`
      + БД доступна; 503 при draining/starting/db_unhealthy.
- [x] Server: при SIGTERM `state.SetDraining()` → sleep 10s
      (drainDelay) → `srv.Shutdown()`. nginx за это время убирает
      backend из upstream по health-check failure.
- [x] Worker: публикует `/api/health` и `/api/ready` на metrics-порту
      :9091 (вместе с /metrics). Goroutine ждёт SIGTERM и вызывает
      `SetDraining()`.
- [x] Dockerfile: `apk add wget` для healthcheck.
- [x] docker-compose: healthcheck на /api/ready для backend (через
      `:8080`) и worker (через `:9091`). frontend.depends_on
      переключён на `service_healthy`.
- [x] `var buildVersion = "dev"` в обоих main.go — перебивается через
      `go build -ldflags "-X main.buildVersion=..."` в build-pipeline.
- [x] Unit-тесты: 6 тестов в `health_test.go` (все случаи: ok,
      draining, not_ready, db_down, draining-precedence, idempotency).

**Tests**: `go test ./... -count=1`: 27 пакетов зелёные.

### Фаза 2: feature flags

**Затрагивает**: новый пакет `backend/internal/features/`,
`configs/features.yaml`.

- [ ] `features.Load(path)` — парсит YAML.
- [ ] Прокидывание `*Flags` в server и worker (через config-loader).
- [ ] Метод `flags.Enabled(key string) bool`.
- [ ] Endpoint `GET /api/features` (для UI: какие фичи включены, чтобы
      рисовать UI conditionally).
- [ ] Документация в comment'ах: «как добавить новый флаг».

**Готовность**: 1 PR, ~150 строк, 0 миграций.

### Фаза 3: API-version header

**Затрагивает**: `backend/internal/httpx/middleware.go` (если есть, или
создать), `frontend/src/api/client.ts`.

- [ ] Backend middleware добавляет `X-Api-Version` в каждый response.
- [ ] Версия из `runtime.BuildInfo()` или env `APP_VERSION`.
- [ ] Frontend читает header, сравнивает со встроенной версией, при
      рассогласовании — toast «обновите страницу».
- [ ] При 426 ответе — full-page toast «обязательное обновление».

**Готовность**: 1 PR, ~60 строк.

### Фаза 4: документация

- [ ] `docs/ops/deploy.md` — процедуры выкатки.
- [ ] `docs/ops/migrations-style.md` — online-friendly правила для
      миграций.
- [ ] Ссылки в `CLAUDE.md` на эти доки.

**Готовность**: 1 PR, чистая документация.

### Фаза 5 (опц.): blue-green через nginx

**Только когда понадобится**. Сейчас при graceful shutdown 10s простой
~10s — допустимо для проекта на стадии MVP/early-prod. Блю-грин — отдельная
задача, когда количество игроков того потребует.

## 5. Приоритеты

**Минимально-нужное для плана 30 (Goal Engine)**:
- Ф.1: health/draining (чтобы при выкатке Goal Engine не было 502).
- Ф.2: feature flags (чтобы можно было включить/выключить флагом).

**Документация (Ф.4)** — параллельно по ходу написания других
планов.

**Ф.3 и Ф.5** — позже, по мере роста проекта.

## 6. Что НЕ делаем

- **Не вводим** Kubernetes/Helm. Один docker-compose, простота.
- **Не делаем** автоматический rollback (CI/CD фича — overkill).
- **Не пишем** свой feature-flag сервис (LaunchDarkly-style). YAML
  достаточно.
- **Не реализуем** session sticky (запросы любого юзера могут идти
  на любой backend instance — у нас и так один instance).
- **Не вводим** distributed tracing — один backend, slog хватает.
- **Не делаем** Canary deploy (% юзеров на новом коде). Feature flag
  + отдельный аккаунт для тестов покрывает 95% случаев.

## 7. Связь с другими планами

- **План 30 (Goal Engine)** — главный потребитель этого плана. Без
  feature flags + health/drain рефакторинг рискованный.
- **Плановые миграции** — все будущие миграции должны следовать
  правилам из Ф.4.
- **План 25** (credits-economy) — если потребуются breaking-change
  API, поможет header-based versioning.

## 8. Открытые вопросы

1. **`APP_VERSION` где брать?** Из git tag, build flag, env var?
2. **draining 10s** — оптимальное число? Можно сделать настраиваемым через
   env.
3. **Feature flags per-user**? (например, A/B тестирование). Сейчас —
   глобальные. Если потребуется — добавить в Ф.2 hash userID для
   детерминированного бакета.
4. **Hot-reload features.yaml** — нужен ли? Restart backend дёшев,
   но hot-reload удобнее. Решение: добавить, если pain-point
   проявится.
