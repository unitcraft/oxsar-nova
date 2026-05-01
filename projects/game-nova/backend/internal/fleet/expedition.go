// EXPEDITION (mission=15) — полёт в неисследованную зону.
//
// Реализует полное портирование с legacy Expedition.class.php.
// Исходы: resources, asteroid, artefact, extra_planet, xSkirmish,
// loss, nothing, battlefield, credit, delay, fast, ship.
// black_hole и unknown не реализованы (не было даже в legacy).
//
// exp_power = expo_tech_level + hours×2 + spy_tech/10 × pow(spy_probes, 0.4)
// Веса исходов зависят от exp_power, hours, visits (штраф повторных посещений).
package fleet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/battle"
	"oxsar/game-nova/internal/battlestats"
	"oxsar/game-nova/internal/config"
	"oxsar/game-nova/internal/event"
	"oxsar/game-nova/pkg/ids"
	"oxsar/game-nova/pkg/rng"
)

// unit IDs для expedition
const (
	unitExpoTech  = 27 // Astrophysics / expo_tech
	unitSpyTech   = 13 // Espionage Technology
	unitSpyProbe  = 38 // Espionage sensor
	unitLF        = 31 // Light Fighter
	unitRecycler  = 37 // Recycler
	unitCruiser   = 33 // Cruiser
	unitDeathstar = 42 // Deathstar
)

// Балансовые константы экспедиций (план 21 блок B).
const (
	// expeditionMinFleetValue — минимум metal-eq флота для экспедиции.
	// 50 000 ≈ 10 Small Transporter'ов или ~1.5 Cruiser. Закрывает
	// фарм-эксплойт BA-003 (отправка 1 LF ради 5M ресурсов).
	expeditionMinFleetValue int64 = 50_000

	// expeditionRewardCapMult — множитель фарм-порога. Суммарная
	// ресурсная награда не может превышать fleet_value × N. План 21 B4.
	expeditionRewardCapMult int64 = 3
)

// ExpeditionHandler — event.Handler для KindExpedition=15.
func (s *TransportService) ExpeditionHandler() event.Handler {
	return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		var pl transportPayload
		if err := json.Unmarshal(e.Payload, &pl); err != nil {
			return fmt.Errorf("expedition: parse payload: %w", err)
		}

		var (
			state       string
			ownerUserID string
			dstGalaxy   int
			dstSystem   int
			departAt    time.Time
			arriveAt    time.Time
			cm, csil, ch int64
		)
		err := tx.QueryRow(ctx, `
			SELECT state, owner_user_id,
			       dst_galaxy, dst_system,
			       depart_at, arrive_at,
			       carried_metal, carried_silicon, carried_hydrogen
			FROM fleets WHERE id = $1 FOR UPDATE
		`, pl.FleetID).Scan(
			&state, &ownerUserID,
			&dstGalaxy, &dstSystem,
			&departAt, &arriveAt,
			&cm, &csil, &ch,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("expedition: read fleet: %w", err)
		}
		if state != "outbound" {
			return nil
		}

		fleetShips, err := readFleetShips(ctx, tx, pl.FleetID)
		if err != nil {
			return fmt.Errorf("expedition: read fleet_ships: %w", err)
		}

		// flight_seconds из payload (записан при создании события в transport.Send).
		// Если поле отсутствует (старые события) — вычисляем из дат.
		flightSeconds := pl.FlightSeconds
		if flightSeconds <= 0 {
			flightSeconds = int64(arriveAt.Sub(departAt).Seconds())
		}
		hours := int(flightSeconds / 3600)

		// exp_power
		expoTech := readResearchLevel(ctx, tx, ownerUserID, unitExpoTech)
		spyTech := readResearchLevel(ctx, tx, ownerUserID, unitSpyTech)
		var spyProbes int64
		for _, s := range fleetShips {
			if s.UnitID == unitSpyProbe {
				spyProbes = s.Count
			}
		}
		expPower := calcExpPower(expoTech, spyTech, spyProbes, hours)

		// visits
		var visits int
		_ = tx.QueryRow(ctx,
			`SELECT COALESCE(visits, 0) FROM expedition_visits WHERE user_id=$1 AND galaxy=$2 AND system=$3`,
			ownerUserID, dstGalaxy, dstSystem,
		).Scan(&visits)

		visitedScale := calcVisitedScale(visits, hours)

		// Обновить счётчик посещений.
		if _, err := tx.Exec(ctx, `
			INSERT INTO expedition_visits (user_id, galaxy, system, visits)
			VALUES ($1, $2, $3, 1)
			ON CONFLICT (user_id, galaxy, system) DO UPDATE SET visits = expedition_visits.visits + 1
		`, ownerUserID, dstGalaxy, dstSystem); err != nil {
			return fmt.Errorf("expedition: upsert visits: %w", err)
		}

		r := rng.New(deriveSeed(pl.FleetID))

		// Подсчёт суммы кораблей и дестаров.
		var totalShips, dsCount int64
		for _, sh := range fleetShips {
			totalShips += sh.Count
			if sh.UnitID == unitDeathstar {
				dsCount += sh.Count
			}
		}

		w := calcExpWeights(expPower, hours, visits, totalShips, dsCount, r)
		outcome := weightedChoice(w, r)

		var (
			reportData map[string]any
		)
		switch outcome {
		case "resource":
			reportData = expResources(ctx, tx, r, expPower, hours, visitedScale,
				pl.FleetID, fleetShips, s.catalog, cm, csil, ch, s.tr)
		case "asteroid":
			reportData = expAsteroid(ctx, tx, r, expPower, hours, visitedScale,
				pl.FleetID, fleetShips, s.catalog, cm, csil, ch, s.tr)
		case "artefact":
			reportData = expArtefact(ctx, tx, r, ownerUserID, s.catalog, s.tr)
		case "extra_planet":
			reportData, err = expExtraPlanet(ctx, tx, r, ownerUserID, s.numGalaxies, s.numSystems, s.tr)
			if err != nil {
				return err
			}
		case "xSkirmish":
			reportData, err = expPirates(ctx, tx, pl.FleetID, ownerUserID, fleetShips, s.catalog, expPower, r, s.tr)
			if err != nil {
				return err
			}
		case "battlefield":
			reportData, err = expBattlefield(ctx, tx, pl.FleetID, ownerUserID, fleetShips, s.catalog, expPower, r, s.tr)
			if err != nil {
				return err
			}
		case "credit":
			reportData, err = expCredit(ctx, tx, r, ownerUserID, expPower, visitedScale)
			if err != nil {
				return err
			}
		case "delay":
			reportData, err = expDelay(ctx, tx, r, pl.ReturnEventID, flightSeconds, s.tr)
			if err != nil {
				return err
			}
		case "fast":
			reportData, err = expFast(ctx, tx, r, pl.ReturnEventID, flightSeconds, s.tr)
			if err != nil {
				return err
			}
		case "ship":
			reportData, err = expShip(ctx, tx, r, pl.FleetID, expPower)
			if err != nil {
				return err
			}
		case "lost":
			reportData, err = expLoss(ctx, tx, r, pl.FleetID, fleetShips)
			if err != nil {
				return err
			}
		default: // nothing
			outcome = "nothing"
			reportData = map[string]any{"message": s.tr("mission", "expeditionNothingFound", nil)}
		}

		reportJSON, _ := json.Marshal(reportData)
		reportID := ids.New()
		if _, err := tx.Exec(ctx, `
			INSERT INTO expedition_reports (id, user_id, fleet_id, outcome, report)
			VALUES ($1, $2, $3, $4, $5)
		`, reportID, ownerUserID, pl.FleetID, outcome, reportJSON); err != nil {
			return fmt.Errorf("expedition: insert report: %w", err)
		}

		subj := s.tr("mission", "expeditionSubject", map[string]string{"outcome": outcome})
		body := s.tr("mission", "expeditionBody", map[string]string{"outcome": outcome})
		if _, err := tx.Exec(ctx, `
			INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body, expedition_report_id)
			VALUES ($1, $2, NULL, 2, $3, $4, $5)
		`, ids.New(), ownerUserID, subj, body, reportID); err != nil {
			return fmt.Errorf("expedition: insert message: %w", err)
		}

		// outcome=lost → флот уничтожен полностью, возврата нет.
		// Для остальных исходов — помечаем как returning, return-event
		// уже создан при отправке.
		newState := "returning"
		if outcome == "lost" {
			newState = "done"
		}
		if _, err := tx.Exec(ctx,
			`UPDATE fleets SET state=$1 WHERE id=$2`, newState, pl.FleetID); err != nil {
			return fmt.Errorf("expedition: update state: %w", err)
		}
		return nil
	}
}

// calcExpPower вычисляет «мощь» экспедиции по формуле legacy.
func calcExpPower(expoTech, spyTech int, spyProbes int64, hours int) float64 {
	spy := float64(spyTech) / 10.0 * math.Pow(float64(spyProbes), 0.4)
	return float64(expoTech) + float64(hours)*2 + spy
}

// calcVisitedScale штраф за повторные посещения системы.
func calcVisitedScale(visits, hours int) float64 {
	var vs float64
	switch {
	case visits >= 20:
		vs = math.Pow(float64(visits), -0.7)
	case visits >= 10:
		vs = math.Pow(float64(visits), -0.5)
	case visits >= 5:
		vs = math.Pow(float64(visits), -0.3)
	case visits >= 3:
		vs = math.Pow(float64(visits), -0.2)
	default:
		vs = 1.0
	}
	hoursScale := math.Pow((float64(hours)+1)/6.0, 1.5)
	return vs * hoursScale
}

// calcExpWeights строит карту весов исходов по формулам legacy.
func calcExpWeights(expPower float64, hours, visits int,
	totalShips, dsCount int64, r *rng.R) map[string]float64 {
	w := map[string]float64{}

	w["resource"] = math.Ceil(100 * math.Pow(1.22, expPower))
	w["asteroid"] = math.Ceil(70 * math.Pow(1.22, expPower))
	if hours == 0 {
		w["asteroid"] += 30
	}

	if hours >= 2 {
		w["ship"] = math.Ceil(20 * math.Pow(1.25, expPower))
	}
	if hours >= 1 {
		w["battlefield"] = math.Ceil(20 * math.Pow(1.25, expPower))
	}
	if hours >= 3 {
		w["xSkirmish"] = math.Ceil(10 * math.Pow(1.26, expPower))
	}
	if hours >= 4 {
		w["artefact"] = math.Ceil(3 * math.Pow(1.28, expPower))
		if xsk := w["xSkirmish"]; xsk > 0 {
			w["artefact"] = math.Min(w["artefact"], xsk/2)
		}
		w["credit"] = math.Ceil(4 * math.Pow(1.28, expPower))
		if xsk := w["xSkirmish"]; xsk > 0 {
			w["credit"] = math.Min(w["credit"], xsk/2)
		}
	}

	visitPower := float64(visits) + expPower/4
	w["delay"] = math.Ceil(30 * math.Pow(1.25, visitPower))
	w["fast"] = math.Ceil(60 * math.Pow(1.25, visitPower))
	w["nothing"] = math.Ceil(40 * math.Pow(1.25, visitPower))
	// B2 (план 21): lost_weight растёт с exp_power, чтобы крупные
	// экспедиции имели пропорционально больший риск. При power=11
	// (новичок) вес 21, при power=20 (хайтек) вес 30. Вероятности:
	// 0.42% → ~1.0% (grows with power). Раньше было фиксированное 10.
	w["lost"] = math.Ceil(10 * (1 + expPower*0.1))

	if totalShips > 10000 {
		w["xSkirmish"] *= 0.5
		w["ship"] *= 0.5
		w["battlefield"] *= 0.5
	}
	if dsCount > 100 {
		w["xSkirmish"] *= 0.1
	}

	// Jitter ±5%
	for k := range w {
		w[k] *= 0.95 + r.Float64()*0.10
	}

	// 0.01% усиление в 10× или обнуление
	for k := range w {
		v := r.Float64()
		if v < 0.0001 {
			w[k] *= 10
		} else if v < 0.0002 {
			w[k] = 0
		}
	}

	return w
}

// weightedChoice выбирает ключ из карты весов через weighted random.
// Порядок ключей фиксирован для детерминированности.
var expOutcomeOrder = []string{
	"resource", "asteroid", "ship", "battlefield", "xSkirmish",
	"artefact", "credit", "delay", "fast", "nothing", "lost",
}

func weightedChoice(w map[string]float64, r *rng.R) string {
	var total float64
	for _, k := range expOutcomeOrder {
		if v := w[k]; v > 0 {
			total += v
		}
	}
	if total <= 0 {
		return "nothing"
	}
	pick := r.Float64() * total
	var acc float64
	for _, k := range expOutcomeOrder {
		if v := w[k]; v > 0 {
			acc += v
			if pick < acc {
				return k
			}
		}
	}
	return "nothing"
}

// calcResK вычисляет базовый ресурсный множитель res_k.
func calcResK(r *rng.R, expPower float64, hours int, visitedScale float64) int64 {
	base := math.Max(0.5, (1+math.Pow(float64(hours), 1.1))*expPower/40*visitedScale)
	jitter := 1.0 - 0.05 + r.Float64()*0.10 // ±5%
	resK := float64(500_000+r.IntN(500_000)) * base * jitter * 2
	// 2% шанс ×100, cap 10_000_000 * base
	if r.Float64() < 0.02 {
		resK *= 100
		cap := 10_000_000 * base
		if resK > cap {
			resK = cap
		}
	}
	return int64(math.Ceil(resK))
}

// expResources — ресурсный исход (metal + silicon + hydrogen).
func expResources(ctx context.Context, tx pgx.Tx, r *rng.R,
	expPower float64, hours int, visitedScale float64,
	fleetID string, ships []unitStack, cat *config.Catalog,
	cm, csil, ch int64, tr trFn) map[string]any {

	totalCap := fleetCargoCap(ships, cat)
	free := totalCap - (cm + csil + ch)
	if free <= 0 {
		return map[string]any{"bonus": tr("mission", "expeditionNoCargoSpace", nil)}
	}

	resK := calcResK(r, expPower, hours, visitedScale)
	metal := resK
	silicon := int64(math.Ceil(float64(resK) / 2 * (0.90 + r.Float64()*0.20)))
	hydrogen := int64(math.Ceil(float64(resK) / 3 * (0.90 + r.Float64()*0.20)))

	// B4 (план 21): reward cap ≤ fleet_value × 3. Защита от эксплойта
	// транспортов (много cargo, малая cost). Пропорциональное срезание.
	rewardCap := fleetValueMetalEq(ships, cat) * expeditionRewardCapMult
	total := metal + silicon + hydrogen
	if rewardCap > 0 && total > rewardCap {
		k := float64(rewardCap) / float64(total)
		metal = int64(float64(metal) * k)
		silicon = int64(float64(silicon) * k)
		hydrogen = int64(float64(hydrogen) * k)
		total = metal + silicon + hydrogen
	}

	// Зажимаем свободным cargo пропорционально.
	if total > free && total > 0 {
		k := float64(free) / float64(total)
		metal = int64(float64(metal) * k)
		silicon = int64(float64(silicon) * k)
		hydrogen = int64(float64(hydrogen) * k)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE fleets SET carried_metal=$1, carried_silicon=$2, carried_hydrogen=$3
		WHERE id=$4
	`, cm+metal, csil+silicon, ch+hydrogen, fleetID); err != nil {
		return map[string]any{"error": err.Error()}
	}
	return map[string]any{"metal": metal, "silicon": silicon, "hydrogen": hydrogen}
}

// expAsteroid — ресурсный исход без водорода (metal + silicon).
func expAsteroid(ctx context.Context, tx pgx.Tx, r *rng.R,
	expPower float64, hours int, visitedScale float64,
	fleetID string, ships []unitStack, cat *config.Catalog,
	cm, csil, ch int64, tr trFn) map[string]any {

	totalCap := fleetCargoCap(ships, cat)
	free := totalCap - (cm + csil + ch)
	if free <= 0 {
		return map[string]any{"bonus": tr("mission", "expeditionNoCargoSpace", nil)}
	}

	resK := calcResK(r, expPower, hours, visitedScale)
	metal := resK
	silicon := int64(math.Ceil(float64(resK) / 2 * (0.90 + r.Float64()*0.20)))

	// B4 (план 21): reward cap ≤ fleet_value × 3.
	rewardCap := fleetValueMetalEq(ships, cat) * expeditionRewardCapMult
	total := metal + silicon
	if rewardCap > 0 && total > rewardCap {
		k := float64(rewardCap) / float64(total)
		metal = int64(float64(metal) * k)
		silicon = int64(float64(silicon) * k)
		total = metal + silicon
	}

	if total > free && total > 0 {
		k := float64(free) / float64(total)
		metal = int64(float64(metal) * k)
		silicon = int64(float64(silicon) * k)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE fleets SET carried_metal=$1, carried_silicon=$2
		WHERE id=$3
	`, cm+metal, csil+silicon, fleetID); err != nil {
		return map[string]any{"error": err.Error()}
	}
	return map[string]any{"metal": metal, "silicon": silicon}
}

// expArtefact — вставить случайный артефакт в state=held.
func expArtefact(ctx context.Context, tx pgx.Tx, r *rng.R, userID string,
	cat *config.Catalog, tr trFn) map[string]any {
	if len(cat.Artefacts.Artefacts) == 0 {
		return map[string]any{"message": tr("mission", "expeditionNoArtefacts", nil)}
	}
	artIDs := make([]int, 0, len(cat.Artefacts.Artefacts))
	for _, spec := range cat.Artefacts.Artefacts {
		artIDs = append(artIDs, spec.ID)
	}
	idx := r.IntN(len(artIDs))
	artID := artIDs[idx]
	if _, err := tx.Exec(ctx, `
		INSERT INTO artefacts_user (id, user_id, planet_id, unit_id, state, acquired_at)
		VALUES ($1, $2, NULL, $3, 'held', now())
	`, ids.New(), userID, artID); err != nil {
		return map[string]any{"error": err.Error()}
	}
	return map[string]any{"artefact_id": artID}
}

// expExtraPlanet — создаёт временную планету (expires 12–24ч). План
// 72.1 ч.12: numGalaxies/numSystems задают диапазон случайной генерации
// координат (раньше hardcoded 1..8 / 1..500).
func expExtraPlanet(ctx context.Context, tx pgx.Tx, r *rng.R, userID string, numGalaxies, numSystems int, tr trFn) (map[string]any, error) {
	// План 20 Ф.7 + ADR-0005: лимит = max(computer_tech+1, astro/2+1).
	computerLvl := readComputerLevel(ctx, tx, userID)
	astroLvl := readResearchLevel(ctx, tx, userID, unitAstroTech)
	maxPlanets := computerLvl + 1
	if astroLimit := astroLvl/2 + 1; astroLimit > maxPlanets {
		maxPlanets = astroLimit
	}
	var curPlanets int
	if err := tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM planets WHERE user_id=$1 AND destroyed_at IS NULL AND is_moon=false AND expires_at IS NULL`,
		userID).Scan(&curPlanets); err != nil {
		return nil, fmt.Errorf("expExtraPlanet: count: %w", err)
	}
	if curPlanets >= maxPlanets {
		return map[string]any{
			"message": tr("mission", "expeditionPlanetLimitReached", map[string]string{
				"current": fmt.Sprintf("%d", curPlanets),
				"max":     fmt.Sprintf("%d", maxPlanets),
			}),
		}, nil
	}

	for attempt := 0; attempt < 50; attempt++ {
		g := r.IntN(numGalaxies) + 1
		sys := r.IntN(numSystems) + 1
		pos := r.IntN(13) + 2 // 2..14

		var exists bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM planets
				WHERE galaxy=$1 AND system=$2 AND position=$3 AND is_moon=false AND destroyed_at IS NULL
			)
		`, g, sys, pos).Scan(&exists); err != nil {
			return nil, fmt.Errorf("expExtraPlanet: check slot: %w", err)
		}
		if exists {
			continue
		}

		rCoord := rng.New(coordsSeed(g, sys, pos))
		diameter := positionDiameter(pos, rCoord)
		pType := planetTypeOf(pos, rCoord)
		tempMin, tempMax := positionTemp(pos, rCoord)

		// Срок: 12ч + random до 12ч.
		expiresAt := time.Now().UTC().Add(12*time.Hour + time.Duration(r.IntN(int(12*time.Hour))))

		newID := ids.New()
		if _, err := tx.Exec(ctx, `
			INSERT INTO planets (id, user_id, is_moon, name, galaxy, system, position,
			                     diameter, used_fields, planet_type, temperature_min, temperature_max,
			                     metal, silicon, hydrogen, expires_at)
			VALUES ($1, $2, false, 'Expedition Colony', $3, $4, $5, $6, 0, $7, $8, $9, 0, 0, 0, $10)
		`, newID, userID, g, sys, pos, diameter, pType, tempMin, tempMax, expiresAt); err != nil {
			return nil, fmt.Errorf("expExtraPlanet: insert: %w", err)
		}
		// Планируем event KindExpirePlanet=65 на expires_at, чтобы
		// удалить планету вовремя (альтернатива раз-в-час крону).
		payload := fmt.Sprintf(`{"planet_id":"%s"}`, newID)
		if _, err := tx.Exec(ctx, `
			INSERT INTO events (id, user_id, planet_id, kind, state, fire_at, payload)
			VALUES ($1, $2, $3, 65, 'wait', $4, $5)
		`, ids.New(), userID, newID, expiresAt, payload); err != nil {
			return nil, fmt.Errorf("expExtraPlanet: schedule expire: %w", err)
		}
		return map[string]any{
			"planet_id":  newID,
			"galaxy":     g,
			"system":     sys,
			"position":   pos,
			"expires_at": expiresAt,
		}, nil
	}
	return map[string]any{"message": tr("mission", "expeditionNoPlanetPos", nil)}, nil
}

// expPirates — PvE-битва с флотом пиратов, масштабированным по exp_power.
func expPirates(ctx context.Context, tx pgx.Tx, fleetID, ownerUserID string,
	ships []unitStack, cat *config.Catalog, expPower float64, r *rng.R, tr trFn) (map[string]any, error) {
	atkUnits := stacksToBattleUnits(ships, cat, false)
	if len(atkUnits) == 0 {
		return map[string]any{"message": tr("mission", "expeditionNoShips", nil)}, nil
	}

	pirateCount := calcPirateCount(ships, cat, expPower, r)

	lfSpec := findShipSpec(cat, unitLF)
	pirateSide := battle.Side{
		UserID:   "pirates",
		IsAliens: true, // NPC-сторона: ApplyBattleResult её скипнет.
		Units: []battle.Unit{{
			UnitID:   unitLF,
			Quantity: pirateCount,
			Front:    0,
			Attack:   float64(lfSpec.Attack),
			Shield:   float64(lfSpec.Shield),
			Shell:    float64(lfSpec.Shell),
		}},
	}
	input := battle.Input{
		Seed:      deriveSeed(fleetID),
		Rounds:    6,
		Attackers: []battle.Side{{UserID: ownerUserID, Units: atkUnits}},
		Defenders: []battle.Side{pirateSide},
	}
	report, err := battle.Calculate(input)
	if err != nil {
		return nil, fmt.Errorf("expedition pirates: battle: %w", err)
	}
	if _, err := applyAttackerLosses(ctx, tx, fleetID, ships, report.Attackers[0].Units); err != nil {
		return nil, fmt.Errorf("expedition pirates: losses: %w", err)
	}
	// План 72.1.1: зачислить опыт/потери игроку. battleID="" — у
	// экспедиций нет записи в battle_reports (legacy: assaultid=0
	// для NPC). Idempotency обеспечивается event-loop уровнем.
	if err := battlestats.ApplyBattleResult(ctx, tx, report, ""); err != nil &&
		!errors.Is(err, battlestats.ErrAlreadyApplied) {
		return nil, fmt.Errorf("expedition pirates: apply battle result: %w", err)
	}
	return map[string]any{
		"winner":       report.Winner,
		"rounds":       report.Rounds,
		"pirate_fleet": tr("mission", "expeditionPirates", map[string]string{"count": fmt.Sprintf("%d", pirateCount)}),
	}, nil
}

// expBattlefield — бой с повреждённым флотом противника (shell_percent=0.5).
func expBattlefield(ctx context.Context, tx pgx.Tx, fleetID, ownerUserID string,
	ships []unitStack, cat *config.Catalog, expPower float64, r *rng.R, tr trFn) (map[string]any, error) {
	atkUnits := stacksToBattleUnits(ships, cat, false)
	if len(atkUnits) == 0 {
		return map[string]any{"message": tr("mission", "expeditionNoShips", nil)}, nil
	}

	// Генерируем повреждённый флот из LF + Cruiser пропорционально expPower.
	enemyPower := math.Max(1, expPower*(0.3+r.Float64()*0.5))
	lfCount := int64(math.Ceil(enemyPower * 5))
	if lfCount < 1 {
		lfCount = 1
	}
	lfSpec := findShipSpec(cat, unitLF)
	enemyUnit := battle.Unit{
		UnitID:   unitLF,
		Quantity: lfCount,
		Front:    0,
		Attack:   float64(lfSpec.Attack) * 0.5, // shell_percent=0.5
		Shield:   float64(lfSpec.Shield) * 0.5,
		Shell:    float64(lfSpec.Shell) * 0.5,
	}

	input := battle.Input{
		Seed:      deriveSeed(fleetID),
		Rounds:    6,
		Attackers: []battle.Side{{UserID: ownerUserID, Units: atkUnits}},
		Defenders: []battle.Side{{
			UserID:   "battlefield_enemy",
			IsAliens: true, // NPC: ApplyBattleResult её скипнет.
			Units:    []battle.Unit{enemyUnit},
		}},
	}
	report, err := battle.Calculate(input)
	if err != nil {
		return nil, fmt.Errorf("expedition battlefield: battle: %w", err)
	}
	if _, err := applyAttackerLosses(ctx, tx, fleetID, ships, report.Attackers[0].Units); err != nil {
		return nil, fmt.Errorf("expedition battlefield: losses: %w", err)
	}
	// План 72.1.1: зачислить опыт/потери. battleID="" (см. expPirates).
	if err := battlestats.ApplyBattleResult(ctx, tx, report, ""); err != nil &&
		!errors.Is(err, battlestats.ErrAlreadyApplied) {
		return nil, fmt.Errorf("expedition battlefield: apply battle result: %w", err)
	}

	// При победе игрока — добавляем выживших противников в fleet_ships.
	result := map[string]any{
		"winner": report.Winner,
		"rounds": report.Rounds,
		"enemy":  tr("mission", "expeditionEnemyDamaged", map[string]string{"count": fmt.Sprintf("%d", lfCount)}),
	}
	if report.Winner == "attacker" && len(report.Defenders) > 0 {
		for _, du := range report.Defenders[0].Units {
			if du.QuantityEnd > 0 {
				if _, err := tx.Exec(ctx, `
					INSERT INTO fleet_ships (fleet_id, unit_id, count)
					VALUES ($1, $2, $3)
					ON CONFLICT (fleet_id, unit_id) DO UPDATE SET count = fleet_ships.count + EXCLUDED.count
				`, fleetID, du.UnitID, du.QuantityEnd); err != nil {
					return nil, fmt.Errorf("expedition battlefield: add ships: %w", err)
				}
				result["captured"] = du.QuantityEnd
			}
		}
	}
	return result, nil
}

// expCredit — начислить кредиты игроку.
func expCredit(ctx context.Context, tx pgx.Tx, r *rng.R,
	userID string, expPower, visitedScale float64) (map[string]any, error) {
	// Сумма покупок кредитов за последние 3 дня (таблица может отсутствовать).
	var buyCredit float64
	_ = tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM credit_purchases
		WHERE user_id=$1 AND created_at > now() - interval '3 days'
	`, userID).Scan(&buyCredit) // ошибка = buyCredit останется 0

	lo := 10 + math.Min(100, buyCredit/10) + expPower/2
	hi := 29 + math.Min(300, buyCredit) + expPower*2
	if hi <= lo {
		hi = lo + 1
	}
	raw := (lo + r.Float64()*(hi-lo)) * visitedScale * 0.7
	// Округление до десятков + случайные единицы 5..9.
	credit := int64(math.Ceil(raw/10)*10) + int64(5+r.IntN(5))
	if credit < 5 {
		credit = 5
	}

	if _, err := tx.Exec(ctx,
		`UPDATE users SET credits = credits + $1 WHERE id=$2`,
		credit, userID); err != nil {
		return nil, fmt.Errorf("expedition credit: update: %w", err)
	}
	return map[string]any{"credits": credit}, nil
}

// expDelay — сдвинуть fire_at события возврата вправо (10–30% от времени полёта).
func expDelay(ctx context.Context, tx pgx.Tx, r *rng.R,
	returnEventID string, flightSeconds int64, tr trFn) (map[string]any, error) {
	if returnEventID == "" {
		return map[string]any{"message": tr("mission", "expeditionDelayNoReturnEvent", nil)}, nil
	}
	delta := int64(math.Round(float64(flightSeconds) * (0.10 + r.Float64()*0.20)))
	if _, err := tx.Exec(ctx,
		`UPDATE events SET fire_at = fire_at + ($1 * interval '1 second') WHERE id=$2`,
		delta, returnEventID); err != nil {
		return nil, fmt.Errorf("expDelay: update: %w", err)
	}
	return map[string]any{"delay_seconds": delta}, nil
}

// expFast — сдвинуть fire_at события возврата влево (10–60% от времени полёта).
func expFast(ctx context.Context, tx pgx.Tx, r *rng.R,
	returnEventID string, flightSeconds int64, tr trFn) (map[string]any, error) {
	if returnEventID == "" {
		return map[string]any{"message": tr("mission", "expeditionFastNoReturnEvent", nil)}, nil
	}
	delta := int64(math.Round(float64(flightSeconds) * (0.10 + r.Float64()*0.50)))
	if _, err := tx.Exec(ctx,
		`UPDATE events SET fire_at = fire_at - ($1 * interval '1 second') WHERE id=$2`,
		delta, returnEventID); err != nil {
		return nil, fmt.Errorf("expFast: update: %w", err)
	}
	return map[string]any{"fast_seconds": delta}, nil
}

// expShip — добавить корабли в fleet_ships флота игрока.
func expShip(ctx context.Context, tx pgx.Tx, r *rng.R,
	fleetID string, expPower float64) (map[string]any, error) {
	count := int64(1 + r.IntN(int(math.Max(1, math.Ceil(expPower/3)))))
	// Случайно: recycler или light_fighter.
	unitID := unitLF
	if r.IntN(2) == 0 {
		unitID = unitRecycler
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO fleet_ships (fleet_id, unit_id, count)
		VALUES ($1, $2, $3)
		ON CONFLICT (fleet_id, unit_id) DO UPDATE SET count = fleet_ships.count + EXCLUDED.count
	`, fleetID, unitID, count); err != nil {
		return nil, fmt.Errorf("expShip: insert: %w", err)
	}
	return map[string]any{"unit_id": unitID, "count": count}, nil
}

// expLoss — полное уничтожение флота (legacy Expedition::expeditionLost,
// sendBack=false). Удаляет все fleet_ships и полагается на то, что
// вызывающий выставит fleets.state='done' — тогда return-event, когда
// сработает, просто молча завершится (ReturnHandler имеет ранний
// выход на state='done', см. events.go:223).
//
// Мы сознательно НЕ удаляем return-event, чтобы в таблице events
// осталась полная история жизни флота: запись с kind=20, state
// которого пройдёт путь wait→ok. Это нужно для будущего аудита
// проблемных ситуаций.
//
// До 2026-04-24 здесь снималось 5..20% — частичные потери. Это было
// упрощением (см. simplifications.md). План 17 B1 портировал legacy.
func expLoss(ctx context.Context, tx pgx.Tx, r *rng.R, fleetID string,
	ships []unitStack) (map[string]any, error) {
	_ = r // rng больше не нужен (раньше использовался для frac)
	losses := map[int]int64{}
	for _, sh := range ships {
		losses[sh.UnitID] = sh.Count
	}
	if _, err := tx.Exec(ctx,
		`DELETE FROM fleet_ships WHERE fleet_id=$1`, fleetID); err != nil {
		return nil, fmt.Errorf("expLoss: delete fleet_ships: %w", err)
	}
	return map[string]any{"lost": losses, "total_destroyed": true}, nil
}

// --- helpers ---

// readResearchLevel читает уровень технологии из таблицы research.
func readResearchLevel(ctx context.Context, tx pgx.Tx, userID string, unitID int) int {
	var lvl int
	_ = tx.QueryRow(ctx,
		`SELECT COALESCE(level, 0) FROM research WHERE user_id=$1 AND unit_id=$2`,
		userID, unitID).Scan(&lvl)
	return lvl
}

// fleetCargoCap суммирует cargo capacity всех кораблей флота.
func fleetCargoCap(ships []unitStack, cat *config.Catalog) int64 {
	var cap int64
	for _, s := range ships {
		for _, spec := range cat.Ships.Ships {
			if spec.ID == s.UnitID {
				cap += spec.Cargo * s.Count
				break
			}
		}
	}
	return cap
}

// fleetValueMetalEq — суммарная стоимость флота в metal-eq (M+Si+H).
// Используется как нижняя планка награды экспедиции (план 21 B4).
func fleetValueMetalEq(ships []unitStack, cat *config.Catalog) int64 {
	var sum int64
	for _, s := range ships {
		spec := findShipSpec(cat, s.UnitID)
		sum += (spec.Cost.Metal + spec.Cost.Silicon + spec.Cost.Hydrogen) * s.Count
	}
	return sum
}

// findShipSpec возвращает spec корабля по unit_id, или пустой spec.
func findShipSpec(cat *config.Catalog, unitID int) config.ShipSpec {
	for _, spec := range cat.Ships.Ships {
		if spec.ID == unitID {
			return spec
		}
	}
	return config.ShipSpec{}
}

// calcPirateCount вычисляет количество LF в пиратском флоте.
// Масштаб по суммарной атаке флота игрока; min 3, max 500.
func calcPirateCount(ships []unitStack, cat *config.Catalog, expPower float64, r *rng.R) int64 {
	lfAttack := float64(50) // базовая атака LF (fallback)
	var totalAttack float64
	if cat != nil {
		lfSpec := findShipSpec(cat, unitLF)
		if lfSpec.Attack > 0 {
			lfAttack = float64(lfSpec.Attack)
		}
		for _, s := range ships {
			for _, spec := range cat.Ships.Ships {
				if spec.ID == s.UnitID {
					totalAttack += float64(spec.Attack) * float64(s.Count)
					break
				}
			}
		}
	}
	// expPower добавляет бонус к силе пиратов.
	pirateAttack := totalAttack*(0.05+r.Float64()*0.10) + expPower*lfAttack
	count := int64(pirateAttack / lfAttack)
	if count < 3 {
		count = 3
	}
	if count > 500 {
		count = 500
	}
	return count
}
