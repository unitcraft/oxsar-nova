package alien

import (
	"math"
	"time"

	"oxsar/game-nova/pkg/rng"
)

// IsAttackTime — порт PHP AlienAI::isAttackTime (строки 15-22).
//
// Возвращает true если переданное время — четверг.
// Origin: `date("w") == 4`. Day-of-week PHP считает в TZ сервера;
// в nova принимаем явно — caller передаёт уже нужную TZ.
func IsAttackTime(t time.Time) bool {
	return t.Weekday() == time.Thursday
}

// RandRoundRange — порт PHP randRoundRange($min, $max).
//
// Возвращает целое в [min, max], при условии min > 0. Используется
// для выбора длительности полёта/удержания (origin: round до сек).
func RandRoundRange(minSec, maxSec int64, r *rng.R) int64 {
	if minSec >= maxSec {
		return minSec
	}
	span := maxSec - minSec + 1
	return minSec + int64(r.IntN(int(span)))
}

// RandRoundRangeDur — то же, но для time.Duration.
func RandRoundRangeDur(min, max time.Duration, r *rng.R) time.Duration {
	if min >= max {
		return min
	}
	span := max.Nanoseconds() - min.Nanoseconds() + 1
	return min + time.Duration(r.Uint64()%uint64(span))
}

// RandFloatRange — порт PHP randFloatRange($min, $max).
//
// Возвращает float64 в [min, max).
func RandFloatRange(min, max float64, r *rng.R) float64 {
	if min >= max {
		return min
	}
	return min + r.Float64()*(max-min)
}

// HoldingDuration выбирает случайную длительность HOLDING
// (origin: randRoundRange(ALIEN_HALTING_MIN_TIME, ALIEN_HALTING_MAX_TIME)).
func HoldingDuration(cfg Config, r *rng.R) time.Duration {
	return RandRoundRangeDur(cfg.HaltingMinTime, cfg.HaltingMaxTime, r)
}

// FlightDuration — длительность полёта alien-флота к цели
// (origin: randRoundRange(ALIEN_FLY_MIN_TIME, ALIEN_FLY_MAX_TIME)).
//
// В nova реальное время полёта определяется ещё и расстоянием
// (alienFlightDuration в internal/alien/alien.go), но AlienAI
// в origin задаёт целевое окно прибытия независимо от расстояния —
// это семантика «пришельцы появляются из глубокого космоса».
func FlightDuration(cfg Config, r *rng.R) time.Duration {
	return RandRoundRangeDur(cfg.FlyMinTime, cfg.FlyMaxTime, r)
}

// ChangeMissionDelay — задержка для CHANGE_MISSION_AI
// (origin AlienAI.class.php:221-226):
//
//   60% — randRoundRange(ALIEN_CHANGE_MISSION_MIN_TIME,
//                        ALIEN_CHANGE_MISSION_MAX_TIME);
//         capped до flightTime - 10 сек
//   40% — randRoundRange(flightTime - 30, flightTime - 10).
//
// Возвращает offset от now (положительная Duration).
func ChangeMissionDelay(cfg Config, flight time.Duration, r *rng.R) time.Duration {
	if flight <= 30*time.Second {
		return flight - 10*time.Second
	}
	var d time.Duration
	if r.IntN(100) < 60 {
		d = RandRoundRangeDur(cfg.ChangeMissionMinTime, cfg.ChangeMissionMaxTime, r)
		max := flight - 10*time.Second
		if d > max {
			d = max
		}
	} else {
		d = RandRoundRangeDur(flight-30*time.Second, flight-10*time.Second, r)
	}
	if d < 0 {
		d = 0
	}
	return d
}

// HoldingExtension — длительность продления HOLDING за платёж
// (origin AlienAI.class.php:993):
//
//   end_time += 2*60*60 * paid_credit / 50
//   end_time = min(end_time, start_time + ALIEN_HALTING_MAX_REAL_TIME)
//
// В nova paid_credit — это оксары (R1, ADR-0009).
//
// Возвращает новый holds_until, не превышающий cap-а.
func HoldingExtension(cfg Config, startAt, holdsUntil time.Time, paidHard int64) time.Time {
	if paidHard <= 0 {
		return holdsUntil
	}
	add := time.Duration(float64(paidHard)*cfg.HoldingPaySecondsPerCredit) * time.Second
	out := holdsUntil.Add(add)
	cap := startAt.Add(cfg.HaltingMaxRealTime)
	if out.After(cap) {
		return cap
	}
	return out
}

// MaxRealEndAt — абсолютный потолок HOLDING (start + 15 дней).
func MaxRealEndAt(cfg Config, startAt time.Time) time.Time {
	return startAt.Add(cfg.HaltingMaxRealTime)
}

// PowerScaleNormal — масштаб силы для обычного дня (0.9..1.1).
func PowerScaleNormal(cfg Config, r *rng.R) float64 {
	return RandFloatRange(cfg.PowerScaleMin, cfg.PowerScaleMax, r)
}

// PowerScaleThursday — масштаб силы в четверг (1.5..2.0).
func PowerScaleThursday(cfg Config, r *rng.R) float64 {
	return RandFloatRange(cfg.ThursdayPowerMin, cfg.ThursdayPowerMax, r)
}

// PowerScaleAfterControlTimes — формула роста силы при цепочке
// CHANGE_MISSION_AI (origin AlienAI.class.php:884):
//
//   power_scale = 1 + control_times * 1.5
//
// Растёт квадратично через цепочку (controlTimes++ на каждой смене).
func PowerScaleAfterControlTimes(controlTimes int) float64 {
	return 1.0 + float64(controlTimes)*1.5
}

// HoldingAISubphaseDuration — длительность подфазы HOLDING_AI
// (origin AlienAI.class.php:974):
//
//   duration = clamp(min(12h, 30s*times) ... max(24h, 60s*times))
//
// Где times — control_times.
func HoldingAISubphaseDuration(cfg Config, controlTimes int, r *rng.R) time.Duration {
	hi := maxDur(cfg.HaltingMaxTime, time.Duration(60*controlTimes)*time.Second)
	lo := minDur(cfg.HaltingMinTime, time.Duration(30*controlTimes)*time.Second)
	if lo > hi {
		lo = hi
	}
	return RandRoundRangeDur(lo, hi, r)
}

// minDur / maxDur — для time.Duration (math.Min не работает с time.Duration).
func minDur(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

func maxDur(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

// CalcGrabAmount — сколько оксаритов украсть у игрока
// (origin AlienAI.class.php:667):
//
//   grab = round(user_credit * 0.01 * randFloat(MIN%, MAX%), 2)
//        = user_credit * (0.0008..0.001)
//
// Возвращает grab; 0 если не подходит порог.
func CalcGrabAmount(cfg Config, userCredit int64, r *rng.R) int64 {
	if userCredit <= cfg.GrabMinCredit {
		return 0
	}
	pct := 0.01 * RandFloatRange(cfg.GrabCreditMinPercent, cfg.GrabCreditMaxPercent, r)
	grab := math.Round(float64(userCredit) * pct)
	if grab <= 0 {
		return 0
	}
	return int64(grab)
}

// CalcGiftAmount — сколько оксаритов подарить игроку
// (origin AlienAI.class.php:739-740):
//
//   max_gift = ALIEN_MAX_GIFT_CREDIT * randFloat(0.98, 1.02)
//   gift = min(max_gift, user_credit * 0.01 * randFloat(5, 10))
func CalcGiftAmount(cfg Config, userCredit int64, r *rng.R) int64 {
	maxG := float64(cfg.MaxGiftCredit) * RandFloatRange(0.98, 1.02, r)
	cap := math.Round(maxG)
	pct := 0.01 * RandFloatRange(cfg.GiftCreditMinPercent, cfg.GiftCreditMaxPercent, r)
	g := math.Round(float64(userCredit) * pct)
	if g > cap {
		g = cap
	}
	if g <= 0 {
		return 0
	}
	return int64(g)
}
