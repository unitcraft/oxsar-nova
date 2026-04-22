# План реализации battle_bonus

## Контекст

`battle_bonus` — тип эффекта артефакта, дающий боевой бонус (атака/щит/броня)
на всё время активности артефакта. Сейчас возвращает `ErrUnsupported` в
`effects.go`. Реализация разблокирована после завершения M4 (боевой движок готов).

Данные из legacy: `game/Artefact.class.php` + `consts.php` (§5.10.1 ТЗ).

**Legacy UI:** http://localhost:8080/login.php (test / quoYaMe1wHo4xaci) → `d:\Sources\oxsar2\www\templates\standard\*.tpl`

---

## Шаги

### 1. `backend/internal/config/catalog.go` — расширить `ArtefactEffect`

Добавить поля только для `battle_bonus`:

```go
type ArtefactEffect struct {
    // существующие поля...
    Type          string  `yaml:"type"`
    Field         string  `yaml:"field,omitempty"`
    Op            string  `yaml:"op,omitempty"`
    Value         float64 `yaml:"value,omitempty"`
    ActiveValue   float64 `yaml:"active_value,omitempty"`
    InactiveValue float64 `yaml:"inactive_value,omitempty"`

    // только для type=battle_bonus:
    BattleAttack float64 `yaml:"battle_attack,omitempty"` // множитель атаки, напр. 1.15
    BattleShield float64 `yaml:"battle_shield,omitempty"` // множитель щита
    BattleShell  float64 `yaml:"battle_shell,omitempty"`  // множитель брони
}
```

### 2. `backend/internal/artefact/effects.go` — добавить `BattleModifier`

Новый тип и функция (без IO):

```go
type BattleModifier struct {
    AttackMul float64 // итоговый множитель атаки (1.0 = без изменений)
    ShieldMul float64
    ShellMul  float64
}

// ComputeBattleModifier суммирует эффекты нескольких активных battle_bonus
// артефактов в один итоговый множитель.
func ComputeBattleModifier(specs []config.ArtefactSpec) BattleModifier {
    m := BattleModifier{AttackMul: 1, ShieldMul: 1, ShellMul: 1}
    for _, s := range specs {
        if s.Effect.Type != "battle_bonus" {
            continue
        }
        if s.Effect.BattleAttack != 0 {
            m.AttackMul *= s.Effect.BattleAttack
        }
        if s.Effect.BattleShield != 0 {
            m.ShieldMul *= s.Effect.BattleShield
        }
        if s.Effect.BattleShell != 0 {
            m.ShellMul *= s.Effect.BattleShell
        }
    }
    return m
}
```

В `computeChanges` убрать `battle_bonus` из `ErrUnsupported` — вернуть `nil, nil`
(эффект не материализуется в БД, применяется in-memory во время боя).

### 3. `backend/internal/artefact/service.go` — `ActiveBattleModifiers`

Запрос активных battle_bonus артефактов пользователя внутри транзакции:

```go
func (s *Service) ActiveBattleModifiers(ctx context.Context, tx pgx.Tx, userID string) (BattleModifier, error) {
    rows, err := tx.Query(ctx, `
        SELECT unit_id FROM artefacts_user
        WHERE user_id = $1 AND state = 'active'
    `, userID)
    // ... собрать unit_id → spec из каталога, filter type=battle_bonus
    // вернуть ComputeBattleModifier(specs)
}
```

### 4. `backend/internal/fleet/attack.go` — применить модификаторы

После `readUserTech` для атакующего:

```go
battleMod, err := artefactSvc.ActiveBattleModifiers(ctx, tx, attackerUserID)
// ...
atkSide := battle.Side{
    UserID: attackerUserID,
    Tech:   attackerTech,
    Units:  applyBattleMod(stacksToBattleUnits(attackerShips, s.catalog, false), battleMod),
}
```

Новая функция `applyBattleMod`:

```go
func applyBattleMod(units []battle.Unit, m artefact.BattleModifier) []battle.Unit {
    for i := range units {
        for ch := range units[i].Attack {
            units[i].Attack[ch] *= m.AttackMul
        }
        for ch := range units[i].Shield {
            units[i].Shield[ch] *= m.ShieldMul
        }
        units[i].Shell *= m.ShellMul
    }
    return units
}
```

`TransportService` нужно расширить полем `artefacts *artefact.Service`.

### 5. `configs/artefacts.yml` — добавить battle_bonus артефакты

Три артефакта из legacy (ID из `consts.php`):

```yaml
  war_machine:        # id из legacy
    id: 310
    name: "War Machine"
    effect:
      type: battle_bonus
      battle_attack: 1.1    # +10% атака
    stackable: true
    lifetime_seconds: 604800

  energy_shield:
    id: 311
    name: "Energy Shield"
    effect:
      type: battle_bonus
      battle_shield: 1.1    # +10% щит
    stackable: true
    lifetime_seconds: 604800

  titanium_hull:
    id: 312
    name: "Titanium Hull"
    effect:
      type: battle_bonus
      battle_shell: 1.1     # +10% броня
    stackable: true
    lifetime_seconds: 604800
```

> **Важно:** ID и множители нужно уточнить по `d:\Sources\oxsar2\consts.php`
> и `game/Artefact.class.php` перед мержем. Не менять баланс без ADR.

### 6. `backend/internal/artefact/effects_test.go` — обновить тест

`TestComputeChanges_BattleBonus` — сейчас ожидает `ErrUnsupported`.
После шага 2 должен ожидать `nil, nil`.

Добавить тест на `ComputeBattleModifier`:
- нет артефактов → все множители 1.0
- один артефакт battle_attack 1.1 → AttackMul=1.1, остальные 1.0
- два стакающихся → AttackMul перемножается

---

## Зависимости / блокеры

- `TransportService` должен получить `*artefact.Service` — нужно передать в конструктор в `cmd/worker`.
- ID артефактов из legacy надо верифицировать по `oxsar2/consts.php`.
- Защитник не получает battle_bonus (артефакт работает только на флот атакующего) — подтвердить по legacy.

## Что НЕ входит в этот план

- `one_shot` артефакты — отдельная задача.
- ACS (несколько атакующих) — M5.
- UI разблокировки «Артефакты на миссию» — после реализации бэкенда.
