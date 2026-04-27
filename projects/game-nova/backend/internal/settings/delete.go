package settings

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/httpx"
)

const (
	deletionCodeLen   = 8
	deletionTTL       = 10 * time.Minute
	deletionMaxAtt    = 5
	deletionRateLimit = 3 // запросов кода в час
)

// алфавит без похожих символов (0/O, 1/I/l).
var codeAlphabet = []byte("ABCDEFGHJKMNPQRSTUVWXYZ23456789")

var (
	errCodeExpired    = errors.New("deletion code expired")
	errCodeInvalid    = errors.New("deletion code invalid")
	errTooManyAttempts = errors.New("too many attempts")
	errNoCode         = errors.New("deletion code not requested")
	errRateLimit      = errors.New("too many code requests")
)

// AutoMsgSender — узкий интерфейс для отправки сообщения с кодом.
type AutoMsgSender interface {
	SendDirect(ctx context.Context, tx pgx.Tx, userID string, folder int, title, body string) error
}

// WithAutoMsg подключает автомессадж-сервис для доставки кодов.
func (h *Handler) WithAutoMsg(a AutoMsgSender) *Handler {
	h.automsg = a
	return h
}

// RequestDeletionCode POST /api/me/deletion/code — генерирует одноразовый код,
// пишет хэш в БД, отправляет код пользователю системным сообщением.
func (h *Handler) RequestDeletionCode(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}

	// Rate-limit: 3 запроса в час по issued_at.
	var recentCount int
	if err := h.pool.QueryRow(r.Context(), `
		SELECT COUNT(*) FROM account_deletion_codes
		WHERE user_id = $1 AND issued_at > now() - interval '1 hour'
	`, uid).Scan(&recentCount); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	if recentCount >= deletionRateLimit {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, errRateLimit.Error()))
		return
	}

	code, err := generateCode()
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	hash, err := auth.HashPassword(code)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	expiresAt := time.Now().Add(deletionTTL)

	_, err = h.pool.Exec(r.Context(), `
		INSERT INTO account_deletion_codes (user_id, code_hash, issued_at, expires_at, attempts)
		VALUES ($1, $2, now(), $3, 0)
		ON CONFLICT (user_id) DO UPDATE SET
			code_hash = EXCLUDED.code_hash,
			issued_at = now(),
			expires_at = EXCLUDED.expires_at,
			attempts = 0
	`, uid, hash, expiresAt)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}

	// Отправить код системным сообщением (folder=13 SYSTEM).
	if h.automsg != nil {
		title := h.tr("settings", "deletionCode.title", nil)
		body := h.tr("settings", "deletionCode.body", map[string]string{
			"code":      code,
			"expiresAt": expiresAt.Format("15:04 02.01.2006"),
		})
		_ = h.automsg.SendDirect(r.Context(), nil, uid, 13, title, body)
	}

	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{
		"expires_at": expiresAt.UTC().Format(time.RFC3339),
		"ttl_seconds": int(deletionTTL.Seconds()),
	})
}

// ConfirmDeletion DELETE /api/me body: {"code":"XXXXXXXX"} — проверяет код и
// выполняет soft-delete.
func (h *Handler) ConfirmDeletion(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	if req.Code == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "code required"))
		return
	}

	err := h.performDeletion(r.Context(), uid, req.Code)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, errNoCode):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "code not requested"))
	case errors.Is(err, errCodeExpired):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "code expired"))
	case errors.Is(err, errTooManyAttempts):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "too many attempts, request new code"))
	case errors.Is(err, errCodeInvalid):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid code"))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

func (h *Handler) performDeletion(ctx context.Context, uid, code string) error {
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback(ctx)

	var codeHash string
	var expiresAt time.Time
	var attempts int
	err = tx.QueryRow(ctx, `
		SELECT code_hash, expires_at, attempts FROM account_deletion_codes
		WHERE user_id = $1 FOR UPDATE
	`, uid).Scan(&codeHash, &expiresAt, &attempts)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errNoCode
		}
		return fmt.Errorf("read code: %w", err)
	}

	if expiresAt.Before(time.Now()) {
		_, _ = tx.Exec(ctx, `DELETE FROM account_deletion_codes WHERE user_id = $1`, uid)
		if err := tx.Commit(ctx); err != nil {
			return err
		}
		return errCodeExpired
	}

	if attempts >= deletionMaxAtt {
		_, _ = tx.Exec(ctx, `DELETE FROM account_deletion_codes WHERE user_id = $1`, uid)
		if err := tx.Commit(ctx); err != nil {
			return err
		}
		return errTooManyAttempts
	}

	ok, err := auth.VerifyPassword(code, codeHash)
	if err != nil {
		return fmt.Errorf("verify: %w", err)
	}
	if !ok {
		if _, err := tx.Exec(ctx, `
			UPDATE account_deletion_codes SET attempts = attempts + 1 WHERE user_id = $1
		`, uid); err != nil {
			return fmt.Errorf("bump attempts: %w", err)
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
		return errCodeInvalid
	}

	// Успех: soft-delete.
	if _, err := tx.Exec(ctx, `
		UPDATE users SET
			deleted_at = now(),
			username = '[deleted_' || substr(id::text, 1, 8) || ']',
			email = '[deleted_' || substr(id::text, 1, 8) || ']',
			alliance_id = NULL
		WHERE id = $1
	`, uid); err != nil {
		return fmt.Errorf("soft-delete users: %w", err)
	}
	// Закрыть открытые лоты игрока — ресурсы остаются в таблице,
	// игровая логика (market.Accept/Cancel) их игнорирует по owner_id.
	if _, err := tx.Exec(ctx, `
		UPDATE market_lots SET state = 'cancelled'
		WHERE seller_id = $1 AND state = 'open'
	`, uid); err != nil {
		// market_lots может отсутствовать в некоторых окружениях — не критично.
		_ = err
	}
	// Удалить код.
	if _, err := tx.Exec(ctx, `DELETE FROM account_deletion_codes WHERE user_id = $1`, uid); err != nil {
		return fmt.Errorf("delete code: %w", err)
	}

	return tx.Commit(ctx)
}

// generateCode возвращает 8-символьный код из безопасного алфавита.
func generateCode() (string, error) {
	buf := make([]byte, deletionCodeLen)
	raw := make([]byte, deletionCodeLen)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	for i, b := range raw {
		buf[i] = codeAlphabet[int(b)%len(codeAlphabet)]
	}
	return string(buf), nil
}
