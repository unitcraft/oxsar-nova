package auth

// DUPLICATE: этот файл скопирован между Go-модулями oxsar/game-nova,
// oxsar/auth и oxsar/portal. При любом изменении синхронизируйте КОПИИ:
//   - projects/game-nova/backend/internal/auth/jwksloader.go
//   - projects/auth/backend/internal/auth/jwksloader.go
//   - projects/portal/backend/internal/auth/jwksloader.go
// Причина дубля: каждый домен — отдельный go.mod, без shared-модуля.

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"oxsar/game-nova/pkg/jwtrs"
)

// LoadVerifier загружает публичный ключ RSA с JWKS URL и возвращает Verifier.
// Делает retry с экспоненциальной задержкой (как OpenPostgres).
func LoadVerifier(ctx context.Context, jwksURL string) (*jwtrs.Verifier, error) {
	// Нормализуем URL: если передан базовый URL без пути — добавляем
	if !strings.HasSuffix(jwksURL, ".json") && !strings.Contains(jwksURL, "/.well-known/") {
		jwksURL = strings.TrimRight(jwksURL, "/") + "/.well-known/jwks.json"
	}

	delay := 500 * time.Millisecond
	var lastErr error
	for attempt := 1; attempt <= 6; attempt++ {
		data, err := fetchJWKS(ctx, jwksURL)
		if err == nil {
			return jwtrs.NewVerifierFromJWKS(data)
		}
		lastErr = err
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
			delay *= 2
		}
	}
	return nil, fmt.Errorf("load jwks from %s: %w", jwksURL, lastErr)
}

func fetchJWKS(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jwks fetch status %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 64*1024))
}
