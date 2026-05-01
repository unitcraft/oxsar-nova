// Package monitor — endpoint /api/monitor-planet — мониторинг
// активностей чужой планеты через здание STAR_SURVEILLANCE
// (план 72.1.20, legacy `MonitorPlanet.class.php`).
//
// Логика по шагам (legacy паритет):
//  1. Принять planet_id в query.
//  2. Найти главную луну сканера (с UNIT_STAR_SURVEILLANCE >=1, та же
//     галактика что и target.galaxy).
//  3. Делегировать TransportService.Phalanx — он списывает водород,
//     проверяет range, возвращает все флоты в системе.
//  4. Отфильтровать только те, где src_planet или dst_planet = target.
//  5. Если target.user_id != scanner и target.spyware >= scanner.spyware/2:
//     отправить AutoMsg `surveillanceDetected` цели в folder=11.
//
// Это самостоятельный endpoint, не пересекающийся с фалангой по
// системе (/api/phalanx — system-level), потому что legacy /MonitorPlanet
// — planet-level (фильтр src/dst = target.id).
package monitor

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/automsg"
	"oxsar/game-nova/internal/fleet"
	"oxsar/game-nova/internal/httpx"
	"oxsar/game-nova/internal/i18n"
	"oxsar/game-nova/internal/repo"
)

// unitSpyware — id исследования UNIT_SPYWARE (legacy const = 13).
const unitSpyware = 13

type Handler struct {
	db        repo.Exec
	transport *fleet.TransportService
	automsg   *automsg.Service
	bundle    *i18n.Bundle
}

func NewHandler(db repo.Exec, t *fleet.TransportService) *Handler {
	return &Handler{db: db, transport: t}
}

// WithAutoMsg подключает automsg для уведомления цели о детектированном
// сканировании (legacy `MSG_SURVEILLANCE_DETECTED`, folder=11).
// Если nil — Monitor работает без уведомлений.
func (h *Handler) WithAutoMsg(am *automsg.Service) *Handler {
	h.automsg = am
	return h
}

// WithBundle — i18n для текста AutoMsg.
func (h *Handler) WithBundle(b *i18n.Bundle) *Handler {
	h.bundle = b
	return h
}

// MonitorResult — DTO ответа.
type MonitorResult struct {
	TargetPlanet planetInfo         `json:"target_planet"`
	Scanner      planetInfo         `json:"scanner"`
	Events       []fleet.PhalanxScan `json:"events"`
	Detected     bool               `json:"detected"` // отправлен ли surveillance-msg цели
}

type planetInfo struct {
	PlanetID  string `json:"planet_id"`
	Name      string `json:"name"`
	UserID    string `json:"user_id"`
	Username  string `json:"username,omitempty"`
	Galaxy    int    `json:"galaxy"`
	System    int    `json:"system"`
	Position  int    `json:"position"`
}

// Monitor GET /api/monitor-planet?id=<planet_id>
//
// Возвращает события флота связанные с указанной планетой (src или dst).
// Использует scan через Phalanx (списывает 5000H с источника).
func (h *Handler) Monitor(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	planetID := r.URL.Query().Get("id")
	if planetID == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "id required"))
		return
	}

	ctx := r.Context()

	// 1. Target планета.
	var target planetInfo
	err := h.db.Pool().QueryRow(ctx, `
		SELECT p.id, p.name, p.user_id, COALESCE(u.username,''),
		       p.galaxy, p.system, p.position
		FROM planets p
		LEFT JOIN users u ON u.id = p.user_id
		WHERE p.id = $1 AND p.destroyed_at IS NULL
	`, planetID).Scan(&target.PlanetID, &target.Name, &target.UserID,
		&target.Username, &target.Galaxy, &target.System, &target.Position)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.WriteError(w, r, httpx.ErrNotFound)
			return
		}
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}

	// 2. Найти scanner — луну текущего юзера в той же галактике с
	// UNIT_STAR_SURVEILLANCE >=1. Берём первую попавшуюся (если их
	// несколько — UI потом может позволить выбор).
	var scanner planetInfo
	err = h.db.Pool().QueryRow(ctx, `
		SELECT p.id, p.name, p.user_id, '' AS username,
		       p.galaxy, p.system, p.position
		FROM planets p
		JOIN buildings b ON b.planet_id = p.id AND b.unit_id = 55
		WHERE p.user_id = $1 AND p.is_moon = true
		  AND p.destroyed_at IS NULL AND p.galaxy = $2
		  AND b.level >= 1
		ORDER BY b.level DESC, p.created_at ASC
		LIMIT 1
	`, uid, target.Galaxy).Scan(&scanner.PlanetID, &scanner.Name,
		&scanner.UserID, &scanner.Username, &scanner.Galaxy,
		&scanner.System, &scanner.Position)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest,
				"no moon with star_surveillance in target galaxy"))
			return
		}
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}

	// 3. Phalanx-скан системы (он спишет 5000H, проверит range и т.д.).
	events, err := h.transport.Phalanx(ctx, uid, scanner.PlanetID, target.Galaxy, target.System)
	if err != nil {
		// Используем те же error mappings что fleet/handler.go::Phalanx.
		switch {
		case errors.Is(err, fleet.ErrPhalanxNotAMoon),
			errors.Is(err, fleet.ErrPhalanxNotInstalled),
			errors.Is(err, fleet.ErrPhalanxDifferentGalax):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		case errors.Is(err, fleet.ErrPhalanxOutOfRange),
			errors.Is(err, fleet.ErrPhalanxNoHydrogen):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		default:
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		}
		return
	}

	// 4. Отфильтровать события только связанные с target.PlanetID.
	// Phalanx возвращает по координатам, но src_planet_id / dst_planet_id
	// в DTO нет — есть только координаты. Сравниваем по target.galaxy/
	// system/position.
	filtered := make([]fleet.PhalanxScan, 0, len(events))
	for _, ev := range events {
		matchesSrc := ev.SrcGalaxy == target.Galaxy &&
			ev.SrcSystem == target.System &&
			ev.SrcPosition == target.Position
		matchesDst := ev.DstGalaxy == target.Galaxy &&
			ev.DstSystem == target.System &&
			ev.DstPosition == target.Position
		if matchesSrc || matchesDst {
			filtered = append(filtered, ev)
		}
	}

	// 5. Surveillance-detected: legacy `MonitorPlanet.class.php:255`
	// если target.user_id != scanner.user_id и
	// target.spyware >= scanner.spyware / 2 → AutoMsg в folder=11.
	detected := false
	if target.UserID != "" && target.UserID != uid && h.automsg != nil && h.bundle != nil {
		var targetSpyware, myySpyware int
		_ = h.db.Pool().QueryRow(ctx,
			`SELECT COALESCE(level,0) FROM research WHERE user_id=$1 AND unit_id=$2`,
			target.UserID, unitSpyware).Scan(&targetSpyware)
		_ = h.db.Pool().QueryRow(ctx,
			`SELECT COALESCE(level,0) FROM research WHERE user_id=$1 AND unit_id=$2`,
			uid, unitSpyware).Scan(&myySpyware)
		if targetSpyware*2 >= myySpyware { // эквивалент target >= my/2 без деления
			lang := h.userLang(ctx, target.UserID)
			vars := map[string]string{
				"galaxy":   strconv.Itoa(target.Galaxy),
				"system":   strconv.Itoa(target.System),
				"position": strconv.Itoa(target.Position),
				"planet":   target.Name,
			}
			title := h.bundle.Tr(lang, "autoMessages", "surveillanceDetected.title", vars)
			body := h.bundle.Tr(lang, "autoMessages", "surveillanceDetected.body", vars)
			if err := h.automsg.SendDirect(ctx, nil, target.UserID, automsg.FolderSurveillance, title, body); err == nil {
				detected = true
			}
		}
	}

	out := MonitorResult{
		TargetPlanet: target,
		Scanner:      scanner,
		Events:       filtered,
		Detected:     detected,
	}
	httpx.WriteJSON(w, r, http.StatusOK, out)
}

func (h *Handler) userLang(ctx context.Context, userID string) i18n.Lang {
	var lang string
	_ = h.db.Pool().QueryRow(ctx,
		`SELECT language FROM users WHERE id=$1`, userID).Scan(&lang)
	if lang == "" {
		return i18n.LangRu
	}
	return i18n.Lang(lang)
}
