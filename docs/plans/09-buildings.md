# План: Экран построек — доработка BuildingsScreen

## Статус: ЗАВЕРШЁН (задачи 1–7 реализованы; задача 8 лунные здания — P2 не делаем)

## Контекст

Legacy страница `/game.php/Constructions` — основной экран строительства зданий на планете.

**Источник legacy:** `d:\Sources\oxsar2\www\templates\standard\constructions.tpl` + `Constructions.class.php` + `Construction.class.php`

**Nova экран:** `frontend/src/features/buildings/BuildingsScreen.tsx`

---

## Сравнение: legacy vs nova

### 1. Что есть в nova (реализовано)

| Функция | Статус |
|---------|--------|
| Сетка карточек зданий с иконкой, именем, уровнем | ✅ |
| Стоимость следующего уровня (M / Si / H) с подсветкой нехватки | ✅ |
| Дефицит ресурсов (красный дельта) | ✅ |
| Время постройки следующего уровня | ✅ |
| Текущее производство здания (добыча/ч, энергия) для шахт и электростанций | ✅ |
| Очередь строительства: прогресс-бар, countdown, кнопка отмены | ✅ |
| Добавление в очередь / кнопка «Построить» / «→ ур. N» | ✅ |
| Бейдж планетарных факторов (build_factor, produce_factor) | ✅ |
| Отмена с возвратом ресурсов (95%, первые 15 сек — 100%) | ✅ |
| Скелетон при загрузке | ✅ |
| Toast-уведомления об успехе/ошибке | ✅ |

### 2. Что есть в legacy, но отсутствует в nova

| Функция | Приоритет | Комментарий |
|---------|-----------|-------------|
| **Требования для здания (пререквизиты)** | P1 | Legacy показывает список зданий/исследований нужных для разблокировки; nova не показывает ничего когда здание заблокировано |
| **Фильтр «Показать недоступные»** | P1 | Legacy: checkbox `show_all_units`. В nova все здания показываются всегда — нет скрытия по требованиям |
| **Максимальный уровень здания** | P2 | Legacy блокирует апгрейд при `level >= MAX_BUILDING_LEVEL`; nova не проверяет максимум |
| **Описание здания** | P2 | Legacy показывает краткое описание под именем (`_DESC`) и ссылку на полное (`ConstructionInfo`) |
| **Энергопотребление при строительстве** | P2 | Legacy показывает `energy_required` рядом со стоимостью; nova не показывает |
| **Нано-фабрика (id=7)** | P2 | Отсутствует в nova `BUILDINGS`; в legacy делит время постройки на `2^nano_factory_level` |
| **Лунные здания (mode=5)** | P2 | `moon_base`, `moon_robotic_factory`, `star_gate`, `star_surveillance` — только на луне; nova не реализует |
| **Снос здания (Demolish)** | P3 | Legacy: понижение уровня с возвратом части ресурсов |
| **Детальный экран здания** | P3 | Legacy `/game.php/ConstructionInfo/{id}` — таблица prod/cons по уровням ±7 |
| **«Начать немедленно» за кредиты (VIP)** | P4 | Кнопка ускорения строительства за кредиты; намеренно пропускаем |
| **Множественная очередь** | — | Намеренное упрощение (1 слот); задокументировано в simplifications.md |

### 3. Расхождения в данных

| Параметр | Legacy | Nova | Действие |
|----------|--------|------|----------|
| Ракетная шахта id | 53 (`UNIT_ROCKET_STATION`) | 13 (`missile_silo`) | ⚠️ Критично — исправить |
| id=13 в legacy | `UNIT_SPYWARE` (research, mode=2) | `missile_silo` (building) | Конфликт типов |
| `metal_storage` costBase | metal:2000 | metal:1000 | Сверить с buildings.yml |
| `nano_factory` (id=7) | В очереди (mode=1) | Отсутствует в BUILDINGS | Добавить |

---

## Структура legacy (что передаётся в шаблон)

### Карточка здания:
- `name` — ссылка на ConstructionInfo
- `level` — текущий уровень
- `added_level` — бонус уровня от артефактов
- `can_build` — bool (требования выполнены)
- `required_constructions` — список требований если `can_build=false`
- `metal_required`, `silicon_required`, `hydrogen_required`, `energy_required`
- `metal_notavailable`, ... — дефицит (отрицательный)
- `productiontime` — форматированное время постройки
- `upgrade` — кнопка/ссылка или текст ошибки

### Очередь:
- `name` — название + целевой уровень
- `cancel_link` — кнопка отмены + countdown JS
- `vip_link` — кнопка «немедленно за кредиты»
- `event_pb_value` — % прогресса

---

## План реализации

### Задача 1 — Исправить id ракетной шахты [КРИТИЧНО]

**Проблема:** `missile_silo` в nova имеет id=13, но в legacy:
- id=13 = `UNIT_SPYWARE` (research, mode=2)
- id=53 = `UNIT_ROCKET_STATION` (здание, mode=1)

**Действия:**
1. Проверить в PostgreSQL: `SELECT * FROM buildings WHERE unit_id=13 OR unit_id=53`
2. Исправить `configs/buildings.yml`: `missile_silo.id: 53`
3. Исправить `frontend/src/api/catalog.ts`: id=53, key=`rocket_station`
4. Написать миграцию `UPDATE buildings SET unit_id=53 WHERE unit_id=13` если id уже в БД

**Файлы:** `configs/buildings.yml`, `frontend/src/api/catalog.ts`, `migrations/`

---

### Задача 2 — Требования (пререквизиты) на карточке [P1]

**Описание:** Когда здание недоступно, legacy показывает «Требуется: X ур.N». Nova ничего не показывает — кнопка просто disabled.

**Backend** — расширить `GET /api/planets/{id}/buildings`:
```json
{
  "levels": { "1": 5 },
  "requirements_unmet": {
    "53": [
      { "kind": "building", "unit_id": 3, "name": "Верфь", "required": 3, "current": 2 }
    ]
  }
}
```

**Frontend** — если здание заблокировано, показать под именем:
```
🔒 Верфь ур.3 (у вас: 2)
```

**Файлы:** `backend/internal/building/handler.go`, `backend/internal/building/service.go`, `frontend/src/features/buildings/BuildingsScreen.tsx`, `frontend/src/api/types.ts`

---

### Задача 3 — Фильтр «Показать недоступные» [P1]

**Действия (только Frontend):**
1. `useState<boolean>(false)` — `showLocked`
2. Скрывать здания с `level === 0` И невыполненными требованиями если `!showLocked`
3. Toggle-кнопка в шапке: «👁 Все здания» / «Доступные»
4. Сохранять в `localStorage('buildings-show-locked')`

**Файлы:** `frontend/src/features/buildings/BuildingsScreen.tsx`

---

### Задача 4 — Максимальный уровень [P2]

**Действия:**
1. Добавить в `configs/buildings.yml` поле `max_level` (default: 50)
2. Backend: проверять в `Enqueue` — если `targetLevel > spec.MaxLevel` → `ErrMaxLevelReached`
3. Frontend: показывать badge «MAX» вместо кнопки когда `level >= maxLevel`

**Файлы:** `configs/buildings.yml`, `backend/internal/building/service.go`, `frontend/src/features/buildings/BuildingsScreen.tsx`

---

### Задача 5 — Нано-фабрика (id=7) [P2]

**Описание:** Отсутствует в nova. В legacy делит время строительства: `duration / 2^nano_level`.

**Действия:**
1. Добавить в `configs/buildings.yml`:
   ```yaml
   nano_factory:
     id: 7
     cost_base: { metal: 1000000, silicon: 500000, hydrogen: 100000 }
     cost_factor: 2.0
   ```
2. Добавить в `frontend/src/api/catalog.ts` в `BUILDINGS`
3. В `backend/internal/building/service.go` — учитывать `nano_factory` уровень в `BuildDuration`:
   ```go
   duration = duration / math.Pow(2, float64(nanoLevel))
   ```

**Файлы:** `configs/buildings.yml`, `frontend/src/api/catalog.ts`, `backend/internal/building/service.go`

---

### Задача 6 — Описание здания на карточке [P2]

**Действия:**
1. Добавить поле `description: string` в `BUILDINGS` в `catalog.ts`
2. Показывать под именем мелким шрифтом (1–2 строки, ellipsis)

Примеры:
- Рудник металла: «Добывает металл из недр планеты»
- Фабрика роботов: «Ускоряет строительство зданий»
- Верфь: «Производит корабли и оборонительные системы»

**Файлы:** `frontend/src/api/catalog.ts`, `frontend/src/features/buildings/BuildingsScreen.tsx`

---

### Задача 7 — Энергопотребление при строительстве [P2]

**Описание:** Legacy показывает `energy_required` в таблице ресурсов. Nova не отображает.

**Действия:**
1. Добавить `energy_cost: int` в `buildings.yml` для зданий с ненулевым значением
2. Backend: включать в ответ Levels
3. Frontend: если `energy_cost > 0` — показывать `⚡{energy_cost}` в блоке стоимости

**Файлы:** `configs/buildings.yml`, `backend/internal/building/handler.go`, `frontend/src/features/buildings/BuildingsScreen.tsx`

---

### Задача 8 — Лунные здания [P2]

**Здания только для луны (is_moon=true):**
- `moon_base` (id=54)
- `star_surveillance` (id=55)
- `star_gate` (id=56)
- `moon_robotic_factory` (id=57)

**Действия:**
1. Добавить в `configs/buildings.yml` с пометкой `moon_only: true`
2. Добавить в `frontend/src/api/catalog.ts` массив `MOON_BUILDINGS`
3. В `BuildingsScreen.tsx`: выбирать `BUILDINGS` или `MOON_BUILDINGS` по `planet.is_moon`
4. Backend: `Enqueue` проверяет `moon_only` и `ismoon` планеты

**Файлы:** `configs/buildings.yml`, `frontend/src/api/catalog.ts`, `frontend/src/features/buildings/BuildingsScreen.tsx`, `backend/internal/building/service.go`

---

### Задача 9 — Снос здания (Demolish) [P3]

**Здания с demolish:** metalmine, silicon_lab, hydrogen_lab, solar_plant, robotic_factory, shipyard, storage x3, research_lab, rocket_station, repair_factory, nano_factory.

**Формула:** `cost(current_level) / demolish_factor` — возврат ресурсов.

**Действия:**
1. Backend: `POST /api/planets/{id}/buildings/{unitId}/demolish`
2. Проверка: `level > 0`, не в очереди, `demolish_factor > 0`
3. Frontend: кнопка «Снести» с подтверждением

**Файлы:** `configs/buildings.yml` (+demolish), `backend/internal/building/`, `frontend/src/features/buildings/BuildingsScreen.tsx`

---

### Задача 10 — Детальный экран здания [P3]

**Описание:** Модальное окно (не отдельная страница) с таблицей prod/cons по уровням ±7 и временем постройки каждого уровня.

**Действия:**
1. Backend: `GET /api/planets/{id}/buildings/{unitId}/info`
2. Frontend: `BuildingInfoModal.tsx` — открывается по клику на иконку/имя

**Файлы:** `backend/internal/building/handler.go`, `frontend/src/features/buildings/BuildingInfoModal.tsx`

---

## Приоритеты

| # | Задача | Приоритет | Усилие |
|---|--------|-----------|--------|
| 1 | Исправить id ракетной шахты (53 vs 13) | КРИТИЧНО | S |
| 2 | Требования на карточке | P1 | M |
| 3 | Фильтр «Показать недоступные» | P1 | S |
| 4 | Максимальный уровень | P2 | S |
| 5 | Нано-фабрика (id=7) | P2 | M |
| 6 | Описание здания | P2 | S |
| 7 | Энергопотребление при строительстве | P2 | S |
| 8 | Лунные здания | P2 | L |
| 9 | Снос здания | P3 | L |
| 10 | Детальный экран BuildingInfoModal | P3 | L |

---

## Намеренные упрощения vs legacy (не реализуем)

| Функция | Причина |
|---------|---------|
| Множественная очередь (N слотов) | Упрощение §5.3 ТЗ — 1 слот достаточно |
| VIP-ускорение за кредиты | Монетизация — отдельное решение |
| Альтернативные скины иконок | Нет необходимости |

---

## Файлы, которые будут изменены

| Файл | Изменение |
|------|-----------|
| `configs/buildings.yml` | +nano_factory, +max_level, +demolish, +moon buildings, +energy_cost |
| `frontend/src/api/catalog.ts` | +nano_factory, +MOON_BUILDINGS, исправить id ракетной шахты, +description |
| `frontend/src/features/buildings/BuildingsScreen.tsx` | +фильтр, +требования, +max уровень, +описание, +луна |
| `frontend/src/api/types.ts` | +requirements_unmet в ответе buildings |
| `backend/internal/building/service.go` | +nano_factory в BuildDuration, +проверка max_level |
| `backend/internal/building/handler.go` | +requirements_unmet в ответе, +demolish endpoint, +info endpoint |
| `migrations/` | Исправить unit_id ракетной шахты если нужно |
| `docs/simplifications.md` | Зафиксировать: 1 слот очереди, нет VIP-ускорения |
