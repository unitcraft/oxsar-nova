package aiadvisor

import (
	"fmt"
	"sort"

	"oxsar/game-nova/internal/config"
	"oxsar/game-nova/internal/economy"
	"oxsar/game-nova/pkg/ids"
)

// scoringInputs — DI для scoreCandidates (тесты дают ин-мемори каталоги).
//
// PointsK — коэффициенты очков (Building/Research/Unit). Берутся из
// cfg.Game.Points; нужны для расчёта delta_score кандидата.
type scoringInputs struct {
	Catalog *config.Catalog
	PointsK config.PointsCoefficients

	// BuildSecondsByPlanet — время постройки следующего уровня каждого
	// здания на каждой планете (учитывает robotic/nano factory). Заполняется
	// в Compute через building.Service.BuildSecondsMap, чтобы scoring
	// оставался pure-функцией без БД.
	BuildSecondsByPlanet map[string]map[int]int
}

// scoreCandidates — pure-функция: snapshot + стратегия → отсортированный
// топ-3 ScoredAction (по убыванию Score).
//
// Поддерживаемые категории по фазам:
//   Ф.1в: "building" (только Economy в этой фазе);
//   Ф.1г: "research" + Economy для других стратегий;
//   Ф.2:  миссии (атака/шпионаж/транспорт/экспедиция);
//   Ф.3:  ACS, биржа, профессия;
//   Ф.5:  StrategyAuto — выбирает целевую функцию через detectPhase.
func scoreCandidates(snap PlayerSnapshot, strategy Strategy, inputs scoringInputs) []Recommendation {
	if strategy == StrategyAuto {
		strategy = detectPhase(snap)
	}

	var recs []Recommendation
	if inputs.Catalog != nil {
		recs = append(recs, scoreBuildings(snap, strategy, inputs)...)
	}

	sort.SliceStable(recs, func(i, j int) bool {
		return recs[i].Score > recs[j].Score
	})

	if len(recs) > 3 {
		recs = recs[:3]
	}
	return recs
}

// scoreBuildings перебирает все здания всех планет и оценивает кандидатов.
//
// Кандидат отсеивается если:
//   - очередь планеты занята (BuildingQueueBusy);
//   - здание moon-only, а планета не луна (или наоборот);
//   - текущий уровень >= MaxLevel (если задан);
//   - на следующий уровень не хватает ресурсов (по прогнозу 1ч);
//   - использованы все поля (UsedFields >= MaxFields) — поле освобождает
//     только демонтаж, который scoring не предлагает.
//
// Score выгоды зависит от стратегии. На уровне Ф.1в реализован Economy:
//   score = (delta_production_score) + (delta_score_points / build_seconds)
// где delta_production_score — прирост ресурсов в секунду * coefficient;
// delta_score_points — прирост очков игрока за конкретный уровень.
//
// Стратегии Military/Defense/Expansion отдают очень низкий score любому
// чисто экономическому зданию; они получат свой scoring в Ф.1г-Ф.5
// (приоритет shipyard/defense зданий, лабораторий и т.д.).
func scoreBuildings(snap PlayerSnapshot, strategy Strategy, inputs scoringInputs) []Recommendation {
	var out []Recommendation
	cat := inputs.Catalog

	for _, ps := range snap.Planets {
		if ps.BuildingQueueBusy {
			continue
		}
		if ps.UsedFields >= ps.MaxFields {
			continue
		}

		buildSeconds := inputs.BuildSecondsByPlanet[ps.ID]

		for key, spec := range cat.Buildings.Buildings {
			if spec.MoonOnly && !ps.IsMoon {
				continue
			}
			if !spec.MoonOnly && ps.IsMoon {
				// На луне строим только moon-only здания.
				continue
			}
			curLvl := ps.Buildings[spec.ID]
			nextLvl := curLvl + 1
			if spec.MaxLevel > 0 && nextLvl > spec.MaxLevel {
				continue
			}

			// Стоимость следующего уровня (формула совпадает с
			// building.Service.BuildCostsMap).
			cost := economy.CostForLevel(economy.Cost{
				Metal:    spec.CostBase.Metal,
				Silicon:  spec.CostBase.Silicon,
				Hydrogen: spec.CostBase.Hydrogen,
			}, spec.CostFactor, nextLvl)
			// Affordability — против прогноза на 1 час, чтобы автопилот
			// предлагал тот ход, к которому игрок сможет приступить
			// сразу или через короткое ожидание (ресурсы добываются).
			if !affordable(cost, ps.ResourcesIn1h) {
				continue
			}

			seconds, ok := buildSeconds[spec.ID]
			if !ok || seconds <= 0 {
				continue // нет данных по времени постройки — пропускаем
			}

			deltaPoints := inputs.PointsK.Building * float64(cost.Metal+cost.Silicon+cost.Hydrogen)
			deltaProduction := productionDelta(spec, curLvl, nextLvl)

			score := buildingScore(strategy, deltaPoints, deltaProduction, float64(seconds))
			if score <= 0 {
				continue
			}

			out = append(out, Recommendation{
				ID:         ids.New(),
				Category:   "building",
				ActionType: "build_" + key,
				PlanetID:   ps.ID,
				UnitID:     spec.ID,
				Params: map[string]any{
					"unit_key":     key,
					"target_level": nextLvl,
				},
				Score:       score,
				Description: buildingDescription(key, ps.Name, nextLvl),
				Benefit:     buildingBenefit(deltaPoints, deltaProduction, seconds),
			})
		}
	}
	return out
}

// affordable возвращает true если ресурсов достаточно для оплаты cost.
func affordable(cost economy.Cost, have Resources) bool {
	return have.Metal >= float64(cost.Metal) &&
		have.Silicon >= float64(cost.Silicon) &&
		have.Hydrogen >= float64(cost.Hydrogen)
}

// productionDelta возвращает суммарный прирост добычи в секунду при
// переходе с curLvl на nextLvl, в виртуальных «единицах ресурса/сек».
//
// Для зданий без BaseRatePerHour (включая solar_plant — он даёт энергию,
// не ресурс) возвращает 0.
//
// Эта функция не учитывает factor планеты, температуру, исследования и
// артефакты — это оценка, а не точная цифра. В реальной игре
// производство будет ниже за счёт energy_ratio < 1, но автопилот
// сравнивает кандидатов между собой, а не предсказывает абсолютный объём.
func productionDelta(spec config.BuildingSpec, curLvl, nextLvl int) float64 {
	if spec.BaseRatePerHour == nil {
		return 0
	}
	cur := economy.ProductionPerHour(*spec.BaseRatePerHour, curLvl, 1.0)
	next := economy.ProductionPerHour(*spec.BaseRatePerHour, nextLvl, 1.0)
	delta := next - cur
	if delta < 0 {
		return 0
	}
	return delta / 3600.0 // /час → /сек
}

// buildingScore — формула выгоды в зависимости от стратегии.
//
// Economy: prioritise приращение производства (за сек). Score-points
// идут как небольшая добавка для tiebreaker'а между не-ресурсными
// зданиями (роботы/верфь/лаба — без production-эффекта, но дают очки).
//
// Прочие стратегии в Ф.1в возвращают только score-points (низкий score
// для ресурсных шахт). В Ф.1г-Ф.5 они получат собственные веса.
func buildingScore(strategy Strategy, deltaPoints, deltaProductionPerSec, buildSeconds float64) float64 {
	if buildSeconds <= 0 {
		return 0
	}
	pointsPerSec := deltaPoints / buildSeconds
	switch strategy {
	case StrategyEconomy:
		// Production доминирует; pointsPerSec — добавка ~10% веса.
		return deltaProductionPerSec + pointsPerSec*0.1
	default:
		// Для Military/Defense/Expansion ресурсное здание не приоритетно;
		// возвращаем чисто очки в секунду — это даст плохой ранг по
		// сравнению с shipyard/lab/defense, которые войдут в Ф.1г-Ф.5.
		return pointsPerSec
	}
}

// buildingDescription — человекочитаемое описание для UI (Description).
//
// Берётся ключ юнита и подставляется в шаблон. i18n будет добавлен в
// Ф.4 (фронт переведёт по ActionType + Params).
func buildingDescription(key, planetName string, nextLevel int) string {
	return fmt.Sprintf("Построить %s ур.%d на %q", key, nextLevel, planetName)
}

// buildingBenefit — строка с ожидаемой выгодой, для UI.
func buildingBenefit(deltaPoints, deltaProductionPerSec float64, seconds int) string {
	dur := formatDuration(seconds)
	if deltaProductionPerSec > 0 {
		// Производственное здание — основная выгода в добыче ресурсов.
		return fmt.Sprintf("+%.1f ед/сек (≈+%.1f/час), %s",
			deltaProductionPerSec, deltaProductionPerSec*3600, dur)
	}
	return fmt.Sprintf("+%.1f очков, %s", deltaPoints, dur)
}

// formatDuration — компактный «4ч 20мин» / «35мин» / «50сек».
func formatDuration(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%dс", seconds)
	}
	if seconds < 3600 {
		return fmt.Sprintf("%dмин", seconds/60)
	}
	h := seconds / 3600
	m := (seconds % 3600) / 60
	if m == 0 {
		return fmt.Sprintf("%dч", h)
	}
	return fmt.Sprintf("%dч %dмин", h, m)
}

// detectPhase — эвристика фазы развития для StrategyAuto.
//
// Реализация в Ф.5 — пока возвращает Economy как безопасный дефолт
// (большинство игроков на ранней стадии экономика-ориентированы).
func detectPhase(_ PlayerSnapshot) Strategy {
	return StrategyEconomy
}
