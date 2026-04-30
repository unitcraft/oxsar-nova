// Package simulator — HTTP handler для боевого симулятора.
//
// План 72.1 ч.20.11.7: симулятор прогоняет num_sim итераций,
// возвращает агрегированную сводку (SimStats) + сохраняет последний
// бой в battle_reports (is_simulation=true). Фронт показывает сводку
// в стиле legacy simulator.tpl и ссылку «Отчёт о сражении» на
// /battle-report/{id} (анонимный публичный просмотр).
//
// ADR-0002: rendering на frontend, backend возвращает структурированный
// JSON.
package simulator

import (
	"encoding/json"
	"net/http"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/battle"
	"oxsar/game-nova/internal/httpx"
	"oxsar/game-nova/internal/repo"
	"oxsar/game-nova/pkg/ids"
)

type Handler struct {
	db repo.Exec
}

func NewHandler(db repo.Exec) *Handler {
	return &Handler{db: db}
}

// RunResponse — ответ POST /api/simulator/run.
type RunResponse struct {
	ID     string         `json:"id"`     // UUID последнего боя в battle_reports
	Stats  battle.SimStats `json:"stats"` // агрегат по num_sim итераций
	Report battle.Report  `json:"report"` // последний бой целиком (для round-by-round, опционально)
}

// Run POST /api/simulator/run — N прогонов, агрегат + сохранение последнего.
func (h *Handler) Run(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserID(r.Context()); !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	var in battle.Input
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	if len(in.Attackers) == 0 || len(in.Defenders) == 0 {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "attackers and defenders required"))
		return
	}
	n := in.NumSim
	if n < 1 {
		n = 1
	}
	if n > 100 {
		n = 100
	}

	stats, last, err := battle.MultiRun(in, n)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}

	id := ids.New()
	reportJSON, _ := json.Marshal(last)
	_, err = h.db.Pool().Exec(r.Context(), `
		INSERT INTO battle_reports (
			id, attacker_user_id, defender_user_id, planet_id,
			seed, winner, rounds,
			debris_metal, debris_silicon,
			loot_metal, loot_silicon, loot_hydrogen,
			report, is_simulation
		) VALUES ($1, NULL, NULL, NULL, $2, $3, $4, $5, $6, 0, 0, 0, $7, true)
	`, id, int64(last.Seed), last.Winner, last.Rounds,
		last.DebrisMetal, last.DebrisSilicon,
		reportJSON,
	)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}

	httpx.WriteJSON(w, r, http.StatusOK, RunResponse{
		ID:     id,
		Stats:  stats,
		Report: last,
	})
}
