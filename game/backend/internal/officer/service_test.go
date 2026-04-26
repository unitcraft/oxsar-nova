package officer

import "testing"

func TestAllowedField(t *testing.T) {
	t.Parallel()
	valid := []string{
		"exchange_rate", "research_factor", "build_factor",
		"produce_factor", "energy_factor", "storage_factor",
	}
	for _, f := range valid {
		if !allowedField(f) {
			t.Errorf("expected %q to be allowed", f)
		}
	}
	invalid := []string{"", "admin", "password", "user_id", "credit", "score"}
	for _, f := range invalid {
		if allowedField(f) {
			t.Errorf("expected %q to be rejected", f)
		}
	}
}
