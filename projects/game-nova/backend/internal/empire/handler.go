// Package empire реализует GET /api/empire — сводная таблица всех планет игрока.
package empire

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/httpx"
)

// PlanetRow — данные одной планеты для empire-таблицы.
type PlanetRow struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Galaxy   int    `json:"galaxy"`
	System   int    `json:"system"`
	Position int    `json:"position"`
	IsMoon   bool   `json:"is_moon"`

	Diameter   int `json:"diameter"`
	UsedFields int `json:"used_fields"`
	TempMin    int `json:"temp_min"`
	TempMax    int `json:"temp_max"`

	Metal    float64 `json:"metal"`
	Silicon  float64 `json:"silicon"`
	Hydrogen float64 `json:"hydrogen"`

	Buildings map[int]int   `json:"buildings"` // unit_id → level
	Ships     map[int]int64 `json:"ships"`      // unit_id → count
	Defense   map[int]int64 `json:"defense"`    // unit_id → count
}

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler { return &Handler{pool: pool} }

// GetAll GET /api/empire — все планеты игрока с зданиями, флотом, обороной.
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}

	planets, err := h.loadPlanets(r.Context(), uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"planets": planets})
}

func (h *Handler) loadPlanets(ctx context.Context, userID string) ([]PlanetRow, error) {
	// 1. Загружаем базовые данные планет.
	rows, err := h.pool.Query(ctx, `
		SELECT id, name, galaxy, system, position, is_moon,
		       diameter, used_fields, temperature_min, temperature_max,
		       COALESCE(metal, 0), COALESCE(silicon, 0), COALESCE(hydrogen, 0)
		FROM planets
		WHERE user_id = $1 AND destroyed_at IS NULL
		ORDER BY is_moon, galaxy, system, position
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var planets []PlanetRow
	var ids []string
	for rows.Next() {
		var p PlanetRow
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Galaxy, &p.System, &p.Position, &p.IsMoon,
			&p.Diameter, &p.UsedFields, &p.TempMin, &p.TempMax,
			&p.Metal, &p.Silicon, &p.Hydrogen,
		); err != nil {
			return nil, err
		}
		p.Buildings = make(map[int]int)
		p.Ships = make(map[int]int64)
		p.Defense = make(map[int]int64)
		planets = append(planets, p)
		ids = append(ids, p.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(planets) == 0 {
		return []PlanetRow{}, nil
	}

	// Индекс для быстрого поиска по ID.
	idx := make(map[string]*PlanetRow, len(planets))
	for i := range planets {
		idx[planets[i].ID] = &planets[i]
	}

	// 2. Здания.
	bRows, err := h.pool.Query(ctx, `
		SELECT planet_id, unit_id, level FROM buildings
		WHERE planet_id = ANY($1) AND level > 0
	`, ids)
	if err != nil {
		return nil, err
	}
	defer bRows.Close()
	for bRows.Next() {
		var pid string
		var uid, level int
		if err := bRows.Scan(&pid, &uid, &level); err != nil {
			return nil, err
		}
		if p, ok := idx[pid]; ok {
			p.Buildings[uid] = level
		}
	}
	if err := bRows.Err(); err != nil {
		return nil, err
	}

	// 3. Флот.
	sRows, err := h.pool.Query(ctx, `
		SELECT planet_id, unit_id, count FROM ships
		WHERE planet_id = ANY($1) AND count > 0
	`, ids)
	if err != nil {
		return nil, err
	}
	defer sRows.Close()
	for sRows.Next() {
		var pid string
		var uid int
		var count int64
		if err := sRows.Scan(&pid, &uid, &count); err != nil {
			return nil, err
		}
		if p, ok := idx[pid]; ok {
			p.Ships[uid] = count
		}
	}
	if err := sRows.Err(); err != nil {
		return nil, err
	}

	// 4. Оборона.
	dRows, err := h.pool.Query(ctx, `
		SELECT planet_id, unit_id, count FROM defense
		WHERE planet_id = ANY($1) AND count > 0
	`, ids)
	if err != nil {
		return nil, err
	}
	defer dRows.Close()
	for dRows.Next() {
		var pid string
		var uid int
		var count int64
		if err := dRows.Scan(&pid, &uid, &count); err != nil {
			return nil, err
		}
		if p, ok := idx[pid]; ok {
			p.Defense[uid] = count
		}
	}
	if err := dRows.Err(); err != nil {
		return nil, err
	}

	return planets, nil
}
