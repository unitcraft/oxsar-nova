# План 18: Боевые технологии — применение модификаторов в бою

## Проблема

Gun/Shield/Shell tech прокачиваются за ресурсы и время, но в бою **ничего не делают**.
`Tech` передаётся в `battle.Side`, используются только `Ballistics` и `Masking` —
поля `Gun`, `Shield`, `Shell` нигде не применяются.

Игрок качает Weapons Tech level 5 за ~12 800 металла + часы реального времени,
ожидая +50% атаки — получает ничего. Это серьёзный обман ожиданий.

Источник истины: oxsar2-java `assault/` — там tech-модификаторы применяются
к `Attack`, `Shield`, `Shell` каждого юнита перед боем.

## Формулы (legacy)

```
effectiveAttack = baseAttack × (1 + gunTech × 0.10)
effectiveShield = baseShield × (1 + shieldTech × 0.10)
effectiveShell  = baseShell  × (1 + shellTech × 0.10)
```

+10% за уровень — стандарт OGame, подтверждён legacy-кодом.

## Шаги реализации

### Шаг 1 — Найти точку применения в engine.go

Файл: `backend/internal/battle/engine.go`.

Tech-модификаторы нужно применять при инициализации `unitState` из `battle.Unit`,
**до первого раунда** — не в каждом раунде (статический множитель).

Искать функцию, которая строит начальный `unitState` из `Side.Units`.

### Шаг 2 — Применить модификаторы

```go
// при построении unitState из u (battle.Unit) и side.Tech:
gunFactor    := 1.0 + float64(side.Tech.Gun)*0.10
shieldFactor := 1.0 + float64(side.Tech.Shield)*0.10
shellFactor  := 1.0 + float64(side.Tech.Shell)*0.10

// Attack — массив [3]float64 (каналы)
for i := range u.Attack {
    st.attack[i] = u.Attack[i] * gunFactor
}
st.shield     = u.Shield * shieldFactor   // аналогично если Shield — массив
st.shellTotal = u.Shell * shellFactor * float64(u.Quantity)
```

Точная сигнатура зависит от внутренних типов `unitState` — уточнить при реализации.

### Шаг 3 — Убедиться что Tech передаётся в Side корректно

Найти вызовы `battle.Calculate` в `fleet/attack.go` и `alien/helpers.go`.
Проверить, что `Side.Tech.Gun/Shield/Shell` заполняются из реального уровня
исследований игрока (запрос к БД).

Искать: `battle.Side{`, `Tech:`.

### Шаг 4 — Тесты

Добавить юнит-тест в `backend/internal/battle/engine_test.go`:

```go
// При gun_tech=5: атака должна быть на 50% выше базовой
// Проверить что юнит с tech.Gun=5 наносит в 1.5× больше урона
// чем тот же юнит с tech.Gun=0
```

Проверить что бой с tech=0 даёт результаты идентичные текущим (регрессия).

### Шаг 5 — Обновить §12 таблицы в balance-analysis.md

Строку «Боевые технологии | Структурно есть, но НЕ применяются» → «✅ применяются».

## Что НЕ меняется

- Ballistics/Masking — уже работают, не трогать
- Формулы rapidfire — не затронуты
- Конфиги — не меняются

## Проверка готовности

- [ ] `engine_test.go`: тест gun_tech=5 даёт +50% урона
- [ ] `engine_test.go`: тест shell_tech=5 даёт +50% брони
- [ ] Регрессия: бой с tech=0 идентичен текущему
- [ ] `fleet/attack.go`: `Side.Tech.Gun/Shield/Shell` заполняются из БД
- [ ] `make test` зелёный
