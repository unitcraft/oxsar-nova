package portalsvc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// IdentityClient вызывает identity-service для server-to-server
// операций. План 72.2: получение handoff-токена для перехода
// портал → game-фронт.
//
// Использует RSA-JWT юзера (forwarded из Authorization header).
// Identity-service сам валидирует токен через JWKS-self-loop.
type IdentityClient struct {
	identityURL string
	httpClient  *http.Client
}

// NewIdentityClient создаёт клиент для identity-service.
func NewIdentityClient(identityURL string) *IdentityClient {
	return &IdentityClient{
		identityURL: identityURL,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

// ErrIdentityUnavailable — identity-service недоступен или вернул 5xx.
var ErrIdentityUnavailable = errors.New("portalsvc: identity unavailable")

// ErrUnauthorized — identity вернул 401 (токен битый/просрочен).
var ErrUnauthorized = errors.New("portalsvc: identity unauthorized")

// IssueHandoffToken запрашивает у identity-service одноразовый handoff-
// токен для переключения юзера в указанную вселенную. План 72.2.
//
// Семантика: identity-service кладёт в Redis ключ handoff:<token> со
// значением user_id (TTL 30s, single-use). Token потом обменивается
// game-фронтом через POST /auth/token/exchange.
//
// userToken — Authorization header (полностью, с "Bearer ").
// Возвращает handoff_token (строка); пустая строка только при ошибке.
func (c *IdentityClient) IssueHandoffToken(ctx context.Context, userToken, universeID string) (string, error) {
	if c.identityURL == "" {
		return "", errors.New("portalsvc: identity URL not configured")
	}
	if userToken == "" {
		return "", ErrUnauthorized
	}
	if universeID == "" {
		return "", errors.New("portalsvc: universe_id required")
	}

	body, _ := json.Marshal(map[string]string{"universe_id": universeID})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.identityURL+"/auth/universe-token", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", userToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrIdentityUnavailable, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// продолжаем ниже
	case http.StatusUnauthorized:
		return "", ErrUnauthorized
	default:
		return "", fmt.Errorf("%w: identity returned %d", ErrIdentityUnavailable, resp.StatusCode)
	}

	var out struct {
		HandoffToken string `json:"handoff_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if out.HandoffToken == "" {
		return "", fmt.Errorf("%w: empty handoff_token in response", ErrIdentityUnavailable)
	}
	return out.HandoffToken, nil
}
