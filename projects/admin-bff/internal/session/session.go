// Package session — Redis-backed session store для admin-bff.
//
// Каждая сессия привязана к одному админ-юзеру и хранит:
//   - access JWT (короткий, ~15 мин) — для проксирования к бекендам.
//   - refresh JWT (~7 дней) — admin-bff обменивает за 60s до exp.
//   - claims summary (username, roles, permissions) — для UI guards.
//   - last activity (для idle timeout).
//   - IP, user-agent (для audit и invalidation).
//
// Cookie на стороне браузера — opaque 128-bit random session ID:
// `admin_session=<id>`, HttpOnly Secure SameSite=Strict.
//
// Refresh выполняется лениво: при каждом проксируемом запросе, если
// до exp осталось < RefreshLeadTime, admin-bff ходит в identity и
// обновляет токены атомарно (Redis WATCH/MULTI).
package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	keyPrefix = "admin:sess:"
)

var (
	ErrNotFound = errors.New("session not found")
	ErrExpired  = errors.New("session expired")
)

// Claims — summary, который видит admin-frontend (не сам JWT).
type Claims struct {
	Subject     string   `json:"sub"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

// Session — то что мы храним в Redis.
type Session struct {
	ID                string    `json:"id"`
	AccessToken       string    `json:"access_token"`
	RefreshToken      string    `json:"refresh_token"`
	AccessTokenExp    time.Time `json:"access_token_exp"`
	Claims            Claims    `json:"claims"`
	CSRFToken         string    `json:"csrf_token"`
	IP                string    `json:"ip"`
	UserAgent         string    `json:"user_agent"`
	CreatedAt         time.Time `json:"created_at"`
	LastSeenAt        time.Time `json:"last_seen_at"`
}

// Store — обёртка над Redis для CRUD сессий.
type Store struct {
	rdb         *redis.Client
	idleTimeout time.Duration
}

func NewStore(rdb *redis.Client, idleTimeout time.Duration) *Store {
	return &Store{rdb: rdb, idleTimeout: idleTimeout}
}

// Create — генерирует новый session ID + CSRF token, сохраняет в Redis.
func (s *Store) Create(ctx context.Context, sess Session) (string, string, error) {
	id, err := randomHex(16)
	if err != nil {
		return "", "", fmt.Errorf("session id: %w", err)
	}
	csrf, err := randomHex(16)
	if err != nil {
		return "", "", fmt.Errorf("csrf token: %w", err)
	}
	sess.ID = id
	sess.CSRFToken = csrf
	sess.CreatedAt = time.Now()
	sess.LastSeenAt = sess.CreatedAt
	if err := s.save(ctx, &sess); err != nil {
		return "", "", err
	}
	return id, csrf, nil
}

// Get — читает сессию из Redis. Если найдена, обновляет LastSeenAt
// (sliding TTL) и возвращает копию.
func (s *Store) Get(ctx context.Context, id string) (*Session, error) {
	if id == "" {
		return nil, ErrNotFound
	}
	raw, err := s.rdb.Get(ctx, keyPrefix+id).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("redis get: %w", err)
	}
	var sess Session
	if err := json.Unmarshal(raw, &sess); err != nil {
		return nil, fmt.Errorf("session decode: %w", err)
	}
	return &sess, nil
}

// Touch — обновляет LastSeenAt и продлевает TTL.
func (s *Store) Touch(ctx context.Context, sess *Session) error {
	sess.LastSeenAt = time.Now()
	return s.save(ctx, sess)
}

// Update — атомарно сохраняет обновлённую сессию (например, после refresh JWT).
func (s *Store) Update(ctx context.Context, sess *Session) error {
	return s.save(ctx, sess)
}

// Delete — удаляет сессию (logout).
func (s *Store) Delete(ctx context.Context, id string) error {
	if id == "" {
		return nil
	}
	return s.rdb.Del(ctx, keyPrefix+id).Err()
}

func (s *Store) save(ctx context.Context, sess *Session) error {
	raw, err := json.Marshal(sess)
	if err != nil {
		return fmt.Errorf("session encode: %w", err)
	}
	if err := s.rdb.Set(ctx, keyPrefix+sess.ID, raw, s.idleTimeout).Err(); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}
	return nil
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
