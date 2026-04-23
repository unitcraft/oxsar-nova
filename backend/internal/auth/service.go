package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/repo"
	"github.com/oxsar/nova/backend/pkg/ids"
)

// Ошибки доменного слоя. Handler переводит их в HTTP-статусы.
var (
	ErrUserExists        = errors.New("auth: user already exists")
	ErrInvalidCredential = errors.New("auth: invalid credentials")
	ErrUserNotFound      = errors.New("auth: user not found")
)

// StarterPlanetAssigner создаёт первую планету для игрока.
// Интерфейс (а не прямой импорт планеты) — чтобы auth не зависел от
// всего planet-пакета и не было циклов.
type StarterPlanetAssigner interface {
	Assign(ctx context.Context, userID string) (planetID string, err error)
}

// Service — регистрация и логин.
type Service struct {
	db       repo.Exec
	jwt      *JWTIssuer
	starter  StarterPlanetAssigner
	automsg  AutoMsgSender
	referral ReferralProcessor
}

// AutoMsgSender — опциональная зависимость для приветственных
// сообщений. Если nil — Register не отправляет WELCOME/STARTER_GUIDE.
type AutoMsgSender interface {
	Send(ctx context.Context, tx pgx.Tx, userID, key string, vars map[string]string) error
}

// ReferralProcessor — опциональная зависимость; обрабатывает реферальную
// регистрацию. Если nil — реф. бонусы не начисляются.
type ReferralProcessor interface {
	ProcessRegistration(ctx context.Context, newUserID, referrerID string) error
}

// NewService. starter/automsg/referral могут быть nil — тогда соответствующая
// подфункциональность не выполняется (полезно для unit-тестов auth).
func NewService(db repo.Exec, jwt *JWTIssuer, starter StarterPlanetAssigner, automsg AutoMsgSender) *Service {
	return &Service{db: db, jwt: jwt, starter: starter, automsg: automsg}
}

// WithReferral подключает реферальный процессор.
func (s *Service) WithReferral(r ReferralProcessor) *Service {
	s.referral = r
	return s
}

// RegisterInput — вход регистрации.
type RegisterInput struct {
	Username   string
	Email      string
	Password   string
	ReferredBy string // userID реферера (опционально)
}

// User — минимальная проекция для auth-слоя; полноценная модель живёт в
// пакете user.
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// Register создаёт пользователя и стартовую планету (планета — ответственность
// пакета planet, здесь только запись в users).
func (s *Service) Register(ctx context.Context, in RegisterInput) (User, Tokens, error) {
	username := strings.TrimSpace(in.Username)
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if len(username) < 3 || len(username) > 24 {
		return User{}, Tokens{}, fmt.Errorf("auth: username length 3..24")
	}
	if !strings.Contains(email, "@") {
		return User{}, Tokens{}, fmt.Errorf("auth: invalid email")
	}
	if len(in.Password) < 8 {
		return User{}, Tokens{}, fmt.Errorf("auth: password >= 8 chars")
	}

	hash, err := HashPassword(in.Password)
	if err != nil {
		return User{}, Tokens{}, err
	}

	userID := ids.New()
	err = s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO users (id, username, email, password_hash, created_at)
			VALUES ($1, $2, $3, $4, $5)
		`, userID, username, email, hash, time.Now().UTC())
		if err != nil {
			// pg unique violation = 23505
			if strings.Contains(err.Error(), "23505") {
				return ErrUserExists
			}
			return fmt.Errorf("insert user: %w", err)
		}
		return nil
	})
	if err != nil {
		return User{}, Tokens{}, err
	}

	// Выдача стартовой планеты. Делаем ПОСЛЕ фиксации пользователя,
	// а не внутри той же транзакции: если поиск координат почему-то
	// застопорится, пользователь уже создан, можно ретраить планету
	// отдельно. Игрок без планеты — валидное состояние (см. §5.15
	// vacation-режим и scenarios экспедиций с потерей последней
	// планеты).
	var planetName, planetCoords string
	if s.starter != nil {
		planetID, err := s.starter.Assign(ctx, userID)
		if err != nil {
			return User{}, Tokens{}, fmt.Errorf("assign starter planet: %w", err)
		}
		// Читаем имя и координаты для WELCOME-сообщения. Не critical —
		// если запрос упадёт, WELCOME просто не отправится.
		_ = s.db.Pool().QueryRow(ctx, `
			SELECT name, galaxy || ':' || system || ':' || position
			FROM planets WHERE id = $1
		`, planetID).Scan(&planetName, &planetCoords)
	}

	// Реферальная регистрация. Сбой не должен блокировать регистрацию.
	if s.referral != nil && in.ReferredBy != "" {
		_ = s.referral.ProcessRegistration(ctx, userID, in.ReferredBy)
	}

	// Автомесседжи (WELCOME + STARTER_GUIDE). Отправка вне транзакции
	// (tx=nil): сбой в message не должен откатывать регистрацию.
	if s.automsg != nil {
		vars := map[string]string{
			"username":    username,
			"planet_name": planetName,
			"coords":      planetCoords,
		}
		_ = s.automsg.Send(ctx, nil, userID, "WELCOME", vars)
		_ = s.automsg.Send(ctx, nil, userID, "STARTER_GUIDE", nil)
	}

	toks, err := s.jwt.Issue(userID)
	if err != nil {
		return User{}, Tokens{}, err
	}
	return User{ID: userID, Username: username, Email: email}, toks, nil
}

// Login проверяет (email или username)+password и возвращает токены.
func (s *Service) Login(ctx context.Context, email, password string) (User, Tokens, error) {
	login := strings.ToLower(strings.TrimSpace(email))

	var id, username, emailRead, hash string
	var bannedAt *time.Time
	err := s.db.Pool().QueryRow(ctx, `
		SELECT id, username, email, password_hash, banned_at FROM users
		WHERE email = $1 OR lower(username) = $1
	`, login).Scan(&id, &username, &emailRead, &hash, &bannedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, Tokens{}, ErrInvalidCredential
		}
		return User{}, Tokens{}, fmt.Errorf("select user: %w", err)
	}
	if bannedAt != nil {
		return User{}, Tokens{}, errors.New("auth: account banned")
	}

	ok, err := VerifyPassword(password, hash)
	if err != nil {
		return User{}, Tokens{}, err
	}
	if !ok {
		return User{}, Tokens{}, ErrInvalidCredential
	}

	toks, err := s.jwt.Issue(id)
	if err != nil {
		return User{}, Tokens{}, err
	}
	return User{ID: id, Username: username, Email: emailRead}, toks, nil
}

// Refresh выпускает новые токены по валидному refresh-токену.
func (s *Service) Refresh(refresh string) (Tokens, error) {
	uid, err := s.jwt.Parse(refresh, "refresh")
	if err != nil {
		return Tokens{}, ErrInvalidCredential
	}
	return s.jwt.Issue(uid)
}
