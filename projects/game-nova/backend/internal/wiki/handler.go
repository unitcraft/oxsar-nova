package wiki

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"oxsar/game-nova/internal/httpx"
)

// Handler — HTTP-адаптер.
//
//	GET /api/wiki                   — список категорий (index первого уровня)
//	GET /api/wiki/{category}        — список страниц категории
//	GET /api/wiki/{category}/{slug} — одна страница с markdown
//
// Публичный (не требует auth) — см. план 19 Ф.6: SEO и доступ без логина.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Index GET /api/wiki
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	cats, err := h.svc.List()
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	if cats == nil {
		cats = []Category{}
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"categories": cats})
}

// Category GET /api/wiki/{category}
func (h *Handler) Category(w http.ResponseWriter, r *http.Request) {
	cat := chi.URLParam(r, "category")
	if cat == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "category required"))
		return
	}
	pages, err := h.svc.ListCategory(cat)
	switch {
	case err == nil:
		if pages == nil {
			pages = []Page{}
		}
		httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"category": cat, "pages": pages})
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrBadPath):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// Page GET /api/wiki/{category}/{slug}
func (h *Handler) Page(w http.ResponseWriter, r *http.Request) {
	cat := chi.URLParam(r, "category")
	slug := chi.URLParam(r, "slug")
	if cat == "" || slug == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "category and slug required"))
		return
	}
	path := strings.Trim(cat, "/") + "/" + strings.Trim(slug, "/")
	p, err := h.svc.Get(path)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusOK, p)
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrBadPath):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}
