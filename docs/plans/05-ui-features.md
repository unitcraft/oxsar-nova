# План E: UI и игровые фичи

---

## История (завершено)

### ✅ UI обзор — основные экраны (план 04, итерации 20–48)
Все P1-экраны реализованы:
- ResourceScreen: факторы производства, слайдеры, пресеты, auto-save
- BuildingsScreen: очередь, требования, max_level, nano_factory, энергия
- ResearchScreen: очередь, требования, описания
- ShipyardScreen: очередь, отмена, rapidfire/боевые характеристики, описания
- RepairScreen: ремонт кораблей и defense
- GalaxyScreen: координаты, альянс, активность, статусы игроков, тултипы
- FleetScreen: отправка флота, все миссии, recall
- MessagesScreen: папки (личные/бой/шпионаж/экспедиции/система), compose, delete
- BattleReportsScreen, ExpeditionReportsScreen, EspionageReportsScreen
- AchievementsScreen с прогресс-барами
- MarketScreen, ArtmarketScreen
- ChatScreen (игровой чат)
- AllianceScreen: создание/вступление/управление, ранги, заявки, отношения
- TutorialScreen (шаги 1–7 с наградами)
- OfficersScreen: активация, auto-renew, взаимоисключающие группы
- AdminScreen: управление шаблонами автосообщений

### ✅ Планеты и колонизация (план 03, итерация ~28)
- Типы планет по позиции (trockenplanet, dschjungelplanet, normaltempplanet и др.)
- Temperature formula: base = 110 - pos×14, spread ±10
- Размер планеты по позиции: pos 1-3 малые, 4-12 стандарт, 13-15 большие
- API возвращает `planet_type`, `temperature_min/max`, `diameter`

### ✅ Ассеты луны (план 05, итерация ~28)
- `moon.jpg` (переименовано из `mond.jpg`), `moon-icon.svg`

### ✅ Оптимизация UX (план 08, итерация ~45)
Реализовано:
- Optimistic updates: Market, ResourceScreen, ArtefactsScreen, OfficersScreen
- Skeleton loaders для всех экранов
- Keyboard shortcuts: Alt+H/B/R/M, Esc, Ctrl+S
- Web Vitals metrics
- UpdateResourceFactors с CASE/WHEN (batch)
- Idempotency key для fleet/send, market/accept, artmarket/buy

### ✅ API-качество (план 10, итерация ~42)
- Idempotency-Key через Redis для POST-ендпоинтов
- Структурированные ошибки (httpx.ErrBadRequest / ErrForbidden / ErrInternal)
- OpenAPI spec обновлён

### ✅ Unit info (план 12, итерация ~40)
- Dedicated UnitInfoScreen (tab) вместо модальных окон
- Формулы производства, требования, описания для зданий/исследований/кораблей
- URL hash навигация (#building-1, #ship-31 и т.д.)

---

## Открытые задачи

### E.1 Tutorial → Achievement объединение (план 02, приоритет: MEDIUM)

**Проблема:** Две параллельные системы: `tutorial/` (квест-цепочка) и `achievement/` (пассивные).
Дублирование кода и UI. (В legacy-инстансе Dominator оба отключены: `ACHIEVEMENTS_ENABLED=false`,
`TUTORIAL_ENABLED=false` — но в nova они есть и нужны.)

**Шаг 1** — Миграция: `ALTER TABLE achievement_defs ADD COLUMN category TEXT DEFAULT 'passive'`

**Шаг 2** — `backend/internal/achievement/service.go`: добавить `CheckAllStarter()` —
триггеры для стартовых ачивок (постройка шахты, солнечной станции, лаборатории и т.д.)

**Шаг 3** — Seed-миграция: Starter-ачивки в `achievement_defs` с `category='starter'`
(STARTER_BUILD_METALMINE, STARTER_BUILD_SOLARPLANT и т.д., 6–7 шагов цепочки)

**Шаг 4** — Удалить `backend/internal/tutorial/` и роут `/api/tutorial`

**Шаг 5** — `frontend/src/features/tutorial/TutorialScreen.tsx` → удалить.
`AchievementsScreen.tsx` — добавить фильтр по категориям (starter / passive).

**Проверка готовности:**
- [ ] `achievement_defs.category` в БД
- [ ] `CheckAllStarter` в service.go
- [ ] Starter-ачивки в seed-миграции
- [ ] Пакет `tutorial/` удалён
- [ ] `AchievementsScreen` показывает стартовые + пассивные
- [ ] `make test` зелёный

---

### E.2 Начальный старт игрока (приоритет: MEDIUM)

**Контекст:** В legacy (Dominator) игрок стартует с готовым набором зданий, исследований и кораблей.
В nova неизвестно как реализован начальный старт — нужно проверить.

**Параметры из legacy `params.php` (Dominator):**

Начальные здания (`INITIAL_BUILDINGS`):
- Metal Mine 2, Silicon Lab 2, Hydrogen Lab 2, Solar Plant 4
- Robotic Factory 2, Shipyard 2, Research Lab 2
- Defense Factory 1, Repair Factory 1

Начальные исследования (`INITIAL_RESEARCHES`):
- Computer Tech 1, Energy Tech 1, Combustion Engine 2

Начальный флот (`INITIAL_UNITS`):
- 20 Small Transporter, 10 Light Fighter, 10 Recycler, 3 Colony Ship, 10 Espionage Probe

Стартовые ресурсы: Metal 1000, Silicon 500, Hydrogen 0.
Домашняя планета: 18 800 полей (`HOME_PLANET_SIZE`).

**Шаги:**
1. Проверить `backend/internal/auth/register.go` — создаёт ли начальные здания/исследования/флот
2. Если нет — добавить `InitialPlanetSetup(ctx, tx, userID, planetID)` с параметрами выше
3. Добавить `HOME_PLANET_SIZE` (18800 полей) в config

**Проверка готовности:**
- [ ] Новый игрок получает начальные здания, исследования, флот
- [ ] Домашняя планета имеет 18800 полей
- [ ] Тест регистрации нового игрока

---

### E.3 Галактика — P2/P3 (план 15, приоритет: LOW)

Реализованные части: альянс-теги, активность, статусы, тултипы, moon diameter/temp.

Отложенные:
- **Star Surveillance**: мониторинг систем (подписка на оповещения об активности)
- **Атака ракетами из галактики**: launch → coordinates без отдельного флота
- **Водород для полётов в галактике**: fuel consumption display
- **Экспедиции из галактики**: кнопка "Отправить экспедицию" из CellView

---

### E.4 Реальное время постройки в UnitInfo (план 13-real-build-time, приоритет: LOW)

`UnitInfoScreen` показывает базовое время постройки, не учитывая уровни
robotic_factory и nano_factory текущего игрока.

**Шаг 1** — `GET /api/planets/{id}/build-time?unit_id=N&level=M` — endpoint для расчёта
реального времени с учётом зданий планеты.

**Шаг 2** — `UnitInfoScreen.tsx`: показывать "Время постройки на [planetName]: Xч Yм"
(динамически, зависит от выбранной планеты).

**Проверка готовности:**
- [ ] Backend endpoint build-time
- [ ] Frontend использует реальное время из API
- [ ] Тест формулы BuildTime

---

### E.5 Vacation mode (режим отпуска) (приоритет: LOW)

**Контекст:** В legacy (consts.php):
- `VACATION_DISABLE_TIME = 30 дней` — через 30 дней без входа включается защита
- `LAST_TIME_ON_VACATION_DISABLE = 20 дней` — мин. интервал между отпусками
- Пока в отпуске: флот нельзя отправить, атака на игрока невозможна

В nova не реализована.

**Шаг 1** — Миграция: `ALTER TABLE users ADD COLUMN vacation_since TIMESTAMPTZ`
**Шаг 2** — `backend/internal/player/vacation.go`: `SetVacation / UnsetVacation` с проверкой интервала
**Шаг 3** — `fleet/attack.go`: проверка vacation у цели
**Шаг 4** — `FleetScreen.tsx`: кнопка "Режим отпуска" в ProfileScreen

**Проверка готовности:**
- [ ] `users.vacation_since` в БД
- [ ] `SetVacation` проверяет минимальный интервал
- [ ] Атака на игрока в отпуске возвращает 403
- [ ] `make test` зелёный
