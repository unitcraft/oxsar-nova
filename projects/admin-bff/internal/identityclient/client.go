// Package identityclient — HTTP-клиент к identity-service.
//
// admin-bff не знает про логику identity, только обменивается с ним:
//   - POST /auth/login    — username+password → access+refresh tokens.
//   - POST /auth/refresh  — refresh-token → новые access+refresh tokens.
//   - POST /auth/logout   — invalidate refresh-token.
//
// JWT не парсим тут (claims уже есть в ответе identity либо мы
// извлекаем их без проверки подписи — identity-сервис уже подписал).
package identityclient

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

// Tokens — пара access+refresh с metadata.
type Tokens struct {
	AccessToken    string    `json:"access_token"`
	RefreshToken   string    `json:"refresh_token"`
	AccessTokenExp time.Time `json:"-"`
	Subject        string    `json:"-"`
	Username       string    `json:"-"`
	Roles          []string  `json:"-"`
	Permissions    []string  `json:"-"`
}

// Login — username+password → tokens. Заполняет parsed claims из access JWT.
func (c *Client) Login(ctx context.Context, username, password string) (*Tokens, error) {
	body, _ := json.Marshal(map[string]string{
		"login":    username,
		"password": password,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/auth/login", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("identity login: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrInvalidCredentials
	}
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("identity login: %d %s", resp.StatusCode, string(raw))
	}
	var t Tokens
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return nil, fmt.Errorf("decode login response: %w", err)
	}
	if err := parseClaims(&t); err != nil {
		return nil, err
	}
	return &t, nil
}

// Refresh — обмен refresh-token на новую пару (rotation).
func (c *Client) Refresh(ctx context.Context, refreshToken string) (*Tokens, error) {
	body, _ := json.Marshal(map[string]string{"refresh_token": refreshToken})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/auth/refresh", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("identity refresh: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrInvalidToken
	}
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("identity refresh: %d %s", resp.StatusCode, string(raw))
	}
	var t Tokens
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return nil, fmt.Errorf("decode refresh response: %w", err)
	}
	if err := parseClaims(&t); err != nil {
		return nil, err
	}
	return &t, nil
}

// Logout — приглашает identity отозвать refresh-token.
// Best-effort: ошибки логируются вызывающим, но не блокируют local-cleanup.
func (c *Client) Logout(ctx context.Context, refreshToken string) error {
	body, _ := json.Marshal(map[string]string{"refresh_token": refreshToken})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/auth/logout", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("identity logout: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 500 {
		return fmt.Errorf("identity logout: %d", resp.StatusCode)
	}
	return nil
}

// parseClaims — извлекает sub/username/roles/permissions/exp из access JWT.
// Подпись НЕ проверяем — токен только что выдан тем же identity, и трогаем
// его внутри trusted-zone. Frontend всё равно видит только summary, не сам JWT.
func parseClaims(t *Tokens) error {
	parts := strings.Split(t.AccessToken, ".")
	if len(parts) != 3 {
		return fmt.Errorf("malformed access token")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("decode access payload: %w", err)
	}
	var c struct {
		Sub         string   `json:"sub"`
		Username    string   `json:"username"`
		Roles       []string `json:"roles"`
		Permissions []string `json:"permissions"`
		Exp         int64    `json:"exp"`
	}
	if err := json.Unmarshal(payload, &c); err != nil {
		return fmt.Errorf("unmarshal access claims: %w", err)
	}
	t.Subject = c.Sub
	t.Username = c.Username
	t.Roles = c.Roles
	t.Permissions = c.Permissions
	if c.Exp > 0 {
		t.AccessTokenExp = time.Unix(c.Exp, 0)
	}
	return nil
}
