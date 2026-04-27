package authsvc

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"oxsar/identity/pkg/jwtrs"
)

// План 44 (152-ФЗ). Integration-тесты для consent flow и DeleteAccount.
//
// Конвенция проекта (как в game-nova/internal/auth/ensure_user_test.go):
// требует IDENTITY_TEST_DB_URL с прокатанными миграциями (0001..0004).
// Если переменная не задана — t.Skip; CI / docker-стенд должны её
// задавать.

func setupTestService(t *testing.T) (*Service, *pgxpool.Pool) {
	t.Helper()
	dbURL := os.Getenv("IDENTITY_TEST_DB_URL")
	if dbURL == "" {
		t.Skip("IDENTITY_TEST_DB_URL not set; skipping integration test")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa generate: %v", err)
	}
	iss := jwtrs.NewIssuer(key, 15*time.Minute, 30*24*time.Hour)
	return New(pool, iss), pool
}

func randHex(t *testing.T, n int) string {
	t.Helper()
	b := make([]byte, n/2+1)
	if _, err := rand.Read(b); err != nil {
		t.Fatalf("rand: %v", err)
	}
	return hex.EncodeToString(b)[:n]
}

// Лишний раз страхуемся: каждый тест работает на своём username/email,
// чтобы не конфликтовать ни с другими тестами, ни с лежащими в стенде
// данными. После теста — DELETE FROM users CASCADE на test-юзеру (FK
// ON DELETE CASCADE из 0001/0004 уберёт consent/oauth/refresh).
func cleanupUser(t *testing.T, pool *pgxpool.Pool, username string) {
	t.Helper()
	_, _ = pool.Exec(context.Background(),
		`DELETE FROM users WHERE username = $1`, username)
}

// TestRegister_NoConsent_Rejected — без consent_accepted Register должен
// вернуть ErrConsentRequired и НЕ создать запись users (откат транзакции
// не нужен — мы вообще не доходим до tx).
func TestRegister_NoConsent_Rejected(t *testing.T) {
	svc, pool := setupTestService(t)
	username := "consent_no_" + randHex(t, 8)
	t.Cleanup(func() { cleanupUser(t, pool, username) })

	_, _, err := svc.Register(context.Background(), RegisterInput{
		Username:        username,
		Email:           username + "@example.test",
		Password:        "password123",
		ConsentAccepted: false,
	})
	if !errors.Is(err, ErrConsentRequired) {
		t.Fatalf("err = %v, want ErrConsentRequired", err)
	}

	// Юзера в БД быть не должно.
	var n int
	if err := pool.QueryRow(context.Background(),
		`SELECT count(*) FROM users WHERE username = $1`, username).Scan(&n); err != nil {
		t.Fatalf("count users: %v", err)
	}
	if n != 0 {
		t.Errorf("users count = %d, want 0", n)
	}
}

// TestRegister_WithConsent_PersistsConsent — happy-path: consent_accepted=true
// и terms_accepted=true, валидный input → user создан; в user_consents две
// записи (pdn_processing + offer_acceptance) с правильными атрибутами.
func TestRegister_WithConsent_PersistsConsent(t *testing.T) {
	svc, pool := setupTestService(t)
	username := "consent_ok_" + randHex(t, 8)
	t.Cleanup(func() { cleanupUser(t, pool, username) })

	const wantIP = "203.0.113.42" // RFC5737 TEST-NET-3
	const wantUA = "oxsar-nova-test/1.0"

	beforeReg := time.Now().Add(-time.Second)
	u, toks, err := svc.Register(context.Background(), RegisterInput{
		Username:         username,
		Email:            username + "@example.test",
		Password:         "password123",
		ConsentAccepted:  true,
		TermsAccepted:    true,
		ConsentIP:        wantIP,
		ConsentUserAgent: wantUA,
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if u.ID == "" || toks.Access == "" {
		t.Fatalf("empty user/tokens: %+v %+v", u, toks)
	}

	// Проверяем обе записи в user_consents (pdn + offer).
	checkConsent := func(consentType, wantVersion string) {
		t.Helper()
		var (
			gotVersion string
			gotIP      string
			gotUA      string
			acceptedAt time.Time
		)
		err := pool.QueryRow(context.Background(), `
			SELECT consent_text_version, host(accepted_ip), accepted_user_agent, accepted_at
			FROM user_consents
			WHERE user_id = $1 AND consent_type = $2
		`, u.ID, consentType).Scan(&gotVersion, &gotIP, &gotUA, &acceptedAt)
		if err != nil {
			t.Fatalf("select consent %s: %v", consentType, err)
		}
		if gotVersion != wantVersion {
			t.Errorf("%s version = %q, want %q", consentType, gotVersion, wantVersion)
		}
		if gotIP != wantIP {
			t.Errorf("%s ip = %q, want %q", consentType, gotIP, wantIP)
		}
		if gotUA != wantUA {
			t.Errorf("%s ua = %q, want %q", consentType, gotUA, wantUA)
		}
		if acceptedAt.Before(beforeReg) {
			t.Errorf("%s accepted_at = %v, want >= %v", consentType, acceptedAt, beforeReg)
		}
	}
	checkConsent(ConsentTypePDNProcessing, PrivacyPolicyVersion)
	checkConsent(ConsentTypeOfferAcceptance, TermsVersion)
}

// TestRegister_NoTerms_Rejected — план 47: без TermsAccepted Register
// должен вернуть ErrTermsRequired и не создать ни user, ни consent.
func TestRegister_NoTerms_Rejected(t *testing.T) {
	svc, pool := setupTestService(t)
	username := "terms_no_" + randHex(t, 8)
	t.Cleanup(func() { cleanupUser(t, pool, username) })

	_, _, err := svc.Register(context.Background(), RegisterInput{
		Username:        username,
		Email:           username + "@example.test",
		Password:        "password123",
		ConsentAccepted: true,
		TermsAccepted:   false,
	})
	if !errors.Is(err, ErrTermsRequired) {
		t.Fatalf("err = %v, want ErrTermsRequired", err)
	}
	var n int
	if err := pool.QueryRow(context.Background(),
		`SELECT count(*) FROM users WHERE username = $1`, username).Scan(&n); err != nil {
		t.Fatalf("count users: %v", err)
	}
	if n != 0 {
		t.Errorf("users count = %d, want 0", n)
	}
}

// TestRegister_DuplicateUser_NoConsentLeaked — если username/email
// конфликтует, второй Register должен упасть с ErrUserExists и НЕ
// оставить осиротевшую запись consent (вся вставка в одной tx).
func TestRegister_DuplicateUser_NoConsentLeaked(t *testing.T) {
	svc, pool := setupTestService(t)
	username := "consent_dup_" + randHex(t, 8)
	t.Cleanup(func() { cleanupUser(t, pool, username) })

	in := RegisterInput{
		Username:        username,
		Email:           username + "@example.test",
		Password:        "password123",
		ConsentAccepted: true,
		TermsAccepted:   true,
	}
	if _, _, err := svc.Register(context.Background(), in); err != nil {
		t.Fatalf("first Register: %v", err)
	}

	// План 47: после успешной регистрации в user_consents должно быть
	// две записи (pdn_processing + offer_acceptance) на одного юзера.
	countConsents := func() int {
		t.Helper()
		var n int
		if err := pool.QueryRow(context.Background(), `
			SELECT count(*) FROM user_consents
			WHERE user_id = (SELECT id FROM users WHERE username = $1)
		`, username).Scan(&n); err != nil {
			t.Fatalf("count consents: %v", err)
		}
		return n
	}
	if got := countConsents(); got != 2 {
		t.Fatalf("after first Register consent count = %d, want 2", got)
	}

	// Повторная регистрация с тем же username — должна упасть и НЕ добавить
	// consent (tx откатится).
	_, _, err := svc.Register(context.Background(), in)
	if !errors.Is(err, ErrUserExists) {
		t.Fatalf("err = %v, want ErrUserExists", err)
	}
	if got := countConsents(); got != 2 {
		t.Errorf("after duplicate Register consent count = %d, want still 2", got)
	}
}

// TestDeleteAccount_AnonymizesAndIsIdempotent — DeleteAccount меняет
// email/username/password_hash и проставляет deleted_at; повторный
// вызов не падает и не меняет уже анонимизированные данные.
func TestDeleteAccount_AnonymizesAndIsIdempotent(t *testing.T) {
	svc, pool := setupTestService(t)
	username := "delete_me_" + randHex(t, 8)
	t.Cleanup(func() { cleanupUser(t, pool, username) })

	u, _, err := svc.Register(context.Background(), RegisterInput{
		Username:        username,
		Email:           username + "@example.test",
		Password:        "password123",
		ConsentAccepted: true,
		TermsAccepted:   true,
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	if err := svc.DeleteAccount(context.Background(), u.ID); err != nil {
		t.Fatalf("DeleteAccount: %v", err)
	}

	var (
		gotUsername string
		gotEmail    string
		gotHash     string
		deletedAt   *time.Time
	)
	err = pool.QueryRow(context.Background(), `
		SELECT username, email, password_hash, deleted_at
		FROM users WHERE id = $1
	`, u.ID).Scan(&gotUsername, &gotEmail, &gotHash, &deletedAt)
	if err != nil {
		t.Fatalf("select user: %v", err)
	}
	if deletedAt == nil {
		t.Errorf("deleted_at is NULL, want timestamp")
	}
	if gotHash != "" {
		t.Errorf("password_hash = %q, want empty", gotHash)
	}
	if gotUsername == username {
		t.Errorf("username not anonymized: %q", gotUsername)
	}
	if gotEmail == username+"@example.test" {
		t.Errorf("email not anonymized: %q", gotEmail)
	}

	// Идемпотентность: повторный DeleteAccount не должен вернуть ошибку
	// и не должен переписать уже анонимизированный username (UPDATE
	// фильтрует по deleted_at IS NULL).
	usernameAfterFirst := gotUsername
	if err := svc.DeleteAccount(context.Background(), u.ID); err != nil {
		t.Errorf("second DeleteAccount: %v", err)
	}
	var usernameAfterSecond string
	if err := pool.QueryRow(context.Background(),
		`SELECT username FROM users WHERE id = $1`, u.ID,
	).Scan(&usernameAfterSecond); err != nil {
		t.Fatalf("select after second delete: %v", err)
	}
	if usernameAfterSecond != usernameAfterFirst {
		t.Errorf("second DeleteAccount changed username: %q -> %q",
			usernameAfterFirst, usernameAfterSecond)
	}
}

// TestDeleteAccount_LoginRefuses — после удаления Login со старыми
// credentials возвращает ErrInvalidCredential (через WHERE deleted_at
// IS NULL в SELECT). Это критично: refresh-токены, выпущенные до
// удаления, тоже не должны обновляться.
func TestDeleteAccount_LoginRefuses(t *testing.T) {
	svc, pool := setupTestService(t)
	username := "delete_login_" + randHex(t, 8)
	t.Cleanup(func() { cleanupUser(t, pool, username) })

	const password = "password123"
	u, _, err := svc.Register(context.Background(), RegisterInput{
		Username:        username,
		Email:           username + "@example.test",
		Password:        password,
		ConsentAccepted: true,
		TermsAccepted:   true,
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	// До удаления — Login работает.
	if _, _, err := svc.Login(context.Background(), username, password); err != nil {
		t.Fatalf("Login before delete: %v", err)
	}

	if err := svc.DeleteAccount(context.Background(), u.ID); err != nil {
		t.Fatalf("DeleteAccount: %v", err)
	}

	// После удаления — Login падает с ErrInvalidCredential. Используем
	// именно username, потому что email уже не тот; но и username тоже
	// был анонимизирован, так что точное совпадение не пройдёт.
	if _, _, err := svc.Login(context.Background(), username, password); !errors.Is(err, ErrInvalidCredential) {
		t.Errorf("Login after delete: err = %v, want ErrInvalidCredential", err)
	}
}
