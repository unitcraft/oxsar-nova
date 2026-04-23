# План A: Боевой движок — технологии и артефакты

---

## История (завершено)

### ✅ M4.1 Боевой движок — базовый порт (итерация 20–21)
- Портирован `Units.processAttack` shield-блок из oxsar2-java
- Реализован rapidfire, ablation (damaged units), ballistics/masking
- Battle report, loot, debris fields, moon-chance
- Тесты: property-based + golden файлы

### ✅ M4.4 Rapidfire — полная таблица (план 17, итерация ~50)
- Добавлены frigate, bomber, star_destroyer, lancer, plasma_turret
- Rapidfire против Deathstar (lancer ×3, plasma_turret ×2) взят из na_rapidfire
- Frigate → transporters/heavy fighter/cruiser/battleship
- Bomber → rocket launcher/light laser/heavy laser/ion cannon/lancer

### ✅ Боевые артефакты — battle_bonus (nil, nil) (план 01, итерация ~48)
- `computeChanges` возвращает `(nil, nil)` для battle_bonus — эффект применяется
  отдельно через `ComputeBattleModifier`, не через изменение полей планеты
- Тест `TestComputeChanges_BattleBonus` исправлен

---

## Открытые задачи

### A.1 Боевые технологии — применение модификаторов (приоритет: HIGH)

**Проблема:** Gun/Shield/Shell tech прокачиваются за ресурсы, но в бою **ничего не делают**.
`Side.Tech.Gun/Shield/Shell` нигде не применяются в `battle/engine.go`.
Игрок качает Weapons Tech 5 за ~12 800 металла — получает 0 эффекта.

**Формула (legacy oxsar2-java):**
```
effectiveAttack = baseAttack × (1 + gunTech × 0.10)
effectiveShield = baseShield × (1 + shieldTech × 0.10)
effectiveShell  = baseShell  × (1 + shellTech × 0.10)
```

**Шаг 1** — `backend/internal/battle/engine.go`: найти построение `unitState` из `battle.Unit`.
Применять tech-факторы **один раз при инициализации** (статически, не per-раунд):
```go
gunFactor    := 1.0 + float64(side.Tech.Gun)*0.10
shieldFactor := 1.0 + float64(side.Tech.Shield)*0.10
shellFactor  := 1.0 + float64(side.Tech.Shell)*0.10
for i := range u.Attack { st.attack[i] = u.Attack[i] * gunFactor }
st.shell = u.Shell * shellFactor * float64(u.Quantity)
```

**Шаг 2** — Проверить `fleet/attack.go` и `alien/helpers.go`: убедиться что
`Side.Tech.Gun/Shield/Shell` заполняются из БД (`research` table).

**Шаг 3** — Тесты в `backend/internal/battle/engine_test.go`:
- gun_tech=5 → урон в 1.5× больше чем tech=0
- shell_tech=5 → броня в 1.5× выше
- Регрессия: бой с tech=0 идентичен текущему

**Проверка готовности:**
- [ ] `engine_test.go`: gun_tech=5 даёт +50% урона
- [ ] `engine_test.go`: shell_tech=5 даёт +50% брони
- [ ] Регрессия: tech=0 идентичен текущему
- [ ] `fleet/attack.go` и `alien/helpers.go`: Gun/Shield/Shell заполняются из БД
- [ ] `make test` зелёный

---

### A.2 Артефакт battle_bonus — применение в бою (приоритет: MEDIUM)

**Проблема:** Артефакты war_machine/energy_shield/titanium_hull существуют в конфиге,
но не влияют на бой — `ComputeBattleModifier` есть, но не вызывается из `attack.go`.

**Шаг 1** — `backend/internal/config/catalog.go`, расширить `ArtefactEffect`:
```go
BattleAttack float64 `yaml:"battle_attack,omitempty"`
BattleShield float64 `yaml:"battle_shield,omitempty"`
BattleShell  float64 `yaml:"battle_shell,omitempty"`
```

**Шаг 2** — `backend/internal/artefact/effects.go`: убедиться что `ComputeBattleModifier` реализован.

**Шаг 3** — `backend/internal/artefact/service.go`:
```go
func (s *Service) ActiveBattleModifiers(ctx context.Context, tx pgx.Tx, userID string) (BattleModifier, error)
```

**Шаг 4** — `backend/internal/fleet/attack.go`: применить модификаторы к юнитам атакующего:
```go
battleMod, _ := artefactSvc.ActiveBattleModifiers(ctx, tx, attackerUserID)
atkSide := battle.Side{Units: applyBattleMod(stacksToBattleUnits(...), battleMod)}
```
`TransportService` получает поле `artefacts *artefact.Service`.

**Шаг 5** — `configs/artefacts.yml`: добавить battle_bonus артефакты.
ID и множители уточнить по `d:\Sources\oxsar2\consts.php` перед мержем.
Пример:
```yaml
war_machine:
  id: 310
  effect:
    type: battle_bonus
    battle_attack: 1.1
```

**Шаг 6** — Тесты `ComputeBattleModifier`: нет артефактов → все 1.0; один +10% → 1.1; два стакаются.

**Зависимость:** A.1 рекомендуется сделать первым (создаёт паттерн).

**Проверка готовности:**
- [ ] `ArtefactEffect.BattleAttack/Shield/Shell` в catalog.go
- [ ] `ComputeBattleModifier` и `ActiveBattleModifiers` реализованы
- [ ] `applyBattleMod` вызывается в `fleet/attack.go`
- [ ] ID артефактов верифицированы по `oxsar2/consts.php`
- [ ] Тесты ComputeBattleModifier
- [ ] `make test` зелёный
