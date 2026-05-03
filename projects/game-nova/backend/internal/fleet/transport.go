// TRANSPORT — миссия «отвезти ресурсы на другую координату».
// Source of truth: oxsar2/www/game/EventHandler.class.php + Mission.class.php.
// EVENT_TRANSPORT = 7.
//
// Логика (упрощённая версия OGame):
//   1) Валидируем: флот/ресурсы есть, координаты валидны, цель — своя
//      планета или планета другого игрока (не умер, не udestroyed).
//   2) Считаем время полёта по самому МЕДЛЕННОМУ кораблю флота.
//   3) Списываем ресурсы с исходной планеты + «отрезаем» корабли от
//      её стока (ships.count).
//   4) Создаём запись fleets + fleet_ships.
//   5) Ставим два события: EVENT_TRANSPORT @ arrive_at (разгрузка у
//      цели + обратный путь) и EVENT_RETURN @ return_at (корабли
//      возвращаются домой).
package fleet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/artefact"
	"oxsar/game-nova/internal/config"
	"oxsar/game-nova/internal/event"
	"oxsar/game-nova/internal/galaxy"
	"oxsar/game-nova/internal/i18n"
	"oxsar/game-nova/internal/repo"
	"oxsar/game-nova/pkg/ids"
)

// TransportService реализует TRANSPORT. Отдельная структура, а не метод
// общего Service — чтобы зависимости (каталог, game speed) были явными
// и чтобы не путать с Mission-интерфейсом, который используется для
// будущих миссий (SPY/ATTACK/COLONIZE).
type TransportService struct {
	db                repo.Exec
	catalog           *config.Catalog
	speed             float64 // GAMESPEED
	numGalaxies       int     // план 72.1 ч.12 — лимит из universes.yaml
	numSystems        int     // план 72.1 ч.12 — кольцевая топология систем
	artefact          *artefact.Service
	bundle            *i18n.Bundle
	maxPlanets        int // MAX_PLANETS override (0 = computer_tech+1)
	protectionPeriod  int // seconds new player is protected
	bashingPeriod     int // seconds window for bashing count (legacy BASHING_PERIOD)
	bashingMaxAttacks int // max attacks per attacker→defender in window (legacy BASHING_MAX_ATTACKS)
	dailyQuests       DailyQuestProgresser // план 17 D — hook для прогресса quest при Send
}

// DailyQuestProgresser — узкий интерфейс к dailyquest.Service.
// Используется только в Send для инкремента прогресса fleet_mission
// quest'ов. Если nil — hook'и no-op.
type DailyQuestProgresser interface {
	IncrementProgress(ctx context.Context, userID, conditionType string,
		delta int, matcher func(condValue json.RawMessage) bool) error
}

// SetDailyQuestSvc — wire-up из server/main.go.
func (s *TransportService) SetDailyQuestSvc(p DailyQuestProgresser) {
	s.dailyQuests = p
}

func NewTransportService(db repo.Exec, cat *config.Catalog, gameSpeed float64, artefactSvc *artefact.Service, numGalaxies, numSystems int) *TransportService {
	if gameSpeed <= 0 {
		gameSpeed = 1
	}
	return &TransportService{
		db:          db,
		catalog:     cat,
		speed:       gameSpeed,
		numGalaxies: numGalaxies,
		numSystems:  numSystems,
		artefact:    artefactSvc,
	}
}

func NewTransportServiceWithConfig(db repo.Exec, cat *config.Catalog, gameSpeed float64, artefactSvc *artefact.Service, numGalaxies, numSystems, maxPlanets, protectionPeriod int) *TransportService {
	svc := NewTransportService(db, cat, gameSpeed, artefactSvc, numGalaxies, numSystems)
	svc.maxPlanets = maxPlanets
	svc.protectionPeriod = protectionPeriod
	return svc
}

// WithBundle подключает i18n-бандл для текстов сообщений.
func (s *TransportService) WithBundle(b *i18n.Bundle) *TransportService {
	s.bundle = b
	return s
}

// trFn — тип функции-переводчика для передачи в package-level helpers expedition.go.
type trFn = func(group, key string, vars map[string]string) string

// tr — helper: возвращает перевод на русском (язык пользователя пока не
// читается в event-хендлерах; пользуемся fallback-ru до реализации Ф.3.x).
func (s *TransportService) tr(group, key string, vars map[string]string) string {
	return bundleTr(s.bundle)(group, key, vars)
}

// bundleTr возвращает функцию-переводчик для package-level хелперов,
// которые не имеют доступа к TransportService (finalizeAttack, tryCreateMoon).
func bundleTr(b *i18n.Bundle) func(group, key string, vars map[string]string) string {
	return func(group, key string, vars map[string]string) string {
		if b == nil {
			return "[" + group + "." + key + "]"
		}
		return b.Tr(i18n.LangRu, group, key, vars)
	}
}

// SetBashingLimits — настройки антибашинга (план 17 A1).
// Legacy consts.dm.local.php: BASHING_PERIOD=18000 (5h), BASHING_MAX_ATTACKS=4.
// 0/0 отключает проверку (для тестов).
func (s *TransportService) SetBashingLimits(periodSec, maxAttacks int) {
	s.bashingPeriod = periodSec
	s.bashingMaxAttacks = maxAttacks
}

// TransportInput — запрос от UI на отправку флота.
// Разница между миссиями — кинд события прибытия:
//   - mission=7  → KindTransport=7
//   - mission=10 → KindAttackSingle=10
//   - mission=12 → KindAttackAlliance=12 (ACS); ACSGroupID присоединяет к группе,
//                  пустой ACSGroupID создаёт новую группу
//
// Если mission=0 в payload — считаем 7 (обратная совместимость).
type TransportInput struct {
	UserID       string
	SrcPlanetID  string
	Dst          galaxy.Coords
	Mission      int   // 7=TRANSPORT, 8=COLONIZE, 9=RECYCLING, 10=ATTACK, 11=SPY, 12=ACS, 15=EXPEDITION, 17=HOLDING
	ACSGroupID   string // только для mission=12; пусто → создать новую группу
	ColonyName   string // для mission=8 (COLONIZE); пусто → «Colony»
	Ships        map[int]int64 // unit_id -> count
	CarryMetal   int64
	CarrySilicon int64
	CarryHydro   int64
	SpeedPercent int // 10..100
	// План 72.1.47: HOLDING mission (legacy `Mission.class.php::sendFleet` L.1487).
	// Длительность удержания на цели в часах (clamp 0..99). После прибытия
	// флот стоит на dst в режиме holding, по истечении возвращается домой.
	HoldingHours int
}

// Ошибки доменного слоя.
var (
	ErrInvalidDispatch   = errors.New("fleet: invalid dispatch")
	ErrNotEnoughShips    = errors.New("fleet: not enough ships on source planet")
	ErrNotEnoughCarry    = errors.New("fleet: carried resources exceed balance")
	ErrExceedCargoCap    = errors.New("fleet: carried resources exceed total cargo capacity")
	ErrTargetNotFound    = errors.New("fleet: target coords are empty")
	ErrPlanetOwnership   = errors.New("fleet: source planet not owned by user")
	ErrUnknownShip       = errors.New("fleet: unknown ship unit_id")
	ErrFleetNotFound     = errors.New("fleet: not found")
	ErrFleetNotRecallable = errors.New("fleet: cannot recall in current state")
	ErrFleetSlotsExceeded = errors.New("fleet: no free fleet slots (improve computer_tech)")
	ErrTargetOnVacation   = errors.New("fleet: target player is on vacation (protected)")
	ErrSenderOnVacation   = errors.New("fleet: you are on vacation, cannot send fleets")
	ErrBashingLimit       = errors.New("fleet: bashing limit reached (too many attacks on this player)")
	ErrPositionNotAllowed = errors.New("fleet: POSITION only to own planets or ally/NAP targets")
	ErrExpeditionSlotsFull = errors.New("fleet: no free expedition slots (improve astro_tech)")
)

// Send — запуск TRANSPORT или ATTACK_SINGLE в зависимости от
// in.Mission. Возвращает созданный Fleet (без ID кораблей,
// только базовые поля — UI достаточно).
func (s *TransportService) Send(ctx context.Context, in TransportInput) (Fleet, error) {
	if in.Mission == 0 {
		in.Mission = int(event.KindTransport) // обратная совместимость
	}
	if !isValidMission(in.Mission) {
		return Fleet{}, fmt.Errorf("%w: mission %d not supported",
			ErrInvalidDispatch, in.Mission)
	}
	if err := in.Dst.Validate(s.numGalaxies, s.numSystems); err != nil {
		return Fleet{}, fmt.Errorf("%w: %v", ErrInvalidDispatch, err)
	}
	if in.SpeedPercent < 10 || in.SpeedPercent > 100 {
		return Fleet{}, fmt.Errorf("%w: speed_percent 10..100", ErrInvalidDispatch)
	}
	if len(in.Ships) == 0 {
		return Fleet{}, fmt.Errorf("%w: no ships selected", ErrInvalidDispatch)
	}
	if in.CarryMetal < 0 || in.CarrySilicon < 0 || in.CarryHydro < 0 {
		return Fleet{}, fmt.Errorf("%w: negative carry", ErrInvalidDispatch)
	}
	for _, c := range in.Ships {
		if c <= 0 {
			return Fleet{}, fmt.Errorf("%w: ship count must be positive", ErrInvalidDispatch)
		}
	}

	// Суммарные характеристики флота: cargo, min(speed), consume.
	specs, err := s.collectShipSpecs(in.Ships)
	if err != nil {
		return Fleet{}, err
	}
	totalCargo := int64(0)
	minSpeed := math.MaxInt
	totalConsume := 0
	totalFleetValue := int64(0)
	for _, sp := range specs {
		totalCargo += sp.spec.Cargo * sp.count
		if sp.spec.Speed > 0 && sp.spec.Speed < minSpeed {
			minSpeed = sp.spec.Speed
		}
		totalConsume += sp.spec.Fuel * int(sp.count)
		totalFleetValue += (sp.spec.Cost.Metal + sp.spec.Cost.Silicon + sp.spec.Cost.Hydrogen) * sp.count
	}
	if totalCargo < in.CarryMetal+in.CarrySilicon+in.CarryHydro {
		return Fleet{}, ErrExceedCargoCap
	}
	if minSpeed == math.MaxInt {
		return Fleet{}, fmt.Errorf("%w: fleet has no speed (empty specs?)", ErrInvalidDispatch)
	}
	// Экспедиция: минимум 50k metal-eq флота. Это отсекает
	// фарм-эксплойт BA-003 — отправку 1 LF (4k) ради 5M ресурсов.
	// 50k ≈ 10 Small Transporter'ов или 1.5 Cruiser — разумный минимум.
	// План 21 блок B1 + комментарий в 21-gameplay-hardening.md.
	if event.Kind(in.Mission) == event.KindExpedition && totalFleetValue < expeditionMinFleetValue {
		return Fleet{}, fmt.Errorf("%w: expedition requires min %d metal-eq fleet (have %d)",
			ErrInvalidDispatch, expeditionMinFleetValue, totalFleetValue)
	}

	var out Fleet
	err = s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		src, err := s.readSrcPlanet(ctx, tx, in.SrcPlanetID)
		if err != nil {
			return err
		}
		if src.userID != in.UserID {
			return ErrPlanetOwnership
		}
		// План 20 Ф.1: игрок в отпуске не может отправлять флоты.
		var senderOnVacation bool
		if err := tx.QueryRow(ctx,
			`SELECT vacation_since IS NOT NULL FROM users WHERE id=$1`,
			in.UserID).Scan(&senderOnVacation); err == nil && senderOnVacation {
			return ErrSenderOnVacation
		}
		if src.metal < in.CarryMetal || src.silicon < in.CarrySilicon || src.hydrogen < in.CarryHydro {
			return ErrNotEnoughCarry
		}
		// Fleet slots (план 20 Ф.2): maxSlots = 1 + floor(computer_tech / 6).
		// Expedition (15) и delivery_units (21) считаются отдельно
		// (см. isFleetSlotMission).
		if isFleetSlotMission(in.Mission) {
			if err := s.checkFleetSlots(ctx, tx, in.UserID); err != nil {
				return err
			}
		}
		// Expedition slots (план 20 Ф.7 + ADR-0005):
		//   max(1, floor(sqrt(astro_tech))).
		// При стартовом astro=2 → 1 слот. Чем выше astro — тем больше
		// одновременных экспедиций.
		if event.Kind(in.Mission) == event.KindExpedition {
			if err := s.checkExpeditionSlots(ctx, tx, in.UserID); err != nil {
				return err
			}
		}

		// Цель должна существовать для всех миссий, кроме COLONIZE
		// (создаёт планету) и EXPEDITION (летит в неисследованную зону).
		if requiresExistingTarget(in.Mission) {
			if err := s.checkTargetExists(ctx, tx, in.Dst); err != nil {
				return err
			}
		}
		// POSITION: только на свои планеты/луны или планеты ally/nap.
		if event.Kind(in.Mission) == event.KindPosition {
			if err := s.checkPositionTarget(ctx, tx, in.UserID, in.Dst); err != nil {
				return err
			}
		}
		// План 20 Ф.1: для агрессивных миссий проверяем, что target
		// не в отпуске. ROCKET_ATTACK идёт через другой путь
		// (rocket/service.go).
		if isAggressiveMission(in.Mission) {
			if err := s.checkTargetNotOnVacation(ctx, tx, in.Dst); err != nil {
				return err
			}
		}
		// MOON_DESTROY: цель должна быть луной + во флоте есть DS.
		if isMoonDestroyMission(in.Mission) {
			if !in.Dst.IsMoon {
				return fmt.Errorf("%w: moon-destroy target must be a moon", ErrInvalidDispatch)
			}
			if in.Ships[unitDeathstar] <= 0 {
				return fmt.Errorf("%w: moon-destroy fleet must contain Deathstar", ErrInvalidDispatch)
			}
		}
		// План 17 A1: антибашинг для атак (ATTACK/ACS/MOON_DESTROY).
		// SPY не считается (см. isAttackMission).
		if isAttackMission(in.Mission) {
			if err := s.checkBashingLimit(ctx, tx, in.UserID, in.Dst); err != nil {
				return err
			}
		}

		// Время полёта. Формула OGame (упрощённая):
		//   t = 10 + 3500 / speed_percent * sqrt(10 * dist / min_speed)
		// Мы используем ту же форму, с /GAMESPEED.
		dist := float64(galaxy.Distance(
			galaxy.Coords{Galaxy: src.galaxy, System: src.system, Position: src.position},
			in.Dst,
			s.numSystems,
		))
		duration := transportDuration(dist, minSpeed, in.SpeedPercent, s.speed)

		depart := time.Now().UTC()
		arrive := depart.Add(duration)
		returnAt := arrive.Add(duration)
		// План 72.1.47: HOLDING (kind=17) — флот стоит на dst HoldingHours
		// часов (clamp 0..99) перед возвратом. Legacy `Mission.class.php`
		// L.1487: data.duration = min(99,max(0,hours))*3600.
		var holdingHours int
		if event.Kind(in.Mission) == event.KindHolding {
			holdingHours = in.HoldingHours
			if holdingHours < 0 {
				holdingHours = 0
			}
			if holdingHours > 99 {
				holdingHours = 99
			}
			returnAt = arrive.Add(time.Duration(holdingHours)*time.Hour + duration)
		}

		// План 72.1.48 (доделка): для HOLDING-флота вычисляем
		// back_consumption и max_control_times.
		//   back_consumption = ceil(totalConsume × dist / 35000) — нижняя
		//   оценка H, нужного на возврат. Legacy `Mission.class.php`
		//   использует точную формулу из calcFleetParams; у nova fuel
		//   не списывается с планет вовсе (упрощение), поэтому минимум
		//   достаточно для unload-проверки.
		//   max_control_times = 1 + floor(comp_tech_owner/6) (legacy
		//   `NS::getMaxFleetControls`). Если на чужой планете —
		//   добавляется comp_tech_location/6, но это применимо только
		//   когда фактический owner-локации ≠ owner флота; на момент
		//   Send цель ещё не достигнута, поэтому считаем по owner'у.
		//   Когда / если в будущем добавим runtime-recalc, можно учесть.
		var backConsumption int64
		var maxControlTimes int
		if event.Kind(in.Mission) == event.KindHolding {
			backConsumption = int64(math.Ceil(float64(totalConsume) * dist / 35000.0))
			if backConsumption < 0 {
				backConsumption = 0
			}
			var compTech int
			_ = tx.QueryRow(ctx,
				`SELECT COALESCE(level,0) FROM research WHERE user_id=$1 AND unit_id=109`,
				in.UserID,
			).Scan(&compTech)
			maxControlTimes = 1 + compTech/6
		}

		// Списываем ресурсы и корабли. FOR UPDATE у планеты
		// обеспечивает, что параллельный tick или другая отправка
		// не создадут гонки.
		if err := s.chargeResources(ctx, tx, in.SrcPlanetID, in.CarryMetal, in.CarrySilicon, in.CarryHydro); err != nil {
			return err
		}
		if err := s.chargeShips(ctx, tx, in.SrcPlanetID, in.Ships); err != nil {
			return err
		}

		// Для ACS (mission=12): определяем acs_group_id и arrive_at группы.
		acsGroupID := ""
		if event.Kind(in.Mission) == event.KindAttackAlliance {
			acsGroupID, arrive, returnAt, err = s.resolveACSGroup(ctx, tx, in, arrive, returnAt)
			if err != nil {
				return err
			}
		}

		// Записываем флот.
		// План 72.1.48: для KindHolding пишем max_control_times и
		// back_consumption (для остальных типов миссий — 0).
		fleetID := ids.New()
		if acsGroupID != "" {
			if _, err := tx.Exec(ctx, `
				INSERT INTO fleets (id, owner_user_id, src_planet_id,
				                    dst_galaxy, dst_system, dst_position, dst_is_moon,
				                    mission, state, depart_at, arrive_at, return_at,
				                    carried_metal, carried_silicon, carried_hydrogen,
				                    speed_percent, acs_group_id,
				                    max_control_times, back_consumption)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'outbound', $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
			`, fleetID, in.UserID, in.SrcPlanetID,
				in.Dst.Galaxy, in.Dst.System, in.Dst.Position, in.Dst.IsMoon,
				in.Mission,
				depart, arrive, returnAt,
				in.CarryMetal, in.CarrySilicon, in.CarryHydro,
				in.SpeedPercent, acsGroupID,
				maxControlTimes, backConsumption,
			); err != nil {
				return fmt.Errorf("insert fleet: %w", err)
			}
		} else {
			if _, err := tx.Exec(ctx, `
				INSERT INTO fleets (id, owner_user_id, src_planet_id,
				                    dst_galaxy, dst_system, dst_position, dst_is_moon,
				                    mission, state, depart_at, arrive_at, return_at,
				                    carried_metal, carried_silicon, carried_hydrogen,
				                    speed_percent,
				                    max_control_times, back_consumption)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'outbound', $9, $10, $11, $12, $13, $14, $15, $16, $17)
			`, fleetID, in.UserID, in.SrcPlanetID,
				in.Dst.Galaxy, in.Dst.System, in.Dst.Position, in.Dst.IsMoon,
				in.Mission,
				depart, arrive, returnAt,
				in.CarryMetal, in.CarrySilicon, in.CarryHydro,
				in.SpeedPercent,
				maxControlTimes, backConsumption,
			); err != nil {
				return fmt.Errorf("insert fleet: %w", err)
			}
		}
		for unitID, count := range in.Ships {
			if _, err := tx.Exec(ctx, `
				INSERT INTO fleet_ships (fleet_id, unit_id, count)
				VALUES ($1, $2, $3)
			`, fleetID, unitID, count); err != nil {
				return fmt.Errorf("insert fleet_ships: %w", err)
			}
		}

		// Events: прибытие и возврат.
		// Для ACS (mission=12): одно событие KindAttackAlliance c acs_group_id в payload.
		// Все флоты группы используют одинаковый arrive_at; handler читает всю группу.
		colonyName := in.ColonyName
		if colonyName == "" {
			colonyName = "Colony"
		}
		returnEventID := ids.New()
		flightSeconds := int64(duration.Seconds())
		arrivePayload := map[string]any{
			"fleet_id":        fleetID,
			"carried":         map[string]int64{"metal": in.CarryMetal, "silicon": in.CarrySilicon, "hydrogen": in.CarryHydro},
			"acs_group_id":    acsGroupID,
			"colony_name":     colonyName,
			"return_event_id": returnEventID,
			"flight_seconds":  flightSeconds,
			"holding_hours":   holdingHours, // план 72.1.47: для KindHolding
		}
		if _, err := event.Insert(ctx, tx, event.InsertOpts{
			UserID:  &in.UserID,
			Kind:    event.Kind(in.Mission),
			FireAt:  arrive,
			Payload: arrivePayload,
		}); err != nil {
			return fmt.Errorf("insert arrive event: %w", err)
		}
		if _, err := event.Insert(ctx, tx, event.InsertOpts{
			ID:      returnEventID,
			UserID:  &in.UserID,
			Kind:    event.KindReturn,
			FireAt:  returnAt,
			Payload: arrivePayload,
		}); err != nil {
			return fmt.Errorf("insert return event: %w", err)
		}

		// Уведомление защитника за 10 минут до прибытия атакующего флота.
		if event.Kind(in.Mission) == event.KindAttackSingle ||
			event.Kind(in.Mission) == event.KindAttackAlliance {
			warnAt := arrive.Add(-10 * time.Minute)
			if warnAt.After(depart) {
				if _, err := event.Insert(ctx, tx, event.InsertOpts{
					UserID:  &in.UserID,
					Kind:    event.KindRaidWarning,
					FireAt:  warnAt,
					Payload: map[string]any{"fleet_id": fleetID},
				}); err != nil {
					return fmt.Errorf("insert raid warning event: %w", err)
				}
			}
		}

		out = Fleet{
			ID:           fleetID,
			OwnerUserID:  in.UserID,
			SrcPlanetID:  in.SrcPlanetID,
			DstGalaxy:    in.Dst.Galaxy,
			DstSystem:    in.Dst.System,
			DstPosition:  in.Dst.Position,
			DstIsMoon:    in.Dst.IsMoon,
			Mission:      in.Mission,
			State:        "outbound",
			DepartAt:     depart,
			ArriveAt:     arrive,
			ReturnAt:     &returnAt,
			Carry:        Resources{Metal: in.CarryMetal, Silicon: in.CarrySilicon, Hydrogen: in.CarryHydro},
			SpeedPercent: in.SpeedPercent,
			Ships:        in.Ships,
		}
		return nil
	})
	if err != nil {
		return out, err
	}
	// План 17 D: инкремент прогресса fleet_mission quest'ов. После
	// commit'а транзакции — quest progress не должен зависеть от
	// успеха основной TX. Если update упадёт — игрок просто не увидит
	// прогресса, но флот уже отправлен.
	if s.dailyQuests != nil {
		mission := in.Mission
		_ = s.dailyQuests.IncrementProgress(ctx, in.UserID, "fleet_mission", 1,
			func(cv json.RawMessage) bool {
				var c struct {
					Mission int `json:"mission"`
				}
				if err := json.Unmarshal(cv, &c); err != nil {
					return false
				}
				return c.Mission == mission
			})
	}
	return out, nil
}

// List возвращает активные флоты игрока (не done и не cancelled).
func (s *TransportService) List(ctx context.Context, userID string) ([]Fleet, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT id, owner_user_id, src_planet_id, dst_galaxy, dst_system, dst_position,
		       dst_is_moon, mission, state, depart_at, arrive_at, return_at,
		       carried_metal, carried_silicon, carried_hydrogen, speed_percent,
		       acs_group_id, control_times, max_control_times, back_consumption
		FROM fleets
		WHERE owner_user_id = $1 AND state IN ('outbound', 'hold', 'returning')
		ORDER BY depart_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list fleets: %w", err)
	}
	defer rows.Close()

	var out []Fleet
	for rows.Next() {
		var f Fleet
		if err := rows.Scan(&f.ID, &f.OwnerUserID, &f.SrcPlanetID, &f.DstGalaxy,
			&f.DstSystem, &f.DstPosition, &f.DstIsMoon, &f.Mission, &f.State,
			&f.DepartAt, &f.ArriveAt, &f.ReturnAt,
			&f.Carry.Metal, &f.Carry.Silicon, &f.Carry.Hydrogen, &f.SpeedPercent,
			&f.ACSGroupID, &f.ControlTimes, &f.MaxControlTimes, &f.BackConsumption,
		); err != nil {
			return nil, err
		}
		if err := s.loadShips(ctx, s.db.Pool(), &f); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// IncomingFleet — краткое описание вражеского флота летящего к планете игрока.
type IncomingFleet struct {
	ID          string    `json:"id"`
	Mission     int       `json:"mission"`
	DstGalaxy   int       `json:"dst_galaxy"`
	DstSystem   int       `json:"dst_system"`
	DstPosition int       `json:"dst_position"`
	DstIsMoon   bool      `json:"dst_is_moon"`
	ArriveAt    time.Time `json:"arrive_at"`
}

// ListIncoming возвращает чужие флоты с hostile-миссиями (атака, АКС-атака),
// летящие к планетам данного игрока.
func (s *TransportService) ListIncoming(ctx context.Context, userID string) ([]IncomingFleet, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT f.id, f.mission, f.dst_galaxy, f.dst_system, f.dst_position, f.dst_is_moon, f.arrive_at
		FROM fleets f
		JOIN planets p ON p.galaxy = f.dst_galaxy
		             AND p.system  = f.dst_system
		             AND p.position = f.dst_position
		             AND p.is_moon  = f.dst_is_moon
		             AND p.destroyed_at IS NULL
		WHERE p.user_id = $1
		  AND f.owner_user_id <> $1
		  AND f.mission IN ($2, $3)
		  AND f.state = 'outbound'
		  AND f.arrive_at > NOW()
		ORDER BY f.arrive_at ASC
	`, userID, int(event.KindAttackSingle), int(event.KindAttackAlliance))
	if err != nil {
		return nil, fmt.Errorf("list incoming fleets: %w", err)
	}
	defer rows.Close()
	var out []IncomingFleet
	for rows.Next() {
		var f IncomingFleet
		if err := rows.Scan(&f.ID, &f.Mission, &f.DstGalaxy, &f.DstSystem,
			&f.DstPosition, &f.DstIsMoon, &f.ArriveAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// Recall — досрочно вернуть флот домой. Доступно только пока флот
// летит к цели (state='outbound'): после прибытия recall не имеет
// смысла (груз уже выгружен / корабли в пути обратно). Симметрия:
// время возврата = (now - depart_at), т.е. сколько флот прошёл от
// источника, столько же будет лететь домой.
//
// Изменения в БД:
//   - fleets.state='returning', arrive_at=now, return_at=now+elapsed;
//   - удаляем wait-событие KindTransport (чтобы ArriveHandler не
//     сработал на «мертвом» флоте);
//   - переносим wait-событие KindReturn на новое return_at.
//
// ReturnHandler у нас устойчив к любому state ≠ 'done' и сам перенесёт
// ресурсы + корабли обратно на src_planet.
func (s *TransportService) Recall(ctx context.Context, userID, fleetID string) (Fleet, error) {
	var out Fleet
	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var (
			ownerID, srcPlanet, state string
			depart                    time.Time
		)
		err := tx.QueryRow(ctx, `
			SELECT owner_user_id, src_planet_id, state, depart_at
			FROM fleets WHERE id = $1 FOR UPDATE
		`, fleetID).Scan(&ownerID, &srcPlanet, &state, &depart)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrFleetNotFound
			}
			return fmt.Errorf("read fleet: %w", err)
		}
		if ownerID != userID {
			return ErrPlanetOwnership
		}
		if state != "outbound" {
			return ErrFleetNotRecallable
		}

		now := time.Now().UTC()
		elapsed := now.Sub(depart)
		if elapsed < time.Second {
			elapsed = time.Second
		}
		newReturn := now.Add(elapsed)

		if _, err := tx.Exec(ctx, `
			UPDATE fleets SET state='returning', arrive_at=$2, return_at=$3
			WHERE id = $1
		`, fleetID, now, newReturn); err != nil {
			return fmt.Errorf("update fleet: %w", err)
		}
		// Отменяем arrive-событие: просто удаляем (state='wait' значит
		// воркер его ещё не взял; FOR UPDATE SKIP LOCKED в воркере
		// гарантирует, что пересечений нет).
		if _, err := tx.Exec(ctx, `
			DELETE FROM events
			WHERE kind = $2 AND state = 'wait'
			  AND (payload->>'fleet_id') = $1
		`, fleetID, int(event.KindTransport)); err != nil {
			return fmt.Errorf("delete arrive event: %w", err)
		}
		// Переносим return-событие.
		if _, err := tx.Exec(ctx, `
			UPDATE events SET fire_at = $2
			WHERE kind = $3 AND state = 'wait'
			  AND (payload->>'fleet_id') = $1
		`, fleetID, newReturn, int(event.KindReturn)); err != nil {
			return fmt.Errorf("reschedule return event: %w", err)
		}

		// Собираем проекцию для ответа — тот же SELECT, что и List.
		row := tx.QueryRow(ctx, `
			SELECT id, owner_user_id, src_planet_id, dst_galaxy, dst_system, dst_position,
			       dst_is_moon, mission, state, depart_at, arrive_at, return_at,
			       carried_metal, carried_silicon, carried_hydrogen, speed_percent
			FROM fleets WHERE id = $1
		`, fleetID)
		if err := row.Scan(&out.ID, &out.OwnerUserID, &out.SrcPlanetID,
			&out.DstGalaxy, &out.DstSystem, &out.DstPosition, &out.DstIsMoon,
			&out.Mission, &out.State, &out.DepartAt, &out.ArriveAt, &out.ReturnAt,
			&out.Carry.Metal, &out.Carry.Silicon, &out.Carry.Hydrogen, &out.SpeedPercent,
		); err != nil {
			return fmt.Errorf("read updated fleet: %w", err)
		}
		return s.loadShips(ctx, tx, &out)
	})
	return out, err
}

// checkFleetSlots — проверка лимита одновременных флотов (план 20 Ф.2).
// Legacy: NS.class.php:1871 — maxSlots = 1 + floor(computer_tech / 6).
// Считаются только state='outbound' mission NOT IN (15 expedition,
// 29 artefact-delivery). Return-флоты (state='returning') не считаются
// — они уже летят домой и не занимают слот.
func (s *TransportService) checkFleetSlots(ctx context.Context, tx pgx.Tx, userID string) error {
	used, maxSlots, err := readFleetSlots(ctx, tx, userID)
	if err != nil {
		return err
	}
	if used >= maxSlots {
		return fmt.Errorf("%w: %d/%d slots used", ErrFleetSlotsExceeded, used, maxSlots)
	}
	return nil
}

// readFleetSlots — читает текущее и максимальное количество слотов.
// Используется и для check, и для отображения в UI.
func readFleetSlots(ctx context.Context, q interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, userID string) (used, max int, err error) {
	var computerLvl int
	err = q.QueryRow(ctx,
		`SELECT level FROM research WHERE user_id=$1 AND unit_id=$2`,
		userID, unitComputerTech).Scan(&computerLvl)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return 0, 0, fmt.Errorf("read computer_tech: %w", err)
	}
	max = 1 + computerLvl/6

	// EXPEDITION (event.KindExpedition=15) и legacy artefact-delivery
	// mission=29 не считаются как занятые слоты. См. isFleetSlotMission.
	const missionLegacyArtefactDelivery = 29
	err = q.QueryRow(ctx, `
		SELECT COUNT(*) FROM fleets
		WHERE owner_user_id = $1
		  AND state = 'outbound'
		  AND mission NOT IN ($2, $3)
	`, userID, int(event.KindExpedition), missionLegacyArtefactDelivery).Scan(&used)
	if err != nil {
		return 0, 0, fmt.Errorf("count fleets: %w", err)
	}
	return used, max, nil
}

// Slots — возвращает used/max слотов флота для userID.
func (s *TransportService) Slots(ctx context.Context, userID string) (used, max int, err error) {
	return readFleetSlots(ctx, s.db.Pool(), userID)
}

// checkExpeditionSlots — план 20 Ф.7 + ADR-0005.
//   maxSlots = max(1, floor(sqrt(astro_tech)))
//   При astro=2 → 1 слот, astro=4 → 2, astro=9 → 3, astro=16 → 4 и т.д.
// Считаются только текущие выполняемые экспедиции (state='outbound',
// mission=15). Возврат не считается — он не блокирует слот.
func (s *TransportService) checkExpeditionSlots(ctx context.Context, tx pgx.Tx, userID string) error {
	var astroLvl int
	_ = tx.QueryRow(ctx,
		`SELECT level FROM research WHERE user_id=$1 AND unit_id=$2`,
		userID, unitAstroTech).Scan(&astroLvl)
	maxSlots := int(math.Sqrt(float64(astroLvl)))
	if maxSlots < 1 {
		maxSlots = 1
	}
	var used int
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM fleets
		WHERE owner_user_id = $1
		  AND state = 'outbound'
		  AND mission = $2
	`, userID, int(event.KindExpedition)).Scan(&used); err != nil {
		return fmt.Errorf("check expedition slots: %w", err)
	}
	if used >= maxSlots {
		return fmt.Errorf("%w: %d/%d slots used (astro_tech=%d)",
			ErrExpeditionSlotsFull, used, maxSlots, astroLvl)
	}
	return nil
}

// loadShips читает состав флота из fleet_ships и заполняет Fleet.Ships.
func (s *TransportService) loadShips(ctx context.Context, q interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}, f *Fleet) error {
	rows, err := q.Query(ctx,
		`SELECT unit_id, count FROM fleet_ships WHERE fleet_id = $1`, f.ID)
	if err != nil {
		return fmt.Errorf("load fleet ships: %w", err)
	}
	defer rows.Close()
	ships := make(map[int]int64)
	for rows.Next() {
		var unitID int
		var count int64
		if err := rows.Scan(&unitID, &count); err != nil {
			return err
		}
		ships[unitID] = count
	}
	if err := rows.Err(); err != nil {
		return err
	}
	f.Ships = ships
	return nil
}

// --- internal helpers ---

type shipEntry struct {
	spec  config.ShipSpec
	count int64
}

func (s *TransportService) collectShipSpecs(ships map[int]int64) ([]shipEntry, error) {
	out := make([]shipEntry, 0, len(ships))
	for id, count := range ships {
		spec, ok := s.findShipByID(id)
		if !ok {
			return nil, fmt.Errorf("%w: id=%d", ErrUnknownShip, id)
		}
		out = append(out, shipEntry{spec: spec, count: count})
	}
	return out, nil
}

func (s *TransportService) findShipByID(id int) (config.ShipSpec, bool) {
	for _, spec := range s.catalog.Ships.Ships {
		if spec.ID == id {
			return spec, true
		}
	}
	return config.ShipSpec{}, false
}

type srcPlanet struct {
	userID   string
	galaxy   int
	system   int
	position int
	metal    int64
	silicon  int64
	hydrogen int64
}

func (s *TransportService) readSrcPlanet(ctx context.Context, tx pgx.Tx, planetID string) (srcPlanet, error) {
	var p srcPlanet
	err := tx.QueryRow(ctx, `
		SELECT user_id, galaxy, system, position, metal, silicon, hydrogen
		FROM planets WHERE id = $1 AND destroyed_at IS NULL
		FOR UPDATE
	`, planetID).Scan(&p.userID, &p.galaxy, &p.system, &p.position,
		&p.metal, &p.silicon, &p.hydrogen)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return p, ErrPlanetOwnership // «нет источника» ≈ «не твоя»
		}
		return p, fmt.Errorf("read src planet: %w", err)
	}
	return p, nil
}

func (s *TransportService) checkTargetExists(ctx context.Context, tx pgx.Tx, c galaxy.Coords) error {
	var exists bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM planets
			WHERE galaxy = $1 AND system = $2 AND position = $3 AND is_moon = $4
			  AND destroyed_at IS NULL
		)
	`, c.Galaxy, c.System, c.Position, c.IsMoon).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check target: %w", err)
	}
	if !exists {
		return ErrTargetNotFound
	}
	return nil
}

// checkPositionTarget — план 20 Ф.3: POSITION (mission=6) разрешён
// только на свои планеты/луны или на планеты игроков-союзников
// (alliance_relationships.relation IN ('ally','nap') AND status='active').
func (s *TransportService) checkPositionTarget(ctx context.Context, tx pgx.Tx,
	attackerID string, dst galaxy.Coords) error {
	var targetUserID *string
	err := tx.QueryRow(ctx, `
		SELECT user_id FROM planets
		WHERE galaxy=$1 AND system=$2 AND position=$3 AND is_moon=$4
		  AND destroyed_at IS NULL
	`, dst.Galaxy, dst.System, dst.Position, dst.IsMoon).Scan(&targetUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrTargetNotFound
		}
		return fmt.Errorf("position: read target: %w", err)
	}
	if targetUserID == nil {
		return ErrPositionNotAllowed // астероид/пустой слот
	}
	if *targetUserID == attackerID {
		return nil // собственная планета/луна
	}
	// Проверка alliance relation.
	var allied bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM users ua
			JOIN users ut ON ut.id = $2
			JOIN alliance_relationships r
			  ON (r.alliance_id = ua.alliance_id AND r.target_alliance_id = ut.alliance_id)
			  OR (r.alliance_id = ut.alliance_id AND r.target_alliance_id = ua.alliance_id)
			WHERE ua.id = $1
			  AND ua.alliance_id IS NOT NULL
			  AND ut.alliance_id IS NOT NULL
			  AND r.relation IN ('ally', 'nap')
			  AND r.status = 'active'
		)
	`, attackerID, *targetUserID).Scan(&allied)
	if err != nil {
		return fmt.Errorf("position: check relation: %w", err)
	}
	if !allied {
		return ErrPositionNotAllowed
	}
	return nil
}

// checkBashingLimit — план 17 A1. Легаси NS.class.php:2285,
// consts.dm.local.php: BASHING_PERIOD=18000 (5h), BASHING_MAX_ATTACKS=4.
// Считаем все атаки (mission=10 ATTACK_SINGLE, 12 ATTACK_ALLIANCE)
// от attacker → любая планета defender за последние BASHING_PERIOD
// секунд, включая pending (state='outbound') и finished (arrive_at > now-5h).
// Если ≥ maxAttacks — отказ.
func (s *TransportService) checkBashingLimit(ctx context.Context, tx pgx.Tx,
	attackerID string, dst galaxy.Coords) error {
	if s.bashingMaxAttacks <= 0 || s.bashingPeriod <= 0 {
		return nil // проверка отключена
	}
	// Найдём user_id владельца целевой планеты.
	var defenderID *string
	err := tx.QueryRow(ctx, `
		SELECT user_id FROM planets
		WHERE galaxy=$1 AND system=$2 AND position=$3 AND is_moon=$4
		  AND destroyed_at IS NULL
	`, dst.Galaxy, dst.System, dst.Position, dst.IsMoon).Scan(&defenderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil // цели нет — проверку другого слоя повторять не будем
		}
		return fmt.Errorf("bashing: read defender: %w", err)
	}
	if defenderID == nil || *defenderID == attackerID {
		return nil // пусто (asteroid) или self-attack (обрабатывается отдельно)
	}
	// COUNT атак attacker → все планеты defender в окне.
	var cnt int
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM fleets f
		JOIN planets p
		  ON p.galaxy = f.dst_galaxy AND p.system = f.dst_system
		 AND p.position = f.dst_position AND p.is_moon = f.dst_is_moon
		 AND p.destroyed_at IS NULL
		WHERE f.owner_user_id = $1
		  AND p.user_id = $2
		  AND f.mission IN ($4, $5)
		  AND (
		    f.state = 'outbound'
		    OR (f.arrive_at IS NOT NULL AND f.arrive_at > now() - ($3 * interval '1 second'))
		  )
	`, attackerID, *defenderID, s.bashingPeriod,
		int(event.KindAttackSingle), int(event.KindAttackAlliance)).Scan(&cnt)
	if err != nil {
		return fmt.Errorf("bashing: count: %w", err)
	}
	if cnt >= s.bashingMaxAttacks {
		return fmt.Errorf("%w: %d/%d attacks in last %ds",
			ErrBashingLimit, cnt, s.bashingMaxAttacks, s.bashingPeriod)
	}
	return nil
}

// checkTargetNotOnVacation — для ATTACK_SINGLE (10), SPY (11),
// ATTACK_ALLIANCE (12), ROCKET_ATTACK (16) проверяет, что владелец
// целевой планеты не в режиме отпуска (план 20 Ф.1). Возвращает
// ErrTargetOnVacation если да.
func (s *TransportService) checkTargetNotOnVacation(ctx context.Context, tx pgx.Tx, c galaxy.Coords) error {
	var onVacation bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM planets p
			JOIN users u ON u.id = p.user_id
			WHERE p.galaxy=$1 AND p.system=$2 AND p.position=$3 AND p.is_moon=$4
			  AND p.destroyed_at IS NULL
			  AND u.vacation_since IS NOT NULL
		)
	`, c.Galaxy, c.System, c.Position, c.IsMoon).Scan(&onVacation)
	if err != nil {
		return fmt.Errorf("check vacation: %w", err)
	}
	if onVacation {
		return ErrTargetOnVacation
	}
	return nil
}

func (s *TransportService) chargeResources(ctx context.Context, tx pgx.Tx, planetID string, m, si, h int64) error {
	if _, err := tx.Exec(ctx, `
		UPDATE planets
		SET metal = metal - $1, silicon = silicon - $2, hydrogen = hydrogen - $3
		WHERE id = $4
	`, m, si, h, planetID); err != nil {
		return fmt.Errorf("charge carry: %w", err)
	}
	return nil
}

func (s *TransportService) chargeShips(ctx context.Context, tx pgx.Tx, planetID string, ships map[int]int64) error {
	for unitID, count := range ships {
		// Проверяем достаточность и списываем одним запросом (RETURNING
		// чтобы увидеть, сколько осталось — если < 0, откатимся).
		var remaining int64
		err := tx.QueryRow(ctx, `
			UPDATE ships SET count = count - $1
			WHERE planet_id = $2 AND unit_id = $3
			RETURNING count
		`, count, planetID, unitID).Scan(&remaining)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("%w: unit_id=%d", ErrNotEnoughShips, unitID)
			}
			return fmt.Errorf("charge ship %d: %w", unitID, err)
		}
		if remaining < 0 {
			return fmt.Errorf("%w: unit_id=%d", ErrNotEnoughShips, unitID)
		}
	}
	return nil
}

// transportDuration — OGame-like формула полёта.
func transportDuration(distance float64, minSpeed, speedPercent int, gameSpeed float64) time.Duration {
	if minSpeed <= 0 {
		return time.Minute
	}
	if speedPercent <= 0 {
		speedPercent = 100
	}
	raw := 10 + 3500.0*math.Sqrt(10*distance/float64(minSpeed))/float64(speedPercent)
	if gameSpeed > 0 {
		raw /= gameSpeed
	}
	if raw < 1 {
		raw = 1
	}
	return time.Duration(raw * float64(time.Second))
}

// resolveACSGroup возвращает acs_group_id, скорректированные arrive и returnAt для ACS-флота.
// Если in.ACSGroupID задан — присоединяемся к группе (используем её arrive_at).
// Иначе создаём новую группу с arrive_at текущего флота.
func (s *TransportService) resolveACSGroup(ctx context.Context, tx pgx.Tx, in TransportInput,
	arrive, returnAt time.Time) (string, time.Time, time.Time, error) {
	if in.ACSGroupID != "" {
		// План 72.1.48: проверка accepted invitation (или сам leader).
		// Без accepted_at — нельзя присоединяться (legacy: только
		// явно приглашённые через formation_invitation).
		var allowed bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM acs_groups
				WHERE id=$1 AND leader_user_id=$2
			) OR EXISTS (
				SELECT 1 FROM acs_invitations
				WHERE acs_group_id=$1 AND user_id=$2 AND accepted_at IS NOT NULL
			)
		`, in.ACSGroupID, in.UserID).Scan(&allowed); err != nil {
			return "", time.Time{}, time.Time{}, fmt.Errorf("acs invite check: %w", err)
		}
		if !allowed {
			return "", time.Time{}, time.Time{},
				fmt.Errorf("%w: not invited to ACS group (need accepted_at)", ErrInvalidDispatch)
		}
		// Join: проверяем что группа существует и цель совпадает.
		var gArriveAt time.Time
		err := tx.QueryRow(ctx, `
			SELECT arrive_at FROM acs_groups
			WHERE id=$1 AND target_galaxy=$2 AND target_system=$3
			  AND target_position=$4 AND target_is_moon=$5
			FOR UPDATE
		`, in.ACSGroupID, in.Dst.Galaxy, in.Dst.System, in.Dst.Position, in.Dst.IsMoon).
			Scan(&gArriveAt)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return "", time.Time{}, time.Time{},
					fmt.Errorf("%w: acs group not found or target mismatch", ErrInvalidDispatch)
			}
			return "", time.Time{}, time.Time{}, fmt.Errorf("acs join: %w", err)
		}
		// Флот прибывает вместе с группой (arrive_at группы ≥ arrive_at флота).
		if gArriveAt.After(arrive) {
			arrive = gArriveAt
		} else {
			// Флот медленнее группы — обновляем arrive_at группы.
			if _, err := tx.Exec(ctx,
				`UPDATE acs_groups SET arrive_at=$1 WHERE id=$2`, arrive, in.ACSGroupID); err != nil {
				return "", time.Time{}, time.Time{}, fmt.Errorf("acs update arrive: %w", err)
			}
		}
		dur := arrive.Sub(time.Now().UTC())
		returnAt = arrive.Add(dur)
		return in.ACSGroupID, arrive, returnAt, nil
	}

	// Создаём новую группу.
	groupID := ids.New()
	if _, err := tx.Exec(ctx, `
		INSERT INTO acs_groups (id, target_galaxy, target_system, target_position, target_is_moon, arrive_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, groupID, in.Dst.Galaxy, in.Dst.System, in.Dst.Position, in.Dst.IsMoon, arrive); err != nil {
		return "", time.Time{}, time.Time{}, fmt.Errorf("acs create group: %w", err)
	}
	return groupID, arrive, returnAt, nil
}
