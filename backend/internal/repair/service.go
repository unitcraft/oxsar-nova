// Package repair — ремонтная фабрика (legacy ext/page/ExtRepair).
//
// M2-MVP: только DISASSEMBLE здоровых юнитов (без damaged/shell_percent),
// batch-режим (всё одной очередью, без per-unit перепланирования, как в
// legacy — см. ExtEventHandler::disassemble).
//
// Экономика (источник: ExtRepair::setDisassembleUnitRequirements):
//   required = ceil(base * 0.2 / 10) * 10      -- списывается при enqueue
//   return   = ceil(base * 0.9 / 10) * 10      -- зачисляется при finish
//   earn     = return - required               -- чистая прибыль (=base*0.7)
//   duration = BuildDuration(0, {m:base*0.1, s:base*0.1, h:0}, ...)
//
// REPAIR (восстановление повреждённых) отложен до M4, когда в таблице
// ships появится значимый damaged_count/shell_percent (после порта боя).
//
// Event: kind=51 (EVENT_DISASSEMBLE, совпадает с legacy).
package repair

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/config"
	"github.com/oxsar/nova/backend/internal/economy"
	"github.com/oxsar/nova/backend/internal/event"
	"github.com/oxsar/nova/backend/internal/planet"
	"github.com/oxsar/nova/backend/internal/repo"
	"github.com/oxsar/nova/backend/internal/requirements"
	"github.com/oxsar/nova/backend/pkg/ids"
)

var (
	ErrUnknownUnit       = errors.New("repair: unknown unit")
	ErrNotEnoughRes      = errors.New("repair: not enough resources for disassembly cost")
	ErrNotEnoughShips    = errors.New("repair: not enough units to disassemble")
	ErrPlanetOwnership   = errors.New("repair: planet not owned by user")
	ErrNoRepairBuilding  = errors.New("repair: repair_factory required")
	ErrInvalidCount      = errors.New("repair: invalid count")
	ErrQueueItemNotFound = errors.New("repair: queue item not found")
	ErrNothingToRepair   = errors.New("repair: no damaged units")
)

type Service struct {
	db      repo.Exec
	planets *planet.Service
	catalog *config.Catalog
	reqs    *requirements.Checker
	gameSpd float64
}

func NewService(db repo.Exec, p *planet.Service, cat *config.Catalog, reqs *requirements.Checker, gameSpeed float64) *Service {
	if gameSpeed <= 0 {
		gameSpeed = 1
	}
	return &Service{db: db, planets: p, catalog: cat, reqs: reqs, gameSpd: gameSpeed}
}

type QueueItem struct {
	ID             string    `json:"id"`
	PlanetID       string    `json:"planet_id"`
	UserID         string    `json:"user_id"`
	UnitID         int       `json:"unit_id"`
	IsDefense      bool      `json:"is_defense"`
	Mode           string    `json:"mode"`
	Count          int64     `json:"count"`
	ReturnMetal    int64     `json:"return_metal"`
	ReturnSilicon  int64     `json:"return_silicon"`
	ReturnHydrogen int64     `json:"return_hydrogen"`
	PerUnitSeconds int       `json:"per_unit_seconds"`
	StartAt        time.Time `json:"start_at"`
	EndAt          time.Time `json:"end_at"`
	Status         string    `json:"status"`
}

// EnqueueDisassemble ставит юниты в очередь на разбор. Семантика batch:
// списываем сразу count×required_cost и count юнитов из ships/defense,
// по завершении события зачисляем count×return_cost.
//
// Предусловия:
//   - планета принадлежит userID;
//   - на планете есть repair_factory уровня >= 1;
//   - на планете count юнитов указанного типа;
//   - метал/кремний/водород покрывают required_cost * count.
func (s *Service) EnqueueDisassemble(ctx context.Context, userID, planetID string, unitID int, count int64) (QueueItem, error) {
	if count <= 0 {
		return QueueItem{}, ErrInvalidCount
	}
	key, baseCost, isDefense, ok := s.lookupUnit(unitID)
	if !ok {
		return QueueItem{}, ErrUnknownUnit
	}

	p, err := s.planets.Get(ctx, planetID)
	if err != nil {
		return QueueItem{}, err
	}
	if p.UserID != userID {
		return QueueItem{}, ErrPlanetOwnership
	}

	// Per-unit required/return (legacy ceil(_/10)*10).
	req := scalePerUnit(baseCost, 0.2)
	ret := scalePerUnit(baseCost, 0.9)
	totalReq := multiplyCost(req, count)

	var item QueueItem
	err = s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// 1. repair_factory >= 1.
		repairSpec, hasRepair := s.catalog.Buildings.Buildings["repair_factory"]
		if !hasRepair {
			// Каталог не содержит repair_factory — считаем, что фича недоступна
			// на этом балансе. Лучше явная ошибка, чем тихое поведение.
			return ErrNoRepairBuilding
		}
		var repairLvl int
		err := tx.QueryRow(ctx,
			`SELECT level FROM buildings WHERE planet_id=$1 AND unit_id=$2`,
			planetID, repairSpec.ID,
		).Scan(&repairLvl)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("repair_factory level: %w", err)
		}
		if repairLvl < 1 {
			return ErrNoRepairBuilding
		}

		// 2. Зависимости юнита (на всякий случай: disassemble не строит,
		//    но если requirements снесены — логичнее блокировать).
		if err := s.reqs.Check(ctx, tx, key, userID, planetID); err != nil {
			return err
		}

		// 3. Хватает ли юнитов. Таблица зависит от is_defense.
		stockTable := "ships"
		if isDefense {
			stockTable = "defense"
		}
		var stock int64
		err = tx.QueryRow(ctx,
			`SELECT count FROM `+stockTable+` WHERE planet_id=$1 AND unit_id=$2 FOR UPDATE`,
			planetID, unitID,
		).Scan(&stock)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotEnoughShips
			}
			return fmt.Errorf("read stock: %w", err)
		}
		if stock < count {
			return ErrNotEnoughShips
		}

		// 4. Хватает ли ресурсов на required. Читаем внутри tx (FOR UPDATE),
		//    чтобы избежать TOCTOU при параллельных enqueue.
		var curMetal, curSilicon, curHydrogen float64
		if err := tx.QueryRow(ctx,
			`SELECT metal, silicon, hydrogen FROM planets WHERE id=$1 FOR UPDATE`,
			planetID,
		).Scan(&curMetal, &curSilicon, &curHydrogen); err != nil {
			return fmt.Errorf("read planet res: %w", err)
		}
		if int64(curMetal) < totalReq.Metal ||
			int64(curSilicon) < totalReq.Silicon ||
			int64(curHydrogen) < totalReq.Hydrogen {
			return ErrNotEnoughRes
		}

		// 5. Списание юнитов и required-ресурсов.
		if _, err := tx.Exec(ctx,
			`UPDATE `+stockTable+` SET count = count - $1 WHERE planet_id=$2 AND unit_id=$3`,
			count, planetID, unitID,
		); err != nil {
			return fmt.Errorf("charge stock: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE planets SET metal = metal - $1, silicon = silicon - $2, hydrogen = hydrogen - $3
			WHERE id = $4
		`, totalReq.Metal, totalReq.Silicon, totalReq.Hydrogen, planetID); err != nil {
			return fmt.Errorf("charge res: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO res_log (user_id, planet_id, reason, delta_metal, delta_silicon, delta_hydrogen)
			VALUES ($1, $2, 'disassemble_cost', $3, $4, $5)
		`, userID, planetID, -totalReq.Metal, -totalReq.Silicon, -totalReq.Hydrogen); err != nil {
			return fmt.Errorf("res_log: %w", err)
		}

		// 6. Время. Legacy: NS::getBuildingTime(base*0.1, base*0.1, mode).
		//    Мы используем ту же формулу, что и shipyard — BuildDuration,
		//    но вход — «надбавка 10%» за юнит.
		perUnit := economy.BuildDuration(1,
			economy.Cost{
				Metal:    int64(math.Round(float64(baseCost.Metal) * 0.1)),
				Silicon:  int64(math.Round(float64(baseCost.Silicon) * 0.1)),
				Hydrogen: 0,
			},
			repairLvl, 0, s.gameSpd)
		perUnitSec := int(math.Max(1, math.Round(perUnit.Seconds())))
		totalDur := time.Duration(perUnitSec) * time.Duration(count) * time.Second
		start := time.Now().UTC()
		end := start.Add(totalDur)

		// 7. Запись очереди.
		totalRet := multiplyCost(ret, count)
		id := ids.New()
		if _, err := tx.Exec(ctx, `
			INSERT INTO repair_queue
				(id, planet_id, user_id, unit_id, is_defense, mode, count,
				 return_metal, return_silicon, return_hydrogen,
				 per_unit_seconds, start_at, end_at, status)
			VALUES ($1, $2, $3, $4, $5, 'disassemble', $6, $7, $8, $9, $10, $11, $12, 'running')
		`, id, planetID, userID, unitID, isDefense, count,
			totalRet.Metal, totalRet.Silicon, totalRet.Hydrogen,
			perUnitSec, start, end); err != nil {
			return fmt.Errorf("insert queue: %w", err)
		}

		// 8. Событие.
		payload, _ := json.Marshal(map[string]any{
			"queue_id": id,
			"mode":     "disassemble",
		})
		if _, err := tx.Exec(ctx, `
			INSERT INTO events (id, user_id, planet_id, kind, state, fire_at, payload)
			VALUES ($1, $2, $3, $4, 'wait', $5, $6)
		`, ids.New(), userID, planetID, event.KindDisassemble, end, payload); err != nil {
			return fmt.Errorf("insert event: %w", err)
		}

		item = QueueItem{
			ID: id, PlanetID: planetID, UserID: userID, UnitID: unitID,
			IsDefense: isDefense, Mode: "disassemble", Count: count,
			ReturnMetal: totalRet.Metal, ReturnSilicon: totalRet.Silicon, ReturnHydrogen: totalRet.Hydrogen,
			PerUnitSeconds: perUnitSec, StartAt: start, EndAt: end, Status: "running",
		}
		return nil
	})
	return item, err
}

// EnqueueRepair чинит ВСЕХ damaged-юнитов одного типа (unit_id) на
// планете. Batch-семантика: списываем required-ресурсы сразу, при
// finish сбрасываем damaged_count/shell_percent к 0.
//
// Формула (legacy ExtRepair::setRepairUnitRequirements):
//   struct_scale = 0.1 × (100 - shell_percent) / 100
//   required_{m,s,h} = ceil(base × struct_scale / 10) × 10
//   required_time   = buildTime(base×0.1, base×0.1, mode)
//
// В M4.4c упрощаем: берём усреднённый shell_percent из ships
// (у нас в M4.1+ сохраняем его на уровне stack'а). Если на стэке
// есть N damaged-юнитов с общим shell_percent — платим за них по
// одной формуле, это корректно в рамках нашей модели (1 damaged на
// stack — см. commitDamage в battle/engine.go).
//
// Defense repair не поддерживается: legacy-таблица defense не
// хранит damaged. Если защита разрушена — её строят заново через
// shipyard.
func (s *Service) EnqueueRepair(ctx context.Context, userID, planetID string, unitID int) (QueueItem, error) {
	_, baseCost, isDefense, ok := s.lookupUnit(unitID)
	if !ok {
		return QueueItem{}, ErrUnknownUnit
	}
	if isDefense {
		// Защита в M4.4c не чинится.
		return QueueItem{}, ErrUnknownUnit
	}

	p, err := s.planets.Get(ctx, planetID)
	if err != nil {
		return QueueItem{}, err
	}
	if p.UserID != userID {
		return QueueItem{}, ErrPlanetOwnership
	}

	var item QueueItem
	err = s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// repair_factory >= 1.
		repairSpec, hasRepair := s.catalog.Buildings.Buildings["repair_factory"]
		if !hasRepair {
			return ErrNoRepairBuilding
		}
		var repairLvl int
		err := tx.QueryRow(ctx,
			`SELECT level FROM buildings WHERE planet_id=$1 AND unit_id=$2`,
			planetID, repairSpec.ID,
		).Scan(&repairLvl)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("repair_factory level: %w", err)
		}
		if repairLvl < 1 {
			return ErrNoRepairBuilding
		}

		// Берём текущий damaged_count + shell_percent из ships FOR UPDATE.
		var damaged int64
		var shellPct float64
		err = tx.QueryRow(ctx, `
			SELECT damaged_count, shell_percent
			FROM ships WHERE planet_id=$1 AND unit_id=$2
			FOR UPDATE
		`, planetID, unitID).Scan(&damaged, &shellPct)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNothingToRepair
			}
			return fmt.Errorf("read ships: %w", err)
		}
		if damaged <= 0 {
			return ErrNothingToRepair
		}
		if shellPct < 0 {
			shellPct = 0
		}
		if shellPct > 100 {
			shellPct = 100
		}

		// Формула стоимости ремонта.
		structScale := 0.1 * (100.0 - shellPct) / 100.0
		reqPerUnit := config.ResCost{
			Metal:    ceil10(float64(baseCost.Metal) * structScale),
			Silicon:  ceil10(float64(baseCost.Silicon) * structScale),
			Hydrogen: ceil10(float64(baseCost.Hydrogen) * structScale),
		}
		totalReq := multiplyCost(reqPerUnit, damaged)

		if int64(p.Metal) < totalReq.Metal ||
			int64(p.Silicon) < totalReq.Silicon ||
			int64(p.Hydrogen) < totalReq.Hydrogen {
			return ErrNotEnoughRes
		}

		// Списываем ресурсы (юнитов НЕ снимаем — они остаются в стоке,
		// просто с damaged_count).
		if _, err := tx.Exec(ctx, `
			UPDATE planets
			SET metal=metal-$1, silicon=silicon-$2, hydrogen=hydrogen-$3
			WHERE id=$4
		`, totalReq.Metal, totalReq.Silicon, totalReq.Hydrogen, planetID); err != nil {
			return fmt.Errorf("charge res: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO res_log (user_id, planet_id, reason, delta_metal, delta_silicon, delta_hydrogen)
			VALUES ($1, $2, 'repair_cost', $3, $4, $5)
		`, userID, planetID, -totalReq.Metal, -totalReq.Silicon, -totalReq.Hydrogen); err != nil {
			return fmt.Errorf("res_log: %w", err)
		}

		// Время. То же, что для disassemble: buildTime(base×0.1, base×0.1).
		perUnit := economy.BuildDuration(1,
			economy.Cost{
				Metal:    int64(math.Round(float64(baseCost.Metal) * 0.1)),
				Silicon:  int64(math.Round(float64(baseCost.Silicon) * 0.1)),
				Hydrogen: 0,
			},
			repairLvl, 0, s.gameSpd)
		perUnitSec := int(math.Max(1, math.Round(perUnit.Seconds())))
		totalDur := time.Duration(perUnitSec) * time.Duration(damaged) * time.Second
		start := time.Now().UTC()
		end := start.Add(totalDur)

		id := ids.New()
		if _, err := tx.Exec(ctx, `
			INSERT INTO repair_queue
				(id, planet_id, user_id, unit_id, is_defense, mode, count,
				 return_metal, return_silicon, return_hydrogen,
				 per_unit_seconds, start_at, end_at, status)
			VALUES ($1, $2, $3, $4, false, 'repair', $5, 0, 0, 0, $6, $7, $8, 'running')
		`, id, planetID, userID, unitID, damaged, perUnitSec, start, end); err != nil {
			return fmt.Errorf("insert queue: %w", err)
		}

		payload, _ := json.Marshal(map[string]any{
			"queue_id": id,
			"mode":     "repair",
		})
		if _, err := tx.Exec(ctx, `
			INSERT INTO events (id, user_id, planet_id, kind, state, fire_at, payload)
			VALUES ($1, $2, $3, $4, 'wait', $5, $6)
		`, ids.New(), userID, planetID, event.KindRepair, end, payload); err != nil {
			return fmt.Errorf("insert event: %w", err)
		}

		item = QueueItem{
			ID: id, PlanetID: planetID, UserID: userID, UnitID: unitID,
			IsDefense: false, Mode: "repair", Count: damaged,
			PerUnitSeconds: perUnitSec, StartAt: start, EndAt: end, Status: "running",
		}
		return nil
	})
	return item, err
}

// DamagedUnit — один stack кораблей с ненулевым damaged.
type DamagedUnit struct {
	UnitID       int     `json:"unit_id"`
	Count        int64   `json:"count"`
	Damaged      int64   `json:"damaged"`
	ShellPercent float64 `json:"shell_percent"`
}

// ListDamaged возвращает всех damaged-юнитов на планете (ships
// table, damaged_count > 0). Используется UI репейр-экрана, чтобы
// показать что можно чинить.
func (s *Service) ListDamaged(ctx context.Context, planetID string) ([]DamagedUnit, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT unit_id, count, damaged_count, shell_percent
		FROM ships
		WHERE planet_id = $1 AND damaged_count > 0
		ORDER BY unit_id
	`, planetID)
	if err != nil {
		return nil, fmt.Errorf("list damaged: %w", err)
	}
	defer rows.Close()
	var out []DamagedUnit
	for rows.Next() {
		var u DamagedUnit
		if err := rows.Scan(&u.UnitID, &u.Count, &u.Damaged, &u.ShellPercent); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// List возвращает активные задания на планете.
func (s *Service) List(ctx context.Context, planetID string) ([]QueueItem, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT id, planet_id, user_id, unit_id, is_defense, mode, count,
		       return_metal, return_silicon, return_hydrogen,
		       per_unit_seconds, start_at, end_at, status
		FROM repair_queue
		WHERE planet_id=$1 AND status IN ('queued','running')
		ORDER BY start_at
	`, planetID)
	if err != nil {
		return nil, fmt.Errorf("list queue: %w", err)
	}
	defer rows.Close()
	var out []QueueItem
	for rows.Next() {
		var q QueueItem
		if err := rows.Scan(&q.ID, &q.PlanetID, &q.UserID, &q.UnitID, &q.IsDefense, &q.Mode, &q.Count,
			&q.ReturnMetal, &q.ReturnSilicon, &q.ReturnHydrogen,
			&q.PerUnitSeconds, &q.StartAt, &q.EndAt, &q.Status); err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	return out, rows.Err()
}

func (s *Service) lookupUnit(unitID int) (key string, cost config.ResCost, isDefense bool, ok bool) {
	for k, spec := range s.catalog.Ships.Ships {
		if spec.ID == unitID {
			return k, spec.Cost, false, true
		}
	}
	for k, spec := range s.catalog.Defense.Defense {
		if spec.ID == unitID {
			return k, spec.Cost, true, true
		}
	}
	return "", config.ResCost{}, false, false
}

// scalePerUnit — legacy ceil(base * k / 10) * 10. Возвращает per-unit
// стоимость с округлением вверх до десятков (как в oxsar2, чтобы баланс
// не разошёлся).
func scalePerUnit(base config.ResCost, k float64) config.ResCost {
	return config.ResCost{
		Metal:    ceil10(float64(base.Metal) * k),
		Silicon:  ceil10(float64(base.Silicon) * k),
		Hydrogen: ceil10(float64(base.Hydrogen) * k),
	}
}

func multiplyCost(c config.ResCost, n int64) config.ResCost {
	return config.ResCost{
		Metal:    c.Metal * n,
		Silicon:  c.Silicon * n,
		Hydrogen: c.Hydrogen * n,
	}
}

func ceil10(v float64) int64 {
	if v <= 0 {
		return 0
	}
	return int64(math.Ceil(v/10.0) * 10)
}
