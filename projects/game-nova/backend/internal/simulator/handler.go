// Package simulator — HTTP handler для боевого симулятора.
//
// План 72.1 ч.20.11: симуляция сохраняется в battle_reports
// (is_simulation=true), endpoint возвращает {id, report}. Юзер
// (или анонимный гость) перенаправляется на /battle-report/{id}
// который доступен публично без auth.
//
// ADR-0002: симулятор — порт legacy oxsar2-java/Assault.java
// (rendering отчёта на frontend, не в backend).
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

// Run POST /api/simulator/run — запускает бой и сохраняет в БД.
// Возвращает {id, report}; frontend редиректит на /battle-report/{id}.
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
	if in.NumSim > 100 {
		in.NumSim = 100
	}
	report, err := battle.Calculate(in)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}

	// Сохраняем симуляцию в battle_reports (is_simulation=true).
	id := ids.New()
	reportJSON, _ := json.Marshal(report)
	_, err = h.db.Pool().Exec(r.Context(), `
		INSERT INTO battle_reports (
			id, attacker_user_id, defender_user_id, planet_id,
			seed, winner, rounds,
			debris_metal, debris_silicon,
			loot_metal, loot_silicon, loot_hydrogen,
			report, is_simulation
		) VALUES ($1, NULL, NULL, NULL, $2, $3, $4, $5, $6, 0, 0, 0, $7, true)
	`, id, int64(report.Seed), report.Winner, report.Rounds,
		report.DebrisMetal, report.DebrisSilicon,
		reportJSON,
	)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}

	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{
		"id":     id,
		"report": report,
	})
}
