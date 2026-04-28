package alien

// План 66 Ф.6: golden-итерации формул AlienAI.
//
// Загружает testdata/golden_alien_ai.json (сгенерирован
// projects/game-origin-php/tools/dump-alien-ai.php) и для каждого
// кейса вызывает соответствующий Go-helper. Для range-кейсов
// проверяет [expected_min, expected_max]; для exact-кейсов
// (expected_min == expected_max) — точное совпадение.
//
// Почему range, а не bit-в-bit: PHP mt_rand (Mersenne Twister) ≠
// Go pkg/rng (xorshift64*). Бит-совместимость потребует портирования
// mt_rand в Go (R8/Ф.6 ТЗ — future work, см. shuffle.go:115).
// Golden работает на уровне ИНВАРИАНТОВ ФОРМУЛЫ origin AlienAI.
//
// Auto-skip если testdata-файл отсутствует (CI без PHP-генератора).

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"oxsar/game-nova/pkg/rng"
)

type goldenCase struct {
	ID          string         `json:"id"`
	Group       string         `json:"group"`
	Fn          string         `json:"fn"`
	Input       map[string]any `json:"input"`
	ExpectedMin float64        `json:"expected_min"`
	ExpectedMax float64        `json:"expected_max"`
	Comment     string         `json:"comment"`
}

const goldenPath = "testdata/golden_alien_ai.json"

func loadGolden(t *testing.T) []goldenCase {
	t.Helper()
	data, err := os.ReadFile(goldenPath)
	if errors.Is(err, fs.ErrNotExist) {
		t.Skipf("%s not present; regenerate via projects/game-origin-php/tools/dump-alien-ai.php",
			goldenPath)
	}
	if err != nil {
		t.Fatalf("read %s: %v", goldenPath, err)
	}
	var cases []goldenCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatalf("parse %s: %v", goldenPath, err)
	}
	if len(cases) < 50 {
		t.Fatalf("golden file has %d cases, want ≥50 (план 66 Ф.6)", len(cases))
	}
	return cases
}

// inputInt извлекает целое число из input map (json парсит как float64).
func inputInt(t *testing.T, in map[string]any, key string) int64 {
	t.Helper()
	v, ok := in[key]
	if !ok {
		t.Fatalf("missing input key %q", key)
	}
	f, ok := v.(float64)
	if !ok {
		t.Fatalf("input %q: want number, got %T", key, v)
	}
	return int64(f)
}

func TestGolden_AllCases(t *testing.T) {
	cases := loadGolden(t)
	cfg := DefaultConfig()

	groupCounts := make(map[string]int)
	for _, c := range cases {
		groupCounts[c.Group]++
		t.Run(c.ID, func(t *testing.T) {
			runGolden(t, c, cfg)
		})
	}

	// Гарантия покрытия групп — план требует ≥5 групп.
	if len(groupCounts) < 5 {
		t.Errorf("only %d groups in golden; want ≥5 (план 66 Ф.6)", len(groupCounts))
	}
	t.Logf("golden cases: %d total across %d groups: %v",
		len(cases), len(groupCounts), groupCounts)
}

func runGolden(t *testing.T, c goldenCase, cfg Config) {
	t.Helper()

	// Helper для exact / range проверки.
	checkInt64 := func(got int64) {
		gotF := float64(got)
		// Exact (expected_min == expected_max).
		if c.ExpectedMin == c.ExpectedMax {
			if gotF != c.ExpectedMin {
				t.Fatalf("[%s] %s exact mismatch: got=%d want=%v (%s)",
					c.ID, c.Fn, got, c.ExpectedMin, c.Comment)
			}
			return
		}
		if gotF < c.ExpectedMin || gotF > c.ExpectedMax {
			t.Fatalf("[%s] %s out of range: got=%d want ∈[%v..%v] (%s)",
				c.ID, c.Fn, got, c.ExpectedMin, c.ExpectedMax, c.Comment)
		}
	}
	checkFloat := func(got, eps float64) {
		if c.ExpectedMin == c.ExpectedMax {
			if absF(got-c.ExpectedMin) > eps {
				t.Fatalf("[%s] %s exact mismatch: got=%v want=%v eps=%v (%s)",
					c.ID, c.Fn, got, c.ExpectedMin, eps, c.Comment)
			}
			return
		}
		if got < c.ExpectedMin-eps || got > c.ExpectedMax+eps {
			t.Fatalf("[%s] %s out of range: got=%v want ∈[%v..%v] (%s)",
				c.ID, c.Fn, got, c.ExpectedMin, c.ExpectedMax, c.Comment)
		}
	}
	checkDur := func(got time.Duration) {
		// expected_min/max в секундах.
		gotSec := float64(got / time.Second)
		if gotSec < c.ExpectedMin || gotSec > c.ExpectedMax {
			t.Fatalf("[%s] %s out of range: got=%vs want ∈[%v..%v]s (%s)",
				c.ID, c.Fn, gotSec, c.ExpectedMin, c.ExpectedMax, c.Comment)
		}
	}

	switch c.Fn {
	case "CalcGrabAmount":
		credit := inputInt(t, c.Input, "user_credit")
		seed := uint64(inputInt(t, c.Input, "seed"))
		got := CalcGrabAmount(cfg, credit, rng.New(seed))
		checkInt64(got)

	case "CalcGiftAmount":
		credit := inputInt(t, c.Input, "user_credit")
		seed := uint64(inputInt(t, c.Input, "seed"))
		got := CalcGiftAmount(cfg, credit, rng.New(seed))
		checkInt64(got)

	case "HoldingExtension":
		startTs := inputInt(t, c.Input, "start_ts")
		holdsTs := inputInt(t, c.Input, "holds_ts")
		paid := inputInt(t, c.Input, "paid_hard")
		start := time.Unix(startTs, 0).UTC()
		holds := time.Unix(holdsTs, 0).UTC()
		got := HoldingExtension(cfg, start, holds, paid)
		// expected_min/max — unix-секунды результата.
		checkInt64(got.Unix())

	case "PowerScaleAfterControlTimes":
		ct := int(inputInt(t, c.Input, "control_times"))
		got := PowerScaleAfterControlTimes(ct)
		checkFloat(got, 1e-9)

	case "HoldingDuration":
		seed := uint64(inputInt(t, c.Input, "seed"))
		got := HoldingDuration(cfg, rng.New(seed))
		checkDur(got)

	case "FlightDuration":
		seed := uint64(inputInt(t, c.Input, "seed"))
		got := FlightDuration(cfg, rng.New(seed))
		checkDur(got)

	case "ChangeMissionDelay":
		flightSec := inputInt(t, c.Input, "flight_seconds")
		seed := uint64(inputInt(t, c.Input, "seed"))
		got := ChangeMissionDelay(cfg, time.Duration(flightSec)*time.Second, rng.New(seed))
		// Range проверка в секундах.
		checkDur(got)

	case "HoldingAISubphaseDuration":
		ct := int(inputInt(t, c.Input, "control_times"))
		seed := uint64(inputInt(t, c.Input, "seed"))
		got := HoldingAISubphaseDuration(cfg, ct, rng.New(seed))
		checkDur(got)

	case "WeakenedTechLevel":
		level := int(inputInt(t, c.Input, "level"))
		seed := uint64(inputInt(t, c.Input, "seed"))
		got := weakenedTechLevel(level, rng.New(seed))
		checkInt64(int64(got))

	default:
		t.Fatalf("[%s] unknown fn %q — add dispatcher case in golden_test.go",
			c.ID, c.Fn)
	}
}

func absF(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// TestGolden_FilePath — sanity: путь существует или явно skip.
// Помогает поймать опечатку в goldenPath.
func TestGolden_FilePath(t *testing.T) {
	abs, err := filepath.Abs(goldenPath)
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	if _, err := os.Stat(abs); errors.Is(err, fs.ErrNotExist) {
		t.Skipf("%s missing (regenerate via dump-alien-ai.php)", abs)
	}
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
}

// _ — компилятор-страж: fmt-import нужен для будущих golden-кейсов с
// форматированным сообщением (сейчас Errorf-строки используют его в
// runGolden косвенно).
var _ = fmt.Sprintf
