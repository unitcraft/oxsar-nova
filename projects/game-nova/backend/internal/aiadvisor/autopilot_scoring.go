package aiadvisor

import (
	"sort"

	"oxsar/game-nova/internal/config"
)

// scoringInputs — DI для scoreCandidates (тесты дают моки/ин-мемори каталоги).
type scoringInputs struct {
	Catalog *config.Catalog
}

// scoreCandidates — pure-функция: snapshot + стратегия → отсортированный
// список ScoredAction (по убыванию Score).
//
// Ф.1в реализует категории "building" и "research". Ф.2 добавит миссии.
// Ф.3 — ACS, биржу, профессию. Стратегия "auto" в Ф.5 выбирает целевую
// функцию по эвристике фазы (см. detectPhase).
//
// На уровне Ф.1а функция возвращает пустой slice; это позволяет
// проверить весь pipeline (Enqueue → воркер → Compute → ai_advisor_log
// status='ready' с пустым массивом recs → Result), не дожидаясь
// реализации scoring.
func scoreCandidates(snap PlayerSnapshot, strategy Strategy, inputs scoringInputs) []Recommendation {
	if strategy == StrategyAuto {
		strategy = detectPhase(snap)
	}

	var recs []Recommendation

	// Ф.1в: building candidates.
	// Ф.1г: research candidates.
	// Ф.2:  mission candidates.
	// Ф.3:  acs/exchange/profession candidates.

	sort.SliceStable(recs, func(i, j int) bool {
		return recs[i].Score > recs[j].Score
	})

	if len(recs) > 3 {
		recs = recs[:3]
	}
	return recs
}

// detectPhase — эвристика фазы развития для StrategyAuto.
// Возвращает целевую стратегию, под которой scoreCandidates будет
// ранжировать кандидатов.
//
// Реализация в Ф.5 — пока возвращает Economy как безопасный дефолт.
func detectPhase(_ PlayerSnapshot) Strategy {
	return StrategyEconomy
}
