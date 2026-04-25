package features

import (
	"encoding/json"
	"net/http"
)

// Handler — HTTP-адаптер. Public endpoint, без auth: фронтенд читает
// при загрузке, чтобы решать, какой UI показывать.
type Handler struct {
	set *Set
}

func NewHandler(s *Set) *Handler { return &Handler{set: s} }

// List GET /api/features — возвращает все флаги (для UI и debug).
//
//	{
//	  "features": {
//	    "goal_engine":     {"enabled": true,  "description": "..."},
//	    "experimental":    {"enabled": false, "description": "..."}
//	  },
//	  "enabled": ["goal_engine"]
//	}
//
// `enabled` — отсортированный список ключей с enabled=true (для
// быстрых проверок на фронте без перебора map).
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	out := struct {
		Features map[string]Flag `json:"features"`
		Enabled  []string        `json:"enabled"`
	}{
		Features: All(h.set),
		Enabled:  EnabledKeys(h.set),
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}
