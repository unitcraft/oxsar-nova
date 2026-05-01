// Package simulator — HTTP handler для боевого симулятора.
//
// План 72.1 ч.20.11.7: симулятор прогоняет num_sim итераций,
// возвращает агрегированную сводку (SimStats) + сохраняет последний
// бой в battle_reports (is_simulation=true). Фронт показывает сводку
// в стиле legacy simulator.tpl и ссылку «Отчёт о сражении» на
// /battle-report/{id} (анонимный публичный просмотр).
//
// План 72.1 ч.20.11.8: cost юнитов backend подгружает из configs
// (ships.yml + defense.yml) и проставляет в Side.Units перед расчётом —
// фронт не обязан знать стоимости (они — backend balance).
//
// ADR-0002: rendering на frontend, backend возвращает структурированный
// JSON.
package simulator

import (
	"encoding/json"
	"net/http"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/battle"
	"oxsar/game-nova/internal/config"
	"oxsar/game-nova/internal/httpx"
	"oxsar/game-nova/internal/repo"
	"oxsar/game-nova/pkg/ids"
)

type Handler struct {
	db  repo.Exec
	cat *config.Catalog
}

func NewHandler(db repo.Exec, cat *config.Catalog) *Handler {
	return &Handler{db: db, cat: cat}
}

// RunResponse — ответ POST /api/simulator/run.
type RunResponse struct {
	ID     string          `json:"id"`     // UUID последнего боя в battle_reports
	Stats  battle.SimStats `json:"stats"`  // агрегат по num_sim итераций
	Report battle.Report   `json:"report"` // последний бой целиком (для round-by-round)
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
	// План 87 / BA-007: симулятор всегда «бой на планете» (legacy:
	// SIM_PLANET_ID=1, planetid != 0). Опыт без штрафа bpc *= 0.5.
	in.HasPlanet = true
	n := in.NumSim
	if n < 1 {
		n = 1
	}
	if n > 100 {
		n = 100
	}

	// Заполняем cost из catalog (фронт его не передаёт — он не должен
	// знать backend-баланс). Без этого SimStats.AttackerLost* / Exp = 0.
	costByID := h.unitCosts()
	for si := range in.Attackers {
		fillCosts(in.Attackers[si].Units, costByID)
	}
	for si := range in.Defenders {
		fillCosts(in.Defenders[si].Units, costByID)
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

// unitCosts — индекс id → cost по объединённому catalog (ships+defense).
func (h *Handler) unitCosts() map[int]battle.UnitCost {
	out := make(map[int]battle.UnitCost, len(h.cat.Ships.Ships)+len(h.cat.Defense.Defense))
	for _, s := range h.cat.Ships.Ships {
		out[s.ID] = battle.UnitCost{
			Metal: s.Cost.Metal, Silicon: s.Cost.Silicon, Hydrogen: s.Cost.Hydrogen,
		}
	}
	for _, d := range h.cat.Defense.Defense {
		out[d.ID] = battle.UnitCost{
			Metal: d.Cost.Metal, Silicon: d.Cost.Silicon, Hydrogen: d.Cost.Hydrogen,
		}
	}
	return out
}

func fillCosts(units []battle.Unit, costs map[int]battle.UnitCost) {
	for i := range units {
		if c, ok := costs[units[i].UnitID]; ok {
			units[i].Cost = c
		}
	}
}
