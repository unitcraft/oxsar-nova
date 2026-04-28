# Roadmap Report — декомпозиция в будущие планы (Ф.4+Ф.5)

**Дата**: 2026-04-28
**Контекст**: артефакт плана 62, итоговая сводка. На основе журнала
D-NNN (46 записей), U-NNN (15), X-NNN (22), S-NNN (55) и
alien-ai A1-A14 — даёт **серию будущих планов 63+** для
ремастера origin на nova-backend.

**Стратегия принята до плана 62**: ремастер origin на nova-backend
с pixel-perfect клоном UI. Этот файл — не выбор стратегии, а
**раскладка её на конкретные планы**.

> **Обновление 2026-04-28 (план 75)**: путь `projects/game-origin/`
> освобождён под новый React-фронт ремастера. Текущая legacy-PHP
> реализация переименована в `projects/game-origin-php/`. Серия
> планов 64-74 пишется уже с правильными путями: новый фронт →
> `projects/game-origin/frontend/`, ссылки на legacy →
> `projects/game-origin-php/...`.

---

## Часть I. Сводка по объёму

### По расхождениям

| Категория | Записей | Обязательные | Опциональные |
|---|---|---|---|
| D-NNN (журнал) | 46 | 28 | 18 |
| U-NNN (UI-функции) | 15 | 10 | 5 |
| X-NNN (UX-микрологика) | 22 | 12 | 10 |
| S-NNN (экраны origin) | 55 | 50 (для прода) | 5 (admin/dev) |
| A1-A14 (AlienAI) | 14 | 10 | 4 |

### По объёму работ (сумма)

- **Backend в nova**: ~12-16 недель (origin.yaml override, новые модули,
  расширение event-loop, alien AI до полного, биржа, телепор,
  3 описания альянса, гранулярные ранги, global mail, и т.п.)
- **Frontend (origin pixel-perfect клон)**: ~12-16 недель (55
  экранов на React + воссоздание layout/themes из game-origin)
- **CI / тестирование (screenshot-diff)**: ~2-3 недели
- **Deploy / DNS / config (origin universe)**: ~1 неделя

**Итого**: 27-36 недель (6-9 месяцев) команды из 1-2 разработчиков.
Часть фронта и бэка можно параллелить.

---

## Часть I.5. Сквозные правила реализации (применять во ВСЕХ планах 64-74)

Эти правила — обязательны для каждого плана серии. В отдельных планах
**не дублируются**, но при реализации соблюдаются. Если правило
требует исключения — обосновать в самом плане явно.

### Правило R0. Геймплей nova ЗАМОРОЖЕН (фундаментальное)

**Текущая механика, баланс, экономика, формулы боя, формулы
производства, прогрессия, RF-таблица, стоимости, времена, очки,
дроп, ачивки, AI и весь остальной геймплей game-nova
(uni01/uni02 и любые modern-вселенные) — не меняются** при работе
над серией планов 64-74.

Это правило перекрывает любые соблазны «попутно подправить»:

- ❌ «В origin Bomber RF=20, в nova=12 — давай вернём 20» — **нет**.
  План 18 ребаланса остаётся в силе для modern-вселенных.
  Нужен RF=20 — он попадает в `configs/balance/origin.yaml` для
  вселенной origin, дефолт nova остаётся 12.
- ❌ «В origin температура влияет на производство водорода — добавим
  это и в nova» — **нет**. Если эту механику не было в nova, она
  не появляется в modern-вселенных. Только в origin (через
  override / `internal/origin/economy/`).
- ❌ «Раз уж переписываем event-loop под origin, заодно поправлю
  formula атаки в nova» — **нет**. Если изменение касается
  modern-вселенных, оно требует **отдельного плана** вне серии
  64-74 с обоснованием.
- ❌ «Унифицируем достижения: возьмём список из origin, заменим
  nova-набор» — **нет**. Достижения origin идут в свой набор;
  nova-достижения остаются.

**Что разрешено:**

- ✅ Добавлять **новые механики** в nova-backend, которые
  применимы ко всем вселенным как **общий знаменатель** (биржа
  артефактов, расширенная дипломатия, телепорт планеты,
  achievements engine), но **поведение этих механик в modern-
  вселенных** определяется отдельно, не копируется из origin.
- ✅ Параметризовывать существующие коды-пути для поддержки
  origin-чисел через `LoadFor(universeCode)` + override-YAML.
- ✅ Рефакторить код nova под per-universe configurability,
  если это **не меняет наблюдаемое поведение** для
  modern-вселенных (тесты nova должны оставаться зелёными).
- ✅ Ремастер origin — приводит origin-вселенную к классическому
  oxsar2-балансу. modern-вселенные свой баланс не трогают.

**Why R0:**

- nova — это уже работающий продукт с принятыми решениями
  (планы 17/18/20/21 ребалансы, план 03 экономика, план 09
  event-system). Перепроектировать их посреди ремастера = двойной
  риск, два эпика в одном.
- Игроки, которые придут на uni01 — приходят на современный nova.
  Они не ждут классику. Если хотят классику — выбирают origin.
- Расхождения origin vs nova не означают «nova ошиблась» — это
  **дизайн-решение**: разные вселенные = разные правила.

**Применение:**

Каждый план серии 64-74, прежде чем что-то менять в game-nova,
должен явно ответить:
- Это изменение применимо к **modern-вселенным**? Если да — вне
  scope серии, нужен отдельный план.
- Это изменение применимо **только к origin**? Тогда — через
  override-файл / `internal/origin/`-модуль / условие `if
  universeCode == "origin"`.
- Это **нейтральная инфраструктура** (loader, event-handler
  scaffolding, миграция схемы под новые опциональные поля)?
  Допустимо, при условии что modern-тесты остаются зелёными.

### Правило R1. Имена в БД и Go — современный nova-стиль, не калька с origin

**Когда применяется**: при создании любых новых полей, таблиц,
колонок, констант, JSON-ключей, OpenAPI-схем в рамках планов 64-74.

**Принцип**: origin/legacy-имена служат **семантическим референсом**
(«что хранится»), но **не источником конкретных идентификаторов**.
Именуем по конвенциям nova:

| Источник в origin | Стиль origin | Стиль nova (правильно) |
|---|---|---|
| `na_user.userid` | venda с префиксом таблицы, lower no-underscore | `users.id` (UUIDv7), без префикса таблицы в колонке |
| `na_planet.galaxy/system/planet_position` | три отдельные `tinyint`/`smallint` | `planets.coords` (composite type) или `coords_galaxy/coords_system/coords_position` (snake_case с осмысленным префиксом) |
| `referer_id` (orig: invited_by) | непоследовательно | `referrer_id` (правильное английское написание) |
| `tag` (alliance) | плоское поле | `tag` ок, но добавить `tag_normalized` для поиска |
| `chargeMetal`, `prodMetal` (PHP-выражения как строка) | varbinary(255) с DSL | предвычисленные числа в YAML/таблицу `building_costs` (плоско, без DSL) |
| `holding_at`, `holding_until` (alien) | timestamp без таймзоны | `holds_until_at` (с `_at` суффиксом по nova-конвенции, `TIMESTAMP WITH TIME ZONE`) |
| `cnt`/`num`/`count` | сокращения | `count`, `quantity`, `amount` — полные слова |
| `is_*` boolean | смесь `is_x` и `x_flag` | строго `is_x` или `has_x` (по семантике) |

**Конкретные конвенции nova** (закреплены практикой game-nova):

1. **snake_case** для SQL (колонки, таблицы) и JSON-ключей в API.
2. **camelCase** для Go-полей (через `json:"snake_case"` теги) и TS.
3. **Полные слова** вместо сокращений (`count` not `cnt`,
   `position` not `pos`, `quantity` not `qty`, `description` not
   `desc`).
4. **Без префикса таблицы в колонках** (`users.id`, не
   `users.user_id`). Если нужен FK, тогда `users.alliance_id`,
   `planets.user_id` — префикс по сущности, на которую ссылаемся.
5. **`_at` для timestamp**: `created_at`, `updated_at`, `deleted_at`,
   `expires_at`, `holds_until_at`. Всегда `TIMESTAMP WITH TIME ZONE`
   (Postgres `TIMESTAMPTZ`).
6. **`_id` для FK**: `user_id`, `planet_id`, `alliance_id`.
   Тип — `UUID` (UUIDv7 предпочтительно). Не `int`/`bigint` для
   новых таблиц.
7. **`is_*` / `has_*`** для boolean: `is_open`, `is_deleted`,
   `has_premium`. Не `*_flag`, не `*_yn`.
8. **Числа без суффиксов единиц** в имени колонки если единица
   очевидна из домена. `oxsar` (не `oxsar_amount`), `metal` (не
   `metal_qty`). Исключение — когда смысл неоднозначен:
   `expires_in_seconds` понятнее, чем `expires_in`.
9. **Английский язык**, без транслита. `oxsar` — название валюты,
   а не транслит. `expedition`, не `expediciya`.
10. **Множественное число для таблиц**: `users`, `planets`,
    `alliances`. Единственное для type-таблиц (`unit_type`).

**Что НЕ переносим из origin**:
- Префикс таблиц `na_` (это namespace из MySQL legacy oxsar2).
- DSL-формулы как строки (`varbinary(255)`) — конвертируем в числа.
- Магические числа в enum-полях (`mode=1`/`2`/`3`) — заменяем на
  читаемые `unit_type` enum-значения.
- Колонки-флаги типа `del_flag`, `act_flag` — заменяем на
  `is_deleted`, `is_active`.
- Кириллицу в именах (если найдётся) — английский всегда.

**Применимость**:
- К **новым** таблицам/колонкам, создаваемым в планах 64-74 — да,
  обязательно.
- К **существующим** таблицам в game-nova — оставляем как есть
  (миграция имён — отдельная задача, не в рамках этих планов).
- К данным, мигрируемым из origin (если когда-либо понадобится) —
  при импорте мапим origin-имена → nova-имена в самой миграции.
  Внутри nova хранится в nova-стиле.

**Если есть сомнение** — смотри текущие nova-таблицы в
`projects/game-nova/backend/migrations/` и `internal/<domain>/` для
прецедента. Если прецедента нет — выбери понятнее и оставь
комментарий в миграции почему.

**Особый случай: игровая валюта.**

Все новые поля под валюту — **строго по ADR-0009 / плану 58**, а
НЕ по legacy `na_user.credit`. Двухвалютная модель уже зафиксирована
архитектурно, и серия 64-74 не должна её ломать.

| Сущность | Где живёт | Колонка | Тип |
|---|---|---|---|
| Hard premium (за рубли) | `billing.wallets` | `oxsar` | `bigint NOT NULL DEFAULT 0` |
| Soft premium (per universe) | `game-nova.users` | `oxsarit` | `bigint NOT NULL DEFAULT 0` |
| Журнал hard-операций | `billing.wallet_transactions` | `oxsar` (signed) + `kind`, `idempotency_key`, `metadata jsonb`, `created_at` | — |
| Журнал soft-операций | `game-nova.oxsarit_transactions` | `oxsarit` (signed) + `kind`, `source jsonb`, `created_at` | — |
| Реферальные выплаты | — | `referral_payouts.oxsarit` (только soft, по плану 59) | — |

**Конкретные правила для валюты:**

- Имена колонок — `oxsar` / `oxsarit` без постфиксов
  (НЕ `oxsar_amount`, НЕ `oxsar_balance`, НЕ `credit`).
- Тип — `bigint` (целые), не `numeric`/`float`.
- Цены премиум-фич в конфигах — с префиксом валюты:
  `officers.yaml: oxsar_price`, не просто `price`.
- Hard и soft журналируются **отдельно** (юр-разделение): hard в
  billing-сервисе под ст. 437 ГК, soft в game-nova под ст. 1062 ГК.
  НЕ смешивать в одной таблице.
- Перевод origin `users.credit` (если миграция данных понадобится)
  → `users.oxsarit` (по плану 58 Ф.var-B). НЕ создавать колонку
  `credit` в новых таблицах под legacy-привычку.

**Источники истины** (читать перед работой над валютой):
- [docs/adr/0009-currency-rebranding.md](../../adr/0009-currency-rebranding.md)
  — архитектурное решение + полная таблица колонок.
- [docs/plans/58-currency-rebranding.md](../../plans/58-currency-rebranding.md)
  — миграционный план + smart-pay механика.
- [docs/plans/59-referral-program.md](../../plans/59-referral-program.md)
  — реферальная программа в оксаритах.

### Правило R2. OpenAPI как источник истины

Любой новый endpoint = сначала схема в `projects/game-nova/api/openapi.yaml`,
потом сгенерированные клиенты на фронте, потом реализация на backend.
Не наоборот.

### Правило R3. Логирование обязательно

Любой новый event-handler / service-метод в Go использует `log/slog`
с полями `user_id`, `planet_id`, `event_id`, `trace_id` (см. CLAUDE.md
§Go). Без этого PR не принимается.

### Правило R4. Тесты — golden + property-based для боя/экономики

Любые правки в `battle/`, `economy/`, `event/` требуют golden-тестов
и (где применимо) property-based через rapid. Покрытие изменённых
строк ≥ 85%.

### Правило R5. Frontend pixel-perfect (только для плана 72)

Фронт origin воспроизводит legacy визуально 1:1 (CSS, layout, цвета,
шрифты) на новом React-стеке. Новые UX-улучшения — отдельная
итерация после старта.

### Правило R6. Новые API проектируем с нуля по современным стандартам

Когда план серии 64-74 вводит новый endpoint — origin-API служит
**семантическим референсом** (что эта операция делает в игре), но
**не источником конкретных URL/методов/параметров**. Проектируем
API по нынешним индустриальным стандартам, не калькой с
`?go=Page&action=...`.

**Конкретные требования:**

1. **REST-стиль, ресурсо-ориентированный**:
   - origin: `?go=Mission&action=send` → nova: `POST /api/fleet/missions`
   - origin: `?go=Constructions&action=build&id=N` → nova: `POST /api/planets/{planetId}/buildings/{buildingId}/queue`
   - Ресурсы во множественном числе (`/missions`, `/buildings`),
     ID в path, не query.
2. **HTTP-методы по семантике**:
   - GET — чтение без побочных эффектов.
   - POST — создание ресурса / нерезидентное действие
     (`/cancel`, `/repair`).
   - PUT/PATCH — обновление существующего ресурса.
   - DELETE — удаление.
   - НЕ использовать GET для действий с побочными эффектами
     (origin часто грешит этим).
3. **JSON / OpenAPI первым** (R2). HTML-фрагменты не возвращаем.
4. **Версионирование**: `/api/v1/...` если уже есть, иначе
   подкатегория-namespace (`/api/exchange/lots`,
   `/api/alliances/{id}/diplomacy`). Решение — по существующему
   стилю nova-API (см. `projects/game-nova/api/openapi.yaml`).
5. **Идемпотентность платежей** (R1 валюта + план 38) — через
   `Idempotency-Key` header для всех операций со списанием
   ресурсов или валюты.
6. **Pagination**: cursor-based (через `?cursor=...&limit=...`),
   не offset/page для списков с >100 элементов. Для коротких
   списков можно без пагинации.
7. **Errors**: единый формат
   `{"error": {"code": "ERROR_CODE", "message": "..."}}`,
   как в существующих nova-handler'ах. HTTP-коды по семантике
   (400 валидация, 401 auth, 403 forbidden, 404 not found,
   409 conflict, 429 rate-limit, 500 server-error).
8. **Authorization** — Bearer JWT всегда, никаких legacy-cookie
   (план 63 RFC 6749).
9. **Имена параметров** — snake_case в JSON, snake_case в query
   (consistency с БД и SQL).
10. **Никаких backend-адаптеров под legacy-имена.** origin-фронт
    (план 72) сразу пишется на новые API без обёрток.

**Источник правил**: фактически — это **R2** (OpenAPI первым) +
дополнительные конкретные практики. Origin-API смотрим только
чтобы понять «что делает эта операция игроку», саму форму API
проектируем заново.

### Правило R7. Backward compat технических интерфейсов — N/A

В проде нет игроков (план 36 «0 живых юзеров»; план 62 ещё раз
подтвердил). До публичного запуска (плана 74) **технические
интерфейсы** можно ломать смело: API-схемы, БД-миграции,
конфиг-форматы, имена полей DTO, payload event'ов, сигнатуры
Go-методов. Без feature-flag'ов и миграционных мостов. Удалять
старый код сразу, не оставлять «// removed in plan N»
комментариев.

**Это правило НЕ распространяется на геймплей** — он зафиксирован
правилом **R0** (геймплей nova заморожен). R7 не отменяет R0:
«можно ломать API» **не означает** «можно поменять формулу боя
или RF-таблицу в nova».

Что разрешено по R7:
- ✅ Поменять структуру JSON-ответа `/api/fleet/missions` без
  поддержки старой версии.
- ✅ Переименовать колонку БД, удалить старую (если не нужна).
- ✅ Изменить сигнатуру Go-метода `internal/fleet/...`.
- ✅ Поменять формат event-payload в JSONB колонке.

Что **НЕ разрешено** по R7 (попадает под R0):
- ❌ Поменять число в `configs/units.yml` для modern-вселенных.
- ❌ Поменять формулу `economy.MetalProduction()` в nova.
- ❌ Удалить механику nova (даже если она «лишняя» по сравнению
  с origin).
- ❌ Поменять баланс боя «попутно с рефактором event-loop'а».

После публичного запуска (плана 74) — R7 теряет силу: вводится
backward compat / deprecation policy для API. R0 продолжает
действовать всегда (если нужно поменять nova-баланс — отдельный
ребаланс-план с ADR).

### Правило R8. Метрики Prometheus + структурированные логи

В дополнение к R3 (slog) — каждый новый event-handler / service-метод
/ endpoint регистрирует **метрики Prometheus**:

- `counter` для подсчёта событий (вызовов / ошибок / успехов).
- `histogram` для длительностей (latency обработки события /
  endpoint'а).

Особо важно для:
- event-loop (план 65) — видеть load, длительность handler'ов,
  очередь событий.
- AlienAI (план 66) — видеть тики, длительности AI-итераций.
- Биржа артефактов (план 68) — операции / latency / провалы.

Имена метрик — snake_case с префиксом `oxsar_<domain>_` (по
существующему стилю game-nova). Подробнее — план 09 Ф.2.

### Правило R9. Идемпотентность всех мутирующих API

Любой POST/PUT/DELETE endpoint, который **изменяет состояние**
(создаёт ресурс, тратит валюту/ресурсы, меняет данные игрока),
**должен** поддерживать идемпотентность через
`Idempotency-Key` header.

Закрывает класс багов:
- Двойной клик игрока (повторная отправка флота).
- Retry на сетевом таймауте (платёж не должен списать дважды).
- Парные обработчики event-loop'а (мультивоспроизведение).

Реализация — по образцу billing (план 38):
- Клиент генерирует UUIDv7 в header `Idempotency-Key`.
- Сервер хранит `idempotency_key` + `response_hash` в журнальной
  таблице (например, `wallet_transactions.idempotency_key`).
- При повторном запросе с тем же ключом — возвращается
  тот же ответ без повторного выполнения.

GET-endpoint'ы освобождены (по определению идемпотентны).

### Правило R10. Per-universe изоляция данных

Любая таблица БД с **per-universe данными** (планеты, флоты, юниты,
здания, события, отчёты, чат, альянсы) **должна содержать колонку
`universe_id`** (UUID FK на `universes.id`).

Любой SELECT/UPDATE/DELETE по такой таблице **должен** фильтровать
по `universe_id`:

```go
// ❌ Плохо — может утечь между вселенными:
db.Exec("UPDATE planets SET ...")

// ✅ Хорошо — изоляция per-universe:
db.Exec("UPDATE planets SET ... WHERE universe_id = $1", uniID)
```

Защита от случайной утечки данных uni01 в origin (или наоборот).
Особенно критично, если игрок имеет аккаунты в двух вселенных
одновременно — данные не должны пересекаться.

Какие таблицы — per-universe (НЕ исчерпывающий список):
`planets`, `fleets`, `buildings`, `units`, `research`, `events`,
`battle_reports`, `espionage_reports`, `expedition_results`,
`chat_messages`, `alliances`, `alliance_members`,
`alliance_applications`, `alliance_relations`,
`oxsarit_transactions`, `goal_progress`.

Какие — НЕ per-universe (cross-universe / global):
`users` (identity-уровень), `wallets` (billing — глобальный hard-кошелёк
по ADR-0009), `wallet_transactions`, `user_consents`, `user_reports`
(модерация — план 56), `mailboxes` (mail — план 57 единый inbox).

При создании любой новой таблицы в планах 64-74 — явно решить и
задокументировать: per-universe или cross-universe. Если
per-universe — добавить FK + index на `universe_id` сразу.

### Правило R11. Анти-абуз rate-limiting

Любой публичный endpoint, который игрок может **спамить**
(особенно UGC: чат, отчёты, заявки в альянс, поиск, регистрация
событий) — обязан иметь rate-limiting middleware.

Готовые механизмы в game-nova (используем существующие):
- План 46 Ф.4: in-memory rate-limit для чата (10 msg/min).
- План 56: rate-limit для report-API.
- planned: единый middleware на основе планов 46/56 — переиспользуем.

Каждый план серии 64-74, когда вводит публичный endpoint, должен
явно указать:
- Rate-limit (например, `60 req/min/user` или `10 req/min/IP`).
- Что происходит при превышении (HTTP 429 + retry-after header).
- Anti-fraud cap (где применимо — например, биржа план 68 имеет
  cap `EXCH_SELLER_MAX_PROFIT 1000%`).

### Правило R12. Локализация i18n с самого начала

Все user-facing строки (тексты ошибок API, тексты UI, шаблоны
сообщений почты, push-уведомлений, achievements-описаний) идут
через **i18n bundle**, не хардкодятся.

В game-nova существует `projects/game-nova/configs/i18n/` —
переиспользуем структуру.

Конкретные требования:
- ❌ Не хардкодить русские строки в Go-handler'ах:
  `errors.New("неверный пароль")`.
- ✅ Использовать i18n-ключ:
  `errors.New(i18n.Tr(ctx, "auth.invalid_password"))`.
- На старте серии — только русский, но архитектура готова к
  английскому/другим без переписывания кода.
- Имена ключей — snake_case, namespace по домену:
  `auth.invalid_password`, `fleet.mission_too_far`,
  `alliance.tag_taken`.

Это упрощает добавление английского / казахского / узбекского
позже без рефакторинга.

### Правило R13. Observability event-payload'ов

Когда план 65 (или другой) добавляет новый Kind в event-loop'е —
payload **обязательно типизированный** (Go-struct), не сырой
`map[string]any` / `[]byte`.

```go
// ❌ Плохо — payload как непрозрачный JSONB:
type EventRow struct {
    Kind    Kind
    Payload []byte  // что внутри — никто не знает
}

// ✅ Хорошо — typed payload через sqlc / json.Unmarshal в struct:
type AlienAttackPayload struct {
    AttackerFleetID string `json:"attacker_fleet_id"`
    TargetPlanetID  string `json:"target_planet_id"`
    PowerScale      int    `json:"power_scale"`
}
```

JSON Schema валидация при INSERT (можно через Postgres JSONB +
CHECK или Go-валидатор перед вставкой). Если schema нарушена —
ошибка явная, не «упало в handler через час».

Закрывает класс багов: «забыл поле в payload, в проде событие
сериализуется, но handler крашится при десериализации».

### Правило R14. Migration policy для legacy-PHP-данных

Если когда-то понадобится **импортировать данные** из
`projects/game-origin-php/` в nova (например, перенести историю
боёв origin для архивов), маппинг **legacy-PHP-имён → nova-имён**
делается в самой миграции (`up.sql` или импорт-скрипт), не в
Go-runtime коде.

```sql
-- ❌ Плохо: дать nova-коду читать поля legacy-имён:
SELECT na_user.userid AS user_id FROM na_user;

-- ✅ Хорошо: миграция делает rename + transform на этапе импорта:
INSERT INTO users (id, ...) SELECT na_user.userid, ... FROM na_user;
-- После миграции nova-код видит только nova-имена.
```

Внутри nova-runtime — **только nova-стиль** (R1). Legacy-PHP-имена
не утекают за границу импорт-скрипта.

На текущей стадии (0 игроков) — миграция данных не требуется.
Правило заранее, на случай если в будущем понадобится перенос.

### Правило R15. Делаем без упрощений, как для прода

Каждый план серии 64-74 реализуется в **прод-готовом** качестве:
полные тесты, обработка ошибок, метрики, безопасность,
производительность, документация. **Никаких** «MVP-сокращений»,
«TODO: позже», «работает на dev, в проде доделаем», «упростим до
запуска».

**Что это означает на практике:**

- ✅ **Тесты с самого начала**, не «потом добавим».
  Покрытие изменённых строк ≥ 70% (R4 для домена ≥ 85%).
  Новый код без тестов в коммит не идёт.
- ✅ **Обработка ошибок везде**, не «happy path работает,
  edge case'ы потом». Все error-paths покрыты тестами.
- ✅ **Метрики Prometheus** (R8) — со старта, не «когда заметим
  проблемы в проде».
- ✅ **Идемпотентность** (R9) — со старта на всех мутирующих
  endpoint'ах, не «когда увидим двойные платежи».
- ✅ **Per-universe изоляция** (R10) — со старта на всех новых
  таблицах, не «когда заметим утечки».
- ✅ **Rate-limiting** (R11) — со старта на всех публичных
  endpoint'ах, не «когда придёт первый спам-бот».
- ✅ **i18n** (R12) — со старта, не «пока хардкод, потом
  переведём».
- ✅ **Документация** API в OpenAPI (R2) — синхронно с кодом,
  не «потом задокументируем».
- ✅ **Производительность**: индексы БД на FK и часто-фильтруемые
  колонки, разумный N+1, batch-операции где нужно. Не «оптимизируем
  потом».
- ✅ **Безопасность**: валидация всех inputs, защита от SQL
  injection (через sqlc / параметры), CSRF где применимо, JWT-проверка
  scopes/roles. Не «дыры замажем после запуска».

**Что это НЕ означает** (избегать gold-plating):

- ❌ Не делаем «избыточно красивые» API ради красоты — функциональная
  достаточность важнее эстетики.
- ❌ Не вводим инфраструктуру, которой не пользуемся (например,
  observability через Jaeger без необходимости — slog + Prometheus
  достаточно).
- ❌ Не оптимизируем то, что не является узким местом (premature
  optimization).
- ❌ Не покрываем случаи «вдруг кто-то когда-то» (YAGNI).

**Прод-качество ≠ over-engineering.** Граница: то, что **точно
понадобится** при наличии реальных игроков (тесты, метрики,
изоляция, rate-limit) — делаем со старта. То, что **может
понадобиться** в будущем (Jaeger, GraphQL, sharding) — не делаем
без явной необходимости.

**Если есть упрощение** — оно явно фиксируется в
[`docs/simplifications.md`](../../simplifications.md) с
обоснованием **причины** и **планом возврата**. Без этого правила
упрощение считается **багом**, не «MVP-вариантом».

**Why R15:**

- Серия 64-74 готовит вселенную к **публичному запуску** (план 74).
- Игроки origin приходят с ожиданием от ремастера качества —
  «не хуже, чем nova».
- Технический долг, накопленный в виде упрощений, превращается в
  кризисы поддержки после запуска (см. историю плана 9 — фазы
  надёжности и наблюдаемости вводились **после** того, как стало
  больно; здесь делаем сразу).

---

## Часть II. Декомпозиция в будущие планы

Нумерация ориентировочная — согласовать с текущим состоянием
docs/plans на момент старта.

### План 64: origin.yaml override + per-universe balance loading

**Что**: параметризация балансовых констант nova через
override-схему. Modern-вселенные работают на дефолтных YAML без
изменений. Origin получает override-файл. Все 🟡 расхождения
D-NNN из категории формула.

**Содержит**:
- `configs/balance/origin.yaml` — override с числовыми
  предвычисленными значениями charge_*, basic_*, prod_* для всех
  buildings/units/research/defense origin
- Парсер origin-формул в Go (для предвычисления)
- Расширение `internal/balance/` с функциями `LoadDefaults()` и
  `LoadFor(universeCode)` (deep-merge override поверх дефолта)
- Без изменений в БД — идентификация по `universes.code`

**Закрывает**: D-026, D-027 (RF алиенов в origin.yaml), D-028
(спец-юниты), D-030 (per-building cost_factor), D-022 (prod_factor).

**Объём**: 2 недели.
**Зависит от**: ничего.
**Блокирует**: 65, 66.

---

### План 65: Расширение event-loop (legacy события)

**Что**: реализовать недостающие Kind-ы и расширить существующие.

**Содержит**:
- Implement KindDemolishConstruction (D-031, D-NNN объявлен но без
  handler)
- Implement KindDeliveryUnits, KindDeliveryResources,
  KindDeliveryArtefacts (D-035)
- Implement KindStargateTransport / KindStargateJump (план 20 Ф.5)
- Implement KindAttackDestroyBuilding, KindAttackAllianceDestroyBuilding (D-037)
- Implement KindTeleportPlanet (D-032, U-009)
- Implement KindArtefactDisappear
- (опц.) KindRunSimAssault (D-034)
- Pruner и идемпотентность всех новых handler'ов

**Закрывает**: D-031..D-037, частично U-009.

**Объём**: 3-4 недели.
**Зависит от**: 64 (origin.yaml — для balance numbers).

---

### План 66: AlienAI до полного паритета с legacy

**Что**: достроить план 15 этап 3 — реализовать все 8 EVENT_ALIEN_*
с полным AI-движком из origin (1127 строк → ~800 строк Go).

**Содержит** (см. `alien-ai-comparison.md`):
- Реализовать `KindAlienFlyUnknown` handler (грабёж/подарок/атака)
- Реализовать `KindAlienGrabCredit` как отдельный сценарий
- Реализовать `KindAlienChangeMissionAI` (control_times, power_scale)
- Расширить `KindAlienHoldingAI` до 8 действий из 2 (с заглушками
  для 6 неактивных, как в origin)
- Алгоритм `generateFleet()` (target_power, итеративное добавление)
- 5 алиен-кораблей UNIT_A_* в `configs/units.yml` под флагом
- Четверг-множитель ×5 / ×1.5..2.0 (вынос в origin.yaml)
- `findTarget` / `findCreditTarget` с критериями
- `shuffleKeyValues` (случайное ослабление техник)
- Платный выкуп удержания

**Закрывает**: D-036, alien-ai-comparison.md A1-A14.

**Объём**: 3 недели.
**Зависит от**: 64 (alien-units в origin.yaml).

---

### План 67: Расширение alliance-системы

**Что**: добавить недостающие фичи альянсов.

**Содержит**:
- 3 описания альянса (`description_external/internal/apply`) — D-041, U-015
- Передача лидерства (`abandonAlly`) — D-040, U-004
- Гранулярные права рангов (`alliance_ranks` таблица с
  permissions JSONB) — D-014, U-005
- Полнотекстовый поиск альянсов с фильтрами — U-012
- Альянсный лог активности (`alliance_audit_log`) — U-013
- (custom logo альянса U-011 — отдельный план или после
  storage/moderation)

**Закрывает**: D-014, D-040, D-041, U-004, U-005, U-012, U-013, U-015.

**Объём**: 2-3 недели.
**Зависит от**: ничего критичного.

---

### План 68: Биржа артефактов (Exchange/Stock)

**Что**: новая cross-universe фича — player-to-player биржа
артефактов. Главное расхождение с origin (D-039 — 3 контроллера,
1220+757+850 строк PHP в origin → новый модуль в nova).

**Содержит**:
- `internal/exchange/` модуль (~2000 строк Go)
- 5+ endpoint'ов:
  - `GET /api/exchange/lots` (список с фильтрами)
  - `POST /api/exchange/lots` (создать)
  - `GET /api/exchange/lots/{id}` (детали)
  - `POST /api/exchange/lots/{id}/buy`
  - `DELETE /api/exchange/lots/{id}` (отозвать)
  - `GET /api/exchange/stats` (статистика)
- БД-схема: `exchange_lots`, `exchange_history`
- Event-loop: KindExchExpire, KindExchBan
- Premium-механика (Знак торговца — артефакт)
- Frontend: 3 экрана (список, детали, создание)

**Закрывает**: D-039, U-001, X-017, X-020.

**Объём**: 3-4 недели.
**Зависит от**: ничего критичного (можно параллелить).

---

### План 69: Расширение domain-полей в nova

**Что**: миграции для legacy-полей пользователя.

**Содержит**:
- `users.max_points`, опц. `dm_points`, `be_points`, `of_points` (D-001)
- `users.protected_until` (D-004) + проверки в attack
- `users.is_observer` или role 'observer' (D-005)
- `users.profession_changed_at` (D-008)
- `users.race` + `configs/races.yml` (D-021)
- `users.last_global_chat_read_at`, `last_ally_chat_read_at`,
  `chat_language` (D-020)
- `users.home_planet_id` (D-019)
- `users.last_planet_teleport_at` (D-016)
- `users.account_deletion_scheduled_at` (D-003 для origin)
- `users.ui_theme`, `ui_pack` (D-007)

**Закрывает**: D-001, D-003 (для legacy), D-004, D-005, D-007,
D-008, D-019, D-020, D-021.

**Объём**: 2 недели (миграции + handler updates).

---

### План 70: Achievements расширение (legacy + общий движок)

**Что**: расширить goal engine под ачивки origin.

**Содержит**:
- Загрузка ~100 ачивок из `na_achievement_datasheet` в `configs/goals.yml`
- Расширение `goal_defs` под условия типа `req_points`,
  `req_u_points`, `bonus_metal`, `bonus_*_unit`
- UI: `frontend/src/features/achievements/` с прогрессом и
  раскрытием полным условий (как в origin)

**Закрывает**: D-017.

**Объём**: 1-2 недели.
**Зависит от**: goal engine (уже реализован в nova).

---

### План 71: UX-микрологика origin → nova-frontend

**Что**: применить X-NNN записи на nova-frontend (для всех
вселенных, не только origin).

**Содержит** (приоритеты):
- ⭐ X-001 (дефицит ресурсов с скобками `(нужно X)`),
- ⭐ X-003 (показ требований при `can_build = false`),
- ⭐ X-010 (энергодефицит красным),
- X-002 (потребление красным),
- X-013 (added_level +/- зелёное/красное),
- X-021 (счётчик новых ачивок),
- X-014 (ремонтные поля),
- X-007 (нет слотов с подсчётом),
- X-008 (статус артефактов),
- X-009 (расширенный helptip),
- остальные 12 X-NNN

**Закрывает**: X-001..X-022.

**Объём**: 2-3 недели.
**Зависит от**: ничего.

---

### План 72: Origin-фронт — pixel-perfect клон (главный)

**Что**: новый Vite-bundle `projects/origin-frontend/` —
pixel-perfect воспроизведение visual style game-origin на React.

**Содержит**:
- Bootstrap проекта (Vite + TS + TanStack Query + Zustand + TipTap)
- Воссоздание layout (3-frame: leftMenu + main + header)
- Перенос ассетов (icons, themes, colors из public/css/, images/)
  с проверкой лицензий
- Реализация всех 50 prod-экранов (S-001..S-050) на nova-API:
  - **Spring 1**: Main, Constructions, Research, Shipyard,
    Galaxy, Mission, Empire, Empire (~7 экранов)
  - **Spring 2**: Alliance (12 шаблонов), Resource, Market,
    Repair, Battlestats, Fleet operations (~10 экранов)
  - **Spring 3**: Artefacts, ArtefactMarket, ArtefactInfo,
    BuildingInfo, UnitInfo, Techtree, Records, Statistics,
    Achievements, Daily quests (~10 экранов)
  - **Spring 4**: Friends, MSG, Chat, ChatAlly, Notepad,
    Search, Officer, Profession, Settings, Tutorial,
    UserAgreement, Changelog, Support, Widgets,
    AdvTechCalculator (~13 экранов)
  - **Spring 5**: Simulator, RocketAttack, MonitorPlanet,
    ResTransferStats, Stock/Exchange (~5 экранов; зависит от 67)
- Только русский язык в первой итерации
- BBCode чата выкидывается → TipTap
- Адаптив, тёмная тема, новшества — **после старта**

**Закрывает**: S-001..S-055 (кроме admin S-039, S-043, S-044, S-053).

**Объём**: 12-16 недель (3-4 месяца). Самый большой план серии.
**Зависит от**: 64, 65, 66, 67, 68, 69, 70, 71 (вся backend-готовность);
57 (mail/TipTap). Может быть **частично** запущен раньше — экраны,
backend которых уже готов.

---

### План 73: Screenshot-diff CI (Playwright + visual regression)

**Что**: автоматизированное сравнение origin-фронта со
скриншотами эталонного game-origin.

**Содержит**:
- Скрипт снятия эталонов с запущенного game-origin
  (localhost:8092) — все 50 экранов
- Playwright-тесты на новый origin-фронт
- pixelmatch threshold (например, 0.5%)
- CI-job: запускается на PR
- Регламент обновления эталонов при намеренных изменениях

**Закрывает**: пороги качества плана 72 (паритет визуала).

**Объём**: 2 недели.
**Зависит от**: 72 (хотя бы первые экраны).

---

### План 74: origin deploy + DNS + config

**Что**: подъём вселенной origin как третей рядом с uni01/uni02.

**Содержит**:
- DNS / поддомен (имя по ADR-0010 — открытый вопрос)
- Свой Vite-bundle deploy (CDN)
- CORS / `ALLOWED_ORIGINS` расширение
- `universes.code = 'origin'` (или другое из ADR-0010);
  override активируется автоматически наличием
  `configs/balance/origin.yaml`
- Регистрация в registry-системе (план 36)
- Smoke-тесты после деплоя

**Закрывает**: запуск ремастера в проде.

**Объём**: 1 неделя.
**Зависит от**: 72, 73.

---

## Часть III. Зависимости между планами

```
64 (origin.yaml) ──┬──→ 65 (event-loop)
                   ├──→ 66 (AlienAI)
                   └──→ 69 (domain fields)

67 (alliance) ─────────────┐
68 (exchange) ─────────────┤
70 (achievements) ─────────┤
71 (UX-микрологика) ───────┤
                           ▼
57 (mail-service, готов)──→ 72 (origin-фронт) ──→ 73 (CI) ──→ 74 (deploy)
                            ▲
                  53 (admin-bff, готов) — для S-039/S-043/S-044
                  38, 42, 54 (billing, готов) — для S-032 Payment
```

---

## Часть IV. Risk register

| Риск | Митигация |
|---|---|
| Точность баланса при предвычислении формул origin → origin.yaml | Golden-tests на 5+ ключевых уровней каждого здания, сравнение с PHP `eval()` через скрипт |
| Тихая регрессия 🟣 при миграции уни01/uni02 на новые поля D-001..D-025 | Все миграции nullable; CI на текущие fixtures |
| AlienAI расхождения сложно отловить (тестируется днями) | Property-based тесты + golden-логи на 50+ итераций |
| pixel-perfect клон будет «уплывать» при правках UI | Screenshot-diff CI (План 73) — баг сразу виден |
| Шрифты/иконки origin не попадают под лицензию | Аудит лицензий ДО плана 72. Замена несовместимых на open-source аналоги |
| BBCode-исход в чате могут жаловаться legacy-игроки | Документировать как осознанное решение в release notes |
| Лидерство альянса передачи может быть злоупотреблено | Email-подтверждение через identity (как в D-003) |
| Биржа артефактов — risc P2W если без лимитов | Премиум-маркер (Знак торговца) + cap на цену (EXCH_SELLER_MAX_PROFIT 1000% из legacy) |
| Производительность event-loop при 51 типе событий | Уже решено в плане 09 (адаптивный воркер до 1000/цикл) |

---

## Часть V. Что НЕ делать (явный отказ)

| Фича | Решение | Причина |
|---|---|---|
| BBCode-чат | выкидывается, заменяется TipTap | Уже принято — план 57 |
| 6 заглушек HOLDING_AI (Repair/AddUnits/...) | не реализуем в nova | В origin они тоже no-op |
| Officer-юниты как боевые | оставляем nova-модель (subscription) | D-015 — устаревшая legacy-механика |
| `delete INT(10)` auto-deletion в users | для всех вселенных через email-коды | D-003 — безопаснее |
| `templatepackage` per-user тёмные темы для origin | `users.ui_theme` enum, не свободная строка | UGC-мрак |
| Турниры (D-031, U-002) в первой итерации | отдельный план **после** плана 74 | Не блокирует ремастер |
| Кастомные logo альянса (U-011) | отдельный план после storage/moderation | UGC требует инфраструктуры |
| EditConstruction / EditUnit / TestAlienAI | переход в admin-frontend (план 53) | dev/admin only |
| **Ачивки в origin-фронте** (S-Achievements) | **не реализуем в первой итерации**; в nova ачивки уже есть, для origin реализуем позже отдельным планом | Сократить scope плана 72; план 70 (расширение goal engine под классические ачивки oxsar2) **отложен** до пост-запуска origin |
| **Туториал в origin-фронте** (S-Tutorial) | **не реализуем в первой итерации**; реализуем позже отдельным планом | Сократить scope плана 72; nova имеет свой onboarding, для origin вернёмся к туториалу после старта |
| **Баннеры и рекламные тексты в origin-фронте** | **не переносим** из legacy-PHP | В legacy-PHP origin есть рекламные блоки / баннеры / промо-тексты в шапке/футере/между секциями. В новый фронт **не копируются** — ни визуально, ни функционально. Решение по монетизации origin принимается отдельно после запуска (если будет). На старте origin-фронт чист от рекламы. |

---

## Часть VI. Известные неизвестные (требуют ещё одного round'а)

1. **Имя поддомена origin-вселенной** — `origin.oxsar-nova.ru` /
   `classic.oxsar-nova.ru` / иное. Имя `universes.code` мы
   зафиксировали как **`origin`**, но поддомен решает ADR-0010
   (открытый).
2. **Точные цвета палитры origin** — взять из `style.css`
   программно или вручную. План 72 решит.
3. **Шрифты в origin** — какие именно, лицензии. План 72 решит
   аудитом.
4. **Какие именно из 100+ ачивок origin переносить** — нужен
   отбор приоритетных vs «исторических». План 70 решит.
5. **Лицензии иконок origin** — критичный блокер. Аудит до плана 72.
6. **Чат фан-аут с TipTap-payload** — план 32 готов, но
   протестировать на ~100 одновременных пользователей до плана 72.

---

## Часть VII. Сводный график (Gantt-стиль)

```
Месяц 1-2:   [64 origin.yaml]──┐
                                │
Месяц 2-3:   [65 event-loop]   │  [67 alliance]   [68 exchange]
                                │
Месяц 3-4:   [66 AlienAI]──────┤  [70 achievements]
                                │
Месяц 4-5:   [69 domain fields]┘  [71 UX-микрологика]
                                
Месяц 5-9:   [72 origin-фронт — pixel-perfect клон]
              (5 spring'ов по 3-4 недели каждый)
                                
Месяц 9-10:  [73 CI screenshot-diff]
                                
Месяц 10:    [74 origin deploy]
```

**Минимальный путь к запуску** (если нет U-001 биржи и U-002
турниров): 6-7 месяцев.
**Полный паритет с origin-фичами**: 9-10 месяцев.

---

## Часть VIII. Матрица «D-NNN → план»

| D-NNN | Категория | План |
|---|---|---|
| D-001 (multi-points) | домен | 69 |
| D-002 (vacation семантика) | домен | 68 (нужна аккуратная миграция) |
| D-003 (account deletion) | домен | 69 (origin ветка) |
| D-004 (protection_time) | домен/механика | 69 |
| D-005 (observer) | домен | 69 |
| D-006 (umi координаты) | домен | 73 (миграционный скрипт) |
| D-007 (UI customization) | домен | 68 + 71 (frontend) |
| D-008 (profession + prof_time) | домен | 69 |
| D-009 (activation tokens) | инфра | identity-svc (готов) |
| D-010 (last activity) | домен | документация в 68 |
| D-011 (battle reports) | домен | 73 (миграционный скрипт) |
| D-012 (espionage reports) | домен | 73 (миграционный скрипт) |
| D-013 (event state) | event-loop | 73 (миграционный скрипт) |
| D-014 (alliance ranks) | домен/механика | 67 |
| D-015 (officer units) | домен | **отказ** (Часть V) |
| D-016 (planet teleport rate-limit) | домен | 64 (вместе с teleport) |
| D-017 (achievements) | домен/механика | 70 |
| D-018 (asteroid slots) | домен/механика | (опционально, после 73) |
| D-019 (home planet hp) | домен | 69 |
| D-020 (chat read tracking) | домен | 69 |
| D-021 (race) | домен/механика | 68 (поле) + опц. план о бонусах |
| D-022 (prod_factor per planet) | домен/формула | 64 |
| D-023 (event payload serialization) | event-loop | 73 (миграционный скрипт) |
| D-024 (event chains) | event-loop | 64 (вместе с alien chains) |
| D-025 (user agreement) | инфра | identity-svc (готов) |
| D-026 (формула DSL источник) | формула | 64 |
| D-027 (RF алиенов) | формула | 63 + 65 |
| D-028 (спец-юниты) | формула/домен | 64 |
| D-029 (температура водорода) | формула | 64 |
| D-030 (charge экспоненты) | формула | 64 |
| D-031 (TOURNAMENT) | event-loop/механика | **отказ** в первой итерации (Часть V) |
| D-032 (TELEPORT_PLANET) | event-loop/механика | 65 |
| D-033 (TEMP_PLANET) | event-loop | 64 верификация (вероятно ✅) |
| D-034 (RUN_SIM_ASSAULT) | event-loop | **отказ** (Часть V) |
| D-035 (DELIVERY_ARTEFACTS) | event-loop/механика | 65 |
| D-036 (alien chains) | event-loop/механика | 66 |
| D-037 (ATTACK_DESTROY_BUILDING) | event-loop/механика | 65 |
| D-038 (ALIEN_ATTACK_CUSTOM) | event-loop | 53 (admin-bff) |
| D-039 (биржа артефактов) | api/механика | 68 |
| D-040 (передача лидерства) | api | 67 |
| D-041 (3 описания альянса) | api/домен | 67 |
| D-042 (global mail альянса) | api/механика | 66 (после 57) |
| D-043 (Phalanx) | api | 64 верификация |
| D-044 (ResTransferStats) | api | 67 |
| D-045 (ExchangeOpts) | api/механика | (опционально, после 73) |
| D-046 (artefact-image PHP-GD) | assets | 72 |

**Все 46 D-NNN маппированы** на конкретный план или явный отказ.

---

## References

- [comparison.md](comparison.md) — сводные таблицы по 8 категориям
- [divergence-log.md](divergence-log.md) — все 46 D-NNN
- [nova-ui-backlog.md](nova-ui-backlog.md) — U-NNN + X-NNN
- [origin-ui-replication.md](origin-ui-replication.md) — S-NNN
- [alien-ai-comparison.md](alien-ai-comparison.md) — A1-A14
- [formula-dsl.md](formula-dsl.md) — origin DSL
- [origin-inventory.md](origin-inventory.md), [nova-inventory.md](nova-inventory.md)
