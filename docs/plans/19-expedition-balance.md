# План 19: Баланс экспедиций — expo_tech + масштабирование пиратов

## Проблема

Два независимых пробела в механике экспедиций (`backend/internal/fleet/expedition.go`):

### 19.1 expo_tech не влияет на экспедиции

Исследование `expo_tech` (id=27) существует, стоит ресурсы, прокачивается —
но в `expedition.go` нигде не читается и не используется. Игрок качает бесполезное
исследование.

В legacy (`Expedition.class.php`) `expo_tech` влияет на:
- `exp_power` — суммарную «мощь» экспедиции, от которой зависят вероятности исходов
- Ресурсный множитель `res_scale`
- Шансы на артефакт и корабли

Формулы из legacy (см. `docs/legacy-game-reference.md`, §Мощь и вероятности):
```
exp_power = expo_tech_level
          + expedition_hours × 2
          + spy_tech / 10 × pow(spy_probes, 0.4)

resourceDiscovery   = ceil(100 × 1.22^exp_power)
artefactDiscovery   = ceil(3   × 1.28^exp_power)   [hours >= 4]
```

**Упрощённая реализация** (без полного портирования вероятностной модели legacy):
```go
expoLevel := readResearchLevel(ctx, tx, ownerUserID, 27)
// Увеличить вероятность resources и artefact пропорционально уровню
resourceWeight += expoLevel * 2   // level 5 → +10% к весу resources
artefactWeight += expoLevel       // level 5 → +5%
// Ресурсный множитель
fraction := baseFraction * (1.0 + float64(expoLevel)*0.05)
```

### 19.2 Пираты фиксированные — 5 light_fighter для любого флота

Средний игрок с 10+ крейсерами уничтожит пиратов без потерь — исход «pirates»
не несёт риска. Нужно масштабировать по силе флота игрока.

Формула:
```go
// Сила пиратов ~ 5–15% от суммарной атаки флота игрока
piratePower := calcFleetAttack(ships, cat) * (0.05 + r.Float64()*0.10)
pirateCount := int64(piratePower / 50)  // 50 = атака 1 LF
pirateCount  = max(3, min(500, pirateCount))
```

## Шаги реализации

### Шаг 1 — expo_tech: читать уровень исследования

В `expedition.go` найти функцию расчёта исхода. Добавить запрос уровня `expo_tech`
(id=27) из таблицы `researches` для текущего владельца флота.

```go
var expoLevel int
_ = tx.QueryRow(ctx,
    `SELECT level FROM researches WHERE user_id=$1 AND unit_id=27`,
    ownerUserID,
).Scan(&expoLevel)
```

### Шаг 2 — expo_tech: применить к весам и ресурсному множителю

Найти таблицу весов исходов в `expedition.go` (`resources`, `artefact`, `pirates`…).
Применить модификаторы из §19.1.

Не трогать legacy-константы вероятностей — только добавить поправку сверху.

### Шаг 3 — Пираты: calcFleetAttack

Добавить вспомогательную функцию `calcFleetAttack(ships map[int]int64, cat *config.Catalog) float64`:
суммирует `spec.Attack × count` для всех юнитов флота.

### Шаг 4 — Пираты: применить масштабирование

В ветке `pirates` заменить фиксированные `5 × light_fighter` на формулу из §19.2.
Использовать `cat.Ships.Ships` для поиска `light_fighter` по ключу.

### Шаг 5 — Тесты

- Тест: при expoLevel=0 результат не хуже текущего (нет регрессии)
- Тест: при expoLevel=10 вес `resources` > вес при expoLevel=0
- Тест: `calcFleetAttack` для известного флота возвращает ожидаемое значение
- Тест: pirateCount при слабом флоте = 3 (min), при сильном растёт

## Что НЕ меняется

- Вероятности остальных исходов (artefact, extra_planet, loss, nothing) — без изменений
- Механика самого боя с пиратами — без изменений
- Ресурсный исход: формула бонуса от cargo — без изменений (отдельный вопрос)

## Проверка готовности

- [ ] `expo_tech` уровень читается из БД в expedition.go
- [ ] При expoLevel > 0 вес `resources` и `artefact` выше
- [ ] Пираты масштабируются по силе флота (min 3, max 500 LF-эквивалентов)
- [ ] `make test` зелёный
- [ ] Обновить §8 и §13.4 в `balance-analysis.md`
