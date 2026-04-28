# Инвентарь game-origin (Ф.1 плана 62)

**Дата сборки**: 2026-04-28
**Источники**: `projects/game-origin-php/`, БД `oxsar_db` через `docker-mysql-1`.

Это полный инвентарь PHP-проекта game-origin (clean-room порт legacy oxsar2,
план 43). Используется как базис для сравнения с game-nova и составления
журнала расхождений `divergence-log.md`.

---

## Классы `src/game/` (без подкаталогов models, page, cronjob, xml)

28 классов, всего ~28 000 строк. Перечислены с docstring-описанием
из заголовка файла.

| Класс | Строк | Назначение |
|---|---|---|
| AccountActivation | 95 | Account activation function |
| AccountCreator | 316 | Create accounts |
| AchievementsService | 1044 | Achievements system |
| AlienAI | 1127 | Class for handling the alien AI |
| AllianceList | 294 | Helper for alliance list pages |
| AllyPageParser | 405 | Parse the alliance page |
| Artefact | 1876 | Class for handling the artefacts |
| Assault | 947 | Combat handler — starts the Java-App `Assault.jar` |
| AutoMsg | 1228 | Generates auto-report after a fleet event |
| CMS | 100 | Light CMS method |
| EspionageReport | 412 | Generates espionage report |
| EventHandler | 3573 | **Сердце event-loop'а** — обрабатывает все 75 EVENT_* типов |
| Exchange | 1220 | Exchange (биржа артефактов) |
| ExpedPlanetCreator | 149 | Extension of PlanetCreator class |
| Expedition | 1160 | Expedition simulator |
| LostPassword | 152 | Sends email if password/username has been forgotten |
| MemCacheHandler | 110 | Handles caching using memcache |
| Menu | 560 | Menu class — генерирует меню по XML |
| NS | 2409 | Главный orchestrator + `isFirstRun` (memcached TTL=2s) |
| OnboardingService | 95 | Onboarding для нового юзера |
| Participant | 636 | Represents an assault participant |
| PasswordChanger | 70 | Changes the password if the key passed the check |
| Planet | 1230 | Planet class — loads planet data, updates production |
| PlanetCreator | 515 | Creates new planets and moons |
| PointRenewer | 312 | Recalculates points |
| Relation | 248 | Loads all relations of a user and his alliance |
| Uni | 123 | Represents a universe |
| UserList | 262 | Helper for user list pages |

---

## Контроллеры `src/game/page/` (55 штук)

Все контроллеры наследуют `Page.class.php`. Каждый соответствует одному
URL-экрану (через `?go=<Name>` или PATH_INFO `/<Name>`). Action'ы —
protected/public методы, используют `setPostAction`/`setGetAction`.

### Achievements (207 строк)
- **Назначение**: Displays Achievements
- **Actions**: `__construct`, `achievementGetBonus`, `achievementHideAjax`, `achievementInfo`, `achievementProcess`, `achievementsAvaliable`, `achievementsDone`, `achievementsProfile`, `achievementsRecalc`, `index`, `redirectBack`
- **Шаблоны**: `blank`

### AdvTechCalculator (58 строк)
- **Назначение**: Advanced tech calculator module
- **Actions**: `__construct`, `index`
- **Шаблоны**: `adv_tech_calc`

### Alliance (1413 строк)
- **Назначение**: Found alliances, shows alliance page and manage it
- **Actions**: `__construct`, `abandonAlly`, `acceptRelation`, `allyPage`, `allySearch`, `apply`, `applyRelationship`, `bbcode`, `cancleApplication`, `candidates`, `deleteAllyRelationApplication`, `determineRelation`, `diplomacy`, `foundAlliance`, `foundAllyForm`, `getDiploStatus`, `getRankSelect`, `getRights`, `globalMail`, `index`, `leaveAlliance`, `manageAlliance`, `manageCadidates`, `manageRanks`, `memberList`, `referFounderStatus`, `refuseRelation`, `relApplications`, `updateAllyName`, `updateAllyPrefs`, `updateAllyTag`, `writeApplication`, `writeApplicationPost`
- **Шаблоны**: `ally`, `ally_diplomacy`, `allypage_own`, `allysearch`, `applications`, `apply`, `foundally`, `globalmail`, `manage_ally`, `manage_ranks`, `memberlist`, `relation_applications`

### ArtefactInfo (261 строк)
- **Назначение**: Shows history of unique artefacts
- **Actions**: `__construct`, `activateArtefact`, `deactivateArtefact`, `index`, `showInfo`, `useArtefact`
- **Шаблоны**: `artefactinfo`

### ArtefactMarket (322 строк)
- **Назначение**: Construction & buildings page (sic — комментарий устарел; реально: рынок артефактов)
- **Actions**: `__construct`, `buyArtefact`, `index`
- **Шаблоны**: `artefactmarket2`

### ArtefactMarketOld (195 строк)
- **Назначение**: Allows to buy artefacts (legacy)
- **Actions**: `__construct`, `buyArtefactCred`, `buyArtefactRes`, `index`
- **Шаблоны**: `artefactmarket`

### Artefacts (392 строк)
- **Назначение**: Displays artefacts owned by user
- **Actions**: `__construct`, `activateArtefact`, `deactivateArtefact`, `index`, `showArtefact`
- **Шаблоны**: `artefacts`

### Battlestats (343 строк)
- **Назначение**: Shows battle statistics for users
- **Actions**: `__construct`, `index`, `showBattles`
- **Шаблоны**: `battlestats`

### BuildingInfo (330 строк)
- **Назначение**: Shows infos about a building, demolish function
- **Actions**: `__construct`, `index`, `packCurrentConstruction`, `packCurrentResearch`, `showInfo`
- **Шаблоны**: `buildinginfo`

### Changelog (66 строк)
- **Назначение**: Changelog and easter egg
- **Actions**: `__construct`, `index`
- **Шаблоны**: `changelog`

### Chat (133 строк)
- **Назначение**: Chat module
- **Actions**: `__construct`, `checkRO`, `index`, `sendMessage`
- **Шаблоны**: `chat`

### ChatAlly (146 строк)
- **Назначение**: Ally chat module
- **Actions**: `__construct`, `checkRO`, `index`, `sendMessage`
- **Шаблоны**: `chatally`

### ChatPro (33 строк)
- **Назначение**: Зарезервировано (не используется)
- **Actions**: `__construct`, `index`
- **Шаблоны**: (нет)

### Construction (498 строк)
- **Назначение**: Common helpers для проверки построек (не отдельный экран — общая база)
- **Actions**: `__construct`, `canShowAllUnits`, `checkResources`, `eventType`, `getChartData`, `setRequieredResources`, `updateUserImagePak`, `updateUserShowAllUnits`
- **Шаблоны**: (нет)

### Constructions (613 строк)
- **Назначение**: Construction & buildings page
- **Actions**: `__construct`, `abort`, `constructionInfo`, `demolish`, `index`, `upgradeConstruction`, `upgradeConstructionVIP`
- **Шаблоны**: `constructions`

### EditConstruction (353 строк)
- **Назначение**: Administrator interface to modify construction data
- **Actions**: `__construct`, `addRequirement`, `deleteRequirement`, `getResourceSelect`, `index`, `saveConstruction`
- **Шаблоны**: `edit_construction`

### EditUnit (356 строк)
- **Назначение**: Administrator interface to modify unit data
- **Actions**: `__construct`, `addRequirement`, `deleteRequirement`, `getEnginesList`, `getRapidFire`, `getShipSelect`, `index`, `saveConstruction`
- **Шаблоны**: `edit_unit`

### Empire (143 строк)
- **Назначение**: Empire module — обзор всех планет
- **Actions**: `__construct`, `index`
- **Шаблоны**: `empire`

### Exchange (154 строк)
- **Назначение**: Exchange module (новая биржа)
- **Actions**: (через index, всё в обработчиках Stock/StockNew)
- **Шаблоны**: `exchange`

### ExchangeOpts (224 строк)
- **Назначение**: Exchange admin module
- **Actions**: (через index)
- **Шаблоны**: `exchange`

### FleetAjax (187 строк)
- **Назначение**: Sends fleet via AjaxRequest (для Galaxy/Mission UI)
- **Actions**: `__construct`, `espionage`, `format`, `index`
- **Шаблоны**: (Ajax — нет шаблона)

### Friends (177 строк)
- **Назначение**: Shows and manages friend list
- **Actions**: `__construct`, `acceptRequest`, `addToBuddylist`, `index`, `removeFromList`
- **Шаблоны**: `buddylist`

### Galaxy (318 строк)
- **Назначение**: Shows galaxy
- **Actions**: `__construct`, `inMissileRange`, `index`, `setCoordinatesByGet`, `setCoordinatesByPost`, `subtractHydrogen`, `validateInputs`
- **Шаблоны**: `galaxy`

### Logout (53 строк)
- **Назначение**: Clears user cache and disables session
- **Actions**: `__construct`
- **Шаблоны**: (нет)

### MSG (489 строк)
- **Назначение**: Shows messages
- **Actions**: `__construct`, `checkSendMessage`, `createNewMessage`, `deleteAllMessages`, `deleteMessages`, `fixMessageLink`, `index`, `readFolder`, `sendMessage`
- **Шаблоны**: `folder`, `messages`, `writemessages`

### Main (631 строк)
- **Назначение**: Main page — overview and planet preferences
- **Actions**: `__construct`, `changePlanetOptions`, `homePlanetRequired`, `index`, `parseEvent`, `planetOptions`, `retreatFleet`
- **Шаблоны**: `main`, `planetOptions`

### Market (282 строк)
- **Назначение**: Market — обмен ресурсов M/Si/H/Credit
- **Actions**: `Credit_ex`, `Hydrogen_ex`, `Metal_ex`, `Silicon_ex`, `__construct`, `index`, `marketCredit`, `marketHydrogen`, `marketMetal`, `marketSilicon`
- **Шаблоны**: `market`, `market_credit`, `market_hydrogen`, `market_metal`, `market_silicon`

### Mission (3096 строк) — **самый большой контроллер**
- **Назначение**: Allows the user to send fleets to a mission
- **Actions**: `__construct`, `canAttack`, `canDestroyAttack`, `canExpedition`, `canHalt`, `canMakeStarJump`, `canRecycle`, `canSpy`, `canTeleportCurPlanet`, `canUseFleetSlot`, `controlFleet`, `controlFleetPost`, `executeJump`, `formation`, `getControlComis`, `getControlResource`, `getFormations`, `getRealShips`, `holdingSelectCoords`, `holdingSendFleet`, `index`, `invite`, `isNeutronAffectorFound`, `isNormalFleet`, `loadResourcesToFleet`, `retreatFleet`, `selectCoordinates`, `selectMission`, `sendFleet`, `setExpoAndFleetRules`, `starGateDefenseJump`, `starGateJump`, `unloadResourcesFromFleet`
- **Шаблоны**: `alliance_attack`, `mission_control`, `missions`, `missions2`, `missions3`, `missions4`, `stargatejump`

### Moderator (280 строк)
- **Назначение**: Allows moderators to change user data and manage bans
- **Actions**: `__construct`, `annulBan`, `annulRO`, `index`, `proceed`, `proceedBan`, `proceedRO`
- **Шаблоны**: `moderate_user`

### MonitorPlanet (268 строк)
- **Назначение**: Star surveillance — show planet's events
- **Actions**: `__construct`, `index`, `subtractHydrogen`, `validate`
- **Шаблоны**: `monitor_planet`

### Notepad (64 строк)
- **Назначение**: Displays Notepad
- **Actions**: `__construct`, `index`, `saveNotes`
- **Шаблоны**: `notes`

### Officer (57 строк)
- **Назначение**: Officer module — наём офицеров
- **Actions**: `__construct`, `hireOfficer`, `index`
- **Шаблоны**: `officer`

### Page (539 строк)
- **Назначение**: Abstract base class для всех контроллеров
- **Actions**: `__construct`, `addArg`, `addGetArg`, `addPostArg`, `checkForTW`, `proceedRequest`, `quit`, `resetActions`, `setGetAction`, `setPostAction`, `setPostActions`
- **Шаблоны**: (нет)

### Payment (513 строк)
- **Назначение**: Payment module (legacy — будет заменено на billing-service)
- **Actions**: `__construct`, `calcSignature`, `checkA1csv`, `checkPaymentA1Lock`, `createA1ser`, `generateCreditStringForOdnoklassniki`, `index`, `payment2pay_step1`, `payment2pay_step2`, `paymentA1`, `paymentA1step2`, `paymentRobokassa`, `paymentVkontakte`, `paymentWebmoney`
- **Шаблоны**: `blank`, `payment`, `payment2pay_step1`, `payment2pay_step2`, `paymentA1`, `paymentA1step2`, `paymentRobokassa`, `paymentWebmoney`

### Preferences (497 строк)
- **Назначение**: Allows the user to change the preferences
- **Actions**: `__construct`, `disableUmode`, `index`, `resendActivationMail`, `updateDeletion`, `updateUserData`
- **Шаблоны**: `preferences`

### Profession (131 строк)
- **Назначение**: Allows the user to change the profession
- **Actions**: `__construct`, `changeProfession`, `index`
- **Шаблоны**: `profession`

### Ranking (258 строк)
- **Назначение**: Shows ranking for users and alliances
- **Actions**: `__construct`, `allianceRanking`, `dmpointsRanking`, `epointsRanking`, `getRanking`, `index`, `maxpointsRanking`, `playerRanking`
- **Шаблоны**: `allystats`, `playerstats`

### Records (135 строк)
- **Назначение**: Shows all available constructions and their requirements (records)
- **Actions**: `__construct`, `index`
- **Шаблоны**: `records`

### Repair (688 строк)
- **Назначение**: Repair page
- **Actions**: `__construct`, `abortDefense`, `abortDisassemble`, `abortEvent`, `abortRepair`, `index`, `isRepair`, `order`, `setDisassembleUnitRequirements`, `setRepairUnitRequirements`, `startDefenseVIP`, `startDisassembleVIP`, `startEventVIP`, `startRepairVIP`
- **Шаблоны**: (использует общие)

### ResTransferStats (245 строк)
- **Назначение**: Shows resources transfer statistics
- **Actions**: `__construct`, `index`, `showResTransfer`
- **Шаблоны**: `restransfers`

### Research (522 строк)
- **Назначение**: Research page
- **Actions**: `__construct`, `abort`, `index`, `researchInfo`, `upgradeResearch`, `upgradeResearchVIP`
- **Шаблоны**: `research`

### Resource (242 строк)
- **Назначение**: Resources page
- **Actions**: `__construct`, `index`, `loadBuildingData`, `updateResources`
- **Шаблоны**: `resource`

### RocketAttack (206 строк)
- **Назначение**: Starting rocket attacks
- **Actions**: `__construct`, `index`, `sendRockets`
- **Шаблоны**: `rocket_attack`

### Search (183 строк)
- **Назначение**: Search the universe
- **Actions**: `__construct`, `allianceSearch`, `index`, `planetSearch`, `playerSearch`, `seek`
- **Шаблоны**: `ally_search_result`, `player_search_result`, `searchheader`

### Shipyard (624 строк)
- **Назначение**: Shipyard page
- **Actions**: `__construct`, `abortDefense`, `abortEvent`, `abortShipyard`, `index`, `order`, `startDefenseVIP`, `startEventVIP`, `startShipyardVIP`
- **Шаблоны**: `shipyard`

### Simulator (749 строк)
- **Назначение**: Assault simulation
- **Actions**: `__construct`, `addParticipant`, `index`, `resetAssault`, `simulate`
- **Шаблоны**: `simulator`

### Stock (757 строк)
- **Назначение**: Stock module (legacy биржа)
- **Actions**: `__construct`, `ban`, `buyLot`, `deleteExpireEvent`, `index`, `premiumLot`, `recall`, `recallByGET`, `showLotDetails`, `showLots`
- **Шаблоны**: `lot_details`, `stock`

### StockNew (850 строк)
- **Назначение**: Handles new stock (новая биржа)
- **Actions**: `__construct`, `addArtefacts`, `addLot`, `index`, `lotOptions`, `lotOptions2`, `selectFleet`, `showExchanges`
- **Шаблоны**: `empty`, `lot_details`, `missions`, `stock_new_1`, `stock_new_2`, `stock_new_3`

### Support (31 строк)
- **Назначение**: Support page
- **Actions**: `__construct`, `index`
- **Шаблоны**: `support`

### Techtree (128 строк)
- **Назначение**: Shows all available constructions and their requirements
- **Actions**: `__construct`, `index`
- **Шаблоны**: `techtree`

### TestAlienAI (32 строк)
- **Назначение**: Test page для AlienAI (admin-only)
- **Actions**: `__construct`, `index`
- **Шаблоны**: (нет)

### Tutorial (67 строк)
- **Назначение**: Displays Tutorials
- **Actions**: `__construct`, `index`
- **Шаблоны**: `tutorials`

### UnitInfo (167 строк)
- **Назначение**: Shows infos about a unit
- **Actions**: `__construct`, `index`, `showInfo`
- **Шаблоны**: `unitinfo`

### UserAgreement (115 строк)
- **Назначение**: Displays User Agreement
- **Actions**: `__construct`, `actionAgree`, `getChildAgreements`, `index`
- **Шаблоны**: `user_agreemet`

### Widgets (31 строк)
- **Назначение**: Displays Widgets
- **Actions**: `__construct`, `index`
- **Шаблоны**: `widgets`

---

## Cron-jobs (`src/game/cronjob/`)

| Класс | Строк | Назначение |
|---|---|---|
| CleanSessions | 34 | Deletes all sessions from cache folder and disable sessions in DB |
| PointClean | 80 | Cleans points |
| RemoveGalaxyGarbage | 42 | Removes destroyed planets |
| RemoveInactiveUser | 77 | Deletes inactive users |

---

## Models (`src/game/models/`)

| Класс | Строк | Назначение |
|---|---|---|
| Model | 393 | Базовый ORM-класс |
| Unit | 68 | Модель юнита |
| Items | 146 | Модель предметов |
| Structure | 48 | Модель структуры |

---

## События

### Полный список типов EVENT_* (75 уникальных)

Определены в `projects/game-origin-php/config/consts.php:402-461` как `define()`.

**Конструктивные** (строительство/исследование):
- EVENT_BUILD_CONSTRUCTION (1), EVENT_DEMOLISH_CONSTRUCTION (2),
  EVENT_RESEARCH (3), EVENT_BUILD_FLEET (4), EVENT_BUILD_DEFENSE (5),
  EVENT_REPAIR, EVENT_DISASSEMBLE, EVENT_ROCKET_ATTACK

**Флотские**:
- EVENT_POSITION, EVENT_TRANSPORT, EVENT_DELIVERY_UNITS,
  EVENT_DELIVERY_RESOURSES, EVENT_DELIVERY_ARTEFACTS,
  EVENT_COLONIZE, EVENT_COLONIZE_RANDOM_PLANET,
  EVENT_COLONIZE_NEW_USER_PLANET, EVENT_RECYCLING, EVENT_HOLDING,
  EVENT_RETURN, EVENT_HALT, EVENT_EXPEDITION, EVENT_TELEPORT_PLANET,
  EVENT_STARGATE_TRANSPORT, EVENT_STARGATE_JUMP

**Боевые**:
- EVENT_ATTACK_SINGLE, EVENT_ATTACK_DESTROY_BUILDING,
  EVENT_ATTACK_DESTROY_MOON, EVENT_ATTACK_ALLIANCE,
  EVENT_ATTACK_ALLIANCE_DESTROY_BUILDING,
  EVENT_ATTACK_ALLIANCE_DESTROY_MOON, EVENT_ALLIANCE_ATTACK_ADDITIONAL,
  EVENT_SPY, EVENT_MOON_DESTRUCTION

**Артефакты**:
- EVENT_ARTEFACT_EXPIRE, EVENT_ARTEFACT_DISAPPEAR, EVENT_ARTEFACT_DELAY

**Биржа**:
- EVENT_EXCH_EXPIRE, EVENT_EXCH_BAN

**Инопланетяне (AlienAI)**:
- EVENT_ALIEN_FLY_UNKNOWN (33), EVENT_ALIEN_HOLDING (34),
  EVENT_ALIEN_ATTACK (35), EVENT_ALIEN_HALT (36),
  EVENT_ALIEN_GRAB_CREDIT (37), EVENT_ALIEN_ATTACK_CUSTOM (38),
  EVENT_ALIEN_HOLDING_AI (80), EVENT_ALIEN_CHANGE_MISSION_AI (81)

**Турниры**:
- EVENT_TOURNAMENT_SCHEDULE, EVENT_TOURNAMENT_RESCHEDULE,
  EVENT_TOURNAMENT_PARTICIPANT (зарезервированы — обработчики в
  EventHandler пока не реализованы)

**Прочее**:
- EVENT_TEMP_PLANET_DISAPEAR, EVENT_RUN_SIM_ASSAULT

**Статусы обработки**:
- EVENT_PROCESSED_WAIT (0), EVENT_PROCESSED_START (1),
  EVENT_PROCESSED_ERROR (2), EVENT_PROCESSED_OK (3)

**Служебные тайминги**:
- EVENT_BLOCK_END_TIME, EVENT_BATCH_PROCESS_TIME (10s),
  EVENT_BATCH_CONSOLE_PROCESS_TIME (20s), EVENT_MARK_FIRST_FLEET,
  EVENT_MARK_LAST_FLEET

### Обработчики в EventHandler.class.php (3573 стр)

| EVENT_TYPE | Метод | Строки | Что делает |
|---|---|---|---|
| EVENT_BUILD_CONSTRUCTION | `build()` | 2205-2255 | Применяет уровень здания, начисляет очки |
| EVENT_DEMOLISH_CONSTRUCTION | `demolish()` | 2257-2289 | Понижает уровень здания, возвращает очки |
| EVENT_RESEARCH | `research()` | 2291-2327 | Завершает исследование, начисляет очки |
| EVENT_BUILD_FLEET / EVENT_BUILD_DEFENSE | `shipyard()` | 2329-2424 | Добавляет корабли/защиту, обновляет очки |
| EVENT_REPAIR | `repair()` | 2426-2429 | Ремонт повреждённых юнитов |
| EVENT_DISASSEMBLE | `disassemble()` | 2431-2505 | Разборка кораблей с возвратом ресурсов |
| EVENT_POSITION | `position()` | 2581-2679 | Размещение флота на координатах |
| EVENT_TELEPORT_PLANET | `teleportPlanet()` | 2507-2579 | Телепорт планеты на новые координаты |
| EVENT_TRANSPORT / DELIVERY_* | `transport()` | 2681-2767 | Доставка ресурсов/юнитов/артефактов |
| EVENT_COLONIZE / RANDOM / NEW_USER | `colonize()` | 2769-2967 | Создание новой колонии |
| EVENT_RECYCLING | `recycling()` | 2969-3003 | Переработка обломков |
| EVENT_ATTACK_SINGLE / DESTROY_* | `attack()` | 2191-2202 | Запуск Assault через Java JAR |
| EVENT_SPY | `spy()` | 3005-3051 | Создание шпионского отчёта |
| EVENT_ATTACK_ALLIANCE / DESTROY_* | `allianceAttack()` | 3053-3085 | Альянсовая атака с консолидацией флотов |
| EVENT_HALT | `halt()` | 3087-3095 | Остановка флота на полпути |
| EVENT_HOLDING | `holding()` | 3097-3118 | Удержание (consumes hydrogen) |
| EVENT_MOON_DESTRUCTION | `moonDestruction()` | 3120-3124 | Уничтожение луны |
| EVENT_EXPEDITION | `expedition()` | 3126-3251 | Случайный исход экспедиции (13 типов) |
| EVENT_ROCKET_ATTACK | `rocketAttack()` | 3133-3154 | Ракетная атака |
| EVENT_RETURN | `fReturn()` | 3156-3251 | Возврат флота на исходную планету |
| EVENT_ARTEFACT_EXPIRE | `artefactExpire()` | 3253-3257 | Истечение эффекта артефакта |
| EVENT_ARTEFACT_DISAPPEAR | `artefactDisappear()` | 3259-3263 | Артефакт исчезает |
| EVENT_ARTEFACT_DELAY | `artefactDelay()` | 3265-3270 | Задержка артефакта |
| EVENT_EXCH_EXPIRE | `exchangeExpire()` | 3272-3285 | Истечение лота на бирже |
| EVENT_EXCH_BAN | (пустой) | — | Бан на бирже (служебное) |
| EVENT_TEMP_PLANET_DISAPEAR | `destroyPlanet()` | 3287-3329 | Уничтожение временной планеты по TTL |
| EVENT_RUN_SIM_ASSAULT | `runSimAssault()` | 3331-3354 | Запуск симулятора |
| EVENT_ALIEN_FLY_UNKNOWN | `alienFlyUnknown()` | 3532-3535 | Делегирует в `AlienAI::onFlyUnknownEvent()` |
| EVENT_ALIEN_GRAB_CREDIT | `alienGrabCredit()` | 3538-3541 | → `AlienAI::onGrabCreditEvent()` |
| EVENT_ALIEN_HOLDING | `alienHolding()` | 3544-3547 | → `AlienAI::onHoldingEvent()` |
| EVENT_ALIEN_HOLDING_AI | `alienHoldingAI()` | 3550-3553 | → `AlienAI::onHoldingAIEvent()` |
| EVENT_ALIEN_CHANGE_MISSION_AI | `alienChangeMissionAI()` | 3556-3559 | → `AlienAI::onChangeMissionAIEvent()` |
| EVENT_ALIEN_ATTACK | `alienAttack()` | 3562-3565 | → `AlienAI::onAttackEvent()` |
| EVENT_ALIEN_HALT | `alienHalt()` | 3568-3571 | → `AlienAI::onHaltEvent()` |
| EVENT_TOURNAMENT_* | (заглушки) | — | Зарезервировано, не реализовано |

### Обработчики вне EventHandler

- `AlienAI.class.php` (1127 стр) — `onFlyUnknownEvent`, `onAttackEvent`,
  `onGrabCreditEvent`, `onHaltEvent`, `onHoldingEvent`,
  `onChangeMissionAIEvent`, `onHoldingAIEvent` — детально в
  `alien-ai-comparison.md`.
- `Artefact.class.php` (1876 стр) — создание EVENT_ARTEFACT_*
  событий (линии 270, 281, 521, 597, 619, 770, 894, 913, 1298, 1311).
- `Expedition.class.php:1063` — создание EVENT_EXPEDITION.
- `Exchange.class.php` — EVENT_DELIVERY_* и EVENT_EXCH_EXPIRE.

### Места создания событий (call-sites)

| EVENT_TYPE | Где создаётся | Когда |
|---|---|---|
| EVENT_BUILD_CONSTRUCTION | `page/Constructions.class.php:491` | Игрок ставит здание в очередь |
| EVENT_DEMOLISH_CONSTRUCTION | `page/Constructions.class.php:600` | Снос здания |
| EVENT_RESEARCH | `page/Research.class.php:475` | Запуск исследования |
| EVENT_BUILD_FLEET / EVENT_BUILD_DEFENSE | `page/Shipyard.class.php:560` | Заказ кораблей/защиты |
| EVENT_REPAIR | `page/Repair.class.php:606` | Запуск ремонта |
| EVENT_POSITION / TRANSPORT / COLONIZE / ATTACK / RECYCLING / HALT / SPY | `page/Mission.class.php:1682` | Отправка флота |
| EVENT_COLONIZE_NEW_USER_PLANET | `OnboardingService.class.php`, `page/Main.class.php` | Onboarding нового игрока |
| EVENT_SPY | `page/FleetAjax.class.php:159` | AJAX-запуск шпионского зонда |
| EVENT_EXPEDITION | `Expedition.class.php:1063` | Отправка экспедиции |
| EVENT_ROCKET_ATTACK | `page/RocketAttack.class.php:200` | Запуск ракеты |
| EVENT_STARGATE_JUMP | `page/Mission.class.php:2150` | Прыжок через Stargate |
| EVENT_RETURN | `EventHandler` (внутри других обработчиков) | Автомат после миссии |
| EVENT_HOLDING | `EventHandler:3091, 3463, 3519` | Прибытие флота к цели |
| EVENT_ARTEFACT_DELAY / DISAPPEAR / EXPIRE | `Artefact.class.php`, `page/ArtefactMarketOld.class.php:138,185` | Жизненный цикл артефакта |
| EVENT_EXCH_EXPIRE | `page/Stock.class.php:133`, `page/StockNew.class.php:806` | Создание лота на бирже |
| EVENT_RUN_SIM_ASSAULT | `page/Simulator.class.php:426` | Запуск симулятора |
| EVENT_ALIEN_* | `AlienAI.class.php` (170, 207, 229, 804, 831, 836, 1066) | Cron-вызов AlienAI::checkAlientNeeds |
| EVENT_TEMP_PLANET_DISAPEAR | `Expedition.class.php` | Создание временной планеты |
| EVENT_TOURNAMENT_* | (нет call-site — не реализовано) | — |

### Параметры обработки событий

- **Batch processing**:
  - `EVENT_BATCH_PROCESS_TIME = 10s` (web)
  - `EVENT_BATCH_CONSOLE_PROCESS_TIME = 20s` (cron/console)
- **Очистка старых записей** (random 1/1000):
  - `EVENT_PROCESSED_OK` → удаление через 7 дней
  - `EVENT_PROCESSED_ERROR` → удаление через 10 дней
  - `EVENT_PROCESSED_START` → удаление через 3 дня (stuck recovery)
- **Блокировка во время отпуска (umode)**: 38 типов событий не
  обрабатываются для игроков в umode.
- **Отмена и возврат ресурсов**:
  - `EV_ABORT_SAVE_TIME = 15s` (полный возврат)
  - Constructions: до 95% (`EV_ABORT_MAX_BUILD_PERCENT`)
  - Shipyard: до 70%
  - Repair/Disassemble: до 70%
  - Полёты: до 90%

---

## База данных

**216 таблиц** в `oxsar_db`. Доступ:
```bash
docker exec docker-mysql-1 mysql -uoxsar_user -poxsar_pass oxsar_db \
  -e "SHOW TABLES;"
```

### Системные таблицы
`na_user`, `na_password`, `na_alliance`, `na_planet`, `na_galaxy`,
`na_sessions`, `na_config`, `na_languages`, `na_permissions`,
`na_group2permission`, `na_page`, `na_user_online`, `na_user_states`,
`na_user2group`, `na_usergroup`, `na_user_agreement`, `na_global_user_id`
(пл. 36).

### Балансовые / постройки
`na_construction` (главная — формулы как DSL-строки),
`na_building2planet`, `na_building2planet_tmp`, `na_building2planet_tmp2`,
`na_research2user`, `na_unit2shipyard`, `na_ship_datasheet`,
`na_rapidfire`, `na_ship2engine`, `na_engine`, `na_attack_formation`,
`na_artefact_datasheet`, `na_artefact2user`, `na_artefact_history`,
`na_artefact_probobility`, `na_artefact_used`, `na_artefact2user_tmp`.

### События и миссии
`na_events`, `na_event_aliens`, `na_event_dest`, `na_event_src`,
`na_fleet2assault`.

### Боевая система
`na_assault`, `na_assault_ext`, `na_assault_ext2`, `na_assault_stat`,
`na_assaultparticipant`, `na_attack_formation`, `na_formation_invitation`,
`na_sim_*` (зеркальный набор для симулятора).

### Экспедиции
`na_expedition_found_units`, `na_expedition_log`, `na_expedition_stats`,
`na_expedition_stats_day`, `na_expedition_stats_ext`,
`na_expedition_stats_old2`, `na_expedition_stats_old3`,
`na_expedition_stats_olddata`, `na_expedition_used`.

### Чат / сообщения
`na_chat`, `na_chat2ally`, `na_chat2ally_ext`, `na_chat2ally_stat`,
`na_chat_ext`, `na_chat_ro`, `na_chat_ro_u`, `na_chat_tmp`,
`na_message`, `na_message_ext`, `na_folder`, `na_sendreport`.

### Биржа
`na_exchange`, `na_exchange_lots`, `na_exchange_stats`,
`na_exchange_tmp`, `na_u2exchange`, `na_user2exchange`.

### Альянсы / отношения
`na_alliance`, `na_alliance_tmp`, `na_allyrank`, `na_allyapplication`,
`na_ally_relationships`, `na_ally_relationships_application`,
`na_user2ally`.

### Ресурсы и логи
`na_res_log` + множество `na_res_log_*` (game_credit, gift_stats,
grab_stats, hack, hack_dub, premium_stats, stats, stats_month,
type_stats), `na_res_transfer`.

### Достижения / турниры
`na_achievement_datasheet`, `na_achievements2user`, `na_tournament`.

### Бан / модерация
`na_ban`, `na_ban_u`, `na_ban_u_ext`, `na_chat_ro`, `na_chat_ro_u`.

### Платежи / премиум
`na_payments`, `na_payments_ext`, `na_payment_stats`,
`na_payment_stats_month`, `na_payment_user_stats`,
`na_payment_user_stats_month`, `na_credit_bonus_item`, `na_officer`,
`na_officer_ext`.

### Прочее
`na_buddylist`, `na_notes`, `na_registration`, `na_requirements`,
`na_free_planets`, `na_moon_creation_stats`, `na_moon_destroy_stats`,
`na_moon_destroy_stats_old`, `na_moon_destroy_stats_old2`,
`na_ships_log`, `na_units_destroyed_stats`, `na_tracks`, `na_tutorials`,
`na_tutorial_states`, `na_tutorial_states_category`,
`na_user_reg_stats`, `na_user_reg_stats_month`,
`na_user_imgpak_ext`, `na_user_ext`, `na_user_experience`,
`na_user_copy`, `na_user_tmp`, `na_yii_cron`, `na_yii_log`,
`na_yii_log_info`, `na_phrases`, `na_phrases_tmp`, `na_phrasesgroups`,
`na_log`, `na_loginattempts`, `na_cronjob`, `na_cronjob_ext`,
`na_stargate_jump`, `na_social_network_user`, `na_sim_*` (симулятор).

---

## Миграции `migrations/` (5 файлов)

| Файл | Назначение |
|---|---|
| `001_schema.sql` | Создание всех 216 таблиц БД (DDL) |
| `002_data.sql` | Seed-данные (mysqldump из live oxsar2) |
| `003_add_global_user_id.sql` | Добавление `na_user.global_user_id` (varbinary(36)) для связки с identity-service (план 36) |
| `004_username_forbidden_phrase.sql` | Фраза-маркер для регистрации запрещённого username (149-ФЗ, план 46/48) |
| `005_drop_referral.sql` | Удаление legacy реферальной системы (план 60 — реферальная программа теперь на portal-backend) |

---

## Runtime-генерируемые ассеты

Найден через `grep -rln "imagecreate\|imagepng\|imagejpeg" projects/game-origin-php/`:

- **`public/artefact-image.php`** (153 строки) — PHP-GD endpoint,
  отдающий композитную PNG артефакта: фон постройки/исследования +
  иконка типа артефакта + уровень. Кеш в `$CACHE_DIR/{name}{level}{art_name}.jpg`.
  URL: `/artefact-image.php?cid=N&level=N&typeid=N`.
  Используется в шаблонах артефактов через `Artefact.class.php` →
  `setViewParams`, поле `artefact["image"]`.

Других `imagecreate*`/`imagettftext` PHP-генераторов не найдено.

---

## Ключевые файлы конфигурации

- `config/consts.php` — все define-константы (>500 штук, в т.ч.
  EVENT_*, UNIT_*, ALIEN_*, MSG_*, RES_*, EXCH_*, MOON_*, FLEET_*).
- `config/params.php` — параметры стартовых ресурсов, лимиты,
  ссылки.
- `src/bd_connect_info.php` — креды БД.
- `src/global.inc.php` — bootstrap.
- `src/common.inc.php` — общие хелперы.

---

## Шаблоны (`src/templates/standard/`) — 125 .tpl

Полный визуальный инвентарь — отдельный артефакт
`origin-ui-replication.md` (Ф.3).
