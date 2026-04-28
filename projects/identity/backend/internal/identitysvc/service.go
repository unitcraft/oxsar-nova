// Package identitysvc реализует логику Auth Service: регистрация, логин,
// refresh, управление global credits.
package identitysvc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"oxsar/identity/internal/auth"
	"oxsar/identity/internal/moderation"
	"oxsar/identity/internal/repo"
	"oxsar/identity/pkg/ids"
	"oxsar/identity/pkg/jwtrs"
)

var (
	ErrUserExists        = errors.New("identitysvc: user already exists")
	ErrInvalidCredential = errors.New("identitysvc: invalid credentials")
	ErrUserBanned        = errors.New("identitysvc: account banned")
	ErrConsentRequired   = errors.New("identitysvc: pdn consent required")
	// ErrTermsRequired — план 47: пользователь не принял Договор-оферту,
	// Правила игры и Политику возврата.
	ErrTermsRequired = errors.New("identitysvc: terms acceptance required")
	// ErrUsernameForbidden — план 46 (149-ФЗ): имя содержит запрещённый
	// корень из UGC-blacklist.
	ErrUsernameForbidden = errors.New("identitysvc: username forbidden")
)

// PrivacyPolicyVersion — версия Политики конфиденциальности, которую
// принимает пользователь чекбоксом при регистрации (план 44, 152-ФЗ).
// Поднимаем при существенных изменениях документа — старые согласия
// останутся в БД с прежней версией для аудита.
const PrivacyPolicyVersion = "1.0"

// TermsVersion — единая версия пакета документов оферты (Договор-оферта,
// Правила игры, Политика возврата) — план 47. Версионируем пакет
// единым тегом, потому что они логически меняются вместе. При
// существенном изменении любого из трёх — поднимаем версию.
const TermsVersion = "1.0"

// ConsentTypePDNProcessing — обработка персональных данных по 152-ФЗ.
const ConsentTypePDNProcessing = "pdn_processing"

// ConsentTypeOfferAcceptance — акцепт пакета документов оферты
// (Договор-оферта, Правила игры, Политика возврата) — план 47.
const ConsentTypeOfferAcceptance = "offer_acceptance"

// User — публичная проекция пользователя.
//
// План 38 Ф.5: GlobalCredits удалён — баланс в billing-service
// (GET /billing/wallet/balance). Чтобы получить вместе с профилем —
// клиент делает 2 запроса (или backend frontend-агрегатор делает 1).
type User struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
}

// RegisterInput — вход регистрации.
//
// План 44 (152-ФЗ): ConsentAccepted — флаг согласия на обработку ПДн.
// План 47: TermsAccepted — флаг акцепта Договора-оферты, Правил игры и
// Политики возврата. Оба согласия обязательны и фиксируются отдельными
// записями в user_consents (consent_type = pdn_processing /
// offer_acceptance) для аудита.
// ConsentIP / ConsentUserAgent — атрибуты согласия (заполняет HTTP handler).
type RegisterInput struct {
	Username         string
	Email            string
	Password         string
	ConsentAccepted  bool
	TermsAccepted    bool
	ConsentIP        string
	ConsentUserAgent string
}

// План 38 Ф.5: SpendInput, CreditTx, ErrInsufficientFunds удалены —
// перенесены в billing-service.

// Service — основной сервис Auth Service.
type Service struct {
	db        *repo.PG
	iss       *jwtrs.Issuer
	blacklist *moderation.Blacklist
	rbac      *RBACService
}

// New создаёт Service.
func New(pool *pgxpool.Pool, iss *jwtrs.Issuer) *Service {
	return &Service{db: repo.New(pool), iss: iss}
}

// WithBlacklist подключает blacklist для проверки никнеймов (план 46).
// Если nil — проверка отключена (на dev/test допустимо).
func (s *Service) WithBlacklist(bl *moderation.Blacklist) *Service {
	s.blacklist = bl
	return s
}

// WithRBAC подключает RBAC service для обогащения JWT-claims
// permissions при выпуске токенов (план 52 Ф.2).
func (s *Service) WithRBAC(rbac *RBACService) *Service {
	s.rbac = rbac
	return s
}

// Register создаёт нового пользователя.
//
// План 44 (152-ФЗ): без явного согласия на обработку ПДн регистрация
// запрещена. Согласие фиксируется в user_consents в одной транзакции
// с insert users — нельзя оказаться с пользователем без согласия.
//
// План 47: дополнительно требуется акцепт Договора-оферты, Правил игры и
// Политики возврата (TermsAccepted). Оба согласия пишутся в user_consents
// разными записями (pdn_processing + offer_acceptance) в той же
// транзакции, что и users — атомарность сохраняется.
func (s *Service) Register(ctx context.Context, in RegisterInput) (User, jwtrs.Tokens, error) {
	username := strings.TrimSpace(in.Username)
	email := strings.ToLower(strings.TrimSpace(in.Email))

	if !in.ConsentAccepted {
		return User{}, jwtrs.Tokens{}, ErrConsentRequired
	}
	if !in.TermsAccepted {
		return User{}, jwtrs.Tokens{}, ErrTermsRequired
	}
	if len(username) < 3 || len(username) > 24 {
		return User{}, jwtrs.Tokens{}, fmt.Errorf("identitysvc: username length 3..24")
	}
	// План 46 (149-ФЗ): проверка никнейма по UGC-blacklist.
	if s.blacklist != nil {
		if forbidden, _ := s.blacklist.IsForbidden(username); forbidden {
			return User{}, jwtrs.Tokens{}, ErrUsernameForbidden
		}
	}
	if !strings.Contains(email, "@") {
		return User{}, jwtrs.Tokens{}, fmt.Errorf("identitysvc: invalid email")
	}
	if len(in.Password) < 8 {
		return User{}, jwtrs.Tokens{}, fmt.Errorf("identitysvc: password >= 8 chars")
	}

	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return User{}, jwtrs.Tokens{}, err
	}

	// accepted_ip — INET в Postgres; пустая строка не прокатит, передаём NULL.
	var consentIP any
	if in.ConsentIP != "" {
		consentIP = in.ConsentIP
	}
	var consentUA any
	if in.ConsentUserAgent != "" {
		consentUA = in.ConsentUserAgent
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
		_, err = tx.Exec(ctx, `
			INSERT INTO user_consents
				(user_id, consent_type, consent_text_version, accepted_ip, accepted_user_agent)
			VALUES ($1, $2, $3, $4, $5)
		`, userID, ConsentTypePDNProcessing, PrivacyPolicyVersion, consentIP, consentUA)
		if err != nil {
			return fmt.Errorf("insert pdn consent: %w", err)
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO user_consents
				(user_id, consent_type, consent_text_version, accepted_ip, accepted_user_agent)
			VALUES ($1, $2, $3, $4, $5)
		`, userID, ConsentTypeOfferAcceptance, TermsVersion, consentIP, consentUA)
		if err != nil {
			return fmt.Errorf("insert offer consent: %w", err)
		}
		return nil
	})
	if err != nil {
		return User{}, jwtrs.Tokens{}, err
	}

	u := User{ID: userID, Username: username, Email: email, Roles: []string{"player"}}
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
		SELECT id, username, email, password_hash, roles, banned_at, deleted_at
		FROM users WHERE email = $1 OR lower(username) = $1
	`, login).Scan(&u.ID, &u.Username, &u.Email, &hash, &u.Roles, &bannedAt, &deletedAt)
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
		SELECT id, username, email, roles
		FROM users WHERE id = $1 AND deleted_at IS NULL AND banned_at IS NULL
	`, claims.Subject).Scan(&u.ID, &u.Username, &u.Email, &u.Roles)
	if err != nil {
		return jwtrs.Tokens{}, ErrInvalidCredential
	}

	return s.issueTokens(ctx, u)
}

// GetUser возвращает профиль пользователя по ID.
func (s *Service) GetUser(ctx context.Context, userID string) (User, error) {
	var u User
	err := s.db.Pool().QueryRow(ctx, `
		SELECT id, username, email, roles
		FROM users WHERE id = $1 AND deleted_at IS NULL
	`, userID).Scan(&u.ID, &u.Username, &u.Email, &u.Roles)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, fmt.Errorf("identitysvc: user not found")
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

// План 38 Ф.5: SpendCredits/AddCredits/CreditBalance/CreditHistory удалены.
// Кошельки в billing-service (см. internal/billing/wallet.go).

// DeleteAccount анонимизирует пользователя в identity-db (план 44, ст. 14 152-ФЗ).
//
// По 152-ФЗ достаточно деперсонализации (ст. 3 п. 9), а не физического
// удаления записи: email и username заменяются на технические значения,
// password_hash обнуляется (вход больше невозможен), deleted_at = now()
// — что отсекает пользователя в Login/Refresh/GetUser (см. WHERE-условия).
//
// Идемпотентность: если пользователь уже удалён — return без ошибки.
//
// Связанные данные (refresh_tokens, oauth_accounts, credit_transactions,
// universe_memberships, user_consents) удалятся каскадом по ON DELETE
// CASCADE — но мы ничего не DELETE-аем, чтобы не разрывать FK-инварианты
// и сохранить аудит-историю. Refresh-токены становятся невалидными, потому
// что при Refresh пользователь отсекается через deleted_at IS NULL.
//
// Полное удаление игровых объектов (планет, флотов, рейтингов) — задача
// игровых сервисов; в identity-service владелец становится «удалённым», и они
// сами разруливают (story-параллельно flow в game-nova/settings/delete.go,
// который анонимизирует игровую таблицу users).
//
// Уникальность анонимизированного username/email: используем последний
// блок UUIDv7 (12 hex символов = 48 случайных бит). Префикс UUIDv7
// — это millisecond timestamp, два юзера, созданные в ту же миллисекунду,
// дают одинаковый prefix и ловят 23505 на UNIQUE(username) при коротком
// `id[:8]`. Последний блок — гарантированно случаен.
func (s *Service) DeleteAccount(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("identitysvc: empty user id")
	}
	// UUIDv7: tttttttt-tttt-7xxx-yxxx-xxxxxxxxxxxx — берём последний блок.
	idx := strings.LastIndex(userID, "-")
	if idx < 0 || idx >= len(userID)-1 {
		return fmt.Errorf("identitysvc: invalid user id")
	}
	suffix := userID[idx+1:]
	tag, err := s.db.Pool().Exec(ctx, `
		UPDATE users SET
			deleted_at    = now(),
			username      = '[deleted_' || $2 || ']',
			email         = '[deleted_' || $2 || ']',
			password_hash = ''
		WHERE id = $1 AND deleted_at IS NULL
	`, userID, suffix)
	if err != nil {
		return fmt.Errorf("anonymize user: %w", err)
	}
	// 0 строк — пользователь либо не найден, либо уже удалён. Идемпотентно.
	_ = tag.RowsAffected()
	return nil
}

// ChangePassword проверяет текущий пароль и устанавливает новый. План 36 Critical-6.
// Минимальная длина нового пароля — 8 символов (как при регистрации).
func (s *Service) ChangePassword(ctx context.Context, userID, current, newPwd string) error {
	if len(newPwd) < 8 {
		return fmt.Errorf("identitysvc: password >= 8 chars")
	}
	var currentHash string
	err := s.db.Pool().QueryRow(ctx,
		`SELECT password_hash FROM users WHERE id = $1 AND deleted_at IS NULL`, userID,
	).Scan(&currentHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidCredential
		}
		return fmt.Errorf("get current hash: %w", err)
	}
	ok, err := auth.VerifyPassword(current, currentHash)
	if err != nil || !ok {
		return ErrInvalidCredential
	}
	newHash, err := auth.HashPassword(newPwd)
	if err != nil {
		return err
	}
	_, err = s.db.Pool().Exec(ctx,
		`UPDATE users SET password_hash = $1 WHERE id = $2`, newHash, userID,
	)
	if err != nil {
		return fmt.Errorf("update hash: %w", err)
	}
	return nil
}

// issueTokens выпускает токены для пользователя.
//
// План 52 Ф.2: если RBAC service подключён — заполняем claims roles/
// permissions из user_roles + role_permissions (динамическая модель).
// Иначе fallback на u.Roles (устаревшее поле users.roles[]).
func (s *Service) issueTokens(ctx context.Context, u User) (jwtrs.Tokens, error) {
	universes, _ := s.GetActiveUniverses(ctx, u.ID)

	roles := u.Roles
	var perms []string
	if s.rbac != nil {
		uid, err := uuid.Parse(u.ID)
		if err == nil {
			if rbacRoles, err := s.rbac.GetUserRoleNames(ctx, uid); err == nil && len(rbacRoles) > 0 {
				roles = rbacRoles
			}
			if rbacPerms, err := s.rbac.GetUserPermissions(ctx, uid); err == nil {
				perms = rbacPerms
			}
		}
	}

	return s.iss.Issue(jwtrs.IssueInput{
		UserID:          u.ID,
		Username:        u.Username,
		ActiveUniverses: universes,
		Roles:           roles,
		Permissions:     perms,
	})
}
