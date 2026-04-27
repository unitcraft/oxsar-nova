# План 58: Ребрендинг валюты — Оксары + Оксариты (двухвалютная схема)

**Дата**: 2026-04-27 (обновлён 2026-04-27 после согласования архитектуры)
**Статус**: Активный
**Зависимости**: ADR-0009 ([docs/adr/0009-currency-rebranding.md](../adr/0009-currency-rebranding.md))
— архитектурное решение и детальный анализ. Текущие планы:
[06-credits-ai-advisor.md](06-credits-ai-advisor.md),
[25-credits-economy.md](25-credits-economy.md),
[38-billing-service.md](38-billing-service.md),
[42-yookassa.md](42-yookassa.md),
[47-offer-tos.md](47-offer-tos.md) §5,
[17-gameplay-improvements.md](17-gameplay-improvements.md) (alien-механики).

---

## Цель

Перевести проект на двухвалютную схему **Оксары** (hard, в billing) +
**Оксариты** (soft, в game-nova per universe). Обоснование, юридический
анализ, UX и edge-cases — в ADR-0009.

В этом плане — конкретные шаги реализации.

---

## Что меняем

### 1. БД — переименование и новые таблицы

**billing**:
- В существующей таблице кошелька (если есть из плана 38 — `wallets`):
  колонка `credit_balance` → `oxsar` (`bigint` вместо `numeric`).
  Если её нет — создать `wallets` с `oxsar`.
- Новая таблица `wallet_transactions`:
  - `id UUID PRIMARY KEY`,
  - `user_id UUID NOT NULL`,
  - `oxsar BIGINT NOT NULL` (signed, отрицательные — расход),
  - `kind TEXT NOT NULL` (purchase / charge / refund / admin_adjust),
  - `idempotency_key TEXT UNIQUE`,
  - `metadata JSONB`,
  - `created_at TIMESTAMPTZ DEFAULT now()`.

**game-nova**:
- `users.credit` → `users.oxsarit` (rename + change type to `bigint`).
- Новая таблица `oxsarit_transactions`:
  - `id UUID PRIMARY KEY`,
  - `user_id UUID NOT NULL`,
  - `universe_id TEXT NOT NULL`,
  - `oxsarit BIGINT NOT NULL` (signed),
  - `kind TEXT NOT NULL` (spend / battle_loss / alien_loss / alien_gift / event_reward / achievement / daily_login / referral / admin_adjust),
  - `source JSONB` (контекст: `event_id`, `battle_id`, `referrer_id` и т.п.),
  - `created_at TIMESTAMPTZ DEFAULT now()`.

**identity** (если есть упоминания credit):
- Если в плане 36 был эндпоинт `/auth/credits/*` (по плану 38 он
  должен был переехать в billing) — проверить, что упоминаний нет.
  Если остались — удалить.

**game-origin** (PHP):
- Если в game-origin БД есть `users.credit` — переименовать в `users.oxsarit`
  (в этой же миграции с game-nova, если БД общая).
- Если БД отдельная — отдельная миграция.

**Атомарность деплоя**: используем **maintenance window 5–15 минут**
или blue-green деплой (план 31 zero-downtime). Все backend-серверы
обновляются ОДНОВРЕМЕННО с миграцией.

### 2. Backend — Smart-pay и charge API

**billing-service** новый эндпоинт (internal-token):

```
POST /internal/billing/wallet/charge
{
  "user_id": "uuid",
  "amount": 20,
  "purpose": "officer_purchase",
  "metadata": {"officer_id": "...", "universe_id": "uni01"},
  "idempotency_key": "uuid"
}
→ 200 {"new_balance": 4980, "transaction_id": "..."}
→ 402 Payment Required {"reason": "insufficient_balance", "balance": 5}
→ 409 Conflict (idempotency key уже использован — вернуть старый ответ)
```

**game-nova** — обновить логику покупки премиум-фичи (офицеры,
AI-советник, премиум-features из плана 06/25):

```go
// Псевдокод:
func BuyOfficer(ctx, userID, officerID) error {
    price := config.OfficerPrices[officerID]  // например 50

    // Шаг 1: сколько списать с оксаритов
    user := getUser(ctx, userID)
    fromOxsarites := min(price, user.Oxsarites)
    fromOxsars := price - fromOxsarites

    // Шаг 2: транзакция в game-nova-БД
    tx := beginTx(ctx)
    if fromOxsarites > 0 {
        tx.users.UpdateOxsarites(userID, -fromOxsarites)
        tx.oxsarit_transactions.Insert(...)
    }

    // Шаг 3: если не хватило оксаритов, списать остаток из billing
    if fromOxsars > 0 {
        idemKey := uuid.New()
        resp, err := billingClient.WalletCharge(ctx, userID, fromOxsars, idemKey, ...)
        if err == ErrInsufficientBalance {
            tx.Rollback()
            return ErrNotEnoughFunds
        }
        // ok
    }

    tx.Commit()
    applyOfficerEffect(ctx, userID, officerID)
    return nil
}
```

Опция «оплатить только из кошелька» — параметр запроса
`pay_from: "wallet_only"`, тогда `fromOxsarites = 0`,
`fromOxsars = price`.

**Идемпотентность critical**: при сбое HTTP-вызова billing.charge
(timeout, 5xx) — в game-nova остаётся `pending_transaction`
запись, через retry-cron periodically повторяется. Это
описано в плане 38 / 42 для платежей, переиспользовать паттерн.

### 3. Frontend — UI для двух валют

**game-nova/frontend**:
- `BalanceBadge` компонент → две части:
  - левая: «⬢ N оксаритов (uni01)»;
  - разделитель `|`;
  - правая: «🪙 M оксаров (кошелёк)».
- TanStack Query keys: `['billing', 'wallet']` для оксаров;
  `['game-nova', 'me']` для оксаритов (поле `oxsarit`).
- Покупка премиум-фичи: модалка с разбивкой
  «Списано 30 оксаритов + 20 оксаров. Опция: ☐ Только из кошелька».

**portal/frontend**:
- Шапка: только «🪙 5000 оксаров (кошелёк)».
- Клик → модалка «Оксариты во вселенных: uni01: 30, uni02: 0, ...».
- Shop UI: «Купить пакет 1000 оксаров за 100 ₽».

**game-origin** (пока PHP, в будущем переписан на Go+React):
- Шаблоны `*.tpl` — обновить терминологию.
- Логика покупки — через тот же billing.charge endpoint.

**Plural rules** (CLDR ru) для UI-строк:

```yaml
# projects/game-nova/configs/i18n/ru.yml
oxsars:
  count_one: "{n} оксар"
  count_few: "{n} оксара"
  count_many: "{n} оксаров"
oxsarit:
  count_one: "{n} оксарит"
  count_few: "{n} оксарита"
  count_many: "{n} оксаритов"
```

Имена namespace (`oxsar`, `oxsarit`) — единственное число,
согласованно с именами колонок в БД и Go-структурами.

Frontend подбирает форму по числу через i18n-библиотеку (i18next /
react-intl / Format.JS).

### 4. Юридические документы (обновляются)

- `docs/legal/offer.md` §5 «Виртуальная валюта»: переименовать
  «Кредиты» на «Оксары и Оксариты». Описать обе валюты, разное
  правовое положение (оксары — деньги, ст. 437 ГК; оксариты —
  игровой ресурс, ст. 1062 ГК). Версия документа `1.0` → `1.0.1`.
  **Существующие `user_consents` с version='1.0' остаются валидными** —
  это переименование, не изменение существенных условий (ADR-0009 §2).
- `docs/legal/refund-policy.md` — обновить терминологию: рефанд
  применим к оксарам (hard), не к оксаритам.
- `docs/legal/privacy-policy.md` — упомянуть оба типа в категориях
  обрабатываемых данных (если применимо).
- `docs/legal/game-rules.md` (план 47) — отдельный пункт «Оксариты —
  игровой ресурс, может теряться от инопланетян и событий по
  правилам игры».

### 5. Активные технические планы (обновляются)

- **Файлы переименовать:**
  - `docs/plans/25-credits-economy.md` → `25-oxsars-economy.md`;
  - `docs/plans/06-credits-ai-advisor.md` → `06-oxsars-ai-advisor.md`.
- **Содержимое обновить** (терминология, добавить раздел про оксариты):
  - 25, 06, 38, 42, 47.
- **Memory-трекер** `completed_plans.md` — обновить ссылки на
  переименованные планы.

### 6. OpenAPI / SDK / типы

- `projects/game-nova/api/openapi.yaml` — поля `credit` →
  `oxsar` / `oxsarit` соответственно. Schemas обновлены.
- Перегенерировать TypeScript-клиенты во всех frontend.
- Аналогично portal-API, billing-API.

### 7. Метрики и логирование

- Prometheus-метрики:
  - `credits_purchased_total` → `oxsar_purchased_total`;
  - новые: `oxsarit_earned_total{kind="alien_gift|battle|achievement"}`,
    `oxsarit_lost_total{kind="alien_loss|spend"}`.
- Slog-поля: `user_credit` → `user_oxsar` (для логов billing) /
  `user_oxsarit` (для game-nova).
- Audit-log записи (план 14) — обновить шаблоны сообщений.

### 8. Миграция существующих балансов

Если в проде есть пользователи с накопленными `credit`:
- Все существующие `users.credit` копируются как hard:
  `wallets.oxsar := users.credit`.
- `users.oxsarit := 0` (оксариты — новая сущность, начальное 0).

В dev-окружении — пропустить (тестовые данные сбрасываются).

### 9. Жизненный цикл (см. ADR-0009 §8)

Реализовать обработку для каждого события:

| Событие | Где обрабатывается |
|---|---|
| YooKassa webhook payment.succeeded | billing-service: `wallet_transactions.kind='purchase'` |
| Smart-pay charge | billing-service: `wallet_transactions.kind='charge'` + game-nova: `oxsarit_transactions.kind='spend'` |
| Награда за бой | game-nova battle-engine: `oxsarit_transactions.kind='battle_reward'` |
| Кража инопланетянами | game-nova alien-engine: `oxsarit_transactions.kind='alien_loss'` |
| Подарок инопланетян | `oxsarit_transactions.kind='alien_gift'` |
| Daily login bonus | scheduler-tick: `oxsarit_transactions.kind='daily_login'` |
| Achievement unlock | achievements-engine: `oxsarit_transactions.kind='achievement'` |
| Реферал-награда | (см. план для рефералов) `oxsarit_transactions.kind='referral'` |
| Рефанд (план 47) | billing-service: `wallet_transactions.kind='refund'` |
| Бан за нарушение | game-nova admin: `oxsarit_transactions.kind='admin_adjust'` (с metadata о причине) |
| Удаление аккаунта (план 44) | billing-service deperson: hard в пределах 14 дней — рефанд непотраченного, остальное — деперсонализация (sender_id затирается, запись хранится 5 лет по 402-ФЗ); soft — сгорают |

---

## Источники оксаритов на старте (план 25 / 06 — обновляется)

По решению: **A на старте, B по мере накопления контента** (см. ADR-0009).

**Старт (Ф.X плана 58 — фиксируется в коде):**
- Достижения (achievements engine, план 06): за каждое — N оксаритов.
- Подарок инопланетян (alien-engine, планы 17 / 15): редкие события.
- Daily login bonus (scheduler, опционально по плану 25): N оксаритов
  за каждый день.

**Расширение (отдельные геймплейные планы в будущем):**
- Награды за экспедиции (план 02);
- Ивенты галактики (план 17 F);
- Турниры/рейтинги (план 17 E);
- Daily quests (план 17 D).

Точные числа выдачи — в `configs/economy.yaml` (или эквивалент),
балансировка — отдельная задача.

---

## Чего НЕ делаем

- **Не вводим биржу hard ↔ soft.**
- **Не вводим вторую hard-валюту.**
- **Не меняем формулы баланса** AI-advisor / офицеров.
- **Не покупаются оксариты за рубли** (это фундаментальное юр-разделение).
- **Не отнимаются оксариты в pvp между игроками** (только инопланетяне/события).
- **Не вводим лутбоксы.**
- **Не уведомляем пользователей автоматически** о ребрендинге —
  это маркетинговая задача (банер на portal, новость, FAQ).
- **Не переписываем историю коммитов** или legacy-документы.

---

## Этапы

### Ф.1. Подготовка миграций БД

- Найти все места: `grep -rn "credit" projects/*/migrations/`.
- Подготовить миграции:
  - billing: создать `wallets` с колонкой `oxsar` (или
    переименовать существующую если уже есть из плана 38) +
    `wallet_transactions`.
  - game-nova: `users.credit → users.oxsarit` (rename + bigint) +
    `oxsarit_transactions`.
  - game-origin (если применимо): то же.
- Прогнать в dev-БД.

### Ф.2. Backend — billing.charge API

- Реализовать `POST /internal/billing/wallet/charge` с идемпотентностью.
- Тесты: charge при insufficient balance, idempotency-replay,
  конкурентные запросы.

### Ф.3. Backend — game-nova smart-pay

- Логика: списание оксаритов → fallback на оксары.
- Опция `pay_from: "wallet_only"`.
- Retry-cron для pending_transaction.
- Unit-тесты.

### Ф.4. Backend — game-nova oxsarite earnings

- Battle-engine: emit `oxsarit_transactions` при награде.
- Alien-engine: emit при подарке/краже.
- Achievements-engine: emit при unlock.
- Daily login (если используется): emit.

### Ф.5. Frontend — UI

- BalanceBadge с двумя значениями.
- Modal для покупки премиум-фичи с разбивкой.
- Plural rules.
- Shop UI: пакеты оксаров.

### Ф.5b. Корректировка локализационных текстов (i18n)

Сквозная замена терминологии «кредит/credit» → «оксар/оксарит» во
всех языковых ресурсах. Это **самая объёмная** часть фронтенд-работы,
выделена в отдельную фазу для контроля.

**Где искать строки:**

1. `projects/game-nova/configs/i18n/{ru,en}.yml` — основной i18n.
2. `projects/portal/frontend/src/i18n/` (если используется) или
   inline-строки в `.tsx`-файлах.
3. `projects/admin-frontend/src/i18n/` — админка (план 53).
4. `projects/game-origin/src/templates/standard/*.tpl` —
   Smarty-шаблоны legacy (фразы вида «у вас N кредитов», «купите
   кредиты», «кредитов не хватает»).
5. `projects/game-origin/src/db/na_phrases` (если используется
   таблица фраз legacy для перевода) — обновить русские/английские
   значения.
6. PHP-код game-origin: hardcoded-фразы в `src/game/page/*.class.php`
   (поиск по `"кредит"` и аналогам).
7. Системные сообщения в email/уведомлениях (план 57 mail-service —
   когда будет реализован, шаблоны системных писем).

**Что заменяется:**

| Было | Стало |
|---|---|
| `credit` (key/имя группы переводов) | `oxsar` (для billing-операций) или `oxsarit` (для game-nova-операций) |
| «N кредитов» | через plural rules: «N оксаров» / «N оксаритов» |
| «Купить кредиты» / «Buy credits» | «Купить оксары» / «Buy oxsars» |
| «Пополнить кредиты» | «Пополнить оксары» |
| «У вас недостаточно кредитов» | в зависимости от контекста: «недостаточно оксаритов» (если про soft-баланс в игре) или «недостаточно средств на счёте» (если smart-pay не справился ни с soft, ни с hard) |
| «Стоимость в кредитах» | «Стоимость в оксарах» (если hard-only) или «Стоимость» (нейтрально, без указания валюты — т.к. smart-pay) |
| «Бонусные кредиты за {достижение}» | «Бонусные оксариты за {достижение}» (это всегда soft) |
| «Реферальный бонус» — суммы в кредитах | в оксаритах (план 59 фиксирует реф-награды как soft) |

**Что НЕ заменяется:**

- Названия лицензий и юр-документов: текст оферты обновляется по
  Ф.7 этого плана, не в i18n.
- Названия в исторических документах (`docs/project-creation.txt`,
  `docs/simplifications.md`, `docs/ui/dev-log.md`) — план 49
  гигиены ПДн допускает исторический «кредит» в дневнике.
- Имена методов кода (Go-функций), они меняются в Ф.2-Ф.4 этого
  плана отдельно — это код, не i18n.

**Plural rules для русского:**

```yaml
oxsar:
  count_one: "{n} оксар"
  count_few: "{n} оксара"
  count_many: "{n} оксаров"
oxsarit:
  count_one: "{n} оксарит"
  count_few: "{n} оксарита"
  count_many: "{n} оксаритов"
```

Для английского — стандартный CLDR:
```yaml
oxsar:
  count_one: "{n} oxsar"
  count_other: "{n} oxsars"
oxsarit:
  count_one: "{n} oxsarit"
  count_other: "{n} oxsarites"
```

**Системные сообщения (push, email, in-game):**

Перечень шаблонов уведомлений, требующих обновления:
- «Покупка прошла успешно: +{n} оксаров на ваш счёт» (билинг
  webhook).
- «Списано {n} оксаритов и {m} оксаров за {услугу}» (smart-pay).
- «Получено {n} оксаритов: {источник}» (награды, ивенты, рефералы).
- «Недостаточно средств для покупки {услуги}. Требуется {N}
  оксаров, у вас {M}.» (smart-pay fail).
- «Возврат: {n} оксаров возвращены на карту» (рефанд).
- Реферал-уведомления (план 59): welcome-бонус, % от покупки.

**Verification:**

После корректировки прогнать поиск:

```bash
grep -ri "кредит" projects/*/frontend/ projects/game-origin/src/templates/ \
  projects/game-nova/configs/i18n/
grep -ri "credit" projects/*/frontend/src/i18n/ \
  projects/admin-frontend/src/i18n/
```

Должны остаться только:
- комментарии в коде (внутреннее обсуждение, не показывается);
- legacy-imports/types для backward-compatibility (если есть);
- цитаты из исторических планов.

Все user-facing строки переведены.

### Ф.6. OpenAPI и регенерация клиентов

- Обновить openapi.yaml.
- Перегенерировать TypeScript-клиенты.
- Прогнать E2E (план 13).

### Ф.7. Юридические документы

- Обновить offer.md (→ v1.0.1), refund-policy.md, privacy-policy.md,
  game-rules.md.
- Проверить: существующие user_consents v=1.0 остаются валидными.

### Ф.8. Документация и переименование

- Переименовать файлы 25 / 06.
- Обновить содержимое 25/06/38/42/47.
- Memory-трекер обновить.

### Ф.9. Миграция продовых данных (если есть)

- Скрипт миграции `users.credit → wallets.oxsar + users.oxsarit=0`.

### Ф.10. Финализация

- Smoke-тест end-to-end: регистрация → покупка пакета → smart-pay
  при покупке офицера → корректное отображение балансов.
- `git status --short` → коммитим только своими файлами поимённо.
- Запись в `docs/project-creation.txt` — итерация 58.
- Серия коммитов:
  - `feat(billing): wallet.charge API + oxsar`;
  - `feat(game-nova): smart-pay + oxsarit + transactions`;
  - `feat(frontend): UI для двух валют`;
  - `docs(legal): offer v1.0.1 + game-rules § оксариты`;
  - `docs(plans): обновить 25/06/38/42/47, переименовать файлы`.

---

## Тестирование

- Все Go-тесты во всех модулях зелёные.
- E2E (план 13): регистрация → пополнение через mock-YooKassa →
  smart-pay при покупке офицера → корректные балансы.
- Юр-документы рендерятся на портале (`/offer`, `/refund`, `/privacy`,
  `/game-rules`) с обновлённой терминологией.
- `docs/ops/legal-compliance-audit.md` повторный прогон → пробелов нет.

---

## Известные риски

1. **Атомарность деплоя.** Митигация: maintenance window или blue-green
   деплой по плану 31.
2. **Frontend-кэш у активных игроков.** Митигация: версионирование build,
   soft-reload.
3. **Спор «я считал что мои оксариты не должны были украсть».** Митигация:
   tooltip в UI «оксариты — игровой ресурс, может теряться»;
   game-rules §X.
4. **Маркетинг-уведомление.** Маркетинговая задача отдельно.

---

## Объём

5–7 коммитов, ~800–1500 строк изменений. Это **не** просто rename —
двухвалютная схема + smart-pay + транзакции — полноценная архитектурная
работа.

Время выполнения: **8–14 часов агента**.

---

## Когда запускать

После выполнения **базовых платёжных задач** (план 42 ЮKassa,
план 38 billing — ✅), но **до публичного запуска** — чтобы не
переименовывать на проде с уже накопленной базой пользователей.

Рекомендуемый порядок относительно текущей очереди:
1. Закрыть юр-планы (44 Ф.4 РКН — ручной шаг, 45 — ручной шаг).
2. Закрыть план 50 (game-origin gaps — Ф.1, Ф.4, Ф.5).
3. Закрыть план 56 (reports → portal).
4. **Выполнить план 58 (это).**
5. Маркетинговый анонс ребрендинга.
6. Публичный запуск.
