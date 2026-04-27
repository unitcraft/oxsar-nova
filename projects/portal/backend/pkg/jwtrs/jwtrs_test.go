package jwtrs_test

// DUPLICATE: этот файл скопирован между Go-модулями oxsar/game-nova,
// oxsar/auth и oxsar/portal. При любом изменении синхронизируйте КОПИИ:
//   - projects/game-nova/backend/pkg/jwtrs/jwtrs_test.go
//   - projects/auth/backend/pkg/jwtrs/jwtrs_test.go
//   - projects/portal/backend/pkg/jwtrs/jwtrs_test.go
// Причина дубля: каждый домен — отдельный go.mod, без shared-модуля.

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"testing"
	"time"

	"oxsar/portal/pkg/jwtrs"
)

func newTestIssuer(t *testing.T) (*jwtrs.Issuer, *jwtrs.Verifier) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	iss := jwtrs.NewIssuer(key, 15*time.Minute, 30*24*time.Hour)
	ver := jwtrs.NewVerifierFromKey(iss.PublicKey())
	return iss, ver
}

func TestIssueAndParse(t *testing.T) {
	iss, ver := newTestIssuer(t)

	in := jwtrs.IssueInput{
		UserID:          "user-123",
		Username:        "StarLord",
		GlobalCredits:   500,
		ActiveUniverses: []string{"uni01"},
		Roles:           []string{"player"},
	}
	toks, err := iss.Issue(in)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	claims, err := ver.Parse(toks.Access, "access")
	if err != nil {
		t.Fatalf("parse access: %v", err)
	}
	if claims.Subject != "user-123" {
		t.Errorf("subject = %q, want user-123", claims.Subject)
	}
	if claims.Username != "StarLord" {
		t.Errorf("username = %q, want StarLord", claims.Username)
	}
	if claims.GlobalCredits != 500 {
		t.Errorf("credits = %d, want 500", claims.GlobalCredits)
	}
	if len(claims.ActiveUniverses) != 1 || claims.ActiveUniverses[0] != "uni01" {
		t.Errorf("universes = %v, want [uni01]", claims.ActiveUniverses)
	}

	// Refresh token must parse as "refresh"
	_, err = ver.Parse(toks.Refresh, "refresh")
	if err != nil {
		t.Fatalf("parse refresh: %v", err)
	}

	// Wrong kind must fail
	_, err = ver.Parse(toks.Access, "refresh")
	if err == nil {
		t.Error("expected error for wrong kind")
	}
}

func TestJWKSRoundTrip(t *testing.T) {
	iss, _ := newTestIssuer(t)

	jwks := jwtrs.IssuerToJWKS(iss)
	data, err := json.Marshal(jwks)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	ver, err := jwtrs.NewVerifierFromJWKS(data)
	if err != nil {
		t.Fatalf("verifier from jwks: %v", err)
	}

	in := jwtrs.IssueInput{UserID: "u1", Username: "x"}
	toks, err := iss.Issue(in)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	claims, err := ver.Parse(toks.Access, "access")
	if err != nil {
		t.Fatalf("parse after jwks round-trip: %v", err)
	}
	if claims.Subject != "u1" {
		t.Errorf("subject = %q", claims.Subject)
	}
}

func TestHandoffToken(t *testing.T) {
	h := jwtrs.NewHandoffToken("user-42")
	if h.Token == "" {
		t.Error("empty token")
	}
	if h.UserID != "user-42" {
		t.Errorf("userID = %q", h.UserID)
	}
	if h.ExpiresAt.Before(time.Now()) {
		t.Error("already expired")
	}
}
