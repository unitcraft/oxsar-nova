# ADR-0005: Astrophysics (ASTRO_TECH) — новая технология (отклонение от legacy)

- Status: Accepted
- Date: 2026-04-25
- План: [docs/plans/20-legacy-port.md](../plans/20-legacy-port.md) Ф.7

## Context

В legacy oxsar2 нет технологии Astrophysics. Колонизация ограничена
только `computer_tech` (max планет = computer_tech + 1), а количество
одновременных экспедиций не ограничено вовсе.

Это создаёт две проблемы для nova:

1. **Эксплойт экспедиций уже частично закрыт** через min_fleet+lost+cap
   ([balance/audit.md](../balance/audit.md) BA-003 + план 21 блок B),
   но игрок всё ещё может слать **N экспедиций параллельно** с разных
   координат, если у него есть computer-слоты. Лимит на одновременные
   экспедиции — стандартная OGame-механика, которую legacy не имеет.
2. **Колонизация безлимитна по сути** — игрок может довести
   `computer_tech` до 10 и иметь 11 планет. В OGame с 2010 это
   ограничено через Astrophysics, чтобы лимит планет не зависел от
   слотов флота (computer_tech = слоты флота, astro = лимит планет).

Spec §12.5 oxsar-spec.txt описывает Astrophysics как новую технологию,
которая **должна** быть в nova.

## Decision

Добавить новую технологию `astro_tech` (id=112, mode=research) с
формулами:

```
colony_limit       = max(computer_tech + 1, floor(astro_level / 2) + 1)
expedition_slots   = max(1, floor(sqrt(astro_level)))
```

### Ключевые отклонения от плана 20 Ф.7

План 20 Ф.7 предлагал **жёстко** заменить `computer_tech+1` на
`astro/2+1`. Это ломало бы существующих игроков с прокачанным
computer_tech (потеряли бы 2-3 слота колоний). Мы используем
**MAX**: игроки с computer_tech ничего не теряют, astro даёт
**дополнительные** колонии поверх.

План также предлагал миграцией дать всем `astro_level=2`. Мы делаем
это **только для существующих игроков** (миграция 0061), новые
регистрации получают `astro=0` — поведение совпадает с прежним
(одна стартовая планета, один слот экспедиций как `max(1, sqrt(0))`).

### Конкретные правки

1. **`configs/research.yml`**: добавлен `astro_tech` (id=112,
   cost_factor=1.75 как у expo_tech).
2. **`configs/units.yml`**: `astro_tech` в группе research.
3. **`configs/construction.yml`**: mode=2, basic
   metal=6000/silicon=12000/hydrogen=6000, ×1.75/level
   (×1.5 от `expo_tech`, дороже базовой технологии).
4. **`configs/requirements.yml`**: `research_lab >= 4` и
   `expo_tech >= 3`.
5. **Migration 0061**: `INSERT INTO research SELECT u.id, 112, 2 …` —
   только для существующих игроков, у которых записи ещё нет.
6. **`fleet/colonize.go::expExtraPlanet`**: лимит = MAX(computer+1,
   astro/2+1).
7. **`fleet/transport.go::checkExpeditionSlots`**: новая проверка
   при mission=15 — COUNT outbound expeditions ≥
   `max(1, floor(sqrt(astro)))` → `ErrExpeditionSlotsFull` (409).

### Поведение для разных игроков

| astro | computer_tech | colony_limit | expedition_slots |
|---:|---:|---:|---:|
| 0 (новичок) | 0 | 1 | 1 |
| 0 (новичок) | 1 | 2 | 1 |
| 2 (существующий после миграции) | 0 | 2 | 1 |
| 2 | 3 (старый игрок) | 4 | 1 |
| 4 | 0 | 3 | 2 |
| 9 | 0 | 5 | 3 |
| 16 | 0 | 9 | 4 |

## Consequences

### Плюсы

- **Не ломает существующих игроков** — мигрируются с astro=2, лимиты
  только растут.
- **Новые игроки** получают тот же опыт что и раньше (одна планета,
  одна экспедиция). Только теперь у них **второй путь прокачки**:
  computer_tech (даёт слоты флота + колонии) или astro_tech (даёт
  колонии + параллельные экспедиции).
- **Закрывает эксплойт паралельных экспедиций** — без astro игрок
  может слать только одну экспедицию за раз. Это многократно усиливает
  фикс BA-003.

### Минусы

- **Отклонение от legacy** — в oxsar2 Astrophysics не было.
- **Новый ресурс-sink** — игроки тратят M+Si+H на технологию, которой
  не было раньше.

### Альтернативы отвергнуты

1. **Не добавлять astro вообще.** Эксплойт паралельных экспедиций
   остаётся, прогрессия колоний скучная.
2. **Жёстко заменить computer на astro** (как в плане 20 Ф.7). Ломает
   существующих игроков; ADR-кризис.
3. **Лимит экспедиций через computer_tech** (как fleet slots). Смешивает
   две разные механики — fleet slots уже отдельно (план 20 Ф.2).

## Проверка

```bash
cd backend
go test -count=1 ./internal/fleet/ ./internal/planet/
# Должны пройти.
```

Manual smoke:
- Новый игрок: try POST /api/fleet mission=15 — успех (1/1 slot).
  Try ещё одну параллельно — `ErrExpeditionSlotsFull` 409.
- Существующий игрок (после миграции): astro=2 → 1 слот, 2 колонии.
- После прокачки astro до 4 → 2 слота, 3 колонии.

## Откат

```sql
-- Down-migration в 0061_astro_tech_starter.sql:
DELETE FROM research WHERE unit_id = 112;
```

Затем удалить astro_tech из configs/research.yml, units.yml,
construction.yml, requirements.yml. Откатить код в colonize.go,
expedition.go, transport.go.
