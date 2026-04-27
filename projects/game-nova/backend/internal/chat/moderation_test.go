package chat

import (
	"testing"

	"oxsar/game-nova/internal/moderation"
)

// План 46 Ф.4: проверка rate-limit и blacklist в Handler без БД.

func TestAllowSend_RateLimit(t *testing.T) {
	h := NewHandler(nil, nil)
	uid := "user-1"
	for i := 0; i < rateLimitCount; i++ {
		if !h.allowSend(uid) {
			t.Fatalf("expected allow at i=%d", i)
		}
	}
	if h.allowSend(uid) {
		t.Fatal("expected rate limit on (rateLimitCount+1)-st send")
	}
	// Другому пользователю не мешаем.
	if !h.allowSend("user-2") {
		t.Error("expected allow for different user")
	}
}

func TestContainsForbidden(t *testing.T) {
	h := NewHandler(nil, nil)
	if h.containsForbidden("anything") {
		t.Error("nil blacklist must allow")
	}
	bl := moderation.NewBlacklist([]string{"героин"})
	h = h.WithBlacklist(bl)
	if !h.containsForbidden("Покупайте героин!") {
		t.Error("forbidden phrase must be blocked")
	}
	if h.containsForbidden("Привет всем") {
		t.Error("clean phrase must pass")
	}
}
