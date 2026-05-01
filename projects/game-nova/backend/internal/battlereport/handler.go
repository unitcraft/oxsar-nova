// Package battlereport — endpoint'ы для чтения боевых отчётов
// (план 72.1 ч.20.8 — battle viewer).
//
// Боевые отчёты хранятся в таблице battle_reports (миграция 0009).
// Этот пакет предоставляет:
//   GET /api/users/me/battles            — список моих боёв (cursor-paginated)
//   GET /api/battle-reports/{id}         — детали отчёта (с правами:
//                                            attacker_user_id или defender_user_id
//                                            или ACS-participant)
//
// Авторизация: bearer token, юзер должен быть участником боя.
package battlereport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/httpx"
	"oxsar/game-nova/internal/repo"
)

type Handler struct {
	db repo.Exec
}

func NewHandler(db repo.Exec) *Handler {
	return &Handler{db: db}
}

// ListItem — суммарная информация для строки в таблице.
type ListItem struct {
	ID            string    `json:"id"`
	AttackerID    *string   `json:"attacker_user_id,omitempty"`
	DefenderID    *string   `json:"defender_user_id,omitempty"`
	Winner        string    `json:"winner"`
	Rounds        int       `json:"rounds"`
	DebrisMetal   int64     `json:"debris_metal"`
	DebrisSilicon int64     `json:"debris_silicon"`
	LootMetal     int64     `json:"loot_metal"`
	LootSilicon   int64     `json:"loot_silicon"`
	LootHydrogen  int64     `json:"loot_hydrogen"`
	IsAttacker    bool      `json:"is_attacker"`
	At            time.Time `json:"at"`
}

// ListMine GET /api/users/me/battles?limit=20&cursor=<at>
func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	limit := 20
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	cursorAt := time.Now()
	if v := r.URL.Query().Get("cursor"); v != "" {
		if t, err := time.Parse(time.RFC3339Nano, v); err == nil {
			cursorAt = t
		}
	}

	rows, err := h.db.Pool().Query(r.Context(), `
		SELECT id, attacker_user_id, defender_user_id, winner, rounds,
		       debris_metal::bigint, debris_silicon::bigint,
		       loot_metal::bigint, loot_silicon::bigint, loot_hydrogen::bigint,
		       at
		FROM battle_reports
		WHERE (attacker_user_id = $1 OR defender_user_id = $1)
		  AND is_simulation = false
		  AND at < $2
		ORDER BY at DESC
		LIMIT $3
	`, uid, cursorAt, limit)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	defer rows.Close()

	out := []ListItem{}
	var lastAt time.Time
	for rows.Next() {
		var it ListItem
		if err := rows.Scan(
			&it.ID, &it.AttackerID, &it.DefenderID, &it.Winner, &it.Rounds,
			&it.DebrisMetal, &it.DebrisSilicon,
			&it.LootMetal, &it.LootSilicon, &it.LootHydrogen,
			&it.At,
		); err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			return
		}
		it.IsAttacker = it.AttackerID != nil && *it.AttackerID == uid
		out = append(out, it)
		lastAt = it.At
	}
	if err := rows.Err(); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}

	resp := map[string]any{"battles": out}
	if len(out) == limit {
		resp["next_cursor"] = lastAt.Format(time.RFC3339Nano)
	}
	httpx.WriteJSON(w, r, http.StatusOK, resp)
}

// GetByID GET /api/battle-reports/{id} — публичный анонимный endpoint
// (план 72.1 ч.20.11). Любой пользователь по ссылке может посмотреть
// результат боя или симуляции. Permission check снят — отчёты
// идентифицируются непредсказуемым UUID v7.
//
// Возвращает {report, started_at} — фронт показывает «Флоты соперников
// встрелись в <started_at> часов:» (план 72.1 ч.20.11.12).
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var reportRaw []byte
	var at time.Time
	err := h.db.Pool().QueryRow(r.Context(), `
		SELECT report, at
		FROM battle_reports
		WHERE id = $1
	`, id).Scan(&reportRaw, &at)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.WriteError(w, r, httpx.ErrNotFound)
			return
		}
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	wrapped := struct {
		Report    json.RawMessage `json:"report"`
		StartedAt time.Time       `json:"started_at"`
	}{
		Report:    reportRaw,
		StartedAt: at,
	}
	_ = json.NewEncoder(w).Encode(wrapped)
}

// контекст для совместимости (не используется явно, но импорт нужен).
var _ = context.Background
