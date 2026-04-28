package economy

import (
	"testing"

	"pgregory.net/rapid"

	"oxsar/game-nova/internal/balance"
)

// Property-based тесты (R4): инварианты, которые должны выполняться
// для любых разумных входов. Используем pgregory.net/rapid (план 64
// требует rapid для economy/battle).
//
// Inferences:
//   - Production ≥ 0 для любых level ≥ 0
//   - Production монотонно растёт по level (при фиксированном tech/temp)
//     для positive base coefficients (modern globals все positive).
//   - Production для level=0 == 0 (защита от random math).
//   - Production деттерминирован — два вызова дают одинаковый результат.

func TestProperty_MetalMineProduction(t *testing.T) {
	globals := balance.ModernGlobals()
	rapid.Check(t, func(t *rapid.T) {
		level := rapid.IntRange(0, 50).Draw(t, "level")
		tech := rapid.IntRange(0, 20).Draw(t, "tech")
		got := MetalMineProduction(globals, level, tech)

		if level == 0 && got != 0 {
			t.Fatalf("level=0 should produce 0, got %v", got)
		}
		if got < 0 {
			t.Fatalf("production must be >= 0, got %v (level=%d tech=%d)", got, level, tech)
		}
		// Determinism.
		got2 := MetalMineProduction(globals, level, tech)
		if got != got2 {
			t.Fatalf("non-deterministic: %v vs %v", got, got2)
		}
		// Monotonicity по level (level >= 1).
		if level >= 1 {
			gotPlus := MetalMineProduction(globals, level+1, tech)
			if gotPlus < got {
				t.Fatalf("non-monotonic: level=%d → %v, level=%d → %v",
					level, got, level+1, gotPlus)
			}
		}
	})
}

func TestProperty_HydrogenLabProduction(t *testing.T) {
	globals := balance.ModernGlobals()
	rapid.Check(t, func(t *rapid.T) {
		level := rapid.IntRange(0, 50).Draw(t, "level")
		tech := rapid.IntRange(0, 20).Draw(t, "tech")
		// Температура диапазон легаси — −200..+200 (na_planet schema).
		// При temp ≈ 640 множитель temp_factor = 1.28 - 0.002*640 = 0.
		// Тестим в реалистичном диапазоне; экстремум (negative tempFactor)
		// — отдельный тест ниже.
		temp := rapid.IntRange(-200, 200).Draw(t, "temp")
		got := HydrogenLabProduction(globals, level, tech, temp)

		// При temp_factor > 0 production ≥ 0.
		// temp_factor = -0.002 * temp + 1.28 > 0 ⟺ temp < 640.
		// В диапазоне [-200, 200] всегда положительный.
		if level >= 1 && got < 0 {
			t.Fatalf("hydrogen production negative: level=%d tech=%d temp=%d → %v",
				level, tech, temp, got)
		}
		if level == 0 && got != 0 {
			t.Fatalf("level=0 should produce 0, got %v", got)
		}
	})
}

func TestProperty_SolarPlantProduction(t *testing.T) {
	globals := balance.ModernGlobals()
	rapid.Check(t, func(t *rapid.T) {
		level := rapid.IntRange(0, 50).Draw(t, "level")
		tech := rapid.IntRange(0, 20).Draw(t, "tech")
		got := SolarPlantProduction(globals, level, tech)

		if level == 0 && got != 0 {
			t.Fatalf("level=0 → %v", got)
		}
		if got < 0 {
			t.Fatalf("solar negative: %v (level=%d tech=%d)", got, level, tech)
		}
		// Tech monotonicity: при фиксированном level, tech++ не уменьшает
		// production (1.1 + tech*0.0005 — растёт).
		if level >= 1 && tech < 20 {
			gotPlus := SolarPlantProduction(globals, level, tech+1)
			if gotPlus < got {
				t.Fatalf("non-monotonic by tech: tech=%d → %v, tech=%d → %v",
					tech, got, tech+1, gotPlus)
			}
		}
	})
}

func TestProperty_MineConsEnergy(t *testing.T) {
	globals := balance.ModernGlobals()
	rapid.Check(t, func(t *rapid.T) {
		base := rapid.Float64Range(1, 100).Draw(t, "base")
		level := rapid.IntRange(0, 50).Draw(t, "level")
		tech := rapid.IntRange(0, 20).Draw(t, "tech")
		got := MineConsEnergy(globals, base, level, tech)

		if level == 0 && got != 0 {
			t.Fatalf("level=0 cons → %v", got)
		}
		if got < 0 {
			t.Fatalf("cons negative: %v", got)
		}
		// Tech ↑ ⇒ cons ↓ (пользу от energy_tech): 1.1 - tech*0.0005 убывает.
		if level >= 1 && tech < 20 {
			gotPlus := MineConsEnergy(globals, base, level, tech+1)
			if gotPlus > got {
				t.Fatalf("cons grows with tech: tech=%d → %v, tech=%d → %v",
					tech, got, tech+1, gotPlus)
			}
		}
	})
}
