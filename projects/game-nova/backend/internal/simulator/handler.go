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
	"math"
	"net/http"

	"oxsar/game-nova/internal/artefact"
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

	// План 72.1.34 ч.C: применяем battle artefacts (legacy
	// `Simulator` строки 124-163, 6 артефактов с
	// effect_type=ARTEFACT_EFFECT_TYPE_BATTLE). Каждый артефакт
	// в Side.BattleArtefactIDs мапится на ArtefactSpec, аккумулируется
	// через ComputeBattleModifier (multiplier на attack/shield/shell),
	// применяется к юнитам стороны.
	h.applyBattleArtefacts(in.Attackers)
	h.applyBattleArtefacts(in.Defenders)

	stats, last, err := battle.MultiRun(in, n)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}

	// План 72.1.45 §6: реальная формула destroy_chance из Java Assault.java
	// L.937: `5 × DS_count^0.3` clamped to [0, 25]. Уровень здания-цели в
	// формулу не входит (legacy не использует его). Применимо только если
	// `target_building_id`/`level` заданы и атакующий выжил с ≥1 DS.
	// Записываем в last.BuildingDestroyChance + last.TargetDestroyed (chance
	// ≥ 25% → true для simulator-репорта; реальный бой бросает RNG).
	if in.TargetBuildingID > 0 && in.TargetBuildingLevel > 0 {
		const unitDeathstarID = 42
		var dsCount int64
		if last.Winner == "attackers" {
			for _, side := range last.Attackers {
				for _, u := range side.Units {
					if u.UnitID == unitDeathstarID {
						dsCount += u.QuantityEnd
					}
				}
			}
		}
		if dsCount > 0 {
			// 5 * pow(ds, 0.3) clamp [0, 25]
			chance := 5.0 * math.Pow(float64(dsCount), 0.3)
			if chance > 25.0 {
				chance = 25.0
			} else if chance < 0 {
				chance = 0
			}
			last.BuildingDestroyChance = chance
			last.TargetDestroyed = chance >= 25.0
		}
	}

	id := ids.New()
	reportJSON, _ := json.Marshal(last)
	// План 72.1.31: симуляции явно не имеют moon/alien-флагов (там
	// нет реальной планеты-цели). DEFAULT false в схеме покрывает,
	// но для будущей консистентности INSERT задаёт явные значения.
	_, err = h.db.Pool().Exec(r.Context(), `
		INSERT INTO battle_reports (
			id, attacker_user_id, defender_user_id, planet_id,
			seed, winner, rounds,
			debris_metal, debris_silicon,
			loot_metal, loot_silicon, loot_hydrogen,
			report, is_simulation,
			has_aliens, moon_created, is_moon
		) VALUES ($1, NULL, NULL, NULL, $2, $3, $4, $5, $6, 0, 0, 0, $7, true,
		          false, false, false)
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

// applyBattleArtefacts применяет battle_bonus от артефактов из
// Side.BattleArtefactIDs к юнитам. Bonus умножается на attack/shield/shell
// каждого юнита (legacy effect_type=ARTEFACT_EFFECT_TYPE_BATTLE).
//
// Пустой/несуществующий ID игнорируется. Stack >max_stacks обрезается
// (legacy soft-cap, у нас просто все ID применяются — пользователь сам
// проверяет лимит на UI).
func (h *Handler) applyBattleArtefacts(sides []battle.Side) {
	for si := range sides {
		ids := sides[si].BattleArtefactIDs
		if len(ids) == 0 {
			continue
		}
		specs := make([]config.ArtefactSpec, 0, len(ids))
		for _, id := range ids {
			for _, sp := range h.cat.Artefacts.Artefacts {
				if sp.ID == id {
					specs = append(specs, sp)
					break
				}
			}
		}
		if len(specs) == 0 {
			continue
		}
		mod := artefact.ComputeBattleModifier(specs)
		// Применяем к каждому юниту side.
		for ui := range sides[si].Units {
			u := &sides[si].Units[ui]
			u.Attack *= mod.AttackMul
			u.Shield *= mod.ShieldMul
			u.Shell *= mod.ShellMul
		}
	}
}
