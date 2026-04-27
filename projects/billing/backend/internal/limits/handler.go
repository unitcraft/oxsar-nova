package limits

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"oxsar/billing/internal/httpx"
)

// Handler — HTTP-адаптер для limits API. Публичный endpoint
// /api/billing/limits/status работает без auth; admin endpoints
// проверяют permissions из JWT-claims (вешаются billing.RequirePermission
// через middleware в main.go).
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// statusResponse — публичный ответ для frontend.
//
// Сообщение нейтральное, не раскрывает причину (план 54 §Ф.5):
// клиент не должен знать, лимит это, технический сбой или admin override.
type statusResponse struct {
	Active  bool   `json:"active"`
	Message string `json:"message,omitempty"`
}

// Status — GET /api/billing/limits/status (без auth).
//
// Кеш: используется in-process cache внутри Service (TTL 30s), поэтому
// под нагрузкой N запросов от frontend → 1 SELECT в БД. Для дополнительной
// защиты от brute-force на публичных endpoints — rate-limit middleware
// в main.go (60 req/min/IP).
func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	active, err := h.svc.IsActive(r.Context())
	resp := statusResponse{Active: active}
	if err != nil {
		// Не раскрываем internals в публичный API.
		resp.Active = false
		resp.Message = PublicMessage
		httpx.WriteJSON(w, r, http.StatusOK, resp)
		return
	}
	if !active {
		resp.Message = PublicMessage
	}
	httpx.WriteJSON(w, r, http.StatusOK, resp)
}

// adminStatusResponse — расширенный ответ для admin (видит всё).
type adminStatusResponse struct {
	Active           bool       `json:"active"`
	RevenueYTDKop    int64      `json:"revenue_ytd_kop"`
	RevenueYTDRub    int64      `json:"revenue_ytd_rub"`
	HardStopKop      int64      `json:"hard_stop_kop"`
	HardStopRub      int64      `json:"hard_stop_rub"`
	Percent          float64    `json:"percent"`
	LastChangedAt    *time.Time `json:"last_changed_at,omitempty"`
	LastChangedBy    *string    `json:"last_changed_by,omitempty"`
	LastChangeReason string     `json:"last_change_reason,omitempty"`
	AutoDisabledAt   *time.Time `json:"auto_disabled_at,omitempty"`
}

// AdminStatus — GET /api/admin/billing/limits/status.
// Permission: billing:reports:read.
func (h *Handler) AdminStatus(w http.ResponseWriter, r *http.Request) {
	if !hasPermission(r, "billing:reports:read") {
		httpx.WriteError(w, r, &httpx.Error{
			Status: http.StatusForbidden, Code: "forbidden",
			Message: "missing permission: billing:reports:read",
		})
		return
	}
	st, err := h.svc.GetState(r.Context())
	if err != nil {
		httpx.WriteError(w, r, &httpx.Error{
			Status: http.StatusInternalServerError, Code: "internal",
			Message: err.Error(),
		})
		return
	}
	resp := adminStatusResponse{
		Active:           st.Active,
		RevenueYTDKop:    st.RevenueYTDKop,
		RevenueYTDRub:    st.RevenueYTDKop / 100,
		HardStopKop:      st.HardStopKop,
		HardStopRub:      st.HardStopKop / 100,
		Percent:          st.Percent,
		LastChangedAt:    st.LastChangedAt,
		LastChangeReason: st.LastChangeReason,
		AutoDisabledAt:   st.AutoDisabledAt,
	}
	if st.LastChangedBy != nil {
		s := st.LastChangedBy.String()
		resp.LastChangedBy = &s
	}
	httpx.WriteJSON(w, r, http.StatusOK, resp)
}

// overrideRequest — body для POST /api/admin/billing/limits/override.
type overrideRequest struct {
	Action string `json:"action"`           // "enable" | "disable"
	Reason string `json:"reason"`
}

// AdminOverride — POST /api/admin/billing/limits/override.
// Permission: billing:admin:override.
//
// Body: {"action":"enable"|"disable","reason":"..."}
// Audit-запись пишется внутри SetActive (транзакционно).
func (h *Handler) AdminOverride(w http.ResponseWriter, r *http.Request) {
	if !hasPermission(r, "billing:admin:override") {
		httpx.WriteError(w, r, &httpx.Error{
			Status: http.StatusForbidden, Code: "forbidden",
			Message: "missing permission: billing:admin:override",
		})
		return
	}
	var req overrideRequest
	if err := decodeJSON(r, &req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	if req.Reason == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "reason is required"))
		return
	}
	var active bool
	switch req.Action {
	case "enable":
		active = true
	case "disable":
		active = false
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "action must be 'enable' or 'disable'"))
		return
	}
	actorID, ok := actorIDFromContext(r)
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	ip := remoteIPString(r)
	err := h.svc.SetActive(r.Context(), active, actorID, req.Reason, ip, r.UserAgent())
	if err != nil {
		if errors.Is(err, errReasonRequired()) {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "reason is required"))
			return
		}
		httpx.WriteError(w, r, &httpx.Error{
			Status: http.StatusInternalServerError, Code: "internal",
			Message: err.Error(),
		})
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]string{"status": "ok"})
}

// revenueResponse — для GET /api/admin/billing/reports/revenue.
type revenueBucket struct {
	Period           string  `json:"period"`             // RFC3339-date or year-month
	TotalKop         int64   `json:"total_kop"`
	TotalRub         int64   `json:"total_rub"`
	CountPurchases   int64   `json:"count_purchases"`
	AvgPurchaseRub   float64 `json:"avg_purchase_rub"`
}

// AdminRevenue — GET /api/admin/billing/reports/revenue?from=&to=&granularity=
// Permission: billing:reports:read.
//
// granularity: "day" | "month" (default month). from/to — RFC3339 date.
func (h *Handler) AdminRevenue(w http.ResponseWriter, r *http.Request) {
	if !hasPermission(r, "billing:reports:read") {
		httpx.WriteError(w, r, &httpx.Error{
			Status: http.StatusForbidden, Code: "forbidden",
			Message: "missing permission: billing:reports:read",
		})
		return
	}
	gran := r.URL.Query().Get("granularity")
	if gran != "day" && gran != "month" {
		gran = "month"
	}
	from, to, err := parseFromTo(r, h.svc.cfg.Timezone)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	rows, err := h.svc.QueryRevenueBuckets(r.Context(), from, to, gran)
	if err != nil {
		httpx.WriteError(w, r, &httpx.Error{
			Status: http.StatusInternalServerError, Code: "internal",
			Message: err.Error(),
		})
		return
	}
	out := make([]revenueBucket, 0, len(rows))
	for _, b := range rows {
		out = append(out, revenueBucket{
			Period:         b.Period,
			TotalKop:       b.TotalKop,
			TotalRub:       b.TotalKop / 100,
			CountPurchases: b.Count,
			AvgPurchaseRub: b.AvgRub,
		})
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"buckets": out})
}

// AdminCSVExport — GET /api/admin/billing/reports/csv?from=&to=
// Permission: billing:reports:csv_export.
//
// Streaming CSV для бухгалтерии. Headers: payment_id, user_id, amount_rub,
// paid_at (ISO), provider, package_id.
func (h *Handler) AdminCSVExport(w http.ResponseWriter, r *http.Request) {
	if !hasPermission(r, "billing:reports:csv_export") {
		httpx.WriteError(w, r, &httpx.Error{
			Status: http.StatusForbidden, Code: "forbidden",
			Message: "missing permission: billing:reports:csv_export",
		})
		return
	}
	from, to, err := parseFromTo(r, h.svc.cfg.Timezone)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition",
		`attachment; filename="billing-`+from.Format("2006-01-02")+
			"_"+to.Format("2006-01-02")+`.csv"`)
	if err := h.svc.StreamPaymentsCSV(r.Context(), w, from, to); err != nil {
		// Headers уже отправлены — просто прерываем тело.
		// Логируется внутри StreamPaymentsCSV.
		_ = err
	}
}

// === helpers ===

// hasPermission — берёт permissions из request context (добавляются auth-
// middleware в main.go при валидации JWT). DUPLICATE-pattern с identity
// (см. план 52).
func hasPermission(r *http.Request, perm string) bool {
	for _, p := range PermissionsFromCtx(r) {
		if p == perm {
			return true
		}
	}
	return false
}

// actorIDFromContext — UUID юзера из JWT-claims.
func actorIDFromContext(r *http.Request) (uuid.UUID, bool) {
	v := r.Context().Value(CtxKeyUserID{})
	if s, ok := v.(string); ok && s != "" {
		if id, err := uuid.Parse(s); err == nil {
			return id, true
		}
	}
	return uuid.Nil, false
}

func remoteIPString(r *http.Request) *string {
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		// Берём первый.
		for i := 0; i < len(v); i++ {
			if v[i] == ',' {
				s := v[:i]
				return &s
			}
		}
		return &v
	}
	if v := r.Header.Get("X-Real-IP"); v != "" {
		return &v
	}
	if r.RemoteAddr != "" {
		return &r.RemoteAddr
	}
	return nil
}

// parseFromTo — RFC3339-date `from` и `to` из query. Default — текущий год
// в timezone.
func parseFromTo(r *http.Request, tz *time.Location) (time.Time, time.Time, error) {
	now := time.Now().In(tz)
	from := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, tz)
	to := time.Date(now.Year(), 12, 31, 23, 59, 59, 0, tz)
	if v := r.URL.Query().Get("from"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return from, to, errors.New("invalid 'from' date (expected YYYY-MM-DD)")
		}
		from = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, tz)
	}
	if v := r.URL.Query().Get("to"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return from, to, errors.New("invalid 'to' date (expected YYYY-MM-DD)")
		}
		to = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, tz)
	}
	if !from.Before(to) {
		return from, to, errors.New("'from' must be before 'to'")
	}
	// Защита от слишком больших окон (для CSV-экспорта).
	if to.Sub(from) > 5*365*24*time.Hour {
		return from, to, errors.New("range too large (max 5 years)")
	}
	return from, to, nil
}

// errReasonRequired — для switch-case в AdminOverride. Не экспортируется,
// помогает не плодить публичные sentinel-errors.
func errReasonRequired() error { return errors.New("reason is required") }

// === auxiliary types для context (DUPLICATE pattern, выставляются
// auth-middleware в main.go) ===

// CtxKeyPermissions — ключ для []string permissions в request context.
// Auth-middleware (в main.go) кладёт claims.Permissions сюда после
// валидации JWT.
type CtxKeyPermissions struct{}

// CtxKeyUserID — ключ для UUID-string юзера. Auth-middleware кладёт
// claims.Subject (string).
type CtxKeyUserID struct{}

// PermissionsFromCtx — экспортируется для использования middleware в
// main.go и handler-ах.
func PermissionsFromCtx(r *http.Request) []string {
	v, _ := r.Context().Value(CtxKeyPermissions{}).([]string)
	return v
}

// decodeJSON — стандартный json.Decode с ограничением body 1 MB.
func decodeJSON(r *http.Request, v any) error {
	if r.Body == nil {
		return errors.New("empty body")
	}
	defer r.Body.Close()
	r.Body = http.MaxBytesReader(nil, r.Body, 1<<20)
	return json.NewDecoder(r.Body).Decode(v)
}
