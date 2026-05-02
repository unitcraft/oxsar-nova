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
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
	// План 72.1.50 ч.5 (72.1.10 wave 3): legacy `Battlestats.class.php`
	// рендерит координаты + название планеты в каждой строке.
	// Источник — `battle_reports.planet_id` → `planets.name/galaxy/system/position`.
	// Может быть NULL для экспедиционных/ACS-битв без конкретного target_planet
	// и для legacy-записей до миграции 0009 (если такие есть).
	PlanetName *string `json:"planet_name,omitempty"`
	Galaxy     *int    `json:"galaxy,omitempty"`
	System     *int    `json:"system,omitempty"`
	Position   *int    `json:"position,omitempty"`
	IsMoonTarget *bool `json:"is_moon_target,omitempty"`
}

// ListMine GET /api/users/me/battles?limit=20&cursor=<at>
// ListMine GET /api/users/me/battles
//
// Query-параметры (план 72.1.10 ч.A3 — порт legacy
// Battlestats.class.php::showBattles):
//
//	limit, cursor              базовая пагинация (cursor = at-RFC3339Nano).
//	date_min, date_max         ограничение по at (RFC3339).
//	user_filter                uuid оппонента (атакер ИЛИ дефендер,
//	                            при условии что юзер тоже участник).
//	alliance_filter            uuid альянса оппонента.
//	show_drawn=false           скрыть ничьи (winner='draw').
//	show_aliens=true           включить бои с aliens
//	                            (battle_reports.has_aliens=true).
//	show_no_destroyed=false    скрыть «пустые» бои (без debris и loot).
//	new_moon=true              только бои с появлением луны
//	                            (moon_created=true).
//	moon_battle=true           только бои на луне (is_moon=true).
//	sort_field=date|rounds|debris|loot
//	sort_order=asc|desc
//
// Дефолты: show_drawn=true, show_aliens=false, show_no_destroyed=true,
// new_moon=false, moon_battle=false, sort_field=date, sort_order=desc
// (как в legacy `index()`).
func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	q := r.URL.Query()
	limit := 20
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	cursorAt := time.Now()
	if v := q.Get("cursor"); v != "" {
		if t, err := time.Parse(time.RFC3339Nano, v); err == nil {
			cursorAt = t
		}
	}

	// Динамический WHERE — собираем conditions + args. $1..$N
	// генерируется по индексу.
	args := []any{uid, cursorAt}
	// План 72.1.50 ч.5 (72.1.10 wave 3): SQL получил JOIN с planets,
	// поэтому все колонки battle_reports префиксированы `b.` для
	// устранения ambiguity (есть пересечения: is_moon, galaxy, system,
	// position у planets).
	conds := []string{
		"(b.attacker_user_id = $1 OR b.defender_user_id = $1)",
		"b.is_simulation = false",
		"b.at < $2",
	}
	addArg := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}

	// План 72.1.10 wave 2: legacy date-range clamping
	// (Battlestats.class.php:65-88). См. clampDateRange ниже.
	dateMin, dateMax := clampDateRange(q.Get("date_min"), q.Get("date_max"), time.Now().UTC())
	conds = append(conds, "b.at >= "+addArg(dateMin), "b.at <= "+addArg(dateMax))
	if v := q.Get("user_filter"); v != "" {
		ph := addArg(v)
		conds = append(conds,
			"(b.attacker_user_id = "+ph+" OR b.defender_user_id = "+ph+")")
	}
	if v := q.Get("alliance_filter"); v != "" {
		ph := addArg(v)
		// Оппонент = тот, кто НЕ uid. Альянс оппонента совпадает с
		// заданным.
		conds = append(conds, `EXISTS (
			SELECT 1 FROM users u
			WHERE u.alliance_id = `+ph+`
			  AND u.id IN (b.attacker_user_id, b.defender_user_id)
			  AND u.id != $1
		)`)
	}

	// Boolean-фильтры (по дефолту в legacy).
	showDrawn := parseBool(q.Get("show_drawn"), true)
	if !showDrawn {
		conds = append(conds, "b.winner != 'draw'")
	}
	showAliens := parseBool(q.Get("show_aliens"), false)
	if !showAliens {
		conds = append(conds, "b.has_aliens = false")
	}
	showNoDestroyed := parseBool(q.Get("show_no_destroyed"), true)
	if !showNoDestroyed {
		conds = append(conds,
			"(b.loot_metal + b.loot_silicon + b.loot_hydrogen + b.debris_metal + b.debris_silicon) > 0")
	}
	if parseBool(q.Get("new_moon"), false) {
		conds = append(conds, "b.moon_created = true")
	}
	if parseBool(q.Get("moon_battle"), false) {
		conds = append(conds, "b.is_moon = true")
	}

	sortField, sortOrder := sortClause(q.Get("sort_field"), q.Get("sort_order"))

	limitArg := addArg(limit)
	// План 72.1.50 ч.5 (72.1.10 wave 3): LEFT JOIN planets для
	// `planet_name` и координат в DTO. Условия WHERE остаются на
	// `battle_reports` (b.). Используем LEFT JOIN — если planet_id NULL
	// (legacy запись или экспедиция), строка всё равно попадает в результат.
	sql := `
		SELECT b.id, b.attacker_user_id, b.defender_user_id, b.winner, b.rounds,
		       b.debris_metal::bigint, b.debris_silicon::bigint,
		       b.loot_metal::bigint, b.loot_silicon::bigint, b.loot_hydrogen::bigint,
		       b.at,
		       p.name, p.galaxy, p.system, p.position, p.is_moon
		FROM battle_reports b
		LEFT JOIN planets p ON p.id = b.planet_id
		WHERE ` + strings.Join(conds, " AND ") + `
		ORDER BY ` + sortField + ` ` + sortOrder + `
		LIMIT ` + limitArg

	rows, err := h.db.Pool().Query(r.Context(), sql, args...)
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
			&it.PlanetName, &it.Galaxy, &it.System, &it.Position, &it.IsMoonTarget,
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

// parseBool — парсит query-флаг с дефолтом. Принимает "1"/"true"/
// "yes" как true; "0"/"false"/"no" как false; пусто/невалидно — def.
func parseBool(v string, def bool) bool {
	switch strings.ToLower(v) {
	case "1", "true", "yes":
		return true
	case "0", "false", "no":
		return false
	default:
		return def
	}
}

// hiddenDays / windowDays — legacy константы из
// Battlestats.class.php:65-68. На прод-сервере (OXSAR_RELEASED &&
// !DEATHMATCH && !isAdmin) последние 3 дня скрыты, окно — 15 дней
// перед скрытыми. Админ-bypass будет добавлен при появлении isAdmin
// контекста на этом эндпоинте (сейчас в nova нет).
const (
	battlesHiddenDays = 3
	battlesWindowDays = 15
)

// clampDateRange реализует legacy date-range clamping
// (Battlestats.class.php:65-88). Серверная защита: даже если клиент
// пришлёт более широкий диапазон, мы его сужаем. Пустые/невалидные
// строки — дефолтный полный диапазон [now-18d .. now-3d]. План 72.1.10
// wave 2.
func clampDateRange(rawMin, rawMax string, now time.Time) (time.Time, time.Time) {
	maxAllowed := now.Add(-time.Duration(battlesHiddenDays) * 24 * time.Hour)
	minAllowed := now.Add(-time.Duration(battlesHiddenDays+battlesWindowDays) * 24 * time.Hour)

	dMin := minAllowed
	if t, err := time.Parse(time.RFC3339, rawMin); err == nil {
		if t.Before(minAllowed) {
			t = minAllowed
		}
		if t.After(maxAllowed) {
			t = maxAllowed
		}
		dMin = t
	}
	dMax := maxAllowed
	if t, err := time.Parse(time.RFC3339, rawMax); err == nil {
		if t.After(maxAllowed) {
			t = maxAllowed
		}
		if t.Before(minAllowed) {
			t = minAllowed
		}
		dMax = t
	}
	if dMin.After(dMax) {
		dMin, dMax = dMax, dMin
	}
	return dMin, dMax
}

// sortClause whitelist'ит legacy sort-поля для ORDER BY. Возвращает
// SQL-фрагмент для поля и направление "ASC"/"DESC".
//
// План 72.1.50 ч.5 (72.1.10 wave 3): добавлено `planet_name` →
// `p.name` (legacy `Battlestats.class.php:96`). JOIN planets теперь
// всегда в SQL (для DTO-полей planet_name/galaxy/system/position),
// поэтому и сортировка по планете тоже доступна. NULL-safe ASC/DESC
// — NULLS LAST для DESC, NULLS FIRST для ASC (поведение postgres
// по умолчанию: NULLS FIRST для ASC, LAST для DESC). Для записей
// без planet_id (экспедиция / legacy) planet_name=NULL.
func sortClause(field, order string) (string, string) {
	col := "b.at"
	switch field {
	case "rounds":
		col = "b.rounds"
	case "debris":
		col = "(b.debris_metal + b.debris_silicon)"
	case "loot":
		col = "(b.loot_metal + b.loot_silicon + b.loot_hydrogen)"
	case "outcome":
		col = "b.winner"
	case "moon":
		col = "b.is_moon"
	case "planet_name":
		col = "p.name"
	}
	dir := "DESC"
	if order == "asc" {
		dir = "ASC"
	}
	return col, dir
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
