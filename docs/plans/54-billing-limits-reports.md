# План 54: Billing limits + admin-отчёты

**Дата**: 2026-04-27
**Статус**: Возобновлён 2026-04-27, **только backend-фазы**.
UI-часть («Пополнить кредиты» → «Пополнить оксары») отложена и
выполняется в составе плана 58 Ф.5b (i18n-корректировка).
**Зависимости**: **План 51** (rename), **План 52** (RBAC), **План 53**
(admin-frontend, backend-часть нужна для admin-bff RBAC) должны быть
выполнены или готовы к моменту своих операций.
**Связь с планом 58** (ADR-0009): backend плана 54 работает с суммами
**в рублях** (лимит ФНС 2.4 млн ₽/год), не зависит от названия валюты.
В коде допустимо использовать существующее имя `credit` — оно будет
переименовано на `oxsar` в составе плана 58. Frontend «Пополнить кредиты»
не трогаем — в плане 58 Ф.5b будет полная i18n-замена.
**Связанные документы**: [38-billing-service.md](38-billing-service.md),
[42-yookassa.md](42-yookassa.md), [52-rbac-unification.md](52-rbac-unification.md),
[53-admin-frontend.md](53-admin-frontend.md).

---

## Зачем

### Лимит самозанятого (НПД, ФЗ-422)

Самозанятый в РФ имеет годовой лимит дохода **2 400 000 ₽**. При
превышении — статус НПД автоматически снимается, и доход облагается
13% НДФЛ как у физлица или по правилам ИП-УСН (если статус ИП есть).
Нарушение лимита — это юридический и налоговый риск.

Чтобы НЕ перелимитить, нужна автоматическая защита:
- Жёсткий потолок отключения **2 300 000 ₽** (буфер 100 000 ₽ от
  лимита ФНС, ~4% запас на edge-cases в платежах в полёте).
- Soft-warning уровни: 80%, 90%, 95% — для предупреждения админа
  заранее.

### Admin-отчёты

В план 38 (billing-service) уже реализована основа: orders, payments,
refunds, providers (mock, yookassa). Но **нет UI для просмотра**:
- Сколько заработано за период (день/неделя/месяц/год).
- Список транзакций с фильтрами (status, provider, юзер).
- Возвраты и chargebacks.
- Reconciliation status.
- Export для бухгалтерии.

## Архитектура

### Лимит — application logic

**Источник правды для лимита** — таблица `billing.payments` в
billing-service (а не provider-side). Считаем **net revenue**:

```
revenue_ytd = SUM(amount_kop) WHERE status='succeeded' AND year(captured_at)=current
            - SUM(amount_kop) WHERE status='refunded' AND year(refunded_at)=current
```

(Период — календарный год по МСК, как считает ФНС.)

### Конфигурация лимитов

Через **ENV** (production-grade, перезагружается при restart):

```bash
SELF_EMPLOYED_ANNUAL_LIMIT_KOP=240000000  # 2.4M ₽ — лимит ФНС
HARD_STOP_THRESHOLD_KOP=230000000          # 2.3M ₽ — отключение пополнения
WARN_THRESHOLD_PERCENT_80=true             # email при 80% (1.84M)
WARN_THRESHOLD_PERCENT_90=true             # email при 90% (2.07M)
WARN_THRESHOLD_PERCENT_95=true             # email при 95% (2.185M)
LIMIT_CHECK_TIMEZONE=Europe/Moscow         # для year boundary
```

ENV-конфиг можно override через runtime (план 52: superadmin
endpoint в identity для глобальных flags), но это опционально.

### Hard-stop — defence in depth

**3 уровня проверки** (если хоть один сработал — пополнение блокируется):

1. **Frontend** (`portal/frontend` или где «Пополнить оксары»):
   - Перед отображением кнопки «Пополнить» дёргает
     `GET /api/billing/limits/status` → `{ active: bool, message?: string }`.
   - Если `active=false` → показывает disabled-button с message
     («Пополнение временно недоступно. Попробуйте позже.»).
   - Это UX, не security — бэкенд всё равно перепроверит.

2. **Billing-service в `BuildPayURL`** (защита от stale frontend
   cache):
   - Перед созданием платежа в провайдере → `LimitsService.IsActive(ctx)`.
   - Если false → `return ErrLimitReached` (HTTP 503 на API-ответе с
     нейтральным message).

3. **Periodic reconciliation job** (15-минутный cron в
   billing-service):
   - Считает revenue_ytd → сравнивает с HARD_STOP_THRESHOLD_KOP.
   - При превышении → ставит global flag `payments_disabled=true` в
     `billing.system_state` таблице.
   - При уменьшении (refunds) → flag сам не сбрасывается; нужен manual
     override через admin-UI («Включить пополнение обратно»).

**Pending платежи**: уже инициированные YooKassa-платежи (статус
`pending` или `waiting_for_capture`) **не отменяем** — добиваем до
конца, иначе клиент потеряет деньги. Это значит **revenue может
превысить HARD_STOP_THRESHOLD_KOP на размер pending-платежей** (~10-50
тыс ₽), отсюда буфер 100к от лимита ФНС.

### Алерты при приближении к лимиту

**Каналы**:
- **Email** — обязательный, на адреса из ENV `BILLING_ALERT_EMAILS`
  (CSV).
- **Telegram bot** — опциональный, через ENV `BILLING_TELEGRAM_BOT_TOKEN`
  + `BILLING_TELEGRAM_CHAT_ID`.
- **Prometheus alert** — `billing_revenue_ytd_kopeks` exposed как
  metric, alerting rules в `deploy/prometheus/alerts.yml`:
  ```yaml
  - alert: BillingRevenueAt80Percent
    expr: billing_revenue_ytd_kopeks > 192000000
  - alert: BillingRevenueAt90Percent
    expr: billing_revenue_ytd_kopeks > 216000000
  - alert: BillingRevenueAt95Percent
    expr: billing_revenue_ytd_kopeks > 228000000
  - alert: BillingHardStopActivated
    expr: billing_payments_disabled == 1
  ```

**Trigger logic** (избегаем спама):
- Каждое пороговое значение шлёт email **один раз** (state в
  `billing.alert_state` table: `last_sent_threshold INT`).
- При новом году (1 января МСК) state сбрасывается, всё начинается
  заново.

### Возвраты (refunds)

**По закону НПД**: возврат уменьшает annual revenue (учитывается со
знаком минус в «Мой налог»). Значит:
- `refund.succeeded` event → revenue_ytd уменьшается.
- Если был hard-stop, после refund'а revenue может упасть ниже
  threshold → но мы НЕ авто-снимаем блокировку (manual override через
  admin-UI), чтобы избежать «yo-yo» отключений.

Возвраты **разрешены законом**: ФЗ-422 ст. 8 п. 4 — корректировка
дохода в случае возврата. YooKassa: `POST /v3/refunds`. Mock-provider
тоже поддерживает возвраты.

В админ-UI:
- Кнопка «Возврат» рядом с каждым `succeeded` платежом.
- Form: amount (default = full), reason textarea (обязательно),
  notify_user checkbox.
- Audit запись в `billing.refunds_audit`.

## Endpoints в billing-service

```
# Frontend (юзер) — публичный
GET    /api/billing/limits/status                  # {active, message?, threshold_percent?}
                                                    # без auth, кеш 30 сек

# Admin — RequirePermission("billing:read")
GET    /api/admin/billing/dashboard                # сводка: revenue YTD, today, top providers
GET    /api/admin/billing/revenue?period=...       # графики: by day/week/month
GET    /api/admin/billing/payments?...             # список с фильтрами
GET    /api/admin/billing/payments/{id}            # детали
GET    /api/admin/billing/refunds?...              # список возвратов

# Admin — RequirePermission("billing:refund")
POST   /api/admin/billing/payments/{id}/refund     # body: {amount_kop, reason, notify_user}

# Admin — RequirePermission("billing:limits")
GET    /api/admin/billing/limits                   # текущая конфигурация + state
PUT    /api/admin/billing/limits/active            # body: {active: bool, reason: string}
                                                    # ручное отключение/включение

# Admin — RequirePermission("billing:reports")
GET    /api/admin/billing/export.csv?period=...    # выгрузка для бухгалтерии
GET    /api/admin/billing/audit                    # audit log billing-операций
```

Все admin endpoints проверяют JWT permissions из RBAC (план 52),
логируют в `billing.audit_log`, rate-limit 120 req/min per admin.

## Pages в admin-frontend

(Дополнения к плану 53.)

### `/billing` (Dashboard)

```
┌─────────────────────────────────────────────────────────┐
│  Billing                                                 │
├─────────────────────────────────────────────────────────┤
│ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────────────┐│
│ │ Revenue │ │ Limit   │ │ Today   │ │ Status          ││
│ │ YTD     │ │ Used    │ │         │ │                 ││
│ │ 1.84M ₽ │ │ 76.7%   │ │ 12,400₽ │ │ ✓ Active        ││
│ │ ↑ +18%  │ │ ▓▓▓▓▓▓░ │ │ ↑ +8%   │ │ Last check 2m   ││
│ └─────────┘ └─────────┘ └─────────┘ └─────────────────┘│
│                                                          │
│ ┌─Revenue last 30 days──────────────────────────────────┐│
│ │ [recharts line chart, daily revenue, RUB]            ││
│ └─────────────────────────────────────────────────────┘│
│                                                          │
│ ┌─Recent payments───────────────────┐ ┌─Alerts────────┐│
│ │ table: 10 last payments           │ │ 80% threshold ││
│ │                                   │ │ reached 5d ago││
│ └───────────────────────────────────┘ └───────────────┘│
└─────────────────────────────────────────────────────────┘
```

Метрики автообновляются каждые 30 сек (TanStack Query staleTime).

### `/billing/payments`

- TanStack Table со sortable columns: id, date, user, amount,
  provider, status, captured_at.
- Фильтры в шапке: date-range, status (pending/succeeded/failed/
  refunded), provider (yookassa/mock/...), search-by-orderID.
- Bulk actions: export selected to CSV.
- Click-row → `/billing/payments/{id}` modal/sheet с деталями
  (full payload, webhooks history, related refunds).

### `/billing/refunds`

- Аналогичная таблица для refunds.
- Фильтры: date-range, original-payment search.
- Export CSV.

### `/billing/limits`

- **Текущее состояние**: active/inactive (большой статус-bar).
- **Конфигурация** (read-only из ENV, для display):
  - Лимит ФНС: 2 400 000 ₽
  - Hard-stop порог: 2 300 000 ₽
  - Текущий revenue YTD: ... ₽ (...% от лимита)
- **Manual override**:
  - Toggle «Active» (требует confirmation modal с reason textarea).
  - Last manual change: timestamp + actor + reason.
- **Year-boundary preview**: дата сброса (1 января UTC+3), countdown.
- **Alerts history**: какие пороги были сработаны в этом году.

### `/billing/reports`

- Генератор отчётов:
  - Period: месяц / квартал / год (с date-picker).
  - Format: CSV / PDF / JSON.
  - Включить refunds: yes/no.
  - Группировать: by day / by user / by provider.
- Export → download.

## Backend implementation

### `internal/limits/`

```go
package limits

type Service interface {
    IsActive(ctx context.Context) (active bool, message string, err error)
    GetState(ctx context.Context) (LimitState, error)
    SetActive(ctx context.Context, active bool, actorID uuid.UUID, reason string) error
    GetRevenueYTD(ctx context.Context) (kop int64, err error)
}

type LimitState struct {
    Active             bool
    RevenueYTDKop      int64
    LimitKop           int64       // 240000000
    HardStopKop        int64       // 230000000
    UsedPercent        float64
    LastCalculatedAt   time.Time
    ManuallyOverridden bool
    ManualReason       string
    ManualActorID      *uuid.UUID
}
```

- `IsActive` дёргается на каждый `BuildPayURL` → быстрый (cached).
- Cache: in-process, TTL 30 сек (refresh from DB).
- `SetActive(false, ...)` помещает запись в `audit_billing_admin`.

### Reconciliation job

```go
func RunReconciliationLoop(ctx, interval=15*time.Minute) {
    for { ... }
}
```

- Считает revenue_ytd через SQL.
- Сравнивает с thresholds → шлёт alert если новый порог пройден
  (state в `billing.alert_state`).
- При revenue ≥ HARD_STOP → авто `SetActive(false, ..., "auto: limit reached")`.

### Migrations

```sql
-- 010_limits_state.sql
CREATE TABLE billing.system_state (
    id                  SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id=1),  -- singleton
    payments_active     BOOLEAN NOT NULL DEFAULT TRUE,
    last_changed_by     UUID,
    last_changed_at     TIMESTAMPTZ,
    last_change_reason  TEXT,
    auto_disabled_at    TIMESTAMPTZ
);
INSERT INTO billing.system_state (id, payments_active) VALUES (1, true);

-- 011_alert_state.sql
CREATE TABLE billing.alert_state (
    year                INTEGER PRIMARY KEY,
    threshold_80_sent   TIMESTAMPTZ,
    threshold_90_sent   TIMESTAMPTZ,
    threshold_95_sent   TIMESTAMPTZ,
    threshold_hard_sent TIMESTAMPTZ
);

-- 012_billing_audit.sql
CREATE TABLE billing.audit_log (
    id           BIGSERIAL PRIMARY KEY,
    actor_id     UUID NOT NULL,
    action       VARCHAR(64) NOT NULL,  -- 'limit:disable', 'limit:enable', 'refund:create'
    target_type  VARCHAR(32),           -- 'payment' | 'limit' | 'system'
    target_id    TEXT,
    payload      JSONB,
    ip_address   INET,
    user_agent   TEXT,
    created_at   TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX ON billing.audit_log(created_at DESC);
CREATE INDEX ON billing.audit_log(actor_id);
CREATE INDEX ON billing.audit_log(action);
```

### Prometheus metrics

```go
billing_revenue_ytd_kopeks       Gauge
billing_revenue_today_kopeks     Gauge
billing_payments_succeeded_total Counter labels: provider
billing_payments_failed_total    Counter labels: provider, reason
billing_refunds_total            Counter
billing_payments_disabled        Gauge (0 or 1)
billing_limit_threshold_percent  Gauge (current % of HARD_STOP)
```

## Прогресс выполнения

**Backend (2026-04-28)**: ✅ закрыт.

- ✅ Ф.1 Database + base service — миграция `0002_limits_and_audit.sql`
  (3 таблицы: `billing_system_state`, `billing_alert_state`,
  `billing_audit_log`) + `internal/limits/service.go` с `IsActive`
  (in-process cache TTL 30s), `GetRevenueYTD`, `SetActive`,
  `AutoDisable` (idempotent), `MarkAlerted`/`AlertedAt`.
- ✅ Ф.2 Hard-stop integration — `billing.CreateOrder` вызывает
  `limits.IsActive()` до `BuildPayURL`, при `false` возвращает
  `ErrLimitReached`. Handler отдаёт **HTTP 503** с нейтральным
  сообщением «Пополнение временно недоступно. Попробуйте позже.».
- ✅ Ф.3 Reconciliation job — `internal/limits/reconciler.go`,
  15-минутный loop (env `LIMIT_CHECK_INTERVAL`). Hard-stop при
  revenue ≥ HARD_STOP, soft-warning при 80/90/95% (один раз/год через
  `MarkAlerted`).
- ✅ Ф.4 Soft-warning notifications — `SlogNotifier` пишет structured
  warn/error (`billing_threshold_reached`, `billing_auto_disabled`).
  Email и Telegram отложены до плана 57 (mail-service); интерфейс
  `Notifier` готов к замене.
- ✅ Ф.5 Public API — `GET /api/billing/limits/status` без auth,
  отдаёт `{"active": bool, "message": "..."}`. Cache внутри Service
  защищает от наплыва (~1 SQL/30s).
- ✅ Ф.6 Admin endpoints — `GET /api/admin/billing/limits/status`
  (расширенный, для админа), `POST /api/admin/billing/limits/override`,
  `GET /api/admin/billing/reports/revenue?from=&to=&granularity=`,
  `GET /api/admin/billing/reports/csv?from=&to=`. RBAC permissions:
  `billing:reports:read`, `billing:admin:override`,
  `billing:reports:csv_export`. Все мутации пишутся в
  `billing_audit_log` транзакционно.
  В `projects/admin-bff/cmd/server/main.go` уже есть proxy
  `/api/admin/billing/*` → billing-service (план 53), доп. правок не
  понадобилось.
- ✅ Ф.7 ENV config — `HARD_STOP_THRESHOLD_KOP`,
  `WARN_THRESHOLD_PERCENT_{80,90,95}`, `LIMIT_CHECK_TIMEZONE`,
  `LIMIT_CHECK_INTERVAL` в `cmd/server/main.go`.
- ✅ Ф.8 Тесты — unit (`HighestPassed`, `startOfYear`, `alertColumn`)
  + integration в Postgres (`BILLING_TEST_DB_URL`): 8 cases service-
  логики + полный flow «1000 платежей → hard-stop → 503 → admin
  override → 1002-й проходит». Все зелёные.

**Frontend**: отложен в план **58 Ф.5b** (i18n-корректировка после
ребрендинга «кредиты» → «оксары»). UI «Пополнить кредиты» в
портале — не трогается.

**Известные ограничения backend-части**:

- Email-уведомления при пересечении порогов — не реализованы (план 57
  `mail-service` ещё не готов). Сейчас только structured slog-warn
  (`event=billing_threshold_reached`/`billing_auto_disabled`) +
  Prometheus gauge `billing_payments_disabled`. Для prod нужна
  Grafana alert rule на эту метрику.
- Refunds в схеме billing-сервиса (план 38) пока не выделены отдельной
  сущностью — `revenue_ytd` считается без вычета возвратов. Это
  **консервативно**: фактический доход самозанятого ≤ revenue_ytd, что
  добавляет ещё буфер к HARD_STOP. Когда план 38 выделит refunds в
  payment_orders, в `GetRevenueYTD` добавится `- SUM(refunds)`.
- Multi-instance billing-сервис не поддерживается (один reconciler =
  один процесс). Для horizontal scale потребуется PG advisory-lock
  (как в `game-nova/scheduler` плана 32).

## Этапы

### Ф.1. Database + base service

- Миграции 010-012.
- `internal/limits/service.go` с базовой реализацией `IsActive`,
  `GetRevenueYTD`, `SetActive`.
- Unit-tests с stub-DB (test postgres).
- Интеграция в `cmd/server/main.go`: register service, wire DI.

### Ф.2. Hard-stop integration

- В `internal/payment/gateway.go`: перед `BuildPayURL` вызов
  `limits.IsActive`. Возврат `ErrLimitReached` (HTTP 503).
- В существующих тестах: добавить case «limit reached → BuildPayURL
  отказывает».
- Frontend (где «Пополнить оксары»): отдельный endpoint
  `GET /api/billing/limits/status` без auth, проверяет `IsActive`,
  возвращает `{active, message}`. UI скрывает/disabled кнопку.

### Ф.3. Reconciliation job

- `internal/limits/reconciler.go` — 15-минутный loop.
- Email-sender через `internal/notify/email.go` (если ещё нет —
  создать через symfony/mailer-эквивалент в Go: gomail.v2).
- Telegram-sender через `internal/notify/telegram.go` (опциональный,
  ENV-driven).
- Tests: timestamp-mock для воспроизведения thresholds.

### Ф.4. Admin endpoints

- `internal/admin/handler.go`: dashboard, payments list, refunds list,
  limit toggle, refund create.
- Middleware `RequirePermission("billing:read"|"billing:refund"|...)`.
- Audit logging (записываем в `billing.audit_log`).
- OpenAPI spec в `projects/billing/api/openapi.yaml`.

### Ф.5. Admin-frontend integration

- В `projects/admin-frontend/src/features/billing/`:
  - `Dashboard.tsx` — карточки + графики.
  - `PaymentsList.tsx` — TanStack Table.
  - `PaymentDetail.tsx` — sheet с деталями.
  - `RefundDialog.tsx` — modal для возврата.
  - `LimitsScreen.tsx` — состояние + override.
  - `ReportsScreen.tsx` — генератор отчётов.
- TanStack Query хуки в `src/api/billing.ts`.
- Routes в admin-router.

### Ф.6. Prometheus + alerts

- Metrics endpoint в billing-service `/metrics` (если ещё нет).
- `deploy/prometheus/alerts.yml`: rules для thresholds.
- Grafana-dashboard (опционально, отдельный план если нужно).

### Ф.7. Reports + CSV export

- Backend: streaming CSV-generator (большие отчёты).
- Frontend: download button с loader, fallback to email-link если
  отчёт > 1 минуты.

### Ф.8. Финализация

- Smoke-сценарий:
  1. Триггерить fake revenue (через тест-бд) до 80% → email пришёл.
  2. До 100% → hard-stop сработал, `IsActive=false`.
  3. UI отображает «Пополнение временно недоступно».
  4. Admin вручную override → `IsActive=true`, audit-запись.
  5. Refund → revenue уменьшилось, но `IsActive` не сбросился авто.
- Документация:
  - `docs/ops/billing-monitoring.md` — runbook на достижение лимита
    (как реагировать админу, как сделать override).
  - `docs/ops/billing-alerts.md` — настройка email/telegram.
  - `docs/architecture/billing-limits.md` — алгоритм лимита и
    threshold-логика.
- `docs/project-creation.txt` — итерация 54.

## Тестирование

### Unit (Go)

- `limits.IsActive` с моком DB.
- Reconciler с mock-clock (transitions между thresholds).
- Refund сервис с моком provider'а.

### Integration (Go + test postgres)

- E2E flow: insert payments → reconciler → expect alert → set 100% →
  expect auto-disable.
- Refund flow: insert payment → refund → expect revenue decrease.

### Frontend (Vitest + msw)

- `LimitsScreen` рендерит state correctly.
- Toggle «Active» вызывает API + audit appears.
- Refund modal: form validation, success toast.

### E2E (Playwright)

- Admin login → dashboard → see revenue.
- Refund flow: open payment → click refund → confirm → see status updated.
- Limit override: toggle off → see disabled state in user-flow.

## Риски

1. **Reconciler пропустил threshold**: если cron не запустился,
   alert не пришёл, и revenue прошёл порог незаметно. Митигация:
   monitoring сам job (Prometheus метрика `billing_reconciler_last_run`).
2. **Pending-платежи в момент hard-stop**: револьвер уменьшен на ~10-50k.
   Митигация: буфер 100k от лимита ФНС.
3. **Refund после hard-stop**: revenue упало ниже threshold, но
   платежи остаются disabled. Это **by design** — ручное включение
   через admin-UI с reason. Иначе yo-yo.
4. **Year boundary** (1 января): reconciler должен правильно
   обработать переход. Митигация: tests с мок-clock на 31 декабря 23:59.
5. **Multiple admins одновременно жмут override**: race condition.
   Митигация: optimistic locking через `version` в `system_state`.
6. **Выгрузка отчётов больших периодов**: streaming CSV, лимит на 100k
   rows для синхронной выгрузки, иначе email-link с ссылкой на
   подписанный URL.

## Out of scope

- Поддержка multi-currency (сейчас только RUB).
- Поддержка multi-tenancy (один продавец).
- Авто-выписка чеков самозанятого через «Мой налог» API (отдельная
  задача, после получения NPD-account).
- Прогноз перелимита через ML/forecasting — overkill, простых
  threshold'ов достаточно.

## Альтернативы (отвергнуты)

- **Лимит на стороне provider'а** (YooKassa shop-level limit): не у
  всех provider'ов есть. Не universal solution.
- **Static config (без БД)** для system_state: невозможно сделать
  manual override через admin без re-deploy.
- **Авто-снятие блокировки после refund**: yo-yo пользовательский
  опыт, плюс race conditions на год boundary.

## Итог

Полная защита от перелимита самозанятого с buffer'ом 100k, defence
in depth (frontend + backend + reconciler), email/Telegram alerts на
80/90/95%, immutable audit log, admin-UI для override и отчётов с
CSV export. Пополнение кредитов автоматически отключается при
достижении 2.3M ₽, ручное включение через профессиональную консоль.
