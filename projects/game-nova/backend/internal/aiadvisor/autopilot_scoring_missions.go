package aiadvisor

import (
	"fmt"
	"math"

	"oxsar/game-nova/pkg/ids"
)

// Файл реализует scoring миссий — действий «отправить флот»: транспорт,
// экспедиция (Ф.2.1), атака, шпионаж (Ф.2.2 — отдельный коммит).
//
// Источники истины формул:
//   - Лимиты слотов фльотов / экспедиций:  fleet/transport.go (1+computer/6,
//     floor(sqrt(astro))).
//   - Минимальная ценность флота для экспедиции: 50_000 metal-eq
//     (expeditionMinFleetValue).
//   - cargo / cost / speed кораблей:        configs/ships.yml.

// Unit IDs кораблей-транспортов и исследований.
// Эти id жёстко зашиты в yaml и совпадают с legacy oxsar2.
const (
	unitSmallTransporter = 29
	unitLargeTransporter = 30

	// research unit ids (для лимитов миссий).
	unitAstrophysics = 27
)

// Минимальная ценность флота экспедиции (metal-eq), чтобы Compute не
// предложил «отправить 1 light fighter» — это анти-эксплойт BA-003
// (см. fleet/expedition.go, expeditionMinFleetValue=50_000).
const minExpeditionFleetValue = 50_000

// scoreMissions добавляет mission-кандидаты в общий пул scoreCandidates.
// На уровне Ф.2.1 поддерживаются: transport (между своими планетами)
// и expedition. Атака и шпионаж — Ф.2.2.
func scoreMissions(snap PlayerSnapshot, strategy Strategy, inputs scoringInputs) []Recommendation {
	if inputs.Catalog == nil {
		return nil
	}
	var out []Recommendation
	out = append(out, scoreTransports(snap, strategy, inputs)...)
	out = append(out, scoreExpeditions(snap, strategy, inputs)...)
	return out
}

// scoreTransports — предлагает перемещение ресурсов с планеты-донора
// на планету-реципиента.
//
// Кандидат:
//   - На планете-доноре есть large/small transporter.
//   - Хотя бы один ресурс заполнен ≥ 80% capacity (профицит).
//   - На целевой планете тот же ресурс < 30% capacity (дефицит).
//
// Score (для Economy высокий, для прочих стратегий низкий — транспорт
// нейтрален к стратегии, но критичен для Economy):
//   score = transferAmount * weight, weight = {Economy: 1e-4, прочие: 5e-6}.
func scoreTransports(snap PlayerSnapshot, strategy Strategy, _ scoringInputs) []Recommendation {
	if len(snap.Planets) < 2 {
		return nil
	}

	weight := 5e-6
	if strategy == StrategyEconomy {
		weight = 1e-4
	}

	var out []Recommendation
	for i, donor := range snap.Planets {
		// Доступная грузоподъёмность на доноре.
		cargo := availableCargo(donor)
		if cargo <= 0 {
			continue
		}

		for j, recipient := range snap.Planets {
			if i == j {
				continue
			}
			rec := bestTransfer(donor, recipient, cargo)
			if rec == nil {
				continue
			}
			rec.Score *= weight
			if rec.Score <= 0 {
				continue
			}
			out = append(out, *rec)
		}
	}
	return out
}

// availableCargo — суммарный грузоподъём всех transporter'ов на планете
// (small=5000, large=25000). Возвращает 0 если транспортных нет или
// планета-луна без транспортных.
func availableCargo(ps PlanetSnapshot) int64 {
	var cargo int64
	if c, ok := ps.Ships[unitSmallTransporter]; ok {
		cargo += c * 5000
	}
	if c, ok := ps.Ships[unitLargeTransporter]; ok {
		cargo += c * 25000
	}
	return cargo
}

// bestTransfer строит рекомендацию-кандидат для одной пары донор→реципиент.
// Возвращает nil, если профицит/дефицит/cargo не сходятся.
func bestTransfer(donor, recipient PlanetSnapshot, cargoLimit int64) *Recommendation {
	type resCandidate struct {
		key         string
		surplus     float64 // сколько лишнего на доноре
		deficit     float64 // сколько не хватает на реципиенте до 50% cap
		donorCap    float64
		recipientCap float64
	}
	cands := []resCandidate{
		{"metal", donor.Resources.Metal, 0, donor.Capacity.Metal, recipient.Capacity.Metal},
		{"silicon", donor.Resources.Silicon, 0, donor.Capacity.Silicon, recipient.Capacity.Silicon},
		{"hydrogen", donor.Resources.Hydrogen, 0, donor.Capacity.Hydrogen, recipient.Capacity.Hydrogen},
	}
	cands[0].deficit = math.Max(0, recipient.Capacity.Metal*0.5-recipient.Resources.Metal)
	cands[1].deficit = math.Max(0, recipient.Capacity.Silicon*0.5-recipient.Resources.Silicon)
	cands[2].deficit = math.Max(0, recipient.Capacity.Hydrogen*0.5-recipient.Resources.Hydrogen)

	var best *resCandidate
	for i := range cands {
		c := &cands[i]
		// Донор: ресурс ≥ 80% capacity.
		if c.donorCap <= 0 || c.surplus < c.donorCap*0.8 {
			continue
		}
		// Реципиент: ресурс < 30% capacity.
		hit := false
		switch c.key {
		case "metal":
			hit = recipient.Resources.Metal < recipient.Capacity.Metal*0.3
		case "silicon":
			hit = recipient.Resources.Silicon < recipient.Capacity.Silicon*0.3
		case "hydrogen":
			hit = recipient.Resources.Hydrogen < recipient.Capacity.Hydrogen*0.3
		}
		if !hit || c.deficit <= 0 {
			continue
		}
		if best == nil || c.deficit > best.deficit {
			best = c
		}
	}
	if best == nil {
		return nil
	}
	// Переносимый объём — min(deficit, surplus-50%cap, cargoLimit).
	canSend := best.surplus - best.donorCap*0.5
	transfer := math.Min(best.deficit, math.Min(canSend, float64(cargoLimit)))
	if transfer < 1000 {
		return nil // совсем мелкие переносы не предлагаем
	}

	carry := map[string]int64{
		"metal":    0,
		"silicon":  0,
		"hydrogen": 0,
	}
	carry[best.key] = int64(transfer)

	return &Recommendation{
		ID:         ids.New(),
		Category:   "mission",
		ActionType: "transport",
		PlanetID:   donor.ID,
		Params: map[string]any{
			"src_planet_id": donor.ID,
			"dst_galaxy":    recipient.Galaxy,
			"dst_system":    recipient.System,
			"dst_position":  recipient.Position,
			"dst_is_moon":   recipient.IsMoon,
			"resource":      best.key,
			"amount":        int64(transfer),
			"carry_metal":   carry["metal"],
			"carry_silicon": carry["silicon"],
			"carry_hydrogen": carry["hydrogen"],
		},
		Score:       transfer, // domain-эффект; weight применяется наружу
		Description: fmt.Sprintf("Транспорт %d %s с %q → %q [%d:%d:%d]",
			int64(transfer), best.key, donor.Name, recipient.Name,
			recipient.Galaxy, recipient.System, recipient.Position),
		Benefit: fmt.Sprintf("Снять заполненность %q и пополнить дефицит %q",
			donor.Name, recipient.Name),
	}
}

// scoreExpeditions — экспедиция в неисследованную зону.
//
// Условия:
//   - У игрока есть исследование astrophysics ≥ 1 (унитid 27).
//   - Есть свободный expedition-слот (max=floor(sqrt(astro)), used=count
//     текущих exp-флотов).
//   - На любой планете-источнике есть флот с metal-eq value ≥ 50_000
//     (анти-фарм-порог BA-003).
//
// Score рассчитывается по упрощённой формуле:
//   score = (astro_level + 5) * fleet_value / 50000
// и далее * stratWeight (Expansion=1.5, Military=0.5, прочие 0.3 — так
// как экспедиция даёт смесь ресурсов/артефактов/планет).
func scoreExpeditions(snap PlayerSnapshot, strategy Strategy, inputs scoringInputs) []Recommendation {
	astroLvl := snap.Research[unitAstrophysics]
	if astroLvl < 1 {
		return nil
	}
	maxSlots := int(math.Sqrt(float64(astroLvl)))
	if maxSlots < 1 {
		maxSlots = 1
	}
	usedSlots := countExpeditionFleets(snap)
	if usedSlots >= maxSlots {
		return nil
	}

	weight := expeditionStrategyWeight(strategy)
	if weight <= 0 {
		return nil
	}

	// Выбираем планету-донор: максимум metal-eq value среди ships.
	bestPlanetID := ""
	bestPlanetName := ""
	var bestValue int64
	for _, ps := range snap.Planets {
		if ps.IsMoon {
			continue // экспедиции отправляются с планет, не с лун
		}
		v := metalEqShipValue(ps.Ships, inputs)
		if v > bestValue {
			bestValue = v
			bestPlanetID = ps.ID
			bestPlanetName = ps.Name
		}
	}
	if bestPlanetID == "" || bestValue < minExpeditionFleetValue {
		return nil
	}

	rawScore := float64(astroLvl+5) * float64(bestValue) / float64(minExpeditionFleetValue)
	score := rawScore * weight

	return []Recommendation{{
		ID:         ids.New(),
		Category:   "mission",
		ActionType: "expedition",
		PlanetID:   bestPlanetID,
		Params: map[string]any{
			"src_planet_id":    bestPlanetID,
			"astrophysics_lvl": astroLvl,
			"fleet_value":      bestValue,
		},
		Score:       score,
		Description: fmt.Sprintf("Отправить экспедицию с планеты %q (астрофизика ур.%d)", bestPlanetName, astroLvl),
		Benefit: fmt.Sprintf("Шанс на ресурсы/артефакты/планету; флот ценой %d metal-eq",
			bestValue),
	}}
}

// expeditionStrategyWeight — стратегический вес экспедиции.
//
// Expansion получает максимум (extra_planet — единственный путь к новым
// планетам кроме колонизации). Economy умеренный (ресурсы); Military и
// Defense меньший — экспедиция отвлекает флот.
func expeditionStrategyWeight(s Strategy) float64 {
	switch s {
	case StrategyExpansion:
		return 1.5
	case StrategyEconomy:
		return 0.5
	case StrategyMilitary, StrategyDefense:
		return 0.3
	}
	return 0.5
}

// countExpeditionFleets — сколько флотов сейчас на миссии экспедиции.
// На уровне snapshot — flight через Mission == int(KindExpedition=15).
func countExpeditionFleets(snap PlayerSnapshot) int {
	const missionExpedition = 15
	n := 0
	for _, f := range snap.Fleets {
		if f.Mission == missionExpedition {
			n++
		}
	}
	return n
}

// metalEqShipValue — суммарная стоимость кораблей в metal-эквиваленте.
// Используется для проверки expeditionMinFleetValue.
//
// Берёт cost.Metal+cost.Silicon+cost.Hydrogen каждого корабля * count.
// Для отсутствующих в каталоге unit-id — 0.
func metalEqShipValue(ships map[int]int64, inputs scoringInputs) int64 {
	if inputs.Catalog == nil {
		return 0
	}
	byID := make(map[int]int64, len(inputs.Catalog.Ships.Ships))
	for _, sp := range inputs.Catalog.Ships.Ships {
		byID[sp.ID] = sp.Cost.Metal + sp.Cost.Silicon + sp.Cost.Hydrogen
	}
	var total int64
	for unitID, count := range ships {
		if v, ok := byID[unitID]; ok {
			total += v * count
		}
	}
	return total
}
