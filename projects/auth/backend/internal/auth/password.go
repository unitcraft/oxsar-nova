// Package auth реализует регистрацию, логин и выпуск JWT.
//
// Пароли хешируются argon2id с персональной солью. Глобальный
// PASSWORD_SALT, который был в oxsar2, не используется (см. §18.15 ТЗ).
package auth

// DUPLICATE: этот файл скопирован между Go-модулями oxsar/game-nova,
// oxsar/auth и oxsar/portal. При любом изменении синхронизируйте КОПИИ:
//   - projects/game-nova/backend/internal/auth/password.go
//   - projects/auth/backend/internal/auth/password.go
//   - projects/portal/backend/internal/auth/password.go
// Причина дубля: каждый домен — отдельный go.mod, без shared-модуля.

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Параметры argon2id. Соответствуют рекомендованным 2024–2026 года
// (OWASP ASVS level 2). Настройки задаются как константы, чтобы при
// смене параметров старые хеши продолжали читаться (они кодируют свои).
const (
	argonTime    = 3
	argonMemory  = 64 * 1024
	argonThreads = 2
	argonKeyLen  = 32
	argonSaltLen = 16
)

// HashPassword возвращает закодированную строку вида
// $argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>.
func HashPassword(password string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("gen salt: %w", err)
	}
	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

// VerifyPassword сравнивает plaintext-пароль с закодированным хешом в
// константное время. Возвращает (ok, error); error != nil — если хеш
// не распарсился.
func VerifyPassword(password, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, errors.New("auth: invalid hash format")
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, fmt.Errorf("auth: parse version: %w", err)
	}
	var memory, time uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return false, fmt.Errorf("auth: parse params: %w", err)
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("auth: decode salt: %w", err)
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("auth: decode hash: %w", err)
	}

	got := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}
