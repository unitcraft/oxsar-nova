# Промпт: выполнить план 76 (nova-frontend exchange UI)

**Дата создания**: 2026-04-28 (перезапись после планов 68/78)
**План**: [docs/plans/76-remaster-nova-frontend-exchange-ui.md](../plans/76-remaster-nova-frontend-exchange-ui.md)
**Зависимости**: ✅ план 68 (биржа backend), ✅ план 78 (раскладка
`frontends/nova/`). Параллелится с планом 72 Ф.3 (Spring 2 origin) —
разные папки.
**Объём**: ~600-1000 строк TS + i18n + тесты, 1-2 коммита.

---

```
Задача: выполнить план 76 — UI биржи артефактов в nova-фронте
(uni01/uni02). Backend готов (план 68), используем как есть.

КОНТЕКСТ:

План 68 закрыт коммитами 59c95b650c (миграция+OpenAPI), b3f566abde
(service+handler+repo), 127d763359 (event-handlers+balance config),
+ Ф.7 финализация (к моменту твоего старта тоже закрыта).

Endpoints биржи (см. projects/game-nova/api/openapi.yaml секция
/api/exchange/*):
- GET /api/exchange/lots — список с filters (artifact_unit_id,
  min_price, max_price, seller_id, status, cursor, limit).
- POST /api/exchange/lots — создать (Idempotency-Key обязателен).
- GET /api/exchange/lots/{id} — детали.
- POST /api/exchange/lots/{id}/buy — купить (Idempotency-Key).
- DELETE /api/exchange/lots/{id} — отозвать (только seller).

ВАЖНЫЕ АРХИТЕКТУРНЫЕ ДЕТАЛИ ИЗ ПЛАНА 68 (agent должен учесть):

1) Артефакты в nova — **row-per-item** (не quantity-based как в
   legacy). Каталог через `unit_id INT` (ARTEFACT_* коды), не
   TEXT-имена. UI отображает имя через i18n-резолв из каталога
   артефактов (nova уже умеет, см. internal/artefact/ или
   эквивалент на фронте).

2) Лот хранит:
   - `artifact_unit_id INT` — код артефакта.
   - `quantity INT` — сколько штук в лоте (агрегированный счётчик;
     N конкретных artefact_id хранятся в exchange_lot_items).
   - `price_oxsarit BIGINT` — цена за весь лот (не per item).
   - `status` (active/sold/cancelled/expired).
   - `seller_user_id`, `buyer_user_id`, `created_at`, `expires_at`,
     `sold_at`.

3) Валюта — оксариты. В БД хранится в `users.credit BIGINT` (легаси-
   имя, семантически = оксариты по ADR-0009; backend это берёт на
   себя). UI показывает «X оксаритов», не «credit». Переименование
   столбца — отдельный план в будущем (не сейчас).

4) Permit «Знак торговца» — **stub, всегда true**. UI не показывает
   permit-gating чек, ErrPermitRequired никогда не возвращается.

5) Антифрод (UI-валидация дублирует backend):
   - max_quantity_per_lot = 100 (CHECK в БД + 422 в handler).
   - max_active_lots_per_user = 10 (422 ErrMaxActiveLots).
   - price_cap_multiplier = 10× от reference (422 ErrPriceCapExceeded).
   - expires_in_hours: min 1, max 168 (7 дней).

ВАЖНО ПРО R5:
- Это nova-фронт, **современный nova-стиль**, НЕ pixel-perfect клон
  legacy.
- UI консистентен с существующими nova-фичами: используй компоненты
  src/ui/Modal.tsx, Toast.tsx, ResourceTicker.tsx, ProgressBar.tsx,
  Skeleton.tsx и пр. Не дублируй стили.
- Origin-фронт получит свой UI биржи в плане 72 спринт 5
  (отдельная сессия) — не лезем туда.

ПЕРЕД НАЧАЛОМ:

1) git status --short. cat docs/active-sessions.md.
   Параллельно ИДЁТ план 72 Ф.3 (Spring 2 origin-фронта) —
   разные папки (frontends/nova vs frontends/origin), конфликта
   быть не должно. На openapi.yaml ни ты, ни 72 Ф.3 не лезут
   (backend закрыт).

2) Прочитай ПОЛНОСТЬЮ:
   - docs/plans/76-remaster-nova-frontend-exchange-ui.md
   - docs/plans/68-remaster-exchange-artifacts.md (backend контекст)
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5» R0-R15
   - docs/adr/0009-currency-rebranding.md
   - projects/game-nova/api/openapi.yaml — секция /api/exchange/*

3) Прочитай выборочно:
   - projects/game-nova/frontends/nova/src/features/market/ если есть
     (близкий аналог по UX) или alliance/ (свежий пример план 67:
     TanStack Query + filters + dialogs)
   - projects/game-nova/frontends/nova/src/ui/ — базовые компоненты
   - projects/game-nova/frontends/nova/src/api/ — паттерн API-клиента
     и query-keys

4) Добавь свою строку в docs/active-sessions.md:
   | <N> | План 76 nova exchange UI | projects/game-nova/frontends/nova/src/features/exchange/, src/i18n/, src/api/exchange.ts | <дата-время> | feat(exchange,frontend): UI биржи в nova (план 76) |

ЧТО НУЖНО СДЕЛАТЬ:

### Ф.1. UI-каркас + API-клиент

- src/api/exchange.ts — TanStack Query функции:
  - listLots(filters, cursor) → GET /api/exchange/lots.
  - getLot(id) → GET /api/exchange/lots/{id}.
  - createLot(input, idempotencyKey) → POST.
  - buyLot(id, idempotencyKey) → POST .../buy.
  - cancelLot(id) → DELETE.
- Query-keys в существующий query-keys.ts: ['exchange','lots',
  filters], ['exchange','lot', id].
- Routes в роутере nova:
  - /exchange → ExchangeListPage,
  - /exchange/lots/:id → ExchangeLotPage,
  - /exchange/new → CreateLotPage.
- Пункт меню «Биржа артефактов» в navigation.

### Ф.2. Список лотов (ExchangeListPage)

- TanStack Query useInfiniteQuery cursor-pagination.
- Filters-панель:
  - select по artifact_unit_id (резолв имени из каталога).
  - range price (min/max oxsarit).
  - select по статусу (active/sold/all). Default = active.
  - search seller (опционально).
  - debounce 300ms.
- Список лотов как карточки или таблица (выбери стиль nova).
- Каждая карточка: имя артефакта, qty, price_oxsarit, seller,
  expires_in. Кнопка «Купить» → modal с Idempotency-Key либо
  navigate на детали.
- Пустое состояние: «Лотов нет, [создайте первый]».
- Кнопка «Создать лот» → /exchange/new.

### Ф.3. Детали лота (ExchangeLotPage)

- TanStack Query useQuery(['exchange','lot', id]).
- Полная информация: артефакт (имя, описание из каталога), qty,
  price_oxsarit, seller (username + ссылка на профиль),
  created_at, expires_at relative.
- Кнопка действия:
  - Если seller==current_user — «Отозвать» (DELETE с confirm-modal).
  - Иначе — «Купить» (modal confirm с Idempotency-Key, баланс
    оксаритов проверяется перед отправкой).
- После успеха — query invalidation + Toast + navigate('/exchange').
- Обработка ошибок:
  - 402 ErrInsufficientOxsarits → «Недостаточно оксаритов».
  - 409 lot уже sold/cancelled → invalidate + redirect.
  - 503 → Toast «Сервис временно недоступен».

### Ф.4. Создание лота (CreateLotPage)

- Форма:
  - select artifact_unit_id из доступных артефактов игрока
    (GET /api/artefacts?state=held — посмотри реальный endpoint в
    openapi.yaml; если другое имя — адаптируй).
  - input quantity (max = quantity available, max 100).
  - input price_oxsarit (positive int).
  - select expires_in_hours: 1/6/24/72/168.
- Live-валидация: quantity и price > 0, quantity ≤ available.
- POST /api/exchange/lots с Idempotency-Key (UUID на каждую сессию
  формы).
- Обработка ошибок:
  - 422 ErrPriceCapExceeded → «Цена превышает рыночную в 10×».
  - 422 ErrMaxActiveLots → «У вас уже max активных лотов (10)».
  - 422 ErrMaxQuantity → «Максимум 100 штук в одном лоте».
  - 400 — общая валидация.
- На успех — navigate('/exchange/lots/{newId}') + Toast.

### Ф.5. Тесты

Vitest + React Testing Library:
- ExchangeListPage.test.tsx — рендер списка, debounce filters,
  infinite scroll вызов next page.
- ExchangeLotPage.test.tsx — happy buy, ошибка 402, отзыв для
  seller'а.
- CreateLotPage.test.tsx — форма validation, успешная отправка,
  ошибка 422 max active lots.

Минимум 6 тестов.

### Ф.6. i18n (R12, КРИТИЧНО)

- Grep projects/game-nova/configs/i18n/{ru,en}.yml на
  exchange|биржа|лот|trade|sell|buy_artefact|sell_artefact.
  План 68 уже добавил 12+ ключей — переиспользуй максимально.
- Новые ключи только для UI-специфики: titles экранов,
  empty-states, button labels, confirmation messages.
- В коммите указать **переиспользовано/новых**, цель ≥ 95%.

### Ф.7. Финализация

- Шапка плана 76 ✅ (Ф.1-Ф.6).
- Запись итерации в docs/project-creation.txt.
- В docs/research/origin-vs-nova/nova-ui-backlog.md:
  U-001 → backend+UI ✅. План 68 уже backend ✅, теперь и UI.
- X-017, X-020 (если касались биржи) → ✅.

══════════════════════════════════════════════════════════════════
ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА R0-R15 — см. блок rules-r0-r15.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/rules-r0-r15.md

Особо важно:
- R0: nova-баланс не меняем. Биржа — общий знаменатель, фича во всех
  вселенных, R0 не нарушен.
- R5: nova-стиль, НЕ pixel-perfect.
- R9: Idempotency-Key на createLot и buyLot.
- R12: grep i18n сначала, цель 95%.
══════════════════════════════════════════════════════════════════

══════════════════════════════════════════════════════════════════
GIT-ИЗОЛЯЦИЯ ОТ ПАРАЛЛЕЛЬНЫХ СЕССИЙ — см. блок git-isolation.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/git-isolation.md

Свои пути для CC_AGENT_PATHS:
- projects/game-nova/frontends/nova/src/features/exchange/
- projects/game-nova/frontends/nova/src/api/exchange.ts
- projects/game-nova/frontends/nova/src/api/query-keys.ts (только новые ключи)
- projects/game-nova/frontends/nova/src/main.tsx или router-файл (только новые routes)
- projects/game-nova/frontends/nova/src/i18n/ru.ts (только exchange.* ключи)
- projects/game-nova/frontends/nova/src/i18n/en.ts (только exchange.* ключи)
- docs/plans/76-remaster-nova-frontend-exchange-ui.md
- docs/research/origin-vs-nova/nova-ui-backlog.md (U-001 closure)
- docs/active-sessions.md
- docs/project-creation.txt (запись итерации)

ВНИМАНИЕ: НЕ трогай:
- projects/game-nova/frontends/origin/ — это origin-фронт (план 72 Ф.3).
- projects/game-nova/backend/ — backend закрыт планом 68.
- projects/game-nova/api/openapi.yaml — должна быть готова после плана 68.
══════════════════════════════════════════════════════════════════

КОММИТЫ:

1-2 коммита:

1) feat(exchange,frontend): UI биржи в nova (план 76)
   — все 3 экрана + тесты + i18n + финализация в одном коммите
   (~600-1000 строк, типичный объём frontend-фичи).

ИЛИ если объём > 1500 строк:
1) feat(exchange,frontend): list + lot UI биржи (план 76 Ф.1-Ф.3)
2) feat(exchange,frontend): create lot + tests + финализация (Ф.4-Ф.7)

Trailer: Generated-with: Claude Code
ВСЕГДА с двойным тире:
git commit -m "..." -- $CC_AGENT_PATHS

ЧЕГО НЕ ДЕЛАТЬ:

- НЕ трогай origin-фронт.
- НЕ меняй backend.
- НЕ меняй openapi.yaml (если не хватает поля — simplifications.md).
- НЕ хардкодь тексты — i18n.
- НЕ забывай Idempotency-Key.
- НЕ забывай про -- в git commit.

УСПЕШНЫЙ ИСХОД:

- 3 экрана работают (/exchange, /exchange/lots/:id, /exchange/new).
- typecheck + build + tests зелёные.
- i18n: 95%+ переиспользование.
- Шапка плана 76: все Ф.1-Ф.6 ✅.
- U-001 в nova-ui-backlog → backend+UI ✅.
- Запись в docs/project-creation.txt.
- Удалена строка из docs/active-sessions.md.

Стартуй.
```
