// Package techtree — дерево требований юнитов/зданий/исследований с учётом прогресса игрока.
package techtree

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/config"
	"oxsar/game-nova/internal/httpx"
)

type Handler struct {
	pool *pgxpool.Pool
	cat  *config.Catalog
}

func NewHandler(pool *pgxpool.Pool, cat *config.Catalog) *Handler {
	return &Handler{pool: pool, cat: cat}
}

type reqDTO struct {
	Kind  string `json:"kind"`
	Key   string `json:"key"`
	Level int    `json:"level"`
	Have  int    `json:"have"`
	Met   bool   `json:"met"`
}

type nodeDTO struct {
	Key          string   `json:"key"`
	Kind         string   `json:"kind"` // building | research | ship | defense
	ID           int      `json:"id"`
	CurrentLevel int      `json:"current_level"`
	Unlocked     bool     `json:"unlocked"`
	Requirements []reqDTO `json:"requirements"`
	// План 72.1.22: legacy `Techtree.class.php` отдельная секция «moon»
	// для лунных построек. moon_only=true → frontend рендерит в отдельной
	// секции под обычными зданиями.
	MoonOnly bool `json:"moon_only,omitempty"`
}

type response struct {
	Nodes []nodeDTO `json:"nodes"`
}

// Get GET /api/techtree?planet_id=... — дерево требований с прогрессом игрока.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	planetID := r.URL.Query().Get("planet_id")
	if planetID == "" {
		// Берём максимум здания по всем планетам пользователя.
		planetID = ""
	}

	// Максимальный уровень здания пользователя (по всем планетам).
	buildingLevels := make(map[string]int)
	{
		rows, err := h.pool.Query(r.Context(), `
			SELECT b.unit_id, MAX(b.level)
			FROM buildings b
			JOIN planets p ON p.id = b.planet_id
			WHERE p.user_id = $1 AND p.destroyed_at IS NULL
			GROUP BY b.unit_id
		`, uid)
		if err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			return
		}
		for rows.Next() {
			var id, level int
			if err := rows.Scan(&id, &level); err != nil {
				rows.Close()
				httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
				return
			}
			if key := h.buildingKeyByID(id); key != "" {
				buildingLevels[key] = level
			}
		}
		rows.Close()
	}

	// Уровни исследований.
	researchLevels := make(map[string]int)
	{
		rows, err := h.pool.Query(r.Context(), `
			SELECT unit_id, level FROM research WHERE user_id = $1
		`, uid)
		if err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			return
		}
		for rows.Next() {
			var id, level int
			if err := rows.Scan(&id, &level); err != nil {
				rows.Close()
				httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
				return
			}
			if key := h.researchKeyByID(id); key != "" {
				researchLevels[key] = level
			}
		}
		rows.Close()
	}

	var nodes []nodeDTO

	// Здания.
	for key, spec := range h.cat.Buildings.Buildings {
		n := nodeDTO{
			Key:          key,
			Kind:         "building",
			ID:           spec.ID,
			CurrentLevel: buildingLevels[key],
			MoonOnly:     spec.MoonOnly,
		}
		n.Requirements, n.Unlocked = h.buildRequirements(key, buildingLevels, researchLevels)
		nodes = append(nodes, n)
	}

	// Исследования.
	for key, spec := range h.cat.Research.Research {
		n := nodeDTO{
			Key:          key,
			Kind:         "research",
			ID:           spec.ID,
			CurrentLevel: researchLevels[key],
		}
		n.Requirements, n.Unlocked = h.buildRequirements(key, buildingLevels, researchLevels)
		nodes = append(nodes, n)
	}

	// Корабли.
	for key, spec := range h.cat.Ships.Ships {
		n := nodeDTO{
			Key:  key,
			Kind: "ship",
			ID:   spec.ID,
		}
		n.Requirements, n.Unlocked = h.buildRequirements(key, buildingLevels, researchLevels)
		nodes = append(nodes, n)
	}

	// Оборона.
	for key, spec := range h.cat.Defense.Defense {
		n := nodeDTO{
			Key:  key,
			Kind: "defense",
			ID:   spec.ID,
		}
		n.Requirements, n.Unlocked = h.buildRequirements(key, buildingLevels, researchLevels)
		nodes = append(nodes, n)
	}

	httpx.WriteJSON(w, r, http.StatusOK, response{Nodes: nodes})
}

func (h *Handler) buildRequirements(key string, bLevels, rLevels map[string]int) ([]reqDTO, bool) {
	reqs, ok := h.cat.Requirements.Requirements[key]
	if !ok || len(reqs) == 0 {
		return []reqDTO{}, true
	}
	out := make([]reqDTO, 0, len(reqs))
	allMet := true
	for _, req := range reqs {
		var have int
		switch req.Kind {
		case "building":
			have = bLevels[req.Key]
		case "research":
			have = rLevels[req.Key]
		}
		met := have >= req.Level
		if !met {
			allMet = false
		}
		out = append(out, reqDTO{
			Kind: req.Kind, Key: req.Key, Level: req.Level, Have: have, Met: met,
		})
	}
	return out, allMet
}

func (h *Handler) buildingKeyByID(id int) string {
	for key, spec := range h.cat.Buildings.Buildings {
		if spec.ID == id {
			return key
		}
	}
	return ""
}

func (h *Handler) researchKeyByID(id int) string {
	for key, spec := range h.cat.Research.Research {
		if spec.ID == id {
			return key
		}
	}
	return ""
}
