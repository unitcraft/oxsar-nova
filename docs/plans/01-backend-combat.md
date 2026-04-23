# План A: Боевой движок — технологии и артефакты

---

## История (завершено)

### ✅ M4.1 Боевой движок — базовый порт (итерация 20–21)
- Портирован `Units.processAttack` shield-блок из oxsar2-java
- Реализован rapidfire, ablation (damaged units), ballistics/masking
- Battle report, loot, debris fields, moon-chance
- Тесты: property-based + golden файлы

### ✅ M4.4 Rapidfire — полная таблица (план 17, итерация ~50)
Синхронизировано с `na_rapidfire` (legacy MySQL):
- Cruiser (33) → Light Fighter (×6), Rocket Launcher (×10)
- Battleship (34) → Espionage Probe (×5), Solar Satellite (×5)
- Frigate (35) → Small/Large Transporter (×3), Heavy Fighter (×7), Cruiser (×4), Battleship (×7)
- Bomber (40) → Rocket Launcher (×20), Light Laser (×20), Heavy Laser (×10), Ion Cannon (×10), Lancer (×5)
- Star Destroyer (41) → Frigate (×2), Light/Heavy Laser (×2), Lancer (×2)
- Lancer (102) → Deathstar (×3)
- Plasma Turret (48) → Deathstar (×2)
- Deathstar (42) → большинство юнитов (×25–1250)

**Не добавлять без подтверждения из na_rapidfire:** Bomber→ShieldDome, HeavyFighter→Rocket/Laser,
Cruiser→LightLaser — предположения, в legacy DB не найдены.

### ✅ Боевые артефакты — battle_bonus (nil, nil) (план 01, итерация ~48)
- `computeChanges` возвращает `(nil, nil)` для battle_bonus — эффект применяется
  отдельно через `ComputeBattleModifier`, не через изменение полей планеты
- Тест `TestComputeChanges_BattleBonus` исправлен

### ✅ Корабли пришельцев в конфиге (план 17, итерация ~50)
- Alien Corvette (200), Screen (201), Paladin (202), Frigate (203), Torpedocarrier (204)
  добавлены в `configs/ships.yml` с параметрами из `na_ship_datasheet`
- Alien ships имеют свою таблицу RF между собой и против юнитов игрока (из na_rapidfire)

---

## Открытые задачи

### A.1 Боевые технологии — применение модификаторов (приоритет: HIGH)

**Проблема:** Gun/Shield/Shell tech прокачиваются за ресурсы, но в бою **ничего не делают**.
`Side.Tech.Gun/Shield/Shell` нигде не применяются в `battle/engine.go`.
Игрок качает Weapons Tech 5 за ~12 800 металла + 2 часа — получает 0 эффекта.
(§7.2, §13.3 balance-analysis.md)

**Формула (legacy oxsar2-java и OGame):**
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
st.shield = u.Shield * shieldFactor
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
ID и множители верифицировать по `d:\Sources\oxsar2\consts.php` перед мержем.
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

---

### A.3 Alien AI — полная state machine (приоритет: LOW)

**Проблема:** Текущий alien AI использует фиксированный флот (5 LF) и нет масштабирования.
В legacy (AlienAI.class.php) — полная state machine с holding, грабёж кредитов, подарки.

**Что не реализовано (из legacy-game-reference.md):**

| Функция | Статус в nova |
|---------|---------------|
| Флот масштабируется по силе игрока (90–110%) | ❌ фиксированные 5 LF |
| Корабли пришельцев (UNIT_A_*, id 200–204) в боях | ❌ |
| День усиленных атак (четверг, ×5 флотов, 150–200% силы) | ❌ |
| Смена миссии в полёте (EVENT_ALIEN_CHANGE_MISSION_AI) | ❌ |
| Режим удержания планеты (HOLDING, 12–24ч) | ❌ |
| Грабёж кредитов (0.0008–0.001% от кредитов игрока) | частично |
| Подарки ресурсов (5%) и кредитов (5%) при прилёте | ❌ |

**Формулы масштабирования (legacy):**
```
target_power = суммарная боевая сила планеты (sum attack × quantity)
fleet_power  = target_power × random(0.9, 1.1)  [обычный день]
fleet_power  = target_power × random(1.5, 2.0)  [четверг]
```

Корабли пришельцев итеративно добавляются от UNIT_A_CORVETTE до UNIT_A_TORPEDOCARIER
пока fleet_power не достигнут. Максимум обломков: `ALIEN_FLEET_MAX_DERBIS = 1 млрд`.

**Разумное упрощение для первой итерации:**
Реализовать только масштабирование мощи (шаг 1–2), остальное оставить на потом.

**Шаг 1** — `backend/internal/alien/helpers.go`: `calcAlienFleetPower(targetPlanetShips map[string]int64)`.
**Шаг 2** — `generateMission` в `alien/`: использовать unit_a_corvette..unit_a_torpedocarier
вместо light_fighter. Добавить проверку `ALIEN_ATTACK_INTERVAL = 6 days`.
**Шаг 3** — `backend/internal/alien/holding.go`: реализовать HOLDING event (min viable).

**Параметры из legacy-game-reference.md:**
- `ALIEN_ATTACK_INTERVAL`: 6 дней (мин. интервал между атаками одного игрока)
- `ALIEN_NORMAL_FLEETS_NUMBER`: 50 флотов одновременно в мире
- `BASHING_MAX_ATTACKS = 4` атаки за 5ч (защита от башинга)
- `PROTECTION_PERIOD = 24ч` — новый игрок защищён от атак

**Проверка готовности:**
- [ ] Alien fleet масштабируется по силе игрока (90–110%)
- [ ] Используются корабли UNIT_A_* вместо light_fighter
- [ ] `ALIEN_ATTACK_INTERVAL` = 6 дней соблюдается
- [ ] `make test` зелёный
