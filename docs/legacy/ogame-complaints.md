# Жалобы игроков OGame

Дистилляция типовых болей и претензий игроков OGame, собранная для дизайна
oxsar-nova. Цель — учесть исторические грабли при проектировании механик
порта, особенно в зонах боя, экономики, ретеншна и монетизации.

Дата сбора: 2026-04-24. Источники — официальные форумы Gameforge
(board.en/us/origin.ogame.gameforge.com), wiki/fandom, обзоры (Steemit,
mmommorpg), Trustpilot, профильные блоги. Это не научная выборка, а сводка
повторяющихся тем из открытых обсуждений; цитаты — пересказ или короткие
фразы с английских веток.

## Геймплей и темп

- **Бесконечный гринд ресурсов** — каждое здание/исследование требует часов
  и дней реального времени; даже late-game игрок «тратит дни на одну
  постройку». _Источник: ogame.life blog, OGame Wiki, форумные guides._
  - Импликация для oxsar-nova: жёстко лимитировать «пустое ожидание» —
    очереди с офлайн-ресурсом, instant-buildings из кредитов внутри лимита,
    ускорение раннего онбординга.
- **Asymmetric PvP / фарм слабых** — топ-игроки «фермят» миды и новичков; в
  старых вселенных порог защиты новичка (50k очков) бессмысленен.
  _Источник: board.en.ogame thread 842864 «Updated New Player Protection»._
  - Импликация: динамическая шкала noob-protection (по медиане сервера, а не
    фикс. очкам), soft-cap на отношение атакующий/защитник.
- **Turtling — мёртвый стиль игры** — «ленивые шахтёры строят защиту в
  избытке и становятся черепахами», которых невозможно ни сломать, ни
  выгнать; PvP не происходит. _Источник: OGame Wiki, форум-гайды._
  - Импликация: затраты на оборону должны иметь diminishing returns;
    moon-destroyer/RIP-эквивалент должен реально пробивать turtle.
- **Длительность вселенной vs. усталость** — «срок жизни новых вселенных —
  максимум 80 дней», старые вселенные «опустошаются с каждой новой».
  _Источник: board.en.ogame thread 843056._
  - Импликация: продумать season/merge цикл заранее, не плодить пустые
    вселенные; механика возвращения игрока в новый сезон с бонусом.
- **Класс-имбаланс (Collector OP)** — Collector в late-game даёт +25%/+10%
  ресурсов, с бустерами до 75–100%; «много игроков уходят из-за этого
  дисбаланса». _Источник: board.us.ogame thread 98263 «class imbalance»._
  - Импликация: классы/расы — только сайдгрейды, симметричный power-budget,
    регулярный балансный аудит с публичными changelog'ами.
- **Экспедиции — лотерея, не стоящая слотов** — «никто не отправит 5–7
  экспедиций ради рандомного лута, когда слоты можно использовать на фарм».
  _Источник: board.us.ogame thread 97251 «Discoverer»._
  - Импликация: экспедиции должны давать стабильный ожидаемый профит +
    редкие jackpot'ы, а не «5 пустых из 7».

## Экономика и ресурсы

- **Дейтерий съедает всё** — «даже маленький флот тратит 20–60M deut в день
  только на fleetsave», что часто превышает дневное производство.
  _Источник: board.en.ogame thread 816997 «Ways of handling deut»._
  - Импликация: пересмотреть расход топлива; рассмотреть «безопасную стоянку»
    (бункер) как альтернативу постоянному перелёту.
- **Инфляция металла, дефицит deut** — поздняя экономика держится на
  «отношениях с шахтёрами» и маркете deut; одиночный игрок не выживает.
  _Источник: «The Ultimate Miner Guide», board.en.ogame._
  - Импликация: рынок ресурсов должен быть ликвидным с самого начала;
    встроенный AMM/NPC-обмен с предсказуемым курсом, чтобы не зависеть от
    клик-картелей.
- **Маркет/трейд костыльный** — нет нормального ордербука, обмены идут
  через личные сообщения и trust. _Источник: общий тон гайдов и форумов._
  - Импликация: сразу проектировать exchange с лимитными ордерами, escrow и
    антифрод-проверкой соотношений.
- **Storage caps и переполнение** — «склады заполнились — ресурсы стоят,
  никаких альтернативных стоков». _Источник: гайды по resource saving._
  - Импликация: overflow-buffer с штрафом, либо автоматическая
    конвертация/сжигание в опыт.
- **Dark Matter pay-to-resources** — «прямая покупка ресурсов на основе
  capacity склада» — «очень pay-to-win опция». _Источник: board.en.ogame
  thread 843056, 847189._
  - Импликация: монетизация не должна продавать ресурсы напрямую;
    конвертировать только в косметику/QoL/ускорение очередей в лимитах.

## Бой

- **Fleet crash — потерять всё за одну ошибку** — «If it sits, it gets
  hit»; одна забытая fleetsave = недели/месяцы прогресса в труху.
  _Источник: OGame Wiki Fleetsaving, форум-гайды._
  - Импликация: смягчить — частичная страховка/инсуренс через кредиты, либо
    «escape pod» механика, при которой возвращается X% флота.
- **Sensor Phalanx и тайминг до секунды** — атакующий с фалангой видит
  возврат флота посекундно и бьёт в окно «между приземлением и взлётом».
  _Источник: OGame Wiki Fleetsaving, ogametips.com._
  - Импликация: рандомизация тайминга прибытия (jitter ±5–15s); либо
    скрытый интервал «выгрузки», в течение которого фалангу не видно.
- **Moon shot RNG** — макс. шанс 20% за 2M debris; «можно сделать дюжины
  попыток без луны» — фундаментально несправедливый барьер прогресса.
  _Источник: OGame Wiki Moonchance, board.us.ogame thread 102451._
  - Импликация: pity-таймер (накопительная вероятность) или
    deterministic-cost альтернатива (накопить N debris → гарантированная
    луна).
- **ACS/координация — клуб для богатых** — синхронизированные удары
  альянсов недоступны соло-игрокам и казуалам; «без альянса ты мясо».
  _Источник: общий тон форумов, board guidelines._
  - Импликация: матчмейкинг для соло-PvP; раунд-режимы, не зависящие от
    группового онлайна.
- **Bashing rules легко эксплуатируются** — лимит 6 атак/24ч обходится
  через несколько аккаунтов альянса; декларация войны — формальность.
  _Источник: OGame Wiki Rules, форумные дискуссии о пушинге._
  - Импликация: bashing-protection считается на уровне атакуемой цели
    (не на пару attacker→target), включает урон, а не только число атак.
- **Debris field — победитель забирает почти всё** — defender теряет 100%
  юнитов, attacker может фармить 30% обломков — поощряет «рейд за лутом»,
  а не «защити дом». _Источник: OGame Wiki Debris Field._
  - Импликация: пересмотреть debris ratio в пользу защитника или ввести
    «aftermath tax» с обломков для соседей.

## Монетизация

- **Officers — обязательная подписка, не косметика** — Commander, Admiral
  и пр. дают очередь построек, +флот-слоты, +2% производства; «без них
  ты буквально слабее». _Источник: board.us.ogame thread 96386, OGame Wiki
  Dark Matter._
  - Импликация: все pay-офицеры заменяются on-grind разлочками; платный
    bundle — только косметика/удобство (темы, сорт-фильтры).
- **Instant-complete за DM — pure P2W** — «абсолютно вопиющие моментальные
  завершения за тёмную материю». _Источник: board.us.ogame thread 97887
  «Absolutely preposterous instant completion»._
  - Импликация: либо нет insta-complete вовсе, либо жёсткий cooldown +
    cap «N раз в сутки», доступный и за in-game валюту.
- **Бесконечный цикл трат** — «опытные друзья ушли, чтобы не быть в
  бесконечном цикле платежей». _Источник: board.en.ogame thread 843056._
  - Импликация: проектируем монетизацию по принципу «battle pass с
    потолком» (фикс. сумма за сезон), а не whale-spending без верха.
- **DM packs резко скейлятся со складом** — «более дорогие пакеты стоят
  сильно непропорционально дешевле — fairness?». _Источник: board.en.ogame
  thread 847189 «Dark Matter cost of resource packages»._
  - Импликация: прозрачное линейное ценообразование пакетов; никаких
    скрытых множителей «pay more — pay less per unit».
- **Аккаунт банят — DM сгорает** — жалобы «забанен без причины, DM и
  прогресс пропали». _Источник: sikayetvar.com / Xolvie review «OGame
  Account Banned Without Reason»._
  - Импликация: при бане — refund/отложенное удаление DM-эквивалента;
    прозрачный appeal-процесс с SLA.

## Социалка и альянсы

- **Альянс-драма / dual accounts топов** — «top-альянс пушит мульти-аккаунты,
  модераторы дают только варн». _Источник: Trustpilot ogame.de отзывы._
  - Импликация: тех. detection мультов (fingerprint, поведение), публичная
    статистика наказаний, automatic action а не «варн от GM по
    настроению».
- **Trolling в чате/сообщениях** — формальные правила есть, но GM-отклик
  медленный; жертвы «терпят и репортят». _Источник: Board Guidelines, форум._
  - Импликация: client-side mute/block + рейт-лимит сообщений новичкам;
    mod-tools с прозрачными action-логами.
- **Pushing/трейд-эксплойт между друзьями** — «передал 100M
  младшему, тот скакнул в топ» — рамки правил размыты.
  _Источник: OGame Wiki Rules, форумные споры._
  - Импликация: формальный лимит трансфера на уровень очков получателя;
    автодетект (anti-pushing rule) с явным расчётом.
- **Внешняя коммуникация (Discord) обязательна** — внутриигровой UI чата
  настолько слаб, что серьёзные альянсы переехали в Discord/TS.
  _Источник: board.us.ogame thread 97153 «Official OGame Discord»._
  - Импликация: встроенный voice/chat не нужен, но альянс-doc, op-board,
    fleet-coordination tools должны быть в самой игре.
- **Decay альянсов = decay сервера** — когда топ-альянс уходит, сервер
  «пустеет»; нет механики переезда/слияния. _Источник: общий тон._
  - Импликация: cross-server merge events; альянс-skins/legacy badges,
    стимулирующие миграцию в новые сезоны вместо ухода.

## UX / UI

- **Интерфейс «clunky», устаревший** — даже игроки-фаны называют UI
  «неуклюжим», требующим множество кликов на типичную операцию.
  _Источник: ogame.life blog, общий тон обзоров._
  - Импликация: keyboard shortcuts, bulk-actions, command-palette
    (Ctrl+K), shareable URL со state галактики.
- **Мобайл — «boyy in the world pain»** — «делать ресурс-трансферы с
  телефона — самая большая боль в мире». _Источник: ogame.ninja
  changelog, форумные жалобы._
  - Импликация: mobile-first layout галактики и флота; PWA с offline-
    кешированием; push-уведомления о входящем флоте.
- **Боты и сторонние тулзы — всем известны, не банят** — «топ-10 игроков
  пользуются ботами, тикеты не работают». _Источник: github.com
  ogame-bots topic, форум board.en.ogame thread 813708 «How to legally
  automate»._
  - Импликация: либо встроить «легальный автоматизм» (idle-rules,
    auto-fleetsave) для всех, либо жёсткий server-side detection
    (статистика action-time distribution).
- **Notifications/alerts только через third-party** — у легитимного игрока
  нет встроенного «вас атакуют через 4ч» — нужно качать сторонние app.
  _Источник: ogame.ninja, форум._
  - Импликация: in-app + push + email уведомления о входящих, fleet
    return, expedition complete; на уровне аккаунта, не device.
- **Отсутствие нормального replay/боевой лог** — combat-report — текстовая
  таблица, нет timeline-визуализации раундов.
  _Источник: общий тон обзоров; OGame Wiki Combat._
  - Импликация: интерактивный battle replay (раунды, прицеливание),
    шарабельный по URL.
- **Slow page transitions / SPA не SPA** — каждый клик = full reload.
  _Источник: общий тон обзоров (Steemit review by enjar)._
  - Импликация: SPA с предзагрузкой соседних экранов; WebSocket для
    real-time таймеров без F5.

## Endgame / retention

- **Топ забетонирован — никто не догонит** — игроки в топе годами, новичкам
  «никогда не подняться». _Источник: ogame.life late-game guides, форум._
  - Импликация: rolling seasons / leagues с relegation; «престиж» механика,
    конвертирующая legacy в стартовый бонус нового сезона.
- **Боль конца игры — нечего делать** — после max research/fleet «остаётся
  фармить inactive accounts по кругу». _Источник: общий тон форумов, обзоры._
  - Импликация: endgame-контент (бессы, raid-target NPC, межгалактические
    события) с обновляемыми целями.
- **Active player count тихо снижается** — «~15k DAU, нишевая аудитория».
  _Источник: mmostats.com/game/ogame, mmommorpg.com._
  - Импликация: проектируем под 1–5k DAU реалистично; auto-scaling
    инфраструктуры под сезонные пики.
- **«Вернулся через год — отстал навсегда»** — нет механики catch-up; даже
  2 недели простоя = «всё, ты больше не конкурентен».
  _Источник: общий тон форумов «why I quit»._
  - Импликация: catch-up XP/research-bonus для returning players, чтобы
    отставание было быстро компенсируемо.
- **Vacation mode — половинчатое решение** — на VM нельзя строить, очки
  замораживаются, но соседи не замораживаются.
  _Источник: OGame Wiki Inactive Players, Vacation Mode._
  - Импликация: VM с накопительным «отложенным производством» (по cap), не
    нулевым ростом; cooldown между активациями ясно прописан.

## Админы / модерация / античит

- **Слабая enforcement ботов** — «detection ручной, GM смотрят логи
  глазами; топовые боты не банятся». _Источник: OGame Wiki Bot, форум
  thread 813708._
  - Импликация: server-side телеметрия action-distribution; ML-флаги +
    auditable rule-engine; публичный бан-репорт.
- **Поддержка платная/медленная** — «без оплаты тех. поддержки нет»,
  «48ч без ответа — норма». _Источник: Trustpilot ogame.de, board
  ticketing rules._
  - Импликация: первый-уровень саппорт через self-serve KB + community
    helpers; 48h SLA публичный, escalation-policy прозрачная.
- **Неравное наказание (Top vs noob)** — «топу варн, новичку перм-бан за
  то же». _Источник: Trustpilot, форумные жалобы на dual-account
  enforcement._
  - Импликация: единый rule-engine с одинаковыми санкциями; appeal
    review независимый от GM, который выписал бан.
- **Бан без объяснения, потеря покупок** — «забанен без причины, DM и
  прогресс пропали». _Источник: sikayetvar.com Xolvie review._
  - Импликация: каждый бан — с указанием правила и evidence-снапшота;
    refund unspent premium currency при подтверждённой ошибке.
- **GM конфликт интересов** — модераторы — игроки тех же вселенных.
  _Источник: общий тон форумных споров._
  - Импликация: GM не играет на сервере, который модерирует; ротация GM
    между регионами; публичный конфликт-disclosure.

## Источники

- https://board.en.ogame.gameforge.com/index.php?thread/843056-pay-to-win-ogame-yin-new-universe/
- https://board.en.ogame.gameforge.com/index.php?thread/830971-ogame-is-not-as-p2w-as-u-think/
- https://board.en.ogame.gameforge.com/index.php?thread/817500-pay-to-win/
- https://board.en.ogame.gameforge.com/index.php?thread/847189-dark-matter-cost-of-resource-packages/
- https://board.en.ogame.gameforge.com/index.php?thread/842864-updated-new-player-protection-system/
- https://board.en.ogame.gameforge.com/index.php?thread/816997-ways-of-handling-deut-consumption-for-fleeter/
- https://board.en.ogame.gameforge.com/index.php?thread/813708-how-to-legally-automate-ogame/
- https://board.en.ogame.gameforge.com/index.php?thread/810303-guide-04-fleetsaving-guide/
- https://board.en.ogame.gameforge.com/index.php?thread/838728-ogame-org-team/
- https://board.en.ogame.gameforge.com/index.php?thread/841580-board-guidelines/
- https://board.en.ogame.gameforge.com/index.php?thread/852579-ogame-org-game-rules-clarifications/
- https://board.us.ogame.gameforge.com/index.php?thread/96386-dark-matter-use/
- https://board.us.ogame.gameforge.com/index.php?thread/97887-absolutely-preposterous-instant-completion-for-dark-matter/
- https://board.us.ogame.gameforge.com/index.php?thread/103903-class-imbalance-discussion/
- https://board.us.ogame.gameforge.com/index.php?thread/98263-will-gameforge-ever-respond-to-the-insane-unbalance-between-classes/
- https://board.us.ogame.gameforge.com/index.php?thread/97251-discoverer-needs-some-adjusting/
- https://board.us.ogame.gameforge.com/index.php?thread/102451-eli5-40-moon-chance/
- https://board.us.ogame.gameforge.com/index.php?thread/105940-40-moon-events/
- https://board.us.ogame.gameforge.com/index.php?thread/97153-official-ogame-discord/
- https://board.origin.ogame.gameforge.com/index.php/Thread/2195-noob-protection-system/
- https://board.origin.ogame.gameforge.com/index.php/Thread/2247-Noob-protection-D/
- https://board.origin.ogame.gameforge.com/index.php/Thread/2288-Noob-protection-seriously-factual-discussion/
- https://board.origin.ogame.gameforge.com/index.php/Thread/11129-Make-Ogame-unplayable-on-Mobile-devices/
- https://board.origin.ogame.gameforge.com/index.php/Thread/10749-Using-Dark-Matter/
- https://board.origin.ogame.gameforge.com/index.php/Thread/7532-Deuterium-consumption/
- https://board.origin.ogame.gameforge.com/index.php/Thread/785-Guide-06-Fleetsaving-Guide/
- https://ogame.fandom.com/wiki/Dark_Matter
- https://ogame.fandom.com/wiki/Fleetsaving
- https://ogame.fandom.com/wiki/Moonchance
- https://ogame.fandom.com/wiki/Moonchance_Strategy
- https://ogame.fandom.com/wiki/Debris_Field
- https://ogame.fandom.com/wiki/Inactive_Players
- https://ogame.fandom.com/wiki/Rules
- https://ogame.fandom.com/wiki/Bot
- https://ogame.fandom.com/wiki/Talk:Newbie_Protection
- https://ogame.life/ogame/blog/ogames-best-strategies-for-late-game-success/
- https://ogame.life/ogame/blog/top-10-ogame-tips-every-player-should-know/
- https://ogame.life/ogame/blog/how-to-increase-your-ranking-in-ogame-fast/
- https://www.ogame.ninja/changelog
- https://steemit.com/review/@enjar/review-or-ogame
- https://www.mmommorpg.com/browser-games/ogame/
- https://mmostats.com/game/ogame
- https://www.trustpilot.com/review/ogame.de
- https://www.trustpilot.com/review/www.ogame.dk
- https://www.sikayetvar.com/en/ogame-us/ogame-account-banned-without-reason-lost-dark-matter-and-progress
- https://www.sikayetvar.com/en/gameforgecom-us
- https://github.com/topics/ogame-bots
