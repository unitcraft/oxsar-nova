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

### ✅ E.1 Tutorial → Achievement объединение

- `backend/internal/tutorial/` удалён
- `achievement/service.go`: `CheckAllStarter()` — стартовые ачивки (STARTER_BUILD_*)
- `AchievementsScreen.tsx`: фильтр по категориям starter/passive

---

### ✅ E.2 Начальный старт игрока

- `planet/starter.go`: `HomePlanetSize=18800`, начальные здания/исследования/флот
- Стартовые ресурсы Metal 1000 / Silicon 500

---

### E.3 Галактика — P2/P3 (план 15, приоритет: LOW)

Реализованные части: альянс-теги, активность, статусы, тултипы, moon diameter/temp.

Отложенные:
- **Star Surveillance**: мониторинг систем (подписка на оповещения об активности)
- **Атака ракетами из галактики**: launch → coordinates без отдельного флота
- **Водород для полётов в галактике**: fuel consumption display
- **Экспедиции из галактики**: кнопка "Отправить экспедицию" из CellView

---

### ✅ E.4 Реальное время постройки в UnitInfo (план 13, итерация ~50)
- `GET /api/planets/{id}/buildings` возвращает `build_seconds` для следующего уровня
- `UnitInfoScreen.tsx` показывает реальное время с учётом robotic/nano factory

---

### ✅ E.5 Vacation mode (режим отпуска) (итерация ~18)
- `users.vacation_since` в БД (миграция 0045)
- `auth/handler.go`: `POST /api/me/vacation`, `DELETE /api/me/vacation`
- `fleet/attack.go`: атака на игрока в отпуске → 403
- `fleet/send.go`: отправка флота в отпуске → 403
