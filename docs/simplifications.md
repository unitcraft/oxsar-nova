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

### [M4.1] Щиты — Java-алгоритм портирован — ЗАКРЫТО (уточнён 2026-04-25)
- Портирован `Units.processAttack` shield-блок (строки 315–427 oxsar2-java):
  `shieldDestroyFactor = clamp(1-turnShield/fullTurnShield, 0.01, 1.0)`,
  `startTurnShield` хранится в `unitState`. Исправлена структура Java-алгоритма:
  выстрелы делятся на «к щиту» (shieldExistFactor) и «к shell» (остаток), а не
  наоборот. `ignoreAttack` вычисляется по базовому щиту без tech-масштабирования.
  Для планетарных щитов (id 49/50) `ignoreAttack=0` как в Java. BA-005 ЗАКРЫТ.

### [M4.1] Регенерация щитов — 100% каждый раунд
- **Где**: `battle/engine.go::regen`.
- **Что**: `turnShield = quantity × Shield` в начале каждого раунда,
  не учитываем partial regen (shieldDamageFactor в Java).
- **Почему**: упрощение для стабильной базы под ablation-тесты.
- **Как чинить**: добавить поле `turnShieldMax` и считать `regen = delta
  (max - current)` с затуханием после массового пробития.
- **Приоритет**: L.

### [M4.1] Multi-channel attack — упрощено до scalar (2026-04-25)
- **Где**: `battle/engine.go`, `battle/types.go`.
- **Что**: было `Attack [3]float64` / `Shield [3]float64` с выбором
  primaryChannel; стало `Attack float64` / `Shield float64`. Поле
  primaryChannel и функция выбора канала удалены.
- **Почему**: 3-канальная система (лазер/ион/плазма vs физ/магн/сил)
  была закомментирована в самом legacy oxsar2-java
  (`Participant.java:530-538`) и не использовалась нигде. Балансная
  матрица `ADV_TECH_MATRIX` существовала только как идея.
- **Как восстановить**: вернуть `[3]float64`, восстановить
  `primaryChannel`, добавить 3 технологии оружия (лазер/ион/плазма) и
  3 типа щита, реализовать матрицу 3×3 эффективности.
- **Приоритет**: L (кандидат на расширенный режим / отдельный сервер,
  не для MVP).

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
- **Где**: `projects/game-nova/backend/internal/battle/testdata/*.json` (пустое).
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

### [M4.x] attacker_* поля юнитов не используются в бою
- **Где**: `backend/internal/battle/engine.go::Front/Ballistics/Masking`,
  `configs/ships.yml`. В `frontend/src/api/catalog.ts` поля
  `attacker_front/ballistics/masking` объявлены в типе, заполнены у
  Deathstar (id=42), но в UI не отображаются и в боевой логике не
  читаются.
- **Что**: в легаси (`oxsar2-java Participant.java`, таблица
  `na_ship_datasheet`) параметры юнита раздельны для атакующей и
  обороняющейся стороны (`attacker_front/attack/shield/ballistics/
  masking` vs обычные). В nova используется единый набор значений
  независимо от роли.
- **Почему**: реальная разница в БД легаси есть только у 2 юнитов из
  ~30 — Deathstar (front 10→9 в атаке) и Alien Screen (front 15→16 в
  атаке). У всех остальных `attacker_* = defender_*`. Deathstar — late-
  game, встречается у 0–2 игроков в партии. Alien Screen — NPC
  пришельцев (AlienAI ещё не реализован). В массовых боях флотов
  эффект разницы `front ±1` тонет.
- **Как чинить**: добавить опциональные поля `attacker_front/attack/
  shield/ballistics/masking` в `ships.yml`, протащить через
  `battle.Unit` и в `engine.go` выбирать значение по роли стороны
  (attacker/defender). Оценка: < 1 дня. Актуально станет при
  реализации AlienAI с атаками по игрокам и при балансировке late-
  game с Deathstar.
- **Приоритет**: L — затрагивает < 1% боёв типичной партии.

### [M4.x] Rapidfire — урезанная заглушка — ЗАКРЫТО
- **Статус**: Закрыто 2026-04-24 (план 18 Фаза 1). До этого
  `configs/rapidfire.yml` содержал ~34 записи (сам файл подписан
  «TODO (M4): перенести один-в-один из oxsar2-java»). Из legacy
  `na_rapidfire` (d:\Sources\oxsar2\sql\new-for-dm\data.sql:9398)
  не было портировано 38 записей для игровых юнитов — в частности
  LF/SF/Cruiser/DS против Lancer (×20/20/35/100), все → Probe/SSat ×5,
  и 9 записей для Shadow Ship как атакующего. Это создавало мнимые
  эксплойты (BA-001, BA-002, BA-004).
- Исправление: портированы все 38 legacy-записей, удалены 2 «изобретения»
  (SD→StrongLaser ×2, DS→Plasma ×50), исправлены 2 числа (SD→LL ×2→×10,
  SD→Lancer ×2→×3). Добавлен Shadow Ship (id=325) в `units.yml/fleet:`.
  Alien rapidfire (id 200–204, 348, 352, 353) намеренно не портирован —
  AI-баланс планируется отдельно (план 24).

### [M4.be_points] be_points накопление-only, use-в-бою отложено
- **Где**: `internal/fleet/attack.go` (накопление), `internal/battle/`
  (отсутствует use).
- **Что**: legacy-механика `be_points` имеет 3 части — накопление при
  бое (`be_points += experience`), use при отправке атаки
  (`attack_lvl_<tech>` от -K до +K, где K = min(20, be_points/100)),
  возврат при cancel. Реализуем только **накопление** (план 72.1 ч.17).
  Поле наполняется при каждом бое и видно на MainScreen, но пока **не
  приносит in-game эффекта** — игрок копит впрок.
- **Почему**: full-implementation требует UI-формы для выбора уровней
  при отправке атаки, валидации, интеграции в `internal/battle/`
  (формула `unit.attack *= (1 + lvl/10)` для GUN-техов, аналогично
  для SHIELD/SHELL/BALLISTICS/MASKING), тестов battle-sim. Это объём
  отдельного плана уровня недели работы. В рамках pixel-perfect
  MainScreen достаточно показывать поле «Накопленный опыт» — оно
  должно быть ненулевым после первых боев.
- **Как чинить**: план 73 «be_points use в бою». См. подробное описание
  в `docs/plans/72.1-post-remaster-stabilization.md` ч.17 раздел про
  `be_points`. Источники формул: legacy `Mission.class.php:1230,1659,1675`,
  `EventHandler.class.php:363,1156`, oxsar2-java `Participant.java:524-525,963`.
- **Приоритет**: M (фича не блокирует игру, но баланс боя без use
  отличается от legacy — атакующий не может бустить юниты).

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

### [EXPEDITION] expeditionLost — ЗАКРЫТО
- **Статус**: закрыто 2026-04-24 (план 17 B1). `expLoss` теперь удаляет
  все `fleet_ships`, удаляет return-event (через `pl.ReturnEventID`) и
  выставляет `state='done'`. Паритет с legacy `Expedition::expeditionLost`
  (`sendBack=false`).

### [EXPEDITION] black_hole (4) и unknown (13) не реализованы
- **Где**: `fleet/expedition.go`.
- **Что**: исходы black_hole и unknown никогда не выбираются.
- **Почему**: black_hole — пустая заглушка даже в legacy (`blackHole()` — пустой метод, нулевой вес);
  unknown — зарезервирован. Полное исчезновение флота в legacy реализовано через `expeditionLost`, не black_hole.
- **Как чинить**: black_hole — новая фича поверх legacy (план 17 блок B1); unknown — план 17 блок B2.
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

### [P85.1] check-duplicates — нет unit-тестов на сам инструмент
- **Где**: `projects/game-nova/backend/cmd/tools/check-duplicates/main.go`.
- **Что упрощено**: нет тестов парсера шапки / нормализации import-prefix /
  diff-печати. Утилита проверяется только smoke'ом (план 85 Ф.4) и фактом
  zero-exit на чистом репо.
- **Почему**: явное решение в плане 85 §«Trade-offs» — «скрипт простой,
  тесты на него — overengineering. Если поломается — починим по факту».
- **Как чинить**: добавить `main_test.go` с фикстурами в `testdata/`
  (group по списку путей; drift в impl-строке; per-module import-prefix).
- **Приоритет**: L. Триггер: первый раз когда инструмент пропустит реальный drift.

### [P85.2] metrics.go в game-nova вынесен из drift-группы (не унифицирован)
- **Где**: `projects/game-nova/backend/pkg/metrics/metrics.go` (DUPLICATE-маркер
  снят); `projects/{identity,portal,billing}/backend/pkg/metrics/metrics.go`
  (группа из 3 копий).
- **Что упрощено**: вместо разделения на общий каркас + game-specific add-on
  (отдельный файл `metrics_game.go` или модуль `pkg/metricscore`) — оставили
  game-nova-копию монолитной и вне drift-чека. Инструмент check-duplicates
  её не сверяет.
- **Почему**: разделение требует выделения общих типов в shared-локацию,
  чего у нас нет (план 85 §«Не цель» запрещает shared-модуль). Текущее решение
  даёт честное отражение реальности: 3 модуля синхронны, 1 — расширен.
- **Как чинить**: при появлении 2-го game-расширения metrics — выделить
  `pkg/metricscore/` (базовый каркас) + `pkg/metrics/` (game-extras), 4 раза
  скопировать metricscore с DUPLICATE-маркером.
- **Приоритет**: L.

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

### [config] Поля Config.Game NumGalaxies/NumSystems — функционал отсутствовал
- **Где**: `projects/game-nova/backend/internal/config/config.go` (поля)
  + сейчас уже подключены в координатной валидации, генерации стартовой
  планеты и расчёте расстояний (план 72.1 часть 12, 2026-04-29).
- **Что было**: поля в `GameConfig` парсились из env, но **в runtime
  никем не читались** — координаты не валидировались, генерация
  планет работала с захардкоженным диапазоном.
- **Что было сделано**: аудит legacy-PHP `projects/game-legacy-php/`
  показал что эти параметры **не мёртвые, а недостающий функционал**:
  валидация координат (clamp galaxy ≤ NUM_GALAXIES, system ≤ NUM_SYSTEMS),
  генерация стартовых планет в нужном диапазоне, кольцевая топология
  систем (`min(|s1-s2|, NUM_SYSTEMS-|s1-s2|)`) для расстояний и времени
  полёта. Реализация — в часть 12 плана 72.1.
- **Приоритет**: closed.

### [config] DEATHMATCH-режим — не реализован, требует детального аудита
- **Где**: `projects/game-nova/backend/internal/config/config.go`
  поле `Config.Game.Deathmatch` — определено, парсится, **в runtime
  никем не читается** (план 72.1 часть 12, 2026-04-29).
- **Что**: в legacy-PHP `DEATHMATCH` — не один флаг, а **набор
  переключателей для 11 фич**, обычно «выключающих» в DM-режиме часть
  игровой системы. По первичному аудиту `consts.php` найдено:

  | # | Константа в legacy | Что делает в DM |
  |---|--------------------|------------------|
  | 1 | `NEWBIE_PROTECTION_ENABLED = !DEATHMATCH` | защита новичков отключается |
  | 2 | `ALIEN_ENABLED = !DEATHMATCH` | NPC-инопланетяне (план 66) выключены |
  | 3 | `EXPEDITION_ENABLED = !DEATHMATCH` | экспедиции выключены |
  | 4 | `MISSION_HALTING_OTHER_ENABLED = !DEATHMATCH` | удержание чужих флотов выключено |
  | 5 | `MAX_BUILDING_LEVEL = DEATHMATCH ? 35 : 40` | потолок зданий ниже |
  | 6 | `MAX_RESEARCH_LEVEL = DEATHMATCH ? 35 : 40` | потолок исследований ниже |
  | 7 | `PROFESSION_CHANGE_MIN_DAYS = DEATHMATCH ? 7 : 14` | смена профессии в 2 раза чаще |
  | 8 | `SHOW_USER_AGREEMENT = !DEATHMATCH` | UA скрыт (быстрый старт) |
  | 9 | `SHOW_DM_POINTS` | особая шкала «DM-очки» (показ override через локальный consts) |
  | 10 | `EXCH_INVIOLABLE = DEATHMATCH ? 0 : 2` | биржа: нет неприкосновенных лотов |
  | 11 | `ADMINS = DEATHMATCH ? [] : [...]` | пустой список админов в DM |

  Заметки:
  - **Не интересуют** (вне scope ремастера / не пригодятся): `ACHIEVEMENTS_ENABLED`
    и `TUTORIAL_ENABLED` — пользователь явно сказал, что эти 2 переключателя
    игнорируем (упоминаются в legacy-аудите для полноты, но не портируем).
  - **Остальные 11 — требуют детального исследования**: каких именно
    точек интеграции в Go-коде касаются, есть ли уже готовый infrastructure
    (например, план 70 ачивки, план 66 alien AI), нужно ли мигрировать
    constant-формулы (35 vs 40, 7 vs 14) на YAML, или они — глобальный
    config независимо от вселенной.
  - Для каждого пункта надо сверить с актуальным состоянием Go-кода
    (план 72.1 часть 11 — аудит legacy/Go-кода был проведён только
    поверхностно; перед реализацией DEATHMATCH-режима — повторить
    точечно по каждому из 11 пунктов).
- **Почему отложили**: это полноценная фича-вселенная (~24+ часов),
  не просто константа. game-nova на 2026-04-29 не имеет ни одной
  DM-вселенной запущенной, бизнес-приоритета нет.
- **Как чинить**:
  1. Детальный аудит каждого из 11 пунктов: что в legacy + что в Go
     (есть/частично/нет) + точки интеграции.
  2. Решение для каждого: портировать (1:1 с legacy), адаптировать
     (game-nova подход) или skip (не нужно в новой игре).
  3. Реализация — отдельным планом (72.2 или 86, по нумерации).
  4. Завязать `cfg.Game.Deathmatch` (или `universes.yaml` поле
     `deathmatch:`) на условную логику в каждой точке.
- **Приоритет**: L — нет вселенной, никто не использует. Но **не
  удалять** поле из Config — оно сразу понадобится при реализации.

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

### [Alien AI] HALT/HOLDING/HOLDING_AI — ЗАКРЫТО (план 15, этапы 1–2)
- Закрыто: цикл ATTACK=35 → HALT=36 → HOLDING=34 → HOLDING_AI=80.
  HALT (12–24ч) переходит в HOLDING (до 15 дней), внутри HOLDING каждые
  12–24ч тикает HOLDING_AI с одним из 2 действий (unload/extract).
  Alien-флот участвует в обороне планеты против сторонних атакующих.
  API платежа `/api/alien/holding/{id}/pay` продлевает HOLDING по
  формуле 2ч/50 кредитов. Остатки: FLY_UNKNOWN=33, GRAB_CREDIT=37,
  CHANGE_MISSION_AI=81 — Этап 3 плана 15.

### [Alien AI] HOLDING_AI — 6 из 8 действий пустые (как в legacy) — ЗАКРЫТО
- **Закрыто**: 2026-04-28, план 66 Ф.4. `HoldingAIHandler` переписан
  в `internal/origin/alien/holding_ai_handler.go` с равновесным
  выбором 1 из 8 веток (origin AlienAI:940-947). 2 активные
  (`SubphaseExtractAlienShips`, `SubphaseUnloadAlienResources`),
  6 заглушек (`SubphaseRepairUserUnits` / `AddUserUnits` /
  `AddCredits` / `AddArtefact` / `GenerateAsteroid` /
  `FindPlanetAfterBattle`) с audit-log на каждом вызове.
  Регистрация в worker'е переключена с `internal/alien` на новый
  пакет.
- **Почему остаются заглушки**: в legacy
  (`AlienAI.class.php:1086–1126`) эти 6 — тоже **пустые тела**.
  Портировать нечего; их семантика «делают ничего, но засчитываются
  тиком (control_times++)» сохранена в нашем handler'е.
- **Когда расширять**: дизайн-вопрос. Если решим давать игрокам
  подарки (артефакты, оксариты, астероиды) в HOLDING — это отдельная
  фича, не «порт».

### [Alien AI] unloadAlienResources — процент от текущих, не от захваченных
- **Где**: `internal/alien/holding.go::unloadAlienResources`.
- **Что**: подарок 7–10% от ТЕКУЩИХ ресурсов планеты. В legacy
  (`AlienAI.class.php:1053–1061`) возвращается процент от РАНЕЕ
  захваченных пришельцами ресурсов (`parent_event["data"][$res]`).
- **Почему**: у нас в payload HALT/HOLDING нет поля `captured_*` —
  loot забирает AttackHandler на этапе боя и пишет только в res_log.
  Хранить отдельно «захваченное пришельцами» = доп. поле payload.
- **Как чинить**: добавить `CapturedMetal/Silicon/Hydrogen` в
  `holdingPayload`, проставлять в spawnHalt из лоута атаки,
  использовать тут.
- **Приоритет**: L — текущая формула даёт похожий по масштабу эффект.

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

### [Alliance план 67 Ф.5.1-5.3] Frontend permissions: видимость только owner'у
- **Где**: `frontend/src/features/alliance/{DescriptionsPanel,RanksPanel,DiplomacyPanel}.tsx`,
  `permissions.ts::hasPerm`.
- **Что упрощено**: `canEdit/canManage` пропсы новых панелей принимают
  `isOwner` от `MyAlliancePanel`. Гранулярная проверка `can_change_description /
  can_manage_ranks / can_manage_diplomacy` для не-owner'а с custom-rank'ом
  на frontend пока **не работает** — все management-кнопки скрыты у не-owner'а
  (выглядит как «всё может только owner»).
- **Почему**: backend DTO `AllianceMember` (`internal/alliance/service.go::Member`)
  не возвращает `rank_id` участника, и нет endpoint'а вида
  `GET /api/alliances/me/permissions`. Без этого frontend не может резолвить
  `permissions JSONB` ранга текущего пользователя, и owner-only — единственный
  безопасный default. На текущем этапе (до создания custom-ranks этим UI)
  defaultная builtin-роль `member` всё равно имеет 0 прав, поэтому owner-only
  отражает реальное состояние.
- **Безопасность**: backend проверяет `Has` в каждом сервисном методе;
  невидимая на UI кнопка не открывает дыру — мутация всё равно вернёт 403.
- **Как чинить**: (а) добавить `rank_id` в DTO Member + расширить
  `permissions.go::LoadMembership` чтобы Member-handler возвращал
  resolved-permissions; ИЛИ (б) добавить `GET /api/alliances/me/permissions`
  → `AlliancePermissions`. Frontend берёт оттуда карту, прокидывает в
  `hasPerm(false, perm, perms)`.
- **Приоритет**: M — пока custom-ranks не используются (никто их не создал
  до этого PR), реальный UX-эффект нулевой; станет важным как только
  владельцы начнут раздавать офицерам права через создаваемый этим PR UI.

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

- **Playwright image pinned к версии 1.48.0** — расхождение с версией
  `@playwright/test` в frontend/package.json даст ошибки «browser not
  found». Нельзя просто обновить пакет без обновления image в
  deploy/Dockerfile.playwright.
  **План возврата**: при bump-е Playwright обновлять обе версии одним
  PR; в идеале — скрипт-линтер, который проверяет совпадение.
  **Приоритет**: L.

- **E2E в Docker без кеша docker buildx** — ЗАКРЫТО (2026-04-24).
  План 16 реализован: `docker/bake-action@v5` + type=gha cache в три
  scope'а (backend/frontend/playwright), разбивка билда на этапы для
  видимости в Actions UI, cache-warm job на push в main. `.dockerignore`
  добавлен (исключает .exe/node_modules/.git/docs — билд-контекст
  1.3 ГБ → ≈10 МБ).
  Ожидаемый профит: e2e cold ~10-12 мин → hit ~3-4 мин.

- **Deep-сценарии Ф.1/Ф.2 отложены** — регистрация, полный Attack-флоу,
  alliance create/invite/leave, messages compose, officers activation,
  art-market buy/sell. Каждый требует точных селекторов форм и/или
  подготовки событий в БД. Smoke-уровень (экран открывается, показывает
  ожидаемый текст) уже покрыт.
  **План возврата**: при первом регрессе — добавлять deep-сценарий
  в тот же домен. Ориентир: 1 deep-сценарий на домен в 2 недели.
  **Приоритет**: M.

- **Offline-тесты не написаны** — `context.setOffline(true)` конфликтует
  с Vite HMR WebSocket (тест зависает). Нужен Playwright middleware,
  пропускающий HMR, или production-build фронта для этих тестов.
  **План возврата**: при необходимости — отдельный project в
  playwright.config с production-билдом.
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

## 2026-04-25 — План 28 Ф.1-4: удалён configs/construction.yml

- **Где**: `configs/construction.yml` (удалён), `backend/internal/config/catalog.go`,
  `backend/internal/planet/service.go`, `backend/cmd/tools/battle-sim/main.go`,
  `backend/cmd/tools/import-datasheets/`.
- **Что упрощено**: единый источник истины для cost/front/ballistics/masking
  кораблей и обороны теперь — `ships.yml`/`defense.yml` (раньше — двойная
  загрузка через construction.yml). `cost_base` исследований —
  `research.yml`. `display_order`/`demolish`/`charge_credit` зданий —
  `buildings.yml`. ConstructionCatalog/ConstructionSpec удалены из Go.
  Удалены `productionRatesApprox` (fallback-путь) и `convert_construction.go`.
- **Почему**: формулы prod/cons/charge давно в `economy/formulas.go`,
  поля в YAML были текстовой документацией. Двойная загрузка
  (construction.yml + ships.yml с `Cost yaml:"-"`) усложняла навигацию.
- **Ф.5-7** (frontend через `/api/catalog`, валидатор) — отложены,
  отдельный заход.
- **Известный десинк**: ключи `metal_mine` (buildings.yml) ↔ `metalmine`
  (бывший construction.yml). После Ф.4 в коде остался только `metal_mine`
  (через `cat.Units` / `buildings.yml`); `metalmine` как ключ больше не
  существует. Константа `economy.IDMetalmine=1` сохранена (это ID, не ключ).
- **Тесты**: `go test ./... -count=1` зелёный, battle-sim regression
  идентичен (lancer-vs-cruiser exchange 0.05, mass-fleet сценарии без
  изменений).

## 2026-04-25 — TODO: rename `metalmine` → `metal_mine` (косметика)

**Где**: 7 файлов, ~15 мест.

**Что**: ключ/имя металлической шахты в коде и конфигах живёт в двух
вариантах:
- `metal_mine` — в `configs/buildings.yml`, `configs/units.yml`,
  `frontend/src/api/catalog.ts` (KEY_MAP), `economy/IDS` (как ID-константа).
- `metalmine` — в `configs/professions.yml`, `economy.IDMetalmine`,
  `economy.MetalmineProdMetal`, тестах, `frontend/.../ProfessionScreen.tsx`,
  имени файла иконки `frontend/public/images/units/metalmine.gif`.

**Рекомендация**: оставить **только `metal_mine`** — согласуется со
всеми другими ключами проекта (`silicon_lab`, `hydrogen_lab`,
`solar_plant`, `metal_storage`, `robotic_factory`, ...). Паритет с
legacy `na_construction.metalmine` больше не цель.

**Что переименовать**:
1. `configs/professions.yml` — `metalmine:` → `metal_mine:` (2 места).
2. `backend/internal/economy/ids.go`:
   - константа `IDMetalmine` → `IDMetalMine`;
   - ключ в `ProfessionKeyToID["metalmine"]` → `["metal_mine"]`.
3. `backend/internal/economy/formulas.go` — функция
   `MetalmineProdMetal` → `MetalMineProdMetal`.
4. `backend/internal/economy/formulas_test.go` — тесты с этим именем.
5. `backend/internal/planet/service.go` — все вызовы
   `economy.IDMetalmine` / `economy.MetalmineProdMetal`.
6. `backend/internal/planet/production_test.go` — тесты.
7. `frontend/src/features/profession/ProfessionScreen.tsx:19` —
   ключ-словарь `metalmine: 'Рудник металла'` → `metal_mine: ...`
   (должен быть синхронен с professions.yml).

**Что НЕ трогать**:
- `frontend/public/images/units/metalmine.gif` — оставить как есть.
  Это имя ассета (нет ценности в переименовании). `KEY_MAP` в
  `frontend/src/api/catalog.ts:170` (`metal_mine: 'metalmine'`) тоже
  остаётся как маппинг ключ→имя файла иконки.
- `legacy_name: METALMINE` в YAML — уже не существует (удалён вместе
  с `construction.yml` в плане 28 Ф.4).
- Документация и dev-log (упоминания `metalmine` как исторического
  факта — оставить).

**Trade-off**: чисто косметическая правка, поведение не меняется.
Тесты должны пройти после rename. Объём — ~15 правок в 7 файлах.

**Приоритет**: low (отложено как «очистка после плана 28»).

## 2026-04-25 — План 27 итерация 2: глубокая ребалансировка юнитов

- **Где**: `configs/ships.yml`, `configs/defense.yml`,
  `configs/rapidfire.yml`, `backend/internal/config/catalog.go`.
- **Что изменилось**: 9 групп изменений (27-F..27-U), все задокументированы
  в [ADR-0008](adr/0008-unit-rebalance-plan27-iter2.md). Главное:
  - Shadow attack 200 → 520 + front 5 (anti-DS-роль работает).
  - Bomber получил RF×3 vs Gauss/Plasma (anti-defense восстановлен).
  - SF cost ↓ до 6k, attack 150 → 120 (ниша anti-LF).
  - Lancer fuel 100 → 400, speed 8000 → 6000, front 8 → 6.
  - Defense shell ×1.5 (RL→3k, Plasma→150k и т.д.).
  - SSat shell 2k → 5k, cost +50% (закрыт эксплойт).
  - Front-тюнинг: транспорты/recycler/probe → 6/5; SD → 9; Bomber → 8;
    shields → 13/14.
  - Удалены per-unit ballistics/masking (dead-fields, движок их не читал).
- **Tests**: `go test ./... -count=1` зелёный, battle-sim --runs=20
  показывает целевые exchange по всем критическим сценариям.
- **Известные ограничения** (после переоценки):
  - DS-флот при 3.5× превосходстве проходит planet-defense с потерями
    ~3% — это **by design**, defense — налог, не барьер. При паритете
    1:1 defense сдерживает (exchange 0.28).
  - DS-как-защитник vs BS+SD без RF неуязвим — это специфика юнитов,
    не дизайн-проблема. Анти-DS в защите работает через Shadow-mass:
    5000 Shadow (25M) vs 5 DS (50M) → атакующий побеждает с exchange
    6.67.
  - Lancer-spam vs lite **починен** через дополнительный nerf attack
    5500 → 4000 (lancer-vs-mixed exchange 0.21 → 0.13, raid убыточен).
    Lancer-как-anti-DS ослаблен — anti-DS теперь через Shadow.

## 2026-04-26 — План 27-V/W: финальный фикс ролей (Variant Б)

После прогона расширенных симуляций (`battle-sim --matrix` +
`--groups`, см. план 27 §17-18) выявлены 2 проблемы:
1. Shadow доминирует чрезмерно (vs LF 90.91, vs Frigate 104.68 при
   равной cost — катастрофа).
2. Lancer стал trap-юнитом после 27-J (vs всё 1v1 0.02-0.13).

**Variant Б (выбран и применён)**:

### 27-V — Shadow Ship cost 5k → 15k
- `configs/ships.yml`: cost 1k+3k+1k → **3k+9k+3k**.
- Эффект: vs мирный флот 7-30 (норма), vs DS **2.22** (целевой
  коридор), vs defense 0.07-0.27 (Shadow проигрывает defense, как
  должен).
- Shadow стал специализированным anti-fleet/anti-DS, не «wunderwaffe».

### 27-W — Lancer attack + cap-per-planet
- `configs/ships.yml`: attack 4000 → **5000**, новое поле
  `max_per_planet: 50`.
- `backend/internal/config/catalog.go`: новое поле `MaxPerPlanet int64`
  в `ShipSpec`.
- `backend/internal/shipyard/service.go`: проверка
  `existing + in_queue + new ≤ cap` в `Enqueue` перед charge.
  Новая ошибка `ErrPlanetCapExceeded`.
- `backend/internal/shipyard/handler.go`: маппинг → HTTP 400.
- Эффект: Lancer боеспособен 1v1 (не trap), Lancer-spam как raid
  невозможен (нужен сбор с 10+ планет).

### Finals
- Все combat-юниты имеют уникальные ниши, **дубликатов нет**.
- Trap-юнитов **нет**.
- Полу-дубль LL≈RL — оставлено, не критично (в OGame так же).
- Сравнение с OGame: совпадает по основным паттернам, oxsar-nova даже
  **лучше** в роли Strong Fighter (получил нишу, тогда как OGame Heavy
  Fighter — trap).
- Tests: `go test ./... -count=1` зелёный.

### TODO (низкий приоритет)
- 27-Y: дать Light Laser уникальную роль (например, RF×3 vs SF). Сейчас
  LL и RL — близкие mass-defense юниты.
- **Frontend**: `frontend/src/api/catalog.ts` имеет hardcoded SHIPS/
  DEFENSE с устаревшими значениями. Нужна синхронизация (задача
  плана 28 Ф.5 — `/api/catalog` endpoint).

## 2026-04-26 — 27-W': откат cap, замена на RF Cruiser→Lancer ×45

**Где**: `configs/rapidfire.yml`, `configs/ships.yml`,
`backend/internal/config/catalog.go`, `backend/internal/shipyard/{service,handler}.go`.

**Что упрощено**: убрана game-механика `max_per_planet` (cap=50 на
Lancer/планета), которая была введена в 27-W. Заменена на чистый
ребаланс — RF Cruiser→Lancer 35→**45** в `rapidfire.yml`.

**Почему**: после sweep-тестов всех характеристик Lancer'а (план 27 §20)
выяснилось, что Lancer-spam закрывается одним RF-числом, без новой
game-механики. Это **проще и чище**:
- Нет нового поля в ShipSpec (`MaxPerPlanet`).
- Нет проверки в shipyard.Enqueue.
- Нет ошибки `ErrPlanetCapExceeded` и её handler-маппинга.
- Согласуется с принципом «counter через RF» (как Cruiser×6 vs LF,
  Bomber×20 vs RL).

**Что удалено**:
- `configs/ships.yml`: поле `max_per_planet: 50` у Lancer.
- `backend/internal/config/catalog.go`: поле `MaxPerPlanet int64` в
  `ShipSpec`.
- `backend/internal/shipyard/service.go`: блок «2.5. Per-planet cap»
  в `Enqueue`, переменная `ErrPlanetCapExceeded`.
- `backend/internal/shipyard/handler.go`: case `ErrPlanetCapExceeded`
  → HTTP 400.

**Tests**: `go test ./... -count=1` — все 26 пакетов зелёные.
Battle-sim: lancer-vs-mixed defender wins 100% (atk loss 96%, def loss
81%, exchange 0.10) — Lancer-spam убыточен.

## 2026-04-26 — security: удалена продажа ресурсов за кредиты (`to_credit`)

**Где**: `backend/internal/market/{service,handler}.go`,
`frontend/src/features/market/MarketScreen.tsx`.

**Что удалено**: направление `direction: "to_credit"` в endpoint
`POST /api/planets/{id}/market/credit`. Это была функция «продать
ресурс (metal/silicon/hydrogen) за premium-кредиты».

**Почему**: уязвимость экономики. Игрок мог:
1. Производить ресурсы через mines (бесплатно, runtime).
2. Конвертировать их в premium-валюту (кредиты).
3. Кредиты тратятся в shop на artefacts/officers/premium-фичи и
   monetary value (через payment-систему обратное направление есть:
   real money → credits).
4. Итог: бесконечный фарминг premium-валюты → ломает payment-балaнс.

**Что осталось**:
- `direction: "from_credit"` (купить ресурс за кредиты) — это
  legitimate-направление: игрок, потративший real-money на credits,
  может конвертировать их в ресурсы. Этот путь **расходный**, не
  фарминг.
- Backend принимает `direction = ""` или `"from_credit"`; `"to_credit"`
  возвращает `ErrInvalidResource` (HTTP 400).
- Frontend отдаёт только `"from_credit"` в payload.
- Поле `Direction` в `CreditExchangeResult` сохранено для совместимости
  (всегда `"from_credit"`).

**Что НЕ тронуто**:
- Обмен ресурс↔ресурс (`/api/planets/{id}/market/exchange`) — нет
  premium-валюты, не уязвимость.
- Покупка credits через payment-систему (real money → credits).
- Чтение баланса credits (`GET /api/artefact-market/credit`).

**Tests**: `go test ./... -count=1` — все 26 пакетов зелёные. Frontend
typecheck не запущен локально (нет npm), правки минимальны и проверены
визуально.

## 2026-04-26 — План 30 Ф.1: Goal Engine backend (за флагом)

**Где**: `backend/internal/goal/` (новый пакет, 9 файлов),
`migrations/0065_goal_progress.sql`, `configs/goals.yml`,
`configs/features.yaml` (новый флаг).

**Что**: backend-движок для замены achievement + dailyquest. Определения
целей — в YAML (как остальной content проекта, согласно плану 28),
БД хранит только `goal_progress` и `goal_rewards_log`.

**Архитектура**:
- `Catalog` загружается из `configs/goals.yml` при старте, валидирует
  все цели (категория, lifecycle, граф зависимостей, циклы).
- `Engine` — Recompute/OnEvent/Claim/MarkSeen/List. Все мутации
  goal_progress в транзакции.
- Conditions через registry: `RegisterSnapshot` / `RegisterCounter`
  типизированные функции в `conditions/`.
- `Rewarder` атомарно зачисляет credits + ресурсы.
- `Notifier` пишет в inbox при completion.
- Period_key: '' permanent, 'YYYY-MM-DD' daily, 'YYYY-Www' weekly (UTC).

**За feature flag** `goal_engine` (см. план 31 Ф.2). Сейчас flag=false,
код мёртв до Ф.5.

**Что в YAML вместо БД** (отличие от первоначального плана):
- Нет таблицы `goal_defs` — определения в `configs/goals.yml`.
- Нет таблицы `goal_dependencies` — поле `requires` в YAML.
- Review через `git diff`, type safety при загрузке, согласовано с
  планом 28.

**Tests**: 9 unit-тестов в catalog_test.go и period_test.go,
29 пакетов всего зелёные.

**Что дальше (Ф.2-7)**: HTTP API, worker hook, перенос данных, UI,
удаление старых пакетов.

## 2026-04-26 — План 31 Ф.2: feature flags

**Где**: `backend/internal/features/` (новый пакет),
`configs/features.yaml`, `backend/cmd/server/main.go`,
`backend/cmd/worker/main.go`, `frontend/src/features/flags.ts`.

**Что**: добавлены feature flags для безопасной выкатки рефакторингов.
YAML-конфиг + Go-пакет + frontend hook.

- `features.Set` иммутабелен после `Load`, atomic-safe.
- `Enabled(s, "key")` fail-closed: unknown / nil → false.
- Endpoint `GET /api/features` (без auth) — для UI conditional render.
- Frontend `useFeatureFlag('goal_engine')` через TanStack Query
  (staleTime 5min, fail-closed во время загрузки).

**Workflow**:
1. Добавить запись в `features.yaml` с `enabled: false`.
2. Обернуть новый код в `if features.Enabled(...)`.
3. Деплой — новый код есть, но не активен.
4. Тест в проде через временное `enabled: true` + restart.
5. Стабильно — `enabled: true` для всех.

**Hot-reload не реализован сознательно**: restart дёшев (план 31 Ф.1
graceful drain), предсказуем. При необходимости — добавить отдельным
итерированием.

**Tests**: 7 unit-тестов в `features_test.go`, 28 пакетов всего зелёные.

## 2026-04-26 — План 31 Ф.1: health/ready + draining state

**Где**: `backend/internal/health/` (новый), `cmd/server/main.go`,
`cmd/worker/main.go`, `backend/Dockerfile`, `deploy/docker-compose.yml`.

**Что**: добавлены liveness/readiness endpoints + draining-state
для zero-downtime deploy. При SIGTERM процесс:
1. `SetDraining()` → `/api/health` и `/api/ready` начинают возвращать 503.
2. Sleep 10s (drainDelay) — даёт nginx/docker-healthcheck время
   заметить unhealthy и убрать инстанс из ротации.
3. `srv.Shutdown()` — graceful 30s timeout для активных запросов.

**Endpoints**:
- `GET /api/health` — liveness, не делает БД-вызовов. 200 ok / 503 draining.
- `GET /api/ready` — readiness с `pool.Ping(ctx)` (timeout 2s).
  503 при draining / starting / db_unhealthy.
- Server: на основном порту (`:8080`).
- Worker: на metrics-порту (`:9091`, рядом с /metrics).

**Docker healthcheck**: добавлен в compose, `wget --spider /api/ready`
с интервалом 10s, retries 3. `frontend.depends_on` переключён на
`service_healthy`.

**buildVersion**: `var buildVersion = "dev"` в server и worker main.go,
перебивается через `go build -ldflags "-X main.buildVersion=..."` в
prod-pipeline. Возвращается в JSON-ответе `/api/health`.

**Tests**: 6 unit-тестов в `health_test.go`, все 27 пакетов зелёные.

## 2026-04-26 — План 32: multi-instance readiness

**Где**: `backend/internal/locks/` (новый), `backend/internal/scheduler/`
(новый), `backend/internal/chat/hub.go`, `cmd/worker/main.go`,
`cmd/server/main.go`, `configs/schedule.yaml`,
`deploy/docker-compose.scaling.yml`, `docs/ops/scaling.md`.

**Что**: backend готов к запуску в N≥2 инстансов:
- Postgres advisory locks (`pg_try_advisory_lock`, FNV-64 hashing)
  через `locks.TryRun(ctx, pool, name, fn)`.
- Scheduler на `robfig/cron/v3` оборачивает каждую job в advisory lock —
  ровно один worker выполняет, остальные status=skip.
- 5 singleton-задач переведены: `alien_spawn`, `inactivity_reminders`,
  `expire_temp_planets`, `event_pruner`, `score_recalc_all`. YAML —
  `configs/schedule.yaml`, ENV-overrides поддерживаются.
- `chat.Hub` рефакторен: Publish → Redis `chat:*`, на каждом инстансе
  subscriber-горутина читает и broadcast'ит локальным WS-клиентам.
  При недоступном Redis — degradation до local-only (single-instance
  поведение, без полного отказа).
- `BootstrapRecalcAllEvent` удалён — scheduler тикает по cron.
- Метрики: `oxsar_scheduler_job_runs_total{job,status}`,
  `_duration_seconds`, `_last_run_timestamp`.

**Trade-off**:
- **Catch-up пропущенных запусков НЕ поддерживается**. Если worker лежал
  во время cron-tick'а — следующий запуск через cron-период.
  Приоритет: L (можно отдельным планом, если потеря критична).
- **Chat при недоступном Redis** деградирует до single-instance:
  клиенты разных backend'ов перестают видеть друг друга, но внутри
  одного инстанса чат работает. Приоритет: L.

**Tests**: 4 unit-теста для `locks`, 13 для `scheduler`, 3 для
`chat.Hub` (single-instance fallback). Все backend-пакеты зелёные.

## 2026-04-26 — План 29 Ф.1-4: magic numbers cleanup

**Где**: `backend/internal/economy/ids.go`, `shipyard/service.go`,
`score/event.go`, `alien/alien.go`, `fleet/transport.go`,
`achievement/service.go`, `planet/service.go`.

**Что упрощено**: магические числа в SQL и Go-коде заменены на
именованные константы из `event/kinds.go` и `economy/ids.go`:

```go
// Было
WHERE kind = 70 AND state = 'wait'
WHERE b.unit_id = 1 AND b.level >= 1

// Стало (через параметр)
WHERE kind = $1                              // event.KindScoreRecalcAll
WHERE b.unit_id = %d                         // fmt.Sprintf + economy.IDMetalmine
```

**Применено**:
- В `economy/ids.go` добавлены: `IDImpulseEngine=21`,
  `IDHyperspaceEngine=22`, `IDTerraformer=58`, `IDMoonLab=350`,
  `IDCombustionEngine=20`. Группировка упорядочена.
- В `shipyard.Enqueue`: `kind = 4/5` → `event.KindBuildFleet/KindBuildDefense`.
- В `score.BootstrapRecalcAllEvent`: `kind = 70` → параметр.
- В `alien.spawnCandidates`: `kind IN (33,34,35,36)` → 4 параметра с
  `KindAlienFlyUnknown/KindAlienHolding/KindAlienAttack/KindAlienHalt`.
- В `fleet/transport.go`: 5 случаев (`mission IN (10,12)`,
  `kind = 7/20`, `mission NOT IN (15, 29)`) → параметры.
- В `achievement/service.go`: 9 случаев `unit_id = N` / `mission = N` /
  `kind = N` → `fmt.Sprintf` с `economy.ID*` / `event.Kind*`.
- В `planet.fillMaxFields`: `IN (58, 350)` → `IN ($2, $3)`.

**Обнаружен и исправлен баг STARTER-достижений** (2026-04-26):
`STARTER_BUILD_SOLARPLANT/METALLURGY/SHIPYARD/LAB` в
`achievement/service.go` исторически проверяли unit_id 3/4/21/22, что
соответствует HydrogenLab/SolarPlant/ImpulseEngine/HyperspaceEngine.
**STARTER_BUILD_SHIPYARD и STARTER_BUILD_LAB никогда не
разблокировались** — unit_id 21/22 это research, а не buildings.
Сверил с legacy oxsar2 (`sql/tutorial.sql` + `na_requirements`):
ожидаемые ID — SolarPlant=4, SiliconLab=2 (металлургический), Shipyard=8,
ResearchLab=12. Поправлено в этом же commit (см. план 29).

**TODO (план 29 Ф.5)**: рассмотреть Mode-enum (`ModeBuilding=1` и т.д.)
если в Go-коде есть `spec.Mode == 3` сравнения.

**Tests**: `go test ./... -count=1` — все 26 пакетов зелёные.

---

## 2026-04-27 — game-origin: упрощения первого запуска (план 37)

При первом запуске PHP-клона oxsar2 в `projects/game-origin-php/` было принято
несколько trade-offs ради быстрой работоспособности главной страницы.
Все они помечены `// TODO plan-37` в коде или комментариями `/* убрано */`.

### 1. Yii-виджеты в шаблонах закомментированы
**Где**: `src/templates/standard/layout.tpl`, `main.tpl`
**Что убрано**: `PrizeWidget`, `NewbieWidget`, `NewsWidget`, `NotifyWidget`,
`TutorialDialog` (все через `Yii::app()->controller->widget(...)`).
**Почему**: Это ContentBox'ы поверх основного UI (плашки про премиум, новичков,
новости, нотификации). Не критичны для рендера страницы.
**План возврата**: 37.5 — заменить на нативные PHP-блоки или включить в Universe Switcher.

### 2. socialUrl() — stub, возвращает URL как есть
**Где**: `src/core/Functions.php` + 9 использований в Stock/MSG/Battlestats и т.д.
**Почему**: Был для соц.сетей (VK/OK/MailRu iframe), у нас OAuth убран в plan-36.
**План возврата**: не требуется — функция корректна как identity для не-социального flow.

### 3. mini_games iframe (7j7.ru) удалён
**Где**: `src/game/page/Main.class.php`
**Почему**: Внешний сервис, к oxsar-nova отношения не имеет.
**План возврата**: не требуется.

### 4. CHtml::link() заменён на нативный <a href>
**Где**: `src/templates/standard/main.tpl`
**Почему**: CHtml — Yii helper, мы Yii убрали.
**План возврата**: не требуется (функционально эквивалентно).

### 5. Universe Switcher — placeholder
**Где**: `src/templates/standard/main.tpl` (бывшая ссылка "Перейти в Niro/Dominator")
**Почему**: Vanilla JS виджет — отдельная задача 37.5 в плане.
**План возврата**: 37.5 — реализовать виджет с балансом кредитов из auth-service.

### 6. Online stats (User_YII::showOnline) убран
**Где**: `src/templates/standard/main.tpl` (под `isAdmin()`)
**Почему**: Yii AR метод. Виден только админам, низкий приоритет.
**План возврата**: 37.6 — простой SQL `SELECT COUNT(DISTINCT userid) FROM na_user WHERE last > UNIX_TIMESTAMP() - 900`.

### 7. View `na_galaxy_new_pos_union2` отключён
**Где**: `migrations/001_schema.sql` (закомментирован)
**Почему**: Использует `UNION ... LIMIT 40 UNION ... LIMIT 50` без скобок —
валидно в MySQL 5.5, ломает MySQL 5.7+. В PHP-коде oxsar-nova этот VIEW не используется.
**План возврата**: при необходимости переписать с подзапросами `(SELECT ... LIMIT 40) UNION (SELECT ... LIMIT 50)`.

### 8. JwtAuth INSERT — только обязательные поля
**Где**: `src/core/JwtAuth.php::lazyJoin()` (16 NOT NULL полей na_user)
**Что**: Создаёт пользователя без планеты, alliance, achievements, stats.
**Почему**: Минимальный INSERT для работы Login flow; полная регистрация —
отдельный шаг (создание стартовой планеты по `Login.util.class.php` legacy).
**План возврата**: 37.6 — порт `UserSetup` (создание планеты, заполнение defaults, начало tutorial).

### 9. `Options` читает только из `na_config` (не из `Config.xml`)
**Где**: `src/core/Options.class.php::loadConfigFile()` отключён
**Почему**: XML-файл `Config.xml` мы не копировали (в legacy он переопределялся БД).
**План возврата**: не требуется (БД — единый источник истины для опций).

### 10. ENGINE=MyISAM/InnoDB FK миграция 003
**Где**: `na_user.global_user_id VARCHAR(36) UNIQUE`
**Почему**: VARBINARY(36) после ALTER в legacy схеме (наследие cp1251 → utf8 миграции),
наш ALTER даёт корректный VARBINARY автоматически.
**План возврата**: не требуется.


## 2026-04-27 — План 36 Ф.11–Ф.12: pre-prod sweep ЗАКРЫТО (Critical-1..6, Functional-8, Nice-10, Космет.13)

После исходной фиксации (Ф.11–Ф.12) проведена ревизия "до прода":
закрыты пп. 1–6 (security/data integrity), 8 (bootstrap retry),
10 (email из JWT) и 13 (git-tracked binaries).

Что осталось как осознанные упрощения — пп. 2 (ротация RSA-ключей,
KeySet) и пп. 11 (OAuth Social Login) внутри секции — это самостоятельные
фичи, не «фикс упрощения», ушли в отдельные итерации. И пп. 7 (голосование
кредитами) ждёт Ф.7.

Детали по каждому пункту — ниже, оригинальный текст сохранён, статусы
переведены в ✅ ЗАКРЫТО.

### 1. RSA-ключ генерируется в контейнере при первом старте — ✅ ЗАКРЫТО (2026-04-27)
**Решение**: добавлены две функции в `pkg/jwtrs`:
- `LoadKey` — только читает, при отсутствии файла — fail-fast.
- `LoadOrGenerateKey` — dev-helper (читает или генерирует).

В `cmd/server/main.go` auth-service выбор по env `AUTH_KEY_AUTOGEN`
(default `"0"` → fail-fast). В `deploy/docker-compose.yml` стоит
`AUTH_KEY_AUTOGEN: "1"` (dev). В `docker-compose.multiverse.yml` env
не задан — там читается `/run/secrets/auth_rsa_key.pem` подложенный
извне, fail-fast при отсутствии.

#### Исходный текст (для истории):

### 1. RSA-ключ генерируется в контейнере при первом старте
**Где**: `projects/auth/backend/cmd/server/main.go::run` через `jwtrs.LoadOrGenerateKey`,
volume `auth-rsa-key` в `deploy/docker-compose.yml`.
**Что**: если файл `/var/lib/auth/rsa_key.pem` отсутствует — auth-service сам
генерирует RSA-2048 и пишет в volume.
**Почему**: для dev удобно (не надо мудохаться с external secret каждый раз).
В `deploy/docker-compose.multiverse.yml` ключ уже подкладывается из
`/run/secrets/auth_rsa_key.pem`, но процесс выдачи ключа не задокументирован.
**Риск**: утечка `auth-rsa-key` volume = компрометация всех JWT.
**План возврата** (до прода): docs/ops/auth-key-management.md — генерация ключа
наружу (openssl), подкладка через Docker secret (compose) или Vault/KMS, ротация
через два-ключевой `KeySet`. Удалить fallback на генерацию из `LoadOrGenerateKey`,
оставить только Load (если файла нет — fail fast).
**Приоритет**: H.

### 2. Нет ротации RSA-ключей
**Где**: `projects/auth/backend/pkg/jwtrs/jwtrs.go::Issuer/Verifier`.
**Что**: в JWKS-выдаче и подписи участвует один ключ. При компрометации сменить
нельзя без выкидывания всех живых JWT.
**Почему**: одного ключа достаточно для MVP, ротация — не первоочередное.
**План возврата**: `KeySet` с `current` (подписывает новые токены) и `previous`
(участвует в верификации в течение grace-period 24h). JWKS отдаёт оба `kid`.
Cron-task раз в N дней генерирует новый, сдвигает current→previous, через 24h
удаляет старый.
**Приоритет**: M.

### 3. Legacy HS256 fallback всё ещё в `cmd/server/main.go` game-nova — ✅ ЗАКРЫТО (2026-04-27)
**Решение**: удалены `internal/auth/jwt.go`, `service.go` (Register/Login/Refresh),
`ratelimit.go` (никем не использовался), legacy маршруты и переменные
`JWT_SECRET`/`AccessTTL`/`RefreshTTL` из `config.go`. `AUTH_JWKS_URL`
обязателен — иначе fail-fast. Тесты `internal/auth` зелёные.
**Где**: `projects/game-nova/backend/cmd/server/main.go:151–164`,
`internal/auth/handler.go::Register/Login/Refresh`,
`internal/auth/service.go::register/login`.
**Что**: если `AUTH_JWKS_URL` не задан, server откатывается на HS256 с
`JWT_SECRET`. В Ф.11 я убрал JWT_SECRET из compose, но сам **код** legacy-режима
живёт. Если кто-то запустит без AUTH_JWKS_URL и с дефолтным секретом — security-bug.
Тесты на старые `Register/Login/Refresh` тоже остались (зелёные, проверяют
мёртвый код).
**Почему**: чистка legacy кода — большая правка (зависимости через `Service`,
referral, rate-limiter, тесты), отложена в финальный sweep после Ф.12.
**План возврата**: удалить целиком: маршруты `/api/auth/login|register|refresh`,
методы `Service.Register/Login/Refresh`, `JWTIssuer`, `password.go::HashPassword`
(unused), `cfg.Auth.JWTSecret`. В `main.go` `useJWKS` всегда true; если
`AUTH_JWKS_URL` пустой — fail fast.
**Приоритет**: H (security risk).

### 4. EnsureUserMiddleware не записывает `universe_memberships` в auth-db — ✅ ЗАКРЫТО (2026-04-27)
**Решение**: `bootstrapNewUser` в `internal/auth/ensure_user.go` теперь
дёргает `POST AUTH_SERVICE_URL/auth/universes/register` после успешного
INSERT в game-db. Параметры `UniverseID` и `AuthServiceURL` передаются
через `EnsureUserConfig`. Проверено: после register в auth-service
запись появляется в `universe_memberships`.
**Где**: `projects/game-nova/backend/internal/auth/ensure_user.go::bootstrapNewUser`.
**Что**: при lazy-create в game-db мы вызываем `starter.Assign` и `automsg.Send`,
но НЕ дёргаем `POST /auth/universes/register` в auth-service. Поэтому таблица
`universe_memberships` в auth-db остаётся пустой, а JWT при выдаче имеет
`active_universes: []`.
**Почему**: я забыл это сделать. План 36 Ф.12 явно указывал.
**Эффект**: Universe Switcher на фронте не знает, в каких вселенных юзер уже играл —
все показываются как «новые».
**План возврата**: в `bootstrapNewUser` после успешного `starter.Assign` сделать
HTTP POST на `auth-service:9000/auth/universes/register` с body
`{user_id, universe_id}`. Новая env `AUTH_INTERNAL_URL` (отличается от
JWKS_URL — может быть private VPC адресом).
**Приоритет**: H (блокирует Universe Switcher логику в multi-стек проде).

### 5. Email в JWT claims — ✅ ЗАКРЫТО (2026-04-27)
**Решение**: поле `Email` удалено из `Claims` и `IssueInput` во всех 3
DUPLICATE-копиях jwtrs. `EnsureUserMiddleware` пишет в game-db `email=NULL`.
Миграция 0068 сделала `users.email` NULLABLE. Если downstream нужен email
(admin-views и т.п.) — берёт через `auth-service GET /auth/me` с тем же
токеном.
**Где**: `pkg/jwtrs/jwtrs.go::Claims.Email` во всех 3 модулях.
**Что**: добавлено поле `email` в access + refresh токены, чтобы lazy-create
в game-db мог сделать INSERT без HTTP-вызова `/auth/me` в auth-service.
**Почему**: одна строка в БД быстрее, чем round-trip к соседнему сервису.
**Риск**: PII (email) попадает в логи если кто-то запишет токен дословно.
GDPR/privacy: токен не должен содержать персональные данные больше
необходимого минимума.
**План возврата**: убрать `Email` из claims. EnsureUserMiddleware при
`RowsAffected==1` делает HTTP `GET /auth/me` к auth-service (token forwarded),
получает email, делает второй INSERT-апдейт (UPDATE users SET email=...).
Альтернатива — не хранить email в game-db вообще (он живёт в auth-db, для
отображения брать оттуда по запросу).
**Приоритет**: M.

### 6. `bootstrapNewUser` асинхронный и без ретраев — ✅ ЗАКРЫТО (2026-04-27)
**Решение**: добавлена `retryBootstrapIfNeeded` в `ensure_user.go`. На
запросах с RowsAffected==0 (юзер уже существует) делает SELECT
`cur_planet_id IS NOT NULL` и при `false` повторно дёргает `Starter.Assign`.
Welcome повторно не отправляем (уже мог быть отправлен).
**Где**: `projects/game-nova/backend/internal/auth/ensure_user.go`.
**Что**: `starter.Assign` и `automsg.Send` запускаются в `go func()` после
`INSERT users`. Если starter падает — юзер в `users` есть, планеты нет.
Lazy-retry на следующем запросе не реализован.
**Почему**: middleware не место для тяжёлой логики, async — единственная адекватная
опция.
**План возврата**: в `EnsureUserMiddleware` дополнительная проверка: если
юзер существует, но `cur_planet_id IS NULL` — выполнить `starter.Assign`
повторно (тоже async). Альтернатива — periodic background-job, который
сканирует подвешенных юзеров.
**Приоритет**: M.

### 7. Голосование за feedback в portal не списывает кредиты
**Где**: `projects/portal/backend/internal/portalsvc/handler.go::VoteFeedback`.
**Что**: `POST /api/feedback/{id}/vote` инкрементирует `vote_count`, но
не дёргает `auth-service POST /auth/credits/spend`. ADR в плане 36 говорил
«100 кредитов за голос».
**Почему**: Ф.7 (Global Credits + платёжные webhook'и) ещё не сделана.
**План возврата**: Ф.7. До неё голосование бесплатное.
**Приоритет**: H (блокер для запуска в прод).

### 8. Refresh-токен не отзывается при logout — ✅ ЗАКРЫТО (2026-04-27)
**Решение**: `POST /auth/logout` принимает refresh-token, валидирует RSA
и кладёт его `jti` в Redis-blacklist на оставшийся TTL. Реализовано
в `internal/authsvc/blacklist.go` (`JTIBlacklist`). `Refresh` handler
проверяет blacklist перед обменом. JWT теперь содержит уникальный
`jti = kind + ":" + uuid` (раньше был общий "access"/"refresh").
Проверено: после logout refresh возвращает 401 «refresh token revoked».
**Где**: `projects/auth/backend/cmd/server/main.go` — нет роута `/auth/logout`,
`internal/authsvc/handler.go` — нет метода `Logout`.
**Что**: refresh-токен (TTL 720h = 30 дней) живёт до истечения. Logout на фронте
просто чистит localStorage; украденный refresh-токен можно использовать.
**Почему**: revocation требует Redis-хранилища revoked-jti, не сделано.
**План возврата**: реализовать `POST /auth/logout` — кладёт `jti` refresh-токена
в Redis с TTL=refreshTTL. Verifier при `kind="refresh"` проверяет наличие jti
в blacklist.
**Приоритет**: H (security, особенно если один аккаунт у нескольких людей).

### 9. Нет rate-limiting на `/auth/login` и `/auth/register` — ✅ ЗАКРЫТО (2026-04-27)
**Решение**: `internal/authsvc/ratelimit.go` — IP-based лимитер на Redis
(INCR + EXPIRE). Лимиты: login 5/min, register 10/min, refresh 30/min,
logout 30/min. IP берётся из `X-Forwarded-For` / `X-Real-IP` (для
работы за nginx). Fail-open при ошибке Redis (production может перейти
на fail-closed). Проверено: 6-й login за минуту получает 429.
**Где**: `projects/auth/backend/cmd/server/main.go` — middleware-цепочка в `r.Use(...)`
не содержит rate-limiter.
**Что**: brute-force атака на login не блокируется. В game-nova был
`auth.NewIPRateLimiter` (20 req/min/IP), я его убрал в Ф.12 как мёртвый код,
в auth-service ничего не появилось.
**Почему**: фокус был на функциональности, не security.
**План возврата**: portировать `IPRateLimiter` в `oxsar/auth/internal/auth`
(либо использовать готовое — `httprate`, `ulule/limiter`). Применить к
`POST /auth/login` (строже — 5/min) и `POST /auth/register` (10/min).
**Приоритет**: H.

### 10. Бинарники game-nova в git — ✅ ЗАКРЫТО (2026-04-27)
**Решение**: `git rm --cached projects/game-nova/backend/{server,worker}`,
в `.gitignore` добавлены глоб-правила
`projects/*/backend/{server,worker,auth-service,portal,testseed}`.
**Где**: `projects/game-nova/backend/{server,worker,*.exe,testseed.exe,...}`
отслеживаются git-ом (`git ls-files | grep -E "backend/(server|worker)$"`).
**Что**: `go build`-артефакты закоммичены. Не моё (старая ошибка), но мешают.
**Почему**: исторически.
**План возврата**: `git rm --cached projects/game-nova/backend/{server,worker,*.exe}`
+ добавить в `.gitignore` правило `projects/*/backend/{server,worker}`,
`*.exe` уже есть в .gitignore.
**Приоритет**: L.

### 11. MySQL strict-mode отключён в game-origin docker-compose
**Где**: `projects/game-origin-php/docker/docker-compose.yml` — `mysql.command: --sql-mode=...` без `STRICT_TRANS_TABLES`
**Что**: MySQL 5.7 по умолчанию в strict-режиме; legacy SQL писалось до этого
и имеет `NOT NULL` колонки без DEFAULT (`na_planet.umi`, возможно другие),
которые рассчитывают на implicit zero. Strict-режим даёт fatal
`Field 'umi' doesn't have a default value` при `INSERT INTO na_planet`
из `PlanetCreator::setRandPos`.
**Почему**: Минимальное вмешательство, ближе к legacy. Альтернатива —
найти все NOT NULL без DEFAULT и либо передавать явно, либо ALTER
DEFAULT через миграцию (рискованно — может оказаться много колонок).
**План возврата**: 37.6+ (фикс игровых дыр) — провести аудит всех `NOT NULL`
без `DEFAULT` в legacy-схеме, добавить `DEFAULT` через миграцию или явные
значения в коде, включить strict обратно.
**Приоритет**: средний (для прода важно — strict ловит реальные баги, но
для PHP-клона как промежуточного этапа допустимо).

### 12. Email::sendMail — только error_log, без реальной отправки
**Где**: `projects/game-origin-php/src/core/util/Email.util.class.php`
**Что**: clean-room rewrite после плана 43 пишет уведомления в
PHP error_log вместо реальной email-доставки.
**Почему**: Не выкатывали SMTP-конфиг в prod до публичного запуска;
для dev-окружения логирование достаточно (видно в `docker compose logs php`).
В legacy Recipe Email-классе была сложная mail() обвязка с множеством
параметров — упростили до 3-аргументного API (to/subject/body), который
покрывает все 4 caller-сайта (AccountCreator при регистрации,
Preferences при смене email/пароля, Preferences resend activation).
**План возврата**: при подготовке к публичному запуску — добавить
`symfony/mailer` в composer.json, реализовать sendMail() через
SMTP/Mailtrap/Brevo, конфиг через ENV (SMTP_HOST/PORT/USER/PASS/FROM).
**Приоритет**: средний — без email регистрация проходит, но юзер не
получает activation-letter / lost-password-letter.

### 13. Cache::buildUserCache — stub, файл-кеш сессий не используется
**Где**: `projects/game-origin-php/src/core/Cache.class.php::buildUserCache`
**Что**: Метод записывает пустой `$item = array()` в session-cache файл
вместо реального снапшота юзера.
**Почему**: Legacy session-таблица `na_sessions` в порте не используется
(аутентификация — JWT через game-nova/auth, lazy-join через JwtAuth).
User::loadData() читает из na_user напрямую по `$_SESSION['userid']`,
session-кеш не нужен.
**План возврата**: оставить как stub. Если в будущем понадобится
session-cache (например, для локальной офлайн-разработки) — реализовать
через Redis или nostalgic file-cache.
**Приоритет**: L (никогда — функционал не нужен).

### 14. Plugin.abstract_class.php — пустой stub
**Где**: `projects/game-origin-php/src/core/plugins/Plugin.abstract_class.php`
**Что**: Файл существует только потому что autoloader-config Recipe
($includingFiles в AutoLoader.php) делает require_once с die() при
отсутствии. Внутри — пустой `abstract class Plugin {}`.
**Почему**: Никто не extends Plugin в проекте (plugin-система Recipe
не используется). Полное удаление файла требует правки autoloader-config'а.
**План возврата**: вместе с rewrite AutoLoader.php (если будет необходим)
или отдельной мини-задачей — убрать `plugins/Plugin.abstract_class.php`
из `$_static_includes` в AutoLoader.php и удалить файл + папку plugins/.
**Приоритет**: L (5 минут работы, делать когда будет повод трогать
AutoLoader).

---

## 2026-04-28 — План 65 Ф.1: KindDemolishConstruction

### [65-Ф.1] Нет публичного API для demolish — только handler
**Где**: `projects/game-nova/backend/internal/event/handlers.go` —
`HandleDemolishConstruction` реализован end-to-end; но
`building.Service.Demolish()` и `POST /api/planets/{id}/demolish`
отсутствуют.
**Что упрощено**: handler принимает события `KindDemolishConstruction`,
кто бы их ни вставил в `events`. Нет ни service-метода, ни REST API,
которые игрок мог бы вызвать.
**Почему**: задача плана 65 Ф.1 — создать **эталонный handler**
для последующих 6 Kind'ов плана 65, не полноценную demolish-фичу.
Полный путь (service + endpoint + i18n + OpenAPI + frontend)
расширил бы scope до ~+400 строк и нарушил «один Kind за сессию».
**План возврата**: отдельный план «building/demolish API» — добавить
`Service.Demolish(ctx, userID, planetID, unitID)` зеркально к
`Enqueue` (списание ресурсов 0%, расчёт времени по building catalog,
INSERT в construction_queue + INSERT events Kind=2), POST endpoint
с rate-limit и Idempotency-Key (R9), UI-кнопка в Constructions
экране. Ориентир — план 67/68/72 (UI-фаза для origin).
**Приоритет**: M — без UI игрок не может снести здание, но handler
уже работает.

---

## 2026-04-28 — План 65 Ф.2: KindDeliveryArtefacts

### [65-Ф.2] Active-артефакт сбрасывается в held без revert эффектов
**Где**: `projects/game-nova/backend/internal/event/handlers.go` —
`HandleDeliveryArtefacts`, ветка `state = 'active' → 'held'`.
**Что упрощено**: при доставке артефакта, который оказался в `active`
(теоретически возможно, если биржевой код не сбросил состояние перед
выставлением лота), handler переводит его в `held` с обнулением
`activated_at`/`expire_at`, но НЕ зовёт `applyChange(revert)`
синхронно — то есть значения колонок в `users.*`/`planets.*`
(если артефакт был типа `factor_user` / `factor_planet`) остаются
с применённой дельтой у старого владельца.
**Почему**: nova вычисляет effect-стек по списку **активных**
артефактов на каждом чтении (см. `artefact/service.go:349`
`ActiveBattleModifiers`, `effects.go`). Для `battle_bonus` revert
не нужен — стек пересобирается. Для `factor_*` (которые
применяются через `applyChange` и остаются в колонках до явного
Deactivate) — да, есть теоретическое расхождение, но биржевая
операция плана 68 обязана ставить артефакт в `held` ДО полёта,
тогда delivery просто переписывает владельца. Полноценный
синхронный revert требует доступа к `artefact.Service` из
event-пакета (циклический импорт) либо вынести `applyChange` в
общий пакет — это +200 строк рефакторинга, нарушает «один Kind
за сессию».
**План возврата**: если в проде поймаем `active`-артефакт в
delivery (метрика `oxsar_delivery_artefacts_active_count`) —
вынести `artefact.RevertChange` в публичный API пакета `artefact`
без зависимости на `Service`, вызывать из handler'а через
конструктор-инъекцию (паттерн `transportSvc.ArriveHandler()`).
**Приоритет**: L — сценарий не должен возникать при корректной
работе биржи плана 68; защита через инвариант «лот = held».

### [65-Ф.2] Per-universe проверка делается через JOIN на каждом артефакте
**Где**: `HandleDeliveryArtefacts` — `SELECT (... universe_id ...) =
(... universe_id ...)` для каждого `artefact_id` в payload.
**Что упрощено**: при `len(ArtefactIDs) = N` делаем N+1 запросов
к БД (1 за артефакт + 1 проверка вселенной + 1 UPDATE), вместо
batch-чтения через `WHERE id = ANY($1)` и одной проверки
вселенной для всех.
**Почему**: типичная доставка — 1-3 артефакта (биржевой лот). При
N=3 это ~9 запросов внутри одной транзакции — приемлемо. Batch
требует динамического SQL для UPDATE `CASE`-логики (active→held)
и усложняет idempotent-skip per-artefact.
**План возврата**: если в проде увидим payload с N>10 — переписать
на batch-SQL.
**Приоритет**: L.

---

## 2026-04-28 — План 66 Ф.3: AlienAI Kind handlers

### [66-Ф.3] ChangeMissionAI replan-mode не пересобирает alien-флот
**Где**: `projects/game-nova/backend/internal/origin/alien/handlers.go`
— `ChangeMissionAIHandler`, ветка `remaining >= ChangeMissionMinTime`.
**Что упрощено**: при срабатывании AI «передумал миссию» (≥8h до
прибытия) handler обновляет `parent.payload`:
- `mode` → random из {Attack, FlyUnknown}
- `power_scale` = `1 + control_times*1.5`
- `control_times++`

**НЕ обновляется** `parent.payload.ships` (alien-флот). В origin
(`AlienAI.class.php:884`) при replan вызывается `generateMission()`
заново, который генерирует новый флот через `generateFleet` под
новый `power_scale`. Без этого power_scale-инкремент не отражается
на фактической силе при бою.
**Почему**: replan требует доступа к `loader.LoadPlanetShips` +
`loader.LoadUserResearches` (для ship_target в `GenerateFleet`),
что разворачивает scope до полноценного `GenerateMission`-pipeline.
Это работа Ф.4 (Spawner-проводка через pgx-Loader).
**План возврата**: Ф.4 плана 66 — выносим `GenerateMission(ctx,
loader, cfg, target)` в отдельную функцию, переиспользуем её
в `internal/alien.Service.Spawn` и в `ChangeMissionAIHandler.replan`.
**Приоритет**: L — power_scale всё равно учитывается в HOLDING_AI
subphase duration (`HoldingAISubphaseDuration` использует
`control_times`); потеря — только в финальном Assault при
изменённой миссии. Сценарий редкий (60% шанс CHANGE_MISSION_AI ×
≥8h до прибытия), не блокирует MVP.

### [66-Ф.3] Spawner internal/alien.Spawn не использует origin/alien.GenerateFleet
**Где**: `projects/game-nova/backend/internal/alien/alien.go::Spawn`,
`scaledAlienFleet` помощник.
**Что упрощено**: текущий nova-spawner использует `scaledAlienFleet`
(простой алгоритм 90-110% от `defPower`), а не `origin/alien.GenerateFleet`
(полный порт PHP:405-622 с поддержкой Death Star / Transplantator /
Armored Terran / Espionage Sensor / Alien Screen).
**Почему**: переключение spawner'а потребует:
- читать catalog для построения `[]ShipSpec` (alien_available_ships);
- читать `target.Ships` через `loader.LoadPlanetShips`;
- мигрировать payload-формат с `alienPayload{tier}` на
  `MissionPayload{ships, control_times, ...}`;
- обновить существующий `internal/alien.Service.AttackHandler` под
  новый payload — это ломает текущие 4 рабочих handler'а
  (Attack/Halt/Holding/HoldingAI).

В Ф.3 цель — добавить новые handlers без регрессий, поэтому
spawner не трогаем.
**План возврата**: Ф.4 плана 66 — после расширения HoldingAI до 8
действий и перевода `holdingPayload` под typed-схему, заменить
`scaledAlienFleet` на `origin/alien.GenerateFleet` через loader.
**Приоритет**: M — текущий nova-флот всё равно работает (тесты
плана 15 зелёные), но не использует UNIT_A_* по тиру и не
учитывает спец-юниты цели. Origin-паритет = после Ф.4.

### [66-Ф.3] Idempotency для Mission-handler'ов через worker, не payload
**Где**: `projects/game-nova/backend/internal/origin/alien/handlers.go`
— все три handler'а (FlyUnknown, GrabCredit, ChangeMissionAI).
**Что упрощено**: handlers не делают повторных no-op проверок
по событию (как demolish'ный `cur <= TargetLevel`). Защита от
двойного грабежа / двойного подарка возложена на worker:
`FOR UPDATE SKIP LOCKED` per-event + `state` enum переходит в
`done`/`error` после handler. Retry на handler не приводит к
повторному списанию, потому что worker не вызывает handler
второй раз для уже-обработанного event.
**Почему**: нет естественного «target_level» в alien-механиках —
грабёж по семантике origin делается ровно один раз, что и
обеспечивает worker. Дополнительная защита (например, флаг
`grabbed_at` в users) увеличила бы schema без выгоды.
**План возврата**: при появлении сценария с ручным retry (admin
заново триггерит alien-event) — добавить `grab_idempotency_key`
в users или в `messages` таблицу. Сейчас не нужно.
**Приоритет**: L — соответствует общему паттерну event-loop в nova
(см. эталон HandleDemolishConstruction).

### [67-Ф.2] RBAC permissions через guards в сервисе, не HTTP-middleware
**Где**: `projects/game-nova/backend/internal/alliance/permissions.go`,
вызовы `Has(ctx, tx, mem, perm)` из методов сервиса.
**Что упрощено**: continuation-промпт описывал «middleware-decorator»
для проверки прав на каждом alliance-action. Реализация — explicit
guards внутри сервисных методов внутри транзакции.
**Почему**: для большинства действий нужно сначала прочитать
`alliance_id` участника из `alliance_members` — а это работа сервиса.
Middleware пришлось бы дублировать SELECT по `alliance_members`
и/или принимать `alliance_id` отдельным параметром (что не для всех
URL применимо: `Leave` не содержит alliance_id в URL). Guards в
сервисе используют тот же tx, что и основная операция, что даёт
консистентность view (перед мутацией).
**План возврата**: если появятся endpoints, где membership можно
определить чисто из URL (`/api/alliances/{id}/...`) и где не нужна
читать-проверка-мутация в одной tx — можно вынести thin middleware,
читающий membership и проверяющий permission из header. Пока
смешанная картина (есть endpoints без alliance_id в URL — Leave,
MyAlliance) — guards дешевле.
**Приоритет**: L — функционально эквивалентно middleware,
переключение тривиально при необходимости.

### [67-Ф.2] alliance_audit_log для disband не пишется
**Где**: `projects/game-nova/backend/internal/alliance/service.go`
`Disband` — нет вызова `writeAuditTx`.
**Что упрощено**: при роспуске альянса (DELETE) запись «alliance_disbanded»
в `alliance_audit_log` не создаётся.
**Почему**: ON DELETE CASCADE на `alliances.id` удаляет все записи
audit-лога этого альянса вместе с самим альянсом — добавленная запись
была бы немедленно потеряна. Сохранять `disbanded` событие без
soft-delete смысла нет.
**План возврата**: если потребуется глобальная история действий
(например, для GM-аудита/анти-чита), создать `alliance_history`
без FK CASCADE, или сделать soft-delete `alliances.deleted_at`. Сейчас
нет потребителя — преждевременная функциональность.
**Приоритет**: L — потеря записи о disband не критична для UI лога
(альянс с удалёнными FK всё равно недоступен через UI).

## 2026-04-28 — План 65 Ф.3-Ф.4: Building Destruction (Kind=26/29)

### [65-Ф.3] Эвристика «у атакующего есть здание сравнимого уровня» НЕ реализована
**Где**: `projects/game-nova/backend/internal/fleet/destroy_building.go` —
`tryDestroyBuilding`, ветка random-выбора цели.
**Что упрощено**: legacy origin (Assault.class.php:253-272, константа
`DESTROY_BUILD_RESULT_MIN_OFFS_LEVEL`) при random-выборе target_building
исключает здания, у которых уровень атакующего ниже defender'a более
чем на N. В nova исключение не реализовано — random выбирает из всех
buildings планеты, кроме UNIT_EXCHANGE/UNIT_NANO_FACTORY (origin-фильтр
сохранён).
**Почему**: эвристика — балансировочный компромисс legacy конкретного
сервера, без аналитики «зачем именно так». Без её портирования миссия
становится более прямолинейной (random eligible building), что
соответствует упрощённому подходу nova. Реализация эвристики потребует
загрузки buildings всех участников ACS-группы и пер-юзер сравнения —
+30-50 строк со сложным purpose, который не воспроизводится без
мотивации legacy-балансера.
**План возврата**: если в боях обнаружится дисбаланс (атакующий c
mining-flot уверенно сносит nano factory у застроенного защитника без
шансов), добавить фильтр через JOIN buildings на attacker(-ов) и
исключать здания, для которых `defender.level - attacker.level >
threshold`. Threshold вынести в `configs/balance/origin.yaml`
(план 64).
**Приоритет**: L — без эвристики миссия работоспособна; восстановление
— на этапе балансовой настройки.

### [65-Ф.3] Не пишем `b_count`/`b_points` декремент при сносе
**Где**: `tryDestroyBuilding` — нет UPDATE `users.points/b_points/b_count`
после понижения уровня.
**Что упрощено**: legacy origin (Assault.class.php:638-643) при
target_destroyed снижает `users.points`, `b_points`, `b_count`. В nova
не делаем — score derived state (план 23): пересчитывается через
`KindScoreRecalcAll` (батч) либо decorator `withScore` per-user после
handler'а.
**Почему**: соответствует общему паттерну event-loop'а в nova
(KindBuildConstruction/KindDemolishConstruction тоже не инкрементят
points в handler'е) — установлено эталоном Ф.1 плана 65.
**План возврата**: не нужен — это согласованное архитектурное решение,
не trade-off. Запись здесь только для прозрачности отличия от legacy.
**Приоритет**: — (документация, не trade-off).

### [66-Ф.4] HoldingAI 1% recheck (`checkAlientNeeds`) не реализуется
**Где**: `projects/game-nova/backend/internal/origin/alien/holding_ai_handler.go::HoldingAIHandler`.
**Что упрощено**: на каждом тике HOLDING_AI origin
(`AlienAI.class.php:1006-1008`) с вероятностью 1% запускает
`checkAlientNeeds()` — глобальный спавн новой alien-миссии. В nova
эта ветка не реализована — спавн идёт через `scheduler.alien_spawn`
(независимая cron-задача, см. `cmd/worker/main.go::sch.Register`).
**Почему**: redundant 1%-trigger из тика HOLDING_AI излишен при
наличии глобального scheduler'а. Архитектура nova явно отделяет
spawn от tick-логики — это здоровее (spawn виден отдельной метрикой,
тестируется изолированно).
**План возврата**: не нужен — сознательное архитектурное расхождение,
не упрощение функциональности (origin тоже фактически использует
кронжоб для regular spawn'а, 1%-recheck в HOLDING_AI был лишним
fallback'ом). Запись здесь только для прозрачности.
**Приоритет**: — (документация, не trade-off).

## 2026-04-28 — План 71 Ф.1: UX-микрологика origin → nova-frontend

### [71-Ф.1] Низкоприоритетные X-NNN отложены (X-004..X-006, X-011, X-015..X-020, X-022)
**Где**: `docs/research/origin-vs-nova/nova-ui-backlog.md` записи
помечены ⏳.
**Что упрощено**: из 22 X-NNN UX-сигналов реализованы 8 ключевых
(X-001, X-002, X-003, X-007, X-008, X-009, X-010, X-013, X-014, X-021;
X-012 покрыт частью X-003). 12 низкоприоритетных отложены: X-004
(прогресс-бар заряда артефактов), X-005 (хранилище переполнено),
X-006 (блокировка ввода кораблей), X-011 (полоса rep_destroyed),
X-015/X-016 (active/max активных артефактов), X-017/X-020 (биржа —
зависит от U-001, биржа не реализована), X-018/X-019 (формы
auth/signup — у nova своя стилизация), X-022 (jQuery UI
ui-state-error — у nova своя дизайн-система, паритет не нужен).
**Почему**: 8 закрытых записей покрывают критичные UX-сигналы
(дефицит ресурсов, требования, энергодефицит, слоты, статусы);
оставшиеся 12 либо требуют ещё не реализованных модулей
(U-001 биржа), либо являются паритетом ради паритета (jQuery UI),
либо ждут DTO-расширений (added_level — после плана 70). Trade-off
зафиксирован в плане 71 как «Приоритет L» с явным обоснованием
по каждой записи.
**План возврата**: возвращать поштучно по запросу. Для биржевых
(X-017, X-020) — после реализации U-001. Для added_level (X-013) —
расширить DTO в openapi.yaml после плана 70 reaчивации achievements.
Для остальных — отдельный план «X-NNN tail» в случае нехватки
сигналов.
**Приоритет**: L.

## 2026-04-28 — План 66 Ф.5: платный выкуп удержания (alien buyout)

### [66-Ф.5] BuyoutBaseOxsars живёт в Go-Config, а не в configs/balance/origin.yaml
**Где**: `projects/game-nova/backend/internal/origin/alien/config.go`
поле `Config.BuyoutBaseOxsars` (default 100); в
`configs/balance/origin.yaml` параметра НЕТ (файла как такового тоже
нет в репо на 2026-04-28).
**Что упрощено**: ТЗ Ф.5 (промпт continuation/plan-66-fase-5-buyout.md)
требовал «параметризуй в `configs/balance/origin.yaml` как
`alien_buyout_base_oxsars`». Реально весь Config alien-сервиса (25+
параметров AlienAI: AttackInterval, GrabMinCredit, …) живёт в Go-коде
с `DefaultConfig()`, а не в YAML; per-universe override через
`balance/<universe>.yaml` запланирован отдельно (Ф.3 plan66 явно
оставила его на потом). Я следую существующей структуре пакета —
добавляю BuyoutBaseOxsars туда же, где остальные alien-параметры.
**Почему**: создание новой инфраструктуры YAML-override-loader для
одного параметра не оправдано. Альтернатива (поднять файл с одним
параметром) ломает paritет с тем, как живут все остальные alien-числа,
и заставит Ф.6/Ф.7 рефакторить.
**План возврата**: вместе с общей миграцией alien-Config на YAML
(в плане Ф.3-отложение Ф.4 при пересборке payload-формата). Когда
balance/origin.yaml появится для других alien-параметров —
BuyoutBaseOxsars переедет туда без изменения Go-API
(`DefaultConfig()` останется fallback'ом).
**Приоритет**: L.

### [66-Ф.5] UPDATE planets SET locked_by_alien=false НЕ делается
**Где**: `internal/origin/alien/buyout_handler.go::Buyout` —
закрывает HOLDING через `UPDATE events SET state='ok'` и удаляет
тики, но не трогает таблицу planets.
**Что упрощено**: ТЗ Ф.5 требовал «разблокировать планету через
`UPDATE planets SET locked_by_alien=false`». В реальной схеме
nova (миграции 0001-0080) колонки `planets.locked_by_alien` НЕТ —
блокировка планеты моделируется самим присутствием активного
KindAlienHolding event'а в `state='wait'`. Все остальные пути
закрытия HOLDING (CloseHoldingIfWiped после битвы,
closeHoldingScattered после извлечения всего флота, естественное
истечение HoldingHandler'а) — тоже не трогают planets, только
events.state. Buyout консистентно с этой моделью.
**Почему**: добавлять колонку под одну фичу неоправданно — другие
HOLDING-пути о ней не знают и не обновляют, источником истины
останутся events. ТЗ Ф.5 написан без знания реальной схемы.
**План возврата**: не требуется. Семантика «планета свободна ⇔
нет активного HOLDING-event» — корректная и единственная в коде.
**Приоритет**: — (документация уточнения, не trade-off).

### [66-Ф.5] Полный 2PC Postgres↔billing не реализован
**Где**: `internal/origin/alien/buyout_handler.go::Buyout` —
billing.Spend выполняется ВНЕ DB-tx (между двумя tx).
**Что упрощено**: если billing успешно списал оксары, а вторая
DB-tx (close HOLDING + delete тиков) упала — оператор увидит
slog.Error `alien_buyout_db_after_spend` для ручного следствия.
Полного distributed-transaction (XA / saga) между Postgres и
billing-микросервисом нет.
**Почему**: окно ~миллисекунды; повторный запрос игрока с тем же
Idempotency-Key вернёт billing-ответ без второго списания, и эта
вторая попытка успешно закроет HOLDING. План 77 (billing-client)
явно построен на идемпотентности по ключу как замене 2PC. Полный
saga-протокол излишен для buyout-объёмов (~₽250-500 типовая цена,
0-100 операций/день в начале).
**План возврата**: при превышении 1000 buyout/день или при первом
реальном случае «billing списал, DB упало» — ввести reconcile-job,
читающий `oxsar_alien_buyout_total{status="error"}` и сверяющий
billing-tx с events. Не требует кода в Buyout — внешний worker.
**Приоритет**: M.

### [65-Ф.6] Артефактный гейтинг телепорта планеты не реализован
**Где**: `internal/planet/teleport_handler.go::Teleport` (план 65 Ф.6).
**Что упрощено**: легаси-гейтинг через `ARTEFACT_PLANET_TELEPORTER`
(consts.php:132, ExtEventHandler.class.php:636) — требование артефакта
на флоте, активация при срабатывании event'а — в nova не реализован.
Единственное условие — оплата оксарами через billing-service (план 77).
**Почему**: явное решение пользователя 2026-04-28 — общий знаменатель
для всех вселенных как премиум-фича через оксары; артефакта в текущем
nova-каталоге нет. Введение артефакта расширило бы scope Ф.6 на
unit-каталог + крафт + биржу артефактов.
**План возврата**: при появлении ARTEFACT_PLANET_TELEPORTER в
configs/units.yaml (если потребуется балансовая differentiation) —
расширить TeleportConfig полем `RequireArtefact bool` и добавить
ветку в HTTP-handler'е. Не требует миграции БД (artefacts_user уже есть).
**Приоритет**: L.

### [65-Ф.6] Полный 2PC Postgres↔billing не реализован (зеркало 66 Ф.5)
**Где**: `internal/planet/teleport_handler.go::Teleport` —
billing.Spend выполняется ВНЕ DB-tx (между pre-check и event.Insert).
**Что упрощено**: если billing успешно списал оксары, а tx с
INSERT event'а упала — handler делает best-effort `Refund` сразу же
(IdempotencyKey + ":refund"). Если и Refund не дошёл (billing
полностью недоступен) — slog.Error с meta для оператора.
**Почему**: то же обоснование, что у buyout (Ф.5) — billing-API
идемпотентен по ключу, окно микросекундное, объёмы не оправдывают saga.
**План возврата**: тот же reconcile-job из Ф.5 покрывает оба
премиум-flow по `oxsar_planet_teleport_total{status="error"}` +
`oxsar_alien_buyout_total{status="error"}`.
**Приоритет**: M.

### [65-Ф.6] Integration-тесты с реальной БД не написаны
**Где**: `internal/event/teleport_handler_test.go`,
`internal/planet/teleport_handler_test.go`.
**Что упрощено**: golden-сценарии (happy-path UPDATE coords + cooldown,
occupied-slot refund, idempotent replay) написаны как pure-предикаты
(property-based) и через мок-handler без БД. Реальные SQL-сценарии
не покрыты тестами.
**Почему**: эталонные `demolish_test.go` и `delivery_artefacts_test.go`
используют схему БД, расходящуюся с актуальной nova-схемой (`planetname`,
`max_fields`, `last_update`, `universe_id` вместо реальных `name`,
`temperature_min/max`, `last_res_update` без universe_id). Те тесты,
видимо, гоняются против отдельной legacy-БД. Писать integration-тест
Ф.6 по любой из двух схем означало бы либо закреплять неработающий
эталон (legacy), либо разойтись с эталонами плана 65 Ф.1-Ф.5.
**План возврата**: после унификации test-fixture-ов плана 65
(перевод demolish/delivery_artefacts на актуальную nova-схему,
отдельным планом) — добавить teleport-integration-test тем же
seedFixture'ом. До тех пор pre-validation + property-based +
type-contract checks обеспечивают покрытие изменённой логики.
**Приоритет**: M.

### [65-Ф.6] origin.yaml override teleport_* не введён
**Где**: `configs/balance/origin.yaml` (комментарий в конце файла).
**Что упрощено**: параметры `teleport_cost_oxsars`,
`teleport_cooldown_hours`, `teleport_duration_minutes` читаются из
ENV (config.GameConfig), а не из per-universe override.
**Почему**: для introduction'а Ф.6 modern-default == origin-default
(50000 / 24h / 0min). Расширение `globalsOverride` под teleport
потребовало бы три новых поля + `applyGlobalsOverride` — больше
кода, чем выигрыш на стадии где origin/modern имеют одинаковые
значения.
**План возврата**: при первом разделении баланса (например, origin
с cost=80000, nova с cost=30000) — добавить три поля в
`balance.Globals` + `globalsOverride`, перенести чтение из
`cfg.Game.Teleport*` в `bundle.Globals.Teleport*`. Изменение
обратимо одним rebase'ом.
**Приоритет**: L.

## 2026-04-28 — План 72 Ф.2 Spring 1: каркас 7 главных экранов origin

Origin-фронт получил router + 7 главных экранов (Main, Constructions,
Research, Shipyard, Galaxy, Mission, Empire) **поверхностной глубины**:
HTML+CSS зеркало legacy `*.tpl + style.css`, базовые мутации (build /
start research / dispatch fleet / cancel queue) с Idempotency-Key,
unit-тесты на форматтеры/валидаторы/router-маршруты. По плану 72
дальнейшая «глубина» доводится итеративно через план 73
(screenshot-diff CI) и Spring 2-5 для остальных групп экранов.

### [P72.S1.A] Empire — нижние блоки (constructions / shipyard / defense / moon / research per-planet) не реализованы
- **Где**: `projects/game-nova/frontends/origin/src/features/empire/EmpireScreen.tsx`.
- **Что упрощено**: legacy `empire.tpl` помимо верхней таблицы планет
  имеет 5 матриц «здание/корабль/оборона/лунные/исследования × все
  планеты». В Spring 1 рендерится только верхняя таблица.
- **Почему**: агрегированный endpoint `GET /api/empire/buildings`
  отсутствует в `openapi.yaml`. Делать N запросов
  `/api/planets/{id}/buildings/queue` × N планет на каждом mount
  Empire-экрана — anti-pattern (план 72 R12 backlog
  «origin-фронт сразу на nova-имена API без backend-адаптеров» —
  значит, нужен один endpoint, а не fan-out).
- **Как чинить**: расширить openapi.yaml `GET /api/empire/buildings`
  → `{ planets: [{ planet_id, buildings: {<id>: <level>}, ships, defense, research_levels }] }`,
  отдельным backend-планом. Затем в EmpireScreen добавить 5 таблиц.
- **Приоритет**: M.

### [P72.S1.B] Main — нет агрегированного «обзора империи»
- **Где**: `src/features/main/MainScreen.tsx`.
- **Что упрощено**: legacy `main.tpl` показывает диаметр текущей
  планеты, температуру, очки/ранг, опыт боев, профессию, серверное
  время. В Spring 1 рендерим только имя/координаты планеты, ресурсы,
  активные миссии, счётчик непрочитанных сообщений.
- **Почему**: эти поля либо отсутствуют в `Planet` schema (diameter,
  temperature, fields), либо не имеют endpoint'а
  (`points`, `rank`, `battle_experience`, `profession`).
- **Как чинить**: расширить `Planet` schema (план 72 backend-итерация
  отдельным spring'ом) + добавить `GET /api/users/me/overview` →
  `{ points, rank, total_users, battle_exp, profession, online_15, online_24h }`.
- **Приоритет**: M.

### [P72.S1.C] Constructions/Research — нет «уровней / стоимости / времени» в UI
- **Где**: `src/features/constructions/ConstructionsScreen.tsx`,
  `src/features/research/ResearchScreen.tsx`.
- **Что упрощено**: legacy `constructions.tpl` для каждого здания
  показывает текущий уровень + стоимость следующего апгрейда + время.
  В Spring 1 рендерим только имя из CATALOG + кнопку «Построить»;
  бэкенд вычислит цену сам, фронт не показывает её до постановки в
  очередь.
- **Почему**: openapi.yaml имеет `GET /api/research` с `levels`
  (агрегированно), но не имеет аналога для зданий по планете.
  `POST /api/planets/{id}/buildings` возвращает `QueueItem` с
  `target_level`, но не «текущий level». Чтобы рисовать «level X →
  X+1» нужен `GET /api/planets/{id}/buildings/levels`.
- **Как чинить**: добавить
  `GET /api/planets/{id}/buildings` → `{ levels: {<id>: <int>} }`
  (по аналогии с `/api/research`). Затем оба экрана отрисуют уровни
  и формулы (формулы — отдельный пакет на фронте, см. план 64
  override-схемы).
- **Приоритет**: M.

### [P72.S1.D] Mission — нет распознавания «доступных кораблей» / расчёта топлива / времени полёта
- **Где**: `src/features/mission/MissionScreen.tsx`.
- **Что упрощено**: legacy `missions.tpl` рассчитывает
  `flight_time = distance / (speed × ship_speed)`, лимит грузоподъёмности
  по `ship.cargo`, расход водорода по `ship.fuel`. Spring 1 показывает
  ввод количества + speed_percent + кнопку «Отправить»; backend сам
  валидирует. Игрок не видит ETA до отправки.
- **Почему**: формулы расстояния/времени/топлива — часть
  балансировочного движка backend (`battle.calculator` / `fleet.dispatch`),
  не вынесены в openapi schema. Дублировать формулы на TS — анти-паттерн
  (R12: один источник истины).
- **Как чинить**: добавить
  `POST /api/fleet/dry-run` → `{ flight_time_seconds, fuel, max_cargo }`
  (без записи в БД, чисто расчёт). На фронте — debounced query при
  изменении ships/dst/speed.
- **Приоритет**: L (UX-улучшение, не блокер MVP).

### [P72.S1.E] Galaxy — нет alliance-tag, без иконок статусов (banned/vacation/inactive)
- **Где**: `src/features/galaxy/GalaxyScreen.tsx`.
- **Что упрощено**: legacy `galaxy.tpl` рисует alliance-аббревиатуру
  при имени игрока + цветовую дифференциацию (banned, vacation,
  inactive7, inactive21). `SystemView.GalaxyCell` в openapi.yaml имеет
  только `owner_username`, не `owner_alliance_tag` и не статусы.
- **Как чинить**: расширить `GalaxyCell` schema полями
  `{ owner_alliance_tag, owner_status: 'active'|'banned'|'vacation'|'inactive7'|'inactive21' }`.
- **Приоритет**: M.

### [P72.S1.F] Shipyard — нет «hide locked» / max-from-resources
- **Где**: `src/features/shipyard/ShipyardScreen.tsx`.
- **Что упрощено**: legacy `shipyard.tpl` ограничивает поле «Количество»
  значением `min(ship_cost_to_max_buildable, current_resources)`.
  Spring 1 показывает только `min={0}`, без max — backend рейзит 400
  если игрок ввёл больше, чем может построить.
- **Как чинить**: либо расширить `inventory` endpoint полем
  `max_buildable: {<id>: <count>}` (вычисляется по текущим ресурсам),
  либо добавить `POST /api/planets/{id}/shipyard/dry-run`. На фронт —
  debounced query на input.
- **Приоритет**: L.

### [P72.S1.G] Pixel-perfect доводка отложена на план 73
- **Где**: все 7 экранов.
- **Что упрощено**: HTML-структура и CSS-классы зеркалят legacy
  (.ntable, .galaxy-browser, .center, .button, cur-planet/cur-moon),
  но точная визуальная сверка (отступы/цвета/spacing на пиксел)
  не проводилась — нет running legacy-стека для скриншотов.
- **Почему**: pixel-perfect feedback loop возможен только с
  screenshot-diff CI (план 73 — Visual regression CI на Playwright),
  ручная сверка скриншотов даёт false negatives.
- **Как чинить**: после завершения плана 73 каждый Spring 1-5 экран
  пройдёт screenshot-diff и баги выявятся автоматически.
- **Приоритет**: M (чисто эстетика, не функционал).

### [P72.S1.H] Tests: unit-only, без рендера компонентов
- **Где**: `*.test.ts` в `src/`.
- **Что упрощено**: тесты покрывают форматтеры, idempotency,
  query-keys, catalog, router-маршруты, mission-валидацию — без
  рендера React-компонентов. Аналогичный подход в nova-фронте.
- **Почему**: добавление `@testing-library/react + jsdom` — отдельный
  инфраструктурный план; Spring 1 не блокировался на нём. Unit-тесты
  ловят 80% регрессий по логике.
- **Как чинить**: при добавлении screenshot-diff CI плана 73
  одновременно подключить testing-library + jsdom для interaction-tests.
- **Приоритет**: L.

**Связанный план**: [docs/plans/72-remaster-origin-frontend-pixel-perfect.md](plans/72-remaster-origin-frontend-pixel-perfect.md)
**Объединённое ТЗ для backend-расширений**: P72.S1.A + .B + .C + .D + .E + .F →
один план «openapi для origin-фронта» с 6 endpoint'ами.

---

## 2026-04-28 — План 72 Ф.3 Spring 2 ч.1: 12 alliance-экранов origin-фронта

### [P72.S2.A] «Текущие заявки» (own applications) — пустой блок
- **Где**: `frontends/origin/src/features/alliance/AllianceOverviewScreen.tsx`.
- **Что упрощено**: legacy `ally.tpl` показывает блок «Текущие заявки»
  пользователя (где он подал заявки на вступление, но альянс не одобрил).
  В origin-фронте этот блок не реализован — backend nova не предоставляет
  endpoint типа `GET /api/users/me/applications`.
- **Почему**: nova-API ориентирован на actor-сторону: alliance-owner видит
  pending applications через `GET /api/alliances/{id}/applications`, но
  applicant своих pending не видит. В Spring 1 закрытие выглядит как
  «в URL переход → экран альянса покажет appInProgress».
- **Как чинить**: расширить openapi.yaml `GET /api/users/me/applications`
  и подмешать блок в AllianceOverviewScreen.
- **Приоритет**: L (UX-удобство, не блокер).

### [P72.S2.B] PATCH alliance name/tag — read-only
- **Где**: `frontends/origin/src/features/alliance/AllianceManageScreen.tsx`.
- **Что упрощено**: legacy `manage_ally.tpl` имеет 2 формы для смены
  tag и name альянса. В origin это поля read-only — backend на
  `/api/alliances/{id}` экспонирует только `GET` и `DELETE`.
- **Почему**: rename альянса — критическая операция (ребрендинг,
  audit-log, рассылки). Backend плана 67 не закрыл её, а в Spring 2
  делать backend-расширения R0 запрещает.
- **Как чинить**: backend-PR на `PATCH /api/alliances/{id}` с rate-limit
  (1 раз в 30 дней) + audit-запись.
- **Приоритет**: M.

### [P72.S2.C] Memberlist preferences (sortBy / showmember) не реализованы
- **Где**: `frontends/origin/src/features/alliance/AllianceManageScreen.tsx`.
- **Что упрощено**: legacy `manage_ally.tpl` имеет настройки
  «Сортировать список членов по очкам/имени» + «Показывать список членов
  всем». В origin-фронте только toggle is_open. Серверной модели нет.
- **Почему**: эти поля в legacy жили в табличке `aks` (ally settings),
  в nova-схеме `alliances` их аналога нет. Отображение списка членов
  доступно владельцу + членам по умолчанию.
- **Как чинить**: добавить колонки `member_visibility`, `member_sort` в
  alliances + расширить openapi.yaml.
- **Приоритет**: L.

### [P72.S2.D] Granular permissions для не-owner'а UI пока не работают
- **Где**: alliance-экраны — кнопки management (manage/ranks/diplomacy/
  descriptions) видны только owner'у.
- **Что упрощено**: nova frontend и origin frontend оба ограничивают
  UI-проверки `isOwner`. Бэкенд проверяет `can_manage_ranks`,
  `can_change_description` и т.д. полноценно — если кнопка случайно
  будет видна, 403 защитит.
- **Почему**: уже зафиксировано в плане 67 (P67.S5.B): Member DTO без
  `rank_id`, поэтому frontend не может разрешить `hasPerm()` без owner'а.
  Origin-фронт наследует то же ограничение.
- **Как чинить**: общий fix с планом 67 — добавить `rank_id` (опционально
  + `effective_perms`) в Member DTO. Один patch на nova и origin.
- **Приоритет**: M.

### [P72.S2.E] Alliance audit/diplomacy/transfer без RTL-тестов
- **Где**: `src/features/alliance/*.test.ts`.
- **Что упрощено**: тесты unit-only (utility-функции, контракт роутов,
  i18n-маппинг). Рендеринг и взаимодействие (например, transfer-flow:
  step1 → code → confirm) не покрыты автотестами.
- **Почему**: testing-library/react не подключён в origin-фронте,
  backend-инвариант лежит в плане 67 alliance-tests + alliance/api.go
  (server-side покрыт).
- **Как чинить**: добавить @testing-library/react + jsdom при подключении
  screenshot-diff CI плана 73. Тогда же — interaction-тесты для transfer/
  ranks/diplomacy.
- **Приоритет**: L.

### [P72.S2.G] Repair и Battlestats — нет nova-API endpoint'ов
- **Где**: `frontends/origin/src/features/repair/RepairScreen.tsx`,
  `frontends/origin/src/features/battlestats/BattleStatsScreen.tsx`.
- **Что упрощено**: nova-API не предоставляет `/api/planets/{id}/repair`
  (повреждённые юниты + ремонт) и `/api/users/me/battles` (детальная
  история боёв). Pixel-perfect-каркас (форма фильтров, таблица) рендерим,
  но список пустой; Battlestats показывает только ранг через
  `/api/highscore/me` как proxy.
- **Почему**: backend repair-домена в nova ещё не написан (legacy ремонт
  завязан на `repair.tpl` контроллер с очередью аналогичной shipyard);
  battlestats — отдельный план агрегации боёв (планы 17/41 не покрывают).
- **Как чинить**: открыть отдельные планы — repair-домен (миграции +
  service + handler) и battlestats-агрегатор (выборка из messages folder=2
  battle reports + индекс по дате).
- **Приоритет**: M (геймплейные фичи).

### [P72.S2.H] Artefact market — DTO без названия артефакта/иконки
- **Где**: `frontends/origin/src/features/market/MarketScreen.tsx`.
- **Что упрощено**: `ArtMarketOffer` в openapi содержит только {id,
  seller_id, unit_id, price, created_at}. Имя артефакта (например,
  «Знак торговца») и описание/иконка/duration не приходят в offer-list
  endpoint, поэтому в UI отображаем «Артефакт #unit_id».
- **Почему**: backend nova `/api/artefact-market/offers` упрощён под
  legacy-механизм EXT_MODE; обогащение требует JOIN с unit-каталогом или
  ArtefactCatalog query на каждой строке.
- **Как чинить**: расширить ArtMarketOffer схемой + добавить join в
  repo_pgx.go артефактного market.
- **Приоритет**: L (UX-улучшение, не блокер).

### [P72.S2.F] Pixel-perfect доводка alliance отложена на план 73
- **Где**: все alliance-экраны.
- **Что упрощено**: HTML-структура и CSS-классы (`ntable`, `center`,
  `idiv`, `false`/`true`, `button`) идентичны legacy-PHP, но
  попиксельная сверка с legacy-screenshots не выполнена. Точная
  визуальная сверка отложена до screenshot-diff CI (план 73).
- **Почему**: то же обоснование что у P72.S1.G — без screenshot-diff
  CI ручной pixel-perfect — это бесконечная регрессия.
- **Как чинить**: план 73 (screenshot-diff CI) сравнит origin-фронт с
  legacy-PHP в Docker-режиме, отчёт ≤ 0.5%.
- **Приоритет**: L.

**Связанный план**: [docs/plans/72-remaster-origin-frontend-pixel-perfect.md](plans/72-remaster-origin-frontend-pixel-perfect.md)

---

## 2026-04-28 — План 72 Ф.4 Spring 3: artefacts/info/techtree/records

### [P72.S3.A] Catalog-endpoints — current-universe-only
- **Где**: `internal/catalog/handler.go` — `GET /api/buildings/catalog/{type}`,
  `/api/units/catalog/{type}`, `/api/artefacts/catalog/{type}`.
- **Что упрощено**: catalog отдаёт modern (nova) данные из
  `internal/economy/formulas.go` + `configs/*.yml` без
  universe-context. Фронтенд не передаёт `?universe=...` и не имеет
  способа выбрать вселенную.
- **Почему**: на 2026-04-28 origin-вселенная не запущена (план 74
  ещё не закрыт). Catalog (params + pre-computed) — read-only общий
  ресурс; R10 требует `universe_id` для **per-universe данных
  игроков** (планеты, юниты), а не для каталога. Modern (nova)
  данные — единственное полезное содержимое сейчас.
- **Как чинить**: при запуске origin-вселенной (план 74) добавить
  query-param `?universe=origin|modern` ИЛИ per-user routing на
  JWT-context (читать `user.current_universe`). Внутри handler —
  выбор `internal/origin/economy/*.go` vs `internal/economy/*.go`
  по universe_id. Ожидаемый объём: +1 if-statement в каждом из
  3 handler'ов + новый field в openapi.
- **Приоритет**: M (фича понадобится при запуске origin, не блокер
  плана 72).

### [P72.S3.B] Pixel-perfect доводка Spring 3 отложена на план 73
- **Где**: все Spring 3 экраны (artefacts / artefact-info /
  building-info / unit-info / techtree / records / ranking).
- **Что упрощено**: HTML-структура и CSS-классы (`ntable`, `center`,
  `idiv`, `false`/`true`, `button`) идентичны legacy-PHP, но
  попиксельная сверка с legacy-screenshots не выполнена.
- **Почему**: то же обоснование что у P72.S1.G и P72.S2.F — без
  screenshot-diff CI ручной pixel-perfect — это бесконечная
  регрессия.
- **Как чинить**: план 73 (screenshot-diff CI).
- **Приоритет**: L.

**Связанный план**: [docs/plans/72-remaster-origin-frontend-pixel-perfect.md](plans/72-remaster-origin-frontend-pixel-perfect.md)

---

## 2026-04-28 — План 68 Ф.1-Ф.7: биржа артефактов (backend)

### [P68.A] Currency: users.credit как оксариты (без переименования)
- **Где**: `internal/exchange/repo_pgx.go` (UPDATE users SET credit=...).
- **Что упрощено**: биржа использует существующую колонку
  `users.credit bigint` как backing-storage для оксаритов вместо
  введения отдельной `users.oxsarit` как требует ADR-0009.
- **Почему**: семантически `credit bigint` уже соответствует ADR
  «оксариты» (soft-currency, ст. 1062 ГК) — все остальные места nova
  (expedition.go, market.fleet_lots, officer.service, goal.rewarder)
  тоже используют credit как soft-currency. Переименование колонки
  затрагивает 10+ файлов — отдельный план migration-серии.
- **Как чинить**: серия миграций credit→oxsarit (rename column +
  обновление всех call-sites), отдельный план после плана 74.
- **Приоритет**: L (имя колонки не влияет на корректность).

### [P68.B] Permit-gating отключён (AlwaysAllowPermit MVP)
- **Где**: `internal/exchange/service.go` PermitChecker DI.
- **Что упрощено**: интерфейс `PermitChecker.HasMerchantPermit()`
  существует, но дефолтная реализация `AlwaysAllowPermit{}` всегда
  возвращает true. ErrPermitRequired остаётся в errors.go и i18n.
- **Почему**: legacy `Artefact::getMerchantMark()` зависит от
  premium-вселенных, которых в nova на текущий момент нет. Гейтинг
  без премиум-инфраструктуры — преждевременная оптимизация.
- **Как чинить**: добавить `DBPermitChecker` (SELECT FROM
  artefacts_user WHERE state='active' AND unit_id=ARTEFACT_MERCHANT_PERMIT)
  и активировать его в DI при включении premium-фич.
- **Приоритет**: L (фича не блокирующая).

### [P68.C] Балансовый конфиг: YAML создан, loader не подключён
- **Где**: `configs/balance/{default,origin}.yaml` секция `exchange`.
- **Что упрощено**: YAML-файлы содержат корректные значения, но
  service.go использует `DefaultConfig()` из Go-кода (значения
  продублированы). Loader из YAML не реализован.
- **Почему**: подключение loader'а требует расширения
  `internal/balance` (которая управляет buildings/research/ships) —
  это touch широко расходится по коду; в плане 68 это вне scope.
- **Как чинить**: пост-фикс плана 68 — `LoadConfigFromYAML(path)
  (Config, error)` + wiring в cmd/server/main.go.
- **Приоритет**: L (значения совпадают; пока YAML — документация).

### [P68.D] Integration-тесты с реальной БД пропущены
- **Где**: `internal/exchange/repo_pgx_test.go` отсутствует.
- **Что упрощено**: PgRepo (450 строк pgx) не покрыт unit-тестами.
  Service.go покрыт через fakeRepo (mock); event-handlers — через
  smoke-тесты (parsing payload).
- **Почему**: Integration-тесты с TEST_DATABASE_URL требуют отдельной
  CI-инфраструктуры (Postgres + миграции в pipeline); плановый объём
  68 уже большой. Service.go покрыт 80-100% по основным методам
  через mock-repo (fakeRepo); pgx-логика — преимущественно SQL-запросы.
- **Как чинить**: создать `repo_pgx_test.go` с auto-skip
  (`if os.Getenv("TEST_DATABASE_URL") == "" { t.Skip(...) }`) и
  тестами CRUD + FOR UPDATE concurrent buy.
- **Приоритет**: M (для prod-готовности).

### [P68.E] Общий oxsarit_transactions журнал не создан
- **Где**: ADR-0009 предполагает таблицу
  `game-nova.oxsarit_transactions` для всех движений soft-currency.
- **Что упрощено**: биржа имеет собственный audit `exchange_history`
  (это требование R13). Общий журнал движений credit (expedition,
  market, officer, goal, exchange) — не реализован.
- **Почему**: тех долг существует и без биржи — `expedition.go:691`
  и др. пишут `UPDATE users SET credit` без audit-trail. Решение
  по общему журналу — отдельный архитектурный план.
- **Как чинить**: отдельный план «реализация ADR-0009 audit-trail»
  с миграцией + рефакторингом всех call-sites.
- **Приоритет**: M (compliance-blocker для аналитики, не для запуска).

**Связанный план**: [docs/plans/68-remaster-exchange-artifacts.md](plans/68-remaster-exchange-artifacts.md)

## 2026-04-28 — План 73 Ф.1+Ф.2: baseline screenshots без JS

### [P73.A] CDN-зависимости заблокированы — JS-countdown в эталонах не работает
- **Где**: `tests/e2e/origin-baseline/baseline.spec.ts` —
  `BLOCKED_HOSTS` (ajax.googleapis.com, fonts.googleapis.com,
  fonts.gstatic.com, counter.yadro.ru, www.liveinternet.ru,
  cakeuniverse.ru).
- **Что упрощено**: legacy-php грузит jQuery 1.5.1 + jQuery-UI
  1.8.14 с ajax.googleapis.com. В headless Chromium эта загрузка
  висит 30-60s или таймаутит. Блокируем CDN — JS не выполняется,
  countdown'ы (стройка/исследование/флот) показывают начальные
  серверные значения, не тикают.
- **Почему**: для baseline-pixel-diff важен статический layout
  (HTML+CSS), а не JS-поведение. Серверный HTML рендерит
  начальное состояние корректно. Альтернатива — поднять local
  proxy с jQuery — overhead не окупается на этапе Ф.1+Ф.2.
- **Как чинить (если понадобится)**: захостить jQuery 1.5.1 +
  jQuery-UI 1.8.14 локально как static asset legacy-php (через
  правку `layout.tpl` либо отдельный nginx-fallback proxy).
  Альтернатива — заменить CDN на локальную копию через
  Playwright `route.fulfill`.
- **Приоритет**: L (Ф.3 pixel-diff применяет масок к динамичным
  зонам; статический snapshot достаточен).

### [P73.B] Smoke-набор 7 экранов из 22 — ЗАКРЫТО (2026-04-28)
- **Где было**: `tests/e2e/origin-baseline/screens.ts::SMOKE_SCREEN_IDS`.
- **Закрытие**: Ф.2.5 догнан в той же дате — снято всех 22 экрана
  через `SMOKE=0 bash tests/e2e/origin-baseline/take-screenshots.sh`.
  Прогон занял 4.7 минуты, 21 ok с первой попытки, 2 flaky
  (S-012-diplomacy, S-012-found) прошли с retry=1, 0 failed. Все PNG
  закоммичены в `tests/e2e/origin-baseline/screenshots/`.

**Связанный план**: [docs/plans/73-remaster-screenshot-diff-ci.md](plans/73-remaster-screenshot-diff-ci.md)

## 2026-04-28 — План 80: smoke выявил баг миграции 0005

### [P80.A] Миграция 0005_rbac_tables.sql валится на CROSS JOIN без ON CONFLICT — ✅ ЗАКРЫТО планом 84 (2026-04-28)
- **Где**: `projects/identity/migrations/0005_rbac_tables.sql` —
  финальный INSERT для роли `superadmin`:
  ```sql
  INSERT INTO role_permissions (role_id, permission_id)
  SELECT r.id, p.id FROM roles r, permissions p;
  ```
- **Что упрощено**: до плана 80 identity-БД стартовала пустой
  (Dockerfile.migrate ссылался на несуществующую `projects/auth/
  migrations`), поэтому миграции 0001-0006 не выполнялись и баг
  не проявлялся. После Ф.2 плана 80 миграция запустилась и упала
  на 0005 с `duplicate key value violates unique constraint
  "role_permissions_pkey"` (SQLSTATE 23505).
- **Почему баг**: `FROM roles r, permissions p` без WHERE — это
  CROSS JOIN для ВСЕХ ролей, включая `support`/`moderator`/
  `admin`/`billing_admin`, которым выше в той же миграции уже
  выданы permissions подмножествами. Конфликт PK
  `(role_id, permission_id)`.
- **Как починили (план 84)**: финальный INSERT переписан как
  `FROM roles r CROSS JOIN permissions p WHERE r.name = 'superadmin'
  ON CONFLICT (role_id, permission_id) DO NOTHING`. WHERE — основной
  фикс (ограничивает scope до superadmin), ON CONFLICT —
  defense-in-depth (идемпотентность повторного запуска и совместимость
  с prod-БД, где 0005 могла быть применена частично).
- **Приоритет**: H (identity-стек не поднимается с нуля).

**Связанный план**: [docs/plans/80-auth-leftovers-cleanup.md](plans/80-auth-leftovers-cleanup.md) (smoke раздел), [docs/plans/84-rbac-migration-0005-hotfix.md](plans/84-rbac-migration-0005-hotfix.md) (фикс).



## 2026-04-28 — План 72 Ф.5 Spring 4 ч.1: communication / notes / search / settings

### [P72.S4.BBCODE] BBCode чата → plain text (отложено в Ф.8)
- **Где**: `projects/game-nova/frontends/origin/src/features/chat/ChatScreen.tsx`.
- **Что упрощено**: legacy chat (`templates/standard/chat.tpl`) использовал
  BBCode-toolbar (`[b]bold[/b]`, `[url]…[/url]`, smileys через
  jQuery-плагин). В Spring 4 Spring origin-фронта чат рендерит сообщение
  как `<span style="white-space: pre-wrap">{body}</span>` — без парсинга
  BBCode и без HTML-render.
- **Почему**: TipTap-интеграция — отдельная фаза Ф.8 плана 72.
  Plain-text безопаснее (нет XSS) и приемлем для переходного периода.
  Существующие BBCode-сообщения из legacy chat_messages таблицы
  отрендерятся как литералы (`[b]hi[/b]`), что не идеально визуально, но
  не ломает UX.
- **Как чинить**: Ф.8 плана 72 — TipTap RichTextEditor + санитайзер
  на backend (см. план 57 mail-service для аналогичной задачи).
- **Приоритет**: M (визуальная косметика, функционально чат работает).

### [P72.S4.SETTINGS] Settings экран не реализует legacy-only поля
- **Где**: `projects/game-nova/frontends/origin/src/features/settings/SettingsScreen.tsx`.
- **Что упрощено**: legacy `templates/standard/preferences.tpl` (176 строк)
  включал поля `templatepackage`, `skin_type`, `user_bg_style`,
  `user_table_style`, `imagepackage`, `show_all_constructions`,
  `show_all_research`, `show_all_shipyard`, `show_all_defense`,
  `planetorder`, `esps`, `ipcheck`. В origin-фронте оставлены только
  поля, поддерживаемые backend-эндпоинтом `/api/settings`: `email`,
  `language`, `timezone`, плюс смена пароля (через identity-service)
  и удаление аккаунта по коду.
- **Почему**: legacy-only поля — это настройки legacy-PHP темы и
  персонификации, которые не имеют смысла в pixel-perfect клонe origin
  (стиль зашит в legacy CSS-классы `.ntable` / `.idiv`). `show_all_*`
  / `planetorder` — не реализовано в backend (нет соответствующих
  колонок в users), реализация = новая backend-фича, не «pixel-perfect
  клон».
- **Как чинить (если понадобится)**: добавить эти настройки сначала
  в backend (миграция + handler), затем в origin-фронт. Не блокирует
  Ф.5 ч.1, не входит в scope Ф.5 ч.2 — TBD при необходимости.
- **Приоритет**: L (legacy-only косметика).

### [P72.S4.MSG_TO] OpenAPI POST /api/messages: исправлено поле `to_username` → `to`
- **Где**: `projects/game-nova/api/openapi.yaml`.
- **Что было**: openapi-схема ошибочно описывала тело как
  `{to_username, subject, body}`. Реальный handler (`internal/message/
  handler.go::Compose`) ожидает `{to, subject, body}`. nova-фронт
  уже корректно отправляет `{to, ...}` (см.
  `features/messages/MessagesScreen.tsx:348`) — это была чистая ошибка
  документации, не runtime-баг.
- **Что сделано**: openapi приведён в соответствие с реальным
  поведением backend в Ф.5 ч.1 (одной строкой в edit'е).
- **Приоритет**: closed (документационный фикс, не trade-off).

### [P72.S4.OFFICER_DTO] OpenAPI Officer/Profession DTO приведены к реальному backend (2026-04-28)
- **Где**: `projects/game-nova/api/openapi.yaml` (схемы Officer,
  Profession, ProfessionInfo + body /api/officers/{key}/activate).
- **Что было**: Officer schema содержала только `{key, active,
  expires_at}`; backend (`internal/officer/service.go`) реально отдаёт
  полный набор полей `title/description/duration_days/cost_credit/
  effect/activated_at/expires_at`. Аналогично Profession описывала
  `{name, description, bonuses}`, а backend (`internal/profession/
  service.go`) отдаёт `{key, label, bonus, malus}` (где bonus/malus —
  целочисленные дельты уровня по техн. ключу). Body activate описывал
  `{planet_id}`, а реальный handler принимает `{auto_renew}`.
- **Что сделано**: openapi приведён в соответствие с реальным
  поведением backend в Ф.5 ч.2 (Officer schema, Profession schema,
  ProfessionInfo, request body activate). Это документационный фикс,
  не trade-off.
- **Приоритет**: closed.

### [P72.S4.WIDGETS] S-046 Widgets закрыт через S-001 Main (R15 ✅, не упрощение)
- **Где**: `projects/game-nova/frontends/origin/src/features/widgets/
  WidgetsRedirect.tsx` — Navigate → / на /widgets.
- **Что**: legacy `templates/standard/widgets.tpl` сам по себе —
  заглушка (Yii widget CurrentEvents удалён в плане 37.5d.9). В origin-
  фронте семантический эквивалент уже агрегирован в S-001 MainScreen
  (Spring 1, коммит 47d1f0ef65): события (`/api/fleet`), непрочитанные
  сообщения (`/api/messages/unread-count`), homeplanet+universe.
- **Почему**: дубликат — нет смысла иметь /widgets и / отдельными
  маршрутами с одинаковым контентом. Современный паттерн — единая
  «главная» с виджетами на ней.
- **Trade-off (R15 ✅, не упрощение)**: визуальное расхождение с
  legacy (нет отдельного /widgets маршрута). В Spring 1 уже зафиксировано
  «pixel-perfect только в рамках реализуемых экранов; semantic
  equivalence важнее визуальной точности дубликатов». /widgets
  делает Navigate → / с dev-notice в console.
- **Как чинить (если понадобится)**: воссоздать legacy-вид как
  отдельный экран, скопировав MainScreen и убрав header. Но проще —
  не делать ничего, /widgets уже работает (через redirect).
- **Приоритет**: closed.

### [P72.S4.CHANGELOG] Changelog как bundled markdown (не backend endpoint)
- **Где**: `projects/game-nova/frontends/origin/src/features/changelog/
  CHANGELOG.md` + `parse.ts` + `ChangelogScreen.tsx`.
- **Что**: список релизов хранится как markdown в bundle origin-фронта,
  а не в БД через `/api/changelog`. Backend такого endpoint'а не имеет.
- **Почему**: changelog меняется при релизах (не из runtime), markdown
  в bundle — стандартный паттерн для редко-меняющегося контента
  (документация / release-notes). Заводить таблицу в БД и админ-панель
  для редактирования релизов — лишняя работа без выгоды.
- **Trade-off**: НЕТ — это **правильный** паттерн, не упрощение. Запись
  в simplifications для документирования факта, а не как «вернуться позже».
- **Как чинить**: ничего не нужно. Если в будущем потребуется
  динамический changelog (например, в админке) — отдельный план.
- **Приоритет**: closed.

### [P72.S4.USER_AGREEMENT] UserAgreement — cross-link на portal (единственный источник истины)
- **Где**: `projects/game-nova/frontends/origin/src/features/user-agreement/
  UserAgreementScreen.tsx`.
- **Что**: /user-agreement в origin-фронте показывает короткую справку
  и ссылку `${VITE_PORTAL_BASE_URL}/user-agreement` в новой вкладке, а
  не дублирует юр-текст inline.
- **Почему**: юр-документ должен иметь **единственный источник истины**,
  иначе при правках появляется риск рассинхрона между порталом и игрой
  (юр-риск, особенно по 149-ФЗ). Портал уже хостит /user-agreement +
  /privacy с актуальным текстом (план 50). Origin-фронт делает
  cross-link — full text живёт на одной площадке.
- **Trade-off**: НЕТ (это правильный паттерн централизации юр-документов).
  Запись для документирования факта.
- **Как чинить (если изменится)**: если решим, что игроки должны видеть
  agreement без выхода из игры — встроить iframe или fetch markdown с
  портала. Но это только при явной потребности.
- **Приоритет**: closed.

### [P72.S4.SUPPORT_CROSS_SERVICE] Support → portal-backend (план 56), не game-nova
- **Где**: `projects/game-nova/frontends/origin/src/api/support.ts` —
  fetch на `${VITE_PORTAL_BASE_URL}/api/reports`.
- **Что**: S-045 Support отправляет POST на portal-backend
  /api/reports, а не на game-nova /api/* (которого не существует —
  reports переехали в portal-backend в плане 56, коммиты 37ae65b430+).
- **Почему**: единый реестр жалоб для всех вселенных (origin / nova /
  будущих) централизован на портале. portal-backend сам управляет
  дедупликацией (ключ {target_type, target_id, user_id, reason}),
  поэтому R9 Idempotency-Key не передаём — это разъяснение, не
  упрощение.
- **Trade-off**: НЕТ. Запись для документирования факта (план 56
  закрыт раньше — но S-045 это первый origin-экран, который реально
  использует portal-backend, отсюда фиксация в simplifications).
- **Приоритет**: closed.

### [P72.S4.OFFICER_NO_BALANCE_FETCH] Officer-экран не показывает текущий баланс кредитов
- **Где**: `projects/game-nova/frontends/origin/src/features/officer/
  OfficerScreen.tsx`.
- **Что**: legacy officer.tpl показывает кнопку «Нанять» и стоимость,
  но не текущий баланс (баланс — в шапке legacy). nova-фронт
  OfficersScreen отдельно фетчит /api/artefact-market/credit для
  показа баланса. В origin-фронте баланс кредитов в Spring 4 ч.2
  не рендерим — он будет доступен в шапке (header) когда финализируем
  AppShell-header в Ф.9.
- **Почему**: pixel-perfect зеркало legacy не требует баланс прямо в
  таблице офицеров. Дополнительный fetch + дублирующий UI-блок —
  отклонение от legacy без явной выгоды.
- **Trade-off**: minor — игрок видит «Активировать (1000 cr)» но не
  знает свой баланс прямо сейчас. Workaround — посмотреть в шапке
  (когда она будет полная) или в /settings.
- **Как чинить**: после финализации header'а в Ф.9 либо добавить
  `<span className="balance">` в officer-table thead, если потребуется.
- **Приоритет**: L.


### [P72.1.1.NO_BATTLE_REPORT_FOR_RAIDS] Alien-рейды и экспедиционные бои не пишут battle_reports

- **Где**: `internal/alien/alien.go`, `internal/fleet/expedition.go`.
- **Что**: `battlestats.ApplyBattleResult(ctx, tx, report, "")`
  вызывается с пустым `battleID` — у этих сценариев нет записи в
  `battle_reports` (alien использует отдельный flow через
  `messages`; экспедиции — через `expedition_reports`). Это
  значит `user_experience.battle_id = NULL`, и UNIQUE
  (battle_id, user_id, is_atter) НЕ блокирует дубликат при
  повторе event'а (NULL != NULL в SQL).
- **Почему**: legacy `oxsar2-java` хранит ВСЕ бои в одной таблице
  `assault` с инкрементальным id; у нас раньше бои разделены по
  поведению на `battle_reports` (PvP) и `messages`-only (рейды),
  что требует более сложной унификации схемы. Чтобы не утроить
  scope подплана 72.1.1, оставляем `battle_id=NULL` для рейдов и
  ловим idempotency event-loop'ом.
- **Trade-off**: minor. Если event переиграется (что крайне редко
  при exactly-once-семантике event-loop'а), defender получит
  опыт и потери дважды. Нагрузка минимальная (event re-process —
  edge-case).
- **Как чинить**: унифицировать `battle_reports` чтобы туда писались
  и рейды/экспедиции, передавать `reportID` в ApplyBattleResult.
  Это отдельный план (потенциально 72.1.2 или 73).
- **Приоритет**: L.

### [P72.1.5.B.DELETION_NO_GRACE] Soft-delete аккаунта вместо grace 7 дней

- **Где**: `internal/settings/delete.go::performDeletion`,
  `frontends/origin/src/features/settings/SettingsScreen.tsx`.
- **Что**: legacy `Preferences.class.php::updateDeletion` ставит флаг
  `users.delete = time() + 604800` (7 дней) — аккаунт помечается на
  удаление, физическое удаление выполняет cron через неделю. Юзер
  может **отменить** удаление до истечения срока (`update_deletion`
  с `delete=0`).
- **В origin**: при подтверждении email-кодом немедленно
  выполняется soft-delete: `UPDATE users SET deleted_at = now(),
  username = '[deleted_<id8>]', email = '[deleted_<id8>]'`. Отмена
  невозможна.
- **Почему упрощено**: email-код уже даёт 24-часовое окно «всё
  обдумать», и отдельный grace 7 дней с cron-задачей увеличивает
  surface для багов (потерянные cron-events, race conditions при
  попытке логина в grace-period). Бизнес-семантика «передумать —
  напиши в саппорт» приемлема для MVP.
- **Trade-off**: minor. Юзер не может self-revert удаление без
  саппорт-обращения (требует ручной правки `deleted_at = NULL` и
  восстановления username/email из бэкапа).
- **Как чинить**: добавить колонку `users.delete_at timestamptz`
  (когда физически удалить), миграция cron-job через 7 дней; в
  ConfirmDeletion вместо немедленного soft-delete ставить
  `delete_at = now() + 7 days`; UI добавить кнопку «Отменить удаление»
  если `delete_at != null && delete_at > now()`.
- **Приоритет**: L (post-MVP).

### [P72.1.5.C.RESEND_ACTIVATION_PORTAL] Повторная отправка email активации делегирована порталу

- **Где**: `projects/identity/` (portal/identity-сервис).
- **Что**: legacy `Preferences.class.php::resendActivationMail`
  отправляет повторное письмо активации в случае непод. юзера.
- **В origin**: вход на game-фронт идёт через handoff с identity-
  сервиса, и ситуация «юзер с непод. email на game-странице» не
  возникает в нормальном flow (identity сначала верифицирует email,
  потом разрешает выписать game-токен).
- **Почему делегировано**: те же доводы, что и в Support/UA
  (см. §20.12 задача 1) — TOS, account activation, password recovery,
  email confirmation — это компетенция identity-сервиса/портала,
  не игрового сервера. Дублирование UI на game-фронте создаёт
  путаницу и повышает риск SSO-рассинхронизации.
- **Trade-off**: none. На portal'е соответствующая страница
  «resend activation» уже планируется отдельно.
- **Как чинить**: не нужно. Если позже понадобится cross-link с
  game-фронта — это будет ссылка `<a href="{PORTAL_URL}/activate-resend">`,
  без логики на стороне game.
- **Приоритет**: closed.

### [P72.1.8.C.PREMIUM_BAN_ADMIN_ONLY] Premium-лоты и Ban [CLOSED by 72.1.27]

**Закрыто планом [72.1.27](plans/72.1.27-stock-premium-ban.md)**
2026-05-01: backend PromoteLot (любому игроку, cost = max(10,
price × 0.5%), max 5 featured 2ч) + BanLot (admin-only, escrow
refund, KindExchangeExpire cancel); миграция 0088 (featured_at +
banned_at + 'banned' status); UI 2 кнопки в StockScreen (⭐
Premium + 🚫 Ban для admin); AutoMsg creditExchangePremium folder=8.

- **Где**: legacy `Stock.class.php::premiumLot, ban`.
- **Что**: legacy биржа имеет два admin-action'а:
  - `premiumLot` — выставить лот как премиум (видимость наверху,
    специальная пометка). Доступ — moderator+.
  - `ban` — admin может забанить лот (скрыть из списка, расследовать
    мошенничество).
- **В origin**: оба отсутствуют.
- **Почему**: это admin-инструменты, а не игровые функции. На старте
  oxsar-nova модерация биржи — ручная (через прямые SQL-операции),
  для штатного процесса admin-UI запланирован в `projects/admin/`
  отдельно.
- **Trade-off**: minor. Игроки не видят разницы между обычным и
  премиум-лотом (в MVP — обычные сортируются по дате). Бан-функция
  — для модератора, не для игрока.
- **Как чинить**: добавить admin-роуты (`POST /api/admin/exchange/lots/{id}/ban`,
  `POST /api/admin/exchange/lots/{id}/premium`) и UI-кнопки в
  admin-фронте (не origin). Это отдельный план в admin-консоли.
- **Приоритет**: L (post-MVP).

### [P88.1] Технические заглушки `expeditionDelayNoReturnEvent` / `expeditionFastNoReturnEvent` попадают в UI

- **Где**: [projects/game-nova/configs/i18n/ru.yml:2676-2677](../projects/game-nova/configs/i18n/ru.yml#L2676-L2677), используются в [projects/game-nova/backend/internal/fleet/expedition.go:724,739](../projects/game-nova/backend/internal/fleet/expedition.go#L724).
- **Что**: при попадании в guard-ветку (отсутствует `return_event_id`)
  пользователь получает `message: "delay: нет return_event_id"` или
  `"fast: нет return_event_id"`. Это код-стиль строка, а не текст для
  игрока.
- **Почему**: guard защищает от теоретически невозможной ситуации
  (event-loop не проставил `return_event_id` при возврате).
  В прод-сценарии не должна срабатывать; если сработает — это баг
  оркестрации событий, а не штатное состояние.
- **Trade-off**: minor. В норме игрок этого не увидит. Если случай
  единичный, маркер «delay/fast: нет return_event_id» помогает
  девелоперу быстро локализовать проблему по логам/скриншоту.
- **Как чинить (в плане 88 не сделано осознанно)**: либо переписать
  как пользовательский текст («Не удалось рассчитать обратный путь.
  Обратитесь в поддержку»), либо превратить во внутреннюю ошибку
  (логировать через `slog.Error` и отдавать общий «не удалось»). Это
  не текстовая, а архитектурная правка — выходит за рамки i18n-pass'а.
- **Приоритет**: L. Затрагивает edge-case, который не должен срабатывать.


## 2026-05-01 — P72.1.13.FRIENDS_UNIDIRECTIONAL: упрощение Friends [CLOSED by 72.1.14]

**Закрыто планом [72.1.14](plans/72.1.14-friends-accept-flow.md)**
2026-05-01: миграция 0086 добавила колонку `accepted` и backfill
симметричных пар; handler реализует accept-flow с mutual
auto-accept; UI показывает 3 секции (mutual / incoming / outgoing).
AutoMsg НЕ реализован — 1:1 с legacy (TODO-комментарии в
`Friends.class.php` фактически не отправляли уведомления).

- **Где**: [`backend/internal/friends/handler.go`](../projects/game-nova/backend/internal/friends/handler.go),
  таблица `friends` (миграция 0053, расширена 0086).
- **Симптом**: legacy [`Friends.class.php`](../projects/game-legacy-php/src/game/page/Friends.class.php)
  реализует двустороннюю дружбу с подтверждением: `buddylist.accepted: 0/1`.
  Запрос на добавление — `accepted=0`, принятие — `UPDATE accepted=1`
  через `setPostAction("accept", "acceptRequest")`. Пока не принято —
  friendship «pending», обе стороны видят разные состояния.
- **Что упрощено** (в плане 11 шаг 5, до §20.12 ревизии): таблица
  `friends (user_id, friend_id, created_at)` без `accepted` flag;
  handler.go явно: «односторонний: добавление не требует
  подтверждения». Игрок A добавил B → A видит B в friends,
  B о факте не знает.
- **Почему `simplification`, а не `design-decision`**: план 72.1
  §20.12 перекрывает старые design-решения и требует строгий
  функциональный паритет с legacy-PHP. Цитата: «Любое отступление
  от legacy (в том числе "pre-existing simplification" из старых
  планов) фиксируется как баг и закрывается, либо явно эскалируется
  отдельным ADR». В перечне запрещённого: «Кнопка/действие в legacy
  работает, в origin не реализовано» (accept), «Поле/колонка в
  legacy есть, в game-nova/origin отсутствует» (`accepted`).
- **План возврата**: миграция `0086_friends_pending.sql` —
  `ALTER TABLE friends ADD COLUMN accepted boolean NOT NULL DEFAULT false;
   CREATE INDEX friends_pending_idx ON friends (friend_id) WHERE NOT accepted;`
  + endpoint `POST /api/friends/{id}/accept` + UI секция «Входящие
  запросы» в `FriendsScreen` + AutoMsg уведомление получателю при
  Add (folder=8 кредиты или новая папка). Legacy при удалении
  записывает MSG (см. TODO в коде Friends.class.php) — также
  воспроизвести.
- **Приоритет**: P3 (Social, не блокирует core gameplay, но §20.12
  требует закрыть либо эскалировать ADR).

### [P88.2] Push-уведомления офицеров подставляют английский id вместо имени

- **Где**: [projects/game-nova/configs/i18n/ru.yml:2725-2729](../projects/game-nova/configs/i18n/ru.yml#L2725-L2729),
  использование в [projects/game-nova/backend/internal/officer/service.go:391,405](../projects/game-nova/backend/internal/officer/service.go#L391).
- **Что**: тексты `officer.renewedSubject` / `officer.expiredSubject`
  начинаются со слова `Officer` (английщина) и подставляют `{{key}}` —
  это `OfficerKey` (`commander`, `engineer`, `geologist` и т.п.). На
  выходе игрок видит, например, «Officer engineer продлён
  автоматически». Английщина + непереведённый id.
- **Почему не правится в плане 88**: правка только текста ничего не
  решит — нужен либо резолв `OfficerKey` → русское имя на бэкенде
  (новый словарь имён в i18n + помощник `tr("officerName", key)`),
  либо переформулировка ключей без подстановки имени («Подписка на
  офицера продлена», «Срок подписки на офицера истёк»). Это
  UX-решение, а не текстовое.
- **Trade-off**: minor. Push-уведомление выглядит коряво, но
  работоспособно (игрок понимает суть).
- **Как чинить**: либо (a) добавить блок `officerName: { commander:
  "Командор", engineer: "Инженер", ... }` в i18n + резолвить в
  `service.go` перед `tr`; либо (b) убрать `{{key}}` из текста.
- **Приоритет**: M. Видимый игроку дефект, простая правка после
  принятия UX-решения.
