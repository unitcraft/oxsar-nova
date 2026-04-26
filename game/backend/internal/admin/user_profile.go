// Package admin — глубокая карточка игрока (план 14 Ф.2).
//
// GET /api/admin/users/{id} — агрегат: базовые поля, планеты, флот в полёте,
// лоты на рынке и арт-рынке, активные офицеры, последние 20 отчётов,
// последние 20 транзакций res_log, последние 20 credit_purchases,
// последние 20 inbox-сообщений, артефакты.
//
// Write-операции (через тот же AuditMiddleware, см. audit.go):
//   POST /api/admin/users/{id}/resources          — добавить/списать ресурсы на планете
//   POST /api/admin/users/{id}/artefacts/grant    — выдать артефакт
//   DELETE /api/admin/users/{id}/artefacts/{aid}  — отобрать артефакт
//
// Все SELECT'ы ограничены LIMIT 20 — админка быстро отвечает даже для
// старых аккаунтов с десятками тысяч записей в истории.

package admin

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/httpx"
)

// UserProfile — полная карточка игрока.
type UserProfile struct {
	// Базовые поля
	ID         string     `json:"id"`
	Username   string     `json:"username"`
	Email      string     `json:"email"`
	Role       string     `json:"role"`
	Credit     int64      `json:"credit"`
	Score      int64      `json:"score"`
	BannedAt   *time.Time `json:"banned_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	LastSeenAt time.Time  `json:"last_seen_at"`

	Planets      []ProfilePlanet      `json:"planets"`
	Fleets       []ProfileFleet       `json:"fleets"`
	MarketLots   []ProfileMarketLot   `json:"market_lots"`
	ArtefactLots []ProfileArtLot      `json:"artefact_lots"`
	Officers     []ProfileOfficer     `json:"officers"`
	Artefacts    []ProfileArtefact    `json:"artefacts"`
	ResLog       []ProfileResLog      `json:"res_log"`
	Purchases    []ProfilePurchase    `json:"purchases"`
	Messages     []ProfileMessage     `json:"messages_recent"`
	Reports      []ProfileReport      `json:"reports_recent"`
}

type ProfilePlanet struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Galaxy   int    `json:"galaxy"`
	System   int    `json:"system"`
	Position int    `json:"position"`
	IsMoon   bool   `json:"is_moon"`
	Metal    int64  `json:"metal"`
	Silicon  int64  `json:"silicon"`
	Hydrogen int64  `json:"hydrogen"`
}

type ProfileFleet struct {
	ID          string    `json:"id"`
	Mission     int       `json:"mission"`
	State       string    `json:"state"`
	DstGalaxy   int       `json:"dst_galaxy"`
	DstSystem   int       `json:"dst_system"`
	DstPosition int       `json:"dst_position"`
	DstIsMoon   bool      `json:"dst_is_moon"`
	ArriveAt    time.Time `json:"arrive_at"`
}

type ProfileMarketLot struct {
	ID           string `json:"id"`
	SellResource string `json:"sell_resource"`
	SellAmount   int64  `json:"sell_amount"`
	BuyResource  string `json:"buy_resource"`
	BuyAmount    int64  `json:"buy_amount"`
	State        string `json:"state"`
}

type ProfileArtLot struct {
	ID          string `json:"id"`
	ArtefactID  string `json:"artefact_id"`
	UnitID      int    `json:"unit_id"`
	PriceCredit int64  `json:"price_credit"`
}

type ProfileOfficer struct {
	Key       string    `json:"key"`
	ExpiresAt time.Time `json:"expires_at"`
}

type ProfileArtefact struct {
	ID     string `json:"id"`
	UnitID int    `json:"unit_id"`
	State  string `json:"state"`
}

type ProfileResLog struct {
	Reason      string     `json:"reason"`
	PlanetID    *string    `json:"planet_id,omitempty"`
	DMetal      int64      `json:"d_metal"`
	DSilicon    int64      `json:"d_silicon"`
	DHydrogen   int64      `json:"d_hydrogen"`
	CreatedAt   time.Time  `json:"created_at"`
}

type ProfilePurchase struct {
	ID         string     `json:"id"`
	PackageKey string     `json:"package_key"`
	Credits    int        `json:"credits"`
	PriceRub   float64    `json:"price_rub"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	PaidAt     *time.Time `json:"paid_at,omitempty"`
}

type ProfileMessage struct {
	ID        string    `json:"id"`
	Folder    int       `json:"folder"`
	Subject   string    `json:"subject"`
	CreatedAt time.Time `json:"created_at"`
	Read      bool      `json:"read"`
}

type ProfileReport struct {
	ID        string    `json:"id"`
	Kind      string    `json:"kind"` // battle / espionage / expedition
	CreatedAt time.Time `json:"created_at"`
}

// GetUserProfile GET /api/admin/users/{id}
func (h *Handler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	uid := chi.URLParam(r, "id")
	if uid == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "user id required"))
		return
	}

	// 1. Базовый блок + score
	var p UserProfile
	err := h.db.Pool().QueryRow(r.Context(), `
		SELECT u.id, u.username, u.email, COALESCE(u.role::text,''), u.credit,
		       COALESCE(s.score, 0), u.banned_at, u.created_at, u.last_seen_at
		FROM users u
		LEFT JOIN scores s ON s.user_id = u.id
		WHERE u.id = $1`, uid).
		Scan(&p.ID, &p.Username, &p.Email, &p.Role, &p.Credit,
			&p.Score, &p.BannedAt, &p.CreatedAt, &p.LastSeenAt)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrNotFound, "user not found"))
		return
	}
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}

	// Остальные блоки — параллельность не нужна, запросы быстрые (LIMIT 20).
	p.Planets = h.loadPlanets(r, uid)
	p.Fleets = h.loadFleets(r, uid)
	p.MarketLots = h.loadMarketLots(r, uid)
	p.ArtefactLots = h.loadArtefactLots(r, uid)
	p.Officers = h.loadOfficers(r, uid)
	p.Artefacts = h.loadArtefacts(r, uid)
	p.ResLog = h.loadResLog(r, uid)
	p.Purchases = h.loadPurchases(r, uid)
	p.Messages = h.loadMessages(r, uid)
	p.Reports = h.loadReports(r, uid)

	httpx.WriteJSON(w, r, http.StatusOK, p)
}

func (h *Handler) loadPlanets(r *http.Request, uid string) []ProfilePlanet {
	rows, err := h.db.Pool().Query(r.Context(), `
		SELECT id, name, galaxy, system, position, is_moon, metal, silicon, hydrogen
		FROM planets WHERE user_id = $1 AND destroyed_at IS NULL
		ORDER BY galaxy, system, position`, uid)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []ProfilePlanet
	for rows.Next() {
		var p ProfilePlanet
		if err := rows.Scan(&p.ID, &p.Name, &p.Galaxy, &p.System, &p.Position,
			&p.IsMoon, &p.Metal, &p.Silicon, &p.Hydrogen); err != nil {
			return out
		}
		out = append(out, p)
	}
	return out
}

func (h *Handler) loadFleets(r *http.Request, uid string) []ProfileFleet {
	rows, err := h.db.Pool().Query(r.Context(), `
		SELECT id, mission, state, dst_galaxy, dst_system, dst_position, dst_is_moon, arrive_at
		FROM fleets
		WHERE user_id = $1 AND state IN ('outbound', 'returning')
		ORDER BY arrive_at ASC LIMIT 50`, uid)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []ProfileFleet
	for rows.Next() {
		var f ProfileFleet
		if err := rows.Scan(&f.ID, &f.Mission, &f.State, &f.DstGalaxy, &f.DstSystem,
			&f.DstPosition, &f.DstIsMoon, &f.ArriveAt); err != nil {
			return out
		}
		out = append(out, f)
	}
	return out
}

func (h *Handler) loadMarketLots(r *http.Request, uid string) []ProfileMarketLot {
	rows, err := h.db.Pool().Query(r.Context(), `
		SELECT id, sell_resource, sell_amount, buy_resource, buy_amount, state
		FROM market_lots WHERE seller_id = $1 AND state = 'open'
		ORDER BY created_at DESC LIMIT 20`, uid)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []ProfileMarketLot
	for rows.Next() {
		var l ProfileMarketLot
		if err := rows.Scan(&l.ID, &l.SellResource, &l.SellAmount, &l.BuyResource,
			&l.BuyAmount, &l.State); err != nil {
			return out
		}
		out = append(out, l)
	}
	return out
}

func (h *Handler) loadArtefactLots(r *http.Request, uid string) []ProfileArtLot {
	rows, err := h.db.Pool().Query(r.Context(), `
		SELECT id, artefact_id, unit_id, price_credit FROM artefact_offers
		WHERE seller_user_id = $1 ORDER BY listed_at DESC LIMIT 20`, uid)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []ProfileArtLot
	for rows.Next() {
		var l ProfileArtLot
		if err := rows.Scan(&l.ID, &l.ArtefactID, &l.UnitID, &l.PriceCredit); err != nil {
			return out
		}
		out = append(out, l)
	}
	return out
}

func (h *Handler) loadOfficers(r *http.Request, uid string) []ProfileOfficer {
	rows, err := h.db.Pool().Query(r.Context(), `
		SELECT officer_key, expires_at FROM officer_active
		WHERE user_id = $1 AND expires_at > now()
		ORDER BY expires_at`, uid)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []ProfileOfficer
	for rows.Next() {
		var o ProfileOfficer
		if err := rows.Scan(&o.Key, &o.ExpiresAt); err != nil {
			return out
		}
		out = append(out, o)
	}
	return out
}

func (h *Handler) loadArtefacts(r *http.Request, uid string) []ProfileArtefact {
	rows, err := h.db.Pool().Query(r.Context(), `
		SELECT id, unit_id, state::text FROM artefacts_user
		WHERE user_id = $1 ORDER BY acquired_at DESC LIMIT 50`, uid)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []ProfileArtefact
	for rows.Next() {
		var a ProfileArtefact
		if err := rows.Scan(&a.ID, &a.UnitID, &a.State); err != nil {
			return out
		}
		out = append(out, a)
	}
	return out
}

func (h *Handler) loadResLog(r *http.Request, uid string) []ProfileResLog {
	rows, err := h.db.Pool().Query(r.Context(), `
		SELECT reason, planet_id, delta_metal, delta_silicon, delta_hydrogen, created_at
		FROM res_log WHERE user_id = $1 ORDER BY created_at DESC LIMIT 20`, uid)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []ProfileResLog
	for rows.Next() {
		var l ProfileResLog
		if err := rows.Scan(&l.Reason, &l.PlanetID, &l.DMetal, &l.DSilicon,
			&l.DHydrogen, &l.CreatedAt); err != nil {
			return out
		}
		out = append(out, l)
	}
	return out
}

func (h *Handler) loadPurchases(r *http.Request, uid string) []ProfilePurchase {
	rows, err := h.db.Pool().Query(r.Context(), `
		SELECT id, package_key, amount_credits, price_rub, status, created_at, paid_at
		FROM credit_purchases WHERE user_id = $1
		ORDER BY created_at DESC LIMIT 20`, uid)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []ProfilePurchase
	for rows.Next() {
		var p ProfilePurchase
		if err := rows.Scan(&p.ID, &p.PackageKey, &p.Credits, &p.PriceRub,
			&p.Status, &p.CreatedAt, &p.PaidAt); err != nil {
			return out
		}
		out = append(out, p)
	}
	return out
}

func (h *Handler) loadMessages(r *http.Request, uid string) []ProfileMessage {
	rows, err := h.db.Pool().Query(r.Context(), `
		SELECT id, folder, subject, created_at, read_at IS NOT NULL
		FROM messages
		WHERE to_user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC LIMIT 20`, uid)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []ProfileMessage
	for rows.Next() {
		var m ProfileMessage
		if err := rows.Scan(&m.ID, &m.Folder, &m.Subject, &m.CreatedAt, &m.Read); err != nil {
			return out
		}
		out = append(out, m)
	}
	return out
}

func (h *Handler) loadReports(r *http.Request, uid string) []ProfileReport {
	// Объединяем три таблицы отчётов через UNION и сортируем по дате.
	rows, err := h.db.Pool().Query(r.Context(), `
		(SELECT id, 'battle' AS kind, created_at FROM battle_reports WHERE user_id = $1)
		UNION ALL
		(SELECT id, 'espionage' AS kind, created_at FROM espionage_reports WHERE owner_id = $1)
		UNION ALL
		(SELECT id, 'expedition' AS kind, created_at FROM expedition_reports WHERE user_id = $1)
		ORDER BY created_at DESC LIMIT 20`, uid)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []ProfileReport
	for rows.Next() {
		var rep ProfileReport
		if err := rows.Scan(&rep.ID, &rep.Kind, &rep.CreatedAt); err != nil {
			return out
		}
		out = append(out, rep)
	}
	return out
}

// ── Write-операции ──────────────────────────────────────────────────

// GrantResources POST /api/admin/users/{id}/resources
// body: {planet_id, metal, silicon, hydrogen} — значения могут быть
// отрицательными (списание). Не даём балансу уйти ниже 0.
func (h *Handler) GrantResources(w http.ResponseWriter, r *http.Request) {
	uid := chi.URLParam(r, "id")
	var req struct {
		PlanetID string `json:"planet_id"`
		Metal    int64  `json:"metal"`
		Silicon  int64  `json:"silicon"`
		Hydrogen int64  `json:"hydrogen"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	if req.PlanetID == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "planet_id required"))
		return
	}
	if req.Metal == 0 && req.Silicon == 0 && req.Hydrogen == 0 {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "at least one delta must be non-zero"))
		return
	}

	// Проверяем, что планета принадлежит игроку.
	tag, err := h.db.Pool().Exec(r.Context(), `
		UPDATE planets
		SET metal    = GREATEST(0, metal    + $2),
		    silicon  = GREATEST(0, silicon  + $3),
		    hydrogen = GREATEST(0, hydrogen + $4)
		WHERE id = $1 AND user_id = $5 AND destroyed_at IS NULL`,
		req.PlanetID, req.Metal, req.Silicon, req.Hydrogen, uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	if tag.RowsAffected() == 0 {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrNotFound, "planet not found or not owned by user"))
		return
	}

	// Пишем в res_log для симметрии с игровыми транзакциями.
	_, _ = h.db.Pool().Exec(r.Context(), `
		INSERT INTO res_log (user_id, planet_id, reason, delta_metal, delta_silicon, delta_hydrogen)
		VALUES ($1, $2, 'admin_grant', $3, $4, $5)`,
		uid, req.PlanetID, req.Metal, req.Silicon, req.Hydrogen)

	w.WriteHeader(http.StatusNoContent)
}

// GrantArtefact POST /api/admin/users/{id}/artefacts/grant
// body: {unit_id}
func (h *Handler) GrantArtefact(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	var req struct {
		UnitID int `json:"unit_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	if req.UnitID <= 0 {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "unit_id required"))
		return
	}

	id := uuid.New().String()
	_, err := h.db.Pool().Exec(r.Context(), `
		INSERT INTO artefacts_user (id, user_id, unit_id, state)
		VALUES ($1, $2, $3, 'held')`, id, userID, req.UnitID)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"id": id})
}

// DeleteArtefact DELETE /api/admin/users/{id}/artefacts/{aid}
func (h *Handler) DeleteArtefact(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	artID := chi.URLParam(r, "aid")

	tag, err := h.db.Pool().Exec(r.Context(), `
		DELETE FROM artefacts_user WHERE id = $1 AND user_id = $2`,
		artID, userID)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	if tag.RowsAffected() == 0 {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrNotFound, "artefact not found or not owned by user"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
