# Упрощения и отложенные доработки

Живой список всех осознанных упрощений, сделанных в ходе итераций.
Каждое упрощение попадает сюда в момент, когда принимается — чтобы
позже не потерять контекст.

Формат:
- **Где** — файл/модуль/миграция.
- **Что упрощено** — конкретика, что НЕ делается.
- **Почему** — trade-off решения на момент принятия.
- **Как чинить** — краткий план на возврат.
- **Приоритет** — L / M / H.

Сортировка внутри доменов — примерно по фактическому порядку
появления. Закрытые (сделанные) упрощения переносим в секцию
"Закрытые" с датой-итерацией.

---

## Battle engine (M4)

### [M4.1] Щиты — Java-алгоритм портирован — ЗАКРЫТО
- Портирован `Units.processAttack` shield-блок (строки 315–427 oxsar2-java):
  `ignoreAttack = shield/100`, `shieldDestroyFactor = clamp(1-turnShield/fullTurnShield, 0.01, 1.0)`,
  `startTurnShield` хранится в `unitState`. Тесты обновлены под Java-поведение (commit 665cdd7).

### [M4.1] Регенерация щитов — 100% каждый раунд
- **Где**: `battle/engine.go::regen`.
- **Что**: `turnShield = quantity × Shield` в начале каждого раунда,
  не учитываем partial regen (shieldDamageFactor в Java).
- **Почему**: упрощение для стабильной базы под ablation-тесты.
- **Как чинить**: добавить поле `turnShieldMax` и считать `regen = delta
  (max - current)` с затуханием после массового пробития.
- **Приоритет**: L.

### [M4.1] Multi-channel attack — один канал на unit-stack
- **Где**: `battle/engine.go::primaryChannel`.
- **Что**: весь stack стреляет каналом с max `Attack[c]`, щит цели
  оценивается только по этому каналу.
- **Почему**: в большинстве ship'ов всё равно один канал ненулевой.
- **Как чинить**: расщепление выстрелов по каналам c учётом всех
  Shield[3] компонентов (Java Units.processAttack делает это
  per-shot).
- **Приоритет**: L (пока нет юнитов с multi-channel атакой).

### [M4.2] Ballistics/masking — детерминированная формула без RNG
- **Где**: `battle/engine.go::applyMasking`.
- **Что**: `missed = floor(shots × factor)` — ровно та же Java-формула,
  но без RNG-roll per shot.
- **Почему**: в Java тоже deterministic (factor применяется к пулу),
  это не упрощение, а соответствие.
- **Как чинить**: не нужно — формула совпадает с legacy.
- **Приоритет**: —

### [M4.3] Ablation — максимум 1 damaged-юнит на stack
- **Где**: `battle/engine.go::commitDamage`.
- **Что**: остаток shell создаёт **один** damaged с `shell_percent`;
  массив разных `shell_percent` на нескольких юнитах одного stack не
  поддерживается.
- **Почему**: Java делает так же — упрощение bookkeeping.
- **Как чинить**: не нужно.
- **Приоритет**: —

### [M4.3] Пропущены golden-тесты против Java-jar
- **Где**: `testdata/battle/*.json` (пустое).
- **Что**: нет cross-verification c `oxsar2-java/assault/dist/
  oxsar2-java.jar`.
- **Почему**: требует JVM в окружении, JavaRandom-адаптер, harness.
- **Как чинить**: добавить `pkg/rng/javarand.go`, прогнать 5–10
  сценариев в обоих движках, diff JSON.
- **Приоритет**: M — при балансировании боя.

### [M4.4a] rapidfireToMap возвращает nil (исправлено)
- **Где**: `fleet/attack.go`.
- **Статус**: Закрыто в итерации 20 (commit c7ae59a). rapidfireToMap
  теперь проксирует cat.Rapidfire.Rapidfire напрямую.

### [M4.4a] Debris=часть loot (исправлено)
- **Статус**: Закрыто в итерации 20. Debris отделено от loot,
  попадает в `debris_fields` на координаты, собирается RECYCLING.

### [M4.4a] moon-chance — ЗАКРЫТО
- **Статус**: Закрыто в итерации 42. tryCreateMoon реализован в
  finalizeAttack: chance=min(20,debris/100000)%, луна создаётся при
  успешном броске если is_moon=false и луны ещё нет.
  Moon Destruction (kind=14) и Stargate (kind=28) — M5+.

### [M4.4a] Нет RES_LOG entries от боя (loot) — ЗАКРЫТО
- Закрыто: `finalizeAttack` (attack.go) и ACS-handler (acs_attack.go) вставляют
  два `res_log` entry при ненулевом лоуте: attacker +loot, defender -loot.
  reason='loot' — уже был в schema comment.

### [M4.4a] Нет unit-тестов AttackHandler — ЧАСТИЧНО ЗАКРЫТО
- Закрыто: `fleet/attack_test.go` — 8 unit-тестов для pure-функций:
  `grabLoot` (4 теста: unlimited cargo, constraint, no survivors, carry reduces free),
  `calcDebris` (2 теста: basic 30%, defense excluded),
  `deriveSeed` (2 теста: deterministic, different inputs → different seeds).
- Остаток: e2e-тест полного цикла AttackHandler + transaction (требует testcontainers).
- **Приоритет**: L — hot-path покрыт, transaction-level тест отложен.

---

## Fleet missions

### [TRANSPORT] recall только для state='outbound'
- **Где**: `fleet/transport.go::Recall`.
- **Что**: после прибытия (arrive) recall невозможен — флот уже
  выполняет миссию или возвращается.
- **Почему**: семантически правильно (нет смысла «отзывать» уже
  разгрузившийся транспорт).
- **Как чинить**: не нужно.
- **Приоритет**: —

### [RECYCLING] Debris один на координаты, без is_moon — ЗАКРЫТО
- Закрыто: migration 0030 добавляет `is_moon bool` в PK debris_fields.
  Все INSERT/UPDATE/SELECT в attack.go, acs_attack.go, events.go обновлены.

### [SPY] Counter-espionage + research>=8 — ЗАКРЫТО
- Закрыто: `buildEspionageReport` добавляет Research при ratio>=8 через
  `readOwnerResearch`. Counter-espionage: `min(defTotal/10, probes)` roll,
  при перехвате всех probes флот уничтожается (commit 306bbaa).

### [COLONIZE] Имя и размер планеты по позиции — ЗАКРЫТО
- Закрыто: `colony_name` в TransportInput/sendRequest/transportPayload.
  UI: поле ввода при mission=8. `positionDiameter(pos, r)`: pos 1-3 → 6000-10000,
  pos 4-12 → 10000-15000, pos 13-15 → 12000-17000. Дефолт имени «Colony».

### [EXPEDITION] Детерминирована по seed от fleetID
- **Где**: `fleet/expedition.go`.
- **Что**: одинаковый fleet_id всегда даёт одинаковый outcome. Это
  не минус (каждый новый flight имеет новый uuid), но тесты с
  фиксированным uuid дадут одинаковые результаты.
- **Почему**: проще тестировать.
- **Как чинить**: не нужно.
- **Приоритет**: —

### [EXPEDITION] black_hole (4) и unknown (13) не реализованы
- **Где**: `fleet/expedition.go`.
- **Что**: исходы black_hole и unknown никогда не выбираются.
- **Почему**: black_hole не реализован даже в legacy; unknown — зарезервирован.
- **Как чинить**: black_hole потребует механики «флот безвозвратно пропадает» — отдельная задача.
- **Приоритет**: L.

### [EXPEDITION] Battlefield — упрощённая генерация врага
- **Где**: `fleet/expedition.go::expBattlefield`.
- **Что**: вместо случайного подмножества кораблей × random(0.3–0.8) от exp_power
  генерируем fleet из LF с shell_percent=0.5 по expPower.
- **Почему**: упрощение bookkeeping; суть (бой с повреждённым флотом) сохранена.
- **Как чинить**: добавить cruiser/bs в состав врага пропорционально expPower.
- **Приоритет**: L.

### [EXPEDITION] credit_purchases таблица отсутствует
- **Где**: `fleet/expedition.go::expCredit`.
- **Что**: `buy_credit` из покупок за 3 дня — запрос к `credit_purchases` таблице.
  Таблица не существует, buyCredit всегда = 0. Формула всё равно работает без неё.
- **Почему**: таблица покупок — часть платёжной системы, которая ещё не реализована.
- **Как чинить**: создать `credit_purchases(user_id, amount, created_at)`, подключить к payment flow.
- **Приоритет**: M — нужно при монетизации.

### [EXPEDITION] expExtraPlanet — временные планеты не чистятся воркером
- **Где**: `fleet/expedition.go::expExtraPlanet`, `migrations/0044_planets_expires_at.sql`.
- **Что**: поле `planets.expires_at` добавлено, значение выставляется, но воркер ещё не удаляет
  планеты с истёкшим `expires_at`.
- **Почему**: воркер-задача требует отдельного event-kind или cron-handler.
- **Как чинить**: добавить `KindExpirePlanet` event или cron-job `DELETE FROM planets WHERE expires_at < now()`.
- **Приоритет**: M.

---

## Market

### [Market] Фиксированные курсы 1:2:4 — ЗАКРЫТО (не упрощение)
- Курсы metal=1, silicon=2, hydrogen=4 — соответствие legacy OGame (не упрощение).
  Ордерная книга (CreateLot/ListLots/CancelLot/AcceptLot, migration 0022) реализована.

### [Market] Только в рамках одной планеты
- **Где**: `internal/market/service.go::Exchange`.
- **Что**: обмен происходит на конкретной планете, нет межпланетного
  swap (ресурсы надо везти).
- **Почему**: так проще, OGame тоже так делает.
- **Как чинить**: не нужно.
- **Приоритет**: —

### [ArtefactMarket] Фильтр «мои» — ЗАКРЫТО
- **Закрыто**: добавлен `GET /api/me` → `{user_id, username}`,
  фильтр сравнивает `seller_user_id === me.user_id`. Кнопка действия
  теперь тоже разделена: Cancel — для своих, Buy — для чужих.

### [ArtefactMarket] Цена через window.prompt — ЗАКРЫТО
- Закрыто: inline price input в строке таблицы (number input + OK/Cancel).
  Enter подтверждает, Escape отменяет. window.prompt удалён.

---

## Rockets

### [Rockets] Anti-ballistic missile (перехват) — ЗАКРЫТО
- Реализовано: `interceptorRocketUnitID=51`, вычитается из rocket_count
  до расчёта урона. ABM-юниты расходуются. Уведомление включает ABM-статистику.

### [Rockets] Урон размазан по всей defense без приоритета — ЗАКРЫТО
- Закрыто: `launchRequest.TargetUnitID int` (0 = без приоритета). `Launch` передаёт
  в payload как `target_unit_id`. `ImpactHandler`: если != 0, весь урон идёт сначала
  в этот стек, overflow → остальным пропорционально.

### [Rockets] Нет silo-limit — ЗАКРЫТО
- Закрыто: `missile_silo` (id=13) добавлен в `configs/buildings.yml`
  (`rocket_capacity_per_level: 10`). `BuildingSpec` получил поле
  `RocketCapacityPerLevel`. `Launch` проверяет `count <= siloLevel × cap`;
  при нарушении — `ErrSiloLimit` → HTTP 400. Если шахта не построена
  (`siloLevel=0`) — лимит не применяется (обратная совместимость для
  стартовых планет без шахты).

---

## Repair

### [Repair] Нельзя чинить defense — ЗАКРЫТО
- Закрыто: migration 0032 добавляет `damaged_count` и `shell_percent` в таблицу defense.
  `applyDefenderLosses` в attack.go теперь записывает damaged/shell для defense симметрично ships.
  `EnqueueRepair` поддерживает оба типа (ships и defense, через `stockTable`).
  `ListDamaged` объединяет ships + defense через UNION ALL. `DamagedUnit` получил поле `is_defense`.

### [Repair] Batch-only (чиним всех damaged одним action)
- **Где**: `internal/repair/service.go::EnqueueRepair`.
- **Что**: кнопка «Починить» чинит N=damaged_count. Нет «починить k
  из N».
- **Почему**: shell_percent на stack один, «частичный ремонт» требует
  усложнения модели.
- **Как чинить**: не нужно пока modelчасть stack'а может иметь разный
  shell_percent.
- **Приоритет**: —

### [Repair] Стоимость считается в момент enqueue — ЗАКРЫТО
- Закрыто: ресурсы планеты читаются внутри транзакции (`FOR UPDATE`) в шаге 4
  вместо pre-tx снапшота из `s.planets.Get`. TOCTOU при параллельных enqueue
  теперь невозможен.

---

## Messages

### [Messages] Read-only inbox (нет compose/reply/delete) — ЗАКРЫТО
- **Статус**: Закрыто в итерации 38. Реализованы POST /api/messages
  (compose), DELETE /api/messages/{id}, UI composer с ComposeForm.

### [Messages] Reply — ЗАКРЫТО
- Закрыто: кнопка «↩ Ответить» в MessageDetail (только для сообщений
  с from_user_id). Pre-fill ComposeForm: to=from_username, subject=«Re: …».

### [Messages] Нет soft-delete — ЗАКРЫТО
- Закрыто: migration 0027 добавляет `deleted_at TIMESTAMPTZ` + частичный индекс
  `WHERE deleted_at IS NULL`. Delete→`UPDATE SET deleted_at=now()`. Inbox и UnreadCount
  фильтруют `deleted_at IS NULL`.

### [Messages] Username в BattleReport — ЗАКРЫТО
- LEFT JOIN users ua/ud добавлен в GetBattleReport. Поля
  attacker_username/defender_username теперь в ответе (commit 10555a6).

### [Messages] Folders — ЗАКРЫТО
- Закрыто: tab-фильтры «Все / Личные / Бой / Шпионаж / Экспедиции / Система»
  в MessagesScreen. Фильтрация на клиенте по `m.folder`.

---

## Achievements

### [Achievements] Lazy-check при GET, не real-time — ЗАКРЫТО
- Закрыто: `withAchievement` decorator в worker/main.go вызывает
  `achSvc.CheckAll` после KindBuildConstruction, KindArtefactExpire,
  KindAttackSingle, KindAttackAlliance, KindColonize. Ошибка не
  прерывает основной handler — только логируется (commit далее).

### [Achievements] Только 5 штук — ЗАКРЫТО
- Закрыто: migration 0026 добавляет 10 новых достижений (FIRST_FLEET,
  FIRST_EXPEDITION, FIRST_RESEARCH, BATTLE_10, FLEET_50, ARTEFACT_MARKET,
  SPY_SUCCESS, RECYCLING, ROCKET_LAUNCH, SCORE_1000). CheckAll расширен
  соответствующими SQL-checks.

### [Achievements] Нет прогресс-баров (N/M) — ЗАКРЫТО
- Закрыто: `progressChecks []progressCheck` в service.go считает on-the-fly
  для BATTLE_10/FLEET_50/SCORE_1000. Entry получила `Progress *int` + `ProgressMax *int`.
  UI показывает «N / max» рядом с описанием для незакрытых числовых достижений.

---

## Officers

### [Officers] Стеккаются с артефактами без suppression — ЧАСТИЧНО ЗАКРЫТО
- Закрыто для officer-vs-officer: migration 0033 добавляет `group_key` в officer_defs.
  ADMIRAL и ENGINEER в группе 'build' — взаимоисключают друг друга при активации.
  `ErrGroupActive` → HTTP 400. Activate читает `group_key` и проверяет активных в группе.
- Остаток: officer+artefact suppression не реализовано — арtefakt short-lived,
  суммирование +0.2 не критично. Потребует `group_key` в artefact_defs.
- **Приоритет**: L → отложено до появления проблем в геймплее.

### [Officers] Нет auto-renew — ЗАКРЫТО
- Закрыто: migration 0029 добавляет `auto_renew bool` в officer_active.
  `Activate(ctx, uid, key, autoRenew bool)` сохраняет флаг + передаёт в payload события.
  `ExpireHandler`: если auto_renew=true и credit >= cost → re-INSERT active, новый event,
  factor остаётся без изменений. Если credit не хватает — обычный expire с пояснением.
  UI: чекбокс «авто» рядом с кнопкой активации.

### [Officers] ADMIRAL — описание исправлено — ЗАКРЫТО
- Описание изменено на «Ускоряет постройку кораблей в верфи на 10%».
  Миграция 0023 исправляет БД, 0015 исправлена для новых установок.

### [Officers] Credit не восстанавливается при expire
- **Где**: `officer/service.go::ExpireHandler`.
- **Что**: при истечении credit НЕ возвращается — это подписка-
  расходник.
- **Почему**: соответствие legacy (и экономики PvE — бонус за
  деньги).
- **Как чинить**: не нужно.
- **Приоритет**: —

---

## AutoMsg

### [AutoMsg] Примитивный шаблонизатор через strings.ReplaceAll
- **Где**: `internal/automsg/service.go::Send`.
- **Что**: подстановка `{{var}}` через `strings.ReplaceAll` для каждой
  пары vars. Нет условных блоков, циклов, форматирования чисел, i18n
  branching. Legacy AutoMsg.class.php — 1228 LOC с полным шаблонизатором.
- **Почему**: все наши текущие шаблоны — простые «привет, {{user}}».
- **Как чинить**: заменить на `text/template` (плюс: safe-escape,
  условия). Но только если шаблоны станут сложнее.
- **Приоритет**: L.

### [AutoMsg] Только event-driven, нет scheduled messages — ЧАСТИЧНО ЗАКРЫТО
- **Статус**: Закрыто в итерации 43 для inactivity-reminder.
  users.last_seen_at обновляется через LastSeenMiddleware (async).
  Воркер ежедневно шлёт INACTIVITY_REMINDER тем, кто не заходил 3+ дней.
  Нет: weekly-digest, event-before-raid (за N минут до прибытия флота).
- **Как чинить** (остаток): ~~event-before-raid~~ — ЗАКРЫТО: `KindRaidWarning=64`
  планируется при Send для mission=10/12 с `fire_at = arrive_at - 10min` (если warnAt > depart).
  RaidWarningHandler читает флот и шлёт сообщение (folder=13/Система) защитнику.
- **Приоритет**: L — inactivity уже есть.

### [AutoMsg] Нет CMS / редактирования через UI — ЗАКРЫТО
- Закрыто: `GET /api/admin/automsgs` + `PUT /api/admin/automsgs/{key}` в admin.Handler.
  AdminScreen получил раздел «Шаблоны сообщений» с inline-редактором (title, folder, body_template).
  Создание новых шаблонов — только через миграцию (ключ PK, произвольный INSERT опасен).

### [AutoMsg] Welcome отправляется вне транзакции регистрации
- **Где**: `auth/service.go::Register`.
- **Что**: `s.automsg.Send(ctx, nil, ...)` с `tx=nil` — после
  фиксации пользователя и планеты. Если Send упадёт, регистрация
  всё равно удалась.
- **Почему**: WELCOME не критичен. Если упадёт — пользователь
  просто не увидит приветствие, но играть сможет.
- **Как чинить**: передать общую `tx` (нужна refactor-итерация
  `Register` — выделить всю операцию в одну InTx).
- **Приоритет**: L.

---

## Economy / Catalog

### [planet.tick] construction.yml — ЗАКРЫТО
- Закрыто: запущен `go run ./cmd/tools/import-datasheets` из na_construction.sql.
  `configs/construction.yml` сгенерирован и закоммичен. DSL-путь (legacy-формулы)
  теперь активен в проде без fallback на приближения.

### [economy] Storage cap — ЗАКРЫТО (было ошибочно открыто)
- `storageCap()` в planet/service.go корректно читает уровни metal_storage/
  silicon_storage/hydrogen_storage и применяет `StorageCapacity()`.
  `clampAdd()` ограничивает прирост. Уже работает.

---

## Infrastructure

### [docker] Frontend dev-mode через bind-mount — ЗАКРЫТО
- Закрыто: `deploy/Dockerfile.frontend-prod` (multi-stage: node builder +
  nginx:1.27-alpine). `deploy/nginx.frontend.conf` — SPA fallback, /api
  proxy на backend:8080, gzip, long-cache для ассетов. `deploy/docker-compose.prod.yml` —
  prod overlay поверх dev-compose (overrides frontend service).

### [docker] Auth rate-limiting — ЗАКРЫТО
- Реализован `auth.RateLimiter` (Redis sliding-window): 20 req/min per IP.
  Применён к `/api/auth/login`, `/api/auth/register`, `/api/auth/refresh`.
  Fail-open при недоступном Redis (commit 1a27dc4).

### [i18n] Только ru/en, en.yml stub
- **Где**: `configs/i18n/en.yml`.
- **Что**: en.yml — заготовка с пустыми значениями. Реально тексты
  только ru.
- **Почему**: сначала legacy-порт, переводы — потом.
- **Как чинить**: ручной перевод или translation workflow.
- **Приоритет**: M — для международного запуска.

---

## Score / Ranking (M5+)

### [score.batch] RecalcUser real-time — ЗАКРЫТО
- Закрыто: `withScore` decorator в worker/main.go вызывает `RecalcUser` после
  KindBuildConstruction/KindResearch/KindBuildFleet/KindBuildDefense.
  RecalcAll (5 мин) остаётся как fallback для прочих событий.

---

## Alliance (M6)

### [Alliance] MVP без рангов и WebSocket-чата
- **Где**: `internal/alliance/service.go`, `features/alliance/AllianceScreen.tsx`.
- **Что**: ~~отношения NAP/WAR/ALLY~~ — ЗАКРЫТО. ~~Встречный acknowledge~~ — ЗАКРЫТО.
  ~~Кастомные ранги~~ — ЗАКРЫТО (migration 0034: `rank_name` в alliance_members;
  `SetMemberRank` + `PATCH /alliances/{id}/members/{uid}/rank`; MembersTable с inline-редактором).
  Остаток: ACS-атаки (требует координации нескольких флотов в реальном времени).
- **Приоритет**: L → ACS — крупная фича M6+.

### [Alliance] Join approval flow — ЗАКРЫТО
- Закрыто: migration 0024 добавляет `is_open` в alliances + таблицу
  `alliance_applications`. Join с is_open=false создаёт заявку;
  owner видит список и вызывает Approve/Reject. UI обновлён (commit 1013f05).

### [Alien AI] GRAB_CREDIT/GIFT_CREDIT — ЗАКРЫТО
- Закрыто: при победе инопланетян берётся 0.08–0.1% кредитов (если >100000);
  при отражении — дарится 5–10% (max 500). Логика в `applyGrabCredit` /
  `applyGiftCredit` внутри AttackHandler. Сообщение дополнено суммой.

### [Alien AI] Без HALT/multi-step state machine (остаток)
- **Где**: `internal/alien/alien.go`.
- **Что**: нет multi-step cycle: ALIEN_FLY=33 → HOLDING=34 → ATTACK=35 → HALT=36 → RETURN.
  Атака материализуется через один event fire_at без промежуточных состояний.
- **Почему**: alien_fleets state machine = сложный конечный автомат (~300 LOC PHP).
- **Как чинить**: добавить `alien_fleets` таблицу с координатами и конечным автоматом.
- **Приоритет**: L — текущая механика атаки уже работает и передаёт суть.

### [Tutorial] Тексты шагов хардкод на русском — ЗАКРЫТО
- Закрыто: ключи `TUTORIAL_STEP_N_TITLE` / `TUTORIAL_STEP_N_DESC` добавлены
  в `configs/i18n/ru.yml`; `TutorialScreen.tsx` использует `tf('Main', key, fallback)`.

### [Tutorial] Нет наград кроме кредитов — ЗАКРЫТО
- Закрыто: `stepResources[6][3]` — metal/silicon/hydrogen за каждый шаг.
  `advanceAndReward` зачисляет ресурсы на первую планету игрока (ORDER BY created_at).
  Шаг 6 даёт 5000M/3000Si/1000H. Кредиты (+10) сохранены.

### [ACS] Loot пропорционально грузоподъёмности — ЗАКРЫТО
- Закрыто: `cargoPerFleet[]` считается через range-итерацию по `s.catalog.Ships.Ships`
  (map keyed by legacy name), матчинг по `spec.ID == st.UnitID`. Loot делится
  пропорционально `cargoPerFleet[i] / totalCargo`. Fallback: если totalCargo==0,
  делим поровну (чтобы избежать деления на ноль).

### [ACS] acs_participants в battle_reports — ЗАКРЫТО
- Закрыто: migration 0025 добавляет `acs_participants jsonb` в battle_reports.
  ACS handler записывает [{user_id, fleet_id}] всех атакующих флотов (commit 0025).

---

## UI Porting (H-план, 2026-04-23)

### [09-Ф5.2] Score.RecalcAll: *_specs таблицы и single-SQL batch не реализованы
- **Где**: `backend/internal/score/`, отсутствует миграция `*_specs`.
- **Что**: RecalcAllEvent теперь вызывается раз в сутки через
  KindScoreRecalcAll event (вместо 5-минутного ticker'а), но сам
  пересчёт всё равно идёт циклом через RecalcUser. Формулу внутри
  calcBuildings/calcResearch переписал на closed-form (O(1) вместо
  O(level)), что даёт существенное ускорение. Но single-SQL batch
  по плану (CTE с JOIN *_specs) не реализован.
- **Почему**: full batch требует новых таблиц building_specs /
  research_specs / ship_specs / def_specs с cost_base_sum + cost_factor
  из YAML + bootstrap на старте. Ощутимый выигрыш только при >10k
  активных игроков (сейчас ~десятки). O(1) closed-form + раз-в-сутки
  достаточно для текущей стадии.
- **Как чинить**: миграция 0059 с `*_specs` таблицами; cmd/tools/sync-specs
  заполняет их из каталога при деплое; RecalcAllBatch выполняет
  single CTE из плана 09 секция «SQL для batch-сверки».
- **Приоритет**: L (пока DAU < 10k).

### [09-Ф2.1] events.trace_id заполняется только в новых INSERT'ах
- **Где**: 11 мест в backend/internal/*/service.go, event-producers.
- **Что**: колонка events.trace_id и индекс добавлены (миграция 0058),
  Event.TraceID читается в worker и прокидывается в context handler'а.
  Но 11 существующих `INSERT INTO events (...)` не передают trace_id —
  поле остаётся NULL для этих событий.
- **Почему**: замена 11 SQL-строк + прокидывание trace.FromContext(ctx)
  в каждый сервис — это отдельная итерация. Пока trace_id работает
  end-to-end только для новых INSERT'ов, которые добавим после.
- **Как чинить**: `event.Insert(ctx, tx, kind, userID, planetID, fireAt, payload)` —
  один helper + замена 11 raw-INSERT'ов. После этого trace_id
  гарантированно цепляется от HTTP-запроса до worker-handler'а.
- **Приоритет**: M.

### [H.1.7] Messages: фаланга-producer отсутствует
- **Где**: `backend/internal/fleet` — сканы фаланги не реализованы.
- **Что**: папка «Фаланга» (folder=11) видима в UI, но не наполняется —
  сам сканер фаланги ещё не реализован в nova-backend.
- **Почему**: это отдельная большая фича, не входит в план 11.
- **Как чинить**: реализовать `KindPhalanxScan` event + UI-триггер в
  galaxy + automsg.SendDirect(folder=11) в обработчике.
- **Приоритет**: M.

### [H.2.11] Payment: custom-сумма пополнения
- **Где**: `frontend/src/features/payment/CreditsScreen.tsx`.
- **Что**: только фиксированные пакеты из `cfg.Payment.Packages`, нет
  поля «своя сумма». Подсказки по пользе пакетов добавлены.
- **Почему**: backend требует `package_key` для валидации через
  RobokassaGateway; произвольная сумма потребует нового flow с
  dynamic-package (с генерацией signature на лету).
- **Как чинить**: расширить payment.Service.CreateOrder дополнительной
  веткой с CustomAmountRub, отдельный PaymentFixedGateway для custom.
- **Приоритет**: L.

---

## UI Testing (план 13, 2026-04-24)

- **Unit-тесты React-компонентов отсутствуют** — вместо них делаем E2E
  на Playwright. Причина: время/риск-профиль. Vitest-каркас уже подключён
  (см. package.json), нужно ~500 LOC каркаса + mocks для 35 экранов.
  Решили: смоук-E2E закрывает «UI не падает» за малые деньги, unit-тесты
  точечно добавляем вместе с bug-fix'ами.
  **План возврата**: по мере наработки компонентной библиотеки, добавить
  unit-покрытие хотя бы для критичных UI-утилит (формы, валидаторы,
  преобразователи ресурсов).
  **Приоритет**: M.

- **testseed идемпотентен только с флагом --reset** — повторный запуск без
  --reset допускает рост таблиц (например, дублирующиеся messages).
  Причина: полный UPSERT со всеми связями требует списка фиксированных
  UUID для ВСЕХ вставляемых строк, включая сообщения/лоты/отчёты. Это
  раздувает код сидера в 3×.
  **План возврата**: если E2E окажутся медленными, сделать fully-idempotent
  (фиксированные UUID для всех строк).
  **Приоритет**: L.

- **Playwright webServer не включён** — конфиг требует ручного запуска
  backend+frontend. Причина: holodny start goose-миграций + воркера
  + vite = 15-30 сек, webServer таймаут будет бомбить в CI. Проще иметь
  отдельный шаг CI «make dev-up && make test-seed» перед Playwright.
  **План возврата**: при стабилизации CI — включить webServer и убрать
  ручной шаг.
  **Приоритет**: L.

- **api-coverage regex matcher грубый** — заменяет `{id}` → `[^\\s'"\`]+?`
  и грепает по литералам. Не отличит `/api/planets/{id}` от
  `/api/planets/foo/bar` в случайной строке. Для текущего репо
  ложноположительных нет, но гарантий нет.
  **План возврата**: если появятся ложные срабатывания — парсить AST
  через ts-morph и искать только вызовы api.get/post.
  **Приоритет**: L.

---

## Закрытые

- **M4.4a.rapidfire** → исправлено в iteration 20 (commit c7ae59a).
- **M4.4a.debris-in-loot** → исправлено в iteration 20 (commit 618cd26).
- **M4.4a.ui-missions** → исправлено в iteration 20.5 (commit 5336f06).
- **M4.4c** REPAIR-режим → iteration 19 (commit 42e4c89).
- **Starter-planet без buildings** → iteration 12 (commit 15be227).
- **planet evalProd не обрабатывал resource='energy'** → iteration 13.

### [UI] Кредиты в шапке — ЗАКРЫТО
- Поле `credit numeric(15,2)` существует в `users` с migration 0001, стартовое значение 5.00.
- Migration 0013 Down (откат) содержит DROP COLUMN — это было ошибочно записано как «поле удалено».
- Реализовано (UI-15): `/api/me` возвращает `credit`, шапка показывает 💳 N cr.

### [UI] Fleet: retreat/formation/прогресс-бар боя не реализованы
- **Ситуация**: legacy показывал кнопки retreat/formation и прогресс-бар при активном бое.
- **Причина**: бой в нашей системе происходит мгновенно при arrive_at (worker event), нет "battle in progress" state.
- **Решение**: не показываем. Все боевые исходы в battle_reports.
- **Возврат**: при реализации ACS с задержкой боя.

### [UI] Score: легенда статусов (i/I/b/v) и отношения с игроком
- **Ситуация**: legacy показывал статусы (неактив/бан/отпуск) и отношения (союзник/враг) в рейтинге.
- **Причина**: сложность реализации, низкий приоритет.
- **Решение**: показываем только альянс [TAG]. Статусы umode=false уже фильтруем (банные не видны).
- **Возврат**: M9 или по запросу.

### [UI] ResourceScreen: FactorInput — ЗАКРЫТО (2026-04-23)
- Заменено на `<input type="range">` + пресеты 0/25/50/75/100, автосохранение по `onMouseUp`/`onTouchEnd`, убраны кнопки «Сохранить» и «Назад».

### План 11 UI Porting Follow-up — 11 упрощений ЗАКРЫТО (2026-04-23)
Все упрощения из H-плана закрыты параллельными итерациями 11.1–11.12:

- **H.2.10 Search навигация с контекстом** (шаг 1, commit 4e0b035).
  GalaxyScreen.initialCoords + ScoreScreen.initialQuery + фильтр по нику.
- **H.1.6 Score координаты гл. планеты** (шаг 8, commit 4e0b035).
  LATERAL JOIN + кликабельные координаты → galaxy.
- **H.1.7 Messages producers** (шаг 2, commit 6641ba8). Добавлен
  `automsg.SendDirect`; интегрирован в payment, referral, alliance,
  artefact. Остался только phalanx-producer (сканер не реализован).
- **H.1.5 Galaxy alliance relations** (шаг 3, commit 9547437). CellView.relation,
  ReadSystem с viewerUserID, цвета rows (ally=зелёный/war=красный/nap=жёлтый).
- **H.3 Delete account через код** (шаг 6, commit недавний). migration 0051
  account_deletion_codes, argon2id, TTL 10 мин, 3/час rate-limit, soft-delete.
- **H.3 Planet sort drag&drop** (шаг 7). migration 0052 planets.sort_order,
  PATCH /api/planets/order, HTML5 draggable без библиотек.
- **H.2.9 Friends** (шаг 5). migration 0053 friends, CRUD endpoints,
  FriendsScreen с онлайн-статусом, ⭐ подсветка в галактике.
- **H.2.6 Records** (шаг 10). `GET /api/records` — топ-1 по зданиям,
  исследованиям, флоту, обороне, очкам + мой % от рекорда.
- **H.2.7 ResTransferStats** (шаг 11). migration 0054 resource_transfers,
  логирование в fleet.ArriveHandler, `GET /api/stats/resource-transfers`,
  вкладка «Торговля» в Score.
- **H.1.9 Fleet market** (шаг 4). migration 0055 market_lots.kind/sell_fleet,
  CreateFleetLot/AcceptFleetLot/CancelFleetLot, UI под-вкладки в LotsPanel.
- **H.2.2 Techtree SVG** (шаг 9). Чистый SVG с layered layout, тумблер
  «🗂 Карточки / 🌐 Граф».
- **H.2.11 Payment подсказки** (шаг 12). packageHint по размеру пакета +
  раздел «На что потратить» с ценами.

Все миграции 0051–0055, go build чистый. Осталась одна незакрытая
запись в UI Porting — phalanx-producer, ждёт реализации самого сканера.
