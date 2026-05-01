// Package empire реализует GET /api/empire — сводная таблица всех планет игрока.
package empire

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/httpx"
)

// PlanetRow — данные одной планеты для empire-таблицы.
type PlanetRow struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Galaxy   int    `json:"galaxy"`
	System   int    `json:"system"`
	Position int    `json:"position"`
	IsMoon   bool   `json:"is_moon"`

	Diameter   int `json:"diameter"`
	UsedFields int `json:"used_fields"`
	TempMin    int `json:"temp_min"`
	TempMax    int `json:"temp_max"`

	Metal    float64 `json:"metal"`
	Silicon  float64 `json:"silicon"`
	Hydrogen float64 `json:"hydrogen"`

	// План 72.1.45: УМИ — virtual research lab level. Legacy
	// `Planet::reseach_virt_lab` (Planet.class.php:806-845):
	// research_lab.level × max(1, moon_lab×10) + IGR-pool bonus.
	UMI int `json:"umi"`

	Buildings map[int]int   `json:"buildings"` // unit_id → level
	Ships     map[int]int64 `json:"ships"`      // unit_id → count
	Defense   map[int]int64 `json:"defense"`    // unit_id → count
}

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler { return &Handler{pool: pool} }

// GetAll GET /api/empire — все планеты игрока с зданиями, флотом,
// обороной + общие исследования (research per-user).
//
// План 72.1.37: legacy `Empire.class.php` показывает 5 вкладок —
// constructions, shipyard, defense, moon, research. Все данные
// уже агрегируются в loadPlanets (buildings/ships/defense per-planet)
// и добавляется `research` (single block для user, не per-planet).
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}

	planets, err := h.loadPlanets(r.Context(), uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	research, err := h.loadResearch(r.Context(), uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{
		"planets":  planets,
		"research": research,
	})
}

// loadResearch возвращает map unit_id → level для всех исследований
// игрока (legacy: общая для всех планет колонка в empire-таблице).
func (h *Handler) loadResearch(ctx context.Context, userID string) (map[int]int, error) {
	out := make(map[int]int)
	rows, err := h.pool.Query(ctx,
		`SELECT unit_id, level FROM research WHERE user_id=$1 AND level > 0`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var uid, lvl int
		if err := rows.Scan(&uid, &lvl); err != nil {
			return nil, err
		}
		out[uid] = lvl
	}
	return out, rows.Err()
}

func (h *Handler) loadPlanets(ctx context.Context, userID string) ([]PlanetRow, error) {
	// 1. Загружаем базовые данные планет.
	rows, err := h.pool.Query(ctx, `
		SELECT id, name, galaxy, system, position, is_moon,
		       diameter, used_fields, temperature_min, temperature_max,
		       COALESCE(metal, 0), COALESCE(silicon, 0), COALESCE(hydrogen, 0)
		FROM planets
		WHERE user_id = $1 AND destroyed_at IS NULL
		ORDER BY is_moon, galaxy, system, position
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var planets []PlanetRow
	var ids []string
	for rows.Next() {
		var p PlanetRow
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Galaxy, &p.System, &p.Position, &p.IsMoon,
			&p.Diameter, &p.UsedFields, &p.TempMin, &p.TempMax,
			&p.Metal, &p.Silicon, &p.Hydrogen,
		); err != nil {
			return nil, err
		}
		p.Buildings = make(map[int]int)
		p.Ships = make(map[int]int64)
		p.Defense = make(map[int]int64)
		planets = append(planets, p)
		ids = append(ids, p.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(planets) == 0 {
		return []PlanetRow{}, nil
	}

	// Индекс для быстрого поиска по ID.
	idx := make(map[string]*PlanetRow, len(planets))
	for i := range planets {
		idx[planets[i].ID] = &planets[i]
	}

	// 2. Здания.
	bRows, err := h.pool.Query(ctx, `
		SELECT planet_id, unit_id, level FROM buildings
		WHERE planet_id = ANY($1) AND level > 0
	`, ids)
	if err != nil {
		return nil, err
	}
	defer bRows.Close()
	for bRows.Next() {
		var pid string
		var uid, level int
		if err := bRows.Scan(&pid, &uid, &level); err != nil {
			return nil, err
		}
		if p, ok := idx[pid]; ok {
			p.Buildings[uid] = level
		}
	}
	if err := bRows.Err(); err != nil {
		return nil, err
	}

	// 3. Флот.
	sRows, err := h.pool.Query(ctx, `
		SELECT planet_id, unit_id, count FROM ships
		WHERE planet_id = ANY($1) AND count > 0
	`, ids)
	if err != nil {
		return nil, err
	}
	defer sRows.Close()
	for sRows.Next() {
		var pid string
		var uid int
		var count int64
		if err := sRows.Scan(&pid, &uid, &count); err != nil {
			return nil, err
		}
		if p, ok := idx[pid]; ok {
			p.Ships[uid] = count
		}
	}
	if err := sRows.Err(); err != nil {
		return nil, err
	}

	// 4. Оборона.
	dRows, err := h.pool.Query(ctx, `
		SELECT planet_id, unit_id, count FROM defense
		WHERE planet_id = ANY($1) AND count > 0
	`, ids)
	if err != nil {
		return nil, err
	}
	defer dRows.Close()
	for dRows.Next() {
		var pid string
		var uid int
		var count int64
		if err := dRows.Scan(&pid, &uid, &count); err != nil {
			return nil, err
		}
		if p, ok := idx[pid]; ok {
			p.Defense[uid] = count
		}
	}
	if err := dRows.Err(); err != nil {
		return nil, err
	}

	// 5. План 72.1.45: УМИ (research_virt_lab) per-planet —
	// legacy `Planet::reseach_virt_lab`.
	//   IGR=0 → research_lab.level (уровень лабы этой планеты).
	//   IGR>0 → SUM top-(1+ign) (lab.level × max(1, moon_lab.level × 10))
	//           по всем планетам user'а, где первой берётся текущая.
	//
	// Для Empire-обзора достаточно простой формулы: УМИ = research_lab
	// конкретной планеты × max(1, moon_lab×10), где moon_lab — на
	// связанной луне (если есть). Эта формула совпадает с per-planet
	// вкладом легаси при IGR=0 и при IGR>0 (bound внутри суммы).
	const unitResearchLab = 12
	const unitMoonLab = 41
	var ignLevel int
	_ = h.pool.QueryRow(ctx,
		`SELECT COALESCE(level, 0) FROM research WHERE user_id=$1 AND unit_id=113`,
		userID,
	).Scan(&ignLevel)

	for i := range planets {
		labLvl := planets[i].Buildings[unitResearchLab]
		// moon_lab у связанной луны (legacy: для не-луны ищем луну в той же позиции).
		moonLabLvl := 0
		if !planets[i].IsMoon {
			for j := range planets {
				if planets[j].IsMoon &&
					planets[j].Galaxy == planets[i].Galaxy &&
					planets[j].System == planets[i].System &&
					planets[j].Position == planets[i].Position {
					moonLabLvl = planets[j].Buildings[unitMoonLab]
					break
				}
			}
		}
		multiplier := 1
		if moonLabLvl*10 > 1 {
			multiplier = moonLabLvl * 10
		}
		planets[i].UMI = labLvl * multiplier
	}

	// Если IGR > 0 — добавляем top-(1+ign) бонус из других планет.
	// Для UI достаточно показать «свой + bonus». Bonus = сумма
	// labLvl × moonMul у других топовых планет.
	if ignLevel > 0 {
		// Сортируем планет-вклады DESC.
		type contrib struct {
			idx int
			val int
		}
		all := make([]contrib, 0, len(planets))
		for i, p := range planets {
			if p.IsMoon {
				continue
			}
			all = append(all, contrib{idx: i, val: p.UMI})
		}
		// Простая сортировка пузырьком DESC (n маленькое).
		for i := 0; i < len(all); i++ {
			for j := i + 1; j < len(all); j++ {
				if all[j].val > all[i].val {
					all[i], all[j] = all[j], all[i]
				}
			}
		}
		// Top (1+ign). Для каждой планеты её UMI = swap into top + bonus.
		topCount := 1 + ignLevel
		if topCount > len(all) {
			topCount = len(all)
		}
		for i := range planets {
			if planets[i].IsMoon {
				continue
			}
			// Вклад текущей планеты + top-(ign) из остальных.
			selfVal := planets[i].UMI
			sum := selfVal
			added := 0
			for _, c := range all {
				if c.idx == i {
					continue
				}
				if added >= ignLevel {
					break
				}
				sum += c.val
				added++
			}
			_ = topCount
			planets[i].UMI = sum
		}
	}

	return planets, nil
}
