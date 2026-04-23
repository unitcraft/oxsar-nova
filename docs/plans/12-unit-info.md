# План 12: Детальная страница здания/исследования (ConstructionInfo)

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

## Задача 5 — Переход на отдельный экран вместо модального окна [P1]

### Проблема
`BuildingInfoModal` и `ResearchInfoModal` реализованы как модальные окна поверх страницы.
Из-за особенностей CSS (`display: flex` + `overflowY: auto` на flex-контейнере) текст
«Подробнее» при раскрытии вываливается за пределы белой панели и рендерится на тёмном
фоне оверлея. Несколько попыток исправить через `max-height`, `grid-template-rows`,
`overflow: hidden` — не дали результата.

### Решение — отдельный экран (как в legacy)
Legacy `ConstructionInfo/{id}` — это **отдельная страница**, не попап.
Делаем аналогично: новый таб `'unit-info'` в навигации nova.

### Реализация

**App.tsx:**
- Добавить `'unit-info'` в тип `Tab`
- Добавить state `infoUnit: { kind: 'building' | 'research'; id: number } | null`
- Рендерить `<UnitInfoScreen>` вместо модала

**UnitInfoScreen** (`frontend/src/features/unit-info/UnitInfoScreen.tsx`):
- Принимает `kind`, `unitId`, `currentLevel`, `onBack`
- Кнопка «← Назад» возвращает на предыдущий таб
- Обычная страница с нормальным скроллом, без фиксированных высот
- Секции: заголовок с иконкой, таблица уровней (+ колонка производства для шахт),
  пресреквизиты, полное описание (просто текст, без анимации)

**BuildingsScreen / ResearchScreen:**
- Клик на карточку/иконку вместо `setInfoUnitId` вызывает `onOpenInfo(kind, id)`
- Проп `onOpenInfo` пробрасывается из `App.tsx`

**Удалить:**
- `BuildingInfoModal.tsx`
- `ResearchInfoModal.tsx`

### Файлы для изменения

| Файл | Изменение |
|------|-----------|
| `frontend/src/App.tsx` | + Tab `'unit-info'`, state `infoUnit`, рендер `UnitInfoScreen` |
| `frontend/src/features/unit-info/UnitInfoScreen.tsx` | **новый файл** — единый экран для зданий и исследований |
| `frontend/src/features/buildings/BuildingsScreen.tsx` | проп `onOpenInfo`, убрать `infoUnitId` state и `BuildingInfoModal` |
| `frontend/src/features/research/ResearchScreen.tsx` | проп `onOpenInfo`, убрать `infoUnitId` state и `ResearchInfoModal` |
| `frontend/src/features/buildings/BuildingInfoModal.tsx` | удалить |
| `frontend/src/features/research/ResearchInfoModal.tsx` | удалить |

---

## Файлы уже изменены (задачи 1–4)

| Файл | Изменение |
|------|-----------|
| `frontend/src/api/catalog.ts` | +`fullDesc` в BuildingEntry/ResearchEntry для всех юнитов |

---

---

## Задача 6 — Улучшение заголовка UnitInfoScreen

- Убрать кнопку «← Назад» (навигация через стандартный таб-бар)
- Увеличить иконку до 128×128
- Вместо короткого описания (`description`/`benefit`) показывать полное (`fullDesc`)
- Убрать блок «Описание» снизу

---

## Задача 7 — Идентификатор юнита в URL

Формат хэша: `#unit-info/building/1` или `#unit-info/research/13`.

- `tabFromHash` парсит три сегмента и возвращает `Tab` + восстанавливает `infoUnit`
- `navigateTo('unit-info')` заменяется на `navigateToUnit(kind, id)` который пишет полный хэш
- При `popstate` парсить хэш и восстанавливать `infoUnit`
- При прямом открытии страницы с хэшем `#unit-info/building/1` экран отображается корректно (уровень берётся 0 если нет данных планеты)

---

## Намеренно не реализуем

- VIP-ускорение за кредиты
- Chart с графиком (достаточно таблицы)
