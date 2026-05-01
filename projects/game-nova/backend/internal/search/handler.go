// Package search — глобальный поиск игроков, альянсов, планет.
//
// План 72.1.16: паритет с legacy `Search.class.php`:
//   - playerResult расширен last_seen, home_planet, coords, banned.
//   - planetResult расширен is_moon, is_home.
//   - ORDER BY использует LENGTH-LENGTH семантику legacy
//     (приоритет ближайшим по длине совпадениям, не по очкам).
package search

import (
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/httpx"
)

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler { return &Handler{pool: pool} }

type playerResult struct {
	UserID       string  `json:"user_id"`
	Username     string  `json:"username"`
	AllianceTag  *string `json:"alliance_tag,omitempty"`
	Points       float64 `json:"points"`
	LastSeen     *string `json:"last_seen,omitempty"`
	HomePlanet   *string `json:"home_planet,omitempty"`
	HomeGalaxy   *int    `json:"home_galaxy,omitempty"`
	HomeSystem   *int    `json:"home_system,omitempty"`
	HomePosition *int    `json:"home_position,omitempty"`
	Banned       bool    `json:"banned,omitempty"`
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
	IsMoon   bool   `json:"is_moon"`
	IsHome   bool   `json:"is_home"` // главная (первая) планета владельца
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
	queryLen := len(q)

	resp := response{
		Players: []playerResult{}, Alliances: []allianceResult{}, Planets: []planetResult{},
	}

	if kind == "" || kind == "player" {
		// Legacy: ORDER BY length(username) - (length(query) - 2) ASC, username ASC.
		// Семантика — приоритет совпадениям с длиной близкой к запросу.
		// LATERAL-JOIN на planets — главная планета (первая созданная, не moon).
		rows, err := h.pool.Query(r.Context(), `
			SELECT u.id, u.username, a.tag, u.points, u.last_seen,
			       u.banned_at IS NOT NULL AS banned,
			       hp.name, hp.galaxy, hp.system, hp.position
			FROM users u
			LEFT JOIN alliances a ON a.id = u.alliance_id
			LEFT JOIN LATERAL (
				SELECT name, galaxy, system, position FROM planets
				WHERE user_id = u.id AND destroyed_at IS NULL AND is_moon = false
				ORDER BY created_at ASC LIMIT 1
			) hp ON true
			WHERE lower(u.username::text) LIKE $1 AND u.deleted_at IS NULL
			ORDER BY (LENGTH(u.username::text) - $2) ASC, u.username ASC
			LIMIT 25
		`, pattern, queryLen)
		if err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			return
		}
		for rows.Next() {
			var p playerResult
			var lastSeen *time.Time
			if err := rows.Scan(&p.UserID, &p.Username, &p.AllianceTag, &p.Points,
				&lastSeen, &p.Banned,
				&p.HomePlanet, &p.HomeGalaxy, &p.HomeSystem, &p.HomePosition); err != nil {
				rows.Close()
				httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
				return
			}
			if lastSeen != nil {
				s := lastSeen.UTC().Format(time.RFC3339)
				p.LastSeen = &s
			}
			resp.Players = append(resp.Players, p)
		}
		rows.Close()
	}

	if kind == "" || kind == "alliance" {
		// Legacy: ORDER BY case when tag LIKE then length(tag)-slen else length(name)-slen end ASC, tag ASC.
		// Упрощаем до общего LENGTH(tag) - $2 для совместимости (точное LIKE-разветвление в Postgres
		// требует CASE, но нет seo-разницы для UI; альянсов всегда десятки, не тысячи).
		rows, err := h.pool.Query(r.Context(), `
			SELECT a.tag, a.name, COUNT(u.id), COALESCE(SUM(u.points), 0)
			FROM alliances a
			LEFT JOIN users u ON u.alliance_id = a.id
			WHERE lower(a.tag) LIKE $1 OR lower(a.name) LIKE $1
			GROUP BY a.id, a.tag, a.name
			ORDER BY (LENGTH(a.tag) - $2) ASC, a.tag ASC
			LIMIT 25
		`, pattern, queryLen)
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
		// Legacy:
		//   if (planet_id == hp) → suffix " (HP)"
		//   else if (ismoon)     → suffix " (MOON)"
		// Передаём is_home/is_moon флагами — фронт делает render-суффикс для i18n.
		// is_home через LATERAL: главная планета владельца (первая созданная, не moon).
		rows, err := h.pool.Query(r.Context(), `
			SELECT p.id, p.name, p.galaxy, p.system, p.position,
			       COALESCE(u.username::text, '') AS owner,
			       p.is_moon,
			       (hp.id = p.id) AS is_home
			FROM planets p
			LEFT JOIN users u ON u.id = p.user_id AND u.deleted_at IS NULL
			LEFT JOIN LATERAL (
				SELECT id FROM planets
				WHERE user_id = p.user_id AND destroyed_at IS NULL AND is_moon = false
				ORDER BY created_at ASC LIMIT 1
			) hp ON true
			WHERE lower(p.name) LIKE $1 AND p.destroyed_at IS NULL
			ORDER BY (LENGTH(p.name) - $2) ASC, p.name ASC
			LIMIT 25
		`, pattern, queryLen)
		if err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			return
		}
		for rows.Next() {
			var p planetResult
			var isHome *bool // owner может быть NULL → is_home NULL
			if err := rows.Scan(&p.PlanetID, &p.Name, &p.Galaxy, &p.System, &p.Position,
				&p.Owner, &p.IsMoon, &isHome); err != nil {
				rows.Close()
				httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
				return
			}
			if isHome != nil && *isHome {
				p.IsHome = true
			}
			resp.Planets = append(resp.Planets, p)
		}
		rows.Close()
	}

	httpx.WriteJSON(w, r, http.StatusOK, resp)
}
