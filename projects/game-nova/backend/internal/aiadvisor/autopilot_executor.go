package aiadvisor

import (
	"context"
	"errors"
	"fmt"

	"oxsar/game-nova/internal/building"
	"oxsar/game-nova/internal/research"
)

// ErrUnsupportedCategory — рекомендация в категории, для которой ещё нет
// executor-а в текущей фазе (см. план 06.1, Ф.1–Ф.3).
var ErrUnsupportedCategory = errors.New("autopilot: unsupported recommendation category")

// ErrRecommendationStale — рекомендация устарела или уже выполнена.
var ErrRecommendationStale = errors.New("autopilot: recommendation is no longer valid")

// executorDeps — зависимости executeRecommendation.
//
// Каждое поле опционально: для тестов (и для каркаса в Ф.1а) можно
// передавать nil; вызов категории, чей сервис равен nil, вернёт
// ErrUnsupportedCategory.
type executorDeps struct {
	Building *building.Service
	Research *research.Service
}

// executeRecommendation создаёт игровое событие, соответствующее
// рекомендации, и возвращает id этого события.
//
// Категории, поддерживаемые на уровне Ф.1:
//   - "building" → building.Service.Enqueue
//   - "research" → research.Service.Enqueue
//
// Прочие категории (миссии, ACS, биржа, профессия) реализуются в Ф.2–Ф.3.
//
// Все валидации (ресурсы, очередь, prereq) делегируются сервисам;
// executor не дублирует их.
func executeRecommendation(ctx context.Context, deps executorDeps, userID string, rec Recommendation) (string, error) {
	switch rec.Category {
	case "building":
		if deps.Building == nil {
			return "", ErrUnsupportedCategory
		}
		item, err := deps.Building.Enqueue(ctx, userID, rec.PlanetID, rec.UnitID)
		if err != nil {
			return "", fmt.Errorf("autopilot: execute building: %w", err)
		}
		// QueueItem.ID — id строки в construction_queue, не в events.
		// Реальный event.id связи с очередью находится в events,
		// но для возврата клиенту достаточно queue id (фронт не делает
		// различия). Если позже понадобится event.id — добавим в Repo.
		return item.ID, nil

	case "research":
		if deps.Research == nil {
			return "", ErrUnsupportedCategory
		}
		item, err := deps.Research.Enqueue(ctx, userID, rec.PlanetID, rec.UnitID)
		if err != nil {
			return "", fmt.Errorf("autopilot: execute research: %w", err)
		}
		return item.ID, nil

	default:
		return "", ErrUnsupportedCategory
	}
}
