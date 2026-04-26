package conditions

import (
	"github.com/oxsar/nova/backend/internal/goal"
)

// EventCountParams — параметры counter-условия "event_count".
//
// Цель инкрементирует прогресс на 1 при каждом event с подходящим
// EventKind. Опциональные фильтры на payload пока не поддерживаются —
// будут добавлены при первой реальной потребности (YAGNI).
//
// Целевой прогресс задаётся через Goal.Target в YAML (например, 5
// шпионских миссий = target=5).
type EventCountParams struct {
	EventKind int `json:"event_kind"`
}

func init() {
	goal.RegisterCounter("event_count", matchEventCount)
}

func matchEventCount(eventKind int, _ []byte, cond goal.ConditionSpec) bool {
	var p EventCountParams
	if err := cond.DecodeParams(&p); err != nil {
		return false
	}
	return p.EventKind == eventKind
}
