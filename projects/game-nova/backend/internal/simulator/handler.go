// Package simulator — HTTP handler для боевого симулятора.
//
// Тонкая обёртка над internal/battle.Calculate (план 72.1 ч.20.7).
// Принимает голый battle.Input, возвращает battle.Report без обращения
// к БД (бой чистая функция). Авторизация — любой залогиненный юзер.
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
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// Run POST /api/simulator/run — запускает бой с переданными
// атакующими/защитниками и возвращает результат.
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
	// Защита от пустых сторон.
	if len(in.Attackers) == 0 || len(in.Defenders) == 0 {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "attackers and defenders required"))
		return
	}
	// Лимит num_sim — в legacy 100 max, ставим тот же.
	if in.NumSim > 100 {
		in.NumSim = 100
	}
	report, err := battle.Calculate(in)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, report)
}
