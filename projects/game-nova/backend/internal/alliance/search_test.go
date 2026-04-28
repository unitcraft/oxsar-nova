package alliance

// План 67 Ф.4: unit-тесты для полнотекстового поиска.

import (
	"testing"
)

func TestSanitizeTSQuery_AllowsAlphanumAndCyrillic(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"hello":        "hello",
		"Hello123":     "Hello123",
		"альянс":       "альянс",
		"АЛЬЯНС":       "АЛЬЯНС",
		"ёлка":         "ёлка",
		"Тест42":       "Тест42",
	}
	for in, want := range cases {
		if got := sanitizeTSQuery(in); got != want {
			t.Errorf("sanitizeTSQuery(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSanitizeTSQuery_StripsTSQueryOperators(t *testing.T) {
	t.Parallel()
	// to_tsquery интерпретирует & | ! ( ) : как операторы.
	// Без санитайзинга «hello & world» падает с syntax error.
	cases := map[string]string{
		"a&b":      "ab",
		"a|b":      "ab",
		"!hello":   "hello",
		"(test)":   "test",
		"a:b":      "ab",
		"a'b\"c":   "abc",
		"foo;bar":  "foobar",
		"привет!?": "привет",
	}
	for in, want := range cases {
		if got := sanitizeTSQuery(in); got != want {
			t.Errorf("sanitizeTSQuery(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSanitizeTSQuery_EmptyOnAllSpecial(t *testing.T) {
	t.Parallel()
	for _, s := range []string{"", "&|!():", "  ", "\t\n"} {
		if got := sanitizeTSQuery(s); got != "" {
			t.Errorf("sanitizeTSQuery(%q) = %q, want empty", s, got)
		}
	}
}

func TestParseIntDefault(t *testing.T) {
	t.Parallel()
	cases := []struct {
		s    string
		def  int
		want int
	}{
		{"", 50, 50},
		{"10", 50, 10},
		{"abc", 50, 50},
		{"-1", 50, -1},
		{"0", 50, 0},
	}
	for _, c := range cases {
		if got := parseIntDefault(c.s, c.def); got != c.want {
			t.Errorf("parseIntDefault(%q, %d) = %d, want %d", c.s, c.def, got, c.want)
		}
	}
}
