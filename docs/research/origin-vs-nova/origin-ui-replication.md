# Origin UI Replication Inventory (S-NNN)

**Дата сборки**: 2026-04-28
**Контекст**: артефакт плана 62 — репродукционный инвентарь
game-origin для будущего pixel-perfect-клонирования на React.
**ВСЕ 55 контроллеров** из `projects/game-legacy-php/src/game/page/`
сопоставлены с экранами S-NNN. Пропуск контроллера = баг, потому
что приведёт к пропуску экрана при разработке клона.

Цель: дать будущему агенту-реализатору **полный список того, что
нужно воспроизвести**, без необходимости проводить повторный аудит.

---

## Экраны (полный список, S-001..S-055)

### S-001. Главный экран (Main)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Main.class.php`
- **Шаблоны**: `templates/standard/main.tpl`, `planetOptions.tpl`
- **URL**: `?go=Main` (+ &action=changePlanetOptions, retreatFleet)
- **Назначение**: Обзор аккаунта, информация о текущей планете,
  активные события строительства/исследования и движения флотов.
- **Layout**: Две колонки (левое меню + основной контент)
- **Основные блоки**: текущая планета и координаты; ресурсы с
  прогресс-барами; очереди стройки с countdown; события флотов;
  быстрые ссылки.
- **Действия**: смена планеты, отступление флота
- **Backend nova-endpoint**: `GET /api/planets/{id}` + `GET /api/events`
- **Сложность воспроизведения**: средняя

### S-002. Исследования (Research)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Research.class.php`
- **Шаблоны**: `templates/standard/research.tpl`
- **URL**: `?go=Research` (+ &action=upgradeResearch, abort, upgradeResearchVIP, researchInfo)
- **Назначение**: Список технологий, очередь, запуск/отмена.
- **Layout**: Таблица с прогресс-барами
- **Основные блоки**: таблица технологий (название/уровень/стоимость),
  текущее исследование с countdown, доступные, VIP-ускорение
- **Действия**: запуск, отмена, VIP, info
- **Backend nova-endpoint**: `POST /api/planets/{id}/research`,
  `GET /api/research`
- **Сложность воспроизведения**: средняя

### S-003. Здания (Constructions)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Constructions.class.php`
- **Шаблоны**: `templates/standard/constructions.tpl`,
  `cons_chart.tpl`
- **URL**: `?go=Constructions` (+ &action=upgradeConstruction, abort,
  demolish, upgradeConstructionVIP, constructionInfo)
- **Назначение**: Список зданий + очередь.
- **Layout**: Таблица с модальными окнами деталей
- **Основные блоки**: таблица зданий, очередь с прогресс-барами,
  стоимость, VIP, демонтаж
- **Действия**: улучшение, отмена, демонтаж, VIP
- **Backend nova-endpoint**: `POST /api/planets/{id}/buildings`,
  `GET/DELETE /api/planets/{id}/buildings/queue{,/{taskId}}`
- **Сложность воспроизведения**: средняя

### S-004. Верфь (Shipyard)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Shipyard.class.php`
- **Шаблоны**: `templates/standard/shipyard.tpl`
- **URL**: `?go=Shipyard` (+ &action=order, abortShipyard, abortDefense, abortEvent, startShipyardVIP, startDefenseVIP, startEventVIP)
- **Назначение**: Постройка флота и обороны.
- **Layout**: Таблица + формы ввода количеств
- **Основные блоки**: переключатель флот/оборона, таблица юнитов
  (название/имеется/стоимость/время), поле количества, итог,
  очередь
- **Действия**: заказ, отмена, VIP
- **Backend nova-endpoint**: `POST /api/planets/{id}/shipyard`,
  `GET /api/planets/{id}/shipyard/{queue,inventory}`
- **Сложность воспроизведения**: высокая (расчёты, очереди)

### S-005. Галактика (Galaxy)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Galaxy.class.php`
- **Шаблоны**: `templates/standard/galaxy.tpl`
- **URL**: `?go=Galaxy&galaxy=X&system=Y` (+ &action=setCoordinatesByGet/Post)
- **Назначение**: Карта галактики, обзор систем, отправка флотов.
- **Layout**: Таблица 15 позиций × колонки
- **Основные блоки**: выбор g/s стрелками, таблица планет
  (координаты, владелец, альянс, метки), быстрые действия,
  поиск координат
- **Действия**: переход, sendFleet (через FleetAjax), inMissileRange,
  subtractHydrogen
- **Backend nova-endpoint**: `GET /api/galaxy/{g}/{s}`
- **Сложность воспроизведения**: высокая (AJAX, динамика)

### S-006. Миссии (Mission)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Mission.class.php` (3096 строк)
- **Шаблоны**: `templates/standard/missions.tpl`, `missions2.tpl`,
  `missions3.tpl`, `missions4.tpl`, `mission_control.tpl`,
  `alliance_attack.tpl`, `stargatejump.tpl`
- **URL**: `?go=Mission` + 30+ actions (selectCoordinates,
  selectMission, sendFleet, retreatFleet, controlFleet,
  starGateJump, formation, invite, holdingSendFleet, etc.)
- **Назначение**: Wizard отправки флотов: координаты → миссия →
  флот → подтверждение.
- **Layout**: Многошаговый wizard (4 шага)
- **Основные блоки**: формы ввода, таблицы выбора, расчёт времени
  и расхода топлива
- **Действия**: 30+ — все варианты миссий, ACS-приглашения, сtargate
- **Backend nova-endpoint**: `GET/POST /api/fleet`,
  `POST /api/fleet/{id}/recall`, `POST /api/stargate`
- **Сложность воспроизведения**: **очень высокая** — 3096 строк
  бизнес-логики

### S-007. Сообщения (MSG)

- **Контроллер**: `projects/game-legacy-php/src/game/page/MSG.class.php`
- **Шаблоны**: `templates/standard/folder.tpl`, `messages.tpl`,
  `writemessages.tpl`
- **URL**: `?go=MSG` (+ &action=sendMessage, deleteMessages,
  deleteAllMessages, createNewMessage, readFolder, fixMessageLink)
- **Назначение**: Личные сообщения.
- **Layout**: Две колонки (папки + список + просмотр)
- **Основные блоки**: папки, таблица писем, просмотр, форма ответа
- **Действия**: отправка, удаление, новое сообщение
- **Backend nova-endpoint**: `GET/POST /api/messages`,
  `DELETE /api/messages/{id}`, `POST /api/messages/{id}/read`,
  `GET /api/messages/unread-count`
- **Сложность воспроизведения**: средняя

### S-008. Чат (Chat)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Chat.class.php`
- **Шаблоны**: `templates/standard/chat.tpl`
- **URL**: `?go=Chat` (+ &action=sendMessage, checkRO)
- **Назначение**: Глобальный чат с BBCode (legacy).
- **Layout**: Колонка (список + форма)
- **Основные блоки**: скролл последних 75 сообщений (BBCode), форма
- **Действия**: sendMessage, checkRO
- **Backend nova-endpoint**: `WS /ws/chat` (план 32; BBCode →
  TipTap)
- **Сложность воспроизведения**: низкая (но BBCode выкидывается —
  заменяется TipTap)

### S-009. Чат альянса (ChatAlly)

- **Контроллер**: `projects/game-legacy-php/src/game/page/ChatAlly.class.php`
- **Шаблоны**: `templates/standard/chatally.tpl`
- **URL**: `?go=ChatAlly` (+ &action=sendMessage, checkRO)
- **Назначение**: Только для членов альянса.
- **Layout**: Колонка (список + форма)
- **Backend nova-endpoint**: `WS /ws/chat` (отдельный канал
  альянса)
- **Сложность воспроизведения**: низкая

### S-010. ChatPro (зарезервирован, не отдельный экран)

- **Контроллер**: `projects/game-legacy-php/src/game/page/ChatPro.class.php` (33 стр)
- **Назначение**: Заглушка для расширенного чата (вынесена в
  `/novax/chat/`). Не отдельный экран — пропускаемый контроллер.
- **Сложность воспроизведения**: не воспроизводится

### S-011. Друзья (Friends)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Friends.class.php`
- **Шаблоны**: `templates/standard/buddylist.tpl`
- **URL**: `?go=Friends` (+ &action=removeFromList, acceptRequest, addToBuddylist)
- **Назначение**: Список друзей + запросы.
- **Layout**: Две колонки (друзья + приглашения)
- **Backend nova-endpoint**: `GET/POST/DELETE /api/friends{,/{id}}`
- **Сложность воспроизведения**: низкая

### S-012. Альянс (Alliance) — самый сложный экран

- **Контроллер**: `projects/game-legacy-php/src/game/page/Alliance.class.php` (1413 стр)
- **Шаблоны**: `ally.tpl`, `allypage_own.tpl`, `memberlist.tpl`,
  `manage_ranks.tpl`, `globalmail.tpl`, `relation_applications.tpl`,
  `ally_diplomacy.tpl`, `allysearch.tpl`, `foundally.tpl`,
  `apply.tpl`, `manage_ally.tpl`, `applications.tpl`
- **URL**: `?go=Alliance` (+ ~30 actions: foundAlliance, apply,
  acceptRelation, candidates, abandonAlly, manageRanks, diplomacy,
  globalMail, allySearch, …)
- **Назначение**: Полный жизненный цикл альянса
- **Layout**: Множественные представления (12 шаблонов)
- **Основные блоки**: страница альянса, члены с рангами,
  дипломатия, заявки, кастомные ранги с правами, global mail,
  передача лидерства
- **Действия**: 30+ (см. инвентарь контроллеров в `origin-inventory.md`)
- **Backend nova-endpoint**: 17 эндпоинтов в nova
  (`/api/alliances/*`), но: **отсутствует** abandonAlly (передача
  лидерства), три описания, гранулярные права рангов, global mail
  (нужен mail-service плана 57)
- **Сложность воспроизведения**: **очень высокая** (12 шаблонов)

### S-013. Артефакты (Artefacts)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Artefacts.class.php`
- **Шаблоны**: `templates/standard/artefacts.tpl`
- **URL**: `?go=Artefacts` (+ &action=activateArtefact, deactivateArtefact, showArtefact)
- **Назначение**: Инвентарь артефактов.
- **Layout**: Таблица
- **Основные блоки**: иконка/название/раритет/уровень/эффект/статус
- **Действия**: активация, деактивация, просмотр
- **Backend nova-endpoint**: `GET /api/artefacts`,
  `POST /api/artefacts/{id}/{activate,deactivate,sell}`
- **Сложность воспроизведения**: средняя

### S-014. Информация об артефакте (ArtefactInfo)

- **Контроллер**: `projects/game-legacy-php/src/game/page/ArtefactInfo.class.php`
- **Шаблоны**: `artefactinfo.tpl`, `artefact_row_info.tpl`
- **URL**: `?go=ArtefactInfo` (+ &action=showInfo, useArtefact, activateArtefact, deactivateArtefact)
- **Назначение**: Модальный просмотр артефакта.
- **Сложность воспроизведения**: низкая

### S-015. Рынок артефактов (ArtefactMarket)

- **Контроллер**: `projects/game-legacy-php/src/game/page/ArtefactMarket.class.php`
- **Шаблоны**: `templates/standard/artefactmarket2.tpl`
- **URL**: `?go=ArtefactMarket` (+ &action=buyArtefact)
- **Backend nova-endpoint**: `GET/POST /api/artefact-market/offers{,/{id}/buy}`
- **Сложность воспроизведения**: средняя

### S-016. Рынок артефактов старый (ArtefactMarketOld)

- **Контроллер**: `projects/game-legacy-php/src/game/page/ArtefactMarketOld.class.php`
- **Шаблоны**: `templates/standard/artefactmarket.tpl`
- **URL**: `?go=ArtefactMarketOld` (+ &action=buyArtefactRes, buyArtefactCred)
- **Назначение**: Legacy-версия (вероятно для уни01-стиля игры
  без оффлайн-опции). В клон попадает через флаг.
- **Сложность воспроизведения**: средняя

### S-017. Боевая статистика (Battlestats)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Battlestats.class.php`
- **Шаблоны**: `templates/standard/battlestats.tpl`,
  `assault_report.tpl`, `_report_button.tpl`
- **URL**: `?go=Battlestats` (+ &action=showBattles)
- **Назначение**: История атак с боевыми отчётами.
- **Backend nova-endpoint**: `GET /api/battlestats`,
  `GET /api/battle-reports/{id}`
- **Сложность воспроизведения**: высокая (assault_report.tpl —
  сложная вёрстка)

### S-018. Информация о здании (BuildingInfo)

- **Контроллер**: `projects/game-legacy-php/src/game/page/BuildingInfo.class.php`
- **Шаблоны**: `templates/standard/buildinginfo.tpl`
- **URL**: `?go=BuildingInfo` (+ &action=showInfo, packCurrentConstruction, packCurrentResearch)
- **Backend nova-endpoint**: derived из `GET /api/research`/buildings
- **Сложность воспроизведения**: низкая

### S-019. Информация о юните (UnitInfo)

- **Контроллер**: `projects/game-legacy-php/src/game/page/UnitInfo.class.php`
- **Шаблоны**: `templates/standard/unitinfo.tpl`
- **URL**: `?go=UnitInfo` (+ &action=showInfo)
- **Backend nova-endpoint**: derived из catalog (frontend/src/api/catalog.ts)
- **Сложность воспроизведения**: низкая

### S-020. Информация об исследовании (встроено в Research)

- **Контроллер**: встроено в `Research.class.php` (researchInfo,
  packCurrentResearch)
- **Сложность**: встроено

### S-021. Tech Tree (Techtree)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Techtree.class.php`
- **Шаблоны**: `templates/standard/techtree.tpl`
- **URL**: `?go=Techtree`
- **Назначение**: Граф зависимостей.
- **Layout**: Граф (диаграмма связей)
- **Backend nova-endpoint**: `GET /api/techtree` (есть в nova)
- **Сложность воспроизведения**: высокая (граф)

### S-022. Достижения (Achievements)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Achievements.class.php`
- **Шаблоны**: `templates/standard/achievement.tpl`,
  `achievements.tpl`
- **URL**: `?go=Achievements` (+ 9 actions)
- **Backend nova-endpoint**: `GET /api/achievements`
- **Сложность воспроизведения**: средняя

### S-023. Рейтинг (Ranking)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Ranking.class.php`
- **Шаблоны**: `playerstats.tpl`, `allystats.tpl`
- **URL**: `?go=Ranking` (+ &action=playerRanking, allianceRanking, dmpointsRanking, epointsRanking, maxpointsRanking)
- **Backend nova-endpoint**: `GET /api/highscore{,/me}`
- **Сложность воспроизведения**: средняя

### S-024. Рекорды (Records)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Records.class.php`
- **Шаблоны**: `templates/standard/records.tpl`
- **URL**: `?go=Records`
- **Backend nova-endpoint**: `GET /api/records`
- **Сложность воспроизведения**: низкая

### S-025. Ресурсы (Resource)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Resource.class.php`
- **Шаблоны**: `templates/standard/resource.tpl`
- **URL**: `?go=Resource` (+ &action=updateResources, loadBuildingData)
- **Backend nova-endpoint**: `GET /api/planets/{id}/resource-report`
- **Сложность воспроизведения**: средняя

### S-026. Рынок ресурсов (Market)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Market.class.php`
- **Шаблоны**: `market.tpl`, `market_credit.tpl`,
  `market_metal.tpl`, `market_silicon.tpl`, `market_hydrogen.tpl`
- **URL**: `?go=Market` (+ &action=Metal_ex, Silicon_ex, Hydrogen_ex, Credit_ex)
- **Backend nova-endpoint**: `GET /api/market/rates`,
  `POST /api/planets/{id}/market/exchange`
- **Сложность воспроизведения**: средняя

### S-027. Опции рынка (ExchangeOpts)

- **Контроллер**: `projects/game-legacy-php/src/game/page/ExchangeOpts.class.php`
- **Шаблоны**: `exchange.tpl`
- **URL**: `?go=ExchangeOpts`
- **Backend nova-endpoint**: вероятно отсутствует — D-NNN
- **Сложность воспроизведения**: низкая

### S-028. Биржа (Exchange)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Exchange.class.php` (154 стр)
- **Шаблоны**: `exchange.tpl`
- **URL**: `?go=Exchange`
- **Назначение**: история обменов / стат.
- **Backend nova-endpoint**: ОТСУТСТВУЕТ полностью — `D-NNN-EXCHANGE`
- **Сложность воспроизведения**: средняя

### S-029. Настройки (Preferences)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Preferences.class.php`
- **Шаблоны**: `templates/standard/preferences.tpl`
- **URL**: `?go=Preferences` (+ &action=updateUserData, disableUmode, resendActivationMail, updateDeletion)
- **Backend nova-endpoint**: `GET /api/me`, `PATCH /api/me/*`
- **Сложность воспроизведения**: средняя

### S-030. Профессия (Profession)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Profession.class.php`
- **Шаблоны**: `templates/standard/profession.tpl`
- **URL**: `?go=Profession` (+ &action=changeProfession)
- **Backend nova-endpoint**: `GET /api/professions`
- **Сложность воспроизведения**: низкая

### S-031. Офицеры (Officer)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Officer.class.php`
- **Шаблоны**: `templates/standard/officer.tpl`
- **URL**: `?go=Officer` (+ &action=hireOfficer)
- **Backend nova-endpoint**: `GET /api/officers`,
  `POST /api/officers/{key}/activate`
- **Сложность воспроизведения**: низкая

### S-032. Платежи (Payment)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Payment.class.php` (513 стр)
- **Шаблоны**: 8 шаблонов (paymentRobokassa, paymentWebmoney,
  paymentVkontakte, paymentA1, ...)
- **URL**: `?go=Payment` (+ много actions для разных шлюзов)
- **Назначение**: Legacy-платежи. В origin-фронте на nova **должен
  быть полностью заменён** на billing-service (план 38, 42).
- **Backend nova-endpoint**: `GET /api/payment/{packages,history}`,
  `POST /api/payment/{order,webhook}`
- **Сложность воспроизведения**: **не воспроизводится** —
  legacy-логика выкидывается, используется единый billing-service

### S-033. User Agreement (UserAgreement)

- **Контроллер**: `projects/game-legacy-php/src/game/page/UserAgreement.class.php`
- **Шаблоны**: `front_user_areement.tpl`, `user_agreemet.tpl`
- **URL**: `?go=UserAgreement` (+ &action=actionAgree)
- **Backend nova-endpoint**: `GET /api/legal/terms` + `POST /api/me/accept-terms`
  (план 47)
- **Сложность воспроизведения**: низкая

### S-034. Changelog

- **Контроллер**: `projects/game-legacy-php/src/game/page/Changelog.class.php`
- **Шаблоны**: `templates/standard/changelog.tpl`
- **URL**: `?go=Changelog`
- **Backend nova-endpoint**: возможно через wiki или отдельный
  ресурс
- **Сложность воспроизведения**: низкая

### S-035. Tutorial

- **Контроллер**: `projects/game-legacy-php/src/game/page/Tutorial.class.php`
- **Шаблоны**: `templates/standard/tutorials.tpl`
- **URL**: `?go=Tutorial`
- **Backend nova-endpoint**: единый goal engine (`internal/goal/`)
- **Сложность воспроизведения**: высокая (highlight UI, JS-логика)

### S-036. Виджеты (Widgets)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Widgets.class.php` (31 стр)
- **Шаблоны**: `templates/standard/widgets.tpl`
- **URL**: `?go=Widgets`
- **Сложность воспроизведения**: средняя (drag-drop)

### S-037. Поддержка (Support)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Support.class.php` (31 стр)
- **Шаблоны**: `templates/standard/support.tpl`
- **URL**: `?go=Support`
- **Backend nova-endpoint**: ссылка на portal (внешняя)
- **Сложность воспроизведения**: низкая

### S-038. Поиск (Search)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Search.class.php`
- **Шаблоны**: `searchheader.tpl`, `player_search_result.tpl`,
  `ally_search_result.tpl`
- **URL**: `?go=Search` (+ &action=playerSearch, allianceSearch, planetSearch, seek)
- **Backend nova-endpoint**: `GET /api/search`
- **Сложность воспроизведения**: низкая

### S-039. Модератор (Moderator)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Moderator.class.php`
- **Шаблоны**: `templates/standard/moderate_user.tpl`
- **URL**: `?go=Moderator` (+ &action=proceedBan, proceedRO, annulBan, annulRO, proceed)
- **Backend nova-endpoint**: модерация — отдельный admin-bff
  (план 53)
- **Сложность воспроизведения**: средняя (только для модераторов)

### S-040. Мониторинг планеты (MonitorPlanet)

- **Контроллер**: `projects/game-legacy-php/src/game/page/MonitorPlanet.class.php`
- **Шаблоны**: `templates/standard/monitor_planet.tpl`
- **URL**: `?go=MonitorPlanet/X`
- **Backend nova-endpoint**: вероятно через `GET /api/phalanx`
  (фаланга, сканирование) и `GET /api/galaxy/{g}/{s}`
- **Сложность воспроизведения**: низкая

### S-041. Блокнот (Notepad)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Notepad.class.php`
- **Шаблоны**: `templates/standard/notes.tpl`
- **URL**: `?go=Notepad` (+ &action=saveNotes)
- **Backend nova-endpoint**: `GET /api/notepad`,
  `PUT /api/notepad`
- **Сложность воспроизведения**: низкая

### S-042. Империя (Empire)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Empire.class.php`
- **Шаблоны**: `templates/standard/empire.tpl`
- **URL**: `?go=Empire`
- **Backend nova-endpoint**: `GET /api/empire`
- **Сложность воспроизведения**: средняя

### S-043. Редактирование здания (EditConstruction, admin)

- **Контроллер**: `projects/game-legacy-php/src/game/page/EditConstruction.class.php`
- **Шаблоны**: `templates/standard/edit_construction.tpl`
- **URL**: `?go=EditConstruction` (+ &action=saveConstruction, addRequirement, deleteRequirement)
- **Назначение**: Админ-инструмент. **Не для прода клона** —
  делается в `admin-frontend` (план 53).
- **Сложность воспроизведения**: не для основного клона

### S-044. Редактирование юнита (EditUnit, admin)

- **Контроллер**: `projects/game-legacy-php/src/game/page/EditUnit.class.php`
- **Шаблоны**: `templates/standard/edit_unit.tpl`
- **URL**: `?go=EditUnit` (+ &action=saveConstruction)
- **Назначение**: Админ-инструмент.
- **Сложность воспроизведения**: не для основного клона

### S-045. Калькулятор технологий (AdvTechCalculator)

- **Контроллер**: `projects/game-legacy-php/src/game/page/AdvTechCalculator.class.php` (58 стр)
- **Шаблоны**: `templates/standard/adv_tech_calc.tpl`
- **URL**: `?go=AdvTechCalculator`
- **Backend nova-endpoint**: client-side только
- **Сложность воспроизведения**: средняя (JS-калькулятор)

### S-046. Симулятор боя (Simulator)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Simulator.class.php` (749 стр)
- **Шаблоны**: `templates/standard/simulator.tpl`
- **URL**: `?go=Simulator` (+ &action=simulate, addParticipant, resetAssault)
- **Backend nova-endpoint**: `POST /api/battle-sim` (есть в nova)
- **Сложность воспроизведения**: высокая

### S-047. Ракетная атака (RocketAttack)

- **Контроллер**: `projects/game-legacy-php/src/game/page/RocketAttack.class.php`
- **Шаблоны**: `templates/standard/rocket_attack.tpl`
- **URL**: `?go=RocketAttack` (+ &action=sendRockets)
- **Backend nova-endpoint**: `POST /api/planets/{id}/rockets/launch`,
  `GET /api/planets/{id}/rockets`
- **Сложность воспроизведения**: средняя

### S-048. Ремонт (Repair)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Repair.class.php` (688 стр)
- **Шаблоны**: использует общие
- **URL**: `?go=Repair` (+ &action=order, abortRepair, abortDefense, abortDisassemble, abortEvent + VIP-варианты)
- **Backend nova-endpoint**: `POST /api/planets/{id}/repair/{repair,disassemble}`,
  `GET /api/planets/{id}/repair/{queue,damaged}`
- **Сложность воспроизведения**: средняя

### S-049. Демонтаж (встроено в Repair)

- **Контроллер**: встроено в `Repair.class.php` (`isRepair()`)
- **Сложность**: встроено

### S-050. Биржа (Stock, legacy)

- **Контроллер**: `projects/game-legacy-php/src/game/page/Stock.class.php` (757 стр)
- **Шаблоны**: `templates/standard/stock.tpl`,
  `lot_details.tpl`
- **URL**: `?go=Stock` (+ &action=buyLot, recall, premiumLot, ban,
  showLots, showLotDetails, deleteExpireEvent)
- **Backend nova-endpoint**: ОТСУТСТВУЕТ — `D-NNN-EXCHANGE`
  (биржа артефактов)
- **Сложность воспроизведения**: высокая

### S-051. Биржа новая (StockNew)

- **Контроллер**: `projects/game-legacy-php/src/game/page/StockNew.class.php` (850 стр)
- **Шаблоны**: `stock_new_1.tpl`, `stock_new_2.tpl`,
  `stock_new_3.tpl`, `lot_details.tpl`, `empty.tpl`,
  `missions.tpl`
- **URL**: `?go=StockNew` (+ &action=addLot, addArtefacts,
  selectFleet, lotOptions, lotOptions2, showExchanges)
- **Backend nova-endpoint**: ОТСУТСТВУЕТ — `D-NNN-EXCHANGE`
- **Сложность воспроизведения**: высокая

### S-052. Статистика обмена ресурсами (ResTransferStats)

- **Контроллер**: `projects/game-legacy-php/src/game/page/ResTransferStats.class.php`
- **Шаблоны**: `templates/standard/restransfers.tpl`
- **URL**: `?go=ResTransferStats` (+ &action=showResTransfer)
- **Backend nova-endpoint**: возможно через `resource_transfers`
  таблицу — нужен endpoint
- **Сложность воспроизведения**: средняя

### S-053. Тестирование AlienAI (TestAlienAI, admin)

- **Контроллер**: `projects/game-legacy-php/src/game/page/TestAlienAI.class.php` (32 стр)
- **Шаблоны**: нет
- **URL**: `?go=TestAlienAI`
- **Назначение**: Dev-only. Не для прода.
- **Сложность воспроизведения**: не для прод-клона

### S-054. FleetAjax (AJAX-эндпоинт, не отдельный экран)

- **Контроллер**: `projects/game-legacy-php/src/game/page/FleetAjax.class.php`
- **Назначение**: AJAX из Galaxy/Mission (espionage, format, index).
- **Сложность**: встроено в S-005, S-006

### S-055. Page, Construction, Logout (служебные классы)

- `Page.class.php` (539 стр) — абстрактный базовый класс
- `Construction.class.php` (498 стр) — общие хелперы для построек
- `Logout.class.php` (53 стр) — `?go=Logout`, очистка сессии

Логаут в origin-фронте на nova: вызов identity logout.

---

## Layout-системы

### Корневой шаблон (`layout.tpl`)

- **DOCTYPE**: XHTML 1.0 Strict
- **Head**:
  - charset UTF-8
  - jQuery 1.5.1 + jQuery UI 1.8.14 (CDN googleapis.com)
  - CSS: `layout.css`, `style.css` + skin-варианты
    (`vkontakte.css`, `fb.css`, `mobi.css`)
  - JS: `main.js`, `jquery.cookie.js`, `jquery.countdown.js`
- **Body** (3 области):
  - **Левое меню** `<div id="leftMenu"><ul>` — основная навигация
    (Constructions, Research, Shipyard, Galaxy, Mission, MSG,
    Alliance, …)
  - **Основной контент** `<div class="main_content">` — экран
  - **Шапка** — ник + текущая планета + ресурсы + быстрые ссылки
- **Футер**: юр-ссылки (план 50), копирайт, версия, online-статус

### Шапка / навигация

- Ник игрока, текущая планета и координаты
- Быстрые ресурсы (Metal, Silicon, Hydrogen, Energy, Credits)
- Быстрые ссылки: Profile, Alliance, Messages, Support, Logout

### Сайдбар (левое меню)

- Иерархическая навигация
- Группировка: Resources/Building, Research/Tech, Fleet/Defense,
  Diplomacy/Alliance, Account
- Активный пункт подсвечен

### Футер

- User Agreement (S-033)
- Privacy policy
- Контактный email
- Версия игры
- Online-статистика

---

## Глобальные визуальные ресурсы

### Шрифты

- **Системные**: Arial, Verdana, sans-serif (стандартные браузерные)
- **Веб-шрифты**: НЕ используются специальные TTF/WOFF
- jQuery UI CSS из googleapis.com (тема `ui-darkness`)

### Палитра

Базовые цвета (из `public/css/style.css`, `layout.css`):

- **Фоны**: `#000000` (основной), `#1a1a1a` (тёмно-серый), `#ffffff`
  (текст)
- **Семантика**:
  - Успех/положительное: `#00ff00` (`.true`, `.available`)
  - Ошибка/дефицит: `#fc3232` (`.false`)
  - Светлая ошибка/notavailable: `#fd7171` (`.notavailable`,
    `.false2`)
  - Warning: `#d57c08` (`.rep_quantity_damage_low`)
  - Голубой/живой (прогресс): `#6cd8bc` (`.rep_alive_over_div`)
- **Кнопки/инпуты**: оранжевые/серые

### Иконки

- **Расположение**: `public/images/`
- **Формат**: PNG основной, GIF для loader'ов
- **Размеры**: 16×16 (списки), 32×32 (кнопки), 48×48 (профили),
  256×256 (логотипы альянсов)
- **Количество**: ~200+ файлов

### Бэкграунды

- `public/images/bg/bg.jpg`, `bg2.gif` — статические текстуры
- На чёрном фоне для глубины

### Стиль кнопок/инпутов

- **Кнопки**: класс `button`, оранжевый/серый фон, hover осветлённый,
  active вдавленный (inset shadow)
- **Инпуты**: светлый фон, серый бордер 1px, focus подсветка,
  padding 5-8px

---

## Типовые таблицы (переиспользуемые компоненты)

### 1. Constructions Table
Колонки: №, Название, Уровень, Статус, Стоимость (M/Si/H), Энергия,
Время, Действия (Upgrade/Cancel/Demolish/VIP/Info)

### 2. Shipyard Table
Иконка, Название, Имеется, Стоимость, Время, Энергия, Поле количества,
Итого, Действия

### 3. Galaxy View Table (15 строк × ~7 колонок)
Позиция, Тип (планета/луна), Название, Владелец, Альянс, Уровень
главного здания, Действия (атаковать, шпионить, сообщение)

### 4. Mission Table (history)
Дата, Тип миссии, Откуда, Куда, Флот (краткое), Статус, Время
прибытия, Действия (отступить, отчёт)

### 5. Research Table
Название, Уровень, Стоимость, Время, Требования (link), Действия

### 6. Stock Table
Название артефакта, Раритет, Цена (с/без скидки), Срок (TTL),
Продавец, Действия (купить, отозвать)

### 7. Messages Table
От/Кому, Тема, Дата, Статус (read/unread), Действия

### 8. Alliance Members Table
Позиция, Ник, Online-статус (badge), Ранг (rank_name), Последний
вход, Деятельность, Действия (профиль, message, kick)

### 9. Ranking Table
Позиция, Ник, Альянс, Очки, Изменение (↑/↓), Действия

### 10. Repair Table
Название юнита, Уровень повреждения (% bar), Стоимость, Время,
Кол-во, Итого, Действия

---

## Сводка по покрытию nova-API

Из 55 контроллеров origin:
- **Полностью покрыты nova-API**: ~20 (Main, Research, Constructions,
  Shipyard, Galaxy, MSG, Friends, Artefacts, Battlestats, Ranking,
  Records, Resource, Market, Preferences, Profession, Officer,
  Search, Notepad, Empire, Repair)
- **Частично покрыты**: ~15 (Alliance — 17/30 действий, Mission —
  не все типы миссий, Achievement — упрощённая модель, Tutorial —
  через goal engine)
- **Полностью отсутствуют**: ~10 (Exchange/ExchangeOpts/Stock/StockNew
  — биржа артефактов, ResTransferStats — статистика переводов,
  RocketAttack — частично, Tournament — отсутствует, MonitorPlanet —
  через phalanx)
- **Не воспроизводятся как клон** (заменяются единым cross-universe):
  Payment (billing-service), Moderator (admin-bff), EditConstruction
  и EditUnit (admin), TestAlienAI (dev-only)

---

## Заметки для агента-реализатора

1. **`?go=Page`** — основной URL-метод. PATH_INFO иногда работает,
   но не везде — см. memory `reference_game_origin_routing.md`.
2. **dev-login.php** — мгновенный вход как `test`/userid=1 для
   проверки экранов (`docs/legacy/game-legacy-access.md`).
3. **Все шаблоны Smarty в `src/templates/standard/`** — 125 файлов.
   Часть — партиалы (`_report_button.tpl`, `before_content.tpl`,
   `front.tpl`), не S-NNN сами.
4. **JS** — в основном jQuery + AJAX. В клоне переписывается на
   React — состояния через Zustand, кеш через TanStack Query.
5. **BBCode чата** — НЕ переносится. Чат в обоих фронтах (nova и
   origin) использует TipTap (план 57).

---

## References

- 55 контроллеров: `projects/game-legacy-php/src/game/page/`
- 125 шаблонов: `projects/game-legacy-php/src/templates/standard/`
- Стили: `projects/game-legacy-php/public/css/`
- Иконки: `projects/game-legacy-php/public/images/`
- [docs/legacy/game-legacy-access.md](../../legacy/game-legacy-access.md)
  — доступ к запущенному game-origin
- [origin-inventory.md](origin-inventory.md) — детально по
  контроллерам с actions
