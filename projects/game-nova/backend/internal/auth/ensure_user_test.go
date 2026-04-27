package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TestIsUsernameConflict — unit-тест классификации pg-ошибок.
func TestIsUsernameConflict(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"random", errors.New("connection refused"), false},
		{"only sqlstate", errors.New("ERROR: duplicate key (SQLSTATE 23505)"), false},
		{"only constraint name", errors.New("violates users_username_key"), false},
		{
			"real pg conflict",
			errors.New(`ERROR: duplicate key value violates unique constraint "users_username_key" (SQLSTATE 23505)`),
			true,
		},
		{
			"different unique (pkey)",
			errors.New(`ERROR: duplicate key value violates unique constraint "users_pkey" (SQLSTATE 23505)`),
			false,
		},
		{
			"email conflict (other UNIQUE)",
			errors.New(`ERROR: duplicate key value violates unique constraint "users_email_key" (SQLSTATE 23505)`),
			false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isUsernameConflict(c.err); got != c.want {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}

// TestEnsureUser_UsernameConflict_Integration — интеграционный тест.
// Проверяет полный flow:
//   1. Insert legacy-юзера в game-db (id=A, username=alice).
//   2. RSAMiddleware-симуляция: claims с другим id=B, тот же username=alice.
//   3. EnsureUserMiddleware → INSERT с id=B → UNIQUE conflict на username.
//   4. Middleware ловит, возвращает 409 (а не пускает дальше до handler).
//
// Требует BILLING_TEST_DB_URL... но это не billing — нужна game-db.
// Используем GAME_TEST_DB_URL для game-db.
func TestEnsureUser_UsernameConflict_Integration(t *testing.T) {
	dbURL := os.Getenv("GAME_TEST_DB_URL")
	if dbURL == "" {
		t.Skip("GAME_TEST_DB_URL not set; skipping integration test")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	// Уникальное имя для теста, чтобы не пересекаться с другими тестами и
	// существующими данными (lazy-create в общем тестовом стенде).
	username := "conflict_test_" + randomHex(8)

	// 1. Legacy-юзер: insert вручную с id=A.
	legacyID := newUUID(t)
	if _, err := pool.Exec(ctx, `
		INSERT INTO users (id, username, email, password_hash)
		VALUES ($1, $2, $3, $4)
	`, legacyID, username, "legacy@x.com", "$argon2id$dummy"); err != nil {
		t.Fatalf("insert legacy: %v", err)
	}
	defer func() {
		_, _ = pool.Exec(context.Background(),
			`DELETE FROM users WHERE username = $1`, username)
	}()

	// 2. Подготовим request с RSA-claims (id=B, тот же username).
	newID := newUUID(t)
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	// Кладём поддельные RSA-claims прямо в ctx через RSAMiddleware-ключ.
	// Используем internal-функцию: claims хранятся в `rsaClaimsKey`.
	// Для теста проще через локальный hack — вместо AuthMiddleware подсунем
	// claims напрямую.
	// (см. middleware.go — rsaClaimsKey не экспортирован; используем
	// контекст через типизированный helper, который есть рядом.)
	// jwtrs.Claims — структура из pkg/jwtrs.

	ctxWithClaims := contextWithFakeRSAClaims(req.Context(), newID, username)
	req = req.WithContext(ctxWithClaims)

	// 3. Запустим middleware. Handler-следующий должен НЕ вызваться,
	// если middleware вернёт 409.
	handlerCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})
	mw := EnsureUserMiddleware(EnsureUserConfig{
		Pool: pool,
	})
	rec := httptest.NewRecorder()
	mw(next).ServeHTTP(rec, req)

	// 4. Проверки.
	if handlerCalled {
		t.Errorf("next handler should NOT be called on username conflict")
	}
	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409 Conflict", rec.Code)
	}
	body := rec.Body.String()
	if !contains(body, "username already taken") && !contains(body, "conflict") {
		t.Errorf("body should mention username conflict, got: %s", body)
	}
}

// TestEnsureUser_CleanFlow_Integration — happy-path: новый юзер,
// никаких legacy-конфликтов, INSERT проходит, handler вызывается.
func TestEnsureUser_CleanFlow_Integration(t *testing.T) {
	dbURL := os.Getenv("GAME_TEST_DB_URL")
	if dbURL == "" {
		t.Skip("GAME_TEST_DB_URL not set; skipping integration test")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	username := "clean_test_" + randomHex(8)
	id := newUUID(t)
	defer func() {
		_, _ = pool.Exec(context.Background(),
			`DELETE FROM users WHERE id = $1`, id)
	}()

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req = req.WithContext(contextWithFakeRSAClaims(req.Context(), id, username))

	handlerCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})
	mw := EnsureUserMiddleware(EnsureUserConfig{
		Pool: pool,
	})
	rec := httptest.NewRecorder()
	mw(next).ServeHTTP(rec, req)

	if !handlerCalled {
		t.Error("next handler should be called on clean flow")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}

	// Проверим, что юзер действительно создан.
	var dbUsername string
	if err := pool.QueryRow(ctx,
		`SELECT username FROM users WHERE id = $1`, id).Scan(&dbUsername); err != nil {
		t.Fatalf("select after middleware: %v", err)
	}
	if dbUsername != username {
		t.Errorf("dbUsername=%q, want %q", dbUsername, username)
	}
}

// TestEnsureUser_RaceCondition_Integration — 50 параллельных запросов
// от одного юзера. Только один INSERT должен пройти (ON CONFLICT id),
// остальные — no-op. handler вызывается у всех.
func TestEnsureUser_RaceCondition_Integration(t *testing.T) {
	dbURL := os.Getenv("GAME_TEST_DB_URL")
	if dbURL == "" {
		t.Skip("GAME_TEST_DB_URL not set; skipping integration test")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	username := "race_test_" + randomHex(8)
	id := newUUID(t)
	defer func() {
		_, _ = pool.Exec(context.Background(),
			`DELETE FROM users WHERE id = $1`, id)
	}()

	mw := EnsureUserMiddleware(EnsureUserConfig{Pool: pool})
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	const N = 50
	results := make(chan int, N)
	for i := 0; i < N; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
			req = req.WithContext(contextWithFakeRSAClaims(req.Context(), id, username))
			rec := httptest.NewRecorder()
			mw(next).ServeHTTP(rec, req)
			results <- rec.Code
		}()
	}
	for i := 0; i < N; i++ {
		code := <-results
		if code != http.StatusOK {
			t.Errorf("[%d] status=%d, want 200 (race shouldn't break clean flow)", i, code)
		}
	}

	// В БД должна быть ровно одна запись.
	var count int
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM users WHERE id = $1`, id).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Errorf("count=%d, want 1 (race created duplicates)", count)
	}
}
