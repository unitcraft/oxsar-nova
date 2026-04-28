package alliance

import (
	"testing"
)

func TestPermissionInJSON(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		perm Permission
		want bool
	}{
		{"empty", ``, PermInvite, false},
		{"true", `{"can_invite":true}`, PermInvite, true},
		{"false", `{"can_invite":false}`, PermInvite, false},
		{"absent_key", `{"can_kick":true}`, PermInvite, false},
		{"non_bool", `{"can_invite":"yes"}`, PermInvite, false},
		{"invalid_json", `{not-json`, PermInvite, false},
		{"empty_obj", `{}`, PermInvite, false},
		{"manage_ranks", `{"can_manage_ranks":true}`, PermManageRanks, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := permissionInJSON([]byte(c.raw), c.perm)
			if got != c.want {
				t.Errorf("permissionInJSON(%q,%s) = %v, want %v", c.raw, c.perm, got, c.want)
			}
		})
	}
}

func TestIsValidPermission(t *testing.T) {
	for _, p := range AllPermissions {
		if !IsValidPermission(string(p)) {
			t.Errorf("IsValidPermission(%q) = false, want true", p)
		}
	}
	for _, bad := range []string{"", "invalid", "can_admin", "Can_Invite"} {
		if IsValidPermission(bad) {
			t.Errorf("IsValidPermission(%q) = true, want false", bad)
		}
	}
}

func TestNormalizeRelation(t *testing.T) {
	cases := []struct {
		in       string
		canon    string
		ok       bool
	}{
		{"friend", "friend", true},
		{"neutral", "neutral", true},
		{"hostile_neutral", "hostile_neutral", true},
		{"nap", "nap", true},
		{"war", "war", true},
		{"ally", "friend", true}, // legacy alias
		{"none", "none", true},
		{"", "", false},
		{"unknown", "", false},
		{"WAR", "", false}, // case-sensitive
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got, ok := normalizeRelation(c.in)
			if ok != c.ok || got != c.canon {
				t.Errorf("normalizeRelation(%q) = (%q,%v), want (%q,%v)", c.in, got, ok, c.canon, c.ok)
			}
		})
	}
}

func TestRelationNeedsAccept(t *testing.T) {
	cases := map[string]bool{
		"friend":          true,
		"neutral":         true,
		"nap":             true,
		"war":             false,
		"hostile_neutral": false,
	}
	for rel, want := range cases {
		if got := relationNeedsAccept(rel); got != want {
			t.Errorf("relationNeedsAccept(%q) = %v, want %v", rel, got, want)
		}
	}
}

func TestDecodePermissions(t *testing.T) {
	got := decodePermissions([]byte(`{"can_invite":true,"can_kick":false,"unknown":1}`))
	if got["can_invite"] != true {
		t.Errorf("can_invite = %v, want true", got["can_invite"])
	}
	if got["can_kick"] != false {
		t.Errorf("can_kick = %v, want false", got["can_kick"])
	}
	// Не-bool значения отбрасываются.
	if _, ok := got["unknown"]; ok {
		t.Error("non-bool key should be dropped")
	}

	// Пустой raw → пустая мапа.
	if len(decodePermissions(nil)) != 0 {
		t.Error("nil raw should yield empty map")
	}
	if len(decodePermissions([]byte("not-json"))) != 0 {
		t.Error("invalid json should yield empty map")
	}
}

func TestValidatePermissions(t *testing.T) {
	good := map[string]bool{"can_invite": true, "can_kick": false}
	if err := validatePermissions(good); err != nil {
		t.Errorf("validatePermissions(good) = %v", err)
	}
	bad := map[string]bool{"can_admin": true}
	if err := validatePermissions(bad); err == nil {
		t.Error("validatePermissions(bad) = nil, want error")
	}
}

func TestValidateRankName(t *testing.T) {
	cases := []struct {
		name string
		ok   bool
	}{
		{"", false},
		{"a", true},
		{"Officer", true},
		{stringRepeat("x", 32), true},
		{stringRepeat("x", 33), false},
	}
	for _, c := range cases {
		err := validateRankName(c.name)
		if (err == nil) != c.ok {
			t.Errorf("validateRankName(len=%d) ok=%v, err=%v", len([]rune(c.name)), c.ok, err)
		}
	}
}

func stringRepeat(s string, n int) string {
	out := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		out = append(out, s...)
	}
	return string(out)
}
