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
		scenario   = flag.String("scenario", "", "имя сценария (lancer-vs-cruiser, ds-vs-fleet, bomber-vs-defense, lancer-vs-ds, cruiser-vs-rl)")
		all        = flag.Bool("all", false, "прогнать все сценарии")
		runs       = flag.Int("runs", 50, "число прогонов на сценарий")
		rounds     = flag.Int("rounds", 6, "макс раундов в одном бою")
		catalogDir = flag.String("configs", "../../../configs", "путь к configs/")
		costOverride multiFlag
	)
	flag.Var(&costOverride, "cost", "переопределить стоимость юнита: ship_key=M/Si/H (можно несколько)")
	flag.Parse()

	if *scenario == "" && !*all {
		return fmt.Errorf("нужен --scenario=<name> или --all")
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

	rf := cat.Rapidfire.Rapidfire

	for _, s := range toRun {
		runScenario(s, rf, *runs, *rounds)
	}
	return nil
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
	}
}

// --- helpers ---

func makeUnit(cat *config.Catalog, key string, qty int64) battle.Unit {
	spec, ok := cat.Ships.Ships[key]
	if !ok {
		panic("unknown ship: " + key)
	}
	return battle.Unit{
		UnitID:   spec.ID,
		Quantity: qty,
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
