# План 10: Детальные модалы зданий и исследований (ConstructionInfo)

## Статус: В РАБОТЕ

## Контекст

Legacy `/game.php/ConstructionInfo/{id}` показывает детальную страницу юнита:
- Полное описание (`_FULL_DESC` из i18n/ru.yml) — длинный текст с механикой
- Chart/таблица производства по уровням (для шахт — металл/кремний/водород в час, для электростанций — энергия)
- Секция сноса (demolish) если здание можно снести
- Пререквизиты (requires)

В nova реализованы `BuildingInfoModal` и `ResearchInfoModal` — но только таблица стоимость/время.

**Все тексты `_FULL_DESC` уже есть** в `configs/i18n/ru.yml` (строки 1100+).

---

## Сравнение nova vs legacy

| Элемент | Legacy ConstructionInfo | Nova modal | Действие |
|---------|------------------------|------------|----------|
| Короткое описание (`_DESC`) | ✅ | ✅ (поле `description` в catalog.ts) | — |
| **Полное описание (`_FULL_DESC`)** | ✅ | ❌ | Добавить |
| Таблица стоимости по уровням | ✅ | ✅ | — |
| Время постройки по уровням | ✅ | ✅ (приближённо, без robo/nano) | — |
| **Производство по уровням** (шахты, электростанции) | ✅ chart | ❌ | Добавить |
| **Пререквизиты** | ✅ | ❌ | Добавить |
| Снос здания (demolish) | ✅ | ❌ | P3, отдельно |
| Ссылка «Купить за кредиты» | ✅ VIP | ❌ | Не делаем |

---

## Задача 1 — Полные описания (_FULL_DESC) в catalog.ts [P1]

Все тексты уже в `configs/i18n/ru.yml`. Нужно перенести в `catalog.ts`.

### Здания — добавить `fullDesc` в `BuildingEntry`:

| key | i18n ключ |
|-----|-----------|
| metal_mine | METAL_MINE_FULL_DESC |
| silicon_lab | SILICON_LAB_FULL_DESC |
| hydrogen_lab | HYDROGEN_LAB_FULL_DESC |
| solar_plant | SOLAR_PLANT_FULL_DESC |
| hydrogen_plant | HYDROGEN_PLANT_FULL_DESC |
| robotic_factory | ROBOTIC_FACTORY_FULL_DESC |
| nano_factory | NANO_FACTORY_FULL_DESC |
| shipyard | SHIPYARD_FULL_DESC |
| metal_storage | METAL_STORAGE_FULL_DESC |
| research_lab | RESEARCH_LAB_FULL_DESC |
| missile_silo | (нет в ru.yml — написать своё) |
| repair_factory | (нет в ru.yml — написать своё) |
| moon_base | MOON_BASE_FULL_DESC |
| moon_robotic_factory | MOON_ROBOTIC_FACTORY_FULL_DESC |

### Исследования — добавить `fullDesc` в `ResearchEntry`:

| key | i18n ключ |
|-----|-----------|
| spyware | SPYWARE_FULL_DESC |
| computer_tech | COMPUTER_TECH_FULL_DESC |
| gun_tech | GUN_TECH_FULL_DESC |
| shield_tech | SHIELD_TECH_FULL_DESC |
| shell_tech | SHELL_TECH_FULL_DESC |
| energy_tech | ENERGY_TECH_FULL_DESC |
| hyperspace_tech | HYPERSPACE_TECH_FULL_DESC |
| combustion_engine | COMBUSTION_ENGINE_FULL_DESC |
| impulse_engine | IMPULSE_ENGINE_FULL_DESC |
| hyperspace_engine | HYPERSPACE_ENGINE_FULL_DESC |
| laser_tech | LASER_TECH_FULL_DESC |
| ion_tech | ION_TECH_FULL_DESC |
| plasma_tech | PLASMA_TECH_FULL_DESC |
| expo_tech | EXPO_TECH_FULL_DESC |
| ballistics_tech | BALLISTICS_TECH_FULL_DESC |
| masking_tech | MASKING_TECH_FULL_DESC |

**Отображение в модале:** сворачиваемый блок под таблицей (`<details><summary>Подробнее</summary>...`). Текст может быть длинным (SPYWARE_FULL_DESC — 5 абзацев).

---

## Задача 2 — Производство по уровням для шахт/электростанций [P1]

В BuildingInfoModal добавить колонку «Производство» для зданий с добычей/энергией.

### Формулы из legacy (economy/production.go):

```
metal_mine:     base_rate_per_hour * level * 1.1^level  (из buildings.yml: base_rate=30)
silicon_lab:    base_rate_per_hour * level * 1.1^level  (base_rate=20)
hydrogen_lab:   base_rate_per_hour * level * 1.1^level  (base_rate=10)
solar_plant:    energy_output_per_level * level * 1.1^level  (out=20)
hydrogen_plant: энергия = floor(22.5 * level * 1.1^level)
```

Показывать в таблице как дополнительную колонку «⚡/ч» или «🟠/ч» рядом со стоимостью.

---

## Задача 3 — Пререквизиты в модале [P2]

Показать в модале зданий и исследований список требований из `catalog.ts` (поле `requires` у ResearchEntry).

Для зданий — из `configs/requirements.yml` (уже загружены через бэкенд в `requirements_unmet`).
Для исследований — из `r.requires` в catalog.ts.

Формат: `🔒 Требуется: Верфь ур.2, Импульсный двигатель ур.3`

---

## Файлы для изменения

| Файл | Изменение |
|------|-----------|
| `frontend/src/api/catalog.ts` | +`fullDesc` в BuildingEntry/ResearchEntry, заполнить для всех юнитов |
| `frontend/src/features/buildings/BuildingInfoModal.tsx` | +fullDesc секция, +production колонка, +requires |
| `frontend/src/features/research/ResearchInfoModal.tsx` | +fullDesc секция, +requires |

---

---

## Задача 4 — Реальное время строительства с учётом роботов/нанитов [P2]

В легаси таблица времени в ConstructionInfo показывает **реальное время** текущей планеты
с учётом уровня фабрики роботов и нано-фабрики. Формула (из `NS::getBuildingTime`):

```
time = (metal + silicon) / 2500.0 / (robo_level + 1) / 2^nano_level / PLANET_CONSTRUCTION_SPEED_FACTOR / 2^(build_factor-1)
```

Для исследований:
```
time = (metal + silicon) / 1000.0 / (1 + lab_level) / RESEARCH_SPEED_FACTOR / 2^(research_factor-1)
```

Сноска «Время указано без учёта фабрики роботов и нано-фабрики» должна исчезнуть.

### Реализация

**Backend** — новый endpoint или расширение существующего:

Вариант: `GET /api/planets/{id}/buildings/time-table?unit_id={id}` возвращает массив
`[{level, secs}, ...]` для 10 уровней вокруг текущего, посчитанный с реальным robo/nano планеты.

Или проще — добавить в BuildingInfoModal данные через уже существующий `build_seconds`
из `/api/planets/{id}/buildings` и `research_seconds` из `/api/research`:
- `build_seconds[unit_id]` = время **следующего** уровня с учётом роботов.
- Нам нужно время для произвольного уровня → нужно передавать robo/nano на фронт
  и считать прямо в TS.

**Предпочтительный вариант (только фронтенд)**:
Добавить в ответ `/api/planets/{id}/buildings` поля:
- `robotic_factory_level: int`
- `nano_factory_level: int`

Тогда `BuildingInfoModal` получает эти уровни через проп и считает время для каждого уровня
по той же формуле что и Go-бэкенд.

Аналогично для исследований — передавать `lab_level` в `ResearchInfoModal`.

### Файлы
| Файл | Изменение |
|------|-----------|
| `backend/internal/building/handler.go` | добавить `robotic_factory_level` и `nano_factory_level` в ответ `/buildings` |
| `frontend/src/features/buildings/BuildingInfoModal.tsx` | принять robo/nano как пропы, пересчитать `buildTimeSecs`, убрать сноску |
| `frontend/src/features/buildings/BuildingsScreen.tsx` | передать robo/nano в `BuildingInfoModal` |
| `frontend/src/features/research/ResearchInfoModal.tsx` | принять lab_level, пересчитать `researchTimeSecs`, убрать сноску |
| `frontend/src/features/research/ResearchScreen.tsx` | передать lab_level в `ResearchInfoModal` |

---

## Намеренно не реализуем

- Снос через модал (P3, отдельная задача 9 плана 09)
- VIP-ускорение за кредиты
- Chart с графиком (достаточно таблицы)
