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

### ✅ A.1 Боевые технологии — применение модификаторов

- `battle/engine.go`: `gunFactor/shieldFactor/shellFactor` применяются при инициализации `unitState`
- `fleet/attack.go`, `alien/helpers.go`: Gun/Shield/Shell заполняются через `economy.IDTech*`
- Тесты `TestGunTech_IncreasesAttack`, `TestShellTech_IncreasesArmor`, регрессия tech=0

---

### ✅ A.2 Артефакт battle_bonus — применение в бою

- `ArtefactEffect.BattleAttack/Shield/Shell` в catalog.go
- `ComputeBattleModifier`, `ActiveBattleModifiers` реализованы в artefact/
- `applyBattleMod` вызывается в `fleet/attack.go`
- ID верифицированы: battle_attack_power=318, battle_shell_power=316, battle_shield_power=317
- Тесты: `TestComputeBattleModifier_*` (4 кейса в effects_test.go)

---

### ✅ A.3 Alien AI — масштабирование флота (план 19, итерация 19)

Реализовано:
- `calcDefPower` — суммарная боевая мощь обороняющейся планеты (attack × quantity)
- `scaledAlienFleet(defPower, rng, cat)` — флот из UNIT_A_* (id 200–204), итеративно
  набирается до targetPower = defPower × random(0.9, 1.1); fallback если каталог пуст
- `ALIEN_ATTACK_INTERVAL = 6 дней` — фильтр в SQL Spawn (NOT EXISTS events за 6 дней)
- Исправлен баг: ID техов 109/110/111 → 15/16/17 (gun/shield/shell из economy.IDTech*)
- Применение профессии в `readUserTech` (alien/helpers.go)

Остаётся за рамками (сложно, нет требований):
- День четверга (×5 флотов, 150–200%)
- HOLDING event (удержание планеты)
- EVENT_ALIEN_CHANGE_MISSION_AI
