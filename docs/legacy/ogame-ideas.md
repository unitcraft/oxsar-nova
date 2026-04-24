# Топ новых идей игроков OGame

Дистилляция предложений и идей сообщества OGame для дизайна **oxsar-nova**
(порт legacy oxsar2 на Go/TS/PG). Дата сбора — 2026-04-24.
Дисклеймер: это срез того, что просят сами игроки на форумах Gameforge,
сабреддите, фан-блогах и тематических ветках. **Популярность ≠ качество идеи**:
часть предложений ломает баланс или конфликтует с философией «медленной
4X-стратегии», часть — реально полезна. Релевантность для oxsar-nova указана
по каждому пункту.

---

## Новые механики (юниты, здания, исследования, события)

- **Wreck Fields (поля обломков от защиты)** — после боя помимо обычного
  debris field появляется «поле обломков» из потерянной защиты, доступное
  только защитнику. Возвращает часть ресурсов вайпнутому игроку, чтобы он
  не бросал аккаунт. _Источник: board.en.ogame.gameforge.com (анонс OGame
  team)_. _Популярность: часто упоминается в feedback-тредах_.
  - Релевантность: **подходит**. Закрывает классическую дыру «один краш —
    игрок ушёл», стыкуется с темой retention в release-roadmap.

- **Holo Domes / ремонт защиты** — здание с уровнями: каждый уровень
  возвращает % уничтоженной защиты после боя и повышает долю «спрятанных»
  ресурсов. Делает оборонительный стиль жизнеспособным.
  _Источник: «Massive Update suggestion (Tech, lategame, defensive combat)»
  на board.en.ogame.gameforge.com_. _Популярность: единичная, но обсуждаемая_.
  - Релевантность: **с оговорками**. Полезно для tortle-стиля, но требует
    ADR — затрагивает базовую формулу боя.

- **Lifeforms-tier-tree (4 расы с уникальными ветками технологий)** —
  4 расы (Humans / Rocktal / Kaelesh / Mecha), 48 зданий и 72 технологии,
  по 12+18 на расу. Каждая раса даёт уникальный бонус (Mecha — быстрее
  верфь, Kaelesh — экспедиции, и т. д.). _Источник: ogame.fandom.com,
  bleedingcool.com_. _Популярность: официально внедрено, активно
  обсуждается_.
  - Релевантность: **с оговорками**. Большой объём, но дать «класс игрока»
    с веткой бонусов — естественный endgame-контент.

- **Новые классы кораблей под анти-meta** — ship-killer/Starcrusher против
  RIP-only флотов, support-корабли с временными щитами-пузырями, destroyer
  с armor-penetration против battleship. _Источник: board.origin.ogame
  («Bag of ideas», «Starcrusher»), OGame Showcases_. _Популярность: часто
  встречается, тема живая_.
  - Релевантность: **подходит**. Прямое расширение боевой матрицы —
    легко добавить через configs/ без слома формул.

- **Trade routes между планетами** — автоматические торговые маршруты,
  завязанные на специализированные здания. _Источник: OGame Showcases
  (developer hints)_. _Популярность: единичные предложения_.
  - Релевантность: **с оговорками**. Перекрывается с «transport templates»
    из QoL-блока — лучше делать шаблонами, чем зданиями.

---

## QoL и автоматизация

- **Очередь стройки на N уровней одной кнопкой («Queue next 5 levels»)** —
  главный QoL-запрос. Сейчас лимит 5 в очереди и каждый клик = 1 уровень,
  что в Lifeforms (40+ уровней одного здания) превращается в кликер.
  _Источник: «Building Queue QoL Improvements Needed» на
  board.en.ogame.gameforge.com_. _Популярность: топ-1 QoL-жалоба_.
  - Релевантность: **подходит**. Дешёвая фича, огромный эффект на UX.

- **Очередь зданий между планетами** — общий планировщик, в котором можно
  собрать «X на планете A, Y на планете B» и запустить. _Источник:
  board.origin.ogame.gameforge.com_. _Популярность: часто_.
  - Релевантность: **подходит**. Хорошо ложится на event-loop воркер.

- **Easy Transport / шаблоны транспортов** — скрипт-фича, которую просят
  встроить: выбираешь нужные ресурсы для постройки/исследования, движок
  сам подбирает грузовики и шлёт с нескольких планет. _Источник:
  «OGame Redesign: Easy Transport»_. _Популярность: десятилетие в топе_.
  - Релевантность: **подходит**. Идеально под TanStack Query + WS-обновления.

- **Sub-second build time для дешёвых юнитов** — снять лимит «не быстрее
  1 сек», особенно для лёгких истребителей и спутников.
  _Источник: «Allow time to build to go under 1sec»_. _Популярность: часто_.
  - Релевантность: **с оговорками**. Только если verfь-формулы выдерживают;
    иначе ломает экономику дешёвых юнитов.

- **Двойная верфь / параллельные линии** — большая партия кораблей делится
  на 2 параллельных потока. _Источник: «Double shipyard»,
  «Improving Shipyard while producing units»_. _Популярность: часто_.
  - Релевантность: **с оговорками**. Меняет балансные формулы, нужен ADR.

- **Cooldown на режим отпуска** — нельзя сразу выйти и сразу снова войти
  в vacation mode (закрывает эксплойт). _Источник: «Wishlist» на
  board.origin.ogame_. _Популярность: единично, но логично_.
  - Релевантность: **подходит**. Строчка в audit.md и одна проверка
    в auth/планетарном сервисе.

---

## Экономика и торговля

- **Player marketplace v2 (с защитой от пушинга)** — внутриигровой рынок
  ресурсов/кораблей/айтемов, отключённый Gameforge в 2020 из-за абуза.
  Сообщество просит вернуть, но с лимитом ratio (3:2:1 ÷ 2:1:1) и анти-push
  правилами. _Источник: ogame.fandom.com/wiki/Marketplace, twitter @OGame_.
  _Популярность: высокая, но контроверсная_.
  - Релевантность: **с оговорками**. В release-roadmap уже стоит как
    риск — внедрять только с anti-push fence (лимит по очкам/возрасту).

- **Лимит выигранных аукционов в день** — чтобы топ-аккаунт не выкупал
  весь аукцион предметов, ограничить N побед/24ч на игрока. _Источник:
  обсуждения на board.us.ogame_. _Популярность: средняя_.
  - Релевантность: **подходит**. Простое анти-доминирование.

- **Crystal mine multiplier 1.6 → 1.5** — экономика косая в сторону металла
  (~4M:1C по факту), сообщество просит выровнять. _Источник: «Wishlist»
  тред_. _Популярность: средняя_.
  - Релевантность: **с оговорками**. Менять баланс запрещено CLAUDE.md
    без ADR — но тема для balance/audit.md.

- **Лимит покупки premium-валюты в первый месяц вселенной** — против
  P2W-старта: в первые 30 дней лимит на покупку Dark Matter / Diamonds,
  чтобы кошелёк не решал стартовый рывок. _Источник: «Pay to Win Ogame
  Yin» на board.en.ogame_. _Популярность: высокая_.
  - Релевантность: **подходит**. Хорошая защита fair-play на старте сезона.

- **Subscription plan ~5€/мес** — пакет QoL-перков (расширенная очередь,
  больше слотов офицера, история боёв) вместо чистого P2W. _Источник:
  длинные треды на форуме en.ogame_. _Популярность: высокая_.
  - Релевантность: **с оговорками**. Монетизационное решение —
    уровень product, не engineering, но закладывать на уровне фичфлагов.

---

## PvP и бой

- **In-game battle simulator** — за 20 лет так и не появился; сообщество
  требует встроенный симулятор боя как у внешних tools (osimulate.com,
  ogame-tools.com). _Источник: ogame.fandom.com/wiki/Simulator, jstar88/opbe
  GitHub_. _Популярность: топ-3 запросов_.
  - Релевантность: **подходит, must-have**. Уже есть отдельный battle-engine
    в Go — простой UI поверх него.

- **3-секундная задержка манёвра после посадки флота** — анти-бот мера:
  бот ловит секундные паузы, человек — нет. _Источник: «Bot issue and
  possible solution?» на board.en.ogame_. _Популярность: средняя_.
  - Релевантность: **с оговорками**. Влияет на legit fleetsave-практики;
    нужен playtest.

- **Precision counter-attack механика** — атакующий со спутниками-разведчиками
  в составе флота сдвигает старт боя на десятки секунд–минут (анти-«секундный
  тайминг»). _Источник: ogame combat NamuWiki_. _Популярность: уже в OGame,
  но просят расширить_.
  - Релевантность: **подходит**. Глубина боевого тайминга без слома формул.

- **Defender deploys reinforcement** — защитник может за секунды до боя
  подкинуть флот или достроить оборону. _Источник: ogame.fandom.com/wiki/
  Combat_. _Популярность: уже частично есть_.
  - Релевантность: **подходит**. Уточнение для боевого движка.

- **Турниры/лиги/ладдер** — формальный соревновательный режим
  (1v1, alliance vs alliance) с сезоном и наградами. _Источник: общий
  тренд в обсуждениях, но формально — пробел_. _Популярность: единичные
  предложения_.
  - Релевантность: **с оговорками**. Хорошо для retention, но это
    отдельный продукт поверх ядра.

- **Новые fleet speeds (5%, 2%)** — расширить набор скоростей флота для
  более дешёвого fleetsave. _Источник: «Wishlist» тред_.
  _Популярность: единично_.
  - Релевантность: **подходит**. Тривиально в configs/.

---

## Социалка и альянсы

- **Бонусы за членство в альянсе** — малые баффы: +% к производству от
  числа активных членов, ускорение исследований от суммарных tech-points
  альянса. _Источник: «Alliances Improvements» на board.origin.ogame,
  Facebook OGame poll_. _Популярность: высокая_.
  - Релевантность: **подходит**. Превращает альянс из чата в gameplay-сущность.

- **Формальная объявка войны в игре (а не на форуме)** — сейчас 12-часовой
  протокол объявки живёт в форумном posts; перенести в UI с автоматическим
  таймером и whitelist-режимом. _Источник: ogame.fandom.com/wiki/War_.
  _Популярность: средняя_.
  - Релевантность: **подходит**. Естественный UI-feature.

- **NAP / pact как игровая сущность** — Non-Attack Pact в виде записи
  в БД с автомеханикой блокировки атак между сторонами. _Источник:
  ogame.fandom.com/wiki/Diplomacy_. _Популярность: средняя_.
  - Релевантность: **с оговорками**. Хорошая идея, но повышает связность
    систем — в фазу 2.

- **Alliance Combat System / совместные атаки** — несколько игроков
  альянса объединяют флоты в одну атаку с общим debris-распределением.
  _Источник: ogame.fandom.com/wiki/Alliance_Combat_System (уже есть)_.
  _Популярность: high uptake_.
  - Релевантность: **подходит**. В legacy oxsar2 этого нет — заметный апгрейд.

- **Альянсовая казна и общие исследования** — общий фонд ресурсов
  с правами доступа + research-ветка, видимая всему альянсу.
  _Источник: «Alliance System» обсуждения_. _Популярность: средняя_.
  - Релевантность: **с оговорками**. Усложняет push-detection.

---

## UX/UI

- **Flat coordinates picker вместо dropdown** — старый OGame имел список
  координат таблицей в один клик; redesign перевёл на dropdown — медленнее.
  _Источник: «OGame Redesign: Flat Coordinates Shortcut List»_.
  _Популярность: высокая среди ветеранов_.
  - Релевантность: **подходит**. Простой UX-fix.

- **Полноценные hotkeys** — Shift-G/S для системы/галактики, Shift-M/K/D
  для аукционов, шорткаты на меню, отправку флота, fleet-save. _Источник:
  «[H] Hotkeys» на board.origin.ogame, OGame UI++_. _Популярность: высокая_.
  - Релевантность: **подходит**. Дешёвая фича на frontend.

- **Mobile-first интерфейс** — нативный mobile app в beta, но многие функции
  не работают; community просит paritet с web. _Источник: «OGame Mobile App
  Changelogs»_. _Популярность: топ-запрос на 2024-2026_.
  - Релевантность: **подходит**. Уже в release-roadmap (responsive UI).

- **Открытый аналог OGame UI++** — браузерное расширение добавляет меню,
  hover-карточки, инлайн-симулятор; пользователи фактически не играют без
  него. _Источник: chrome web store OGame UI++_. _Популярность: де-факто
  стандарт_.
  - Релевантность: **подходит**. Встроить базовый функционал в нативный UI
    с самого начала — конкурентное преимущество.

- **Async planet/moon star indicators** — уведомления о действиях
  противника на твоих координатах через WS, без перезагрузки галактики.
  _Источник: «Wishlist» тред_. _Популярность: средняя_.
  - Релевантность: **подходит**. Уже есть WS-инфра.

---

## Retention и endgame

- **Восстановление после краш-вайпа** — комплекс мер (wreck field,
  «возрождение» флота за % ресурсов, временный щит после потери флота
  >X% состояния). _Источник: feedback-треды OGame team_. _Популярность:
  высокая, главная причина оттока_.
  - Релевантность: **подходит, критично**. В release-roadmap пометить
    как retention-feature.

- **Сезонные/speed серверы с финалом** — turbo-вселенная с понятным
  концом сезона, призами, и переносом косметики в основную. _Источник:
  Server Settings wiki, private OGame servers feature-set_.
  _Популярность: высокая_.
  - Релевантность: **подходит**. Хорошо ложится на multi-tenant архитектуру.

- **Prestige / rebirth** — после конца сезона игрок «перерождается»
  в новой вселенной с косметикой/мини-бонусом. _Источник: общий тренд
  4X-сообщества, не специфично OGame_. _Популярность: единично_.
  - Релевантность: **с оговорками**. Не хардкорный OGame-вайб, но
    хороший hook для нового игрока.

- **Achievements/lore-разблокировки за экспедиции** — экспедиции уже
  дают лут; добавить редкие лор-награды (skin/имя/декор корабля).
  _Источник: ogame.fandom.com/wiki/Expedition + Lifeforms FAQ_.
  _Популярность: средняя_.
  - Релевантность: **подходит**. Дешёвый retention-крючок.

- **Persistent история боёв и stats-кабинет** — личная страница со всей
  историей атак/защит/экспедиций, графики, top-list. _Источник:
  обсуждения hover-карт в UI++_. _Популярность: средняя_.
  - Релевантность: **подходит**. Удобно для маркетинга и стримерства.

---

## Античит и честность

- **Captcha раз в 6 часов** — лёгкий human-check, который сообщество
  готово терпеть в обмен на чистоту серверов. Многие признают, что 50%+
  «играют» через скрипты. _Источник: «Bot issue and possible solution?»
  на board.en.ogame, обсуждения на board.fr.ogame_. _Популярность:
  поляризующая, но активная_.
  - Релевантность: **с оговорками**. UX-trade-off; в audit.md записать
    как опцию для fair-серверов.

- **Динамический детект ботов** — серверная аналитика интервалов
  кликов/секундной точности fleetsave с soft-flagging. _Источник:
  «Reflexion Anti Bot Ogame» на board.fr.ogame_. _Популярность: средняя_.
  - Релевантность: **подходит**. Плотно ложится на event-loop логи.

- **Anti-push fences** — жёсткий лимит на трансферы между аккаунтами:
  по % очков, возрасту аккаунта, ratio ресурсов. _Источник: причина
  отключения Marketplace, ogame.fandom.com/wiki/Marketplace_.
  _Популярность: высокая среди honest-игроков_.
  - Релевантность: **подходит, must-have**. В oxsar-nova внедрить
    с первого дня — иначе marketplace мёртв.

- **Fair-серверы без premium-валюты** — отдельные вселенные без Dark
  Matter/Diamonds для пуристов. _Источник: общий тред о P2W,
  «Pay to Win Ogame Yin»_. _Популярность: средняя_.
  - Релевантность: **подходит**. Server-flag в БД, легко включить.

- **80%-голос на изменение настроек сервера** — community-driven server
  settings: запрос → poll → если 80% за, Gameforge применяет.
  _Источник: ogame.fandom.com/wiki/Server_Settings_. _Популярность:
  высокая_.
  - Релевантность: **с оговорками**. Хорошая идея для прозрачности,
    но операционно тяжёлая.

---

## Источники

- https://board.en.ogame.gameforge.com/index.php — главный английский форум Gameforge
- https://board.en.ogame.gameforge.com/index.php?thread/831903-building-queue-qol-improvements-needed-lifeforms-feedback/ — Building Queue QoL
- https://board.en.ogame.gameforge.com/index.php?thread/823200-massive-update-suggestion-tech-lategame-defensive-combat/ — Massive update suggestion (Holo Domes)
- https://board.en.ogame.gameforge.com/index.php?thread/825985-bot-issue-and-possible-solution/ — Bot issue & captcha
- https://board.en.ogame.gameforge.com/index.php?thread/843056-pay-to-win-ogame-yin-new-universe/ — P2W discussion
- https://board.en.ogame.gameforge.com/index.php?thread/832023-lifeforms-feedback-thread/ — Lifeforms feedback
- https://board.us.ogame.gameforge.com/index.php?thread/96103-double-shipyard/ — Double shipyard
- https://board.us.ogame.gameforge.com/index.php?thread/99684-building-fleet-faster-than-1-second/ — sub-second build
- https://board.us.ogame.gameforge.com/index.php?board/1023-suggestions/ — US Suggestions board
- https://board.origin.ogame.gameforge.com/index.php/Thread/10815-Wishlist/ — Wishlist
- https://board.origin.ogame.gameforge.com/index.php/Thread/10431-Bag-of-ideas-ships-defenses-and-laser-tech/ — Bag of ideas
- https://board.origin.ogame.gameforge.com/index.php/Thread/1759-New-ship-Starcrusher/ — Starcrusher
- https://board.origin.ogame.gameforge.com/index.php/Thread/2154-Improving-Shipyard-while-producing-units/ — Shipyard parallel
- https://board.origin.ogame.gameforge.com/index.php/Thread/11238-Allow-time-to-build-to-go-under-1sec-ships-nanite/ — sub-1s build
- https://board.origin.ogame.gameforge.com/index.php/Thread/1904-OGame-Redesign-Easy-Transport/ — Easy Transport
- https://board.origin.ogame.gameforge.com/index.php/Thread/5246-Fleet-Routine-Transport-Destination-improvement/ — Routine Transport
- https://board.origin.ogame.gameforge.com/index.php/Thread/4144-Alliance-System/ — Alliance system
- https://board.origin.ogame.gameforge.com/index.php/Thread/11374-Alliances-Improvements/ — Alliance improvements
- https://board.origin.ogame.gameforge.com/index.php/Thread/10289-Auction-suggestion/ — Auction suggestion
- https://board.origin.ogame.gameforge.com/index.php/Thread/10729-Adding-bots-to-Ogame-universes-with-low-player-counts/ — Bots in low-pop unis
- https://board.origin.ogame.gameforge.com/index.php/Thread/5275-H-Hotkeys/ — Hotkeys
- https://board.fr.ogame.gameforge.com/index.php?thread/755994-reflexion-anti-bot-ogame/ — Reflexion anti-bot (FR)
- https://forum.origin.ogame.gameforge.com/forum/thread/64-lifeform/ — Lifeform PTS
- https://forum.origin.ogame.gameforge.com/forum/thread/153-lifeforms-faq/ — Lifeforms FAQ
- https://forum.origin.ogame.gameforge.com/forum/thread/172-oglight/ — OGLight tool
- https://ogame.fandom.com/wiki/Marketplace — Marketplace wiki
- https://ogame.fandom.com/wiki/Auctioneer — Auctioneer wiki
- https://ogame.fandom.com/wiki/Alliance_Combat_System — ACS wiki
- https://ogame.fandom.com/wiki/Diplomacy — Diplomacy wiki
- https://ogame.fandom.com/wiki/War — War wiki
- https://ogame.fandom.com/wiki/Combat — Combat wiki
- https://ogame.fandom.com/wiki/Simulator — Simulator wiki
- https://ogame.fandom.com/wiki/Bot — Bot wiki
- https://ogame.fandom.com/wiki/Server_Settings — Server settings wiki
- https://ogame.fandom.com/wiki/Lifeform — Lifeform wiki
- https://ogame.fandom.com/wiki/Expedition — Expedition wiki
- https://chrome.google.com/webstore/detail/ogame-ui++/nhbgpipnadhelnecpcjcikbnedilhddf — OGame UI++ extension
- https://github.com/Azkellas/ogame-mobile — Ogame mobile project
- https://userscripts-mirror.org/scripts/show/81287 — Flat coordinates shortcut
- https://userscripts-mirror.org/scripts/show/83284 — Keyboard shortcuts userscript
- https://osimulate.com/ — OSimulate
- https://simulator.ogame-tools.com/ — OGame-Tools simulator
- https://battlesim.logserver.net/en — BattleSim
- https://github.com/jstar88/opbe — Probabilistic Battle Engine
- https://bleedingcool.com/games/ogame-introduces-the-new-lifeforms-expansion-today/ — Lifeforms launch
- https://x.com/ogame/status/1167346214213443585 — Marketplace announcement
- https://en.namu.wiki/w/%EC%98%A4%EA%B2%8C%EC%9E%84/%EC%A0%84%ED%88%AC — OGame combat namuwiki
