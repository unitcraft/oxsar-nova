package alliance

import "testing"

func TestValidateTag_Valid(t *testing.T) {
	t.Parallel()
	valid := []string{"ABC", "ab1", "XY12Z", "12345", "abcde"}
	for _, tag := range valid {
		if err := validateTag(tag); err != nil {
			t.Errorf("validateTag(%q) = %v, want nil", tag, err)
		}
	}
}

func TestValidateTag_TooShort(t *testing.T) {
	t.Parallel()
	for _, tag := range []string{"", "a", "AB"} {
		if err := validateTag(tag); err == nil {
			t.Errorf("validateTag(%q) = nil, want error (too short)", tag)
		}
	}
}

func TestValidateTag_TooLong(t *testing.T) {
	t.Parallel()
	if err := validateTag("ABCDEF"); err == nil {
		t.Error("validateTag(6 chars) = nil, want error (too long)")
	}
}

func TestValidateTag_InvalidChars(t *testing.T) {
	t.Parallel()
	invalid := []string{"AB-C", "AB C", "АБВ", "ab_c", "ab.c"}
	for _, tag := range invalid {
		if err := validateTag(tag); err == nil {
			t.Errorf("validateTag(%q) = nil, want error (invalid chars)", tag)
		}
	}
}

func TestValidateTag_Unicode(t *testing.T) {
	t.Parallel()
	// Кириллица — 3 руны, но не ASCII alphanumeric.
	if err := validateTag("АБВ"); err == nil {
		t.Error("Cyrillic tag must be rejected")
	}
}

func TestIsAlphanumASCII(t *testing.T) {
	t.Parallel()
	yes := []rune{'A', 'Z', 'a', 'z', '0', '9', 'M', 'm', '5'}
	no := []rune{'-', '_', ' ', 'А', '.', '!', '\x00'}
	for _, r := range yes {
		if !isAlphanumASCII(r) {
			t.Errorf("isAlphanumASCII(%q) = false, want true", r)
		}
	}
	for _, r := range no {
		if isAlphanumASCII(r) {
			t.Errorf("isAlphanumASCII(%q) = true, want false", r)
		}
	}
}

func TestIsDupKey(t *testing.T) {
	t.Parallel()
	if isDupKey(nil) {
		t.Fatal("isDupKey(nil) must be false")
	}
}

func TestErrorSentinels_NotNil(t *testing.T) {
	t.Parallel()
	sentinels := []error{
		ErrNotFound, ErrAlreadyMember, ErrNotMember,
		ErrNotOwner, ErrTagTaken, ErrNameTaken,
		ErrInvalidTag, ErrCannotLeaveOwn,
	}
	for _, err := range sentinels {
		if err == nil {
			t.Errorf("sentinel error is nil: %T", err)
		}
	}
}
