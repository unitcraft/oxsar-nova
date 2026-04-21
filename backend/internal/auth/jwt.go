package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims — кастомные поля JWT.
type Claims struct {
	UserID string `json:"uid"`
	jwt.RegisteredClaims
}

// Tokens — пара access + refresh, отдаваемая клиенту при логине.
type Tokens struct {
	Access  string    `json:"access"`
	Refresh string    `json:"refresh"`
	Expires time.Time `json:"expires"`
}

// JWTIssuer выпускает access/refresh-токены.
type JWTIssuer struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewJWTIssuer(secret string, accessTTL, refreshTTL time.Duration) *JWTIssuer {
	return &JWTIssuer{secret: []byte(secret), accessTTL: accessTTL, refreshTTL: refreshTTL}
}

// Issue создаёт пару токенов для userID.
func (j *JWTIssuer) Issue(userID string) (Tokens, error) {
	now := time.Now().UTC()
	access, err := j.signClaims(userID, now, j.accessTTL, "access")
	if err != nil {
		return Tokens{}, err
	}
	refresh, err := j.signClaims(userID, now, j.refreshTTL, "refresh")
	if err != nil {
		return Tokens{}, err
	}
	return Tokens{Access: access, Refresh: refresh, Expires: now.Add(j.accessTTL)}, nil
}

// Parse валидирует токен и возвращает userID. kind — "access" или "refresh".
func (j *JWTIssuer) Parse(raw, kind string) (string, error) {
	tok, err := jwt.ParseWithClaims(raw, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return j.secret, nil
	})
	if err != nil {
		return "", fmt.Errorf("parse jwt: %w", err)
	}
	claims, ok := tok.Claims.(*Claims)
	if !ok || !tok.Valid {
		return "", errors.New("auth: invalid token")
	}
	if claims.Subject != kind {
		return "", fmt.Errorf("auth: wrong token kind %q, want %q", claims.Subject, kind)
	}
	return claims.UserID, nil
}

func (j *JWTIssuer) signClaims(userID string, now time.Time, ttl time.Duration, kind string) (string, error) {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   kind,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			NotBefore: jwt.NewNumericDate(now),
		},
	})
	return tok.SignedString(j.secret)
}
