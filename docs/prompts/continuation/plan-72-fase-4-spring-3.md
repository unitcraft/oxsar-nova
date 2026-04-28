# Промпт: выполнить план 72 Ф.4 — Spring 3 (артефакты, info-экраны, tech-tree, статистика)

**Дата создания**: 2026-04-28
**План**: [docs/plans/72-remaster-origin-frontend-pixel-perfect.md](../../plans/72-remaster-origin-frontend-pixel-perfect.md)
**Зависимости**: ✅ Ф.1 Bootstrap, ✅ Ф.2 Spring 1, ✅ Ф.3 Spring 2,
✅ план 67 (alliance backend), ✅ план 68 (биржа артефактов backend),
✅ план 78 (раскладка). Параллелится с планами 80 (auth-cleanup) и
73 Ф.1 (screenshots) — разные папки.
**Объём**: ~1500-2500 строк TS + CSS + i18n, 1-2 коммита.

---

```
Задача: реализовать Spring 3 плана 72 — pixel-perfect клон 8
экранов origin-фронта (Artefacts, ArtefactMarket, ArtefactInfo,
BuildingInfo, UnitInfo, Techtree, Records, Statistics, Daily quests).

КОНТЕКСТ:

Ф.1 (Bootstrap, 54fabbdf46), Ф.2 (Spring 1, 47d1f0ef65), Ф.3
(Spring 2, 48ef07cf19+590a68b428) закрыты. Каркас в
projects/game-nova/frontends/origin/ работает: router + 7 главных
+ 22 alliance/resource/market экранов + i18n + auth-store + theme.css.

Spring 3 — 8 экранов из docs/research/origin-vs-nova/origin-ui-
replication.md секции S-025..S-032:
- S-025 Artefacts (мои артефакты — карточки с количеством, описанием,
  кнопками activate/use/sell)
- S-026 ArtefactMarket (артефактный market, EXT_MODE = продажа за
  credit; см. internal/market в game-nova; ЭТО НЕ план 68, это
  старая legacy-механика)
- S-027 ArtefactInfo (страница описания одного артефакта — что делает,
  redirect-цены, требования)
- S-028 BuildingInfo (страница описания одного здания — формула
  стоимости, что даёт)
- S-029 UnitInfo (страница описания одного юнита — статы, rapidfire,
  стоимость)
- S-030 Techtree (граф технологий — визуальное древо зависимостей)
- S-031 Records (рекорды — top-N игроков по разным критериям)
- S-032 Statistics (агрегированная статистика игры — суммарно
  игроков, юнитов, очков)

Daily quests (если есть в legacy и упомянут как S-033 в
origin-ui-replication) — добавь как ~9-й экран если поместится в
объём, иначе отложи на отдельную сессию.

ВАЖНО ПРО S-026 ArtefactMarket vs план 68 ExchangeScreen:

В плане 68 (биржа артефактов) и плане 76 (UI биржи в nova) была
**новая** биржа — P2P-обмен артефактов за оксариты. Это **отдельная
фича**.

S-026 ArtefactMarket — это **legacy EXT_MODE market**: фиксированный
price-list, продажа артефактов системе за credit. См. миграцию
0013_artefact_market.sql и internal/market в game-nova-backend —
там уже есть GET /api/market/offers + POST /api/market/buy.

Это **две разные подсистемы** в backend, обе работают одновременно
в nova. Origin-фронт получит обе:
- S-026 ArtefactMarket — legacy-style, в Spring 3 (этой сессии).
- P2P-биржа артефактов из плана 68 — будет в Spring 5 (отдельная
  сессия, см. план 72 «Spring 5: Stock/Exchange»).

ВАЖНО ПРО S-025 vs план 67/77:

Артефакты — row-per-item (см. план 68 архитектурное уточнение),
state в (held/listed/active/expired/consumed). На S-025 показываем
только state='held' и state='active' (если артефакт уже использован
с TTL). 'listed' (escrow в биржу из плана 68) показывать **отдельной
группой** (если решишь — UI feedback пользователю что артефакт
висит на P2P-бирже; см. план 76 ExchangeScreen).

ПЕРЕД НАЧАЛОМ:

ПЕРВЫМ ДЕЙСТВИЕМ (до любого чтения плана):

1) git status --short. cat docs/active-sessions.md.

2) ОБЯЗАТЕЛЬНО добавь свою строку в раздел «Активные сессии»:
   | <N> | План 72 Ф.4 Spring 3 (8 экранов origin: artefacts/info/techtree/stats) | projects/game-nova/frontends/origin/ | <дата-время> | feat(origin/frontend): Ф.4 Spring 3 — artefacts + info + techtree + stats |

3) Параллельные сессии:
   - Slot 1 (план 80 auth-cleanup) — трогает deploy/, identity/.
     НЕ конфликтует с твоей frontends/origin/.
   - Slot 2 (план 73 Ф.1 baseline) — трогает tests/e2e/origin-baseline/.
     НЕ конфликтует.

ТОЛЬКО ПОСЛЕ — переходи к чтению:

4) Прочитай ПОЛНОСТЬЮ:
   - docs/plans/72-remaster-origin-frontend-pixel-perfect.md
   - docs/research/origin-vs-nova/origin-ui-replication.md секции
     S-025..S-033 (детальное описание)
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5» R0-R15
   - projects/game-nova/api/openapi.yaml — секции
     /api/artefacts (S-025/S-027), /api/market/* (S-026, legacy
     market), /api/buildings/* для info (S-028), /api/units/* для
     info (S-029), /api/research/tree (S-030), /api/records и
     /api/statistics (S-031/S-032). Если каких-то нет — отметь в
     simplifications.md как P72.S3.* и используй mock + TODO.

5) Прочитай выборочно:
   - projects/game-nova/frontends/origin/src/main.tsx (роутер)
   - projects/game-nova/frontends/origin/src/api/ (API-инфра Spring
     1+2)
   - Один из существующих экранов Spring 1 как эталон:
     projects/game-nova/frontends/origin/src/features/main/MainScreen.tsx
   - projects/game-legacy-php/templates/ — *.tpl для каждого S-NNN
     экрана (artefacts.tpl, artefactmarket.tpl, artefactinfo.tpl,
     buildinginfo.tpl, unitinfo.tpl, techtree.tpl, records.tpl,
     statistics.tpl, dailyquests.tpl).

ЧТО НУЖНО СДЕЛАТЬ:

### Архитектурно

1. Routes в роутере:
   - /artefacts → ArtefactsScreen (S-025)
   - /artefact-market → ArtefactMarketScreen (S-026)
   - /artefact/:id → ArtefactInfoScreen (S-027)
   - /building/:type → BuildingInfoScreen (S-028)
   - /unit/:type → UnitInfoScreen (S-029)
   - /techtree → TechtreeScreen (S-030)
   - /records → RecordsScreen (S-031)
   - /statistics → StatisticsScreen (S-032)
   - /daily-quests → DailyQuestsScreen (S-033, если есть в legacy)

2. API-модули в src/api/:
   - artefacts.ts (list, getOne, activate, use)
   - market.ts (offers list, buy) — это LEGACY market, не план 68
     биржа
   - building-info.ts (по существующему GET /api/buildings/:type)
   - unit-info.ts (по существующему GET /api/units/:type)
   - techtree.ts (GET /api/research/tree)
   - records.ts (GET /api/records?type=points/army/economy/etc)
   - statistics.ts (GET /api/statistics — агрегаты по игре)

3. Каждый экран — папка `src/features/<name>/`, главный файл
   `<Name>Screen.tsx` + связанные хуки/типы.

4. **Pixel-perfect**: HTML+CSS-классы зеркалят legacy *.tpl. Точная
   визуальная сверка — план 73 (snapshot Ф.1+Ф.2 параллельно).

5. **i18n (R12)**: grep configs/i18n/{ru,en}.yml на artefact|building|
   unit|tech|record|statistic. Цель ≥95% переиспользования.

6. **Тесты** (vitest + RTL): 1-2 теста на каждый экран. Spring 3
   = 8 экранов = 10-15 тестов минимум.

7. **Idempotency-Key (R9)** на mutation'ах (activate/use/buy).

### Экран за экраном

#### S-025 Artefacts

- TanStack Query useQuery(['artefacts', 'me']) → GET /api/artefacts.
- Группировка: by_state (held / active-with-TTL / listed-on-exchange)
  + by_type (cosmetic/strategic/...).
- Каждая карточка: имя (i18n), description, quantity, действия:
  - «Активировать» (если применимо) → POST /api/artefacts/{id}/activate
  - «Использовать» (если разовое) → POST /api/artefacts/{id}/use
  - «Продать» (legacy market) → navigate /artefact-market?artefact_id=X
  - «Выставить на биржу» (план 68) → navigate /exchange/new?artefact_id=X
- Idempotency-Key на activate/use.

#### S-026 ArtefactMarket (legacy EXT_MODE)

- TanStack Query useInfiniteQuery(['market', 'offers']) → GET
  /api/market/offers.
- Filters: тип артефакта, диапазон цен (credit / oxsariты).
- Карточки: имя артефакта, цена в credit, описание, кнопка «Купить»
  (POST /api/market/buy с Idempotency-Key).
- Список собственных «продаж» (если legacy market поддерживает) —
  опционально.

#### S-027 ArtefactInfo

- /artefact/:id или /artefact/:type — статическая страница
  описания одного артефакта (не «мои», а каталог).
- TanStack Query → GET /api/artefacts/catalog/:type.
- Показывает: имя, описание (длинный текст из i18n), эффект,
  redirect-цены (если применимо в legacy), требования (если есть).

#### S-028 BuildingInfo

- /building/:type — описание здания.
- TanStack Query → GET /api/buildings/catalog/:type.
- Показывает: имя, описание, формула стоимости (металл/кристалл/
  водород по уровню), что даёт (производство, защита, и т.д.).

#### S-029 UnitInfo

- /unit/:type — описание юнита.
- TanStack Query → GET /api/units/catalog/:type.
- Статы: атака, броня, щиты, скорость, груз, rapidfire против
  других юнитов (таблица), стоимость, требования.

#### S-030 Techtree

- /techtree — визуальный граф технологий с зависимостями.
- TanStack Query → GET /api/research/tree.
- Может быть SVG / canvas / простой grid с линиями. Минимум —
  табличное представление с indentation для родитель/дочерний.
- Текущий уровень исследования (из контекста игрока) — подсветка.
- Клик по тех → /research?type=X (Spring 1 экран).

#### S-031 Records

- /records — top-N (обычно 100) по разным критериям:
  - очки (общий ranking)
  - флот
  - экономика
  - бой (kill/death)
  - и т.д.
- TanStack Query → GET /api/records?type=X&limit=100.
- Tabs или select для переключения между критериями.

#### S-032 Statistics

- /statistics — агрегированная статистика игры:
  - всего игроков, активных за 7 дней
  - всего планет, юнитов в игре
  - суммарные ресурсы (если backend отдаёт)
  - кол-во сражений за последние 24h
  - и т.д.
- TanStack Query → GET /api/statistics (агрегаты, кешируется).

#### S-033 Daily quests (опционально)

Если в legacy templates/dailyquests.tpl или подобный есть — реализуй
9-й экран по тому же паттерну. Если нет в legacy — пропусти и
отметь в simplifications.md «P72.S3.X: Daily quests не реализованы
в legacy, отложено».

### Финализация Ф.4

- Шапка плана 72: Ф.4 ✅, Spring 3 закрыт.
- НЕ закрываешь весь план 72 — впереди Ф.5 (Spring 4), Ф.6 (Spring 5),
  Ф.7 (i18n рус), Ф.8 (TipTap чат), Ф.9 (финал).
- Запись итерации в docs/project-creation.txt («72 Ф.4 — Spring 3»).
- В коммите указать i18n переиспользовано/новых.

══════════════════════════════════════════════════════════════════
ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА R0-R15 — см. блок rules-r0-r15.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/rules-r0-r15.md

Особо важно:
- R0: backend nova не меняем.
- R5: pixel-perfect для origin-фронта.
- R9: Idempotency-Key на activate/use/buy.
- R12: i18n grep сначала.
══════════════════════════════════════════════════════════════════

══════════════════════════════════════════════════════════════════
GIT-ИЗОЛЯЦИЯ ОТ ПАРАЛЛЕЛЬНЫХ СЕССИЙ — см. блок git-isolation.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/git-isolation.md

Свои пути для CC_AGENT_PATHS:
- projects/game-nova/frontends/origin/

  (вся папка целиком — твой territory; параллельные планы 80
   (deploy/) и 73 Ф.1 (tests/e2e/origin-baseline/) НЕ пересекаются)

- docs/plans/72-remaster-origin-frontend-pixel-perfect.md
- docs/active-sessions.md
- docs/project-creation.txt
- docs/simplifications.md (если есть новые P72.S3.* записи)

ВНИМАНИЕ: НЕ трогай:
- projects/game-nova/frontends/nova/ — там работает nova-фронт.
- projects/game-nova/backend/ — backend закрыт.
- projects/game-nova/api/openapi.yaml — расширения отметить в
  simplifications.md.
- deploy/ — план 80 параллельно.
- tests/e2e/origin-baseline/ — план 73 Ф.1 параллельно.
══════════════════════════════════════════════════════════════════

КОММИТЫ:

1-2 коммита:

1) feat(origin/frontend): Ф.4 Spring 3 — artefacts + info + techtree
   + stats (план 72)

ИЛИ если объём > 2500 строк:
1) feat(origin/frontend): Ф.4 Spring 3 ч.1 — artefacts + market +
   info-страницы (5 экранов)
2) feat(origin/frontend): Ф.4 Spring 3 ч.2 — techtree + records +
   statistics + финализация

Trailer: Generated-with: Claude Code
ВСЕГДА с двойным тире:
git commit -m "..." -- $CC_AGENT_PATHS

ЧЕГО НЕ ДЕЛАТЬ:

- НЕ менять nova-фронт.
- НЕ менять backend / openapi.yaml.
- НЕ путать ArtefactMarket (legacy EXT_MODE) с ExchangeScreen
  (план 68 P2P биржа). Это две разные подсистемы.
- НЕ переносить рекламу/баннеры из legacy.
- НЕ закрывать весь план 72 — Ф.4 это только Spring 3.
- НЕ забывай Idempotency-Key.
- НЕ забывай про -- в git commit.

УСПЕШНЫЙ ИСХОД:

- 8-9 экранов работают (router → каждый рендерится).
- typecheck + build + tests зелёные.
- Все экраны pixel-perfect клоны legacy.
- i18n: 95%+ переиспользования.
- Шапка плана 72: Ф.4 ✅ (Spring 3).
- Запись в docs/project-creation.txt.
- Удалена строка из docs/active-sessions.md.

Стартуй.
```
