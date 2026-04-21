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

### [M4.1] Щиты — линейная абсорбция вместо Java-алгоритма
- **Где**: `backend/internal/battle/engine.go::applyShots`.
- **Что**: `pool = attack × shots`, тратится сначала на `turnShield`,
  остаток в `turnShell`. Java-алгоритм сложнее: каждый выстрел меньше
  `unit.shield / 100` полностью поглощается без просадки щита
  (`ignoreAttack`), есть `shieldDestroyFactor` с плавным падением.
- **Почему**: линейная модель проще тестировать, достаточно для
  сбалансированных сценариев. Порт полного Java-алгоритма — M4.3+.
- **Как чинить**: порт `Units.processAttack` shield-блока (строки
  362–427 в oxsar2-java).
- **Приоритет**: M.

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

### [M4.4a] Нет RES_LOG entries от боя (loot)
- **Где**: `fleet/attack.go::finalizeAttack`.
- **Что**: loot пишется напрямую в `fleets.carried_*`, без записи
  в `res_log` с reason='attack_loot'.
- **Почему**: `res_log.reason` enum не содержал 'attack_loot',
  ради одной строки не хотелось расширять.
- **Как чинить**: добавить reason + INSERT res_log.
- **Приоритет**: L — только для аудита.

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

### [RECYCLING] Debris один на координаты, без is_moon
- **Где**: `migrations/0010_debris_fields.sql`.
- **Что**: PK (galaxy, system, position), is_moon не учитывается —
  debris над планетой и над луной считаются одним полем.
- **Почему**: упрощение схемы. OGame позволяет обе орбиты.
- **Как чинить**: добавить колонку `is_moon` в PK.
- **Приоритет**: L.

### [SPY] Без counter-espionage и research-уровней (ratio>=8)
- **Где**: `fleet/spy.go::buildEspionageReport`.
- **Что**: нет сбития probes defense'ом, нет видимости
  research-уровней при ratio>=8 (legacy).
- **Почему**: сократили scope.
- **Как чинить**: counter-espionage = случайный roll
  `min(defense_count * 0.1, probes)`, research читается из research
  по списку unit_id.
- **Приоритет**: M.

### [COLONIZE] Нет выбора имени / размера планеты по позиции
- **Где**: `fleet/colonize.go`.
- **Что**: имя hardcoded «Colony», diameter 12800..14800 без учёта
  position (в OGame ближе к звезде меньше).
- **Почему**: MVP-подход.
- **Как чинить**: добавить field `name` в TransportInput + табличку
  «position → diameter-range».
- **Приоритет**: L.

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

### [Market] Фиксированные курсы 1:2:4
- **Где**: `internal/market/service.go::resourceCost`.
- **Что**: metal=1, silicon=2, hydrogen=4 — константы.
- **Почему**: legacy OGame так устроен.
- **Как чинить**: заменить на order-book (полноценный Exchange из
  legacy 1205 LOC).
- **Приоритет**: M — это M6 full-exchange.

### [Market] Только в рамках одной планеты
- **Где**: `internal/market/service.go::Exchange`.
- **Что**: обмен происходит на конкретной планете, нет межпланетного
  swap (ресурсы надо везти).
- **Почему**: так проще, OGame тоже так делает.
- **Как чинить**: не нужно.
- **Приоритет**: —

### [ArtefactMarket] Фильтр «мои» — по username, не по user_id
- **Где**: `features/artmarket/ArtefactMarketScreen.tsx::shown`.
- **Что**: `filter === 'mine'` сравнивает `seller_name` с
  `credit?.toString()` — это мусор.
- **Почему**: на клиенте не знаем свой user_id без отдельного запроса
  `/api/me`.
- **Как чинить**: добавить `/api/me` → `{user_id, username}` или
  парсить JWT-claims на клиенте.
- **Приоритет**: M — сейчас «мои» работает неправильно.

### [ArtefactMarket] Цена через window.prompt
- **Где**: `features/artefacts/ArtefactsScreen.tsx`.
- **Что**: ввод цены через `window.prompt`, без inline-form.
- **Почему**: минимальный UX для MVP.
- **Как чинить**: inline-form в отдельной строке при клике «Продать».
- **Приоритет**: L.

---

## Rockets

### [Rockets] Нет anti-ballistic missile (перехвата)
- **Где**: `internal/rocket/events.go::ImpactHandler`.
- **Что**: все ракеты долетают до цели, счёт без anti-ballistic.
- **Почему**: легко добавить позже, нужен второй unit_id.
- **Как чинить**: в ImpactHandler до применения урона вычитать
  `min(abm_count, interceptors)` из `pl.Count`.
- **Приоритет**: M.

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

### [Repair] Стоимость считается в момент enqueue
- **Где**: `internal/repair/service.go::EnqueueRepair`.
- **Что**: если между enqueue и finish юнит дополнительно повредился
  от второго боя, стоимость не пересчитывается.
- **Почему**: per-unit-seconds очень короткий (1–2 сек), вероятность
  edge-case минимальна.
- **Как чинить**: FOR UPDATE на ships в handler'е + перерасчёт.
- **Приоритет**: L.

---

## Messages

### [Messages] Read-only inbox (нет compose/reply/delete) — ЗАКРЫТО
- **Статус**: Закрыто в итерации 38. Реализованы POST /api/messages
  (compose), DELETE /api/messages/{id}, UI composer с ComposeForm.

### [Messages] Нет reply и нет soft-delete
- **Где**: `internal/message/service.go`, `MessagesScreen.tsx`.
- **Что**: кнопка Reply не реализована (нет pre-fill to/subject из
  оригинального сообщения). Delete — hard DELETE из БД (нет корзины,
  нет sent-folder).
- **Почему**: reply = косметика MVP; soft-delete усложняет schema.
- **Как чинить**: reply — pre-fill ComposeForm из message.from_username
  и «Re: subject». Soft-delete — добавить `deleted_at` column + WHERE
  deleted_at IS NULL в Inbox.
- **Приоритет**: L.

### [Messages] Username в BattleReport/Espionage только UUID
- **Где**: `internal/message/service.go::GetBattleReport`.
- **Что**: `attacker_user_id` в ответе — UUID, без join с users
  для username.
- **Почему**: simple query.
- **Как чинить**: LEFT JOIN users в Get*Report.
- **Приоритет**: L.

### [Messages] Folders не используются в UI
- **Где**: `features/messages/MessagesScreen.tsx`.
- **Что**: все messages в одном inbox, без фильтров по folder (2=battle,
  4=spy, etc).
- **Почему**: для MVP fields достаточно одного списка.
- **Как чинить**: tab'ы по folder.
- **Приоритет**: L.

---

## Achievements

### [Achievements] Lazy-check при GET, не real-time
- **Где**: `internal/achievement/handler.go`.
- **Что**: CheckAll прогоняется при GET /api/achievements.
  Уведомление о разблокировке приходит при следующем заходе на
  экран.
- **Почему**: не инвазивно для handler'ов (не нужно править build/
  fleet/artefact handler'ы).
- **Как чинить**: inject achievement.Service в каждый domain handler
  и вызывать UnlockIfNew после успешного действия в той же транзакции.
- **Приоритет**: M — если захотим real-time toast «Достижение!».

### [Achievements] Только 5 штук
- **Где**: `migrations/0014_achievements.sql`.
- **Что**: seed 5 достижений (FIRST_METAL/SILICON/ARTEFACT/WIN/COLONY).
  Legacy na_phrases содержит 20+ ключей ACHIEVEMENT_*.
- **Почему**: MVP-покрытие.
- **Как чинить**: расширить seed + добавить SQL-checks в CheckAll.
- **Приоритет**: L.

### [Achievements] Нет прогресс-баров (N/M)
- **Где**: `features/achievements/AchievementsScreen.tsx`.
- **Что**: только boolean unlocked/locked. Нет «построил 4 из 10
  metal_mine».
- **Почему**: boolean проще; прогресс требует `progress_json`.
- **Как чинить**: добавить колонку progress + отдельный `Progress()`
  method + прогресс-бар в UI.
- **Приоритет**: L.

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

### [Officers] Нет auto-renew
- **Где**: `officer/service.go::Activate`.
- **Что**: после expire игрок должен вручную активировать снова.
  В legacy была подписка с auto-renew за credit.
- **Почему**: UX проще; избегаем «случайно списало credit».
- **Как чинить**: флаг `auto_renew` в officer_active + новый kind
  события «попытка продлить».
- **Приоритет**: L.

### [Officers] ADMIRAL — build_factor, не attack
- **Где**: seed в `migrations/0015_officers.sql`.
- **Что**: описание говорит «+10% attack», а эффект — на
  `build_factor` (нет поля attack_factor в модели).
- **Почему**: в БД у нас только 6 фактор-полей, attack_factor не
  заведён. Legacy применяет attack-бонус прямо в бою (Participant
  getAttack()).
- **Как чинить**: добавить `users.attack_factor` + учёт в
  `fleet/attack.go::stacksToBattleUnits` (умножать attack).
  Или переписать описание на «+10% build».
- **Приоритет**: M — сейчас название вводит в заблуждение.

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

### [AutoMsg] Только event-driven, нет scheduled messages
- **Где**: отсутствует soaking/scheduler.
- **Что**: шлются только WELCOME/STARTER_GUIDE при регистрации.
  Нет inactivity-reminder, weekly-digest, event-before-raid (за N
  минут до прибытия вражеского флота).
- **Почему**: scheduled автосообщения требуют отдельного cron/worker
  и трекинга last_seen_at.
- **Как чинить**: добавить `users.last_seen_at` + воркер, который
  раз в день ищет неактивных и шлёт reminder.
- **Приоритет**: M — важно для retention.

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

### [planet.tick] Нет production.yml с формулами в runtime
- **Где**: `backend/internal/planet/service.go`.
- **Что**: есть DSL (`productionRatesDSL` через formula-пакет), но
  `construction.yml` не генерируется import-datasheets при старте
  docker → в проде всегда fallback на `productionRatesApprox`.
- **Почему**: import-datasheets запускается руками, не в dev-up.
- **Как чинить**: добавить в `deploy/Dockerfile.migrate` или отдельный
  init-контейнер вызов `cmd/tools/import-datasheets`.
- **Приоритет**: M.

### [economy] Storage cap hardcoded 1e18
- **Где**: `backend/internal/economy/formulas.go`.
- **Что**: `StorageCapacity` игнорирует `metal_storage/silicon_storage/
  hydrogen_storage` уровни → фактического капа нет.
- **Почему**: ресурсы float64, overflow не страшен.
- **Как чинить**: читать capacity из buildings + apply cap.
- **Приоритет**: L.

---

## Infrastructure

### [docker] Frontend dev-mode через bind-mount
- **Где**: `deploy/docker-compose.yml::frontend`.
- **Что**: frontend/src монтируется в контейнер, prod-сборки нет.
- **Почему**: dev-first для итераций. Для публикации нужен nginx +
  собранный bundle.
- **Как чинить**: `deploy/Dockerfile.frontend-prod` с `npm run build`
  и `nginx:alpine`.
- **Приоритет**: M — когда выйдем на внешнее demo.

### [docker] Никакого auth-rate-limiting
- **Где**: `backend/cmd/server/main.go`.
- **Что**: `/api/auth/login` без rate-limit или fail2ban.
- **Почему**: MVP не публикуется.
- **Как чинить**: middleware на redis-counter.
- **Приоритет**: H перед публикацией.

### [i18n] Только ru/en, en.yml stub
- **Где**: `configs/i18n/en.yml`.
- **Что**: en.yml — заготовка с пустыми значениями. Реально тексты
  только ru.
- **Почему**: сначала legacy-порт, переводы — потом.
- **Как чинить**: ручной перевод или translation workflow.
- **Приоритет**: M — для международного запуска.

---

## Score / Ranking (M5+)

### [score.batch] Пересчёт очков раз в 5 минут, не real-time
- **Где**: `backend/cmd/worker/main.go`, `backend/internal/score/service.go::RecalcAll`.
- **Что**: `RecalcAll` запускается горутиной с `time.Ticker(5 * time.Minute)`.
  Между завершением постройки/исследования и обновлением очков — задержка
  до 5 минут. Лидерборд не real-time.
- **Почему**: встраивать `RecalcUser` в каждый domain-handler (building,
  research, shipyard) усложняет handler'ы и добавляет N-запросов в и без
  того тяжёлые транзакции. Для лидерборда 5-минутная задержка приемлема.
- **Как чинить**: вызывать `scoreSvc.RecalcUser(ctx, userID)` в конце
  `HandleBuildConstruction`, `HandleResearch`, `HandleBuildFleet` — после
  основного UPDATE/INSERT. Нужно пробросить `*score.Service` в `event`-пакет
  или сделать callback-хук на `Worker`.
- **Приоритет**: L — для конкурентного PvP нужна точность; сейчас достаточно.

---

## Alliance (M6)

### [Alliance] MVP без заявок, рангов, отношений и WebSocket-чата
- **Где**: `internal/alliance/service.go`, `features/alliance/AllianceScreen.tsx`.
- **Что**: только create / join / leave / disband. Нет: alliance applications
  (заявки на вступление с текстом), кастомных рангов, отношений NAP/WAR/ALLY,
  WebSocket chat (global/alliance/PM), ACS-атак.
- **Почему**: полный scope M6 — ~1200 LOC PHP (alliance/ + chat/). MVP достаточен
  для игрового процесса M6-запуска.
- **Как чинить**: добавить `alliance_applications` + `alliance_relationships` таблицы,
  WebSocket Hub с fan-out в `chat_messages`, эндпоинты ACS-флота.
- **Приоритет**: M.

### [Alliance] Join без заявки — открытый альянс
- **Где**: `alliance/service.go::Join`.
- **Что**: любой игрок может войти в любой альянс напрямую без одобрения owner'а.
- **Почему**: заявки требуют отдельной таблицы + notification flow.
- **Как чинить**: добавить `is_open bool` в alliances + таблицу applications +
  approve/reject эндпоинты.
- **Приоритет**: M — для серьёзного PvP это важно.

---

## Закрытые

- **M4.4a.rapidfire** → исправлено в iteration 20 (commit c7ae59a).
- **M4.4a.debris-in-loot** → исправлено в iteration 20 (commit 618cd26).
- **M4.4a.ui-missions** → исправлено в iteration 20.5 (commit 5336f06).
- **M4.4c** REPAIR-режим → iteration 19 (commit 42e4c89).
- **Starter-planet без buildings** → iteration 12 (commit 15be227).
- **planet evalProd не обрабатывал resource='energy'** → iteration 13.
