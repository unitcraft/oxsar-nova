package aiadvisor

import (
	"fmt"
	"math"
	"sort"

	"oxsar/game-nova/internal/config"
	"oxsar/game-nova/internal/economy"
	"oxsar/game-nova/internal/i18n"
	"oxsar/game-nova/pkg/ids"
)

// scoringInputs — DI для scoreCandidates (тесты дают ин-мемори каталоги).
//
// PointsK — коэффициенты очков (Building/Research/Unit). Берутся из
// cfg.Game.Points; нужны для расчёта delta_score кандидата.
//
// GameSpeed / ResearchSpeed — множители из конфига вселенной.
// Используются в формуле длительности research (1:1 с research.Service).
type scoringInputs struct {
	Catalog *config.Catalog
	PointsK config.PointsCoefficients

	// BuildSecondsByPlanet — время постройки следующего уровня каждого
	// здания на каждой планете (учитывает robotic/nano factory). Заполняется
	// в Compute через building.Service.BuildSecondsMap, чтобы scoring
	// оставался pure-функцией без БД.
	BuildSecondsByPlanet map[string]map[int]int

	// GameSpeed — game-wide speed factor (cfg.Game.Speed).
	GameSpeed float64
	// ResearchSpeed — research-specific factor (cfg.Game.ResearchSpeedFactor).
	ResearchSpeed float64

	// ResearchLabKey / MoonLabKey — ключи зданий-лаборатории.
	// В стандартной игре "research_lab" / "moon_lab"; вынесены параметрами,
	// чтобы тесты могли переиспользовать упрощённый каталог.
	ResearchLabKey string
	MoonLabKey     string

	// Bundle — i18n для перевода названий зданий/исследований в
	// Description рекомендаций. Если nil — fallback на raw-key.
	Bundle *i18n.Bundle
	// Language — язык игрока ('ru'|'en'); выбирается из users.language.
	Language string
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
		recs = append(recs, scoreResearch(snap, strategy, inputs)...)
		recs = append(recs, scoreMissions(snap, strategy, inputs)...)
		recs = append(recs, scoreProfession(snap, strategy)...)
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

			// Prereq: building requires (research lab levels, technologies).
			// Без этой проверки советник предлагал ходы, которые
			// fleet/building.Enqueue потом отбивал «requirement not met».
			if !meetsRequirements(key, ps.Buildings, snap.Research, cat) {
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
				Description: buildingDescription(key, ps.Name, nextLvl, inputs),
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

// meetsRequirements проверяет что у игрока выполнены ВСЕ предусловия
// для строительства/исследования юнита targetKey на конкретной планете.
//
// Источник правды: configs/requirements.yml (Catalog.Requirements).
// Семантика 1:1 с requirements.Checker.Check (см. internal/requirements/):
//   - kind="building": уровень здания на planetBuildings >= req.Level.
//   - kind="research": уровень исследования у игрока >= req.Level.
//
// Если planetBuildings == nil — building-требования НЕ проверяются
// (используется в research-scoring, где building-prereq проверяется
// per-planet отдельно через pickResearchPlanet — там нужна планета
// с research_lab нужного уровня).
//
// Если в каталоге для targetKey нет записи — требований нет (true).
// Если каталог nil — true (тесты могут не передавать requirements).
func meetsRequirements(
	targetKey string,
	planetBuildings map[int]int,
	research map[int]int,
	catalog *config.Catalog,
) bool {
	if catalog == nil {
		return true
	}
	reqs, ok := catalog.Requirements.Requirements[targetKey]
	if !ok || len(reqs) == 0 {
		return true
	}
	for _, r := range reqs {
		switch r.Kind {
		case "building":
			if planetBuildings == nil {
				continue // building-prereq не проверяется на этом уровне
			}
			spec, ok := catalog.Buildings.Buildings[r.Key]
			if !ok {
				continue // неизвестное здание — не блокируем
			}
			if planetBuildings[spec.ID] < r.Level {
				return false
			}
		case "research":
			spec, ok := catalog.Research.Research[r.Key]
			if !ok {
				continue
			}
			if research[spec.ID] < r.Level {
				return false
			}
		}
	}
	return true
}

// scoreTimeExponent — степень в формуле Score = delta_points / time^exp.
//
// 1.0 (linear) переоценивает быстрые мелочи; 0.5 (sqrt) недо-оценивает
// quick wins. Промежуточные значения — компромисс.
//
// Per-категория (см. discussion 2026-05-06):
//   - building: 0.7 — здания дают compound (+10%/уровень добычи),
//     стоит ценить структурные long-term апгрейды; но не sqrt чтобы
//     metal_mine ур.2 (instant, +1.1× добычи) всё ещё попадал в топ.
//   - research: 0.6 — research-цены растут геометрически, длинные
//     апгрейды (hyperspace_tech ур.10+) дают мощные unlock'и,
//     наказывать за 30 секунд глупо.
//   - mission (атака/spy/expedition): 1.0 — миссии короткосрочны,
//     быстрая ARR важнее долгосрочной выгоды; точное время рассчитать
//     для миссии трудно, поэтому фактически time не учитываем
//     (вес = 1 при time=1сек, что совпадает с прошлым поведением).
const (
	scoreTimeExpBuilding = 0.7
	scoreTimeExpResearch = 0.6
	scoreTimeExpMission  = 1.0
)

// timeDiscounted возвращает delta_points / time^exp с защитой от
// деления на 0 (при time<=0 возвращает 0 — кандидат отсеивается).
func timeDiscounted(deltaPoints, seconds, exp float64) float64 {
	if seconds <= 0 {
		return 0
	}
	if exp == 1.0 {
		return deltaPoints / seconds // быстрый путь без math.Pow
	}
	return deltaPoints / math.Pow(seconds, exp)
}


// researchStrategyWeights — стратегический вес группы технологий.
//
// Веса используют шкалу: 1.0 = «целевая для этой стратегии», 0.1 =
// «нейтральная», 0.01 = «вне профиля». Большой разрыв (×100) между
// «целевой» и «вне профиля» нужен потому что научные исследования
// сильно различаются по стоимости (astrophysics дороже weapons на
// порядок), а простое умножение pointsPerSec на 0.2 vs 1.0 не компенсирует
// эту разницу — дорогая технология выиграла бы по абсолютному score.
//
// Дефолт (если ключа нет): 0.1 (нейтральный — research даёт очки, но
// без бонуса под выбранную стратегию).
var researchStrategyWeights = map[string]map[Strategy]float64{
	// Боевые: Military профиль, Defense вторичный.
	"weapons_tech": {StrategyMilitary: 1.0, StrategyDefense: 0.5, StrategyEconomy: 0.01, StrategyExpansion: 0.01},
	"shield_tech":  {StrategyMilitary: 0.7, StrategyDefense: 1.0, StrategyEconomy: 0.01, StrategyExpansion: 0.01},
	"armor_tech":   {StrategyMilitary: 0.7, StrategyDefense: 1.0, StrategyEconomy: 0.01, StrategyExpansion: 0.01},
	// Двигатели — Expansion профиль + Military вторичный.
	"engine_combust": {StrategyMilitary: 0.4, StrategyExpansion: 0.8, StrategyEconomy: 0.05, StrategyDefense: 0.05},
	"engine_impulse": {StrategyMilitary: 0.4, StrategyExpansion: 0.8, StrategyEconomy: 0.05, StrategyDefense: 0.05},
	"engine_hyper":   {StrategyMilitary: 0.5, StrategyExpansion: 0.9, StrategyEconomy: 0.05, StrategyDefense: 0.05},
	// Энергооружие — Military.
	"laser_tech":  {StrategyMilitary: 0.6, StrategyDefense: 0.3, StrategyEconomy: 0.01, StrategyExpansion: 0.01},
	"ion_tech":    {StrategyMilitary: 0.6, StrategyDefense: 0.3, StrategyEconomy: 0.01, StrategyExpansion: 0.01},
	"plasma_tech": {StrategyMilitary: 0.7, StrategyDefense: 0.3, StrategyEconomy: 0.01, StrategyExpansion: 0.01},
	// Шпионаж и компьютер — универсальные.
	"spy_tech":      {StrategyMilitary: 0.3, StrategyDefense: 0.3, StrategyEconomy: 0.2, StrategyExpansion: 0.2},
	"computer_tech": {StrategyMilitary: 0.3, StrategyDefense: 0.2, StrategyEconomy: 0.2, StrategyExpansion: 0.3},
	// Экспансия — astrophysics — единственная путь к новым планетам.
	"astrophysics":                 {StrategyExpansion: 1.0, StrategyMilitary: 0.02, StrategyEconomy: 0.05, StrategyDefense: 0.01},
	"intergalactic_research_network": {StrategyEconomy: 0.4, StrategyMilitary: 0.1, StrategyDefense: 0.1, StrategyExpansion: 0.1},
	// Гравитация — Military (Deathstar/RIP).
	"gravitation_tech": {StrategyMilitary: 0.4, StrategyDefense: 0.1, StrategyEconomy: 0.01, StrategyExpansion: 0.05},
}

// researchWeight возвращает множитель score для технологии при стратегии.
// Дефолт 0.1 — нейтральная неизвестная технология.
func researchWeight(techKey string, strategy Strategy) float64 {
	if m, ok := researchStrategyWeights[techKey]; ok {
		if w, ok2 := m[strategy]; ok2 {
			return w
		}
	}
	return 0.1
}

// scoreResearch перебирает все технологии и оценивает кандидатов.
//
// Кандидат отсеивается если:
//   - Очередь исследований занята (один research на игрока, флаг
//     ResearchQueueBusy у любой планеты — общий);
//   - Текущий уровень >= MaxResearchLevel (40);
//   - Нет планеты с research_lab >= 1 и достаточными ресурсами на
//     прогноз 1ч.
//
// Для подбора планеты-источника берётся та, у которой:
//   1. Есть лаб (research_lab или moon_lab если планета-луна);
//   2. ResourcesIn1h покрывает cost(nextLevel);
//   3. Из подходящих — с максимальным effectiveLab (=labLevel,
//      без учёта IGR — оценка снизу).
//
// Score = pointsPerSec * researchWeight(techKey, strategy).
func scoreResearch(snap PlayerSnapshot, strategy Strategy, inputs scoringInputs) []Recommendation {
	if inputs.Catalog == nil {
		return nil
	}
	// Один research на игрока. Если хоть на одной планете флаг busy — пропускаем.
	if anyResearchBusy(snap) {
		return nil
	}

	gameSpeed := inputs.GameSpeed
	if gameSpeed <= 0 {
		gameSpeed = 1
	}
	researchSpeed := inputs.ResearchSpeed
	if researchSpeed <= 0 {
		researchSpeed = 1
	}
	labKey := inputs.ResearchLabKey
	if labKey == "" {
		labKey = "research_lab"
	}
	moonLabKey := inputs.MoonLabKey
	if moonLabKey == "" {
		moonLabKey = "moon_lab"
	}

	var labSpec, moonLabSpec config.BuildingSpec
	if s, ok := inputs.Catalog.Buildings.Buildings[labKey]; ok {
		labSpec = s
	}
	if s, ok := inputs.Catalog.Buildings.Buildings[moonLabKey]; ok {
		moonLabSpec = s
	}

	var out []Recommendation
	for techKey, spec := range inputs.Catalog.Research.Research {
		curLvl := snap.Research[spec.ID]
		nextLvl := curLvl + 1
		// Ограничение легаси: MAX_RESEARCH_LEVEL=40.
		if nextLvl > 40 {
			continue
		}

		// Prereq: research-prerequisites (laser_tech, energy_tech, ...)
		// и building-prerequisites (research_lab уровень).
		// Building-prereq проверяется per-planet в pickResearchPlanet ниже,
		// но research-prereq глобальный — отсекаем сразу.
		if !meetsRequirements(techKey, nil, snap.Research, inputs.Catalog) {
			continue
		}

		cost := economy.CostForLevel(economy.Cost{
			Metal:    spec.CostBase.Metal,
			Silicon:  spec.CostBase.Silicon,
			Hydrogen: spec.CostBase.Hydrogen,
		}, spec.CostFactor, nextLvl)

		planetID, effectiveLab := pickResearchPlanet(snap, cost, labSpec, moonLabSpec)
		if planetID == "" {
			continue // нет планеты с лабом + ресурсами
		}

		// Длительность по 1:1 формуле research.Service.
		seconds := researchDuration(cost, effectiveLab, gameSpeed, researchSpeed)
		if seconds <= 0 {
			continue
		}

		deltaPoints := inputs.PointsK.Research * float64(cost.Metal+cost.Silicon+cost.Hydrogen)
		// scoreTimeExpResearch=0.6: research-цены растут геометрически,
		// длинные апгрейды (hyperspace_tech ур.10+) дают мощные unlock'и.
		// Менее агрессивная экспонента чем у building (0.6 < 0.7) —
		// research более «инвестиционный» в природе.
		pointsDiscounted := timeDiscounted(deltaPoints, float64(seconds), scoreTimeExpResearch)
		score := pointsDiscounted * researchWeight(techKey, strategy)
		if score <= 0 {
			continue
		}

		planetName := ""
		for _, ps := range snap.Planets {
			if ps.ID == planetID {
				planetName = ps.Name
				break
			}
		}

		out = append(out, Recommendation{
			ID:         ids.New(),
			Category:   "research",
			ActionType: "research_" + techKey,
			PlanetID:   planetID,
			UnitID:     spec.ID,
			Params: map[string]any{
				"unit_key":     techKey,
				"target_level": nextLvl,
			},
			Score:       score,
			Description: researchDescription(techKey, planetName, nextLvl, inputs),
			Benefit:     researchBenefit(deltaPoints, seconds),
		})
	}
	return out
}

// anyResearchBusy — флаг ResearchQueueBusy одинаков на всех планетах
// snapshot'а (research один на игрока), но проверяем все на случай
// рассогласования.
func anyResearchBusy(snap PlayerSnapshot) bool {
	for _, ps := range snap.Planets {
		if ps.ResearchQueueBusy {
			return true
		}
	}
	return false
}

// pickResearchPlanet выбирает планету-источник для исследования.
// Возвращает ("", 0) если ни одна не подходит.
func pickResearchPlanet(snap PlayerSnapshot, cost economy.Cost, labSpec, moonLabSpec config.BuildingSpec) (string, int) {
	bestID := ""
	bestLab := 0
	for _, ps := range snap.Planets {
		var lvl int
		if ps.IsMoon {
			lvl = ps.Buildings[moonLabSpec.ID]
		} else {
			lvl = ps.Buildings[labSpec.ID]
		}
		if lvl < 1 {
			continue
		}
		if !affordable(cost, ps.ResourcesIn1h) {
			continue
		}
		if lvl > bestLab {
			bestLab = lvl
			bestID = ps.ID
		}
	}
	return bestID, bestLab
}

// researchDuration — формула 1:1 c research.Service.Enqueue:
//   t = (m+s) / (1000 * (1 + effectiveLab)) сек, /gameSpeed /researchSpeed,
//   floor=1.
func researchDuration(cost economy.Cost, effectiveLab int, gameSpeed, researchSpeed float64) int {
	resSum := float64(cost.Metal + cost.Silicon)
	raw := resSum / (1000.0 * float64(1+effectiveLab))
	if gameSpeed > 0 {
		raw /= gameSpeed
	}
	if researchSpeed > 0 {
		raw /= researchSpeed
	}
	if raw < 1 {
		raw = 1
	}
	return int(raw + 0.5)
}

func researchDescription(key, planetName string, nextLevel int, inputs scoringInputs) string {
	label := translateUnitKey(inputs, key)
	if planetName == "" {
		return fmt.Sprintf("Исследовать %s ур.%d", label, nextLevel)
	}
	return fmt.Sprintf("Исследовать %s ур.%d (с планеты %q)", label, nextLevel, planetName)
}

func researchBenefit(deltaPoints float64, seconds int) string {
	return fmt.Sprintf("+%.1f очков, %s", deltaPoints, formatDuration(seconds))
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
	// scoreTimeExpBuilding=0.7: умеренное сглаживание времени —
	// длинные апгрейды (Solar Plant ур.20) не теряются под мелочами,
	// но instant-апгрейды (robotic_factory ур.1) всё ещё в топе.
	// Прежняя формула pointsPerSec (linear, exp=1.0) слишком сильно
	// штрафовала структурные ходы.
	pointsDiscounted := timeDiscounted(deltaPoints, buildSeconds, scoreTimeExpBuilding)
	switch strategy {
	case StrategyEconomy:
		// Production доминирует; pointsDiscounted — добавка ~10% веса.
		return deltaProductionPerSec + pointsDiscounted*0.1
	default:
		// Для Military/Defense/Expansion ресурсное здание не приоритетно;
		// возвращаем чисто discounted-очки — это даст плохой ранг по
		// сравнению с shipyard/lab/defense, которые войдут в Ф.1г-Ф.5.
		return pointsDiscounted
	}
}

// buildingDescription — человекочитаемое описание для UI (Description).
//
// Если bundle задан — переводит ключ юнита через i18n (info.<camelKey>);
// иначе fallback на raw-key (поведение Ф.1в без i18n).
func buildingDescription(key, planetName string, nextLevel int, inputs scoringInputs) string {
	label := translateUnitKey(inputs, key)
	return fmt.Sprintf("Построить %s ур.%d на %q", label, nextLevel, planetName)
}

// translateUnitKey ищет локализованное название юнита (здания или
// исследования) по ключу. Используется группа i18n "info" — там лежат
// названия (info.<camelKey>) и описания (info.<camelKey>Desc, ...Full).
//
// Если bundle nil или ключ не найден — возвращает raw-key (читабельно
// в логах и при дев-сборках без i18n).
func translateUnitKey(inputs scoringInputs, snakeKey string) string {
	if inputs.Bundle == nil {
		return snakeKey
	}
	camel := snakeToCamel(snakeKey)
	lang := i18n.Lang(inputs.Language)
	if lang == "" {
		lang = i18n.LangRu
	}
	if inputs.Bundle.Has(lang, "info", camel) {
		return inputs.Bundle.Tr(lang, "info", camel, nil)
	}
	return snakeKey
}

// snakeToCamel конвертирует snake_case → camelCase для i18n-ключей.
//   metal_mine → metalMine
//   hyperspace_tech → hyperspaceTech
//   solar_plant → solarPlant
func snakeToCamel(s string) string {
	if s == "" {
		return s
	}
	out := make([]byte, 0, len(s))
	upperNext := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '_' {
			upperNext = true
			continue
		}
		if upperNext && c >= 'a' && c <= 'z' {
			c -= 'a' - 'A'
		}
		upperNext = false
		out = append(out, c)
	}
	return string(out)
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
// Возвращает целевую стратегию исходя из состояния игрока:
//   1. Меньше 3 планет → Expansion (пока есть куда расти, расти).
//   2. Слабое производство (< 100 metal/sec суммарно) → Economy.
//   3. Текущая ценность военного флота < 200_000 metal-eq → Military
//      (нужно строить флот для защиты/атак).
//   4. Иначе → Defense (наращивать оборонные установки и щиты).
//
// Пороговые значения подобраны на основе типичных F2P-кривых: к 3-й
// планете игрок уже имеет ~50-100 metal/sec, после неё фокус смещается
// в сторону армии; defense становится приоритетом, когда флот собран,
// и нужно защитить ресурсы от соседей.
func detectPhase(snap PlayerSnapshot) Strategy {
	if len(snap.Planets) < 3 {
		return StrategyExpansion
	}

	var totalMetalRate float64
	for _, ps := range snap.Planets {
		totalMetalRate += ps.Rates.Metal
	}
	if totalMetalRate < 100 { // 100 metal/sec суммарно ≈ Earth с metal_mine ур.10
		return StrategyEconomy
	}

	// Ценность флота на всех планетах в metal-eq (по cost).
	var fleetValue int64
	for _, ps := range snap.Planets {
		// Используем те же "цены" что в metalEqShipValue, но без catalog
		// (catalog здесь не доступен — detectPhase pure-функция от snapshot).
		// Считаем по cargo*0.5 как proxy: грубо, но монотонно растёт с флотом.
		for unitID, count := range ps.Ships {
			fleetValue += approxShipValue(unitID) * count
		}
	}
	if fleetValue < 200_000 {
		return StrategyMilitary
	}

	return StrategyDefense
}

// approxShipValue — грубая оценка ценности корабля без catalog.
//
// Используется только в detectPhase для ranking-эвристики; даёт
// stable-relative порядок без абсолютной точности. Значения подобраны
// под configs/ships.yml: small_transporter=4k, large=12k, light_fighter=4k,
// cruiser=27k, battleship=60k, deathstar=10M и т.д.
// Источник: configs/ships.yml. Значения — суммарная metal-eq cost
// корабля. Исправлено 2026-05-06: были неправильные ID (39 был
// помечен bomber, реально это solar_satellite; 35 был colonizer,
// реально frigate). См. configs/ships.yml для актуальной правды.
func approxShipValue(unitID int) int64 {
	switch unitID {
	case 29: // small_transporter
		return 4000
	case 30: // large_transporter
		return 12000
	case 31: // light_fighter
		return 4000
	case 32: // strong_fighter
		return 10000
	case 33: // cruiser
		return 27000
	case 34: // battle_ship
		return 60000
	case 35: // frigate
		return 50000
	case 36: // colony_ship
		return 30000
	case 37: // recycler
		return 16000
	case 38: // espionage_sensor (probe)
		return 1000
	case 39: // solar_satellite
		return 2000
	case 40: // bomber
		return 75000
	case 41: // star_destroyer
		return 125000
	case 42: // death_star
		return 10_000_000
	}
	return 0
}
