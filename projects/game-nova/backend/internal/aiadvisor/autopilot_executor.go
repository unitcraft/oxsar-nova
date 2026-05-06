package aiadvisor

import (
	"context"
	"errors"
	"fmt"

	"oxsar/game-nova/internal/building"
	"oxsar/game-nova/internal/event"
	"oxsar/game-nova/internal/fleet"
	"oxsar/game-nova/internal/galaxy"
	"oxsar/game-nova/internal/profession"
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
	// Profession — для смены профессии (Ф.3).
	Profession *profession.Service
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

	case "profession":
		if deps.Profession == nil {
			return "", ErrUnsupportedCategory
		}
		key, _ := rec.Params["profession_key"].(string)
		if key == "" {
			return "", fmt.Errorf("autopilot: profession: profession_key missing")
		}
		if err := deps.Profession.Change(ctx, userID, key); err != nil {
			return "", fmt.Errorf("autopilot: profession change: %w", err)
		}
		// profession.Change не возвращает event_id (это атомарный UPDATE
		// users + automsg, без события в очереди). Возвращаем синтетический
		// «event id» = ключ профессии, чтобы фронт смог показать результат.
		return "profession:" + key, nil

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
	case "acs_join":
		return executeACSJoin(ctx, fleetSvc, userID, rec)
	default:
		return "", fmt.Errorf("%w: mission action %q", ErrUnsupportedCategory, rec.ActionType)
	}
}

func executeACSJoin(ctx context.Context, fleetSvc *fleet.TransportService, userID string, rec Recommendation) (string, error) {
	src, _ := rec.Params["src_planet_id"].(string)
	groupID, _ := rec.Params["acs_group_id"].(string)
	if src == "" || groupID == "" {
		return "", fmt.Errorf("autopilot: acs_join: src_planet_id or acs_group_id missing")
	}
	dstG, _ := paramInt(rec.Params, "dst_galaxy")
	dstS, _ := paramInt(rec.Params, "dst_system")
	dstP, _ := paramInt(rec.Params, "dst_position")
	dstMoon, _ := rec.Params["dst_is_moon"].(bool)
	// Состав флота из scoring (pickAttackerShips).
	ships := paramShips(rec.Params)
	if len(ships) == 0 {
		return "", fmt.Errorf("autopilot: acs_join: ships missing in params")
	}

	f, err := fleetSvc.Send(ctx, fleet.TransportInput{
		UserID:      userID,
		SrcPlanetID: src,
		Dst: galaxy.Coords{Galaxy: dstG, System: dstS, Position: dstP, IsMoon: dstMoon},
		Mission:      int(event.KindAttackAlliance),
		ACSGroupID:   groupID,
		Ships:        ships,
		SpeedPercent: 100,
	})
	if err != nil {
		return "", fmt.Errorf("autopilot: acs_join send: %w", err)
	}
	return f.ID, nil
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
	// Состав флота определяется scoring (pickAttackerShips) на основе
	// реального наличия. Полный battle-aware выбор — отдельная задача
	// (см. dev-log Ф.2.2).
	ships := paramShips(rec.Params)
	if len(ships) == 0 {
		return "", fmt.Errorf("autopilot: attack: ships missing in params")
	}

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

	// Состав транспортов из scoring (pickTransporterShips) — реально
	// доступные large/small transporter с планеты-донора.
	ships := paramShips(rec.Params)
	if len(ships) == 0 {
		// Fallback: если по какой-то причине scoring не передал ships,
		// считаем по cargo (Send проверит наличие).
		ships = chooseTransportShips(carryM + carryS + carryH)
	}

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
	dstG, _ := paramInt(rec.Params, "dst_galaxy")
	dstS, _ := paramInt(rec.Params, "dst_system")
	dstP, _ := paramInt(rec.Params, "dst_position")
	if dstG == 0 && dstS == 0 && dstP == 0 {
		// scoring обязан задавать координаты (см. scoreExpeditions:
		// dst = system планеты-источника, position=15). Если по какой-то
		// причине их нет — это баг scoring; вернём понятную ошибку.
		return "", fmt.Errorf("autopilot: expedition: target coords missing")
	}

	// Корабли определяются scoring-ом (pickExpeditionShips) на основе
	// реального наличия на планете. Если ships отсутствуют в Params —
	// это баг scoring; вернём понятную ошибку.
	ships := paramShips(rec.Params)
	if len(ships) == 0 {
		return "", fmt.Errorf("autopilot: expedition: ships missing in params")
	}

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

// paramShips извлекает map ship_id → count из Params["ships"].
//
// scoring записывает map[string]any (JSON-friendly: ключи string, значения
// int64). После сериализации/десериализации через JSONB значения становятся
// float64. Парсим обратно в map[int]int64.
func paramShips(p map[string]any) map[int]int64 {
	if p == nil {
		return nil
	}
	raw, ok := p["ships"].(map[string]any)
	if !ok {
		return nil
	}
	out := make(map[int]int64, len(raw))
	for k, v := range raw {
		var id int
		if _, err := fmt.Sscanf(k, "%d", &id); err != nil {
			continue
		}
		switch n := v.(type) {
		case int:
			if n > 0 {
				out[id] = int64(n)
			}
		case int64:
			if n > 0 {
				out[id] = n
			}
		case float64:
			if n > 0 {
				out[id] = int64(n)
			}
		}
	}
	return out
}
