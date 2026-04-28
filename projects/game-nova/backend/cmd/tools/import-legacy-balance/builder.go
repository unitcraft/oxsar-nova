package main

import (
	"fmt"
	"log/slog"
	"math"
	"sort"
	"strings"
	"time"
)

// OverrideDoc — финальная YAML-структура configs/balance/origin.yaml.
// Сериализуется в-вручную через formatYAML (не yaml.Marshal),
// чтобы:
//
//  1. контролировать порядок ключей (сначала legacy meta, потом
//     globals, потом buildings и т.д.)
//  2. чётко форматировать int'ы без научной нотации (yaml.v3 склонен
//     писать 1e+06)
//  3. генерировать осмысленные комментарии в YAML («prod_metal был
//     динамической формулой → see internal/origin/economy/...»).
type OverrideDoc struct {
	GeneratedAt time.Time
	Universe    string
	Version     int
	Globals     map[string]float64
	Buildings   map[string]BuildingOverrideOut
	Research    map[string]ResearchOverrideOut
	Ships       map[string]ShipOverrideOut
	// Buildings/research/ships keys in their canonical order:
	BuildingsOrder []string
	ResearchOrder  []string
	ShipsOrder     []string
}

// ShipOverrideOut — origin-числа для одного ship'а, отличающиеся от nova.
// Применяются как override в configs/balance/origin.yaml.
type ShipOverrideOut struct {
	OriginID      int
	OriginName    string
	Cost          ResCost
	Attack        int
	Shield        int
	Shell         int
	Cargo         int64
	Speed         int
	Fuel          int
	Front         int
}

// BuildingOverrideOut — origin-числа для одного здания. Используется
// для сериализации только тех полей, которые ОТЛИЧАЮТСЯ от nova-дефолта
// (override semantics — отсутствующие ключи остаются дефолтными).
//
// HasDynamicProd / HasDynamicCons — origin DSL содержит {temp}/{tech=N} —
// предвычислить таблицы нельзя; помечаем для будущих Go-функций
// (план 64 Ф.4, internal/origin/economy/).
type BuildingOverrideOut struct {
	OriginID                int
	OriginName              string
	BasicMetal              int64
	BasicSilicon            int64
	BasicHydrogen           int64
	BasicEnergy             int64
	HasBasic                bool
	CostFactor              float64 // средний показатель степени из charge_*; 0 если не выявлен
	HasCostFactor           bool
	HasDynamicProd          bool
	HasDynamicCons          bool
	OriginChargeMetal       string
	OriginChargeSilicon     string
	OriginChargeHydrogen    string
	OriginChargeEnergy      string
	OriginProdFormula       string
	OriginConsFormula       string
	ChargeMetalTable        []int64 // 1..maxLevel (если статика)
	ChargeSiliconTable      []int64
	ChargeHydrogenTable     []int64
}

type ResearchOverrideOut struct {
	OriginID      int
	OriginName    string
	BasicMetal    int64
	BasicSilicon  int64
	BasicHydrogen int64
	HasBasic      bool
	CostFactor    float64
	HasCostFactor bool
}

// DefaultExtensions — добавки в дефолтные configs/units.yml + ships.yml
// + rapidfire.yml для R0-исключения (алиен/спец-юниты во всех
// вселенных).
type DefaultExtensions struct {
	// UnitsAppend — записи для units.yml.fleet (алиен и спец-юниты,
	// которые игроки могут строить — Lancer, Shadow, Transplantator,
	// Collector, Armored Terran). И для units.yml.defense (small/large
	// planet shield).
	UnitsAppend     []UnitEntry
	UnitsDefenseAppend []UnitEntry

	// ShipsAppend — балансовые записи для configs/ships.yml.
	ShipsAppend     map[string]ShipBalance

	// RapidfireAppend — записи в configs/rapidfire.yml для алиен-юнитов.
	// Map shooterID → map[targetID]value.
	RapidfireAppend map[int]map[int]int
}

type UnitEntry struct {
	ID   int
	Key  string
	Name string
}

type ShipBalance struct {
	ID     int
	Cost   ResCost
	Attack int
	Shield int
	Shell  int
	Cargo  int64
	Speed  int
	Fuel   int
	Front  int
}

type ResCost struct {
	Metal    int64
	Silicon  int64
	Hydrogen int64
}

// buildOverride принимает все таблицы origin БД и собирает override-
// документ для configs/balance/origin.yaml.
//
// Алгоритм для buildings/research:
//   1. Для каждой записи na_construction (mode=1 building, mode=2 research):
//      - находим nova-key через novaKeyForName(name)
//      - если name не мэппится (unmapped) → лог warn, skip (не падаем)
//      - если ID — алиен/спец → skip (он уйдёт в DefaultExtensions)
//      - basic_* поля копируем как есть
//      - charge_* парсим DSL; если статичны — извлекаем cost_factor
//        (показатель степени pow) + предвычисляем таблицу 1..maxLevel
//      - prod_*/cons_* — если динамичны (содержат {temp}/{tech=N}) →
//        HasDynamicProd/HasDynamicCons=true, формула сохраняется в
//        OriginProdFormula/OriginConsFormula для документации.
// origin-IDs спец-юнитов, которые УЖЕ ЕСТЬ в дефолтном configs/ships.yml
// с nova-балансом (план 22 + ADR-0007/0008). Для origin их параметры
// переопределяются через configs/balance/origin.yaml override (cost,
// attack — из na_construction/na_ship_datasheet).
//
// armored_terran/ship_transplantator/ship_collector — НЕ в этом списке,
// потому что они отсутствуют в nova-default и идут как добавки.
//
// small_planet_shield/large_planet_shield (354/355) — упомянуты в
// nova units.yml.defense (knownOrphan), но БАЛАНСОВЫХ записей в
// ships.yml/defense.yml нет (план 22 Ф.2.2 — отложен до ADR). Поэтому
// override бессмыслен; импортёр их пропускает.
var existingNovaSpecialUnits = map[int]string{
	102: "lancer_ship",
	325: "shadow_ship",
	200: "unit_a_corvette",
	201: "unit_a_screen",
	202: "unit_a_paladin",
	203: "unit_a_frigate",
	204: "unit_a_torpedocarier",
}

// origin-IDs спец-юнитов, которые ОТСУТСТВУЮТ в дефолтном configs.
// Импортёр их добавляет в дефолтные units.yml/ships.yml (R0-exception).
// 358 (SHIP_ARMORED_TERRAN) — debug-юнит (RF×900 ко всему), включается
// для admin-tooling, но из RF-таблицы исключается (см. план 18 — RF
// 358 пропускается).
var newDefaultSpecialUnits = map[int]string{
	352: "ship_transplantator",
	353: "ship_collector",
	358: "armored_terran",
}

func buildOverride(cons []Construction, ships []ShipDatasheet, rf []RapidfireEntry, maxLevel int, log *slog.Logger) (*OverrideDoc, error) {
	doc := &OverrideDoc{
		GeneratedAt: time.Now().UTC(),
		Universe:    "origin",
		Version:     1,
		Globals: map[string]float64{
			// Origin совпадает с nova ModernGlobals — формулы prod_*
			// origin = nova economy/formulas.go (verify 2026-04-28
			// против live origin docker-mysql-1). Поэтому Globals
			// override-файла оставляем пустым (== дефолт). Если в
			// будущем числа разойдутся, перечислять здесь поле + значение.
		},
		Buildings: map[string]BuildingOverrideOut{},
		Research:  map[string]ResearchOverrideOut{},
		Ships:     map[string]ShipOverrideOut{},
	}

	// Индексируем ship_datasheet для bullet-числами override-ships.
	dsIdx := make(map[int]ShipDatasheet, len(ships))
	for _, s := range ships {
		dsIdx[s.UnitID] = s
	}

	for _, c := range cons {
		switch c.Mode {
		case modeBuilding:
			if isAlienOrSpecial(c.BuildingID) {
				continue
			}
			key := novaKeyForName(c.Name)
			if key == "" {
				continue
			}
			if !isKnownNovaBuilding(key) {
				log.Warn("origin building not in nova default — skipping override",
					slog.String("origin_name", c.Name),
					slog.String("nova_key", key))
				continue
			}
			out, err := buildBuildingOverride(c, maxLevel, log)
			if err != nil {
				return nil, fmt.Errorf("build %s (%s): %w", c.Name, key, err)
			}
			doc.Buildings[key] = out
			doc.BuildingsOrder = append(doc.BuildingsOrder, key)

		case modeResearch:
			key := novaKeyForName(c.Name)
			if key == "" {
				continue
			}
			// Origin может содержать research-ключи, которых нет в
			// nova (например, ARTEFACTS_TECH из oxsar2 в nova не
			// портирован). Override таких ключей бессмыслен — нечего
			// перекрывать. Skip их с warn'ом.
			if !isKnownNovaResearch(key) {
				log.Warn("origin research not in nova default — skipping override",
					slog.String("origin_name", c.Name),
					slog.String("nova_key", key))
				continue
			}
			out, err := buildResearchOverride(c, log)
			if err != nil {
				return nil, fmt.Errorf("research %s (%s): %w", c.Name, key, err)
			}
			doc.Research[key] = out
			doc.ResearchOrder = append(doc.ResearchOrder, key)

		case modeShip, modeDefense:
			// Override параметров для спец-юнитов, которые УЖЕ ЕСТЬ
			// в nova-default с другим балансом (R0): origin cost +
			// attack — другие, переопределяем.
			novaKey, isExisting := existingNovaSpecialUnits[c.BuildingID]
			if !isExisting {
				continue
			}
			ds, ok := dsIdx[c.BuildingID]
			if !ok {
				log.Warn("no ship_datasheet for special unit, skipping override",
					slog.Int("origin_id", c.BuildingID), slog.String("nova_key", novaKey))
				continue
			}
			doc.Ships[novaKey] = ShipOverrideOut{
				OriginID:   c.BuildingID,
				OriginName: c.Name,
				Cost: ResCost{
					Metal:    int64(c.BasicMetal),
					Silicon:  int64(c.BasicSilicon),
					Hydrogen: int64(c.BasicHydrogen),
				},
				Attack: ds.Attack,
				Shield: ds.Shield,
				Shell:  shellFromBasic(c.BasicMetal, c.BasicSilicon, c.BasicHydrogen),
				Cargo:  ds.Capicity,
				Speed:  ds.Speed,
				Fuel:   ds.Consume,
				Front:  ds.Front,
			}
			doc.ShipsOrder = append(doc.ShipsOrder, novaKey)

		default:
			// Mode 5+: служебные / moon. Skip с warn (план 64
			// сосредоточен на core balance).
			continue
		}
	}

	sort.Strings(doc.BuildingsOrder)
	sort.Strings(doc.ResearchOrder)
	sort.Strings(doc.ShipsOrder)

	return doc, nil
}

// shellFromBasic — origin shell вычислялся как 10*(metal+silicon+hydrogen)/1000
// или похожей формулой в legacy. Точная: shell = (basic_metal +
// basic_silicon + basic_hydrogen) / 10. Это эмпирика из легаси-кода
// Functions.inc.php; если нужно точнее — посмотреть consts.php.
//
// Для override используем как best-guess; если nova-shell отличается —
// override его не трогает (origin shell ≈ nova shell в большинстве
// случаев — verify в spot-check).
func shellFromBasic(m, s, h float64) int {
	return int((m + s + h) / 10)
}

func buildBuildingOverride(c Construction, maxLevel int, log *slog.Logger) (BuildingOverrideOut, error) {
	out := BuildingOverrideOut{
		OriginID:      c.BuildingID,
		OriginName:    c.Name,
		BasicMetal:    int64(c.BasicMetal),
		BasicSilicon:  int64(c.BasicSilicon),
		BasicHydrogen: int64(c.BasicHydrogen),
		BasicEnergy:   int64(c.BasicEnergy),
		HasBasic:      c.BasicMetal != 0 || c.BasicSilicon != 0 || c.BasicHydrogen != 0 || c.BasicEnergy != 0,
		OriginChargeMetal:    c.ChargeMetal,
		OriginChargeSilicon:  c.ChargeSilicon,
		OriginChargeHydrogen: c.ChargeHydrogen,
		OriginChargeEnergy:   c.ChargeEnergy,
	}

	// charge_*: ищем cost_factor через метод «inferCostFactor», и
	// предвычисляем таблицу 1..maxLevel.
	if c.ChargeMetal != "" {
		factor, isStatic, err := inferCostFactor(c.ChargeMetal)
		if err != nil {
			return out, fmt.Errorf("parse charge_metal %q: %w", c.ChargeMetal, err)
		}
		if isStatic && factor > 0 {
			out.CostFactor = factor
			out.HasCostFactor = true
			out.ChargeMetalTable = precomputeCharge(c.ChargeMetal, c.BasicMetal, maxLevel)
		}
	}
	if c.ChargeSilicon != "" {
		out.ChargeSiliconTable = precomputeCharge(c.ChargeSilicon, c.BasicSilicon, maxLevel)
	}
	if c.ChargeHydrogen != "" {
		out.ChargeHydrogenTable = precomputeCharge(c.ChargeHydrogen, c.BasicHydrogen, maxLevel)
	}

	// prod_*/cons_*: проверяем динамичность.
	for _, src := range []string{c.ProdMetal, c.ProdSilicon, c.ProdHydrogen, c.ProdEnergy} {
		if strings.TrimSpace(src) == "" {
			continue
		}
		dyn, _ := IsDynamic(src)
		if dyn {
			out.HasDynamicProd = true
			if out.OriginProdFormula == "" {
				out.OriginProdFormula = src
			}
		}
	}
	for _, src := range []string{c.ConsMetal, c.ConsSilicon, c.ConsHydrogen, c.ConsEnergy} {
		if strings.TrimSpace(src) == "" {
			continue
		}
		dyn, _ := IsDynamic(src)
		if dyn {
			out.HasDynamicCons = true
			if out.OriginConsFormula == "" {
				out.OriginConsFormula = src
			}
		}
	}

	return out, nil
}

func buildResearchOverride(c Construction, log *slog.Logger) (ResearchOverrideOut, error) {
	out := ResearchOverrideOut{
		OriginID:      c.BuildingID,
		OriginName:    c.Name,
		BasicMetal:    int64(c.BasicMetal),
		BasicSilicon:  int64(c.BasicSilicon),
		BasicHydrogen: int64(c.BasicHydrogen),
		HasBasic:      c.BasicMetal != 0 || c.BasicSilicon != 0 || c.BasicHydrogen != 0,
	}
	if c.ChargeMetal != "" {
		factor, isStatic, err := inferCostFactor(c.ChargeMetal)
		if err == nil && isStatic && factor > 0 {
			out.CostFactor = factor
			out.HasCostFactor = true
		}
	}
	return out, nil
}

// inferCostFactor извлекает «cost_factor» из charge-формулы вида
// floor({basic} * pow(F, ({level} - 1))) или {basic} * pow(F, ({level} - 1))
// или 50 * pow(F, {level}) (без {basic} — в этом случае возвращаем F
// и просим caller интерпретировать precomputed-table напрямую, а
// CostFactor == F).
//
// Возвращает (factor, isStatic, err):
//   - isStatic = false если формула динамическая ({temp}/{tech})
//   - factor = 0 если формула статика, но не классической формы — caller
//     полагается на precomputed-таблицу
func inferCostFactor(src string) (float64, bool, error) {
	dyn, err := IsDynamic(src)
	if err != nil {
		return 0, false, err
	}
	if dyn {
		return 0, false, nil
	}
	// Берём два уровня (level=10 и level=11) при basic=10000, делим —
	// это и будет F. Высокий basic + высокие уровни — чтобы избежать
	// floor-обрезки на маленьких числах (для level=1,2 при basic=1
	// floor(1.5)=1, floor(1)=1 → ratio=1, ложный «не factor»).
	//
	// Эмпирическая инференция — проще чем regex-матчить структуру
	// формулы (origin использует разные формы: floor({basic}*pow(...)),
	// {basic}*pow(...) без floor, прямые числа N*pow(F,{level}) и т.п.).
	level1, basic1 := 10, 10000.0
	v1, err := EvalNumber(src, VarBinding{Level: &level1, Basic: &basic1})
	if err != nil {
		return 0, true, err
	}
	level2 := 11
	basic2 := 10000.0
	v2, err := EvalNumber(src, VarBinding{Level: &level2, Basic: &basic2})
	if err != nil {
		return 0, true, err
	}
	if v1 <= 0 {
		return 0, true, nil
	}
	factor := v2 / v1
	// Округляем до 0.01 для красивого числа (1.5, 1.6, 2.0, ...).
	factor = math.Round(factor*100) / 100
	return factor, true, nil
}

// precomputeCharge вычисляет таблицу значений charge-формулы для уровней
// 1..maxLevel. Если формула не парсится — возвращает nil (caller
// логирует / решает).
func precomputeCharge(src string, basic float64, maxLevel int) []int64 {
	out := make([]int64, 0, maxLevel)
	for level := 1; level <= maxLevel; level++ {
		l := level
		b := basic
		v, err := EvalNumber(src, VarBinding{Level: &l, Basic: &b})
		if err != nil {
			return nil
		}
		out = append(out, int64(math.Round(v)))
	}
	return out
}

// buildDefaultExtensions создаёт **только новые** алиен/спец-юниты
// для R0-исключения, которых нет в текущих nova-конфигах. План 64 R0-
// исключение: эти юниты доступны во всех вселенных (uni01/uni02/origin).
//
// УЖЕ ЕСТЬ в nova default (через план 22/26 + ADR-0007/0008):
//   - alien fleet 200..204 как unit_a_corvette..unit_a_torpedocarier
//   - lancer_ship (102), shadow_ship (325)
//   - small_planet_shield (354), large_planet_shield (355)
// Их числа nova не совпадают с origin (R0: nova-баланс заморожен) —
// для origin-вселенной они переопределяются в configs/balance/origin.yaml.
//
// ОТСУТСТВУЮТ в nova default — добавляются этим импортёром:
//   - ship_transplantator (352)
//   - ship_collector (353)
//   - armored_terran (358) — debug-юнит, добавляется только в реестр,
//     RF×900 НЕ переносится в default (см. plan-18 comment).
func buildDefaultExtensions(cons []Construction, ships []ShipDatasheet, rf []RapidfireEntry, log *slog.Logger) (*DefaultExtensions, error) {
	ext := &DefaultExtensions{
		ShipsAppend:     make(map[string]ShipBalance),
		RapidfireAppend: make(map[int]map[int]int),
	}

	dsIdx := make(map[int]ShipDatasheet, len(ships))
	for _, s := range ships {
		dsIdx[s.UnitID] = s
	}
	consIdx := make(map[int]Construction, len(cons))
	for _, c := range cons {
		consIdx[c.BuildingID] = c
	}

	addNewUnit := func(originID int, novaKey, displayName string, defenseUnit bool) {
		ds, ok := dsIdx[originID]
		if !ok {
			log.Warn("no ship_datasheet for unit, skipping", slog.Int("origin_id", originID), slog.String("nova_key", novaKey))
			return
		}
		c, hasCost := consIdx[originID]
		entry := UnitEntry{ID: originID, Key: novaKey, Name: displayName}
		if defenseUnit {
			ext.UnitsDefenseAppend = append(ext.UnitsDefenseAppend, entry)
		} else {
			ext.UnitsAppend = append(ext.UnitsAppend, entry)
		}
		bal := ShipBalance{
			ID:     originID,
			Attack: ds.Attack,
			Shield: ds.Shield,
			Cargo:  ds.Capicity,
			Speed:  ds.Speed,
			Fuel:   ds.Consume,
			Front:  ds.Front,
		}
		if hasCost {
			bal.Cost = ResCost{
				Metal:    int64(c.BasicMetal),
				Silicon:  int64(c.BasicSilicon),
				Hydrogen: int64(c.BasicHydrogen),
			}
			bal.Shell = shellFromBasic(c.BasicMetal, c.BasicSilicon, c.BasicHydrogen)
		}
		ext.ShipsAppend[novaKey] = bal
	}

	addNewUnit(352, "ship_transplantator", "Transplantator", false)
	addNewUnit(353, "ship_collector", "Collector", false)
	addNewUnit(358, "armored_terran", "Armored Terran", false)

	// Rapidfire: добавляем только записи для unit-ID, которые алиен/
	// спец, ИЛИ цели которых алиен/спец. Фильтруем:
	//   - debug-юнит 358 (RF×900 ко всему — см. план-18 comment)
	//   - 348 (TRANSMITTER) — есть в legacy `na_construction` но
	//     отсутствует в nova units.yml; импортировать его как
	//     самостоятельный юнит выходит за scope плана 64
	//     (он связан с механикой передачи ресурсов, требует отдельного
	//     плана). RF-запись с участием 348 пропускается.
	const debugTerran = 358
	const transmitter = 348
	skipIDs := map[int]bool{
		debugTerran: true,
		transmitter: true,
	}
	includeIDs := map[int]bool{}
	for _, id := range alienUnitIDs {
		includeIDs[id] = true
	}
	for _, id := range specialUnitIDs {
		if skipIDs[id] {
			continue
		}
		includeIDs[id] = true
	}
	for _, e := range rf {
		if skipIDs[e.UnitID] || skipIDs[e.Target] {
			continue
		}
		if !(includeIDs[e.UnitID] || includeIDs[e.Target]) {
			continue
		}
		if ext.RapidfireAppend[e.UnitID] == nil {
			ext.RapidfireAppend[e.UnitID] = map[int]int{}
		}
		ext.RapidfireAppend[e.UnitID][e.Target] = e.Value
	}

	return ext, nil
}
