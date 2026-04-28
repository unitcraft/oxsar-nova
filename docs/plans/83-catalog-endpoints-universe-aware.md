# План 83: Universe-aware catalog-endpoints

**Дата**: 2026-04-28
**Статус**: Черновик. Запуск **после плана 74** (публичный запуск
origin-вселенной).
**Зависимости**:
- ✅ план 64 (override-схема балансов).
- ✅ план 72 Ф.4 Spring 3 (catalog-endpoints реализованы как
  current-universe-only — modern formulas).
- ⏳ план 74 (публичный запуск origin) — главный триггер.
**Связанные документы**:
- [план 72 §Backend-расширения по требованию](72-remaster-origin-frontend-pixel-perfect.md) —
  текущая реализация catalog-endpoints в Ф.4.
- [план 64](64-remaster-origin-yaml-override.md) — override-схема
  для per-universe балансов.

---

## Зачем

Сейчас (после Ф.4 плана 72) catalog-endpoints отдают **modern
параметры** (`internal/economy/formulas.go` + `configs/*.yml`).
Это корректно пока origin-вселенная не запущена.

Когда план 74 (публичный запуск origin) закроется — игроки начнут
играть в origin-вселенной. Их планеты используют **origin-формулы**
(`internal/origin/economy/*.go`), которые могут давать **другие
числа** для тех же зданий/юнитов (R0-параметризация через override-
схему плана 64).

Если catalog-endpoint к этому моменту останется current-universe-
only, origin-игрок откроет `/building/metal_mine` info-страницу и
увидит **modern** числа (cost/production), а на своей origin-планете
**другие** числа. Это **bug в восприятии**: UI обещает одно, реальные
вычисления дают другое.

План 83 устраняет этот разрыв.

---

## Что НЕ цель

- Не менять frontend-логику — он уже потребляет catalog-endpoints
  (план 72 Ф.4). После расширения backend frontend получает
  правильные числа без изменений.
- Не пересчитывать формулы — они уже разделены на
  `internal/economy/` (modern) и `internal/origin/economy/` (origin).
- Не вводить новые формулы — только routing по universe-context.
- Не дублировать YAML-каталоги — `configs/balance/origin.yaml`
  override-логика уже работает (план 64).

---

## Что делаем

### Universe-routing в catalog-handler'ах

Эталон — `projects/game-nova/backend/internal/balance/loader.go`
(план 64 deep-merge override). Catalog-handler:

1. **Определяет universe-context**:
   - Из query-param `?universe=origin` (приоритет — explicit).
   - ИЛИ из JWT (current-user → planets → universe_id первой
     планеты или текущей).
   - Default — modern (nova) если не определено.

2. **Выбирает формулы**:
   - `universe == 'origin'` → `internal/origin/economy/*.go`.
   - иначе → `internal/economy/formulas.go`.

3. **Выбирает params** через override-loader:
   - `universe == 'origin'` → `configs/buildings.yml` deep-merged с
     `configs/balance/origin.yaml`.
   - иначе — чистый `configs/buildings.yml`.

4. **Формирует pre-computed таблицу** — те же ключевые уровни
   (1, 5, 10, 20, max), но через выбранные формулы.

### Endpoint-протокол

```
GET /api/buildings/catalog/{type}?universe=origin
GET /api/units/catalog/{type}?universe=origin
GET /api/research/tree?universe=origin
GET /api/artefacts/catalog/{type}?universe=origin
```

Без `?universe` — fallback к JWT-context (current planet's universe)
или modern-default.

### Тесты

- Property-based: для каждой пары (modern, origin) проверить что
  catalog-endpoint возвращает именно те числа что выдаёт
  соответствующая formula-функция.
- Golden-test: catalog response для `metal_mine` в universe=origin
  совпадает с числами из `internal/origin/economy/MetalMineProduction`
  для тех же level/tech.
- Integration: запрос `?universe=origin` → response `metal_mine`
  level 5 cost = origin-cost; запрос без universe → modern-cost.

### Frontend (минимальные правки)

- В TanStack Query keys добавить universe в ключ:
  `['buildings', 'catalog', type, universe]` — иначе кэш будет
  смешивать modern/origin данные.
- В useQuery вызовах добавить `?universe=...` параметр (взять из
  current-planet context или Zustand-store).
- В origin-фронте (`projects/game-nova/frontends/origin/`) — просто
  всегда передавать `?universe=origin`.
- В nova-фронте (`projects/game-nova/frontends/nova/`) — определять
  по выбранной планете игрока (если есть universe-switcher).

---

## Этапы

- Ф.1. Universe-context resolver в catalog-handler'ах (общий
  middleware или helper).
- Ф.2. Расширение 4 catalog-endpoints (buildings, units, research,
  artefacts) routing по universe.
- Ф.3. Tests (property-based + golden + integration).
- Ф.4. Frontend: добавить universe в query-keys + параметр запроса.
- Ф.5. Финализация (шапка плана 83 ✅).

---

## Объём

~200-400 строк Go + ~50-100 строк frontend + тесты.
1-2 коммита. ~3-4 часа агента.

---

## Триггеры запуска

- ✅ план 74 (публичный запуск origin) закрыт, ИЛИ
- 🟡 жалоба от тестера/игрока «info-страница в origin показывает
  modern-числа» (5+ запросов).

До этого — план остаётся черновиком.

---

## Risk: что если override origin почти не отличается от modern

В origin-сейчас (по плану 64) override `configs/balance/origin.yaml`
содержит только **специфичные origin-параметры** (alien-юниты,
multipliers). Большинство building/unit/research-параметров —
identical с modern.

Если на момент запуска плана 83 это останется так — расширение
catalog имеет минимальную user-visible разницу. Это не
противопоказание: всё равно нужно сделать чтобы **архитектурно**
endpoint был корректным (universe-aware), даже если data-различие
0%. И на случай когда баланс origin отойдёт от modern (после
запуска любая правка `origin.yaml` → catalog должен подхватить).

---

## Связанные

- План 72 Ф.4 (Spring 3) — current implementation, модернизируется
  планом 83.
- План 64 (override-схема) — backend-инфра уже готова, переиспользуем.
- План 74 (публичный запуск) — главный триггер.
