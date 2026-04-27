package authsvc

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"oxsar/identity/internal/httpx"
	"oxsar/identity/pkg/jwtrs"
	"github.com/redis/go-redis/v9"
)

// Handler — HTTP-адаптер Auth Service.
type Handler struct {
	svc       *Service
	iss       *jwtrs.Issuer
	ver       *jwtrs.Verifier
	rdb       *redis.Client
	blacklist *JTIBlacklist
}

// NewHandler создаёт Handler.
func NewHandler(svc *Service, iss *jwtrs.Issuer, ver *jwtrs.Verifier, rdb *redis.Client) *Handler {
	return &Handler{
		svc:       svc,
		iss:       iss,
		ver:       ver,
		rdb:       rdb,
		blacklist: NewJTIBlacklist(rdb),
	}
}

// Register — POST /auth/register
//
// План 44 (152-ФЗ): consent_accepted обязателен; без него регистрация
// возвращает 400. IP и User-Agent сохраняются в user_consents.
// План 47: terms_accepted обязателен — акцепт Договора-оферты, Правил
// игры и Политики возврата.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Username        string `json:"username"`
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConsentAccepted bool   `json:"consent_accepted"`
		TermsAccepted   bool   `json:"terms_accepted"`
	}
	if err := decodeJSON(r, &in); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	u, toks, err := h.svc.Register(r.Context(), RegisterInput{
		Username:         in.Username,
		Email:            in.Email,
		Password:         in.Password,
		ConsentAccepted:  in.ConsentAccepted,
		TermsAccepted:    in.TermsAccepted,
		ConsentIP:        clientIP(r),
		ConsentUserAgent: r.UserAgent(),
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrUserExists):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, "user already exists"))
		case errors.Is(err, ErrConsentRequired):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "consent to personal data processing is required"))
		case errors.Is(err, ErrTermsRequired):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "acceptance of offer, game rules and refund policy is required"))
		case errors.Is(err, ErrUsernameForbidden):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "имя содержит запрещённое слово"))
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
	// План 36 Critical-4: проверяем blacklist отозванных refresh-токенов.
	claims, err := h.ver.Parse(in.Refresh, "refresh")
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrUnauthorized, "invalid refresh token"))
		return
	}
	if revoked, _ := h.blacklist.IsRevoked(r.Context(), claims.ID); revoked {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrUnauthorized, "refresh token revoked"))
		return
	}
	toks, err := h.svc.Refresh(r.Context(), in.Refresh)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrUnauthorized, "invalid refresh token"))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"tokens": toks})
}

// Logout — POST /auth/logout. Принимает refresh-token, кладёт его jti
// в Redis-blacklist на оставшийся TTL.
//
// Публичный endpoint (не требует access JWT) — на logout фронт мог уже
// потерять access. Достаточно знать refresh, чтобы его отозвать.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Refresh string `json:"refresh"`
	}
	if err := decodeJSON(r, &in); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	claims, err := h.ver.Parse(in.Refresh, "refresh")
	if err != nil {
		// Идемпотентность: невалидный токен → 204 (logout всё равно «работает»).
		w.WriteHeader(http.StatusNoContent)
		return
	}
	expiresAt := time.Time{}
	if claims.ExpiresAt != nil {
		expiresAt = claims.ExpiresAt.Time
	}
	if err := h.blacklist.Revoke(r.Context(), claims.ID, expiresAt); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
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

// DeleteMe — DELETE /auth/users/me (требует JWT).
//
// План 44 (152-ФЗ, ст. 14): право субъекта на удаление своих ПДн.
// Анонимизирует учётную запись в auth-service. Игровые сервисы должны
// иметь свой flow удаления связанных игровых данных (в game-nova —
// /api/me/deletion/code → DELETE /api/me).
func (h *Handler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromCtx(r)
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	if err := h.svc.DeleteAccount(r.Context(), userID); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ChangePassword — POST /auth/password (требует JWT).
// План 36 Critical-6: смена пароля переехала из game-nova/settings в auth-service,
// потому что хеш пароля живёт здесь.
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromCtx(r)
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	var in struct {
		Current string `json:"current"`
		New     string `json:"new"`
	}
	if err := decodeJSON(r, &in); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	err := h.svc.ChangePassword(r.Context(), userID, in.Current, in.New)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredential):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "current password is incorrect"))
		default:
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// JWKS — GET /.well-known/jwks.json
func (h *Handler) JWKS(w http.ResponseWriter, r *http.Request) {
	jwks := jwtrs.IssuerToJWKS(h.iss)
	httpx.WriteJSON(w, r, http.StatusOK, jwks)
}

// План 38 Ф.5: CreditBalance/CreditHistory/SpendCredits handlers удалены.
// Кошельки и платежи живут в billing-service:
//   GET  /billing/wallet/balance
//   GET  /billing/wallet/history
//   POST /billing/wallet/spend (с Idempotency-Key)

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
