package artefact

import (
	"testing"

	"github.com/oxsar/nova/backend/internal/config"
)

func TestComputeChanges_SetSymmetric(t *testing.T) {
	t.Parallel()
	spec := config.ArtefactSpec{
		ID: 300,
		Effect: config.ArtefactEffect{
			Type:          "factor_user",
			Field:         "exchange_rate",
			Op:            "set",
			ActiveValue:   1.03,
			InactiveValue: 1.2,
		},
	}
	apply, err := computeChanges(spec, dirApply)
	if err != nil {
		t.Fatalf("apply err: %v", err)
	}
	if apply.Op != "set" || apply.NewValue != 1.03 {
		t.Fatalf("apply wrong: %+v", apply)
	}

	revert, err := computeChanges(spec, dirRevert)
	if err != nil {
		t.Fatalf("revert err: %v", err)
	}
	if revert.NewValue != 1.2 {
		t.Fatalf("revert wrong: %+v", revert)
	}
}

func TestComputeChanges_AddOppositeSigns(t *testing.T) {
	t.Parallel()
	spec := config.ArtefactSpec{
		ID: 301,
		Effect: config.ArtefactEffect{
			Type:  "factor_all_planets",
			Field: "produce_factor",
			Op:    "add",
			Value: 0.1,
		},
	}
	apply, _ := computeChanges(spec, dirApply)
	revert, _ := computeChanges(spec, dirRevert)
	if apply.Delta+revert.Delta != 0 {
		t.Fatalf("apply+revert must sum to zero: %v + %v", apply.Delta, revert.Delta)
	}
}

func TestComputeChanges_RejectsUnknownField(t *testing.T) {
	t.Parallel()
	spec := config.ArtefactSpec{
		Effect: config.ArtefactEffect{
			Type:  "factor_user",
			Field: "drop_all_tables",
			Op:    "add",
			Value: 1,
		},
	}
	if _, err := computeChanges(spec, dirApply); err == nil {
		t.Fatalf("expected rejection of arbitrary field")
	}
}

func TestComputeChanges_UnsupportedType(t *testing.T) {
	t.Parallel()
	spec := config.ArtefactSpec{
		Effect: config.ArtefactEffect{Type: "one_shot"},
	}
	if _, err := computeChanges(spec, dirApply); err != ErrUnsupported {
		t.Fatalf("expected ErrUnsupported, got %v", err)
	}
}

// battle_bonus не меняет планетарные факторы — computeChanges возвращает (nil, nil).
// Боевые модификаторы читаются отдельно через ComputeBattleModifier.
func TestComputeChanges_BattleBonus(t *testing.T) {
	t.Parallel()
	spec := config.ArtefactSpec{
		Effect: config.ArtefactEffect{Type: "battle_bonus"},
	}
	fc, err := computeChanges(spec, dirApply)
	if err != nil {
		t.Fatalf("unexpected error for battle_bonus: %v", err)
	}
	if fc != nil {
		t.Fatalf("expected nil FactorChange for battle_bonus, got %+v", fc)
	}
}

func TestComputeChanges_UnknownType(t *testing.T) {
	t.Parallel()
	spec := config.ArtefactSpec{
		Effect: config.ArtefactEffect{Type: "totally_unknown"},
	}
	if _, err := computeChanges(spec, dirApply); err == nil {
		t.Fatalf("expected error for unknown effect type")
	}
}

func TestFactorChange_UnknownOp(t *testing.T) {
	t.Parallel()
	spec := config.ArtefactSpec{
		Effect: config.ArtefactEffect{
			Type:  "factor_user",
			Field: "exchange_rate",
			Op:    "multiply", // unsupported op
		},
	}
	if _, err := computeChanges(spec, dirApply); err == nil {
		t.Fatalf("expected error for unsupported op")
	}
}

func TestComputeChanges_PlanetScope(t *testing.T) {
	t.Parallel()
	spec := config.ArtefactSpec{
		Effect: config.ArtefactEffect{
			Type:  "factor_planet",
			Field: "build_factor",
			Op:    "add",
			Value: 0.25,
		},
	}
	ch, err := computeChanges(spec, dirApply)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if ch.Scope != "planet" || ch.Field != "build_factor" {
		t.Errorf("unexpected change: %+v", ch)
	}
}

// TestMaxStacks_SpecField: MaxStacks поле корректно читается из ArtefactSpec.
func TestMaxStacks_SpecField(t *testing.T) {
	t.Parallel()
	spec := config.ArtefactSpec{
		ID:        301,
		Stackable: true,
		MaxStacks: 3,
		Effect: config.ArtefactEffect{
			Type:  "factor_all_planets",
			Field: "produce_factor",
			Op:    "add",
			Value: 0.1,
		},
	}
	if spec.MaxStacks != 3 {
		t.Fatalf("MaxStacks should be 3, got %d", spec.MaxStacks)
	}
	// max_stacks=0 означает лимит не применяется
	noLimit := config.ArtefactSpec{ID: 303, Stackable: true, MaxStacks: 0}
	if noLimit.MaxStacks != 0 {
		t.Fatalf("MaxStacks=0 should mean no limit")
	}
}

// TestErrMaxStacksReached_IsSentinel: ошибка объявлена и не nil.
func TestErrMaxStacksReached_IsSentinel(t *testing.T) {
	t.Parallel()
	if ErrMaxStacksReached == nil {
		t.Fatal("ErrMaxStacksReached must not be nil")
	}
}
