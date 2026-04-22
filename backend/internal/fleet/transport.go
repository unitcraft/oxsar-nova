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

	"github.com/oxsar/nova/backend/internal/config"
	"github.com/oxsar/nova/backend/internal/event"
	"github.com/oxsar/nova/backend/internal/galaxy"
	"github.com/oxsar/nova/backend/internal/repo"
	"github.com/oxsar/nova/backend/pkg/ids"
)

// TransportService реализует TRANSPORT. Отдельная структура, а не метод
// общего Service — чтобы зависимости (каталог, game speed) были явными
// и чтобы не путать с Mission-интерфейсом, который используется для
// будущих миссий (SPY/ATTACK/COLONIZE).
type TransportService struct {
	db      repo.Exec
	catalog *config.Catalog
	speed   float64 // GAMESPEED
}

func NewTransportService(db repo.Exec, cat *config.Catalog, gameSpeed float64) *TransportService {
	if gameSpeed <= 0 {
		gameSpeed = 1
	}
	return &TransportService{db: db, catalog: cat, speed: gameSpeed}
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
	Mission      int   // 7=TRANSPORT, 8=COLONIZE, 9=RECYCLING, 10=ATTACK, 11=SPY, 12=ACS, 15=EXPEDITION
	ACSGroupID   string // только для mission=12; пусто → создать новую группу
	ColonyName   string // для mission=8 (COLONIZE); пусто → «Colony»
	Ships        map[int]int64 // unit_id -> count
	CarryMetal   int64
	CarrySilicon int64
	CarryHydro   int64
	SpeedPercent int // 10..100
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
)

// Send — запуск TRANSPORT или ATTACK_SINGLE в зависимости от
// in.Mission. Возвращает созданный Fleet (без ID кораблей,
// только базовые поля — UI достаточно).
func (s *TransportService) Send(ctx context.Context, in TransportInput) (Fleet, error) {
	if in.Mission == 0 {
		in.Mission = 7 // обратная совместимость
	}
	if in.Mission != 7 && in.Mission != 8 && in.Mission != 9 && in.Mission != 10 &&
		in.Mission != 11 && in.Mission != 12 && in.Mission != 15 {
		return Fleet{}, fmt.Errorf("%w: mission %d not supported",
			ErrInvalidDispatch, in.Mission)
	}
	if err := in.Dst.Validate(); err != nil {
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
	for _, sp := range specs {
		totalCargo += sp.spec.Cargo * sp.count
		if sp.spec.Speed > 0 && sp.spec.Speed < minSpeed {
			minSpeed = sp.spec.Speed
		}
		totalConsume += sp.spec.Fuel * int(sp.count)
	}
	if totalCargo < in.CarryMetal+in.CarrySilicon+in.CarryHydro {
		return Fleet{}, ErrExceedCargoCap
	}
	if minSpeed == math.MaxInt {
		return Fleet{}, fmt.Errorf("%w: fleet has no speed (empty specs?)", ErrInvalidDispatch)
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
		if src.metal < in.CarryMetal || src.silicon < in.CarrySilicon || src.hydrogen < in.CarryHydro {
			return ErrNotEnoughCarry
		}

		// Для TRANSPORT/ATTACK/RECYCLING/SPY требуется существующая
		// цель. Для COLONIZE (mission=8) и EXPEDITION (mission=15)
		// цель может быть любой: COLONIZE создаёт планету, EXPEDITION
		// летит в неисследованную зону.
		if in.Mission != 8 && in.Mission != 15 {
			if err := s.checkTargetExists(ctx, tx, in.Dst); err != nil {
				return err
			}
		}

		// Время полёта. Формула OGame (упрощённая):
		//   t = 10 + 3500 / speed_percent * sqrt(10 * dist / min_speed)
		// Мы используем ту же форму, с /GAMESPEED.
		dist := float64(galaxy.Distance(
			galaxy.Coords{Galaxy: src.galaxy, System: src.system, Position: src.position},
			in.Dst,
		))
		duration := transportDuration(dist, minSpeed, in.SpeedPercent, s.speed)

		depart := time.Now().UTC()
		arrive := depart.Add(duration)
		returnAt := arrive.Add(duration)

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
		fleetID := ids.New()
		if acsGroupID != "" {
			if _, err := tx.Exec(ctx, `
				INSERT INTO fleets (id, owner_user_id, src_planet_id,
				                    dst_galaxy, dst_system, dst_position, dst_is_moon,
				                    mission, state, depart_at, arrive_at, return_at,
				                    carried_metal, carried_silicon, carried_hydrogen,
				                    speed_percent, acs_group_id)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'outbound', $9, $10, $11, $12, $13, $14, $15, $16)
			`, fleetID, in.UserID, in.SrcPlanetID,
				in.Dst.Galaxy, in.Dst.System, in.Dst.Position, in.Dst.IsMoon,
				in.Mission,
				depart, arrive, returnAt,
				in.CarryMetal, in.CarrySilicon, in.CarryHydro,
				in.SpeedPercent, acsGroupID,
			); err != nil {
				return fmt.Errorf("insert fleet: %w", err)
			}
		} else {
			if _, err := tx.Exec(ctx, `
				INSERT INTO fleets (id, owner_user_id, src_planet_id,
				                    dst_galaxy, dst_system, dst_position, dst_is_moon,
				                    mission, state, depart_at, arrive_at, return_at,
				                    carried_metal, carried_silicon, carried_hydrogen,
				                    speed_percent)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'outbound', $9, $10, $11, $12, $13, $14, $15)
			`, fleetID, in.UserID, in.SrcPlanetID,
				in.Dst.Galaxy, in.Dst.System, in.Dst.Position, in.Dst.IsMoon,
				in.Mission,
				depart, arrive, returnAt,
				in.CarryMetal, in.CarrySilicon, in.CarryHydro,
				in.SpeedPercent,
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
		payload, _ := json.Marshal(map[string]any{
			"fleet_id":     fleetID,
			"carried":      map[string]int64{"metal": in.CarryMetal, "silicon": in.CarrySilicon, "hydrogen": in.CarryHydro},
			"acs_group_id": acsGroupID,
			"colony_name":  colonyName,
		})
		if _, err := tx.Exec(ctx, `
			INSERT INTO events (id, user_id, kind, state, fire_at, payload)
			VALUES ($1, $2, $3, 'wait', $4, $5)
		`, ids.New(), in.UserID, in.Mission, arrive, payload); err != nil {
			return fmt.Errorf("insert arrive event: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO events (id, user_id, kind, state, fire_at, payload)
			VALUES ($1, $2, 20, 'wait', $3, $4)
		`, ids.New(), in.UserID, returnAt, payload); err != nil {
			return fmt.Errorf("insert return event: %w", err)
		}

		// Уведомление защитника за 10 минут до прибытия атакующего флота.
		if event.Kind(in.Mission) == event.KindAttackSingle ||
			event.Kind(in.Mission) == event.KindAttackAlliance {
			warnAt := arrive.Add(-10 * time.Minute)
			if warnAt.After(depart) {
				warnPayload, _ := json.Marshal(map[string]any{"fleet_id": fleetID})
				if _, err := tx.Exec(ctx, `
					INSERT INTO events (id, user_id, kind, state, fire_at, payload)
					VALUES ($1, $2, $3, 'wait', $4, $5)
				`, ids.New(), in.UserID, event.KindRaidWarning, warnAt, warnPayload); err != nil {
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
	return out, err
}

// List возвращает активные флоты игрока (не done и не cancelled).
func (s *TransportService) List(ctx context.Context, userID string) ([]Fleet, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT id, owner_user_id, src_planet_id, dst_galaxy, dst_system, dst_position,
		       dst_is_moon, mission, state, depart_at, arrive_at, return_at,
		       carried_metal, carried_silicon, carried_hydrogen, speed_percent
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
		  AND f.mission IN (10, 12)
		  AND f.state = 'outbound'
		  AND f.arrive_at > NOW()
		ORDER BY f.arrive_at ASC
	`, userID)
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
			WHERE kind = 7 AND state = 'wait'
			  AND (payload->>'fleet_id') = $1
		`, fleetID); err != nil {
			return fmt.Errorf("delete arrive event: %w", err)
		}
		// Переносим return-событие.
		if _, err := tx.Exec(ctx, `
			UPDATE events SET fire_at = $2
			WHERE kind = 20 AND state = 'wait'
			  AND (payload->>'fleet_id') = $1
		`, fleetID, newReturn); err != nil {
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
