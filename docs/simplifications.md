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

### [M4.4a] Нет unit-тестов AttackHandler
- **Где**: `fleet/attack.go`.
- **Что**: 500 строк handler'а, только e2e smoke через docker.
- **Почему**: mock `pgx.Tx` + `battle.Calculate` — большая работа.
- **Как чинить**: testcontainers или interface-инъекция с моками.
- **Приоритет**: M — когда начнём рефакторить attack под ACS.

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

### [Rockets] Урон размазан по всей defense без приоритета
- **Где**: `internal/rocket/events.go::ImpactHandler`.
- **Что**: нет `target_unit_id` (конкретная цель), урон распределяется
  по всем defense пропорционально `count × shell`.
- **Почему**: простая модель.
- **Как чинить**: опциональный `target` в launch-payload + приоритет
  этому stack'у.
- **Приоритет**: L.

### [Rockets] Нет silo-limit
- **Где**: `internal/rocket/service.go::Launch`.
- **Что**: в legacy `max_rockets = silo.level × 10`; у нас этого
  ограничения нет (limit только при производстве в shipyard).
- **Почему**: нет здания `missile_silo` в каталоге.
- **Как чинить**: добавить silo в buildings.yml + проверку.
- **Приоритет**: L.

---

## Repair

### [Repair] Нельзя чинить defense
- **Где**: `internal/repair/service.go::EnqueueRepair`.
- **Что**: `isDefense → ErrUnknownUnit`. Только ship можно чинить.
- **Почему**: legacy defense-table не имеет `damaged_count`.
- **Как чинить**: добавить колонки в `defense` и применять repair
  симметрично ships.
- **Приоритет**: L — бой defense'е наносит 0 damaged (ракеты их
  уничтожают целиком).

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

### [Officers] Стеккаются с артефактами без suppression
- **Где**: `internal/officer/service.go::applyFactor`.
- **Что**: если активен артефакт +0.1 к produce_factor и officer
  GEOLOGIST +0.1, итоговая сумма = baseline + 0.2. Legacy может
  иметь «mutually exclusive» группы, у нас — нет.
- **Почему**: простая модель; пересечения редкие (артефакт как
  правило short-lived).
- **Как чинить**: добавить колонку `group` в officer_defs/
  artefact_defs и проверять «уже есть active в этой group'е».
- **Приоритет**: L.

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

### [AutoMsg] Нет CMS / редактирования через UI
- **Где**: шаблоны только в миграции `0016_automsg.sql` seed.
- **Что**: админ не может редактировать тексты через интерфейс.
  Любая правка = новая миграция.
- **Почему**: админ-панели нет (M8).
- **Как чинить**: в рамках Admin panel дать CRUD на automsg_defs.
- **Приоритет**: L — пока только dev правит.

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
- **Что**: ~~отношения NAP/WAR/ALLY~~ — ЗАКРЫТО (migration 0028, SetRelation/GetRelations,
  RelationsPanel в UI). Остаток: кастомные ранги, отношения NAP/WAR/ALLY (встречный
  acknowledge), WebSocket chat (global/alliance/PM), ACS-атаки.
- **Как чинить**: добавить WebSocket Hub с fan-out в `chat_messages`, встречное
  подтверждение отношений, ранги участников.
- **Приоритет**: M.

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

### [Tutorial] Тексты шагов хардкод на русском
- **Где**: `internal/tutorial/service.go::steps`.
- **Что**: title/description шагов — инлайн на русском, не через i18n-ключи.
- **Почему**: 6 строк текста, i18n-интеграция добавит ~30 строк boilerplate.
- **Как чинить**: вынести в `configs/i18n/ru.yml` группу tutorial.steps.*.
- **Приоритет**: L.

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

## Закрытые

- **M4.4a.rapidfire** → исправлено в iteration 20 (commit c7ae59a).
- **M4.4a.debris-in-loot** → исправлено в iteration 20 (commit 618cd26).
- **M4.4a.ui-missions** → исправлено в iteration 20.5 (commit 5336f06).
- **M4.4c** REPAIR-режим → iteration 19 (commit 42e4c89).
- **Starter-planet без buildings** → iteration 12 (commit 15be227).
- **planet evalProd не обрабатывал resource='energy'** → iteration 13.
