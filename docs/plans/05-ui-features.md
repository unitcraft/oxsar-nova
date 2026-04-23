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
Дублирование кода и UI.

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

### E.2 Галактика — P2/P3 (план 15, приоритет: LOW)

Реализованные части: альянс-теги, активность, статусы, тултипы, moon diameter/temp.

Отложенные:
- **Star Surveillance**: мониторинг систем (подписка на оповещения об активности)
- **Атака ракетами из галактики**: launch → coordinates без отдельного флота
- **Водород для полётов в галактике**: fuel consumption display
- **Экспедиции из галактики**: кнопка "Отправить экспедицию" из CellView

---

### E.3 Реальное время постройки в UnitInfo (план 13-real-build-time, приоритет: LOW)

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
