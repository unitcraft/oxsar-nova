// Command battle-sim — прогон сценариев боя на ходу с текущим
// состоянием configs/ для балансовой аналитики.
//
// Использование:
//
//	battle-sim --scenario=lancer-vs-cruiser --runs=100
//	battle-sim --all --runs=100
//
// Что делает:
//   - читает configs/ (ships.yml, defense.yml, rapidfire.yml, construction.yml)
//   - собирает сценарии боёв (Lancer-spam vs Cruiser, DS vs fleet, Bomber vs defense, ...)
//   - прогоняет каждый N раз с разным seed
//   - выдаёт сводку: winrate атакующего, медианные потери по ресурсам,
//     EV/cost (насколько атакующий эффективно тратит ресурсы)
//
// БД не трогает — чистая функция от configs/.
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/oxsar/nova/backend/internal/battle"
	"github.com/oxsar/nova/backend/internal/config"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "battle-sim:", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		scenario   = flag.String("scenario", "", "имя сценария")
		all        = flag.Bool("all", false, "прогнать все сценарии")
		matrix     = flag.Bool("matrix", false, "матрица 1v1: каждый combat-юнит vs каждый при равной metal-eq")
		matrixBudget = flag.Int64("matrix-budget", 10000000, "metal-eq на сторону для --matrix (default 10M)")
		groups     = flag.Bool("groups", false, "группы vs группы (lite/mid/capital/endgame)")
		runs       = flag.Int("runs", 50, "число прогонов на сценарий")
		rounds     = flag.Int("rounds", 6, "макс раундов в одном бою")
		catalogDir = flag.String("configs", "../../../configs", "путь к configs/")
		costOverride  multiFlag
		frontOverride multiFlag
	)
	flag.Var(&costOverride, "cost", "переопределить стоимость юнита: ship_key=M/Si/H")
	flag.Var(&frontOverride, "front", "переопределить front юнита: ship_key=N")
	flag.Parse()

	if *scenario == "" && !*all && !*matrix && !*groups {
		return fmt.Errorf("нужен --scenario=<name>, --all, --matrix или --groups")
	}

	cat, err := config.LoadCatalog(*catalogDir)
	if err != nil {
		return fmt.Errorf("load catalog: %w", err)
	}

	for _, s := range costOverride {
		if err := applyCostOverride(cat, s); err != nil {
			return fmt.Errorf("cost override %q: %w", s, err)
		}
	}
	for _, s := range frontOverride {
		if err := applyFrontOverride(cat, s); err != nil {
			return fmt.Errorf("front override %q: %w", s, err)
		}
	}

	rf := cat.Rapidfire.Rapidfire

	if *matrix {
		runMatrix(cat, rf, *matrixBudget, *runs, *rounds)
		return nil
	}
	if *groups {
		runGroups(cat, rf, *runs, *rounds)
		return nil
	}

	scenarios := buildScenarios(cat)

	var toRun []scn
	if *all {
		toRun = scenarios
	} else {
		for _, s := range scenarios {
			if s.Name == *scenario {
				toRun = []scn{s}
				break
			}
		}
		if len(toRun) == 0 {
			fmt.Fprintln(os.Stderr, "доступные сценарии:")
			for _, s := range scenarios {
				fmt.Fprintf(os.Stderr, "  %s — %s\n", s.Name, s.Descr)
			}
			return fmt.Errorf("сценарий %q не найден", *scenario)
		}
	}

	for _, s := range toRun {
		runScenario(s, rf, *runs, *rounds)
	}
	return nil
}

// matrixUnits — combat-юниты для матрицы (в порядке tier'ов).
var matrixShipUnits = []string{
	"light_fighter",
	"strong_fighter",
	"cruiser",
	"battle_ship",
	"frigate",
	"bomber",
	"star_destroyer",
	"lancer_ship",
	"shadow_ship",
	"death_star",
}

var matrixDefUnits = []string{
	"rocket_launcher",
	"light_laser",
	"strong_laser",
	"ion_gun",
	"gauss_gun",
	"plasma_gun",
}

// runMatrix прогоняет каждую пару юнитов при равной metal-eq budget.
// Печатает компактную таблицу exchange-ratio (def_loss / atk_loss).
func runMatrix(cat *config.Catalog, rf map[int]map[int]int, budget int64, runs, rounds int) {
	allUnits := append([]string{}, matrixShipUnits...)
	allUnits = append(allUnits, matrixDefUnits...)

	header := fmt.Sprintf("=== MATRIX 1v1 @ %d metal-eq, runs=%d ===", budget, runs)
	fmt.Println(header)
	fmt.Println("Cell = exchange ratio (def_loss/atk_loss). >1 = атакующий выгоден; <1 = защитник; — = atk не теряет.")

	// Header row
	fmt.Printf("%-18s |", "atk \\ def")
	for _, d := range allUnits {
		fmt.Printf(" %8s", short(d))
	}
	fmt.Println()
	fmt.Println(strings.Repeat("-", 20+len(allUnits)*9))

	for _, atkKey := range matrixShipUnits {
		fmt.Printf("%-18s |", atkKey)
		for _, defKey := range allUnits {
			res := matrixCell(cat, rf, atkKey, defKey, budget, runs, rounds)
			fmt.Printf(" %8s", res)
		}
		fmt.Println()
	}
}

func short(key string) string {
	if len(key) > 8 {
		// просто первые 8 символов
		return key[:8]
	}
	return key
}

// matrixCell — один прогон atkKey vs defKey при равной metal-eq.
// Возвращает строку: "X.XX" (exchange) или "ATK" (атакующий не теряет
// и побеждает) или "DEF" (защитник не теряет).
func matrixCell(cat *config.Catalog, rf map[int]map[int]int, atkKey, defKey string, budget int64, runs, rounds int) string {
	atkSide := buildSideFromKey(cat, atkKey, budget)
	defSide := buildSideFromKey(cat, defKey, budget)
	if atkSide == nil || defSide == nil {
		return "n/a"
	}

	var atkLosses, defLosses []int64
	atkWins, defWins := 0, 0
	for i := 0; i < runs; i++ {
		in := battle.Input{
			Seed:      uint64(7000 + i),
			Rounds:    rounds,
			Attackers: cloneSides([]battle.Side{*atkSide}),
			Defenders: cloneSides([]battle.Side{*defSide}),
			Rapidfire: rf,
		}
		rep, err := battle.Calculate(in)
		if err != nil {
			return "ERR"
		}
		atkLosses = append(atkLosses, sumLoss(rep.Attackers))
		defLosses = append(defLosses, sumLoss(rep.Defenders))
		switch rep.Winner {
		case "attackers":
			atkWins++
		case "defenders":
			defWins++
		}
	}
	al := median(atkLosses)
	dl := median(defLosses)
	if al == 0 && dl == 0 {
		return "—"
	}
	if al == 0 {
		return "ATK"
	}
	if dl == 0 {
		return "def"
	}
	ratio := float64(dl) / float64(al)
	return fmt.Sprintf("%.2f", ratio)
}

// buildSideFromKey — собрать сторону из единственного юнита, заполнив
// до budget metal-eq.
func buildSideFromKey(cat *config.Catalog, key string, budget int64) *battle.Side {
	if spec, ok := cat.Ships.Ships[key]; ok {
		costPer := spec.Cost.Metal + spec.Cost.Silicon + spec.Cost.Hydrogen
		if costPer <= 0 {
			return nil
		}
		qty := budget / costPer
		if qty < 1 {
			qty = 1
		}
		front := spec.Front
		if front <= 0 {
			front = 10
		}
		u := battle.Unit{
			UnitID:   spec.ID,
			Quantity: qty,
			Front:    front,
			Attack:   float64(spec.Attack),
			Shield:   float64(spec.Shield),
			Shell:    float64(spec.Shell),
			Name:     key,
			Cost: battle.UnitCost{
				Metal: spec.Cost.Metal, Silicon: spec.Cost.Silicon, Hydrogen: spec.Cost.Hydrogen,
			},
		}
		return &battle.Side{UserID: key, Username: key, Units: []battle.Unit{u}}
	}
	if spec, ok := cat.Defense.Defense[key]; ok {
		costPer := spec.Cost.Metal + spec.Cost.Silicon + spec.Cost.Hydrogen
		if costPer <= 0 {
			return nil
		}
		qty := budget / costPer
		if qty < 1 {
			qty = 1
		}
		front := spec.Front
		if front <= 0 {
			front = 10
		}
		u := battle.Unit{
			UnitID:   spec.ID,
			Quantity: qty,
			Front:    front,
			Attack:   float64(spec.Attack),
			Shield:   float64(spec.Shield),
			Shell:    float64(spec.Shell),
			Name:     key,
			Cost: battle.UnitCost{
				Metal: spec.Cost.Metal, Silicon: spec.Cost.Silicon, Hydrogen: spec.Cost.Hydrogen,
			},
		}
		return &battle.Side{UserID: key, Username: key, Units: []battle.Unit{u}}
	}
	return nil
}

// runGroups — групповые тесты при равных бюджетах.
func runGroups(cat *config.Catalog, rf map[int]map[int]int, runs, rounds int) {
	type group struct {
		name string
		// distribution: ключ → доля cost.
		dist map[string]float64
	}
	groups := []group{
		{"lite-fleet", map[string]float64{"light_fighter": 0.6, "strong_fighter": 0.3, "cruiser": 0.1}},
		{"mid-fleet", map[string]float64{"cruiser": 0.4, "battle_ship": 0.4, "frigate": 0.2}},
		{"capital-fleet", map[string]float64{"battle_ship": 0.3, "frigate": 0.2, "star_destroyer": 0.3, "bomber": 0.2}},
		{"endgame-fleet", map[string]float64{"battle_ship": 0.2, "star_destroyer": 0.2, "death_star": 0.5, "frigate": 0.1}},
		{"shadow-fleet", map[string]float64{"shadow_ship": 0.7, "battle_ship": 0.3}},
		{"lancer-fleet", map[string]float64{"lancer_ship": 0.6, "battle_ship": 0.4}},
		{"defense-light", map[string]float64{"rocket_launcher": 0.5, "light_laser": 0.3, "ion_gun": 0.2}},
		{"defense-heavy", map[string]float64{"gauss_gun": 0.4, "plasma_gun": 0.4, "ion_gun": 0.2}},
		{"defense-mixed", map[string]float64{"rocket_launcher": 0.2, "light_laser": 0.1, "gauss_gun": 0.3, "plasma_gun": 0.4}},
	}

	budgets := []int64{10000000, 50000000, 100000000}

	fmt.Println("=== GROUPS vs GROUPS (равные бюджеты) ===")
	fmt.Println("Cell = exchange ratio (def_loss/atk_loss).")

	for _, b := range budgets {
		fmt.Printf("\n--- Budget %d metal-eq ---\n", b)
		fmt.Printf("%-18s |", "atk \\ def")
		for _, g := range groups {
			fmt.Printf(" %14s", g.name)
		}
		fmt.Println()
		fmt.Println(strings.Repeat("-", 20+len(groups)*15))

		for _, atkG := range groups {
			fmt.Printf("%-18s |", atkG.name)
			for _, defG := range groups {
				atkSide := buildSideFromGroup(cat, atkG.dist, b)
				defSide := buildSideFromGroup(cat, defG.dist, b)
				if atkSide == nil || defSide == nil {
					fmt.Printf(" %14s", "n/a")
					continue
				}
				var atkLosses, defLosses []int64
				for i := 0; i < runs; i++ {
					in := battle.Input{
						Seed:      uint64(8000 + i),
						Rounds:    rounds,
						Attackers: cloneSides([]battle.Side{*atkSide}),
						Defenders: cloneSides([]battle.Side{*defSide}),
						Rapidfire: rf,
					}
					rep, err := battle.Calculate(in)
					if err != nil {
						continue
					}
					atkLosses = append(atkLosses, sumLoss(rep.Attackers))
					defLosses = append(defLosses, sumLoss(rep.Defenders))
				}
				al := median(atkLosses)
				dl := median(defLosses)
				cell := ""
				switch {
				case al == 0 && dl == 0:
					cell = "—"
				case al == 0:
					cell = "ATK-clean"
				case dl == 0:
					cell = "def-clean"
				default:
					cell = fmt.Sprintf("%.2f", float64(dl)/float64(al))
				}
				fmt.Printf(" %14s", cell)
			}
			fmt.Println()
		}
	}
}

// buildSideFromGroup — собрать сторону по распределению cost.
func buildSideFromGroup(cat *config.Catalog, dist map[string]float64, budget int64) *battle.Side {
	var units []battle.Unit
	for key, frac := range dist {
		subBudget := int64(float64(budget) * frac)
		side := buildSideFromKey(cat, key, subBudget)
		if side == nil {
			continue
		}
		units = append(units, side.Units...)
	}
	if len(units) == 0 {
		return nil
	}
	return &battle.Side{UserID: "group", Units: units}
}

type scn struct {
	Name      string
	Descr     string
	Attackers []battle.Side
	Defenders []battle.Side
}

func runScenario(s scn, rf map[int]map[int]int, runs, rounds int) {
	fmt.Printf("\n=== %s ===\n%s\n", s.Name, s.Descr)
	fmt.Printf("Attacker cost: %s\n", costLine(s.Attackers))
	fmt.Printf("Defender cost: %s\n", costLine(s.Defenders))

	atkWins, defWins, draws := 0, 0, 0
	var atkLosses, defLosses []int64 // metal-eq per run
	var roundsList []int

	for i := 0; i < runs; i++ {
		in := battle.Input{
			Seed:      uint64(1000 + i),
			Rounds:    rounds,
			Attackers: cloneSides(s.Attackers),
			Defenders: cloneSides(s.Defenders),
			Rapidfire: rf,
		}
		rep, err := battle.Calculate(in)
		if err != nil {
			fmt.Printf("  ERROR run %d: %v\n", i, err)
			continue
		}
		switch rep.Winner {
		case "attackers":
			atkWins++
		case "defenders":
			defWins++
		default:
			draws++
		}
		roundsList = append(roundsList, rep.Rounds)
		atkLosses = append(atkLosses, sumLoss(rep.Attackers))
		defLosses = append(defLosses, sumLoss(rep.Defenders))
	}

	atkCost := totalCost(s.Attackers)
	defCost := totalCost(s.Defenders)

	fmt.Printf("\nRuns: %d\n", runs)
	fmt.Printf("  attacker wins: %d (%.1f%%)\n", atkWins, pct(atkWins, runs))
	fmt.Printf("  defender wins: %d (%.1f%%)\n", defWins, pct(defWins, runs))
	fmt.Printf("  draws:         %d (%.1f%%)\n", draws, pct(draws, runs))
	fmt.Printf("  avg rounds:    %.1f\n", avgInt(roundsList))
	fmt.Printf("  median atk loss: %d (%.1f%% of cost)\n", median(atkLosses), pctF(float64(median(atkLosses)), float64(atkCost)))
	fmt.Printf("  median def loss: %d (%.1f%% of cost)\n", median(defLosses), pctF(float64(median(defLosses)), float64(defCost)))

	// Размен ресурсов: (def loss) / (atk loss). >1 — атакующий эффективнее.
	ml := float64(median(atkLosses))
	dl := float64(median(defLosses))
	if ml > 0 {
		fmt.Printf("  exchange ratio (def_loss/atk_loss): %.2f\n", dl/ml)
	}
}

// --- сценарии ---

func buildScenarios(cat *config.Catalog) []scn {
	u := func(key string, qty int64) battle.Unit { return makeUnit(cat, key, qty) }
	def := func(key string, qty int64) battle.Unit { return makeDefUnit(cat, key, qty) }
	side := func(name string, us ...battle.Unit) battle.Side {
		return battle.Side{UserID: name, Username: name, Units: us}
	}

	return []scn{
		{
			Name:  "lancer-vs-cruiser",
			Descr: "BA-002 проверка: 1000 Lancer (25M-eq) vs 1000 Cruiser (29M-eq). В legacy Cruiser×35 vs Lancer — должен разнести.",
			Attackers: []battle.Side{side("lancer-attacker",
				u("lancer_ship", 1000),
			)},
			Defenders: []battle.Side{side("cruiser-defender",
				u("cruiser", 1000),
			)},
		},
		{
			Name:  "lancer-vs-mixed",
			Descr: "BA-002: 500 Lancer атакуют микс 200 LF + 100 Cruiser + 50 BS. Lancer-spam должен провалиться.",
			Attackers: []battle.Side{side("lancer-attacker",
				u("lancer_ship", 500),
			)},
			Defenders: []battle.Side{side("mixed-fleet",
				u("light_fighter", 200),
				u("cruiser", 100),
				u("battle_ship", 50),
			)},
		},
		{
			Name:  "ds-vs-lancer",
			Descr: "BA-001/002: 1 DS (10M-eq) vs 300 Lancer (7.5M-eq). DS×100 vs Lancer — должен выиграть.",
			Attackers: []battle.Side{side("ds-attacker",
				u("death_star", 1),
			)},
			Defenders: []battle.Side{side("lancer-swarm",
				u("lancer_ship", 300),
			)},
		},
		{
			Name:  "ds-vs-bs-fleet",
			Descr: "BA-001: 1 DS vs 200 Battleship (12M-eq). BS пробивает DS (1000>500), но нет rapidfire.",
			Attackers: []battle.Side{side("ds-attacker",
				u("death_star", 1),
			)},
			Defenders: []battle.Side{side("bs-fleet",
				u("battle_ship", 200),
			)},
		},
		{
			Name:  "ds-vs-bs-sd-fleet",
			Descr: "BA-001: 1 DS vs смесь 100 BS + 50 SD (12.25M-eq). Ни у BS, ни у SD нет rapidfire vs DS.",
			Attackers: []battle.Side{side("ds-attacker",
				u("death_star", 1),
			)},
			Defenders: []battle.Side{side("bs-sd",
				u("battle_ship", 100),
				u("star_destroyer", 50),
			)},
		},
		{
			Name:  "bomber-vs-rl",
			Descr: "Balance: 100 Bomber (9M) vs 3000 RocketLauncher (6M). Bomber×20 vs RL из legacy.",
			Attackers: []battle.Side{side("bomber-attacker",
				u("bomber", 100),
			)},
			Defenders: []battle.Side{side("rl-wall",
				def("rocket_launcher", 3000),
			)},
		},
		{
			Name:  "cruiser-vs-rl",
			Descr: "Balance: 200 Cruiser (5.8M) vs 2000 RocketLauncher (4M). Cruiser×10 vs RL из legacy.",
			Attackers: []battle.Side{side("cruiser-attacker",
				u("cruiser", 200),
			)},
			Defenders: []battle.Side{side("rl-wall",
				def("rocket_launcher", 2000),
			)},
		},
		{
			Name:  "mixed-vs-mixed",
			Descr: "Общий sanity: 500 LF + 200 Cruiser + 50 BS vs 1000 RL + 300 LL + 50 Plasma.",
			Attackers: []battle.Side{side("mixed-atk",
				u("light_fighter", 500),
				u("cruiser", 200),
				u("battle_ship", 50),
			)},
			Defenders: []battle.Side{side("mixed-def",
				def("rocket_launcher", 1000),
				def("light_laser", 300),
				def("plasma_gun", 50),
			)},
		},

		// === ПЛАН 27: ГЛУБОКИЙ АНАЛИЗ ===

		// --- SSat-эксплойт ---
		{
			Name:  "ssat-trap-vs-ds",
			Descr: "Эксплойт: 50000 Solar Satellite (125M-eq) vs 1 DS. SSat должен 'красть' выстрелы DS×1250.",
			Attackers: []battle.Side{side("ds", u("death_star", 1))},
			Defenders: []battle.Side{side("ssat-wall", u("solar_satellite", 50000))},
		},
		{
			Name:  "ssat-trap-vs-ds-fleet",
			Descr: "Защита: 10000 SSat (25M-eq) + 50 BS прикрывает от 5 DS. Проверяем поглощение DS-атак.",
			Attackers: []battle.Side{side("ds-fleet", u("death_star", 5))},
			Defenders: []battle.Side{side("ssat-bs",
				u("solar_satellite", 10000),
				u("battle_ship", 50),
			)},
		},

		// --- Frigate role check ---
		{
			Name:  "frigate-vs-cruiser",
			Descr: "Frigate-role: 200 Frigate (17M-eq) vs 600 Cruiser (17.4M-eq) — Frigate×4 vs Cruiser в legacy.",
			Attackers: []battle.Side{side("frigate", u("frigate", 200))},
			Defenders: []battle.Side{side("cruiser", u("cruiser", 600))},
		},
		{
			Name:  "frigate-vs-bs",
			Descr: "Frigate-role: 200 Frigate (17M-eq) vs 200 BS (12M-eq) — Frigate×7 vs BS в legacy.",
			Attackers: []battle.Side{side("frigate", u("frigate", 200))},
			Defenders: []battle.Side{side("bs", u("battle_ship", 200))},
		},
		{
			Name:  "frigate-vs-sf",
			Descr: "Frigate-role: 100 Frigate (8.5M-eq) vs 1000 Strong Fighter (10M-eq) — Frigate×7 vs SF.",
			Attackers: []battle.Side{side("frigate", u("frigate", 100))},
			Defenders: []battle.Side{side("sf", u("strong_fighter", 1000))},
		},

		// --- Shadow Ship проверка после ADR-0007 ---
		{
			Name:  "shadow-vs-ds",
			Descr: "Shadow anti-DS (ADR-0007): 100 Shadow (500k-eq) vs 1 DS — должны наносить ощутимый урон через RF×70.",
			Attackers: []battle.Side{side("shadow", u("shadow_ship", 100))},
			Defenders: []battle.Side{side("ds", u("death_star", 1))},
		},
		{
			Name:  "shadow-mass-vs-ds",
			Descr: "Shadow mass: 1000 Shadow (5M-eq) vs 1 DS — масштабная атака stealth-флотом.",
			Attackers: []battle.Side{side("shadow-mass", u("shadow_ship", 1000))},
			Defenders: []battle.Side{side("ds", u("death_star", 1))},
		},
		{
			Name:  "shadow-vs-mixed",
			Descr: "Shadow vs средний флот: 500 Shadow (2.5M-eq) vs 100 BS + 50 Cruiser + 200 LF.",
			Attackers: []battle.Side{side("shadow", u("shadow_ship", 500))},
			Defenders: []battle.Side{side("mixed",
				u("battle_ship", 100),
				u("cruiser", 50),
				u("light_fighter", 200),
			)},
		},

		// --- Star Destroyer vs Battleship — есть ли у SD ниша? ---
		{
			Name:  "sd-vs-bs",
			Descr: "SD vs BS (одинаковая metal-eq): 50 SD (6.25M) vs 100 BS (6M) — должны быть близки.",
			Attackers: []battle.Side{side("sd", u("star_destroyer", 50))},
			Defenders: []battle.Side{side("bs", u("battle_ship", 100))},
		},
		{
			Name:  "sd-vs-frigate",
			Descr: "SD имеет RF×2 vs Frigate (legacy). 50 SD (6.25M) vs 100 Frigate (8.5M).",
			Attackers: []battle.Side{side("sd", u("star_destroyer", 50))},
			Defenders: []battle.Side{side("frigate", u("frigate", 100))},
		},

		// --- Mass fleet vs DS (другие способы убить DS) ---
		{
			Name:  "ds-vs-bs-mass",
			Descr: "Mass-BS: 1 DS vs 1000 BS (60M-eq). 6× ресурсов vs DS — должно убивать.",
			Attackers: []battle.Side{side("ds", u("death_star", 1))},
			Defenders: []battle.Side{side("bs-mass", u("battle_ship", 1000))},
		},
		{
			Name:  "ds-vs-bomber-mass",
			Descr: "Bomber vs DS: 200 Bomber (18M) vs 1 DS. У Bomber нет RF vs DS — только Attack=900 пробивает.",
			Attackers: []battle.Side{side("bomber", u("bomber", 200))},
			Defenders: []battle.Side{side("ds", u("death_star", 1))},
		},
		{
			Name:  "ds-vs-plasma-mass",
			Descr: "Plasma защищает планету: 1 DS vs 50 Plasma Gun (6.5M). Plasma RF×2 vs DS.",
			Attackers: []battle.Side{side("ds", u("death_star", 1))},
			Defenders: []battle.Side{side("plasma", def("plasma_gun", 50))},
		},
		{
			Name:  "ds-vs-gauss-wall",
			Descr: "Gauss-wall: 1 DS vs 200 Gauss (7.4M). Gauss attack=1100 (>500), но без RF.",
			Attackers: []battle.Side{side("ds", u("death_star", 1))},
			Defenders: []battle.Side{side("gauss", def("gauss_gun", 200))},
		},

		// --- Defense scaling ---
		{
			Name:  "lf-mass-vs-rl-ll",
			Descr: "Cheap fleet vs cheap defense: 5000 LF (20M) vs 5000 RL + 1000 LL (12M).",
			Attackers: []battle.Side{side("lf", u("light_fighter", 5000))},
			Defenders: []battle.Side{side("def",
				def("rocket_launcher", 5000),
				def("light_laser", 1000),
			)},
		},
		{
			Name:  "bs-vs-plasma",
			Descr: "Anti-defense capital: 200 BS (12M) vs 50 Plasma (6.5M).",
			Attackers: []battle.Side{side("bs", u("battle_ship", 200))},
			Defenders: []battle.Side{side("plasma", def("plasma_gun", 50))},
		},
		{
			Name:  "bomber-vs-plasma",
			Descr: "Bomber специалист по defense: 100 Bomber (9M) vs 50 Plasma (6.5M). Bomber имеет RF vs RL/LL/SL/IG но НЕ vs Plasma.",
			Attackers: []battle.Side{side("bomber", u("bomber", 100))},
			Defenders: []battle.Side{side("plasma", def("plasma_gun", 50))},
		},
		{
			Name:  "bomber-vs-gauss",
			Descr: "Bomber vs Gauss: 100 Bomber (9M) vs 100 Gauss (3.7M). У Bomber нет RF vs Gauss.",
			Attackers: []battle.Side{side("bomber", u("bomber", 100))},
			Defenders: []battle.Side{side("gauss", def("gauss_gun", 100))},
		},

		// --- Strong Fighter — есть ли роль? ---
		{
			Name:  "sf-vs-rl",
			Descr: "SF vs defense: 500 SF (5M) vs 2000 RL (4M).",
			Attackers: []battle.Side{side("sf", u("strong_fighter", 500))},
			Defenders: []battle.Side{side("rl", def("rocket_launcher", 2000))},
		},
		{
			Name:  "sf-vs-lf",
			Descr: "SF vs LF (одинаковая metal-eq): 250 SF (2.5M) vs 600 LF (2.4M).",
			Attackers: []battle.Side{side("sf", u("strong_fighter", 250))},
			Defenders: []battle.Side{side("lf", u("light_fighter", 600))},
		},

		// --- Lancer vs новые цели ---
		{
			Name:  "lancer-vs-bs",
			Descr: "Lancer vs тяжёлый флот: 500 Lancer (55M) vs 1000 BS (60M).",
			Attackers: []battle.Side{side("lancer", u("lancer_ship", 500))},
			Defenders: []battle.Side{side("bs", u("battle_ship", 1000))},
		},
		{
			Name:  "lancer-vs-sf",
			Descr: "Lancer vs SF (ловит на дешёвом флоте): 100 Lancer (11M) vs 1000 SF (10M).",
			Attackers: []battle.Side{side("lancer", u("lancer_ship", 100))},
			Defenders: []battle.Side{side("sf", u("strong_fighter", 1000))},
		},
		{
			Name:  "lancer-vs-plasma",
			Descr: "Lancer vs Plasma: 200 Lancer (22M) vs 100 Plasma (13M).",
			Attackers: []battle.Side{side("lancer", u("lancer_ship", 200))},
			Defenders: []battle.Side{side("plasma", def("plasma_gun", 100))},
		},

		// --- Recycler в бою ---
		{
			Name:  "recycler-in-fleet",
			Descr: "Уязвимость Recycler: 100 BS + 100 Recycler (7.8M) vs 200 Cruiser (5.8M). Recycler — мягкая цель.",
			Attackers: []battle.Side{side("mixed",
				u("battle_ship", 100),
				u("recycler", 100),
			)},
			Defenders: []battle.Side{side("cruiser", u("cruiser", 200))},
		},

		// --- Endgame: огромные армии ---
		{
			Name:  "huge-fleet-vs-huge-fleet",
			Descr: "Endgame: 100 BS + 50 SD + 200 Cruiser vs 100 BS + 50 SD + 200 Cruiser (зеркало).",
			Attackers: []battle.Side{side("a",
				u("battle_ship", 100),
				u("star_destroyer", 50),
				u("cruiser", 200),
			)},
			Defenders: []battle.Side{side("b",
				u("battle_ship", 100),
				u("star_destroyer", 50),
				u("cruiser", 200),
			)},
		},
		{
			Name:  "ds-fleet-vs-ds-fleet",
			Descr: "Endgame mirror: 5 DS + 100 BS vs 5 DS + 100 BS.",
			Attackers: []battle.Side{side("a",
				u("death_star", 5),
				u("battle_ship", 100),
			)},
			Defenders: []battle.Side{side("b",
				u("death_star", 5),
				u("battle_ship", 100),
			)},
		},

		// --- Defense walls и их пределы ---
		{
			Name:  "small-shield-coverage",
			Descr: "Small Shield (front=16) перетягивает огонь: 200 Cruiser vs 1 SS + 500 RL.",
			Attackers: []battle.Side{side("cru", u("cruiser", 200))},
			Defenders: []battle.Side{side("def",
				def("small_shield", 1),
				def("rocket_launcher", 500),
			)},
		},
		{
			Name:  "large-shield-coverage",
			Descr: "Large Shield (front=17) — ультра-приоритетная цель. 1000 BS vs 1 LS + 100 Plasma.",
			Attackers: []battle.Side{side("bs", u("battle_ship", 1000))},
			Defenders: []battle.Side{side("def",
				def("large_shield", 1),
				def("plasma_gun", 100),
			)},
		},

		// --- Light Fighter swarm vs Cruiser (контр) ---
		{
			Name:  "lf-swarm-vs-cruiser",
			Descr: "LF-swarm: 5000 LF (20M) vs 1000 Cruiser (29M). Cruiser RF×6 vs LF.",
			Attackers: []battle.Side{side("lf", u("light_fighter", 5000))},
			Defenders: []battle.Side{side("cru", u("cruiser", 1000))},
		},

		// --- Espionage Sensor в бою ---
		{
			Name:  "esensor-in-fleet",
			Descr: "ESensor мягкая цель: 100 BS + 50 ESensor (6.05M) vs 200 Cruiser (5.8M). Probe rapidfire×5 от всех.",
			Attackers: []battle.Side{side("bs-probe",
				u("battle_ship", 100),
				u("espionage_sensor", 50),
			)},
			Defenders: []battle.Side{side("cru", u("cruiser", 200))},
		},

		// --- Защита: scaling Plasma vs мощного атакующего ---
		{
			Name:  "plasma-wall-vs-bs-mass",
			Descr: "Plasma stack vs BS: 200 Plasma (26M) vs 1000 BS (60M). Какой стек удержит?",
			Attackers: []battle.Side{side("bs", u("battle_ship", 1000))},
			Defenders: []battle.Side{side("plasma", def("plasma_gun", 200))},
		},

		// --- Atk-bias check: возьмёт ли BS+Cru+LF дешевле, чем Bomber+SD ---
		{
			Name:  "trio-vs-defense",
			Descr: "Эффективность атаки: 50 BS + 100 Cru + 500 LF (10.8M) vs 1000 RL + 200 LL + 50 Plasma (8.6M).",
			Attackers: []battle.Side{side("trio",
				u("battle_ship", 50),
				u("cruiser", 100),
				u("light_fighter", 500),
			)},
			Defenders: []battle.Side{side("def",
				def("rocket_launcher", 1000),
				def("light_laser", 200),
				def("plasma_gun", 50),
			)},
		},

		// --- DS-fleet с reasonable количеством vs реалистичной защитой ---
		{
			Name:  "ds-fleet-vs-defended-planet",
			Descr: "Realistic endgame: 10 DS + 100 BS (106M) vs 200 Plasma + 100 Gauss + 5 Large Shield (33M).",
			Attackers: []battle.Side{side("a",
				u("death_star", 10),
				u("battle_ship", 100),
			)},
			Defenders: []battle.Side{side("d",
				def("plasma_gun", 200),
				def("gauss_gun", 100),
				def("large_shield", 5),
			)},
		},

		// --- BS RF чтобы посмотреть BS×5 vs ESensor ---
		{
			Name:  "esensor-stealth-vs-mixed",
			Descr: "Probe-spam: 1000 ESensor (1M) vs 100 Cruiser. Probe не угроза, но забирает выстрелы.",
			Attackers: []battle.Side{side("probe", u("espionage_sensor", 1000))},
			Defenders: []battle.Side{side("cru", u("cruiser", 100))},
		},

		// --- Lancer + BS combo (стоит ли смешивать?) ---
		{
			Name:  "lancer-bs-combo-vs-mixed",
			Descr: "Combo: 100 Lancer + 200 BS (23M) vs 200 Cruiser + 100 BS + 50 SD (18.05M).",
			Attackers: []battle.Side{side("lancer-bs",
				u("lancer_ship", 100),
				u("battle_ship", 200),
			)},
			Defenders: []battle.Side{side("mix",
				u("cruiser", 200),
				u("battle_ship", 100),
				u("star_destroyer", 50),
			)},
		},

		// --- Death Star как защита ---
		{
			Name:  "ds-as-defense-vs-fleet",
			Descr: "Большой флот атакует DS: 500 BS + 200 SD (55M) vs 5 DS (50M).",
			Attackers: []battle.Side{side("fleet",
				u("battle_ship", 500),
				u("star_destroyer", 200),
			)},
			Defenders: []battle.Side{side("ds-def", u("death_star", 5))},
		},

		// --- LT как мул в бою ---
		{
			Name:  "transport-in-fleet",
			Descr: "Transport-уязвимость: 500 LF + 100 LT (12M) vs 200 Cruiser. LT — мягкая цель + rapidfire.",
			Attackers: []battle.Side{side("convoy",
				u("light_fighter", 500),
				u("large_transporter", 100),
			)},
			Defenders: []battle.Side{side("cru", u("cruiser", 200))},
		},

		// --- Shadow + BS (front-test для stealth) ---
		{
			Name:  "shadow-bs-mix-vs-cruiser",
			Descr: "Mix-test: 200 Shadow + 100 BS (7M) vs 200 Cruiser. Низкий front Shadow=7 должен защитить BS.",
			Attackers: []battle.Side{side("mix",
				u("shadow_ship", 200),
				u("battle_ship", 100),
			)},
			Defenders: []battle.Side{side("cru", u("cruiser", 200))},
		},

		// --- Recycler-mass + BS — Recycler front=10 такой же как BS, должен принимать выстрелы ---
		{
			Name:  "recycler-bs-mix",
			Descr: "Recycler в смеси: 100 BS + 50 Recycler (6.5M) vs 200 Cruiser (5.8M). Recycler — front=10.",
			Attackers: []battle.Side{side("mix",
				u("battle_ship", 100),
				u("recycler", 50),
			)},
			Defenders: []battle.Side{side("cru", u("cruiser", 200))},
		},

		{
			Name:  "huge-fleet-vs-ds-defense",
			Descr: "Превосходство 3:1 — 1500 BS + 500 SD (152.5M) vs 5 DS (50M). Должно убивать.",
			Attackers: []battle.Side{side("a",
				u("battle_ship", 1500),
				u("star_destroyer", 500),
			)},
			Defenders: []battle.Side{side("d", u("death_star", 5))},
		},
		{
			Name:  "shadow-mass-vs-ds-defense",
			Descr: "Anti-DS через Shadow: 5000 Shadow (25M) vs 5 DS (50M). Меньше ресурсов, но RF×70.",
			Attackers: []battle.Side{side("a", u("shadow_ship", 5000))},
			Defenders: []battle.Side{side("d", u("death_star", 5))},
		},
		{
			Name:  "lancer-mass-vs-ds-defense",
			Descr: "Anti-DS через Lancer: 1000 Lancer (110M) vs 5 DS (50M). RF×3.",
			Attackers: []battle.Side{side("a", u("lancer_ship", 1000))},
			Defenders: []battle.Side{side("d", u("death_star", 5))},
		},

		// --- Тесты паритета: атакующий ~1.2× ресурсов защитника ---
		{
			Name:  "ds-fleet-vs-equal-defense",
			Descr: "Паритет 1.2:1 — 3 DS + 50 BS (33M) vs 200 Plasma + 100 Gauss (33.7M). Defense должна сдерживать.",
			Attackers: []battle.Side{side("a",
				u("death_star", 3),
				u("battle_ship", 50),
			)},
			Defenders: []battle.Side{side("d",
				def("plasma_gun", 200),
				def("gauss_gun", 100),
			)},
		},
		{
			Name:  "fleet-vs-defense-parity",
			Descr: "Паритет: 100 BS + 200 Cruiser (17.8M) vs 100 Plasma + 50 Gauss (16.85M). Mid-tier паритет.",
			Attackers: []battle.Side{side("a",
				u("battle_ship", 100),
				u("cruiser", 200),
			)},
			Defenders: []battle.Side{side("d",
				def("plasma_gun", 100),
				def("gauss_gun", 50),
			)},
		},

		// --- Колонизация под огнём ---
		{
			Name:  "colony-ship-escort",
			Descr: "Эскорт колонизатора: 5 Colony + 50 BS (3.2M) vs 200 LF.",
			Attackers: []battle.Side{side("escort",
				u("colony_ship", 5),
				u("battle_ship", 50),
			)},
			Defenders: []battle.Side{side("lf", u("light_fighter", 200))},
		},
	}
}

// --- helpers ---

// frontOrDefault — front=10 как default, если в YAML не задан явно.
func frontOrDefault(f int) int {
	if f > 0 {
		return f
	}
	return 10
}

func makeUnit(cat *config.Catalog, key string, qty int64) battle.Unit {
	spec, ok := cat.Ships.Ships[key]
	if !ok {
		panic("unknown ship: " + key)
	}
	return battle.Unit{
		UnitID:   spec.ID,
		Quantity: qty,
		Front:    frontOrDefault(spec.Front),
		Attack:   float64(spec.Attack),
		Shield:   float64(spec.Shield),
		Shell:    float64(spec.Shell),
		Name:     key,
		Cost: battle.UnitCost{
			Metal:    spec.Cost.Metal,
			Silicon:  spec.Cost.Silicon,
			Hydrogen: spec.Cost.Hydrogen,
		},
	}
}

func makeDefUnit(cat *config.Catalog, key string, qty int64) battle.Unit {
	spec, ok := cat.Defense.Defense[key]
	if !ok {
		panic("unknown defense: " + key)
	}
	return battle.Unit{
		UnitID:   spec.ID,
		Quantity: qty,
		Front:    frontOrDefault(spec.Front),
		Attack:   float64(spec.Attack),
		Shield:   float64(spec.Shield),
		Shell:    float64(spec.Shell),
		Name:     key,
		Cost: battle.UnitCost{
			Metal:    spec.Cost.Metal,
			Silicon:  spec.Cost.Silicon,
			Hydrogen: spec.Cost.Hydrogen,
		},
	}
}

func cloneSides(in []battle.Side) []battle.Side {
	out := make([]battle.Side, len(in))
	for i, s := range in {
		s2 := s
		s2.Units = append([]battle.Unit(nil), s.Units...)
		out[i] = s2
	}
	return out
}

func totalCost(sides []battle.Side) int64 {
	var sum int64
	for _, s := range sides {
		for _, u := range s.Units {
			sum += (u.Cost.Metal + u.Cost.Silicon + u.Cost.Hydrogen) * u.Quantity
		}
	}
	return sum
}

func sumLoss(res []battle.SideResult) int64 {
	var sum int64
	for _, r := range res {
		sum += r.LostMetal + r.LostSilicon + r.LostHydrogen
	}
	return sum
}

func costLine(sides []battle.Side) string {
	total := totalCost(sides)
	return fmt.Sprintf("%d metal-eq (%s)", total, unitSummary(sides))
}

func unitSummary(sides []battle.Side) string {
	out := ""
	for _, s := range sides {
		for _, u := range s.Units {
			if out != "" {
				out += ", "
			}
			out += fmt.Sprintf("%d×%s", u.Quantity, u.Name)
		}
	}
	return out
}

func median(xs []int64) int64 {
	if len(xs) == 0 {
		return 0
	}
	s := append([]int64(nil), xs...)
	sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })
	return s[len(s)/2]
}

func avgInt(xs []int) float64 {
	if len(xs) == 0 {
		return 0
	}
	sum := 0
	for _, x := range xs {
		sum += x
	}
	return float64(sum) / float64(len(xs))
}

func pct(n, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(n) / float64(total) * 100
}

func pctF(n, total float64) float64 {
	if total == 0 {
		return 0
	}
	return n / total * 100
}

// multiFlag — flag.Value для повторяющихся --cost=... аргументов.
type multiFlag []string

func (m *multiFlag) String() string     { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error { *m = append(*m, v); return nil }

// applyCostOverride — разобрать "ship_key=M/Si/H" и переписать
// Cost в cat.Ships.Ships[key] (для defense — cat.Defense.Defense[key]).
func applyCostOverride(cat *config.Catalog, spec string) error {
	eq := strings.IndexByte(spec, '=')
	if eq < 0 {
		return fmt.Errorf("нет '=' в %q", spec)
	}
	key := strings.TrimSpace(spec[:eq])
	parts := strings.Split(spec[eq+1:], "/")
	if len(parts) != 3 {
		return fmt.Errorf("ожидалось M/Si/H, получено %q", spec[eq+1:])
	}
	vals := make([]int64, 3)
	for i, p := range parts {
		n, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
		if err != nil {
			return fmt.Errorf("parse %q: %w", p, err)
		}
		vals[i] = n
	}
	newCost := config.ResCost{Metal: vals[0], Silicon: vals[1], Hydrogen: vals[2]}

	if s, ok := cat.Ships.Ships[key]; ok {
		s.Cost = newCost
		cat.Ships.Ships[key] = s
		fmt.Fprintf(os.Stderr, "cost override: ship %s = %d/%d/%d (metal-eq %d)\n",
			key, vals[0], vals[1], vals[2], vals[0]+vals[1]+vals[2])
		return nil
	}
	if d, ok := cat.Defense.Defense[key]; ok {
		d.Cost = newCost
		cat.Defense.Defense[key] = d
		fmt.Fprintf(os.Stderr, "cost override: defense %s = %d/%d/%d\n",
			key, vals[0], vals[1], vals[2])
		return nil
	}
	return fmt.Errorf("юнит %q не найден в ships/defense", key)
}

// applyFrontOverride — "ship_key=N" → переписать Front в ShipSpec/DefenseSpec.
func applyFrontOverride(cat *config.Catalog, spec string) error {
	eq := strings.IndexByte(spec, '=')
	if eq < 0 {
		return fmt.Errorf("нет '=' в %q", spec)
	}
	key := strings.TrimSpace(spec[:eq])
	val, err := strconv.Atoi(strings.TrimSpace(spec[eq+1:]))
	if err != nil {
		return fmt.Errorf("parse front %q: %w", spec[eq+1:], err)
	}
	if s, ok := cat.Ships.Ships[key]; ok {
		s.Front = val
		cat.Ships.Ships[key] = s
		fmt.Fprintf(os.Stderr, "front override: ship %s = %d\n", key, val)
		return nil
	}
	if d, ok := cat.Defense.Defense[key]; ok {
		d.Front = val
		cat.Defense.Defense[key] = d
		fmt.Fprintf(os.Stderr, "front override: defense %s = %d\n", key, val)
		return nil
	}
	return fmt.Errorf("юнит %q не найден в ships/defense", key)
}
