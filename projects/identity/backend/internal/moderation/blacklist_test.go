package moderation

import "testing"

func TestIsForbidden(t *testing.T) {
	bl := NewBlacklist([]string{"героин", "fuck", "admin"})

	cases := []struct {
		in   string
		want bool
	}{
		{"героин", true},
		{"Героин-2024", true},
		{"г е р о и н", true}, // обход через пробелы
		{"FuCk", true},
		{"admin", true},
		{"administrator", true}, // префикс совпадает
		{"героин!", true},
		{"normal_user", false},
		{"Игрок 42", false},
		{"", false},
	}
	for _, c := range cases {
		got, _ := bl.IsForbidden(c.in)
		if got != c.want {
			t.Errorf("IsForbidden(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestEmptyBlacklist(t *testing.T) {
	var bl *Blacklist
	if got, _ := bl.IsForbidden("anything"); got {
		t.Error("nil blacklist must allow everything")
	}
	bl2 := NewBlacklist(nil)
	if got, _ := bl2.IsForbidden("anything"); got {
		t.Error("empty blacklist must allow everything")
	}
}
