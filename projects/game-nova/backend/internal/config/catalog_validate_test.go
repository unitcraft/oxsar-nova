// Валидатор YAML-справочников из configs/ (план 22 Ф.3).
//
// Запускается как обычный go test — ловит рассинхроны между units.yml
// (реестр id/key) и балансными файлами на этапе build, а не при
// попытке игрока построить юнит.
//
// Известные исключения описаны в knownOrphans — это юниты из legacy,
// которые зарегистрированы в units.yml, но не имеют runtime-реализации
// (planetary_shield 354/355, etc.). Они не ломают runtime, но и не
// работают.

package config

import (
	"testing"
)

// knownOrphans — юниты из units.yml, для которых намеренно нет
// балансного файла или requirements. Документируют компромиссы
// с legacy-реестром. Удалять из списка только при реализации.
//
// Решение по реализации: НЕ ДЕЛАЕМ В V1 — см. ADR-0006
// (docs/adr/0006-orphan-units-deferred.md). Список является
// roadmap для v1.x: каждый юнит — отдельная задача с известным
// legacy-эффектом и прикидкой сложности.
//
// Данные из legacy-поиска 2026-04-24 (см. docs/plans/22-configs-cleanup.md
// Ф.2.2). Для каждого указано: что юнит делает в legacy и где в PHP
// реализован эффект (путь относительно d:/Sources/oxsar2/www).
var knownOrphans = map[string]string{
	// Ракеты — формально в units.yml.defense (legacy mode=4), но runtime
	// работает с ними как fleet (event kind=16, ships.yml). Оставлены
	// здесь для совместимости с legacy-реестром. См. план 22 Ф.2.3.
	"interceptor_rocket":    "ракета kind=16, работает через ships.yml",
	"interplanetary_rocket": "ракета kind=16, работает через ships.yml",

	// Планетарные щиты. В legacy: EventHandler.class.php:1716-1722 — дают
	// +10 HP (small) / +40 HP (large) к shield-пулу обороны планеты.
	// Shipyard.class.php: занимают 4/8 полей. План 22 Ф.2.2 — ADR на
	// интеграцию с battle-engine.
	"small_planet_shield": "legacy: +10 HP shield pool, 4 слота (план 22 Ф.2.2)",
	"large_planet_shield": "legacy: +40 HP shield pool, 8 слота (план 22 Ф.2.2)",

	// Ign — «интергалактическая исследовательская сеть»: добавляет лабы
	// других планет аккаунта к текущей при расчёте времени исследований
	// (Planet.class.php:810, 834). MOON_LAB множит бонус ×10.
	"ign": "legacy: virtual lab network across planets (Planet.class.php:810)",

	// gravi уже в research.yml (требуется death_star). В legacy:
	// NS.class.php::getGraviWeaponScale — формула масштаба урона от щита
	// цели. Сейчас у нас gravi = пустое исследование без effect'а в
	// engine.go. ADR: реализовать weapon scale или оставить как
	// «пустое» требование для DS.

	// Terra Former — +5 полей за уровень к max_fields планеты
	// (Planet.class.php:702-705). Простая фича, 1 поле в schema
	// + вычисление в planet.Service. Отдельный план-ADR.
	// terra_former — реализован в плане 23 (field-limit + +5 полей/level).

	// Лунные здания. moon_lab + moon_repair_factory имеют эффект в
	// legacy, moon_hydrogen_lab — пустая константа даже в legacy.
	"moon_lab":            "legacy: alt lab + 5 polей / level (Planet.class.php:693, Research.class.php:200)",
	"moon_repair_factory": "legacy: lunar repair factory (Planet.class.php:406, ext/page/ExtRepair.class.php:276)",
	"moon_hydrogen_lab":   "legacy: только константа в consts.php:288, effect не реализован даже в legacy",

	// Lancer — в legacy тоже чисто константа (consts.php:78, mode=3),
	// никакого использования. У нас применяется только через AlienAI
	// (отдельная механика, не в основной модели).
	"lancer_ship": "legacy: только константа (consts.php:78), у нас — AlienAI юнит",
}

// loadCatalogForTest читает ../../configs (относительный путь от config-пакета).
func loadCatalogForTest(t *testing.T) *Catalog {
	t.Helper()
	cat, err := LoadCatalog("../../../configs")
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}
	return cat
}

// TestValidate_AllUnitsHaveBalance — каждый id из units.yml должен
// быть в соответствующем балансном файле, кроме knownOrphans.
func TestValidate_AllUnitsHaveBalance(t *testing.T) {
	cat := loadCatalogForTest(t)

	check := func(group string, entries []UnitEntry, has func(key string) bool) {
		for _, u := range entries {
			if _, skip := knownOrphans[u.Key]; skip {
				continue
			}
			if !has(u.Key) {
				t.Errorf("%s.%s (id=%d) в units.yml, но отсутствует в балансном файле",
					group, u.Key, u.ID)
			}
		}
	}

	check("buildings", cat.Units.Buildings, func(k string) bool {
		_, ok := cat.Buildings.Buildings[k]
		return ok
	})
	check("moon_buildings", cat.Units.MoonBuildings, func(k string) bool {
		_, ok := cat.Buildings.Buildings[k]
		return ok
	})
	check("research", cat.Units.Research, func(k string) bool {
		_, ok := cat.Research.Research[k]
		return ok
	})
	check("fleet", cat.Units.Fleet, func(k string) bool {
		_, ok := cat.Ships.Ships[k]
		return ok
	})
	check("defense", cat.Units.Defense, func(k string) bool {
		_, ok := cat.Defense.Defense[k]
		return ok
	})
}

// TestValidate_RequirementsReferenceExistingUnits — в requirements.yml
// ссылки на здания/исследования должны существовать.
func TestValidate_RequirementsReferenceExistingUnits(t *testing.T) {
	cat := loadCatalogForTest(t)

	for targetKey, reqs := range cat.Requirements.Requirements {
		// Цель требований должна быть одним из юнитов.
		// (не требуется формально, но полезно — если target misspelled,
		// его Check никогда не сработает)
		if !unitExists(cat, targetKey) {
			t.Errorf("requirements.yml: target %q не существует в units.yml", targetKey)
		}

		for i, r := range reqs {
			switch r.Kind {
			case "building":
				if _, ok := cat.Buildings.Buildings[r.Key]; !ok {
					t.Errorf("requirements[%s][%d]: building %q не существует в buildings.yml",
						targetKey, i, r.Key)
				}
			case "research":
				if _, ok := cat.Research.Research[r.Key]; !ok {
					t.Errorf("requirements[%s][%d]: research %q не существует в research.yml",
						targetKey, i, r.Key)
				}
			default:
				t.Errorf("requirements[%s][%d]: неизвестный kind %q (ожидался building|research)",
					targetKey, i, r.Kind)
			}
		}
	}
}

// TestValidate_RapidfireReferenceExistingShips — from/to в rapidfire.yml
// должны быть существующими id из fleet или defense.
func TestValidate_RapidfireReferenceExistingShips(t *testing.T) {
	cat := loadCatalogForTest(t)

	// Собираем существующие id всех боевых юнитов.
	combatIDs := make(map[int]struct{})
	for _, u := range cat.Units.Fleet {
		combatIDs[u.ID] = struct{}{}
	}
	for _, u := range cat.Units.Defense {
		combatIDs[u.ID] = struct{}{}
	}

	for shooterID, targets := range cat.Rapidfire.Rapidfire {
		if _, ok := combatIDs[shooterID]; !ok {
			t.Errorf("rapidfire.yml: shooter id=%d не существует в units.yml fleet|defense",
				shooterID)
		}
		for targetID := range targets {
			if _, ok := combatIDs[targetID]; !ok {
				t.Errorf("rapidfire.yml: shooter=%d target id=%d не существует в units.yml fleet|defense",
					shooterID, targetID)
			}
		}
	}
}

// TestValidate_NoDuplicateIDs — один id не должен быть в двух местах
// реестра units.yml (кроме допустимых случаев, как planets 53 — это
// здание, а не duplicate).
func TestValidate_NoDuplicateIDs(t *testing.T) {
	cat := loadCatalogForTest(t)

	seen := make(map[int]string) // id → где впервые встретилось
	add := func(group string, entries []UnitEntry) {
		for _, u := range entries {
			if prev, ok := seen[u.ID]; ok {
				t.Errorf("units.yml: id=%d %q (в группе %s) уже есть в группе %s",
					u.ID, u.Key, group, prev)
			}
			seen[u.ID] = group
		}
	}
	add("buildings", cat.Units.Buildings)
	add("moon_buildings", cat.Units.MoonBuildings)
	add("research", cat.Units.Research)
	add("fleet", cat.Units.Fleet)
	add("defense", cat.Units.Defense)
}

// unitExists — проверка наличия key в units.yml (реестр — источник истины)
// или в балансных файлах. Используется для проверки target в requirements:
// допускаются требования к orphan-юнитам (ракеты в units.yml.defense),
// потому что requirements они всё равно должны иметь.
func unitExists(cat *Catalog, key string) bool {
	for _, group := range [][]UnitEntry{
		cat.Units.Buildings, cat.Units.MoonBuildings,
		cat.Units.Research, cat.Units.Fleet, cat.Units.Defense,
	} {
		for _, u := range group {
			if u.Key == key {
				return true
			}
		}
	}
	// Fallback на балансные файлы — для юнитов, которых нет в units.yml
	// (такое тоже ловим через AllUnitsHaveBalance).
	if _, ok := cat.Buildings.Buildings[key]; ok {
		return true
	}
	if _, ok := cat.Research.Research[key]; ok {
		return true
	}
	if _, ok := cat.Ships.Ships[key]; ok {
		return true
	}
	if _, ok := cat.Defense.Defense[key]; ok {
		return true
	}
	return false
}
