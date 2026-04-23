// Package search — глобальный поиск игроков, альянсов, планет.
package search

import (
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oxsar/nova/backend/internal/auth"
	"github.com/oxsar/nova/backend/internal/httpx"
)

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler { return &Handler{pool: pool} }

type playerResult struct {
	UserID      string  `json:"user_id"`
	Username    string  `json:"username"`
	AllianceTag *string `json:"alliance_tag,omitempty"`
	Points      float64 `json:"points"`
}

type allianceResult struct {
	Tag     string  `json:"tag"`
	Name    string  `json:"name"`
	Members int     `json:"members"`
	Points  float64 `json:"points"`
}

type planetResult struct {
	PlanetID string `json:"planet_id"`
	Name     string `json:"name"`
	Galaxy   int    `json:"galaxy"`
	System   int    `json:"system"`
	Position int    `json:"position"`
	Owner    string `json:"owner"`
}

type response struct {
	Players   []playerResult   `json:"players"`
	Alliances []allianceResult `json:"alliances"`
	Planets   []planetResult   `json:"planets"`
}

// Search GET /api/search?q=...&type=player|alliance|planet (type опционально — иначе все).
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserID(r.Context()); !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(q) < 2 {
		httpx.WriteJSON(w, r, http.StatusOK, response{
			Players: []playerResult{}, Alliances: []allianceResult{}, Planets: []planetResult{},
		})
		return
	}
	kind := r.URL.Query().Get("type")
	pattern := "%" + strings.ToLower(q) + "%"

	resp := response{
		Players: []playerResult{}, Alliances: []allianceResult{}, Planets: []planetResult{},
	}

	if kind == "" || kind == "player" {
		rows, err := h.pool.Query(r.Context(), `
			SELECT u.id, u.username, a.tag, u.points
			FROM users u
			LEFT JOIN alliances a ON a.id = u.alliance_id
			WHERE lower(u.username::text) LIKE $1
			ORDER BY u.points DESC
			LIMIT 20
		`, pattern)
		if err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			return
		}
		for rows.Next() {
			var p playerResult
			if err := rows.Scan(&p.UserID, &p.Username, &p.AllianceTag, &p.Points); err != nil {
				rows.Close()
				httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
				return
			}
			resp.Players = append(resp.Players, p)
		}
		rows.Close()
	}

	if kind == "" || kind == "alliance" {
		rows, err := h.pool.Query(r.Context(), `
			SELECT a.tag, a.name, COUNT(u.id), COALESCE(SUM(u.points), 0)
			FROM alliances a
			LEFT JOIN users u ON u.alliance_id = a.id
			WHERE lower(a.tag) LIKE $1 OR lower(a.name) LIKE $1
			GROUP BY a.id, a.tag, a.name
			ORDER BY COALESCE(SUM(u.points), 0) DESC
			LIMIT 20
		`, pattern)
		if err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			return
		}
		for rows.Next() {
			var a allianceResult
			if err := rows.Scan(&a.Tag, &a.Name, &a.Members, &a.Points); err != nil {
				rows.Close()
				httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
				return
			}
			resp.Alliances = append(resp.Alliances, a)
		}
		rows.Close()
	}

	if kind == "" || kind == "planet" {
		rows, err := h.pool.Query(r.Context(), `
			SELECT p.id, p.name, p.galaxy, p.system, p.position, COALESCE(u.username::text, '')
			FROM planets p
			LEFT JOIN users u ON u.id = p.user_id
			WHERE lower(p.name) LIKE $1 AND p.destroyed_at IS NULL
			ORDER BY p.name
			LIMIT 20
		`, pattern)
		if err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			return
		}
		for rows.Next() {
			var p planetResult
			if err := rows.Scan(&p.PlanetID, &p.Name, &p.Galaxy, &p.System, &p.Position, &p.Owner); err != nil {
				rows.Close()
				httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
				return
			}
			resp.Planets = append(resp.Planets, p)
		}
		rows.Close()
	}

	httpx.WriteJSON(w, r, http.StatusOK, resp)
}
