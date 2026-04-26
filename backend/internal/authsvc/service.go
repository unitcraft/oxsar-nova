// Package authsvc реализует логику Auth Service: регистрация, логин,
// refresh, управление global credits.
package authsvc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oxsar/nova/backend/internal/auth"
	"github.com/oxsar/nova/backend/internal/repo"
	"github.com/oxsar/nova/backend/pkg/ids"
	"github.com/oxsar/nova/backend/pkg/jwtrs"
)

var (
	ErrUserExists        = errors.New("authsvc: user already exists")
	ErrInvalidCredential = errors.New("authsvc: invalid credentials")
	ErrUserBanned        = errors.New("authsvc: account banned")
	ErrInsufficientFunds = errors.New("authsvc: insufficient credits")
)

// User — публичная проекция пользователя.
type User struct {
	ID            string   `json:"id"`
	Username      string   `json:"username"`
	Email         string   `json:"email"`
	GlobalCredits int64    `json:"global_credits"`
	Roles         []string `json:"roles"`
}

// RegisterInput — вход регистрации.
type RegisterInput struct {
	Username string
	Email    string
	Password string
}

// SpendInput — списание кредитов.
type SpendInput struct {
	UserID string
	Amount int64
	Reason string // "feedback_vote" | "universe_purchase" | ...
	RefID  string // id предложения, вселенной и т.п.
}

// CreditTx — одна транзакция кредитов.
type CreditTx struct {
	ID        string    `json:"id"`
	Delta     int64     `json:"delta"`
	Reason    string    `json:"reason"`
	RefID     string    `json:"ref_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Service — основной сервис Auth Service.
type Service struct {
	db  *repo.PG
	iss *jwtrs.Issuer
}

// New создаёт Service.
func New(pool *pgxpool.Pool, iss *jwtrs.Issuer) *Service {
	return &Service{db: repo.New(pool), iss: iss}
}

// Register создаёт нового пользователя.
func (s *Service) Register(ctx context.Context, in RegisterInput) (User, jwtrs.Tokens, error) {
	username := strings.TrimSpace(in.Username)
	email := strings.ToLower(strings.TrimSpace(in.Email))

	if len(username) < 3 || len(username) > 24 {
		return User{}, jwtrs.Tokens{}, fmt.Errorf("authsvc: username length 3..24")
	}
	if !strings.Contains(email, "@") {
		return User{}, jwtrs.Tokens{}, fmt.Errorf("authsvc: invalid email")
	}
	if len(in.Password) < 8 {
		return User{}, jwtrs.Tokens{}, fmt.Errorf("authsvc: password >= 8 chars")
	}

	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return User{}, jwtrs.Tokens{}, err
	}

	userID := ids.New()
	err = s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO users (id, username, email, password_hash, created_at)
			VALUES ($1, $2, $3, $4, now())
		`, userID, username, email, hash)
		if err != nil {
			if strings.Contains(err.Error(), "23505") {
				return ErrUserExists
			}
			return fmt.Errorf("insert user: %w", err)
		}
		return nil
	})
	if err != nil {
		return User{}, jwtrs.Tokens{}, err
	}

	u := User{ID: userID, Username: username, Email: email, GlobalCredits: 0, Roles: []string{"player"}}
	toks, err := s.issueTokens(ctx, u)
	if err != nil {
		return User{}, jwtrs.Tokens{}, err
	}
	return u, toks, nil
}

// Login проверяет email+password и возвращает токены.
func (s *Service) Login(ctx context.Context, login, password string) (User, jwtrs.Tokens, error) {
	login = strings.ToLower(strings.TrimSpace(login))

	var u User
	var hash string
	var bannedAt, deletedAt *time.Time
	err := s.db.Pool().QueryRow(ctx, `
		SELECT id, username, email, password_hash, global_credits, roles, banned_at, deleted_at
		FROM users WHERE email = $1 OR lower(username) = $1
	`, login).Scan(&u.ID, &u.Username, &u.Email, &hash, &u.GlobalCredits, &u.Roles, &bannedAt, &deletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, jwtrs.Tokens{}, ErrInvalidCredential
		}
		return User{}, jwtrs.Tokens{}, fmt.Errorf("select user: %w", err)
	}
	if deletedAt != nil {
		return User{}, jwtrs.Tokens{}, ErrInvalidCredential
	}
	if bannedAt != nil {
		return User{}, jwtrs.Tokens{}, ErrUserBanned
	}

	ok, err := auth.VerifyPassword(password, hash)
	if err != nil {
		return User{}, jwtrs.Tokens{}, err
	}
	if !ok {
		return User{}, jwtrs.Tokens{}, ErrInvalidCredential
	}

	toks, err := s.issueTokens(ctx, u)
	if err != nil {
		return User{}, jwtrs.Tokens{}, err
	}
	return u, toks, nil
}

// Refresh выпускает новые токены по валидному refresh-токену.
func (s *Service) Refresh(ctx context.Context, refreshToken string) (jwtrs.Tokens, error) {
	ver := jwtrs.NewVerifierFromKey(s.iss.PublicKey())
	claims, err := ver.Parse(refreshToken, "refresh")
	if err != nil {
		return jwtrs.Tokens{}, ErrInvalidCredential
	}

	var u User
	err = s.db.Pool().QueryRow(ctx, `
		SELECT id, username, email, global_credits, roles
		FROM users WHERE id = $1 AND deleted_at IS NULL AND banned_at IS NULL
	`, claims.Subject).Scan(&u.ID, &u.Username, &u.Email, &u.GlobalCredits, &u.Roles)
	if err != nil {
		return jwtrs.Tokens{}, ErrInvalidCredential
	}

	return s.issueTokens(ctx, u)
}

// GetUser возвращает профиль пользователя по ID.
func (s *Service) GetUser(ctx context.Context, userID string) (User, error) {
	var u User
	err := s.db.Pool().QueryRow(ctx, `
		SELECT id, username, email, global_credits, roles
		FROM users WHERE id = $1 AND deleted_at IS NULL
	`, userID).Scan(&u.ID, &u.Username, &u.Email, &u.GlobalCredits, &u.Roles)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, fmt.Errorf("authsvc: user not found")
		}
		return User{}, fmt.Errorf("get user: %w", err)
	}
	return u, nil
}

// GetActiveUniverses возвращает список вселенных, где есть аккаунт пользователя.
// Читает из таблицы universe_memberships, которую заполняют игровые серверы
// при lazy join.
func (s *Service) GetActiveUniverses(ctx context.Context, userID string) ([]string, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT universe_id FROM universe_memberships WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("query universes: %w", err)
	}
	defer rows.Close()
	var universes []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		universes = append(universes, id)
	}
	return universes, rows.Err()
}

// SpendCredits атомарно списывает кредиты и записывает транзакцию.
func (s *Service) SpendCredits(ctx context.Context, in SpendInput) error {
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `
			UPDATE users SET global_credits = global_credits - $1
			WHERE id = $2 AND global_credits >= $1
		`, in.Amount, in.UserID)
		if err != nil {
			return fmt.Errorf("spend credits: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return ErrInsufficientFunds
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO credit_transactions (id, user_id, delta, reason, ref_id, created_at)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, now())
		`, in.UserID, -in.Amount, in.Reason, in.RefID)
		return err
	})
}

// AddCredits зачисляет кредиты (при оплате или другом начислении).
func (s *Service) AddCredits(ctx context.Context, userID string, amount int64, reason, refID string) error {
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			UPDATE users SET global_credits = global_credits + $1 WHERE id = $2
		`, amount, userID)
		if err != nil {
			return fmt.Errorf("add credits: %w", err)
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO credit_transactions (id, user_id, delta, reason, ref_id, created_at)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, now())
		`, userID, amount, reason, refID)
		return err
	})
}

// CreditBalance возвращает актуальный баланс.
func (s *Service) CreditBalance(ctx context.Context, userID string) (int64, error) {
	var balance int64
	err := s.db.Pool().QueryRow(ctx,
		`SELECT global_credits FROM users WHERE id = $1`, userID,
	).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("credit balance: %w", err)
	}
	return balance, nil
}

// CreditHistory возвращает историю транзакций кредитов.
func (s *Service) CreditHistory(ctx context.Context, userID string, limit, offset int) ([]CreditTx, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := s.db.Pool().Query(ctx, `
		SELECT id, delta, reason, COALESCE(ref_id,''), created_at
		FROM credit_transactions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("credit history: %w", err)
	}
	defer rows.Close()
	var result []CreditTx
	for rows.Next() {
		var tx CreditTx
		if err := rows.Scan(&tx.ID, &tx.Delta, &tx.Reason, &tx.RefID, &tx.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, tx)
	}
	return result, rows.Err()
}

// issueTokens выпускает токены для пользователя.
func (s *Service) issueTokens(ctx context.Context, u User) (jwtrs.Tokens, error) {
	universes, _ := s.GetActiveUniverses(ctx, u.ID)
	return s.iss.Issue(jwtrs.IssueInput{
		UserID:          u.ID,
		Username:        u.Username,
		GlobalCredits:   u.GlobalCredits,
		ActiveUniverses: universes,
		Roles:           u.Roles,
	})
}
