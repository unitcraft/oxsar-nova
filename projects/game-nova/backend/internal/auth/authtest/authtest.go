// Package authtest предоставляет helper'ы для тестов, которым нужно
// поставить userID в контекст так же, как это делает RSAMiddleware,
// но без выпуска JWT и поднятия middleware-стека.
//
// Использовать ТОЛЬКО из _test.go-файлов.
package authtest

import (
	"context"

	"oxsar/game-nova/internal/auth"
)

// WithUserID кладёт userID в context под публичным ключом auth-пакета
// (см. auth.UserIDKey). Возвращает производный контекст.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, auth.UserIDKey, userID)
}
