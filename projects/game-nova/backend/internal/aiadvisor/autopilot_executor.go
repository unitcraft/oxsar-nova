package aiadvisor

import (
	"context"
	"errors"
	"fmt"

	"oxsar/game-nova/internal/building"
	"oxsar/game-nova/internal/event"
	"oxsar/game-nova/internal/fleet"
	"oxsar/game-nova/internal/galaxy"
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
	// Fleet — для миссий (transport/expedition в Ф.2.1, atk/spy в Ф.2.2).
	Fleet *fleet.TransportService
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

	case "mission":
		if deps.Fleet == nil {
			return "", ErrUnsupportedCategory
		}
		return executeMission(ctx, deps.Fleet, userID, rec)

	default:
		return "", ErrUnsupportedCategory
	}
}

// executeMission строит TransportInput и вызывает fleet.TransportService.Send.
//
// Поддерживаемые ActionType:
//   "transport"   → mission=KindTransport (7), переносит ресурсы между своими планетами.
//   "expedition"  → mission=KindExpedition (15), отправка в случайную координату
//                   текущей системы (legacy выбирает random gal/sys/pos в галактике игрока).
//
// Атака и шпионаж — Ф.2.2.
func executeMission(ctx context.Context, fleetSvc *fleet.TransportService, userID string, rec Recommendation) (string, error) {
	switch rec.ActionType {
	case "transport":
		return executeTransport(ctx, fleetSvc, userID, rec)
	case "expedition":
		return executeExpedition(ctx, fleetSvc, userID, rec)
	case "attack":
		return executeAttack(ctx, fleetSvc, userID, rec)
	case "spy":
		return executeSpy(ctx, fleetSvc, userID, rec)
	default:
		return "", fmt.Errorf("%w: mission action %q", ErrUnsupportedCategory, rec.ActionType)
	}
}

func executeAttack(ctx context.Context, fleetSvc *fleet.TransportService, userID string, rec Recommendation) (string, error) {
	src, _ := rec.Params["src_planet_id"].(string)
	if src == "" {
		return "", fmt.Errorf("autopilot: attack: src_planet_id missing")
	}
	dstG, _ := paramInt(rec.Params, "dst_galaxy")
	dstS, _ := paramInt(rec.Params, "dst_system")
	dstP, _ := paramInt(rec.Params, "dst_position")
	if dstG == 0 && dstS == 0 && dstP == 0 {
		return "", fmt.Errorf("autopilot: attack: target coords missing")
	}
	// Скучная эвристика — отправляем все доступные боевые корабли.
	// Реальный выбор состава флота требует глубокой battle-симуляции,
	// что вынесено как отдельная задача (см. dev-log Ф.2.2).
	// Состав ниже — placeholder, fleet.Send проверит наличие в БД.
	ships := map[int]int64{unitLightFighter: 10, unitCruiser: 5}

	f, err := fleetSvc.Send(ctx, fleet.TransportInput{
		UserID:      userID,
		SrcPlanetID: src,
		Dst: galaxy.Coords{Galaxy: dstG, System: dstS, Position: dstP},
		Mission:      int(event.KindAttackSingle),
		Ships:        ships,
		SpeedPercent: 100,
	})
	if err != nil {
		return "", fmt.Errorf("autopilot: attack send: %w", err)
	}
	return f.ID, nil
}

func executeSpy(ctx context.Context, fleetSvc *fleet.TransportService, userID string, rec Recommendation) (string, error) {
	src, _ := rec.Params["src_planet_id"].(string)
	if src == "" {
		return "", fmt.Errorf("autopilot: spy: src_planet_id missing")
	}
	dstG, _ := paramInt(rec.Params, "dst_galaxy")
	dstS, _ := paramInt(rec.Params, "dst_system")
	dstP, _ := paramInt(rec.Params, "dst_position")
	if dstG == 0 && dstS == 0 && dstP == 0 {
		return "", fmt.Errorf("autopilot: spy: target coords missing")
	}
	probes, _ := paramInt64(rec.Params, "probes")
	if probes <= 0 {
		probes = 4
	}
	ships := map[int]int64{unitEspionageProbe: probes}

	f, err := fleetSvc.Send(ctx, fleet.TransportInput{
		UserID:      userID,
		SrcPlanetID: src,
		Dst: galaxy.Coords{Galaxy: dstG, System: dstS, Position: dstP},
		Mission:      int(event.KindSpy),
		Ships:        ships,
		SpeedPercent: 100,
	})
	if err != nil {
		return "", fmt.Errorf("autopilot: spy send: %w", err)
	}
	return f.ID, nil
}

func executeTransport(ctx context.Context, fleetSvc *fleet.TransportService, userID string, rec Recommendation) (string, error) {
	src, _ := rec.Params["src_planet_id"].(string)
	if src == "" {
		return "", fmt.Errorf("autopilot: transport: src_planet_id missing")
	}
	dstG, _ := paramInt(rec.Params, "dst_galaxy")
	dstS, _ := paramInt(rec.Params, "dst_system")
	dstP, _ := paramInt(rec.Params, "dst_position")
	dstMoon, _ := rec.Params["dst_is_moon"].(bool)
	carryM, _ := paramInt64(rec.Params, "carry_metal")
	carryS, _ := paramInt64(rec.Params, "carry_silicon")
	carryH, _ := paramInt64(rec.Params, "carry_hydrogen")

	// Минимальный комплект transporter'ов считаем на стороне Send;
	// здесь шлём 1 large_transporter — Send проверит хватит ли cargo.
	// Если cargo не хватает — Send вернёт ошибку, autopilot её прокинет
	// игроку как «состояние изменилось, попробуйте ещё».
	ships := chooseTransportShips(carryM + carryS + carryH)

	f, err := fleetSvc.Send(ctx, fleet.TransportInput{
		UserID:      userID,
		SrcPlanetID: src,
		Dst: galaxy.Coords{
			Galaxy:   dstG,
			System:   dstS,
			Position: dstP,
			IsMoon:   dstMoon,
		},
		Mission:      int(event.KindTransport),
		Ships:        ships,
		CarryMetal:   carryM,
		CarrySilicon: carryS,
		CarryHydro:   carryH,
		SpeedPercent: 100,
	})
	if err != nil {
		return "", fmt.Errorf("autopilot: transport send: %w", err)
	}
	return f.ID, nil
}

// chooseTransportShips подбирает минимальный комплект транспортов
// под целевой объём груза. Предпочитаем large_transporter
// (25k cargo), добиваем small (5k).
//
// Если кораблей меньше нужного — возвращаем максимально доступное;
// fleet.Send всё равно проверит наличие в транзакции.
func chooseTransportShips(totalCargo int64) map[int]int64 {
	if totalCargo <= 0 {
		return map[int]int64{unitLargeTransporter: 1}
	}
	const largeCargo = 25000
	const smallCargo = 5000
	largeNeeded := totalCargo / largeCargo
	rest := totalCargo % largeCargo
	smallNeeded := (rest + smallCargo - 1) / smallCargo

	ships := map[int]int64{}
	if largeNeeded > 0 {
		ships[unitLargeTransporter] = largeNeeded
	}
	if smallNeeded > 0 {
		ships[unitSmallTransporter] = smallNeeded
	}
	if len(ships) == 0 {
		ships[unitSmallTransporter] = 1
	}
	return ships
}

func executeExpedition(ctx context.Context, fleetSvc *fleet.TransportService, userID string, rec Recommendation) (string, error) {
	src, _ := rec.Params["src_planet_id"].(string)
	if src == "" {
		return "", fmt.Errorf("autopilot: expedition: src_planet_id missing")
	}
	// Для экспедиции легаси выбирает целью неисследованную координату.
	// Автопилот шлёт на текущую планету+1 system (упрощение; мы могли бы
	// читать coords донора из snapshot, но это сделает executor зависимым
	// от planet.Service. Send всё равно проверит валидность координат.)
	// Берём 1 large_transporter — хватит для minFleetValue экспедиции
	// (≥50k metal-eq в самих кораблях; transporter cost=12k).
	dstG, _ := paramInt(rec.Params, "dst_galaxy")
	dstS, _ := paramInt(rec.Params, "dst_system")
	dstP, _ := paramInt(rec.Params, "dst_position")
	if dstG == 0 && dstS == 0 && dstP == 0 {
		// Координаты не передали — на этом уровне executor не может
		// выбрать координаты экспедиции (нужен галактический контекст).
		// Возвращаем понятную ошибку — фронт перепросит у игрока.
		return "", fmt.Errorf("autopilot: expedition: target coords missing")
	}

	// Корабли — берём light_fighter, минимально 13 шт. (13×4000=52000 ≥ 50k порог).
	// Уточнение: executor использует то, что передал scoring; пока shoring
	// не определяет состав флота, шлём 1 large_transporter (12k cost — мало,
	// Send вернёт ErrFleetTooSmall, и UI покажет ошибку «нужно больше флота»).
	// Корректный выбор экспедиционного флота — задача Ф.2.2 (когда добавим
	// fleet-shape selection в scoring).
	ships := map[int]int64{unitLargeTransporter: 5} // 5×12k = 60k, проходит порог.

	f, err := fleetSvc.Send(ctx, fleet.TransportInput{
		UserID:      userID,
		SrcPlanetID: src,
		Dst: galaxy.Coords{
			Galaxy:   dstG,
			System:   dstS,
			Position: dstP,
		},
		Mission:      int(event.KindExpedition),
		Ships:        ships,
		SpeedPercent: 100,
	})
	if err != nil {
		return "", fmt.Errorf("autopilot: expedition send: %w", err)
	}
	return f.ID, nil
}

// paramInt извлекает int из map[string]any (JSON unmarshal даёт float64).
func paramInt(p map[string]any, key string) (int, bool) {
	if p == nil {
		return 0, false
	}
	switch v := p[key].(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	}
	return 0, false
}

func paramInt64(p map[string]any, key string) (int64, bool) {
	if p == nil {
		return 0, false
	}
	switch v := p[key].(type) {
	case int:
		return int64(v), true
	case int64:
		return v, true
	case float64:
		return int64(v), true
	}
	return 0, false
}
