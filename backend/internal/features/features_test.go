package features

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

var osWriteFile = os.WriteFile

func TestParse_Empty(t *testing.T) {
	s, err := ParseBytes([]byte(""))
	if err != nil {
		t.Fatal(err)
	}
	if Enabled(s, "anything") {
		t.Error("empty set should have all flags false")
	}
}

func TestParse_Sample(t *testing.T) {
	yaml := []byte(`
features:
  goal_engine:
    enabled: true
    description: "Новый движок целей"
  experimental_battle:
    enabled: false
    description: "Новая боевая формула"
`)
	s, err := ParseBytes(yaml)
	if err != nil {
		t.Fatal(err)
	}
	if !Enabled(s, "goal_engine") {
		t.Error("goal_engine should be enabled")
	}
	if Enabled(s, "experimental_battle") {
		t.Error("experimental_battle should be disabled")
	}
	if Enabled(s, "unknown") {
		t.Error("unknown flag should be false (fail-closed)")
	}
}

func TestEnabled_NilSet(t *testing.T) {
	// Защита от crash, если Load не вызывался.
	if Enabled(nil, "goal_engine") {
		t.Error("nil set should return false")
	}
}

func TestEnabledKeys_Sorted(t *testing.T) {
	yaml := []byte(`
features:
  zzz:
    enabled: true
  aaa:
    enabled: true
  middle:
    enabled: false
`)
	s, _ := ParseBytes(yaml)
	keys := EnabledKeys(s)
	want := []string{"aaa", "zzz"}
	if !reflect.DeepEqual(keys, want) {
		t.Errorf("EnabledKeys: got %v, want %v", keys, want)
	}
}

func TestAll_Copy(t *testing.T) {
	yaml := []byte(`
features:
  flag_a:
    enabled: true
    description: "A"
`)
	s, _ := ParseBytes(yaml)
	all := All(s)
	if len(all) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(all))
	}
	// Мутация копии не должна влиять на Set.
	all["flag_a"] = Flag{Enabled: false}
	if !Enabled(s, "flag_a") {
		t.Error("All() must return a copy, not original map")
	}
}

func TestLoad_FileNotExist(t *testing.T) {
	// Несуществующий файл — пустой Set, не ошибка.
	s, err := Load("/nonexistent/path/features.yaml")
	if err != nil {
		t.Fatal("missing file should not be error")
	}
	if Enabled(s, "anything") {
		t.Error("should be empty")
	}
}

func TestLoad_FromTempFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "features.yaml")
	yaml := []byte(`
features:
  goal_engine:
    enabled: true
`)
	if err := writeFile(path, yaml); err != nil {
		t.Fatal(err)
	}
	s, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !Enabled(s, "goal_engine") {
		t.Error("loaded flag not enabled")
	}
}

// helper для test (избегаем os.WriteFile в тестах из-за импортов).
func writeFile(path string, data []byte) error {
	return osWriteFile(path, data, 0o644)
}
