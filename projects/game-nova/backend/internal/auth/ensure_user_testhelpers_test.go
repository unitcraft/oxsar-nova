package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"oxsar/game-nova/pkg/jwtrs"
)

// contextWithFakeRSAClaims кладёт в context фиктивные RSA-claims —
// эмулирует поведение RSAMiddleware без необходимости поднимать
// auth-service в тестах.
func contextWithFakeRSAClaims(ctx context.Context, userID, username string) context.Context {
	claims := &jwtrs.Claims{
		Username: username,
		Roles:    []string{"player"},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userID,
		},
	}
	return context.WithValue(ctx, rsaClaimsKey, claims)
}

func newUUID(t *testing.T) string {
	t.Helper()
	return uuid.NewString()
}

func randomHex(n int) string {
	b := make([]byte, n/2+1)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)[:n]
}

func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}
