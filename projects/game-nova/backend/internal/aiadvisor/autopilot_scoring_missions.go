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
	unitLightFighter     = 31
	unitCruiser          = 33
	unitBattleship       = 34
	unitEspionageProbe   = 38

	// research unit ids (для лимитов миссий).
	unitAstrophysics  = 27
	unitEspionageTech = 13
)

// Минимальная ценность флота экспедиции (metal-eq), чтобы Compute не
// предложил «отправить 1 light fighter» — это анти-эксплойт BA-003
// (см. fleet/expedition.go, expeditionMinFleetValue=50_000).
const minExpeditionFleetValue = 50_000

// scoreMissions добавляет mission-кандидаты в общий пул scoreCandidates.
//   Ф.2.1: transport (между своими планетами), expedition.
//   Ф.2.2: attack (упрощённая оценка по очкам), spy (шпионаж соседей).
//   Ф.3:   acs (присоединение к атаке союзника).
func scoreMissions(snap PlayerSnapshot, strategy Strategy, inputs scoringInputs) []Recommendation {
	if inputs.Catalog == nil {
		return nil
	}
	var out []Recommendation
	out = append(out, scoreTransports(snap, strategy, inputs)...)
	out = append(out, scoreExpeditions(snap, strategy, inputs)...)
	out = append(out, scoreAttacks(snap, strategy, inputs)...)
	out = append(out, scoreSpy(snap, strategy, inputs)...)
	out = append(out, scoreACS(snap, strategy)...)
	return out
}

// scoreACS — присоединиться к существующей атаке союзника.
//
// Условия:
//   - Игрок в альянсе (AllianceID != nil — обеспечивается buildSnapshot).
//   - Есть открытые ACS-формации союзников (snap.ACSGroups).
//   - Есть боевые корабли на хотя бы одной планете.
//
// Score: высокий для Military (1.0), низкий для прочих (0.2). ACS даёт
// разделение лута и потерь — выгодно при совместной атаке сильной цели.
func scoreACS(snap PlayerSnapshot, strategy Strategy) []Recommendation {
	if len(snap.ACSGroups) == 0 {
		return nil
	}
	weight := 0.2
	if strategy == StrategyMilitary {
		weight = 1.0
	}

	// Найдём планету с наибольшим боевым флотом.
	srcPlanetID, srcPlanetName := "", ""
	var srcValue int64
	for _, ps := range snap.Planets {
		v := approxAttackerShipValue(ps.Ships)
		if v > srcValue {
			srcValue = v
			srcPlanetID = ps.ID
			srcPlanetName = ps.Name
		}
	}
	if srcValue == 0 {
		return nil
	}

	var out []Recommendation
	for _, g := range snap.ACSGroups {
		// Score базовый = наш флот / 1000 (нормализация). Для ACS не
		// важна сила цели (атаку планирует лидер); важно лишь что мы
		// добавляем к пулу свою силу.
		score := float64(srcValue) / 1000.0 * weight
		if score <= 0 {
			continue
		}
		out = append(out, Recommendation{
			ID:         ids.New(),
			Category:   "mission",
			ActionType: "acs_join",
			PlanetID:   srcPlanetID,
			Params: map[string]any{
				"src_planet_id": srcPlanetID,
				"acs_group_id":  g.ID,
				"dst_galaxy":    g.TargetGalaxy,
				"dst_system":    g.TargetSystem,
				"dst_position":  g.TargetPos,
				"dst_is_moon":   g.TargetIsMoon,
			},
			Score: score,
			Description: fmt.Sprintf("Присоединиться к атаке %q на [%d:%d:%d] с %q",
				g.LeaderName, g.TargetGalaxy, g.TargetSystem, g.TargetPos, srcPlanetName),
			Benefit: fmt.Sprintf("ACS-формация: разделить грабёж и потери (флот %d metal-eq)", srcValue),
		})
	}
	return out
}

// scoreAttacks — упрощённая (без battle.Simulator) оценка атаки.
//
// Кандидат-цель отсеивается:
//   - Сосед в umode/observer/protection — нельзя атаковать.
//   - TotalScore цели > 0.5 × (мои очки) — слишком сильный, риск.
//   - У меня нет атакующих кораблей (light_fighter/cruiser/battleship).
//
// Score = (мои военные очки / очки цели) * weight.
//   weight: Military=1.0, Defense/Economy/Expansion=0.2.
//
// Полный battle-симулятор (с прогнозом потерь и грабежа) вынесен
// в отдельную задачу — это самостоятельная фича объёмом ~500 строк.
// На уровне Ф.2.2 атака предлагается как «есть кого ограбить»,
// без точного количества loot.
func scoreAttacks(snap PlayerSnapshot, strategy Strategy, _ scoringInputs) []Recommendation {
	weight := 0.2
	if strategy == StrategyMilitary {
		weight = 1.0
	}

	// Суммарная атакующая сила игрока (proxy: ценность боевых кораблей).
	var attackerValue int64
	srcPlanetID, srcPlanetName := "", ""
	for _, ps := range snap.Planets {
		val := approxAttackerShipValue(ps.Ships)
		if val > attackerValue {
			attackerValue = val
			srcPlanetID = ps.ID
			srcPlanetName = ps.Name
		}
	}
	if attackerValue == 0 || srcPlanetID == "" {
		return nil
	}

	var out []Recommendation
	for _, n := range snap.Neighbors {
		if n.Umode || n.IsObserver || n.HasProtection {
			continue
		}
		// Защита: не атакуем сильнее себя.
		if snap.Score > 0 && n.TotalScore > 0.5*snap.Score {
			continue
		}
		// Защита: не атакуем тех, кто сильно слабее в bashing-протекшен
		// (у legacy 1:5 порог защищает «новичка» от старшего). Здесь —
		// просто пропускаем целей с очень низкими очками (< 1000).
		if n.TotalScore < 1000 {
			continue
		}

		// Score = (моя сила / сила цели) — больше отдача когда цель
		// много слабее игрока.
		ratio := float64(attackerValue) / math.Max(n.TotalScore, 1)
		score := ratio * weight
		if score <= 0 {
			continue
		}

		out = append(out, Recommendation{
			ID:         ids.New(),
			Category:   "mission",
			ActionType: "attack",
			PlanetID:   srcPlanetID,
			Params: map[string]any{
				"src_planet_id":   srcPlanetID,
				"target_user_id":  n.UserID,
				"dst_galaxy":      n.Galaxy,
				"dst_system":      n.System,
				"dst_position":    n.Position,
				"target_points":   n.TotalScore,
				"attacker_value":  attackerValue,
			},
			Score: score,
			Description: fmt.Sprintf("Атака игрока [%d:%d:%d] (очки %.0f) с %q",
				n.Galaxy, n.System, n.Position, n.TotalScore, srcPlanetName),
			Benefit: fmt.Sprintf("Грабёж ресурсов цели; ваш флот %d metal-eq vs очки цели %.0f",
				attackerValue, n.TotalScore),
		})
	}
	return out
}

// scoreSpy — отправка espionage probes на соседа, чтобы получить
// отчёт об уровне шпионажа/ресурсах/флоте.
//
// Условия:
//   - Сосед не в protection (шпионаж по защищённым тоже работает в
//     legacy, но автопилот предпочитает сначала видимых не-защищённых).
//   - У игрока есть espionage_tech ≥ 1 + хотя бы 1 espionage_probe (id=38).
//
// Score = низкий abs (1 + log(target_points)) * weight; в Military
// и Defense чуть выше, в Economy/Expansion — нейтрально.
func scoreSpy(snap PlayerSnapshot, strategy Strategy, _ scoringInputs) []Recommendation {
	if snap.Research[unitEspionageTech] < 1 {
		return nil
	}
	weight := 0.3
	switch strategy {
	case StrategyMilitary, StrategyDefense:
		weight = 0.6
	}

	srcPlanetID, srcPlanetName := "", ""
	probesAvail := int64(0)
	for _, ps := range snap.Planets {
		if ps.Ships[unitEspionageProbe] > 0 {
			if ps.Ships[unitEspionageProbe] > probesAvail {
				probesAvail = ps.Ships[unitEspionageProbe]
				srcPlanetID = ps.ID
				srcPlanetName = ps.Name
			}
		}
	}
	if srcPlanetID == "" || probesAvail < 1 {
		return nil
	}

	// Кол-во probe per миссия — обычно 4..8 (хороший баланс точности).
	probesPerMission := int64(4)
	if probesAvail < probesPerMission {
		probesPerMission = probesAvail
	}

	var out []Recommendation
	for _, n := range snap.Neighbors {
		if n.Umode || n.IsObserver || n.HasProtection {
			continue
		}
		if n.TotalScore < 100 {
			continue // нечего шпионить
		}
		score := math.Log10(n.TotalScore+10) * weight

		out = append(out, Recommendation{
			ID:         ids.New(),
			Category:   "mission",
			ActionType: "spy",
			PlanetID:   srcPlanetID,
			Params: map[string]any{
				"src_planet_id":  srcPlanetID,
				"target_user_id": n.UserID,
				"dst_galaxy":     n.Galaxy,
				"dst_system":     n.System,
				"dst_position":   n.Position,
				"probes":         probesPerMission,
			},
			Score: score,
			Description: fmt.Sprintf("Шпионаж [%d:%d:%d] (%d probes с %q)",
				n.Galaxy, n.System, n.Position, probesPerMission, srcPlanetName),
			Benefit: fmt.Sprintf("Отчёт о флоте и ресурсах цели (%d probes)", probesPerMission),
		})
	}
	return out
}

// approxAttackerShipValue — суммарная metal-eq стоимость боевых кораблей
// на планете.
//
// Учитываются только бойцы (не транспорты), чтобы score атаки не
// учитывал транспортный тоннаж как «силу».
func approxAttackerShipValue(ships map[int]int64) int64 {
	var v int64
	v += ships[unitLightFighter] * approxShipValue(unitLightFighter)
	v += ships[32] * approxShipValue(32) // heavy_fighter
	v += ships[unitCruiser] * approxShipValue(unitCruiser)
	v += ships[unitBattleship] * approxShipValue(unitBattleship)
	v += ships[39] * approxShipValue(39) // bomber
	v += ships[41] * approxShipValue(41) // destroyer
	v += ships[42] * approxShipValue(42) // deathstar
	return v
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
	// Запоминаем координаты — для expedition target = (galaxy, system,
	// position=15 — позиция за обычными 1..14, в которую летит флот
	// в неисследованную зону, согласно legacy oxsar2).
	bestPlanetID := ""
	bestPlanetName := ""
	var bestValue int64
	var srcGalaxy, srcSystem int
	for _, ps := range snap.Planets {
		if ps.IsMoon {
			continue // экспедиции отправляются с планет, не с лун
		}
		v := metalEqShipValue(ps.Ships, inputs)
		if v > bestValue {
			bestValue = v
			bestPlanetID = ps.ID
			bestPlanetName = ps.Name
			srcGalaxy = ps.Galaxy
			srcSystem = ps.System
		}
	}
	if bestPlanetID == "" || bestValue < minExpeditionFleetValue {
		return nil
	}

	rawScore := float64(astroLvl+5) * float64(bestValue) / float64(minExpeditionFleetValue)
	score := rawScore * weight

	// Координаты экспедиции — позиция 15 в системе планеты-источника.
	// fleet.Send для KindExpedition не проверяет существование цели
	// (см. transport.go:349-355), так что произвольная позиция работает.
	const expeditionPosition = 15

	return []Recommendation{{
		ID:         ids.New(),
		Category:   "mission",
		ActionType: "expedition",
		PlanetID:   bestPlanetID,
		Params: map[string]any{
			"src_planet_id":    bestPlanetID,
			"astrophysics_lvl": astroLvl,
			"fleet_value":      bestValue,
			"dst_galaxy":       srcGalaxy,
			"dst_system":       srcSystem,
			"dst_position":     expeditionPosition,
		},
		Score:       score,
		Description: fmt.Sprintf("Отправить экспедицию с планеты %q [%d:%d:%d] (астрофизика ур.%d)",
			bestPlanetName, srcGalaxy, srcSystem, expeditionPosition, astroLvl),
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
