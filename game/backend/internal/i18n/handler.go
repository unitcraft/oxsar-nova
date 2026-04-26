package i18n

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/oxsar/nova/backend/internal/httpx"
)

// Handler — HTTP-адаптер. Два эндпоинта:
//   GET /api/i18n                — список доступных языков
//   GET /api/i18n/{lang}         — полный словарь
type Handler struct{ b *Bundle }

func NewHandler(b *Bundle) *Handler { return &Handler{b: b} }

// Languages GET /api/i18n
func (h *Handler) Languages(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{
		"languages": h.b.Languages(),
		"fallback":  h.b.fallback,
	})
}

// Locale GET /api/i18n/{lang}
func (h *Handler) Locale(w http.ResponseWriter, r *http.Request) {
	lang := Lang(chi.URLParam(r, "lang"))
	httpx.WriteJSON(w, r, http.StatusOK, h.b.Locale(lang))
}
