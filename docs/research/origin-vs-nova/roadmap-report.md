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

- **Backend в nova**: ~12-16 недель (legacy.yaml, новые модули,
  расширение event-loop, alien AI до полного, биржа, телепор,
  3 описания альянса, гранулярные ранги, global mail, и т.п.)
- **Frontend (origin pixel-perfect клон)**: ~12-16 недель (55
  экранов на React + воссоздание layout/themes из game-origin)
- **CI / тестирование (screenshot-diff)**: ~2-3 недели
- **Deploy / DNS / config (legacy01 universe)**: ~1 неделя

**Итого**: 27-36 недель (6-9 месяцев) команды из 1-2 разработчиков.
Часть фронта и бэка можно параллелить.

---

## Часть II. Декомпозиция в будущие планы

Нумерация ориентировочная — согласовать с текущим состоянием
docs/plans на момент старта.

### План 64: legacy.yaml + per-universe balance loading

**Что**: параметризация балансовых констант nova под профили
`modern` / `legacy`. Все 🟡 расхождения D-NNN из категории формула.

**Содержит**:
- `configs/balance/legacy.yaml` — числовые предвычисленные значения
  charge_*, basic_*, prod_* для всех buildings/units/research/defense
- Парсер origin-формул в Go (для предвычисления)
- Расширение `internal/config/` для загрузки per-universe профиля
- Поле `universes.balance_profile` ('modern'|'legacy')

**Закрывает**: D-026, D-027 (RF алиенов в legacy.yaml), D-028
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
**Зависит от**: 63 (legacy.yaml — для balance numbers).

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
- Четверг-множитель ×5 / ×1.5..2.0 (вынос в legacy.yaml)
- `findTarget` / `findCreditTarget` с критериями
- `shuffleKeyValues` (случайное ослабление техник)
- Платный выкуп удержания

**Закрывает**: D-036, alien-ai-comparison.md A1-A14.

**Объём**: 3 недели.
**Зависит от**: 63 (alien-units в legacy.yaml).

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
- `users.account_deletion_scheduled_at` (D-003 для legacy01)
- `users.ui_theme`, `ui_pack` (D-007)

**Закрывает**: D-001, D-003 (для legacy), D-004, D-005, D-007,
D-008, D-019, D-020, D-021.

**Объём**: 2 недели (миграции + handler updates).

---

### План 70: Achievements расширение (legacy + общий движок)

**Что**: расширить goal engine под legacy-ачивки.

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
вселенных, не только legacy01).

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

### План 74: legacy01 deploy + DNS + config

**Что**: подъём legacy-вселенной как третей рядом с uni01/uni02.

**Содержит**:
- DNS / поддомен (имя по ADR-0010 — открытый вопрос)
- Свой Vite-bundle deploy (CDN)
- CORS / `ALLOWED_ORIGINS` расширение
- `universes.code = 'legacy01'`, `balance_profile = 'legacy'`
- Регистрация в registry-системе (план 36)
- Smoke-тесты после деплоя

**Закрывает**: запуск ремастера в проде.

**Объём**: 1 неделя.
**Зависит от**: 72, 73.

---

## Часть III. Зависимости между планами

```
64 (legacy.yaml) ──┬──→ 65 (event-loop)
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
| Точность баланса при предвычислении формул origin → legacy.yaml | Golden-tests на 5+ ключевых уровней каждого здания, сравнение с PHP `eval()` через скрипт |
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
| `templatepackage` per-user тёмные темы для legacy01 | `users.ui_theme` enum, не свободная строка | UGC-мрак |
| Турниры (D-031, U-002) в первой итерации | отдельный план **после** плана 74 | Не блокирует ремастер |
| Кастомные logo альянса (U-011) | отдельный план после storage/moderation | UGC требует инфраструктуры |
| EditConstruction / EditUnit / TestAlienAI | переход в admin-frontend (план 53) | dev/admin only |

---

## Часть VI. Известные неизвестные (требуют ещё одного round'а)

1. **Имя legacy-вселенной** — `legacy01` ли (по convention) или
   что-то иное (oxsar2 / classic / origin-skin?). Решение в
   ADR-0010 (открытый).
2. **Точные цвета палитры origin** — взять из `style.css`
   программно или вручную. План 72 решит.
3. **Шрифты в legacy** — какие именно, лицензии. План 72 решит
   аудитом.
4. **Какие именно из 100+ ачивок origin переносить** — нужен
   отбор приоритетных vs «исторических». План 70 решит.
5. **Поведение balance_profile в multi-universe DB** — каждая
   вселенная в одной DB или отдельные shard'ы? Уточнить с
   архитектором перед планом 64.
6. **Лицензии иконок origin** — критичный блокер. Аудит до плана 72.
7. **Чат фан-аут с TipTap-payload** — план 32 готов, но
   протестировать на ~100 одновременных пользователей до плана 72.

---

## Часть VII. Сводный график (Gantt-стиль)

```
Месяц 1-2:   [64 legacy.yaml]──┐
                                │
Месяц 2-3:   [65 event-loop]   │  [67 alliance]   [68 exchange]
                                │
Месяц 3-4:   [66 AlienAI]──────┤  [70 achievements]
                                │
Месяц 4-5:   [69 domain fields]┘  [71 UX-микрологика]
                                
Месяц 5-9:   [72 origin-фронт — pixel-perfect клон]
              (5 spring'ов по 3-4 недели каждый)
                                
Месяц 9-10:  [73 CI screenshot-diff]
                                
Месяц 10:    [74 legacy01 deploy]
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
| D-003 (account deletion) | домен | 68 (legacy01 ветка) |
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
