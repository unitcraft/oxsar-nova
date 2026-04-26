// Package records — рекорды сервера: топ-1 по каждой категории.
package records

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oxsar/nova/backend/internal/auth"
	"github.com/oxsar/nova/backend/internal/config"
	"github.com/oxsar/nova/backend/internal/httpx"
)

type Handler struct {
	pool *pgxpool.Pool
	cat  *config.Catalog
}

func NewHandler(pool *pgxpool.Pool, cat *config.Catalog) *Handler {
	return &Handler{pool: pool, cat: cat}
}

type record struct {
	Category   string  `json:"category"` // "building" | "research" | "ship" | "defense" | "score"
	Key        string  `json:"key"`
	UnitID     int     `json:"unit_id,omitempty"`
	HolderID   string  `json:"holder_id"`
	HolderName string  `json:"holder_name"`
	Value      float64 `json:"value"`
	MyValue    float64 `json:"my_value"`
}

// List GET /api/records — топ-1 по каждой постройке/исследованию + очки.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	ctx := r.Context()
	out := []record{}

	for key, spec := range h.cat.Buildings.Buildings {
		if rec, err := h.topBuildingOrResearch(ctx, uid, spec.ID, key, "building", "buildings"); err == nil && rec.HolderID != "" {
			out = append(out, rec)
		}
	}
	for key, spec := range h.cat.Research.Research {
		if rec, err := h.topBuildingOrResearch(ctx, uid, spec.ID, key, "research", "research"); err == nil && rec.HolderID != "" {
			out = append(out, rec)
		}
	}
	for key, spec := range h.cat.Ships.Ships {
		if rec, err := h.topUnit(ctx, uid, spec.ID, key, "ship", "ships"); err == nil && rec.HolderID != "" {
			out = append(out, rec)
		}
	}
	for key, spec := range h.cat.Defense.Defense {
		if rec, err := h.topUnit(ctx, uid, spec.ID, key, "defense", "defense"); err == nil && rec.HolderID != "" {
			out = append(out, rec)
		}
	}
	if rec, err := h.topScore(ctx, uid); err == nil && rec.HolderID != "" {
		out = append(out, rec)
	}

	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"records": out})
}

// topBuildingOrResearch: buildings — MAX(level) по planet_id, research — просто level.
// table: "buildings" | "research"; category: "building" | "research".
func (h *Handler) topBuildingOrResearch(ctx context.Context, uid string, unitID int, key, category, table string) (record, error) {
	rec := record{Category: category, UnitID: unitID, Key: key}
	var sqlTop, sqlMy string
	if table == "buildings" {
		sqlTop = `SELECT u.id, u.username, MAX(b.level)::float AS lvl
			FROM buildings b
			JOIN planets p ON p.id = b.planet_id AND p.destroyed_at IS NULL
			JOIN users u ON u.id = p.user_id AND u.deleted_at IS NULL
			WHERE b.unit_id = $1
			GROUP BY u.id, u.username
			ORDER BY lvl DESC LIMIT 1`
		sqlMy = `SELECT COALESCE(MAX(b.level), 0)::float
			FROM buildings b
			JOIN planets p ON p.id = b.planet_id
			WHERE p.user_id = $1 AND b.unit_id = $2`
	} else {
		sqlTop = `SELECT u.id, u.username, r.level::float
			FROM research r
			JOIN users u ON u.id = r.user_id AND u.deleted_at IS NULL
			WHERE r.unit_id = $1
			ORDER BY r.level DESC LIMIT 1`
		sqlMy = `SELECT COALESCE(level, 0)::float FROM research WHERE user_id = $1 AND unit_id = $2`
	}
	if err := h.pool.QueryRow(ctx, sqlTop, unitID).Scan(&rec.HolderID, &rec.HolderName, &rec.Value); err != nil {
		if err == pgx.ErrNoRows {
			return rec, nil
		}
		return rec, err
	}
	_ = h.pool.QueryRow(ctx, sqlMy, uid, unitID).Scan(&rec.MyValue)
	return rec, nil
}

// topUnit: ships/defense — MAX(SUM(count)) per user.
func (h *Handler) topUnit(ctx context.Context, uid string, unitID int, key, category, table string) (record, error) {
	rec := record{Category: category, UnitID: unitID, Key: key}
	sqlTop := `SELECT u.id, u.username, SUM(s.count)::float AS cnt
		FROM ` + table + ` s
		JOIN planets p ON p.id = s.planet_id AND p.destroyed_at IS NULL
		JOIN users u ON u.id = p.user_id AND u.deleted_at IS NULL
		WHERE s.unit_id = $1
		GROUP BY u.id, u.username
		ORDER BY cnt DESC LIMIT 1`
	if err := h.pool.QueryRow(ctx, sqlTop, unitID).Scan(&rec.HolderID, &rec.HolderName, &rec.Value); err != nil {
		if err == pgx.ErrNoRows {
			return rec, nil
		}
		return rec, err
	}
	sqlMy := `SELECT COALESCE(SUM(s.count), 0)::float
		FROM ` + table + ` s
		JOIN planets p ON p.id = s.planet_id
		WHERE p.user_id = $1 AND s.unit_id = $2`
	_ = h.pool.QueryRow(ctx, sqlMy, uid, unitID).Scan(&rec.MyValue)
	return rec, nil
}

func (h *Handler) topScore(ctx context.Context, uid string) (record, error) {
	rec := record{Category: "score", Key: "total"}
	err := h.pool.QueryRow(ctx,
		`SELECT id, username, points FROM users
		 WHERE deleted_at IS NULL AND umode = false
		 ORDER BY points DESC LIMIT 1`,
	).Scan(&rec.HolderID, &rec.HolderName, &rec.Value)
	if err != nil {
		if err == pgx.ErrNoRows {
			return rec, nil
		}
		return rec, err
	}
	_ = h.pool.QueryRow(ctx, `SELECT COALESCE(points, 0) FROM users WHERE id = $1`, uid).Scan(&rec.MyValue)
	return rec, nil
}
