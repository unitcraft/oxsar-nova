// Package universeswitcher реализует переключение вселенных для игрового сервера.
// Клиент вызывает GET /api/universes/switch?target=uni02, сервер проксирует запрос
// в auth-service для получения handoff-токена и возвращает redirect URL.
package universeswitcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/httpx"
	"oxsar/game-nova/internal/universe"
)

// Handler выдаёт handoff URL для переключения в другую вселенную.
type Handler struct {
	authServiceURL string // http://auth-service:9000
	universeID     string // текущая вселенная, например uni01
	reg            *universe.Registry
	httpClient     *http.Client
}

// New создаёт Handler.
func New(authServiceURL, universeID string, reg *universe.Registry) *Handler {
	return &Handler{
		authServiceURL: authServiceURL,
		universeID:     universeID,
		reg:            reg,
		httpClient:     &http.Client{Timeout: 10 * time.Second},
	}
}

// ListUniverses — GET /api/universes
// Возвращает список всех вселенных из реестра (публичный endpoint).
func (h *Handler) ListUniverses(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{
		"universes":   h.reg.All(),
		"current":     h.universeID,
	})
}

// SwitchUniverse — GET /api/universes/switch?target=<universeID>
// Требует аутентификации. Запрашивает handoff-токен у auth-service
// и возвращает URL для перехода в целевую вселенную.
func (h *Handler) SwitchUniverse(w http.ResponseWriter, r *http.Request) {
	if h.authServiceURL == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "universe switching not configured"))
		return
	}

	targetID := r.URL.Query().Get("target")
	if targetID == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "target universe_id required"))
		return
	}

	target, ok := h.reg.ByID(targetID)
	if !ok {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "unknown universe: "+targetID))
		return
	}
	if target.Status != "active" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "universe not active: "+targetID))
		return
	}

	userID, ok := auth.UserID(r.Context())
	if !ok || userID == "" {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}

	// Получаем Bearer токен из запроса, чтобы передать в auth-service
	authHeader := r.Header.Get("Authorization")

	handoffToken, err := h.requestHandoffToken(r.Context(), authHeader, targetID)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, "failed to get handoff token"))
		return
	}

	redirectURL := fmt.Sprintf("https://%s.oxsar-nova.ru/auth/handoff?code=%s",
		target.Subdomain, handoffToken)

	httpx.WriteJSON(w, r, http.StatusOK, map[string]string{
		"redirect_url":  redirectURL,
		"universe_id":   targetID,
		"universe_name": target.Name,
	})
}

func (h *Handler) requestHandoffToken(ctx context.Context, authHeader, universeID string) (string, error) {
	body, _ := json.Marshal(map[string]string{"universe_id": universeID})
	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		h.authServiceURL+"/auth/universe-token",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth-service returned %d", resp.StatusCode)
	}

	var out struct {
		HandoffToken string `json:"handoff_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.HandoffToken, nil
}
