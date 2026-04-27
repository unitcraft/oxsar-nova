package auth

// DUPLICATE: этот файл скопирован между Go-модулями oxsar/game-nova,
// oxsar/auth, oxsar/portal и oxsar/billing. При любом изменении
// синхронизируйте КОПИИ:
//   - projects/game-nova/backend/internal/auth/password_test.go
//   - projects/auth/backend/internal/auth/password_test.go
// Причина дубля: каждый домен — отдельный go.mod, без shared-модуля.

import "testing"

func TestHashAndVerify(t *testing.T) {
	t.Parallel()
	h, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	ok, err := VerifyPassword("correct horse battery staple", h)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !ok {
		t.Fatalf("expected verify ok")
	}
	bad, err := VerifyPassword("wrong password", h)
	if err != nil {
		t.Fatalf("verify bad: %v", err)
	}
	if bad {
		t.Fatalf("expected verify fail on wrong password")
	}
}

func TestVerifyInvalidFormat(t *testing.T) {
	t.Parallel()
	if _, err := VerifyPassword("x", "not-a-hash"); err == nil {
		t.Fatalf("expected parse error")
	}
}
