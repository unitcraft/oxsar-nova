# Comparison: 8 категорий (Ф.3 плана 62)

**Дата**: 2026-04-28
**Контекст**: артефакт плана 62 — сводные таблицы по 8
категориям. Для каждой строки — ссылка на D-NNN журнал, S-NNN
UI-инвентарь, U-NNN/X-NNN UI-backlog или явная пометка ✅
«идентично, переиспользуем».

Цвета: 🟢 совпадает | 🟡 правка конфига | 🟠 фича origin
(нужно реализовать) | 🔴 формула расходится | 🟣 тихая семантика | 🔵 фича nova (отключаемо)

---

## Категория 1. Домен-модель (сущности и поля)

| Сущность | Origin | Nova | Цвет | D-NNN |
|---|---|---|---|---|
| User: dm_points/points/max_points | 9 полей очков | 6 полей | 🟣 | D-001 |
| User: vacation (umode/umodemin) | 2 INT поля | vacation_since/last_end | 🟣 | D-002 |
| User: account_deletion (delete) | INT timestamp | deletion_codes | 🔴 | D-003 |
| User: protection_time | INT(10) | ОТСУТСТВУЕТ | 🔴 | D-004 |
| User: observer | TINYINT(1) | role enum | 🔴 | D-005 |
| User: координаты (umi) | FLOAT кодированный | g/s/p раздельно | 🟣 | D-006 |
| User: UI customization | 6 полей (themes) | ОТСУТСТВУЕТ | 🔴 | D-007 |
| User: profession + prof_time | 2 поля | только profession | 🔴 | D-008 |
| User: activation/password tokens | 3 varbinary | в identity-svc | 🔴 | D-009 |
| User: last activity | INT(10) | timestamptz | 🟡 | D-010 |
| User: race | TINYINT(3) | ОТСУТСТВУЕТ | 🔴 | D-021 |
| User: chat tracking | last_chat/_ally | ОТСУТСТВУЮТ | 🔴 | D-020 |
| User: home planet (hp) | INT(10) | derived | 🟡 | D-019 |
| User: asteroid slots | INT(10) | ОТСУТСТВУЕТ | 🔴 | D-018 |
| User: planet_teleport_time | INT(11) | ОТСУТСТВУЕТ | 🟡 | D-016 |
| User: user_agreement_read | INT(10) | в identity-svc | 🟡 | D-025 |
| Planet: основные поля | 19 полей | 13 полей | 🟢 (mostly) | — |
| Building: уровни | na_building2planet (33 поля c factor) | buildings (2 поля) | 🟡 | D-022 |
| Building: формулы | DSL-строки в na_construction | YAML+Go | 🟣 | D-026 |
| Battle Report | 45 полей нормализ. | JSONB + 8 полей | 🟣 | D-011 |
| Espionage Report | embedded в events | отдельная таблица | 🟡 | D-012 |
| Alliance: ranks | битовые права | enum + rank_name | 🔴 | D-014 |
| Alliance: descriptions | 3 поля | 1 поле | 🟠 | D-041 |
| Achievement | 32 поля условий | 3 поля + goal engine | 🔴 | D-017 |
| Officer | юниты of_1..of_4 | подписка expires_at | 🔴 | D-015 |
| Event payload | MEDIUMBLOB (PHP serialize) | JSONB | 🟡 | D-023 |
| Event chains | parent_eventid/ally_eventid | в коде | 🟡 | D-024 |
| Event state-machine | 4 поля statuses | ENUM | 🟣 | D-013 |

**Итого по домену**: 28 строк, ~15 расхождений → 25 записей D-NNN
(D-001..D-025).

---

## Категория 2. Формулы и константы

| Формула | Origin | Nova | Совпадает? | D-NNN |
|---|---|---|---|---|
| Источник истины формул | DSL-строки в БД (`na_construction`) | YAML + Go-код | 🟣 фундаментально разное | D-026 |
| Стоимость зданий по уровням | charge_*: `pow(1.5, level-1)` etc | YAML cost_factor | 🟡 | D-030 |
| Время постройки | `basic_*` × формула | YAML time | 🟡 | — |
| Производство ресурсов | prod_* (с tech, level) | Go-формулы | 🟡 | — |
| Производство водорода (с {temp}) | `(-0.002*temp+1.28)` | без temp-mod | 🔴 | D-029 |
| Стоимость технологий | charge_* для mode=2 | YAML | 🟡 | — |
| Время исследований | формулы | YAML | 🟡 | — |
| Стоимость юнитов | basic_* для mode=3 | YAML | 🟡 | — |
| Время постройки флота | shipyard формулы | YAML | 🟡 | — |
| RF-таблица | na_rapidfire (с UNIT_A_*) | rapidfire.yml без алиен-юнитов | 🟠 | D-027 |
| Атака/щит/HP юнитов | na_ship_datasheet | ships.yml | 🟡 | — |
| Алиен-юниты UNIT_A_* | id 200-204 | ОТСУТСТВУЮТ | 🟠 | D-028 |
| Lancer/Shadow/etc спец-юниты | id 102, 325, 352, 353, 354, 355, 358 | частично | 🟠 | D-028 |
| Формула боя (раунды) | Java JAR Assault.jar | Go battle/engine.go | 🟡 нужна golden-проверка | — |
| Энергопотребление зданий | cons_energy формулы | Go | 🟡 | — |
| Лимит полей на планете | basic_size formula | YAML | 🟡 | — |
| Расход топлива | формулы Mission.class.php | Go fleet/economics | 🟡 | — |

**Итого**: 17 строк, 5 D-NNN (D-026..D-030). Основное —
архитектурное расхождение источников истины (D-026).

---

## Категория 3. Механики (фичи)

| Фича | Origin | Nova | Действие |
|---|---|---|---|
| Vacation Mode | ✅ na_user.umode | ✅ план 45 | 🟣 D-002 (семантика) |
| Fleet Slots (Computer Tech) | ✅ | ✅ | 🟢 |
| Миссия POSITION (перебазирование) | ✅ KindPosition реализован | ✅ | 🟢 |
| Сенсорная Фаланга | ✅ MonitorPlanet | ✅ /api/phalanx | 🟡 D-043 |
| Stargate | ✅ STARGATE_TRANSPORT/JUMP | план 20 Ф.5 не реал. | 🟠 |
| Уничтожение Луны | ✅ | ✅ план 20 Ф.6 | 🟢 |
| Astrophysics | ✅ | ✅ | 🟢 |
| IGR network | ✅ | ❓ верифицировать | — |
| Alien Holding (четверг) | полный AI 1127 стр | план 15 упрощён | 🔴 (см. alien-ai-comparison) |
| Achievements | 100+ ачивок | 5 базовых | 🔴 D-017 |
| Daily Quests | ❌ | ✅ план 56 | 🔵 |
| Категориальные рейтинги | u/r/b/e_points | u/r/b_points | 🟣 D-001 |
| Галактические события | ❌ | ✅ план 65 | 🔵 |
| Артефакты | ✅ полная система | ✅ | 🟢 |
| Двухвалютная схема (oxsar/oxsariт) | ✅ | план 58 ребранд | 🟡 |
| AI-советник | ❌ | ✅ план 47 LLM | 🔵 |
| Чат личный/альянс/глобальный | ✅ legacy BBCode | ✅ TipTap (план 32) | 🟢 (BBCode выкидывается) |
| Личные сообщения | ✅ | ✅ | 🟢 |
| Симулятор боя | ✅ Simulator.class.php | ✅ /api/battle-sim | 🟡 verify UI parity |
| Ремонт флота после боя | ✅ Repair.class.php | ✅ /api/repair | 🟡 |
| Биржа артефактов | ✅ Stock/Exchange | ❌ | 🟠 D-039 |
| Турниры | ❌ (зарезерв) | ❌ | 🟠 D-031 |
| Telepor планеты | ✅ EVENT_TELEPORT_PLANET | ❌ | 🟠 D-032 |
| TempPlanet (expires) | ✅ EVENT_TEMP_PLANET_DISAPEAR | ✅ KindExpirePlanet | 🟢 |
| Buddy-list | ✅ | ✅ | 🟢 |
| Профессии | ✅ + смена за кредиты | ✅ план 46 | 🟡 D-008 |
| Расы | ✅ | ❌ | 🟠 D-021 |
| Officer-юниты | ✅ боевые | подписка | 🔴 D-015 |
| Гранулярные права рангов | ✅ битовые | enum | 🔴 D-014 |
| Передача лидерства альянса | ✅ abandonAlly | ❌ | 🟠 D-040 |
| Global mail альянса | ✅ globalMail | ❌ (после 57) | 🟠 D-042 |
| ExchangeOpts (auto-exchange) | ✅ | ❌ | 🟡 D-045 |

**Итого**: 32 строки, основные расхождения уже отражены в D-NNN.

---

## Категория 4. Event-loop

| EVENT (origin) | Kind (nova) | Семантика | D-NNN |
|---|---|---|---|
| EVENT_BUILD_CONSTRUCTION | KindBuildConstruction (1) | 🟢 |
| EVENT_DEMOLISH_CONSTRUCTION | KindDemolishConstruction (2) объявлен, нет handler | 🟠 | — |
| EVENT_RESEARCH | KindResearch (3) | 🟢 |
| EVENT_BUILD_FLEET / DEFENSE | KindBuildFleet/Defense (4,5) | 🟢 |
| EVENT_REPAIR / DISASSEMBLE | KindRepair/Disassemble (50,51) | 🟢 |
| EVENT_POSITION | KindPosition (6) | 🟢 |
| EVENT_TRANSPORT / DELIVERY_* | KindTransport (7) + Delivery* объявлены | 🟡 | D-035 |
| EVENT_COLONIZE / RANDOM / NEW_USER | KindColonize (8) | 🟢 (с ветками) |
| EVENT_RECYCLING | KindRecycling (9) | 🟢 |
| EVENT_ATTACK_SINGLE / DESTROY_* | KindAttack* | 🟡 D-037 (DESTROY_BUILDING нет) |
| EVENT_ATTACK_ALLIANCE / DESTROY_* | KindAttackAlliance* | 🟡 |
| EVENT_SPY | KindSpy (11) | 🟢 |
| EVENT_HALT | KindHalt (13) объявлен, нет handler | 🟠 |
| EVENT_HOLDING | KindHolding (17) объявлен, нет handler | 🟠 |
| EVENT_RETURN | KindReturn (20) | 🟢 |
| EVENT_MOON_DESTRUCTION | KindMoonDestruction (14) объявлен, нет handler | 🟠 |
| EVENT_EXPEDITION | KindExpedition (15) | 🟢 (но 6 типов из 13) |
| EVENT_ROCKET_ATTACK | KindRocketAttack (16) | 🟢 |
| EVENT_STARGATE_TRANSPORT/JUMP | KindStargateTransport/Jump (28,32) объявлены | 🟠 |
| EVENT_ARTEFACT_EXPIRE | KindArtefactExpire (60) | 🟢 |
| EVENT_ARTEFACT_DELAY | KindArtefactDelay (63) | 🟢 |
| EVENT_ARTEFACT_DISAPPEAR | KindArtefactDisappear (61) объявлен, нет handler | 🟠 |
| EVENT_EXCH_EXPIRE | ❌ | 🟠 D-039 |
| EVENT_EXCH_BAN | ❌ | 🟠 D-039 |
| EVENT_TEMP_PLANET_DISAPEAR | KindExpirePlanet (65) | 🟢 D-033 (вер.) |
| EVENT_RUN_SIM_ASSAULT | ❌ | 🟠 D-034 |
| EVENT_TELEPORT_PLANET | ❌ | 🟠 D-032 |
| EVENT_TOURNAMENT_* | ❌ (зарезерв.) | 🟠 D-031 |
| EVENT_ALIEN_FLY_UNKNOWN | KindAlienFlyUnknown (33) объявлен, нет handler | 🟠 D-036 |
| EVENT_ALIEN_HOLDING | KindAlienHolding (34) | 🟢 |
| EVENT_ALIEN_ATTACK | KindAlienAttack (35) | 🟢 |
| EVENT_ALIEN_HALT | KindAlienHalt (36) | 🟢 |
| EVENT_ALIEN_GRAB_CREDIT | KindAlienGrabCredit (37), вложено в Attack | 🟡 D-036 |
| EVENT_ALIEN_ATTACK_CUSTOM | ❌ | 🟡 D-038 |
| EVENT_ALIEN_HOLDING_AI | KindAlienHoldingAI (80) — 2/8 действий | 🟡 |
| EVENT_ALIEN_CHANGE_MISSION_AI | KindAlienChangeMissionAI (81) объявлен, нет handler | 🟠 D-036 |

**Сводка**:
- 51 EVENT_* в origin, 24 реализованных Kind в nova + 8 объявленных
  без handler
- 🟢 совпадают: ~22
- 🟠 нужно реализовать в nova: ~17
- 🟡 семантические нюансы: ~6
- 🔵 nova-only (Daily Quest events, Score recalc и т.д.): 4

См. `alien-ai-comparison.md` для углублённого блока по AI.

---

## Категория 5. API-поверхность

Принцип (план 62): origin-фронт пишется на nova-API без
backend-адаптеров. Расхождения в путях имён НЕ в журнале.

| Действие | Origin (через ?go=) | Nova-API | Покрытие |
|---|---|---|---|
| Получить состояние планеты | Main, Resource | GET /api/planets/{id} | 🟢 |
| Список построек | Constructions | GET /api/planets/{id} | 🟢 |
| Поставить здание | Constructions::upgrade | POST /api/planets/{id}/buildings | 🟢 |
| Список технологий | Research | GET /api/research | 🟢 |
| Запустить исследование | Research::upgrade | POST /api/planets/{id}/research | 🟢 |
| Список юнитов | Shipyard | GET /api/planets/{id}/shipyard/inventory | 🟢 |
| Заказ флота | Shipyard::order | POST /api/planets/{id}/shipyard | 🟢 |
| Отправить флот (миссия) | Mission::sendFleet | GET/POST /api/fleet | 🟢 |
| Отозвать флот | Mission::retreatFleet | POST /api/fleet/{id}/recall | 🟢 |
| Боевой отчёт | Battlestats | GET /api/battle-reports/{id} | 🟢 |
| Шпионский отчёт | через MSG | GET /api/espionage-reports/{id} | 🟢 |
| Галактика | Galaxy | GET /api/galaxy/{g}/{s} | 🟢 |
| Список планет | Empire | GET /api/empire | 🟢 |
| Личные сообщения | MSG | /api/messages/* | 🟢 |
| Альянс — Create/List/Join/Leave | Alliance::found/apply/leave | /api/alliances/* | 🟢 |
| Альянс — три описания | updateAllyPrefs | ❌ | 🟠 D-041 |
| Альянс — передача лидерства | abandonAlly | ❌ | 🟠 D-040 |
| Альянс — гранулярные права | manageRanks | enum + rank_name | 🔴 D-014 |
| Альянс — global mail | globalMail | ❌ (после 57) | 🟠 D-042 |
| Альянс — диплом-статусы | diplomacy | partial /relations | 🟡 |
| Альянс — поиск с фильтрами | allySearch | /api/search | 🟡 (расширить) |
| Альянс — лог активности | ❌ (через msg?) | ❌ | 🟠 |
| Биржа — лоты, создание, покупка | Stock/StockNew | ❌ | 🟠 D-039 |
| Симулятор боя | Simulator | POST /api/battle-sim | 🟢 |
| Ремонт после боя | Repair | /api/repair/* | 🟢 |
| Phalanx | MonitorPlanet | GET /api/phalanx | 🟡 D-043 |
| Магазин/биллинг | Payment | /api/payment/* | 🟢 (план 38/42) |
| WebSocket / push | memcached polling | /ws/chat (план 32) | 🟢 |
| Resource transfers | ResTransferStats | ❌ endpoint | 🟡 D-044 |
| ExchangeOpts (auto-exchange) | ExchangeOpts | ❌ | 🟡 D-045 |

**Сводка**:
- 🟢 ~21 действий покрыты nova-API напрямую
- 🟡 ~5 нужно расширить (поля DTO / новые endpoint'ы)
- 🟠 ~6 отсутствуют — нужны новые endpoint'ы (биржа, передача
  лидерства, global mail и т.д.)
- 🔴 ~1 несовместимо концептуально (alliance ranks с битовыми
  правами)

---

## Категория 6. UI-экраны и UX-микрологика

См. отдельные артефакты:
- `origin-ui-replication.md` — все 55 контроллеров → S-001..S-055
- `nova-ui-backlog.md` — U-NNN (15 функций) и X-NNN (22 микрологики)

| Экран origin (S-NNN) | Nova-аналог | Цвет |
|---|---|---|
| Main, Constructions, Shipyard, Research, Galaxy, Mission, MSG, Friends, Artefacts, Battlestats, Ranking, Records, Resource, Market, Preferences, Profession, Officer, Search, Notepad, Empire, Repair (~21) | features/* | 🟢 |
| Alliance (12 шаблонов) | features/alliance | 🟡 (нужны добавления) |
| Achievement, Tutorial | features/* (упрощено) | 🟡 D-017 |
| Exchange/Stock/StockNew (3 контроллера биржи) | ❌ | 🟠 D-039 |
| Tournament | ❌ | 🟠 D-031 |
| RocketAttack | features/rockets | 🟢 |
| MonitorPlanet (фаланга) | features/galaxy с phalanx | 🟢 |
| ResTransferStats | ❌ | 🟡 D-044 |
| ExchangeOpts | ❌ | 🟡 D-045 |
| Payment (legacy 8 шаблонов) | features/payment (единый билинг) | ✅ выкидываем legacy |
| Moderator | admin-frontend (план 53) | 🟢 |
| EditConstruction / EditUnit / TestAlienAI | admin-only, не для прода клона | — |
| ChatPro | не используется (legacy зарезерв) | — |

UX-микрологика — все 22 X-NNN записи в `nova-ui-backlog.md`.

---

## Категория 7. Runtime-генерируемые ассеты

| Ассет | Origin | Nova | Статус |
|---|---|---|---|
| Изображение артефакта (composite) | `public/artefact-image.php` | ❌ | 🟠 D-046 |
| Аватар игрока | ❌ | ❌ | (опционально) |
| Карта галактики | client-side | client-side | 🟢 |
| Превью отчёта боя | embed в отчёте | embed | 🟢 |
| Логотип альянса | upload (см. U-011) | ❌ | 🟠 |
| Иконки построек по уровню | static PNG | static PNG | 🟢 |

Из `grep imagecreate|imagecopy|imagepng|imagejpeg|imagettftext`
найден **один** PHP-GD генератор: `artefact-image.php`. Остальные
ассеты статические.

---

## Категория 8. Cross-universe сервисы

Архитектурное решение принято: эта категория **не должна** порождать
расхождений. Если порождает — нарушение договорённости.

| Сервис | План | Используется в origin | Статус |
|---|---|---|---|
| Identity (JWT/JWKS, RBAC, users) | 51, 52 | ✅ как в nova | 🟢 |
| Billing (платежи, кошельки) | 38, 42, 54 | ✅ единый | 🟢 (legacy Payment.class.php выкидывается) |
| Portal (новости, предложения) | 36 | ✅ | 🟢 |
| Reports (жалобы, модерация) | 56 | ✅ | 🟢 |
| Mail-service | 57 | ✅ блокирует D-042 | 🟢 (после 57) |
| Admin-bff + admin-frontend | 53 | ✅ единая админка | 🟢 (Moderator.class.php выкидывается) |
| Реферальная программа | 59 | ✅ удалена в плане 60 из origin | 🟢 |
| Moderation blacklist | 46, 48 | ✅ общий YAML | 🟢 |
| Multi-instance / scheduler / chat fan-out | 32 | ✅ | 🟢 (BBCode выкидывается) |

**Итого**: 0 D-NNN записей в этой категории — архитектурная
договорённость соблюдается.

---

## Сводная заключительная таблица

| Категория | Совпадает 🟢 | Правка конфига 🟡 | Реализ. в nova 🟠 | Code-path 🔴 | Тихая 🟣 | Только nova 🔵 |
|---|---|---|---|---|---|---|
| 1. Домен | ~5 | 8 | 0 | 11 | 6 | 0 |
| 2. Формулы | 5 | 10 | 2 | 1 | 0 | 0 |
| 3. Механики | ~12 | 4 | 6 | 4 | 1 | 4 |
| 4. Event-loop | 22 | 6 | 17 | 0 | 0 | 4 |
| 5. API | 21 | 5 | 6 | 1 | 0 | 0 |
| 6. UI | 21 | 5 | 4 | 0 | 0 | 0 |
| 7. Assets | 4 | 0 | 2 | 0 | 0 | 0 |
| 8. Cross-uni | 9 | 0 | 0 | 0 | 0 | 0 |

**Общее ВСЕГО**: 99 элементов сравнения, 46 расхождений в журнале
D-NNN, 15+22=37 UI-кандидатов в `nova-ui-backlog.md`, 14 событий
alien-AI в `alien-ai-comparison.md`, 55 экранов в
`origin-ui-replication.md`.

---

## References

- [origin-inventory.md](origin-inventory.md)
- [nova-inventory.md](nova-inventory.md)
- [divergence-log.md](divergence-log.md)
- [nova-ui-backlog.md](nova-ui-backlog.md)
- [origin-ui-replication.md](origin-ui-replication.md)
- [alien-ai-comparison.md](alien-ai-comparison.md)
- [formula-dsl.md](formula-dsl.md)
