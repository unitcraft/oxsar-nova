package ids

// DUPLICATE: этот файл скопирован между Go-модулями oxsar/game-nova,
// oxsar/identity, oxsar/portal и oxsar/billing. При любом изменении
// синхронизируйте КОПИИ:
//   - projects/game-nova/backend/pkg/ids/ids_test.go
//   - projects/identity/backend/pkg/ids/ids_test.go
//   - projects/portal/backend/pkg/ids/ids_test.go
//   - projects/billing/backend/pkg/ids/ids_test.go
// Причина дубля: каждый домен — отдельный go.mod, без shared-модуля.

import (
	"strings"
	"testing"
)

func TestNew_IsValidUUID(t *testing.T) {
	t.Parallel()
	id := New()
	// UUIDv7 format: 8-4-4-4-12 hex chars separated by hyphens.
	parts := strings.Split(id, "-")
	if len(parts) != 5 {
		t.Fatalf("expected 5 dash-separated parts, got %d: %q", len(parts), id)
	}
	lengths := []int{8, 4, 4, 4, 12}
	for i, p := range parts {
		if len(p) != lengths[i] {
			t.Errorf("part[%d] length = %d, want %d: %q", i, len(p), lengths[i], p)
		}
	}
}

func TestNew_Version7(t *testing.T) {
	t.Parallel()
	id := New()
	// The 13th character (first char of 3rd group) encodes the version, must be '7'.
	parts := strings.Split(id, "-")
	if parts[2][0] != '7' {
		t.Errorf("expected UUIDv7 (version char '7'), got %q in %q", string(parts[2][0]), id)
	}
}

func TestNew_Unique(t *testing.T) {
	t.Parallel()
	seen := map[string]struct{}{}
	for i := 0; i < 1000; i++ {
		id := New()
		if _, dup := seen[id]; dup {
			t.Fatalf("duplicate UUID generated: %q", id)
		}
		seen[id] = struct{}{}
	}
}

func TestNew_Monotonic(t *testing.T) {
	t.Parallel()
	// UUIDv7 IDs generated in sequence should be lexicographically ordered.
	prev := New()
	for i := 0; i < 100; i++ {
		next := New()
		if next < prev {
			t.Fatalf("UUIDv7 not monotonic: %q < %q", next, prev)
		}
		prev = next
	}
}
