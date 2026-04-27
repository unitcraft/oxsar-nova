package session

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestStore(t *testing.T, idle time.Duration) (*Store, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return NewStore(rdb, idle), mr
}

func TestStore_CreateAndGet(t *testing.T) {
	store, _ := newTestStore(t, time.Hour)
	ctx := context.Background()

	in := Session{
		AccessToken:    "access",
		RefreshToken:   "refresh",
		AccessTokenExp: time.Now().Add(15 * time.Minute),
		Claims: Claims{
			Subject:     "user-1",
			Username:    "alice",
			Roles:       []string{"admin"},
			Permissions: []string{"users:read"},
		},
		IP:        "10.0.0.1",
		UserAgent: "test",
	}
	id, csrf, err := store.Create(ctx, in)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if len(id) != 32 || len(csrf) != 32 {
		t.Fatalf("expected 32-hex tokens, got id=%d csrf=%d", len(id), len(csrf))
	}

	got, err := store.Get(ctx, id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Claims.Username != "alice" {
		t.Fatalf("username mismatch: %q", got.Claims.Username)
	}
	if got.AccessToken != "access" {
		t.Fatalf("access token mismatch")
	}
	if got.CSRFToken != csrf {
		t.Fatalf("csrf token mismatch")
	}
	if got.CreatedAt.IsZero() || got.LastSeenAt.IsZero() {
		t.Fatalf("timestamps not set")
	}
}

func TestStore_GetNotFound(t *testing.T) {
	store, _ := newTestStore(t, time.Hour)
	if _, err := store.Get(context.Background(), "nope"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if _, err := store.Get(context.Background(), ""); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound for empty id, got %v", err)
	}
}

func TestStore_Touch(t *testing.T) {
	store, _ := newTestStore(t, time.Hour)
	ctx := context.Background()

	id, _, err := store.Create(ctx, Session{Claims: Claims{Username: "bob"}})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	first, err := store.Get(ctx, id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	// Запоминаем текущее LastSeenAt и подменяем его в Redis на древнее,
	// чтобы Touch гарантированно сдвинул значение вперёд (без сравнения
	// с моментом времени, который зависит от разрешения часов CI).
	first.LastSeenAt = first.LastSeenAt.Add(-time.Hour)
	if err := store.Update(ctx, first); err != nil {
		t.Fatalf("update: %v", err)
	}
	pre := first.LastSeenAt
	if err := store.Touch(ctx, first); err != nil {
		t.Fatalf("touch: %v", err)
	}
	second, err := store.Get(ctx, id)
	if err != nil {
		t.Fatalf("get after touch: %v", err)
	}
	if !second.LastSeenAt.After(pre) {
		t.Fatalf("LastSeenAt not advanced after Touch: pre=%v post=%v",
			pre, second.LastSeenAt)
	}
}

func TestStore_Delete(t *testing.T) {
	store, _ := newTestStore(t, time.Hour)
	ctx := context.Background()

	id, _, err := store.Create(ctx, Session{Claims: Claims{Username: "carol"}})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := store.Delete(ctx, id); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := store.Get(ctx, id); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
	// Idempotent
	if err := store.Delete(ctx, id); err != nil {
		t.Fatalf("delete twice: %v", err)
	}
	if err := store.Delete(ctx, ""); err != nil {
		t.Fatalf("delete empty: %v", err)
	}
}

func TestStore_TTLExpiry(t *testing.T) {
	store, mr := newTestStore(t, 100*time.Millisecond)
	ctx := context.Background()
	id, _, err := store.Create(ctx, Session{Claims: Claims{Username: "dave"}})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	mr.FastForward(200 * time.Millisecond)
	if _, err := store.Get(ctx, id); err != ErrNotFound {
		t.Fatalf("expected expired session to be ErrNotFound, got %v", err)
	}
}
