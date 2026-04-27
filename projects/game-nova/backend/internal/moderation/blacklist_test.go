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
		{"г е р о и н", true},
		{"FuCk", true},
		{"admin", true},
		{"administrator", true},
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

func TestMaskForbidden(t *testing.T) {
	bl := NewBlacklist([]string{"героин"})
	in := "Я продаю героин дёшево!"
	out := bl.MaskForbidden(in)
	if out == in {
		t.Errorf("expected mask applied, got %q", out)
	}
	// Не-запрещённое сообщение не меняется.
	clean := "Привет всем 42!"
	if got := bl.MaskForbidden(clean); got != clean {
		t.Errorf("clean message changed: %q -> %q", clean, got)
	}
}

func TestEmptyBlacklist(t *testing.T) {
	var bl *Blacklist
	if got, _ := bl.IsForbidden("anything"); got {
		t.Error("nil blacklist must allow everything")
	}
}
