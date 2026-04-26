package auth

import (
	"testing"
	"time"
)

func TestJWTRoundtrip(t *testing.T) {
	t.Parallel()
	j := NewJWTIssuer("test-secret", time.Minute, time.Hour)
	toks, err := j.Issue("user-1")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	uid, err := j.Parse(toks.Access, "access")
	if err != nil {
		t.Fatalf("parse access: %v", err)
	}
	if uid != "user-1" {
		t.Fatalf("expected user-1, got %s", uid)
	}
	if _, err := j.Parse(toks.Access, "refresh"); err == nil {
		t.Fatalf("access should not parse as refresh")
	}
}

func TestJWTRejectsBadSignature(t *testing.T) {
	t.Parallel()
	a := NewJWTIssuer("secret-a", time.Minute, time.Hour)
	b := NewJWTIssuer("secret-b", time.Minute, time.Hour)
	toks, _ := a.Issue("user-1")
	if _, err := b.Parse(toks.Access, "access"); err == nil {
		t.Fatalf("b must reject a's token")
	}
}
