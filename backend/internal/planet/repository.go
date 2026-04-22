package planet

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository — CRUD-доступ к planets и связанным таблицам.
// Намеренно сделано тонким: бизнес-логика живёт в Service.
type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (r *Repository) GetByID(ctx context.Context, id string) (Planet, error) {
	var p Planet
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, is_moon, name, galaxy, system, position,
		       diameter, used_fields, planet_type, temperature_min, temperature_max,
		       metal, silicon, hydrogen, last_res_update,
		       solar_satellite_prod, build_factor, research_factor,
		       produce_factor, energy_factor, storage_factor
		FROM planets WHERE id = $1 AND destroyed_at IS NULL
	`, id).Scan(
		&p.ID, &p.UserID, &p.IsMoon, &p.Name, &p.Galaxy, &p.System, &p.Position,
		&p.Diameter, &p.UsedFields, &p.PlanetType, &p.TempMin, &p.TempMax,
		&p.Metal, &p.Silicon, &p.Hydrogen, &p.LastResUpdate,
		&p.SolarSatelliteProd, &p.BuildFactor, &p.ResearchFactor,
		&p.ProduceFactor, &p.EnergyFactor, &p.StorageFactor,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Planet{}, ErrNotFound
		}
		return Planet{}, fmt.Errorf("select planet: %w", err)
	}
	return p, nil
}

func (r *Repository) ListByUser(ctx context.Context, userID string) ([]Planet, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, is_moon, name, galaxy, system, position,
		       diameter, used_fields, planet_type, temperature_min, temperature_max,
		       metal, silicon, hydrogen, last_res_update,
		       solar_satellite_prod, build_factor, research_factor,
		       produce_factor, energy_factor, storage_factor
		FROM planets WHERE user_id = $1 AND destroyed_at IS NULL
		ORDER BY is_moon, galaxy, system, position
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list planets: %w", err)
	}
	defer rows.Close()

	var out []Planet
	for rows.Next() {
		var p Planet
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.IsMoon, &p.Name, &p.Galaxy, &p.System, &p.Position,
			&p.Diameter, &p.UsedFields, &p.PlanetType, &p.TempMin, &p.TempMax,
			&p.Metal, &p.Silicon, &p.Hydrogen, &p.LastResUpdate,
			&p.SolarSatelliteProd, &p.BuildFactor, &p.ResearchFactor,
			&p.ProduceFactor, &p.EnergyFactor, &p.StorageFactor,
		); err != nil {
			return nil, fmt.Errorf("scan planet: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// Rename обновляет имя планеты.
func (r *Repository) Rename(ctx context.Context, planetID string, name string) error {
	res, err := r.pool.Exec(ctx, `
		UPDATE planets SET name = $1 WHERE id = $2 AND destroyed_at IS NULL
	`, name, planetID)
	if err != nil {
		return fmt.Errorf("rename planet: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// SetHome устанавливает эту планету домашней для юзера (обновляет users.cur_planet_id).
func (r *Repository) SetHome(ctx context.Context, userID, planetID string) error {
	res, err := r.pool.Exec(ctx, `
		UPDATE users SET cur_planet_id = $1 WHERE id = $2
	`, planetID, userID)
	if err != nil {
		return fmt.Errorf("set home planet: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Abandon мягко удаляет планету (soft-delete через destroyed_at).
func (r *Repository) Abandon(ctx context.Context, planetID string) error {
	res, err := r.pool.Exec(ctx, `
		UPDATE planets SET destroyed_at = now() WHERE id = $1 AND destroyed_at IS NULL
	`, planetID)
	if err != nil {
		return fmt.Errorf("abandon planet: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Ошибки операций с планетами.
var (
	ErrNotFound            = errors.New("planet: not found")
	ErrInvalidInput        = errors.New("planet: invalid input")
	ErrMoonRestricted      = errors.New("planet: operation not allowed for moons")
	ErrOnlyPlanet          = errors.New("planet: cannot abandon only planet")
	ErrCannotAbandonHome   = errors.New("planet: cannot abandon home planet")
)
