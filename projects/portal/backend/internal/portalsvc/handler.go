package portalsvc

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"oxsar/portal/internal/httpx"
	"oxsar/portal/internal/universe"
)

// Handler — HTTP-адаптер портала.
type Handler struct {
	svc     *Service
	reg     *universe.Registry
	credits *CreditClient
}

// NewHandler создаёт Handler.
func NewHandler(svc *Service, reg *universe.Registry) *Handler {
	return &Handler{svc: svc, reg: reg, credits: NewCreditClient("")}
}

// NewHandlerWithCredits создаёт Handler с клиентом кредитов.
func NewHandlerWithCredits(svc *Service, reg *universe.Registry, authServiceURL string) *Handler {
	return &Handler{svc: svc, reg: reg, credits: NewCreditClient(authServiceURL)}
}

// --- universes ---

// ListUniverses — GET /api/universes
func (h *Handler) ListUniverses(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"universes": h.reg.All()})
}

// --- news ---

// ListNews — GET /api/news?limit=20&offset=0
func (h *Handler) ListNews(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, err := h.svc.ListNews(r.Context(), limit, offset)
	if err != nil {
		httpx.WriteError(w, r, httpx.ErrInternal)
		return
	}
	if items == nil {
		items = []NewsItem{}
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"news": items})
}

// GetNews — GET /api/news/{id}
func (h *Handler) GetNews(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	item, err := h.svc.GetNews(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		httpx.WriteError(w, r, httpx.ErrNotFound)
		return
	}
	if err != nil {
		httpx.WriteError(w, r, httpx.ErrInternal)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, item)
}

// CreateNews — POST /api/news (admin only)
func (h *Handler) CreateNews(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromCtx(r)
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	var in struct {
		Title     string `json:"title"`
		Body      string `json:"body"`
		Published bool   `json:"published"`
	}
	if err := decodeJSON(r, &in); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	item, err := h.svc.CreateNews(r.Context(), userID, in.Title, in.Body, in.Published)
	if err != nil {
		httpx.WriteError(w, r, httpx.ErrInternal)
		return
	}
	httpx.WriteJSON(w, r, http.StatusCreated, item)
}

// --- feedback ---

// ListFeedback — GET /api/feedback?status=approved&limit=20&offset=0
func (h *Handler) ListFeedback(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	posts, err := h.svc.ListFeedback(r.Context(), status, limit, offset)
	if err != nil {
		httpx.WriteError(w, r, httpx.ErrInternal)
		return
	}
	if posts == nil {
		posts = []FeedbackPost{}
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"posts": posts})
}

// GetFeedback — GET /api/feedback/{id}
func (h *Handler) GetFeedback(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	post, err := h.svc.GetFeedback(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		httpx.WriteError(w, r, httpx.ErrNotFound)
		return
	}
	if err != nil {
		httpx.WriteError(w, r, httpx.ErrInternal)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, post)
}

// CreateFeedback — POST /api/feedback
func (h *Handler) CreateFeedback(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromCtx(r)
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	authorName, _ := authorNameFromCtx(r)
	var in struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if err := decodeJSON(r, &in); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	post, err := h.svc.CreateFeedback(r.Context(), userID, authorName, in.Title, in.Body)
	if err != nil {
		switch {
		case errors.Is(err, ErrTooManyProposals):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, err.Error()))
		default:
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		}
		return
	}
	httpx.WriteJSON(w, r, http.StatusCreated, post)
}

// ModerateFeedback — PATCH /api/feedback/{id}/status (admin only)
func (h *Handler) ModerateFeedback(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var in struct {
		Status string `json:"status"`
	}
	if err := decodeJSON(r, &in); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	if err := h.svc.ModerateFeedback(r.Context(), id, in.Status); err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteError(w, r, httpx.ErrNotFound)
			return
		}
		httpx.WriteError(w, r, httpx.ErrInternal)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// VoteFeedback — POST /api/feedback/{id}/vote
// Списывает 100 кредитов через auth-service и добавляет голос.
func (h *Handler) VoteFeedback(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromCtx(r)
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	postID := chi.URLParam(r, "id")

	// Проверяем, что предложение существует и одобрено
	post, err := h.svc.GetFeedback(r.Context(), postID)
	if errors.Is(err, ErrNotFound) {
		httpx.WriteError(w, r, httpx.ErrNotFound)
		return
	}
	if err != nil || post.Status != "approved" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "post not available for voting"))
		return
	}

	// Списываем 100 кредитов через auth-service, затем записываем голос.
	const voteCost = int64(100)
	if err := h.credits.Spend(r.Context(), userID, voteCost, "feedback_vote", postID); err != nil {
		if errors.Is(err, ErrInsufficientCredits) {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, "insufficient credits (need 100)"))
			return
		}
		httpx.WriteError(w, r, httpx.ErrInternal)
		return
	}
	if err := h.svc.VoteFeedback(r.Context(), postID, userID, voteCost); err != nil {
		httpx.WriteError(w, r, httpx.ErrInternal)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"vote_count": post.VoteCount + 1})
}

// ListComments — GET /api/feedback/{id}/comments
func (h *Handler) ListComments(w http.ResponseWriter, r *http.Request) {
	postID := chi.URLParam(r, "id")
	comments, err := h.svc.ListComments(r.Context(), postID)
	if err != nil {
		httpx.WriteError(w, r, httpx.ErrInternal)
		return
	}
	if comments == nil {
		comments = []FeedbackComment{}
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"comments": comments})
}

// AddComment — POST /api/feedback/{id}/comments
func (h *Handler) AddComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromCtx(r)
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	authorName, _ := authorNameFromCtx(r)
	postID := chi.URLParam(r, "id")
	var in struct {
		ParentID *string `json:"parent_id,omitempty"`
		Body     string  `json:"body"`
	}
	if err := decodeJSON(r, &in); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	comment, err := h.svc.AddComment(r.Context(), postID, in.ParentID, userID, authorName, in.Body)
	if err != nil {
		httpx.WriteError(w, r, httpx.ErrInternal)
		return
	}
	httpx.WriteJSON(w, r, http.StatusCreated, comment)
}

func decodeJSON(r *http.Request, into any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	return dec.Decode(into)
}
