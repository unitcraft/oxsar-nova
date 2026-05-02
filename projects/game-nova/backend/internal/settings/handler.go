// Package settings реализует GET/PUT /api/settings — настройки аккаунта.
package settings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/httpx"
	"oxsar/game-nova/internal/i18n"
)

var validTimezones = map[string]bool{
	"UTC": true, "Europe/Moscow": true, "Europe/Kiev": true, "Europe/Minsk": true,
	"Asia/Yekaterinburg": true, "Asia/Novosibirsk": true, "Asia/Vladivostok": true,
	"Asia/Almaty": true, "Europe/Berlin": true, "Europe/London": true,
	"America/New_York": true, "America/Los_Angeles": true, "Asia/Tokyo": true,
}

type Handler struct {
	pool    *pgxpool.Pool
	automsg AutoMsgSender
	bundle  *i18n.Bundle
}

func NewHandler(pool *pgxpool.Pool) *Handler { return &Handler{pool: pool} }

func (h *Handler) WithBundle(b *i18n.Bundle) *Handler {
	h.bundle = b
	return h
}

func (h *Handler) tr(group, key string, vars map[string]string) string {
	if h.bundle == nil {
		return "[" + group + "." + key + "]"
	}
	return h.bundle.Tr(i18n.LangRu, group, key, vars)
}

type settingsResponse struct {
	Email        string  `json:"email"`
	Language     string  `json:"language"`
	Timezone     string  `json:"timezone"`
	VacationSince *string `json:"vacation_since"`
	// План 72.1.55 Task I (P72.S4.SETTINGS subset 1:1): legacy
	// preferences.tpl поля. Default values заданы миграцией 0097.
	ShowAllConstructions bool  `json:"show_all_constructions"`
	ShowAllResearch      bool  `json:"show_all_research"`
	ShowAllShipyard      bool  `json:"show_all_shipyard"`
	ShowAllDefense       bool  `json:"show_all_defense"`
	PlanetOrder          int16 `json:"planet_order"`
	// План 72.1.55.E (effects): esps — int 1..99 (legacy
	// `Preferences.class.php:407` ESPIONAGE_DRONES_DEFAULT clamp);
	// дефолтное число шпионских зондов в Mission spy form.
	Esps    int16 `json:"esps"`
	IpCheck bool  `json:"ipcheck"`
}

// Get GET /api/settings — текущие настройки.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}

	var email, language, timezone string
	var vacationSince *time.Time
	var resp settingsResponse
	err := h.pool.QueryRow(r.Context(), `
		SELECT email, language, timezone, vacation_since,
		       show_all_constructions, show_all_research, show_all_shipyard,
		       show_all_defense, planetorder, esps, ipcheck
		FROM users WHERE id = $1
	`, uid).Scan(&email, &language, &timezone, &vacationSince,
		&resp.ShowAllConstructions, &resp.ShowAllResearch, &resp.ShowAllShipyard,
		&resp.ShowAllDefense, &resp.PlanetOrder, &resp.Esps, &resp.IpCheck)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}

	resp.Email = email
	resp.Language = language
	resp.Timezone = timezone
	if vacationSince != nil {
		s := vacationSince.UTC().Format(time.RFC3339)
		resp.VacationSince = &s
	}
	httpx.WriteJSON(w, r, http.StatusOK, resp)
}

type updateRequest struct {
	Email    *string `json:"email"`
	Language *string `json:"language"`
	Timezone *string `json:"timezone"`
	// План 72.1.55 Task I (P72.S4.SETTINGS subset 1:1).
	ShowAllConstructions *bool  `json:"show_all_constructions"`
	ShowAllResearch      *bool  `json:"show_all_research"`
	ShowAllShipyard      *bool  `json:"show_all_shipyard"`
	ShowAllDefense       *bool  `json:"show_all_defense"`
	PlanetOrder          *int16 `json:"planet_order"`
	Esps                 *int16 `json:"esps"`
	IpCheck              *bool  `json:"ipcheck"`
}

// Update PUT /api/settings — обновить email, language, timezone.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}

	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}

	if req.Email != nil {
		email := strings.ToLower(strings.TrimSpace(*req.Email))
		if !strings.Contains(email, "@") || len(email) < 3 {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid email"))
			return
		}
		// План 72.1.49: AutoMsg при смене email (legacy
		// `Preferences.class.php:305` EMAIL_EMAIL_MESSAGE).
		// Атомарно: SET email + INSERT message в одной транзакции.
		if err := h.setEmailWithNotice(r.Context(), uid, email); err != nil {
			if strings.Contains(err.Error(), "unique") {
				httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, "email already taken"))
			} else {
				httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			}
			return
		}
	}

	if req.Language != nil {
		lang := *req.Language
		if lang != "ru" && lang != "en" {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "unsupported language"))
			return
		}
		if _, err := h.pool.Exec(r.Context(),
			`UPDATE users SET language = $1 WHERE id = $2`, lang, uid); err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			return
		}
	}

	if req.Timezone != nil {
		tz := *req.Timezone
		if !validTimezones[tz] {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "unsupported timezone"))
			return
		}
		if _, err := h.pool.Exec(r.Context(),
			`UPDATE users SET timezone = $1 WHERE id = $2`, tz, uid); err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			return
		}
	}

	// План 72.1.55 Task I (P72.S4.SETTINGS subset 1:1): bool/int prefs.
	// Простые UPDATE — нет сторонних effects (применение в UI/backend
	// отдельно, по мере подключения каждого preference).
	type kv struct {
		col   string
		val   any
		isSet bool
	}
	bools := []kv{
		{"show_all_constructions", req.ShowAllConstructions, req.ShowAllConstructions != nil},
		{"show_all_research", req.ShowAllResearch, req.ShowAllResearch != nil},
		{"show_all_shipyard", req.ShowAllShipyard, req.ShowAllShipyard != nil},
		{"show_all_defense", req.ShowAllDefense, req.ShowAllDefense != nil},
		{"ipcheck", req.IpCheck, req.IpCheck != nil},
	}
	for _, b := range bools {
		if !b.isSet {
			continue
		}
		// reflect-free: значение — *bool, разыменуем через type assertion.
		v := *(b.val.(*bool))
		if _, err := h.pool.Exec(r.Context(),
			"UPDATE users SET "+b.col+" = $1 WHERE id = $2", v, uid); err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			return
		}
	}
	if req.PlanetOrder != nil {
		v := *req.PlanetOrder
		if v < 0 || v > 2 {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "planet_order out of range (0..2)"))
			return
		}
		if _, err := h.pool.Exec(r.Context(),
			`UPDATE users SET planetorder = $1 WHERE id = $2`, v, uid); err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			return
		}
	}
	// План 72.1.55.E (effects): esps int 1..99 (legacy clamp).
	if req.Esps != nil {
		v := *req.Esps
		if v < 1 {
			v = 1
		}
		if v > 99 {
			v = 99
		}
		if _, err := h.pool.Exec(r.Context(),
			`UPDATE users SET esps = $1 WHERE id = $2`, v, uid); err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// План 36 Critical-6: ChangePassword переехал в identity-service (POST /auth/password).
// Хеш пароля живёт в identity-db, в game-db password_hash IS NULL.
// Frontend дёргает identity-service напрямую через vite proxy /auth/password.

func (h *Handler) setEmail(ctx context.Context, uid, email string) error {
	tag, err := h.pool.Exec(ctx,
		`UPDATE users SET email = $1 WHERE id = $2`, email, uid)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("user not found")
	}
	return nil
}

// setEmailWithNotice — план 72.1.49: атомарно меняет email и шлёт
// AutoMsg юзеру (legacy `Preferences.class.php::305 EMAIL_EMAIL_MESSAGE`).
// Если automsg не подключён — fallback на простую setEmail.
const settingsMessageFolder = 1 // системные сообщения (legacy `MSG_SYSTEM`).

func (h *Handler) setEmailWithNotice(ctx context.Context, uid, email string) error {
	if h.automsg == nil || h.bundle == nil {
		return h.setEmail(ctx, uid, email)
	}
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx,
		`UPDATE users SET email = $1 WHERE id = $2`, email, uid)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("user not found")
	}
	title := h.tr("autoMessages", "settingsEmailChanged.title", nil)
	body := h.tr("autoMessages", "settingsEmailChanged.body", map[string]string{
		"email": email,
	})
	if err := h.automsg.SendDirect(ctx, tx, uid, settingsMessageFolder, title, body); err != nil {
		return fmt.Errorf("automsg: %w", err)
	}
	return tx.Commit(ctx)
}
