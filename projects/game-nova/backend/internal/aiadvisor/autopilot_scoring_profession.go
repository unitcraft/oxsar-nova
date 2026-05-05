package aiadvisor

import (
	"fmt"
	"time"

	"oxsar/game-nova/pkg/ids"
)

// scoreProfession — рекомендация сменить профессию под текущую стратегию.
//
// Условия:
//   - Кулдаун профессии прошёл (14 дней с прошлой смены или никогда).
//   - На балансе ≥ 1000 кредитов (стоимость смены, profession.ChangeCost).
//   - Текущая профессия отличается от рекомендованной для стратегии.
//
// Стратегия → профессия:
//   Military  → attacker
//   Defense   → defender
//   Economy   → miner
//   Expansion → universal (без штрафов; даёт ровный buff к движкам)
//   Auto      → детекция в detectPhase до scoring; не трогаем здесь.
//
// Score = mediumPriority. Эта рекомендация редкая (раз в 14 дней),
// но критичная — даёт +1 уровень многим зданиям.
const professionChangeCost = 1000.0
const professionChangeIntervalDays = 14

func scoreProfession(snap PlayerSnapshot, strategy Strategy) []Recommendation {
	target := professionForStrategy(strategy)
	if target == "" {
		return nil
	}
	current := snap.Profession
	if current == "none" {
		current = "universal" // legacy 1:1 (см. profession.Service.Get)
	}
	if current == target {
		return nil
	}
	if snap.Credits < professionChangeCost {
		return nil
	}
	// Кулдаун.
	if snap.ProfessionChangedAt != nil {
		next := snap.ProfessionChangedAt.Add(time.Duration(professionChangeIntervalDays) * 24 * time.Hour)
		if !snap.Now.IsZero() && snap.Now.Before(next) {
			return nil
		}
	}

	// Score: фиксированный, но не слишком высокий — чтобы профессия
	// не подавляла обычные строительные/исследовательские рекомендации.
	// Подбирается так, чтобы профессия попадала в топ-3 при равной
	// силе других кандидатов, но не выигрывала всегда.
	score := 50.0

	return []Recommendation{{
		ID:         ids.New(),
		Category:   "profession",
		ActionType: "change_profession_" + target,
		Params: map[string]any{
			"profession_key": target,
			"current":        current,
			"cost_credits":   professionChangeCost,
		},
		Score:       score,
		Description: fmt.Sprintf("Сменить профессию: %s → %s", professionLabel(current), professionLabel(target)),
		Benefit:     fmt.Sprintf("Под стратегию %s; стоимость %.0f кр", strategyLabel(strategy), professionChangeCost),
	}}
}

func professionForStrategy(s Strategy) string {
	switch s {
	case StrategyMilitary:
		return "attacker"
	case StrategyDefense:
		return "defender"
	case StrategyEconomy:
		return "miner"
	case StrategyExpansion:
		return "universal"
	}
	return ""
}

func professionLabel(key string) string {
	switch key {
	case "universal":
		return "Универсал"
	case "miner":
		return "Шахтёр"
	case "attacker":
		return "Атакер"
	case "defender":
		return "Оборонщик"
	case "tank":
		return "Танк"
	}
	return key
}

func strategyLabel(s Strategy) string {
	switch s {
	case StrategyEconomy:
		return "Экономика"
	case StrategyMilitary:
		return "Военная"
	case StrategyDefense:
		return "Оборона"
	case StrategyExpansion:
		return "Экспансия"
	case StrategyAuto:
		return "Авто"
	}
	return string(s)
}
