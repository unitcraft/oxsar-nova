// Package galaxyevent — глобальные галактические события (план 17 F).
//
// MVP: одно событие за раз, создаётся админом. Эффект применяется в
// расчётах (production, market) через CurrentEffect.
//
// Тип события — строка `kind`:
//   'meteor_storm' — +30% metal production у всех (params: {"metal_mult":1.3})
//   'solar_flare'  — -20% energy (не реализовано в MVP)
//   'trade_forum'  — иной рыночный курс
//   'star_nebula'  — +15% к exp_power
package galaxyevent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Event — галактическое событие.
type Event struct {
	ID        int64           `json:"id"`
	Kind      string          `json:"kind"`
	StartedAt time.Time       `json:"started_at"`
	EndsAt    time.Time       `json:"ends_at"`
	Params    json.RawMessage `json:"params"`
}

var ErrNoActive = errors.New("galaxy_event: no active event")

// Service.
type Service struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// Active возвращает активное событие (ends_at > now()) или ErrNoActive.
func (s *Service) Active(ctx context.Context) (*Event, error) {
	var e Event
	err := s.db.QueryRow(ctx, `
		SELECT id, kind, started_at, ends_at, params
		FROM galaxy_events
		WHERE ends_at > now()
		ORDER BY started_at DESC
		LIMIT 1
	`).Scan(&e.ID, &e.Kind, &e.StartedAt, &e.EndsAt, &e.Params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoActive
		}
		return nil, fmt.Errorf("galaxy_event active: %w", err)
	}
	return &e, nil
}

// MetalMultiplier — какой множитель применить к metal production
// сейчас. По умолчанию 1.0. Используется в planet.applyTickInTx.
//
// Берёт активное событие kind='meteor_storm' и читает params.metal_mult.
// Если события нет или kind другой — 1.0.
func (s *Service) MetalMultiplier(ctx context.Context) float64 {
	e, err := s.Active(ctx)
	if err != nil || e.Kind != "meteor_storm" {
		return 1.0
	}
	var p struct {
		MetalMult float64 `json:"metal_mult"`
	}
	if err := json.Unmarshal(e.Params, &p); err != nil || p.MetalMult <= 0 {
		return 1.0
	}
	return p.MetalMult
}

// Create — создаёт событие. Используется админ-API.
func (s *Service) Create(ctx context.Context, kind string, durationHours int, params map[string]any) (*Event, error) {
	if kind == "" {
		return nil, errors.New("galaxy_event: kind required")
	}
	if durationHours <= 0 || durationHours > 168 {
		return nil, errors.New("galaxy_event: duration must be 1..168 hours")
	}
	if params == nil {
		params = map[string]any{}
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("galaxy_event marshal: %w", err)
	}
	var e Event
	err = s.db.QueryRow(ctx, `
		INSERT INTO galaxy_events (kind, ends_at, params)
		VALUES ($1, now() + ($2 * interval '1 hour'), $3)
		RETURNING id, kind, started_at, ends_at, params
	`, kind, durationHours, paramsJSON).Scan(&e.ID, &e.Kind, &e.StartedAt, &e.EndsAt, &e.Params)
	if err != nil {
		return nil, fmt.Errorf("galaxy_event create: %w", err)
	}
	return &e, nil
}

// Cancel — преждевременно завершить событие (admin).
func (s *Service) Cancel(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE galaxy_events SET ends_at = now() WHERE id = $1 AND ends_at > now()`,
		id)
	if err != nil {
		return fmt.Errorf("galaxy_event cancel: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNoActive
	}
	return nil
}
