// Package jwtrs выпускает и верифицирует JWT с RSA-256.
//
// Auth Service подписывает токены приватным ключом.
// Все остальные сервисы верифицируют публичным ключом через JWKS.
package jwtrs

// DUPLICATE: этот файл скопирован между Go-модулями oxsar/game-nova,
// oxsar/auth и oxsar/portal. При любом изменении синхронизируйте КОПИИ:
//   - projects/game-nova/backend/pkg/jwtrs/jwtrs.go
//   - projects/auth/backend/pkg/jwtrs/jwtrs.go
//   - projects/portal/backend/pkg/jwtrs/jwtrs.go
// Причина дубля: каждый домен — отдельный go.mod, без shared-модуля.

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims — кастомные поля JWT для oxsar-nova.
//
// Email включён в claims, чтобы handoff в game-nova/portal мог
// зеркалить аккаунт в локальной БД без дополнительного запроса
// /auth/me в auth-service. План 36 Ф.12.
type Claims struct {
	Username        string   `json:"username"`
	Email           string   `json:"email"`
	GlobalCredits   int64    `json:"global_credits"`
	ActiveUniverses []string `json:"active_universes"`
	Roles           []string `json:"roles"`
	jwt.RegisteredClaims
}

// Tokens — пара access + refresh, отдаваемая клиенту.
type Tokens struct {
	Access  string    `json:"access"`
	Refresh string    `json:"refresh"`
	Expires time.Time `json:"expires"`
}

// IssueInput — данные для выпуска токенов.
type IssueInput struct {
	UserID          string
	Username        string
	Email           string
	GlobalCredits   int64
	ActiveUniverses []string
	Roles           []string
}

// Issuer выпускает JWT приватным ключом RSA-256.
type Issuer struct {
	privateKey *rsa.PrivateKey
	keyID      string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// Verifier верифицирует JWT публичным ключом RSA-256.
type Verifier struct {
	publicKey *rsa.PublicKey
	keyID     string
}

// LoadOrGenerateKey загружает приватный RSA-ключ из файла или генерирует новый.
func LoadOrGenerateKey(keyPath string) (*rsa.PrivateKey, error) {
	if data, err := os.ReadFile(keyPath); err == nil {
		return parsePrivateKey(data)
	}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generate rsa key: %w", err)
	}
	if err := os.WriteFile(keyPath, encodePrivateKey(key), 0600); err != nil {
		return nil, fmt.Errorf("write rsa key: %w", err)
	}
	return key, nil
}

// NewIssuer создаёт Issuer из приватного ключа.
func NewIssuer(key *rsa.PrivateKey, accessTTL, refreshTTL time.Duration) *Issuer {
	return &Issuer{
		privateKey: key,
		keyID:      kidFromPublicKey(&key.PublicKey),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// NewVerifierFromKey создаёт Verifier из публичного ключа напрямую.
func NewVerifierFromKey(key *rsa.PublicKey) *Verifier {
	return &Verifier{publicKey: key, keyID: kidFromPublicKey(key)}
}

// NewVerifierFromJWKS создаёт Verifier из JWKS JSON.
func NewVerifierFromJWKS(data []byte) (*Verifier, error) {
	var jwks JWKS
	if err := json.Unmarshal(data, &jwks); err != nil {
		return nil, fmt.Errorf("unmarshal jwks: %w", err)
	}
	if len(jwks.Keys) == 0 {
		return nil, errors.New("jwtrs: empty jwks")
	}
	k := jwks.Keys[0]
	pub, err := jwkToPublicKey(k)
	if err != nil {
		return nil, err
	}
	return &Verifier{publicKey: pub, keyID: k.Kid}, nil
}

// PublicKey возвращает публичный ключ Issuer-а.
func (iss *Issuer) PublicKey() *rsa.PublicKey { return &iss.privateKey.PublicKey }

// KeyID возвращает kid ключа.
func (iss *Issuer) KeyID() string { return iss.keyID }

// Issue создаёт пару токенов.
func (iss *Issuer) Issue(in IssueInput) (Tokens, error) {
	now := time.Now().UTC()
	access, err := iss.sign(in, now, iss.accessTTL, "access")
	if err != nil {
		return Tokens{}, err
	}
	refresh, err := iss.sign(in, now, iss.refreshTTL, "refresh")
	if err != nil {
		return Tokens{}, err
	}
	return Tokens{Access: access, Refresh: refresh, Expires: now.Add(iss.accessTTL)}, nil
}

func (iss *Issuer) sign(in IssueInput, now time.Time, ttl time.Duration, kind string) (string, error) {
	roles := in.Roles
	if len(roles) == 0 {
		roles = []string{"player"}
	}
	universes := in.ActiveUniverses
	if universes == nil {
		universes = []string{}
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, &Claims{
		Username:        in.Username,
		Email:           in.Email,
		GlobalCredits:   in.GlobalCredits,
		ActiveUniverses: universes,
		Roles:           roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   in.UserID,
			ID:        kind,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			NotBefore: jwt.NewNumericDate(now),
		},
	})
	tok.Header["kid"] = iss.keyID
	return tok.SignedString(iss.privateKey)
}

// Parse верифицирует токен и возвращает Claims. kind — "access" или "refresh".
func (v *Verifier) Parse(raw, kind string) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(raw, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return v.publicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse jwt: %w", err)
	}
	claims, ok := tok.Claims.(*Claims)
	if !ok || !tok.Valid {
		return nil, errors.New("jwtrs: invalid token")
	}
	if claims.ID != kind {
		return nil, fmt.Errorf("jwtrs: wrong token kind %q, want %q", claims.ID, kind)
	}
	return claims, nil
}

// JWKS — JSON Web Key Set.
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK — одна запись в JWKS.
type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// IssuerToJWKS формирует JWKS из Issuer.
func IssuerToJWKS(iss *Issuer) JWKS {
	return publicKeyToJWKS(iss.PublicKey(), iss.keyID)
}

func publicKeyToJWKS(key *rsa.PublicKey, kid string) JWKS {
	eBytes := big.NewInt(int64(key.E)).Bytes()
	return JWKS{Keys: []JWK{{
		Kty: "RSA",
		Use: "sig",
		Kid: kid,
		Alg: "RS256",
		N:   base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
		E:   base64.RawURLEncoding.EncodeToString(eBytes),
	}}}
}

func jwkToPublicKey(k JWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, fmt.Errorf("jwtrs: decode n: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, fmt.Errorf("jwtrs: decode e: %w", err)
	}
	e := new(big.Int).SetBytes(eBytes)
	if !e.IsInt64() {
		return nil, errors.New("jwtrs: exponent too large")
	}
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: int(e.Int64()),
	}, nil
}

func kidFromPublicKey(key *rsa.PublicKey) string {
	n := base64.RawURLEncoding.EncodeToString(key.N.Bytes())
	if len(n) > 16 {
		return n[:16]
	}
	return n
}

func encodePrivateKey(key *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
}

func parsePrivateKey(data []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("jwtrs: no PEM block found")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// HandoffToken — одноразовый токен для переключения вселенных.
type HandoffToken struct {
	Token     string
	UserID    string
	ExpiresAt time.Time
}

// NewHandoffToken генерирует UUID-токен.
func NewHandoffToken(userID string) HandoffToken {
	id, _ := uuid.NewV7()
	return HandoffToken{
		Token:     id.String(),
		UserID:    userID,
		ExpiresAt: time.Now().UTC().Add(30 * time.Second),
	}
}
