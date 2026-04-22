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

// ErrNotFound — планета отсутствует или уничтожена.
var ErrNotFound = errors.New("planet: not found")
