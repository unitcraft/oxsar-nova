# Аудит configs/ (2026-04-24)

Проверка актуальности YAML-справочников. Вывод: в целом согласованы,
но есть несколько дыр и 3 файла, которые можно убрать из репо.

---

## Что грузится runtime'ом

`backend/internal/config/catalog.go` → `LoadCatalog()` читает **10 файлов**:

| Файл | Размер | Назначение |
|------|--------|------------|
| `units.yml` | 5 КБ | Центральный реестр ID + key + name (buildings/moon/research/fleet/defense) |
| `buildings.yml` | 5 КБ | Баланс зданий (prod/cons, cost, requires) |
| `research.yml` | 1 КБ | Баланс исследований |
| `ships.yml` | 4 КБ | Баланс кораблей |
| `defense.yml` | 1 КБ | Баланс оборонных юнитов |
| `rapidfire.yml` | 1 КБ | Таблица rapidfire |
| `requirements.yml` | 6 КБ | Зависимости (что нужно построить/изучить) |
| `artefacts.yml` | 4 КБ | Артефакты |
| `professions.yml` | 1 КБ | 4 профессии (бонусы/штрафы) |
| `construction.yml` | **24 КБ** | Legacy-формулы na_construction (`ConstructionCatalog`, §1.4 ТЗ) |

---

## Что **не** грузится runtime'ом

`*_generated.yml` — одноразовый выхлоп CLI `backend/cmd/tools/import-datasheets`,
нужен был для первичного импорта из legacy-дампов. Сейчас **никто не читает**:

| Файл | Размер | Что дублирует |
|------|--------|---------------|
| `ships_generated.yml` | 3 КБ | частично `ships.yml` (генератор использует unit_id, не key) |
| `requirements_generated.yml` | 6 КБ | частично `requirements.yml` (формат `needs/level` вместо `kind/key`) |
| `artefacts_meta_generated.yml` | 3 КБ | метаданные артефактов (не используется даже CLI после первой генерации) |

**Проверено**: `grep -rE "_generated" backend/ frontend/` — только `import-datasheets/*.go` пишет в них, но не читает обратно.

---

## Найденные дыры

### 1. Defense.yml неполон

Содержит id **43–50** (8 юнитов). В `units.yml defense` 12 записей:
- **51** `interceptor_rocket`, **52** `interplanetary_rocket` — в `ships.yml` уже есть (ракеты у нас идут через `fleet` с `attack=0`, это межпланетные снаряды из плана 05, kind=16). Формально зарегистрированы как defense в units.yml, но фактически работают через ships.yml → рассинхрон реестра с реализацией.
- **354** `small_planet_shield`, **355** `large_planet_shield` — **нигде не реализованы** (ни в ships, ни в defense, ни в requirements, ни в backend-коде).

### 2. Research.yml пропускает 2 id

- **26** `ign` (Alliance Network) — есть в units.yml, отсутствует в research.yml и requirements.yml. **В коде не ищется**.
- **28** `gravi` (Graviton Tech) — то же самое.

Это legacy-технологии. У них нет ни cost, ни эффектов в коде. Нужно либо
вычеркнуть из `units.yml` (если не планируем), либо реализовать с минимум
cost+effect.

### 3. Requirements.yml неполон

Отсутствуют зависимости для юнитов, которые реально есть в `ships.yml`:

| Юнит | ID | ships.yml | requirements.yml |
|------|:--:|:---------:|:----------------:|
| `star_destroyer` | 41 | ✅ | ❌ |
| `shadow_ship` | 325 | ✅ | ❌ |
| `ion_gun` | 46 | defense ✅ | ❌ |

Практический эффект: игрок может построить юнит без reqs-check, потому что
`requirements.Check(key)` вернёт `nil` (юнит отсутствует в таблице → требований
нет). Для `ion_gun` это значит «строй сразу на любом уровне шиполверфи». Это
бажок.

### 4. Мёртвые id в units.yml

`shadow_ship` (325), `star_destroyer` (41) объявлены в реестре, есть статы
в ships.yml, но **не используются в backend/frontend коде** (0 ссылок по
`grep`). Два варианта:
- это готовые к использованию, но не интегрированные юниты (включить в тесты/sim)
- это «заложенный на будущее» resource, можно временно убрать в комментарий

---

## Рекомендации

### Быстрые (низкий риск)

1. **Добавить `*_generated.yml` в `.gitignore`** и удалить из репо. Это 12 КБ
   устаревших дампов, которые никто не читает. Воспроизводимы через
   `make import-datasheets` из legacy-дампа.

2. **Добавить requirements для `ion_gun`, `star_destroyer`, `shadow_ship`**.
   Например, `ion_gun` требует `shipyard level 4 + ion_tech level 1`.

3. **Добавить в `LoadCatalog` warn-лог**, если какой-то id из `units.yml`
   отсутствует в соответствующем балансном файле. Поможет ловить такие
   рассинхроны на старте процесса.

### Средние (нужно решение геймдизайна)

4. **Planet shields (354, 355)** — решить: реализовать или выкинуть из
   `units.yml`. Сейчас это мёртвый регистровый шум.

5. **Research ign (26) и gravi (28)** — то же самое: реализовать с
   конкретным эффектом или удалить.

6. **Ракеты в defense** — определиться: `interceptor_rocket` и
   `interplanetary_rocket` сейчас работают как fleet (через ships.yml),
   но registered как defense в units.yml. Правильнее — перенести их в
   `fleet` группу `units.yml`, чтобы реестр не врал. Но это меняет публичные
   id-категории, требует ADR.

### Крупные

7. **Валидатор конфигов как separate test**. Добавить `go test ./backend/internal/config/` с проверками:
   - все id из units.yml есть в своём балансном файле (buildings/research/ships/defense)
   - все key в requirements.yml ссылаются на существующие юниты
   - все from/to в rapidfire.yml ссылаются на существующие ship/defense id
   
   Одна-две сотни строк, дешевле чем отлов багов в рантайме.

---

## Что НЕ рекомендую

- **Не удалять `construction.yml`** (24 КБ). Это источник истины legacy-формул
  из `na_construction` (§1.4 ТЗ). Даже если runtime сейчас больше полагается
  на `buildings.yml`, `ConstructionCatalog` — это fallback/reference.

- **Не пытаться автоматически синхронизировать `units.yml` с балансными
  файлами через генератор**. Сначала решить судьбу orphan-id (ign/gravi/
  planet_shields), иначе генератор затянет их как пустышки.
