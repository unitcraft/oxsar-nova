package authsvc

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/oxsar/nova/backend/internal/httpx"
	"github.com/oxsar/nova/backend/pkg/jwtrs"
	"github.com/redis/go-redis/v9"
)

// Handler — HTTP-адаптер Auth Service.
type Handler struct {
	svc     *Service
	iss     *jwtrs.Issuer
	rdb     *redis.Client
}

// NewHandler создаёт Handler.
func NewHandler(svc *Service, iss *jwtrs.Issuer, rdb *redis.Client) *Handler {
	return &Handler{svc: svc, iss: iss, rdb: rdb}
}

// Register — POST /auth/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &in); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	u, toks, err := h.svc.Register(r.Context(), RegisterInput{
		Username: in.Username,
		Email:    in.Email,
		Password: in.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrUserExists):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, "user already exists"))
		default:
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		}
		return
	}
	httpx.WriteJSON(w, r, http.StatusCreated, map[string]any{"user": u, "tokens": toks})
}

// Login — POST /auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Login    string `json:"login"`    // email или username
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &in); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	u, toks, err := h.svc.Login(r.Context(), in.Login, in.Password)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredential):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrUnauthorized, "invalid credentials"))
		case errors.Is(err, ErrUserBanned):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrForbidden, "account banned"))
		default:
			httpx.WriteError(w, r, httpx.ErrInternal)
		}
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"user": u, "tokens": toks})
}

// Refresh — POST /auth/refresh
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Refresh string `json:"refresh"`
	}
	if err := decodeJSON(r, &in); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	toks, err := h.svc.Refresh(r.Context(), in.Refresh)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrUnauthorized, "invalid refresh token"))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"tokens": toks})
}

// Me — GET /auth/me
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromCtx(r)
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	u, err := h.svc.GetUser(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, r, httpx.ErrNotFound)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, u)
}

// JWKS — GET /.well-known/jwks.json
func (h *Handler) JWKS(w http.ResponseWriter, r *http.Request) {
	jwks := jwtrs.IssuerToJWKS(h.iss)
	httpx.WriteJSON(w, r, http.StatusOK, jwks)
}

// CreditBalance — GET /auth/credits/balance
func (h *Handler) CreditBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromCtx(r)
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	balance, err := h.svc.CreditBalance(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, r, httpx.ErrInternal)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]int64{"balance": balance})
}

// CreditHistory — GET /auth/credits/history?limit=50&offset=0
func (h *Handler) CreditHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromCtx(r)
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	txs, err := h.svc.CreditHistory(r.Context(), userID, limit, offset)
	if err != nil {
		httpx.WriteError(w, r, httpx.ErrInternal)
		return
	}
	if txs == nil {
		txs = []CreditTx{}
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"transactions": txs})
}

// SpendCredits — POST /auth/credits/spend (внутренний, только между сервисами)
func (h *Handler) SpendCredits(w http.ResponseWriter, r *http.Request) {
	var in struct {
		UserID string `json:"user_id"`
		Amount int64  `json:"amount"`
		Reason string `json:"reason"`
		RefID  string `json:"ref_id"`
	}
	if err := decodeJSON(r, &in); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	if in.Amount <= 0 {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "amount must be positive"))
		return
	}
	err := h.svc.SpendCredits(r.Context(), SpendInput{
		UserID: in.UserID,
		Amount: in.Amount,
		Reason: in.Reason,
		RefID:  in.RefID,
	})
	if err != nil {
		if errors.Is(err, ErrInsufficientFunds) {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, "insufficient credits"))
			return
		}
		httpx.WriteError(w, r, httpx.ErrInternal)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UniverseToken — POST /auth/universe-token
// Создаёт одноразовый handoff-токен для переключения вселенной.
func (h *Handler) UniverseToken(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromCtx(r)
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	var in struct {
		UniverseID string `json:"universe_id"`
	}
	if err := decodeJSON(r, &in); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	if in.UniverseID == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "universe_id required"))
		return
	}

	ht := jwtrs.NewHandoffToken(userID)
	// Сохраняем в Redis: key=handoff:<token>, value=userID, TTL=30s
	key := "handoff:" + ht.Token
	if err := h.rdb.Set(r.Context(), key, userID, 30*time.Second).Err(); err != nil {
		httpx.WriteError(w, r, httpx.ErrInternal)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]string{
		"handoff_token": ht.Token,
		"universe_id":   in.UniverseID,
	})
}

// TokenExchange — POST /auth/token/exchange
// Игровой сервер обменивает handoff-токен на полноценный JWT.
func (h *Handler) TokenExchange(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Code string `json:"code"` // handoff token
	}
	if err := decodeJSON(r, &in); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}

	key := "handoff:" + in.Code
	userID, err := h.rdb.GetDel(r.Context(), key).Result()
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrUnauthorized, "invalid or expired handoff token"))
		return
	}

	u, err := h.svc.GetUser(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	toks, err := h.svc.issueTokens(r.Context(), u)
	if err != nil {
		httpx.WriteError(w, r, httpx.ErrInternal)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"user": u, "tokens": toks})
}

// RegisterUniverse — POST /auth/universes/register (внутренний, вызывается игровым сервером)
// Регистрирует membership игрока в вселенной при lazy join.
func (h *Handler) RegisterUniverse(w http.ResponseWriter, r *http.Request) {
	var in struct {
		UserID     string `json:"user_id"`
		UniverseID string `json:"universe_id"`
	}
	if err := decodeJSON(r, &in); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	_, err := h.svc.db.Pool().Exec(r.Context(), `
		INSERT INTO universe_memberships (user_id, universe_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, in.UserID, in.UniverseID)
	if err != nil {
		httpx.WriteError(w, r, httpx.ErrInternal)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func decodeJSON(r *http.Request, into any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	return dec.Decode(into)
}
