# UI Development Log — oxsar-nova

Дневник UI-доработок, начатых после завершения ядра игры (бэкенд M1–M5,
battle engine, fleet, economy, chat). Предназначен как исходный материал
для будущей статьи о создании игры.

Структура: хронологические итерации с контекстом принятых решений,
техническими деталями и обоснованием выбора подхода.

---

## Контекст на момент начала UI-фазы (2026-04-22)

К этому моменту:
- **Бэкенд**: Go + PostgreSQL + Redis, ~36 таблиц, REST API + WebSocket
- **Ядро**: экономика, бой (порт oxsar2-java), флот, шпионаж, колонизация,
  рынок, артефакты, офицеры, ачивки, туториал, чат
- **Легаси**: oxsar2 (PHP/Yii 1.1) запущен локально на http://localhost:8080
  как эталон функционала и дизайна
- **Фронтенд**: React + TanStack Query + Zustand, Vite dev-сервер в Docker
- **Проблема**: UI работал но выглядел как "голый" React без дизайна —
  таблицы, стандартные браузерные стили, нет единого визуального языка

---

## Итерация UI-1: Дизайн-система и компоненты (2026-04-22)

### Задача
Создать единый визуальный язык для всего приложения, не зависящий
от UI-библиотек (Material UI, Ant Design и т.п.) — чтобы сохранить
полный контроль над стилями и не тащить лишний bundle.

### Дизайн-система (app.css)
Выбрана тёмная космическая тема: тёмно-синий фон (`#0a0c10`), циановый
акцент (`#63d9ff`), монспейс для чисел.

CSS-переменные:
- `--ox-bg`, `--ox-bg-panel`, `--ox-bg-hover` — иерархия фонов
- `--ox-accent`, `--ox-accent-dim` — основной акцент (cyan)
- `--ox-success`, `--ox-warning`, `--ox-danger` — статусные цвета
- `--ox-font` (Orbitron), `--ox-mono` (JetBrains Mono) — типографика
- `--ox-border`, `--ox-border-hover` — рамки с прозрачностью

Классы компонентов:
- `ox-panel` — карточка с фоном, border, box-shadow, border-radius
- `ox-tabs` — табы с `button[aria-pressed="true"]` вместо `.active`
  (aria-атрибут вместо класса — правильно семантически)
- `ox-badge`, `ox-alert`, `ox-skeleton` — вспомогательные элементы
- `ox-table-responsive` — таблица с горизонтальным скроллом на мобиле
- `ox-cards-grid` — CSS grid для карточек юнитов

### Общие UI-компоненты (src/ui/)
- **Modal** — overlay с анимацией появления
- **Toast** — уведомления (info/success/warning/danger), auto-dismiss 4s
- **ProgressBar** — с вариантами (default/success/warning/danger)
- **Confirm** — диалог подтверждения с `danger`-вариантом
- **Countdown** — живой таймер до даты (пересчёт каждую секунду)
- **ResourceTicker** — тикер ресурсов с интерполяцией по ratePerSec

### Навигация
Переработана с горизонтальных вкладок на вертикальный sidebar:
- Левая панель (160px) с группами: Планета / Флот / Социальное / Рынок
- Основной контент справа (1fr)
- На мобиле sidebar сворачивается (планируется гамбургер-меню)

**Решение**: sidebar лучше горизонтальных вкладок потому что пунктов
много (15+), и в горизонтальном варианте они не помещаются на экране.
Игры типа OGame традиционно используют вертикальную навигацию.

---

## Итерация UI-2: Переработка всех экранов (2026-04-22)

### Подход
Каждый экран переписан с нуля по принципу "card-first":
- Юниты (здания, корабли, оборонка, исследования) — карточки `ox-unit-card`
  с иконкой, названием, уровнем/количеством, стоимостью и кнопкой действия
- Очереди — прогресс-бары с Countdown
- Таблицы данных — `ox-table-responsive`

### Иконки юнитов
Взяты из legacy: `oxsar2/www/images/buildings/std/*.gif` → скопированы
в `frontend/public/images/units/`. 123 файла GIF из legacy skin.

Функция `imageOf(key)` в `catalog.ts` строит URL. Исключения через
`KEY_MAP`: `metal_mine` → `metalmine`, `missile_silo` → `rocket_station`.

### Исправление бесконечной загрузки
**Проблема**: при устаревшем JWT-токене в localStorage все API-запросы
возвращали 401, но приложение не выходило из состояния "Загрузка вселенной".

**Решение**: `api.get` теперь при 401 вызывает `useAuthStore.getState().logout()`,
что сбрасывает стор и показывает экран логина. Дополнительно — fallback
на `planets.isError` в App.tsx с кнопкой "Обновить".

### Замена window.confirm
Все `window.confirm()` заменены на компонент `<Confirm>` с `danger`-prop.
Пример: удаление сообщения в чате, роспуск альянса.

---

## Итерация UI-3: Страница Обзор — сравнение с legacy (2026-04-22)

### Методология
Открыт legacy oxsar2 (`http://localhost:8080/game.php/Main`), прочитан
шаблон `main.tpl` и контроллер `Main.class.php`. Составлен сравнительный
план: `docs/ui-plan-overview.md`.

### Что добавлено на страницу Обзор
**Было**: только список планет с ресурсами и очередями.

**Добавлено**:
1. **Флоты в пути** — список активных флотов с иконкой миссии, координатами
   цели, состоянием (→/←), таймером, прогресс-баром полёта и кнопкой
   "Отозвать" (только для outbound-флотов). Данные из `/api/fleet`.

2. **Характеристики планеты** — диаметр, поля (занято/макс), температура.
   Поля `diameter`, `used_fields`, `temp_min`, `temp_max` уже были в
   Go-модели и JSON-ответе, просто не использовались во фронтенде.

3. **Статистика игрока** — очки и место в рейтинге из `/api/highscore/me`.

4. **Уведомление о сообщениях** — баннер при наличии непрочитанных
   (`/api/messages/unread-count`).

### Что пока не добавлено (P3, требует нереализованных систем)
- Профессия игрока (система профессий не реализована)
- Уровень МГИ (межгалактических исследований через альянсы с артефактами)

---

## Итерация UI-4: Стабильность и UX (2026-04-22)

### Версионирование localStorage
**Проблема**: после каждой пересборки Docker-образа фронтенда у пользователей
мог оставаться устаревший стор Zustand в localStorage, что приводило к
чёрному экрану (crash до рендера).

**Решение**: поле `v: STORAGE_VERSION` в `oxsar-auth`. При несовпадении
версии — автоматический сброс. Текущая версия: `2`.

```typescript
// stores/auth.ts
const STORAGE_VERSION = 2;
// при загрузке: if (obj.v !== STORAGE_VERSION) { localStorage.removeItem(...) }
```

### Error Boundary
Добавлен `RootErrorBoundary` в `main.tsx` — класс-компонент React
(`getDerivedStateFromError`). При любом необработанном исключении в дереве
вместо чёрного экрана показывается:
- Сообщение об ошибке
- Кнопка "Сбросить кэш и перезагрузить" (очищает `localStorage` + `sessionStorage`)

### Autofill-стили
**Проблема**: Chrome/Safari при автозаполнении формы логина перекрашивали
поля в белый цвет, игнорируя CSS `background`.

**Решение**: стандартный trick через `-webkit-box-shadow: inset`:
```css
input:-webkit-autofill {
  -webkit-box-shadow: 0 0 0 100px rgba(5, 13, 24, 0.95) inset !important;
  -webkit-text-fill-color: var(--ox-fg) !important;
  transition: background-color 9999s ease-in-out 0s;
}
```

### Страница логина
Переработана: форма центрирована по экрану, ограничена шириной 400px,
обёрнута в `ox-panel`. Добавлены логотип (🚀 OXSAR) и placeholder-ы
в поля. Убрана зависимость от i18n (ключи `Registration.USERNAME` и т.п.
отсутствуют в словаре legacy — `t()` выдавал `[Registration.USERNAME]`).

---

## Технические решения, принятые в UI-фазе

| Решение | Альтернатива | Почему выбрано |
|---|---|---|
| Самописная дизайн-система | Material UI, Ant Design | Полный контроль стилей, нет лишнего bundle, тема "под космос" |
| `aria-pressed` для активных табов | класс `.active` | Семантика + CSS-селектор без JS |
| Docker для dev-сервера | `npm run dev` на хосте | Воспроизводимость среды, нет проблем с путями Windows |
| Версионирование `localStorage` | Очистка вручную | Автоматический сброс при несовместимом стейте |
| Хардкод текстов в LoginScreen | i18n словарь | Словарь legacy не содержит нужных ключей для формы входа |

---

## Итерация UI-5: P2/P3 Обзор + стоимости на карточках (2026-04-22)

### Обзор (P2)
- **Часы** в шапке: `useServerClock` хук, тикает раз в секунду через `setInterval`
- **Домашняя планета**: иконка 🏠 в `PlanetSwitcher`, правильные эмодзи 🪐/🌑
- **Очки игрока**: расширен `/api/highscore/me` — добавлен метод `PlayerScore` в сервис, хэндлер возвращает `points`
- **Картинки планет**: скопированы из legacy (`images/planets/*.jpg`), функция `planetImageOf(position, id)` в `catalog.ts` генерирует детерминированно по `PlanetPictures.xml`-логике без хранения в БД

### Обзор (P3)
- **Состав флота**: добавлены поля `ships` и `carry` в `FleetRow`, показываются под прогресс-баром

### Стоимости на карточках юнитов
- **catalog.ts**: `BuildingEntry` / `ResearchEntry` с `costBase` + `costFactor`, `CombatEntry` с `cost`, функция `costForLevel(base, factor, level)`; данные сверены с `configs/*.yml`
- **BuildingsScreen**: стоимость следующего уровня на каждой карточке, красный цвет если не хватает ресурсов
- **ResearchScreen**: аналогично
- **ShipyardScreen**: стоимость × количество, пересчитывается динамически при изменении поля ввода

---

## Итерация UI-6: GalaxyScreen — миссии и рейтинг (2026-04-22)

### Бэкенд
- `galaxy/repository.go`: добавлен `owner_rank` в `CellView` — подзапрос `COUNT()+1` аналогично `score.PlayerRank`

### Фронтенд
- **Своя планета**: подсвечивается голубым фоном, иконка 🏠, жирный шрифт
- **Рейтинг игрока**: `#N` рядом с именем в колонке Игрок
- **Кнопки миссий**: 🔭 шпионаж, ⚔️ атака, 📦 транспорт, ♻️ переработка (если есть обломки), 🌑🔭 шпионаж на луну — только для чужих планет
- **Переход на флот**: клик по кнопке миссии переключает таб на `fleet` и предзаполняет координаты + тип миссии через `fleetDst` state в `App.tsx`; `FleetScreen` принимает `initialDst?: InitialDst`

---

## Планируемые доработки UI (backlog)

Выполнено: Обзор P1/P2/P3, стоимости на карточках, GalaxyScreen с миссиями.

Остаток:
- Профессия игрока и МГИ (зависит от нереализованных систем)
- Карусель планет на Обзоре (P3, низкий приоритет)
- Сравнение с legacy: FleetScreen, MarketScreen, ScoreScreen, MessagesScreen
- AdminScreen — переработка
- BattleSimScreen — дизайн

Глобально:
- Мобильный sidebar (гамбургер-меню)
- Production build (сейчас dev-сервер в Docker)
- OpenAPI-генерация типов (`npm run gen:api`) вместо ручных типов в `types.ts`

---

## Итерация UI-7: Атаки, тикер, луна в карусели (2026-04-22)

### Задача
Завершить P1 из `docs/ui-plan-overview.md`: живые тикеры ресурсов,
баннер входящей атаки, луна в карусели.

### Бэкенд
- `fleet/transport.go`: `IncomingFleet` struct + `ListIncoming` — JOIN флоты→планеты по (galaxy,system,position,is_moon), mission IN (10,12), state='outbound', arrive_at > NOW()
- `fleet/handler.go`: `GET /api/fleet/incoming`
- `building/service.go`: `BuildSecondsMap` — время следующего уровня для каждого здания через `economy.BuildDuration`
- `building/handler.go`: `/api/planets/{id}/buildings/levels` теперь отдаёт `{"levels":…,"build_seconds":…}`
- `research/service.go`: `ResearchSecondsMap` — время через `(metal+silicon)/(1000*(1+labLevel))/gameSpeed`
- `research/handler.go`: `/api/research` теперь отдаёт `{"queue":…,"levels":…,"research_seconds":…}`

### Фронтенд
- `OverviewScreen`: красный пульсирующий баннер атаки (`@keyframes ox-pulse-border`), луна в верхнем правом углу карусельной карточки (`position: absolute, top:3, right:3`)
- `BuildingsScreen`: время ⏱ Xч Xм на каждой карточке из `build_seconds`
- `ResearchScreen`: время ⏱ Xч Xм из `research_seconds`
- `App.tsx ResourceTicker`: получает `metalRate/siliconRate/hydrogenRate` из planet

---

## Итерация UI-8: Склад и энергия в шапке (2026-04-22)

### Задача
P1 из плана: показать вместимость хранилищ под каждым ресурсом и
реальную энергию вместо заглушки "+0".

### Бэкенд
- `planet/model.go`: добавлены поля `MetalCap`, `SiliconCap`, `HydrogenCap`, `EnergyProd`, `EnergyCons`, `EnergyRemaining`
- `planet/service.go`: `energyStats(p, levels, tech)` — абсолютные значения энергии (prod через solar_plant + solar_satellite×EnergyFactor; cons через mine+labs); `applyTickInTx` присваивает все 6 новых полей

### Фронтенд
- `types.ts`: добавлены 6 новых полей в `Planet`
- `App.tsx Header`: под каждым ресурсом — подпись `Nk` (вместимость), цвет зелёный/<90% оранжевый/≥100% красный; `ResourceTicker` получает `cap` (останавливает тикер у потолка); энергия `⚡ prod (+remaining)`, красный при дефиците

### Ключевые решения
- `energyStats` считает абсолютные значения отдельно от `energyRatio` (который возвращает множитель производства). Дублирования нет — разные use case.
- cap-подпись выводится только когда `cap > 0` (новые игроки со стартовыми 5000 видят cap сразу)

---

## Итерация UI-9: Ресурсы в карусели + ScoreScreen альянс (2026-04-22)

### Задача
P2 из плана: ресурсы на мини-карточках карусели, альянс в рейтинге.

### OverviewScreen: ресурсы в карусели
- Под координатами каждой планеты-карточки: 🟠Nk 💎Nk 💧Nk.
- `fmtRes()`: ≥1M → "1.2M", ≥1k → "5k", иначе "500".
- Высота карточки auto (убран fixed height: 110).

### ScoreScreen: альянс в таблице
- `score/service.go` Entry.AllianceTag *string + LEFT JOIN alliances в Top().
- Новая колонка "Альянс" показывает [TAG] или —.

---

## Итерации UI-10..14: Серия улучшений экранов (2026-04-22)

### UI-10: Сравнение Fleet/Market/Score с legacy
- Проведён анализ трёх экранов через legacy .tpl шаблоны.
- Итог: наш Market и Score богаче (лоты, категории, viewers). Fleet беднее (retreat/formation заблокированы мгновенным боем).

### UI-11: MessagesScreen — удалить всё
- `DELETE /api/messages?folder=N` — мягкое удаление всей папки.
- Кнопка 🗑 Удалить все с `window.confirm`, учитывает activeFolder.

### UI-12: GalaxyScreen — картинки планет
- `CellView` расширен: `planet_id`, `planet_type`.
- `ReadSystem()` SELECT p.id, p.planet_type + scan.
- В строке таблицы: `<img 24×24>` через `planetImageOf(position, id, type)`.

### UI-13: OfficersScreen — эффект офицера
- `Entry.effect?: Record<string,number>` из JSONB API.
- `fmtEffect()`: `{produce_factor:1.25}` → "Производство +25%".
- Зелёная строка ✦ на карточке офицера.

### UI-14: OverviewScreen — производство в час
- `ResourceCell`: `perHour = ratePerSec * 3600`.
- Показывается "N/ч" под тикером если > 0.
- Соответствует legacy `fNumber(metal_per_hour)`.

---

## Итерация UI-15: Кредиты в шапке (2026-04-22)

### Задача
P2 из плана: показать баланс кредитов в шапке рядом с ресурсами.

### Предыстория
Кредиты (поле `credit numeric(15,2) DEFAULT 5.00`) существовали в `users` с миграции 0001.
Были ошибочно помечены как «удалены» в docs. Задача свелась к реализации отображения.

### Бэкенд
- `auth/handler.go Me`: добавлено `credit float64` в SELECT + scan; возвращается в `map[string]any`.

### Фронтенд
- `App.tsx`: `me` query тип расширен полем `credit: number`.
- Заголовок: `💳 N cr` в `ox-header-right`; целые — без знака, дробные — `.toFixed(2)`.
- `exactOptionalPropertyTypes` обход: spread-conditional `{...(credit !== undefined ? {credit} : {})}`.

### Ключевые решения
- Кредиты — глобальное поле `/api/me`, не привязано к планете → шапка, не панель ресурсов.
- Стиль `var(--ox-accent)` — кредиты как «особый» ресурс.

---

## Итерация UI-16: Добыча/час в зданиях + описание технологий (2026-04-22)

### Задача
Показать производство в час на производственных зданиях (BuildingsScreen) и
описание эффекта каждой технологии (ResearchScreen).

### BuildingsScreen
- Для зданий `metal_mine`, `silicon_lab`, `hydrogen_lab`: показываем `🟠/💎/💧 Nk/ч`
  из `planet.metal_per_sec * 3600` (суммарная добыча планеты с учётом энергии и офицеров).
- Для `solar_plant`, `hydrogen_plant`: `⚡ N` из `planet.energy_prod`.
- `fmtPerHour(v)`: per-sec → `N/ч`, `Nk/ч`, `N.NM/ч`.
- Нет изменений бэкенда — все данные уже в Planet.

### ResearchScreen
- `catalog.ts ResearchEntry`: новое поле `benefit: string` (статичное описание эффекта).
- Показывается курсивом под уровнем технологии.

### Ключевые решения
- Добыча в зданиях — агрегированная по планете, не per-building. Это корректно:
  шахты не производят ресурсы независимо — у них единый энерго-множитель.
- Описания технологий статичны в catalog.ts — нет смысла гонять их через API.

---

## Итерация UI-33: Глобальный баннер входящей атаки (2026-04-22)

### Задача
Входящие атаки отображались только в OverviewScreen. Если игрок находится на другой вкладке,
он может не заметить атаку. В OGame атака видна везде (в шапке или глобальном баннере).

### Реализация
- `App.tsx`: добавлен `useQuery` для `/api/fleet/incoming` с рефетчем 15s.
- Глобальный баннер между Header и контентом: красный пульсирующий блок с иконкой ⚠️,
  координатами цели, таймером прилёта (Countdown).
- Отображается для каждой входящей атаки отдельно (если несколько).
- Всегда видна независимо от текущей вкладки.

---

## Итерация UI-32: Форматирование боевых отчётов (2026-04-22)

### Задача
В OGame боевой отчёт — красивая вёрстка с цветами (атакующий синий, защитник зелёный),
крупным итогом победителя, блоком добычи, юнитами с именами. В oxsar-nova был серый
текст с `#unit_id` вместо имён.

### Реализация
- `BattleReportView`: полный рефакторинг.
  - Заголовок победителя крупным шрифтом с цветом (`--ox-accent` / `--ox-success` / `--ox-warning`).
  - Карточки «Атакующий» / «Защитник» с цветной рамкой.
  - Блок «Добыча» выделен отдельным панелем.
  - `SideLosses`: отображение юнитов с `nameOf(unit_id)` вместо `#N`, потери `−N` красным.
  - Раунды-по-раундам спрятаны в `<details>` (не захламляют основной вид).
  - Кнопка «Атаковать» перенесена в заголовок.
- `UnitMapBlock`: `nameOf(Number(id))` + `toLocaleString('ru-RU')`.
- `EspionageReportView`: ресурсы с emoji иконками.

---

## Итерация UI-31: Preview времени полёта и расхода водорода (2026-04-22)

### Задача
В OGame отправка флота показывает preview: время в пути, расход дейтерия, время возврата.
В oxsar-nova этого не было — игрок не знал, сколько лететь до отправки.

### Реализация
- `catalog.ts`: добавлены поля `speed` и `fuel` к `CombatEntry` (данные из configs/ships.yml).
- `FleetScreen.tsx`: константа `GAME_SPEED = 0.75`; функции `galaxyDistance` и `flightSecs`
  портируют формулу с бэкенда (Go → TS):
  - `dist = 20000*|Δg|` / `2700+95*|Δs|` / `1000+5*|Δp|` / `5`
  - `t = (10 + 3500/speed_pct * sqrt(10*dist/minSpeed)) / GAME_SPEED`
- Расход водорода: OGame-формула `fuel * dist/35000 * (speed/100+1)^2 * count`.
- Блок preview под формой: `⏱ время`, `↩ время туда-обратно`, `💧 N водорода`.
- Обновляется реактивно при изменении кораблей/координат/скорости.

---

## Итерация UI-30: «Не хватает» — дефицит ресурсов на карточках (2026-04-22)

### Задача
В OGame, если ресурсов не хватает на постройку/исследование, кнопка disabled
и отображается "Не хватает X M / Y K". В oxsar-nova кнопка была кликабельной,
а нехватка выражалась только цветом числа — не очевидно на первый взгляд.

### Реализация
- `BuildingsScreen` и `ResearchScreen`: добавлена строка-дефицит под стоимостью.
- Формат: `🟠−N 💎−M` — только ресурсы, которых не хватает, с суммой недостачи.
- Цвет `var(--ox-danger)`, шрифт `var(--ox-mono)`, размер 10px — компактно.
- Логика через `[...].filter(Boolean).join(' ')` — условное построение строки.

---

## Итерация UI-29: Кнопка «Атаковать» из боевого отчёта (2026-04-22)

### Задача
В боевом отчёте не было кнопки «Атаковать [G:S:P]», хотя enemy-координаты
известны из отчёта. В legacy игре такая кнопка есть — позволяет сразу перейти
к отправке флота на цель.

### Проблема
`BattleReport` содержал `planet_id` (UUID), но не координаты. Переводить UUID
в G:S:P на фронтенде нельзя — нет endpoints для обратного поиска.

### Реализация

**Backend (`backend/internal/message/service.go`):**
- `BattleReport` struct: добавлены `DstGalaxy *int`, `DstSystem *int`, `DstPosition *int`.
- SQL запрос в `GetBattleReport`: добавлен `LEFT JOIN planets p ON p.id = br.planet_id`,
  в SELECT добавлены `p.galaxy, p.system, p.position`.

**Frontend (`frontend/src/features/messages/MessagesScreen.tsx`):**
- `BattleReportFull` interface: добавлены `dst_galaxy/dst_system/dst_position` (nullable).
- Добавлен тип `FleetMissionCb` — колбэк `(g, s, pos, isMoon, mission) => void`.
- Проброс `onFleetMission?` через `MessagesScreen → MessageDetail → BattleReportView`.
- В `BattleReportView`: если координаты известны и колбэк передан — кнопка
  `⚔️ Атаковать [G:S:P]`, mission=10 (атака).

**Frontend (`frontend/src/App.tsx`):**
- В рендер `<MessagesScreen>` добавлен `onFleetMission` — устанавливает `fleetDst`
  и переключает таб на `fleet`.

---

## Итерация UI-28: Вкладка «Боевой» в рейтинге (2026-04-22)

### Задача
Бэкенд уже возвращал `e_points` (боевой опыт) в лидерборде, но ScoreScreen
не показывал соответствующую вкладку.

### Реализация
- `ScoreType` расширен: `'e'`.
- `SCORE_TYPES`: добавлена запись `{ value: 'e', label: 'Боевой', icon: '⚔️' }`.
- `Entry.e_points?: number` добавлен в интерфейс.
- `getPoints`: ветка `type === 'e' → e.e_points ?? 0`.
- API уже поддерживал `?type=e` (columnFor в service.go).

---

## Итерация UI-27: Отмена строительства в очереди (2026-04-22)

### Задача
В очереди зданий не было кнопки отмены. API `DELETE /api/planets/{id}/buildings/queue/{taskId}`
существовал, но не использовался в UI. В legacy игре отмена возвращает ресурсы.

### Реализация
- `cancel` useMutation → `api.delete` → `/api/planets/${planet.id}/buildings/queue/${item.id}`.
- Кнопка «✕» (btn-ghost btn-sm) добавлена в каждую строку `QueueRow`.
- `title="Отменить (ресурсы вернутся)"` — подсказка на hover.
- После успеха invalidate `buildings-queue` и `planets` (ресурсы обновятся).

---

## Итерация UI-26: Мелкие UX-исправления и описания артефактов (2026-04-22)

### Задача
Устранить несколько мелких UX-проблем и добавить описание эффектов артефактов.

### Исправления

**App.tsx — query key унификация:**
- `['messages', 'unread-count']` → `['messages-unread']` (совпадает с OverviewScreen).
- Тип ответа `{ unread }` → `{ count }` (правильный по OpenAPI).
- `unread.data?.unread` → `unread.data?.count`.
- Синхронизированы `invalidateQueries` в MessagesScreen.

**App.tsx — мобильная навигация:**
- `chat` → `messages` в BOTTOM_ITEMS (бейдж непрочитанных теперь на правильной кнопке).

**catalog.ts — ArtefactEntry:**
- Добавлен интерфейс `ArtefactEntry` с полями `benefit` и `lifetime`.
- Все 6 артефактов получили описание эффекта и срок действия (7 дней).

**ArtefactsScreen:**
- В карточке под статусом показывается зелёный italic эффект (`meta.benefit`).

**ArtefactMarketScreen:**
- В таблице под именем артефакта показывается мелкий italic эффект.

---

## Итерация UI-25: Замена иконок ресурсов по всему проекту (2026-04-22)

### Задача
Иконка металла ⛏ (кирка) и кремния 🔷 (синий ромб) не передавали суть ресурсов.
Пользователь предложил новую систему: 🟠 металл (расплавленный металл — оранжевый),
💎 кремний (кристалл/алмаз), 💧 водород (капля, без изменений).

### Процесс выбора
Перебрано ~10 вариантов: ⚒ 🪨 🔩 ⚙️ 🔶 🔸 🧱 🪨 и др.
⚙️ отвергнута — похожа на иконку настроек. Финальный выбор — 🟠 + 💎 + 💧.

### Реализация
`sed -i 's/⛏/🟠/g; s/🔷/💎/g'` применён ко всем файлам frontend/src/, docs/ и
project-creation.txt. Затронуто 13 файлов:
App.tsx, BuildingsScreen, GalaxyScreen, MarketScreen, OverviewScreen,
RepairScreen, ResearchScreen, ShipyardScreen, TutorialScreen,
ui-dev-log.md, ui-plan-overview.md, ui-design-spec.md, project-creation.txt.

---

## Итерация UI-24: Предварительный расчёт возврата при разборе кораблей (2026-04-22)

### Задача
RepairScreen/Disassemble показывал «~70% стоимости» как общую фразу.
Игрок не знал, сколько именно ресурсов получит.

### Реализация
- Тип `units` в `DisassembleList` расширен: добавлено `cost?` поле.
- При изменении `draft` (количество) вычисляется `refund = cost * draft * 0.7`.
- Отображается зелёным: `+🟠N 💎N 💧N` под строкой «В наличии».
- Отображается только при `draft > 0` и ненулевых компонентах refund.

---

## Итерация UI-23: Очередь исследований на странице Обзор (2026-04-22)

### Задача
PlanetOverviewCard показывал очереди строительства и верфи, но не исследований.
На главной странице игрок должен видеть все активные процессы планеты.

### Реализация
- Добавлен запрос `rQueue` к `/api/research` (refetchInterval 5s) в PlanetOverviewCard.
- Фильтр `rItems` = незавершённые элементы из queue.
- Рендеринг с `icon="🔬"` и `nameOf(item.unit_id)` → уже включает RESEARCH в lookup.
- `invalidateQueues` обновлён: добавлен `['research']`.
- `hasActivity` теперь учитывает `rItems.length > 0`.

### Trade-off
Исследования — аккаунтовые (не планетарные), но показываются в карточке
первой планеты (selectedPlanet). Это нормально: исследователская лаборатория
привязана к конкретной планете по очереди, и показ на любой планете избыточен.
Для MVP приемлемо — показываем на всех планетах, если есть активная очередь.

---

## Итерация UI-22: Рефакторинг AllianceScreen в дизайн-систему (2026-04-22)

### Задача
AllianceScreen был единственным экраном, использующим сырые HTML-элементы
(`<section>`, `<h2>`, `<h3>`, `<button>` без классов) вместо дизайн-системы.

### Что сделано
- Навигация переведена на `ox-tabs`.
- Карточка альянса — `ox-panel` с заголовком, тегом в accent-цвете и описанием.
- Список членов — `ox-panel` + `ox-table` с заголовком «Состав (N)».
- Отношения с альянсами — отдельный `ox-panel`, цвет статуса по типу:
  НЕН = dim, ВОЙНА = danger, СОЮЗ = success.
- Заявки на вступление — `ox-panel` с inline approve/reject кнопками.
- Форма создания — `ox-panel` с лейблами и правильными input.
- Все `<button>` → `btn/btn-ghost btn-sm`.
- Ошибки переведены на `toast.show()` вместо `<p className="ox-error">`.
- Удалён `useTranslation` (tf-вызовы с fallback — избыточны на этом экране).

---

## Итерация UI-21: Иконки артефактов (2026-04-22)

### Задача
ArtefactsScreen показывал `✨` вместо реальных иконок — у нас есть GIF для всех
6 реализованных артефактов: merchants_mark, catalyst, power_generator, atomic_densifier,
supercomputer, robot_control_system.

### Реализация
- Добавлен импорт `ARTEFACTS` и `imageOf` в ArtefactsScreen.
- В карточке: lookup по `ARTEFACTS.find(x => x.id === a.unit_id)` → `imageOf(meta.key)`.
- Fallback `✨` оставлен для неизвестных unit_id.

---

## Итерация UI-20: Требования к юнитам в карточках (2026-04-22)

### Задача
Показывать «🔒 Требует: X ур.N + Y ур.M» в карточках верфи (корабли/оборона)
и исследований, чтобы игрок понимал, что нужно построить/изучить для разблокировки.

### Реализация
- Добавлен тип `Req { kind, key, level }` и поле `requires?: Req[]` в
  `ResearchEntry` и `CombatEntry` (catalog.ts).
- Заполнены данные требований из `configs/requirements.yml` для всех 16 исследований,
  12 кораблей и 7 единиц обороны — статически в catalog.ts.
- Добавлена функция `fmtReqs(reqs)` + `nameByKey(key)` — превращает ключи в
  читаемые русские имена через lookup по массивам BUILDINGS/RESEARCH/SHIPS/DEFENSE.
- ResearchScreen: `🔒 ...` под benefit-строкой, только для `level === 0`.
- ShipyardScreen: `🔒 ...` под combat-stats, для всех юнитов (у новичка всё заблокировано).

### Trade-off
Требования отображаются статически (всегда) — не проверяем, выполнены ли они
на текущей планете. Это проще и не требует лишних запросов.
`simplifcations.md` обновлять не нужно — это визуальный справочник, не баланс.

---

## Итерация UI-19: Состав флота в таблице активных флотов (2026-04-22)

### Задача
FleetScreen показывал активные флоты без состава кораблей. OverviewScreen уже умел
рендерить `ships: Record<string, number>` с иконками. Нужна была паритетность.

### Реализация
- Добавлено поле `ships?: Record<string, number>` в `FleetRow` (FleetScreen.tsx).
- Добавлена колонка «Состав» в таблицу активных флотов.
- Каждый тип корабля — иконка 14×14 + название + ×N, flex-wrap.
- Импортирован `imageOfId` из catalog.ts (уже использовался в OverviewScreen).
- API `/api/fleet` уже возвращал `ships` — бэкенд не менялся.

### Trade-off
Нет фильтрации «не показывать 0-кораблей» — API возвращает только ненулевые,
поэтому фильтрация не нужна.

---

## Итерация UI-18.1: Замена window.confirm на кастомный диалог (2026-04-22)

### Задача
Убрать нативные браузерные диалоги (запрещены по design-spec §1.4).

### MessagesScreen
- Кнопка «🗑 Удалить все»: `window.confirm(...)` → `setConfirmDelAll(true)`.
- Рендерит `<Confirm>` компонент (из `ui/Confirm.tsx`) когда `confirmDelAll=true`.
- Текст диалога учитывает активную папку (все / конкретная папка).

---

## Итерация UI-18: Грузоподъёмность флота (2026-04-22)

### Задача
Показать грузоподъёмность (cargo) кораблей в верфи и суммарную вместимость
флота при составлении миссии.

### catalog.ts
- `CombatEntry`: новое опциональное поле `cargo?: number`.
- `SHIPS`: добавлены значения из configs/ships.yml для всех кораблей.
  solar_satellite — без cargo (не транспорт).

### ShipyardScreen
- В строке статов карточки: 📦 N если cargo > 0.

### FleetScreen
- `totalCargo`: `SHIPS.reduce(sum + cargo * count)`.
- При mission=7/8 (транспорт/колонизация): подпись «📦 макс. N» рядом с лейблом «Груз».

---

## Итерация UI-18.5: Бонусы офицеров в зданиях/исследованиях (2026-04-22)

### Задача
Показывать активные бонусы офицеров прямо на экранах строительства и исследований.

### Фронтенд
- `types.ts Planet`: добавлены `produce_factor`, `build_factor`, `research_factor` (опциональные).
- `BuildingsScreen`: в шапке — `🏗 +N% строительство` и `🟠 +N% добыча` если factor > 1.
- `ResearchScreen`: `🔬 +N% исследование` если factor > 1.
- Поля уже были в backend Planet model (json:"build_factor" и т.д.) — нулевых изменений бэкенда.

---

## Итерация UI-17: Туториал + рынок артефактов (2026-04-22)

### Задача
Показать награды за каждый шаг туториала и дату листинга на рынке артефактов.

### TutorialScreen
- Статичный массив `STEP_RESOURCES` зеркалит backend `stepResources` (6 шагов).
- Для незакрытых шагов: строка `💳 +10 cr 🟠 +N 💎 +N 💧 +N` зелёным.
- Убран общий «Каждый шаг даёт +10 кредитов» — замещён конкретными цифрами.

### ArtefactMarketScreen
- Новая колонка «Дата» с `listed_at` в формате ДД.ММ.
- Поле уже было в интерфейсе `Offer` но не отображалось.

---

## Итерация UI-18: BuildingsScreen — план 09 (2026-04-23)

### Задача
Сравнение с legacy `/game.php/Constructions` и устранение расхождений.

### Критический баг: missile_silo id
- Nova использовала id=13 для ракетной шахты. В legacy id=13 = `UNIT_SPYWARE` (research).
- Правильный id = 53 (`UNIT_ROCKET_STATION`). Исправлено в configs, catalog, rocket/service.go.

### Нано-фабрика (id=7)
- Здание отсутствовало в nova. Добавлено в buildings.yml и BUILDINGS.
- Backend: `BuildSecondsMap` и `Enqueue` теперь читают nanoLevel из БД и передают в `BuildDuration`.
- `economy.BuildDuration` уже поддерживал параметр — просто всегда получал 0.

### Описания и MAX badge
- Добавлен `description?` в `BuildingEntry`, все 14 зданий получили описания.
- При `level >= maxLevel` кнопка заменяется на текст «MAX».

### Пререквизиты на карточке
- Backend: новый метод `UnmetForTarget` в requirements.Checker (читает без транзакции).
- `RequirementsUnmet` в building.Service собирает unmet для всех зданий планеты.
- `GET /api/planets/{id}/buildings` теперь возвращает `requirements_unmet`.
- Frontend: карточка показывает `🔒 key ур.N (у вас: M)` и кнопку «Заблокировано».

### Фильтр «Только доступные / Все здания»
- `useState` + `localStorage('buildings-show-locked')`.
- По умолчанию скрываются здания с level=0 и невыполненными требованиями.

### Лунные здания
- Добавлены moon_base(54), star_surveillance(55), star_gate(56), moon_robotic_factory(57).
- `BuildingSpec.MoonOnly bool` в config; валидация в `Enqueue` — ErrMoonOnly/ErrPlanetOnly.
- Frontend: `MOON_BUILDINGS[]`, выбор по `planet.is_moon`.

### Не реализовано (P3)
- Задача 9 (снос здания) и Задача 10 (детальный модал) — оставлены как P3.

---

## Итерация UI-35: Детальные модалы зданий и исследований — план 10 (2026-04-23)

### Задача
Сравнение с legacy `ConstructionInfo/{id}`: nova показывала только таблицу стоимость/время.
Legacy дополнительно показывает полное описание (_FULL_DESC), производство по уровням
для шахт/электростанций, и пререквизиты.

### catalog.ts
- `BuildingEntry`: добавлено поле `fullDesc?: string`.
- `ResearchEntry`: добавлено поле `fullDesc?: string`.
- Все 14 зданий получили `fullDesc` из `configs/i18n/ru.yml` (`*_FULL_DESC` ключи).
- Все 16 исследований получили `fullDesc` аналогично.
- Moon buildings: `moon_base` и `star_gate` получили `fullDesc`; `star_surveillance` — нет в i18n.

### BuildingInfoModal
- Новая колонка «производство/ч» в таблице для зданий с добычей/энергией.
  Формула: `floor(base_rate × level × 1.1^level)`. Base rates из configs/buildings.yml:
  metal_mine=30, silicon_lab=20, hydrogen_lab=10, solar_plant=20, hydrogen_plant=22.5.
  Заголовок колонки меняется по типу: 🟠/ч, 💎/ч, 💧/ч, ⚡/ч.
- Секция `<details><summary>Подробнее</summary>…</details>` под таблицей с `fullDesc`.

### ResearchInfoModal
- Секция пререквизитов `🔒 Требуется: …` через `fmtReqs()` под заголовком (если есть).
- Секция `<details>Подробнее</details>` под таблицей с `fullDesc`.

### Не реализовано
- Снос здания через модал (P3, план 09 задача 9).
- Производство в модале исследований (исследования не производят ресурсы напрямую).

## Итерация UI-36: UnitInfoScreen — отдельный экран вместо модальных окон — план 12 (2026-04-23)

### Проблема
`BuildingInfoModal` и `ResearchInfoModal` имели стойкий CSS-баг: при раскрытии блока
«Подробнее» текст вываливался за границы белой панели на тёмный фон оверлея.
Причина — `display: flex` + `gap` на flex-контейнере модала не ограничивает высоту
дочерних элементов при росте. Перепробованы: `max-height`, `grid-template-rows`,
`alignItems: flex-start`, plain-toggle — все дали тот же визуальный результат.

### Решение — отдельная страница (как в legacy)
Legacy `ConstructionInfo/{id}` — это отдельная страница, не попап. Делаем аналогично:
новый `Tab = 'unit-info'` без хэша-навигации (не добавлен в `VALID_TABS` через URL),
новый компонент `UnitInfoScreen`.

### Реализация
- **`UnitInfoScreen.tsx`** — новый единый компонент для зданий и исследований.
  Кнопка «← Назад» возвращает на `fromTab` (buildings/research).
  Секции: заголовок с иконкой, пререквизиты, таблица уровней (+колонка производства),
  сноска, полный текст описания. Нормальный document flow, без фиксированных высот.
- **`App.tsx`**: добавлен `Tab 'unit-info'`, state `infoUnit { kind, id, level, fromTab }`,
  функция `openInfo(kind, id, level)`, рендер `UnitInfoScreen` с колбэком `onBack`.
- **`BuildingsScreen`**: добавлен проп `onOpenInfo(id, level)`, убран `infoUnitId` state.
- **`ResearchScreen`**: аналогично.
- **Удалены**: `BuildingInfoModal.tsx`, `ResearchInfoModal.tsx`.
