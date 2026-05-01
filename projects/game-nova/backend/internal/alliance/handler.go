package alliance

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/httpx"
	"oxsar/game-nova/pkg/idempotency"
	"oxsar/game-nova/pkg/metrics"
)

type Handler struct {
	svc *Service
	rdb *redis.Client // optional, для Idempotency-Key (план 67 R9)
}

func NewHandler(s *Service) *Handler { return &Handler{svc: s} }

// WithRedis подключает Redis для Idempotency-Key middleware на
// PATCH/POST-операциях альянса (R9). Если nil — idempotency no-op.
func (h *Handler) WithRedis(rdb *redis.Client) *Handler {
	h.rdb = rdb
	return h
}

// recordAction увеличивает Prometheus-счётчик действия (R8).
// status: ok|forbidden|error.
func recordAction(action, status string) {
	if metrics.AllianceActions != nil {
		metrics.AllianceActions.WithLabelValues(action, status).Inc()
	}
}

// statusFromErr — для метрик: классификация ошибок.
func statusFromErr(err error) string {
	switch {
	case err == nil:
		return "ok"
	case errors.Is(err, ErrForbidden), errors.Is(err, ErrNotOwner), errors.Is(err, ErrNotMember):
		return "forbidden"
	default:
		return "error"
	}
}

// List GET /api/alliances
//
// Query (план 67 Ф.4, U-012):
//   - q              — полнотекст по name+tag (prefix-match для одного слова,
//     websearch для фраз)
//   - is_open        — true|false, фильтр по открытости
//   - min_members    — минимальное число участников
//   - max_members    — максимальное число участников
//   - limit          — 1..100, default 50
//   - offset         — пагинация, default 0
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := ListFilters{
		Q:      q.Get("q"),
		Limit:  parseIntDefault(q.Get("limit"), 50),
		Offset: parseIntDefault(q.Get("offset"), 0),
	}
	if v := q.Get("is_open"); v != "" {
		b := v == "true" || v == "1"
		f.IsOpen = &b
	}
	if v := q.Get("min_members"); v != "" {
		f.MinMembers = parseIntDefault(v, 0)
	}
	if v := q.Get("max_members"); v != "" {
		f.MaxMembers = parseIntDefault(v, 0)
	}
	list, err := h.svc.List(r.Context(), f)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{
		"alliances": list,
		"limit":     f.Limit,
		"offset":    f.Offset,
	})
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

// Get GET /api/alliances/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	al, members, err := h.svc.Get(r.Context(), id)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusOK, map[string]any{
			"alliance": al,
			"members":  members,
		})
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// My GET /api/alliances/me
func (h *Handler) My(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	al, members, err := h.svc.MyAlliance(r.Context(), uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	if al == nil {
		httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"alliance": nil, "members": nil})
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"alliance": al, "members": members})
}

// Create POST /api/alliances
// Body: {"tag":"TAG","name":"Full Name","description":"..."}
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	var req struct {
		Tag         string `json:"tag"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	al, err := h.svc.Create(r.Context(), uid, req.Tag, req.Name, req.Description)
	recordAction(ActionAllianceCreated, statusFromErr(err))
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusCreated, map[string]any{"alliance": al})
	case errors.Is(err, ErrAlreadyMember):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "already in an alliance"))
	case errors.Is(err, ErrTagTaken):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "tag already taken"))
	case errors.Is(err, ErrNameTaken):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "name already taken"))
	case errors.Is(err, ErrInvalidTag):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "tag must be 3–5 latin letters/digits"))
	case errors.Is(err, ErrNameForbidden):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "name contains forbidden word"))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// Join POST /api/alliances/{id}/join
// Body (optional): {"message":"..."}
func (h *Handler) Join(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	var req struct {
		Message string `json:"message"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	id := chi.URLParam(r, "id")
	joined, err := h.svc.Join(r.Context(), uid, id, req.Message)
	switch {
	case err == nil:
		if joined {
			w.WriteHeader(http.StatusNoContent)
		} else {
			httpx.WriteJSON(w, r, http.StatusAccepted, map[string]string{"status": "application_pending"})
		}
	case errors.Is(err, ErrAlreadyMember):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "already in an alliance"))
	case errors.Is(err, ErrApplicationExists):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "application already pending"))
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// SetOpen PATCH /api/alliances/{id}/open
// Body: {"is_open":true|false}
func (h *Handler) SetOpen(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	var req struct {
		IsOpen bool `json:"is_open"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	id := chi.URLParam(r, "id")
	err := h.svc.SetOpen(r.Context(), uid, id, req.IsOpen)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrNotOwner):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// Applications GET /api/alliances/{id}/applications
func (h *Handler) Applications(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	apps, err := h.svc.Applications(r.Context(), uid, id)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"applications": apps})
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrNotOwner):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// Approve POST /api/alliances/applications/{appID}/approve
func (h *Handler) Approve(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	appID := chi.URLParam(r, "appID")
	err := h.svc.Approve(r.Context(), uid, appID)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrApplicationNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrNotOwner):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrAlreadyMember):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "applicant already joined another alliance"))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// Reject DELETE /api/alliances/applications/{appID}
func (h *Handler) Reject(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	appID := chi.URLParam(r, "appID")
	err := h.svc.Reject(r.Context(), uid, appID)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrApplicationNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrNotOwner):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// Leave POST /api/alliances/leave
func (h *Handler) Leave(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	err := h.svc.Leave(r.Context(), uid)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrNotMember):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "not in an alliance"))
	case errors.Is(err, ErrCannotLeaveOwn):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "owner must disband the alliance first"))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// Disband DELETE /api/alliances/{id}
func (h *Handler) Disband(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	err := h.svc.Disband(r.Context(), uid, id)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrNotOwner):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// SetMemberRank PATCH /api/alliances/{id}/members/{userID}/rank
// Body: {"rank_name":"..."}
func (h *Handler) SetMemberRank(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	allianceID := chi.URLParam(r, "id")
	memberUID := chi.URLParam(r, "userID")
	var body struct {
		RankName string `json:"rank_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	err := h.svc.SetMemberRank(r.Context(), uid, allianceID, memberUID, body.RankName)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrMemberNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrNotOwner):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// GetRelations GET /api/alliances/{id}/relations
func (h *Handler) GetRelations(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	rels, err := h.svc.GetRelations(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"relations": rels})
}

// ProposeRelation PUT /api/alliances/{id}/relations/{target_id}
// Body: {"relation":"nap"|"war"|"ally"|"none"}
// WAR активно сразу; NAP/ALLY — pending до подтверждения target.
func (h *Handler) ProposeRelation(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	targetID := chi.URLParam(r, "target_id")

	var body struct {
		Relation string `json:"relation"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}

	err := h.svc.ProposeRelation(r.Context(), uid, id, targetID, body.Relation)
	recordAction(ActionRelationProposed, statusFromErr(err))
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrNotMember):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrTargetNotFound):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrNotFound, err.Error()))
	case errors.Is(err, ErrNotOwner), errors.Is(err, ErrForbidden):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrInvalidRelation), errors.Is(err, ErrRelationSelf):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// AcceptRelation POST /api/alliances/{id}/relations/{initiator_id}/accept
// Подтверждает входящее NAP/ALLY предложение. {id} — наш альянс, {initiator_id} — кто предложил.
func (h *Handler) AcceptRelation(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	myID := chi.URLParam(r, "id")
	initiatorID := chi.URLParam(r, "initiator_id")

	err := h.svc.AcceptRelation(r.Context(), uid, myID, initiatorID)
	recordAction(ActionRelationAccepted, statusFromErr(err))
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrTargetNotFound), errors.Is(err, ErrNotMember):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrNotOwner), errors.Is(err, ErrForbidden):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// RejectRelation DELETE /api/alliances/{id}/relations/{initiator_id}
// Отклоняет входящее pending предложение.
func (h *Handler) RejectRelation(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	myID := chi.URLParam(r, "id")
	initiatorID := chi.URLParam(r, "initiator_id")

	err := h.svc.RejectRelation(r.Context(), uid, myID, initiatorID)
	recordAction(ActionRelationRejected, statusFromErr(err))
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrNotMember):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrNotOwner), errors.Is(err, ErrForbidden):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// План 67 Ф.2: расширенные handlers (descriptions, ranks, kick, audit).

// GetDescriptions GET /api/alliances/{id}/descriptions
//
// Возвращает 3 описания + legacy поле + контекст вьюера. Видимость
// internal/apply фильтруется в сервисе (см. GetDescriptions).
func (h *Handler) GetDescriptions(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	uid, _ := auth.UserID(r.Context()) // "" если анонимный
	v, err := h.svc.GetDescriptions(r.Context(), uid, id)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusOK, v)
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// UpdateDescriptions PATCH /api/alliances/{id}/descriptions
// Body: {"description_external"?:"...", "description_internal"?:"...", "description_apply"?:"..."}
//
// Право: PermChangeDescription. Поддерживает Idempotency-Key.
func (h *Handler) UpdateDescriptions(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	idem := idempotency.FromRequest(r, h.rdb)
	if idem.Replay(w) {
		return
	}
	id := chi.URLParam(r, "id")
	var body struct {
		External *string `json:"description_external"`
		Internal *string `json:"description_internal"`
		Apply    *string `json:"description_apply"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	in := UpdateDescriptionsInput{External: body.External, Internal: body.Internal, Apply: body.Apply}
	err := h.svc.UpdateDescriptions(r.Context(), uid, id, in)
	recordAction(ActionDescriptionChanged, statusFromErr(err))
	switch {
	case err == nil:
		idem.Record(http.StatusNoContent, nil)
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrNotMember), errors.Is(err, ErrForbidden):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrDescriptionTooLong):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// ListRanks GET /api/alliances/{id}/ranks
func (h *Handler) ListRanks(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	out, err := h.svc.ListRanks(r.Context(), uid, id)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"ranks": out})
	case errors.Is(err, ErrNotMember):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// CreateRank POST /api/alliances/{id}/ranks
// Body: {"name":"...","position":100,"permissions":{"can_invite":true,...}}
//
// Право: PermManageRanks. Поддерживает Idempotency-Key.
func (h *Handler) CreateRank(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	idem := idempotency.FromRequest(r, h.rdb)
	if idem.Replay(w) {
		return
	}
	id := chi.URLParam(r, "id")
	var body struct {
		Name        string          `json:"name"`
		Position    int             `json:"position"`
		Permissions map[string]bool `json:"permissions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	rank, err := h.svc.CreateRank(r.Context(), uid, id, body.Name, body.Position, body.Permissions)
	recordAction(ActionRankCreated, statusFromErr(err))
	switch {
	case err == nil:
		buf := httpx.MarshalJSON(map[string]any{"rank": rank})
		idem.Record(http.StatusCreated, buf)
		httpx.WriteJSONBytes(w, r, http.StatusCreated, buf)
	case errors.Is(err, ErrNotMember), errors.Is(err, ErrForbidden):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrRankNameTaken):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, err.Error()))
	case errors.Is(err, ErrRankNameInvalid), errors.Is(err, ErrInvalidPermission):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// UpdateRank PATCH /api/alliances/{id}/ranks/{rank_id}
// Body: {"name"?:"...","position"?:100,"permissions"?:{...}}
func (h *Handler) UpdateRank(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	idem := idempotency.FromRequest(r, h.rdb)
	if idem.Replay(w) {
		return
	}
	id := chi.URLParam(r, "id")
	rankID := chi.URLParam(r, "rank_id")
	var body struct {
		Name        *string          `json:"name"`
		Position    *int             `json:"position"`
		Permissions *map[string]bool `json:"permissions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	in := UpdateRankInput{Name: body.Name, Position: body.Position, Permissions: body.Permissions}
	err := h.svc.UpdateRank(r.Context(), uid, id, rankID, in)
	recordAction(ActionRankUpdated, statusFromErr(err))
	switch {
	case err == nil:
		idem.Record(http.StatusNoContent, nil)
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrRankNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrNotMember), errors.Is(err, ErrForbidden):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrRankNameTaken):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, err.Error()))
	case errors.Is(err, ErrRankNameInvalid), errors.Is(err, ErrInvalidPermission):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// DeleteRank DELETE /api/alliances/{id}/ranks/{rank_id}
func (h *Handler) DeleteRank(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	rankID := chi.URLParam(r, "rank_id")
	err := h.svc.DeleteRank(r.Context(), uid, id, rankID)
	recordAction(ActionRankDeleted, statusFromErr(err))
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrRankNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrNotMember), errors.Is(err, ErrForbidden):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// AssignMemberRank PATCH /api/alliances/{id}/members/{userID}/rank-id
// Body: {"rank_id": "uuid" | null}
func (h *Handler) AssignMemberRank(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	idem := idempotency.FromRequest(r, h.rdb)
	if idem.Replay(w) {
		return
	}
	id := chi.URLParam(r, "id")
	memberUID := chi.URLParam(r, "userID")
	var body struct {
		RankID *string `json:"rank_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	rankID := ""
	if body.RankID != nil {
		rankID = *body.RankID
	}
	err := h.svc.AssignMemberRank(r.Context(), uid, id, memberUID, rankID)
	recordAction(ActionMemberRankAssigned, statusFromErr(err))
	switch {
	case err == nil:
		idem.Record(http.StatusNoContent, nil)
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrRankNotFound), errors.Is(err, ErrMemberNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrNotMember), errors.Is(err, ErrForbidden):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// Kick DELETE /api/alliances/{id}/members/{userID}
//
// Право: PermKick. Owner кикнуть нельзя; самого себя — Leave.
func (h *Handler) Kick(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	memberUID := chi.URLParam(r, "userID")
	err := h.svc.Kick(r.Context(), uid, id, memberUID)
	recordAction(ActionMemberKicked, statusFromErr(err))
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrMemberNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrNotMember), errors.Is(err, ErrForbidden):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrCannotKickOwner), errors.Is(err, ErrCannotKickSelf):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// ListAudit GET /api/alliances/{id}/audit?action=&actor_id=&limit=&offset=
func (h *Handler) ListAudit(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	filters := AuditFilters{
		Action:  q.Get("action"),
		ActorID: q.Get("actor_id"),
		Limit:   limit,
		Offset:  offset,
	}
	entries, err := h.svc.ListAudit(r.Context(), uid, id, filters)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusOK, map[string]any{
			"entries": entries,
			"limit":   filters.Limit,
			"offset":  filters.Offset,
		})
	case errors.Is(err, ErrNotMember):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// BroadcastMail POST /api/alliances/{id}/broadcast (план 72.1.43).
// Body: { title, body }. Permission: CAN_SEND_GLOBAL_MAIL.
func (h *Handler) BroadcastMail(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	allianceID := chi.URLParam(r, "id")
	var body struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	if body.Title == "" || body.Body == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "title and body required"))
		return
	}
	err := h.svc.BroadcastMail(r.Context(), uid, allianceID, body.Title, body.Body)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrNotMember):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrForbidden):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// UpdateTagName PATCH /api/alliances/{id} (план 72.1.43).
// Body: { tag?, name? }. Только owner.
func (h *Handler) UpdateTagName(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	allianceID := chi.URLParam(r, "id")
	var body struct {
		Tag  string `json:"tag,omitempty"`
		Name string `json:"name,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	if body.Tag == "" && body.Name == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "tag or name required"))
		return
	}
	err := h.svc.UpdateTagName(r.Context(), uid, allianceID, body.Tag, body.Name)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrNotOwner):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrInvalidTag),
		errors.Is(err, ErrTagTaken),
		errors.Is(err, ErrNameTaken),
		errors.Is(err, ErrNameForbidden):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}
