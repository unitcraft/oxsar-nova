# План 13: Реальное время строительства/исследований в модалах

## Статус: ПЛАНИРУЕТСЯ

## Контекст

В легаси `ConstructionInfo/{id}` таблица времени показывает **реальное время текущей планеты**
с учётом уровня фабрики роботов и нано-фабрики.

В nova `BuildingInfoModal` и `ResearchInfoModal` показывают базовое время без ускорений
и внизу стоит сноска-предупреждение («Время указано без учёта фабрики роботов и нано-фабрики»).
Задача — убрать сноску и показывать реальное время как в легаси.

## Формулы (из `NS::getBuildingTime` в `www/game/NS.class.php`)

**Здания:**
```
time_hours = (metal + silicon) / 2500.0 / (robo_level + 1) / 2^nano_level
           / PLANET_CONSTRUCTION_SPEED_FACTOR / 2^(build_factor - 1)
```

**Луна:**
```
time_hours = (metal + silicon) / 2500.0 / (moon_robo_level + 1) / 2^nano_level
           / MOON_CONSTRUCTION_SPEED_FACTOR
```

**Исследования:**
```
time_hours = (metal + silicon) / 1000.0 / (1 + lab_level)
           / RESEARCH_SPEED_FACTOR / 2^(research_factor - 1)
```

Все три формулы уже портированы в `backend/internal/economy/build_duration.go`.
`PLANET_CONSTRUCTION_SPEED_FACTOR` и `RESEARCH_SPEED_FACTOR` = 1.0 в нашей конфигурации.
`build_factor` и `research_factor` — бонусы офицера (уже есть в `Planet`).

## Реализация

### Backend — минимальные изменения

Добавить в ответ `GET /api/planets/{id}/buildings` три поля:

```json
{
  "levels": {...},
  "build_seconds": {...},
  "requirements_unmet": {...},
  "robotic_factory_level": 3,
  "nano_factory_level": 0,
  "moon_robotic_factory_level": 0
}
```

Эти уровни уже читаются внутри `building/service.go` при расчёте `build_seconds` —
нужно просто добавить их в структуру ответа.

Для исследований добавить в ответ `GET /api/research`:

```json
{
  "levels": {...},
  "queue": [...],
  "research_seconds": {...},
  "research_lab_level": 5
}
```

### Frontend — пересчёт в модале

`BuildingInfoModal` получает новые пропы `roboLevel: number` и `nanoLevel: number`.
`buildTimeSecs` пересчитывается по формуле легаси вместо приближения:

```ts
function buildTimeSecs(b: BuildingEntry, level: number, roboLevel: number, nanoLevel: number): number {
  const cost = costForLevel(b.costBase, b.costFactor, level);
  const hours = (cost.metal + cost.silicon) / 2500.0 / (roboLevel + 1) / Math.pow(2, nanoLevel);
  return Math.round(hours * 3600);
}
```

`ResearchInfoModal` получает проп `labLevel: number`:

```ts
function researchTimeSecs(r: ResearchEntry, level: number, labLevel: number): number {
  const cost = costForLevel(r.costBase, r.costFactor, level);
  const hours = (cost.metal + cost.silicon) / 1000.0 / (1 + labLevel);
  return Math.round(hours * 3600);
}
```

Сноски удаляются.

## Файлы для изменения

| Файл | Изменение |
|------|-----------|
| `backend/internal/building/handler.go` | `+robotic_factory_level`, `+nano_factory_level`, `+moon_robotic_factory_level` в ответ `/buildings` |
| `backend/internal/research/handler.go` | `+research_lab_level` в ответ `/research` |
| `frontend/src/features/buildings/BuildingInfoModal.tsx` | пропы `roboLevel`/`nanoLevel`, формула легаси, убрать сноску |
| `frontend/src/features/buildings/BuildingsScreen.tsx` | передать `roboLevel`/`nanoLevel` из ответа в модал |
| `frontend/src/features/research/ResearchInfoModal.tsx` | проп `labLevel`, формула легаси, убрать сноску |
| `frontend/src/features/research/ResearchScreen.tsx` | передать `labLevel` из ответа в модал |

## Проверка

- Открыть модал рудника на планете с роботами ур.5 — время должно совпадать с легаси `/game.php/ConstructionInfo/1`
- Открыть модал исследования на планете с лабораторией ур.7 — сравнить с `/game.php/Research`
- Убедиться что сноска исчезла в обоих модалах
