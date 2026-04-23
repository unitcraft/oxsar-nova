# План 13: Экран верфи — доработка ShipyardScreen

## Статус: В РАБОТЕ (задачи 0–4 + 6 выполнены + hotfix, задача 5 — P3)

## Контекст

Legacy страница `/game.php/Shipyard` — экран строительства кораблей и оборонительных
сооружений. Отдельный раздел для очереди верфи.

**Источник legacy:** `d:\Sources\oxsar2\www\templates\standard\shipyard.tpl` + `Shipyard.class.php`

**Nova экран:** `frontend/src/features/shipyard/ShipyardScreen.tsx`

---

## Сравнение: legacy vs nova

### 1. Что есть в nova (реализовано)

| Функция | Статус |
|---------|--------|
| Очередь верфи: прогресс-бар и countdown для активного задания | ✅ |
| Очередь верфи: список ожидающих с временем завершения | ✅ |
| Табы «Корабли» / «Оборона» | ✅ |
| Карточки юнитов с иконкой, именем, боевыми характеристиками (⚔/🛡/❤) | ✅ |
| Грузоподъёмность (📦) для транспортов | ✅ |
| Требования (🔒) на карточке | ✅ |
| Количество в наличии (В наличии: N) | ✅ |
| Стоимость (M/Si/H) с подсветкой нехватки | ✅ |
| Дефицит ресурсов (красный дельта) | ✅ |
| Input для количества + кнопка «Строить» | ✅ |
| Toast-уведомления об успехе/ошибке | ✅ |
| Проверка требований на backend при enqueue | ✅ |

### 2. Что есть в legacy, но отсутствует в nova

| Функция | Приоритет | Комментарий |
|---------|-----------|-------------|
| **Отмена задания в очереди (cancel)** | P1 | Legacy показывает cancel_link для каждого задания; backend-endpoint отсутствует |
| **Описание юнита** | P1 | Legacy показывает `{description}` под именем; нет в каталоге |
| **Скорость и расход топлива для кораблей** | P2 | Данные уже в SHIPS (speed, fuel), но не отображаются на карточке |
| **Фильтр «Доступные / Все»** | P2 | Legacy: `can_build` bool на каждом юните; nova показывает все всегда |
| **Rapidfire (быстрый огонь)** | P3 | Данные есть в backend catalog; frontend не получает и не показывает |
| **VIP-ускорение за кредиты** | P4 | Legacy: vip_link; намеренно пропускаем |

### 3. Данные есть в каталоге, но не показываются в UI

| Поле | Где хранится | Статус в UI |
|------|-------------|-------------|
| `speed` | `SHIPS[].speed` в catalog.ts | ❌ не показывается |
| `fuel` | `SHIPS[].fuel` в catalog.ts | ❌ не показывается |
| rapidfire | backend `RapidfireCatalog` | ❌ не передаётся на frontend |

---

## Структура legacy (что передаётся в шаблон)

### Очередь:
- `number` — позиция в очереди
- `name` — название юнита + количество
- `cancel_link` — кнопка отмены задания
- `vip_link` — VIP-ускорение
- `event_pb_value` — % прогресса (JS progressbar)

### Карточка юнита:
- `image` — иконка
- `name` — ссылка на страницу юнита
- `description` — краткое описание под именем
- `can_build` — bool (требования выполнены)
- `required_constructions` — текст требований если `can_build=false`
- `metal_required`, `silicon_required`, `hydrogen_required` — стоимость
- `quantity_num` — текущее количество в наличии
- `construct` — поле ввода количества + кнопка «Построить»

---

## План реализации

### Задача 0 — Исправить вёрстку карточек [P1 КРИТИЧНО]

**Проблема:** ShipyardScreen использует самодельную инлайн-вёрстку карточек, которая
расходится со структурой BuildingsScreen и CSS-классами из `app.css`. Видимые дефекты:
- Иконка 64×64 торчит как отдельный блок, не выровнена с телом карточки в `ox-unit-card-header`
- Блок характеристик (⚔/🛡/❤/📦), требования и стоимость — единая свалка без чёткого
  визуального разделения
- Требования занимают 2–3 строки на длинных списках, «съедая» карточку

**Решение:** Привести структуру карточки в ShipyardScreen к той же схеме, что в BuildingsScreen:

```
div.ox-unit-card
  div (flex, gap 10, alignItems flex-start)
    img 64×64  (иконка — кликабельна → UnitInfoScreen)
    div.ox-unit-card-body (flex:1, minWidth:0, overflow:hidden)
      div.ox-unit-card-name   — имя
      div                     — характеристики ⚔/🛡/❤/📦 (flex-wrap)
      div                     — скорость/топливо (только SHIPS, мелко)
      div                     — требования 🔒 (если есть)
      div                     — в наличии: N (если > 0)
      div                     — стоимость (моно, flex-wrap)
      div                     — дефицит (красный, если не хватает)
  div.ox-unit-card-footer
    input[number] + button «Строить»
```

**Файлы:** `frontend/src/features/shipyard/ShipyardScreen.tsx`

---

### Задача 1 — Отмена задания в очереди [P1]

**Описание:** Legacy показывает cancel для каждого задания. В nova нет ни backend-эндпоинта,
ни кнопки на фронте.

**Backend** — добавить endpoint:
```
DELETE /api/planets/{planetId}/shipyard/{queueId}
```
- Проверить, что задание принадлежит планете игрока
- Вернуть ресурсы на планету (100% — верфь не строит «частично»)
- Удалить запись из очереди
- Если задание активное (первое) — пересчитать end_at следующего

**Frontend** — добавить кнопку «✕» в `ShipQueueRow`:
```tsx
<button onClick={() => cancel.mutate(item.id)}>✕</button>
```
- Мутация с `onSuccess`: invalidate `shipyard-queue` и `planets`
- Confirm-диалог не нужен (логика быстрая, ресурсы возвращаются)

**Файлы:** `backend/internal/shipyard/service.go`, `backend/internal/shipyard/handler.go`,
`frontend/src/features/shipyard/ShipyardScreen.tsx`

---

### Задача 2 — Описание юнита на карточке [P1]

**Описание:** Legacy показывает краткое описание под именем.
В nova нет поля `description` ни в SHIPS, ни в DEFENSE.

**Frontend** — добавить `description?: string` в `CombatEntry` в `catalog.ts`
и заполнить для всех кораблей и оборонительных сооружений:

| key | Описание |
|-----|----------|
| small_transporter | Дёшев, перевозит до 5 000 ед. ресурсов |
| large_transporter | Основной грузовоз, 25 000 ед. cargo |
| light_fighter | Дешёвый и быстрый — основа атакующего флота |
| strong_fighter | Более мощная альтернатива лёгкому истребителю |
| cruiser | Эффективен против ракетных установок (rapidfire) |
| battle_ship | Мощный линкор с гиперпространственным двигателем |
| colony_ship | Позволяет колонизировать незанятые планеты |
| recycler | Собирает обломки в полях мусора |
| espionage_sensor | Зонд для разведки — перехватывается при слабом шпионаже |
| solar_satellite | Добавляет энергию без строительства электростанции |
| bomber | Специализируется по обороне противника (rapidfire) |
| death_star | Сильнейший корабль; уничтожает луны |
| rocket_launcher | Базовая и дешёвая оборона |
| light_laser | Лазерная пушка начального уровня |
| strong_laser | Более мощная лазерная пушка |
| gauss_gun | Мощная пушка против тяжёлых кораблей |
| plasma_gun | Наиболее разрушительное орудие обороны |
| small_shield | Купол защищает всю оборону от одного залпа |
| large_shield | Усиленный купол с 10× большим щитом |

**Отображение:** мелким курсивом (`fontSize: 11, fontStyle: 'italic'`) под именем,
2 строки максимум (`overflow: hidden, display: -webkit-box, -webkit-line-clamp: 2`).

**Файлы:** `frontend/src/api/catalog.ts`, `frontend/src/features/shipyard/ShipyardScreen.tsx`

---

### Задача 3 — Скорость и расход топлива на карточке кораблей [P2]

**Описание:** Поля `speed` и `fuel` уже есть в `SHIPS` в `catalog.ts`, но не показываются.

**Frontend** — добавить в блок характеристик карточки корабля (только для SHIPS, не DEFENSE):
```
🚀 12 500   ⛽ 20/ед.
```

Формат: `speed.toLocaleString('ru-RU')`, `fuel` — только если fuel > 0.

**Файлы:** `frontend/src/features/shipyard/ShipyardScreen.tsx`

---

### Задача 4 — Фильтр «Доступные / Все» [P2]

**Описание:** По умолчанию показывать только юниты с выполненными требованиями.
Кнопка «👁 Все» разворачивает недоступные (со стилем `opacity: 0.5`).

**Действия:**
1. `useState<boolean>(false)` — `showLocked` (отдельно для кораблей и обороны)
2. Юнит «заблокирован» если у него есть requires, которые не выполнены
   - Данные о выполнении требований: сравнивать с уровнями зданий и исследований
   - Альтернатива проще: показывать все, заблокированные — с `opacity: 0.5` и disabled-кнопкой
3. Toggle в шапке таба: «👁 Все» / «✅ Доступные»
4. Сохранять в `localStorage('shipyard-show-locked')`

**Примечание:** требования уже есть на карточке (🔒 строка), достаточно CSS-фильтрации.

**Файлы:** `frontend/src/features/shipyard/ShipyardScreen.tsx`

---

### Задача 6 — UnitInfoScreen для кораблей и обороны [P1] ✅

**Описание:** Аналог `game.php/UnitInfo/29` — отдельный экран с полной информацией о
корабле или оборонительном сооружении. Открывается кликом по иконке на карточке верфи.

**Реализовано:**
- Расширен `InfoUnit` в App.tsx: `kind: 'building' | 'research' | 'ship' | 'defense'`
- `parseHash()` теперь распознаёт `#unit-info/ship/29` и `#unit-info/defense/43`
- `openInfo()` принимает `InfoUnit['kind']` вместо жёстко заданных двух значений
- `ShipyardScreen` получил prop `onOpenInfo?: (kind, id) => void`
- Клик по иконке юнита открывает UnitInfoScreen
- `UnitInfoScreen` расширен компонентом `CombatUnitInfo`:
  - Заголовок с иконкой 128×128, описанием
  - Таблица боевых характеристик (⚔/🛡/❤/📦/🚀/⛽/стоимость)
  - Таблица «Быстрый огонь» — по каким целям стреляет N× за раунд
  - Таблица «Уязвим к быстрому огню» — кто стреляет по этому юниту быстро
- `CombatEntry` расширен полем `rapidfire?: Record<number, number>`
- Rapidfire-данные из `configs/rapidfire.yml` добавлены в catalog.ts:
  - Крейсер (33): лёгкий истребитель ×6, ракетная установка ×10
  - Линкор (34): шпионский зонд ×5, солнечный спутник ×5
  - Звезда смерти (42): по всем юнитам (250×–1250×)

**Файлы:** `frontend/src/api/catalog.ts`, `frontend/src/features/unit-info/UnitInfoScreen.tsx`,
`frontend/src/features/shipyard/ShipyardScreen.tsx`, `frontend/src/App.tsx`

---

### Задача 5 — Rapidfire (быстрый огонь) [P3]

**Описание:** В legacy каждый корабль имеет таблицу rapidfire — сколько раз он может
выстрелить по конкретным целям за один раунд боя. Данные есть в backend catalog, но
не передаются на frontend.

**Источник данных:** `backend/internal/catalog/catalog.go` → `RapidfireCatalog.Rapidfire`

**Варианты реализации:**
1. Встроить rapidfire прямо в `SHIPS`/`DEFENSE` в `catalog.ts` как `rapidfire?: Record<number, number>`
2. Или показывать в `UnitInfoScreen` (отдельная задача)

**Отображение:** раздел «Быстрый огонь» в `UnitInfoScreen` (план 12), не на карточке верфи
(карточка и так насыщена).

**Файлы:** `frontend/src/api/catalog.ts`, `frontend/src/features/unit-info/UnitInfoScreen.tsx`

---

## Приоритеты

| # | Задача | Приоритет | Статус |
|---|--------|-----------|--------|
| 0 | Исправить вёрстку карточек | P1 КРИТ | ✅ |
| 1 | Отмена задания в очереди (cancel) | P1 | ✅ |
| 2 | Описание юнита на карточке | P1 | ✅ |
| 3 | Скорость и расход топлива | P2 | ✅ |
| 4 | Фильтр «Доступные / Все» | P2 | ✅ |
| 5 | Rapidfire в UnitInfoScreen (расширить) | P3 | — |
| 6 | UnitInfoScreen для кораблей и обороны | P1 | ✅ |

---

### Hotfix 3 — кнопка «Строить» и инпут при недоступном юните

Если юнит нельзя построить (нет ресурсов или не выполнены требования):
- Кнопка «Строить» — красная (`btn-danger`) и `disabled`
- `<input>` количества — `disabled`

**Файлы:** `frontend/src/features/shipyard/ShipyardScreen.tsx`

---

### Hotfix 2 — пустой список при активном фильтре «Скрыть недоступные»

Когда фильтр включён и все юниты заблокированы — показывать надпись:
«Все корабли требуют выполнения условий. Нажмите „Показать все" чтобы увидеть список.»

**Файлы:** `frontend/src/features/shipyard/ShipyardScreen.tsx`

---

### Hotfix 1 — пустой экран верфи после внедрения фильтра

**Проблема:** логика фильтра была перевёрнута: `showLocked=false` (дефолт) оставлял
только юниты без `requires`, а все корабли и оборона имеют requires → экран пустой.

**Исправление:**
- Дефолт `showLocked=true` (показывать всё)
- `localStorage !== 'false'` вместо `=== 'true'`
- Текст кнопки: «🔒 Скрыть недоступные» / «👁 Показать все»

---

## Намеренные упрощения vs legacy (не реализуем)

| Функция | Причина |
|---------|---------|
| VIP-ускорение за кредиты | Монетизация — отдельное решение |
| Множественная очередь (N слотов) | Упрощение — 1 активное задание, остальные ждут |
| Построить «всё доступное» одним кликом | Нет в legacy, не нужно |

---

## Файлы, которые будут изменены

| Файл | Изменение |
|------|-----------|
| `backend/internal/shipyard/service.go` | +Cancel метод |
| `backend/internal/shipyard/handler.go` | +DELETE endpoint |
| `frontend/src/api/catalog.ts` | +`description` в CombatEntry, заполнить для всех |
| `frontend/src/features/shipyard/ShipyardScreen.tsx` | +cancel-кнопка, +description, +speed/fuel, +фильтр |
| `frontend/src/features/unit-info/UnitInfoScreen.tsx` | +rapidfire секция (задача 5) |
