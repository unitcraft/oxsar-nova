package alien

import "time"

// Config — параметры AlienAI, портированные 1-в-1 из oxsar2-classic
// consts.php:752-770. Все значения — defaults для legacy-вселенной;
// per-universe override будет реализован в Ф.3 плана 66 через
// configs/balance/<universe>.yaml.
//
// R15: значения семантически идентичны origin; менять без ADR
// запрещено (CLAUDE.md / §17 ТЗ).
type Config struct {
	// FleetsNumberNormal — лимит активных alien-миссий в обычные дни
	// (origin: ALIEN_NORMAL_FLEETS_NUMBER).
	FleetsNumberNormal int

	// FleetsNumberAttackTime — лимит в день атаки (четверг)
	// (origin: ALIEN_ATTACK_TIME_FLEETS_NUMBER = 50 × 5).
	FleetsNumberAttackTime int

	// AttackInterval — между alien-событиями на одного игрока
	// (origin: ALIEN_ATTACK_INTERVAL = 6 дней).
	AttackInterval time.Duration

	// GrabCreditInterval — между GRAB_CREDIT-событиями
	// (origin: ALIEN_GRAB_CREDIT_INTERVAL = 10 дней).
	GrabCreditInterval time.Duration

	// FlyMinTime/FlyMaxTime — диапазон времени полёта миссии
	// (origin: ALIEN_FLY_MIN_TIME = 15h, ALIEN_FLY_MAX_TIME = 24h).
	FlyMinTime time.Duration
	FlyMaxTime time.Duration

	// HaltingMinTime/MaxTime — длительность HOLDING (12-24ч).
	HaltingMinTime time.Duration
	HaltingMaxTime time.Duration

	// HaltingMaxRealTime — cap продления платежом
	// (origin: ALIEN_HALTING_MAX_REAL_TIME = 15 дней).
	HaltingMaxRealTime time.Duration

	// ChangeMissionMinTime/MaxTime — задержка для CHANGE_MISSION_AI
	// (origin: 8h..10h).
	ChangeMissionMinTime time.Duration
	ChangeMissionMaxTime time.Duration

	// GrabMinCredit — минимальный остаток кредитов цели для грабежа
	// (origin: ALIEN_GRAB_MIN_CREDIT = 100_000).
	// В nova по ADR-0009 — оксариты, не кредиты; Kind остался
	// исторически KindAlienGrabCredit. Имя поля сохраняем для
	// 1-в-1 чтения origin-кода.
	GrabMinCredit int64

	// GrabCreditMinPercent/MaxPercent — диапазон процента кражи
	// (origin: 0.08..0.10 → 0.08%..0.10% от баланса).
	GrabCreditMinPercent float64
	GrabCreditMaxPercent float64

	// GiftCreditMinPercent/MaxPercent — диапазон процента подарка
	// (origin: 5..10 → 5%..10% от баланса). Множитель ×0.01 как у
	// grab; в Ф.3 уточнится по сценарию подарка.
	GiftCreditMinPercent float64
	GiftCreditMaxPercent float64

	// MaxGiftCredit — cap абсолютного подарка
	// (origin: ALIEN_MAX_GIFT_CREDIT = 500).
	MaxGiftCredit int64

	// FleetMaxDebris — потолок возможной массы для generateFleet
	// (origin: ALIEN_FLEET_MAX_DERBIS = 1e9).
	FleetMaxDebris float64

	// ThursdayPowerMin/Max — диапазон множителя силы флота в день атаки
	// (origin: randFloatRange(1.5, 2.0)).
	ThursdayPowerMin float64
	ThursdayPowerMax float64

	// PowerScaleMin/Max — диапазон множителя силы в обычные дни
	// (origin: randFloatRange(0.9, 1.1)).
	PowerScaleMin float64
	PowerScaleMax float64

	// HoldingPaySecondsPerCredit — длительность продления HOLDING
	// за 1 единицу платежа (origin: 60*60*2 / 50 = 144 сек/кредит).
	// В nova — оксары (hard, ADR-0009 / R1), 1 оксар = 1 единица.
	HoldingPaySecondsPerCredit float64

	// FindTargetUserShipsMin / FindTargetPlanetShipsMin — пороги
	// findTarget (origin: 1000 у юзера, 100 на планете).
	FindTargetUserShipsMin   int64
	FindTargetPlanetShipsMin int64

	// FindCreditTargetUserShipsMin / PlanetShipsMin — пороги
	// findCreditTarget (origin: 300_000 / 10_000).
	FindCreditTargetUserShipsMin   int64
	FindCreditTargetPlanetShipsMin int64

	// SolarSatelliteTargetChance — вероятность выбора цели с
	// solar_satellite-only (origin: 10%).
	SolarSatelliteTargetChance int

	// ChangeMissionChance — вероятность спавна CHANGE_MISSION_AI
	// (origin: 60%).
	ChangeMissionChance int

	// FlyUnknownAttackChance — для onFlyUnknown: 90% становится
	// атакой в isAttackTime; 50% в обычные дни (origin:692).
	FlyUnknownAttackChanceThursday int
	FlyUnknownAttackChanceNormal   int

	// FlyUnknownGiftChance — 5% подарка ресурсами/кредитами
	// (origin:702, 739).
	FlyUnknownGiftChance int

	// FlyUnknownReplanChance — 10% «переадресовать» миссию
	// в новую generateMission через CHANGE_MISSION_AI (origin:774).
	FlyUnknownReplanChance int

	// FlyUnknownGrabChance — 10% при достаточном кредите забрать
	// (origin:665, "mt_rand(1, 100) <= 10").
	FlyUnknownGrabChance int

	// HoldingAIRecheckChance — 1% запустить checkAlientNeeds
	// в onHoldingAIEvent (origin:1006-1008).
	HoldingAIRecheckChance int

	// BuyoutBaseOxsars — фиксированная стоимость платного выкупа
	// HOLDING (план 66 Ф.5).
	//
	// Особенность: это НОВАЯ фича ремастера, в legacy oxsar2
	// (AlienAI.class.php) платного выкупа не существовало — там был
	// только paid_credit для продления окна HOLDING на 2h за 50
	// оксаритов (см. PHP:993, маппится на HoldingPaySecondsPerCredit
	// выше). Buyout же завершает HOLDING полностью и сразу за оксары
	// (hard currency, ADR-0009).
	//
	// Значение 100 выбрано как заметный, но не запретительный
	// price-point: средний игрок имеет 0-50 оксаров pre-purchase
	// (≈ ₽250 пакет), у активных премиум-игроков 200-1000+. Без ADR
	// «отклонение от legacy» — самой формулы в legacy нет, см.
	// docs/plans/66-remaster-alien-ai-full-parity.md «Ф.5 — итог».
	BuyoutBaseOxsars int64
}

// DefaultConfig возвращает Config с дефолтами 1-в-1 из origin
// consts.php:752-770 (зафиксировано на 2026-04-28).
//
// Используется в Ф.1+Ф.2 для прямого вызова helper'ов. В Ф.3
// runtime будет читать config из configs/balance/<universe>.yaml
// и накладывать override поверх дефолтов.
func DefaultConfig() Config {
	return Config{
		FleetsNumberNormal:             50,
		FleetsNumberAttackTime:         250,
		AttackInterval:                 6 * 24 * time.Hour,
		GrabCreditInterval:             10 * 24 * time.Hour,
		FlyMinTime:                     15 * time.Hour,
		FlyMaxTime:                     24 * time.Hour,
		HaltingMinTime:                 12 * time.Hour,
		HaltingMaxTime:                 24 * time.Hour,
		HaltingMaxRealTime:             15 * 24 * time.Hour,
		ChangeMissionMinTime:           8 * time.Hour,
		ChangeMissionMaxTime:           10 * time.Hour,
		GrabMinCredit:                  100_000,
		GrabCreditMinPercent:           0.08,
		GrabCreditMaxPercent:           0.10,
		GiftCreditMinPercent:           5,
		GiftCreditMaxPercent:           10,
		MaxGiftCredit:                  500,
		FleetMaxDebris:                 1_000_000_000,
		ThursdayPowerMin:               1.5,
		ThursdayPowerMax:               2.0,
		PowerScaleMin:                  0.9,
		PowerScaleMax:                  1.1,
		HoldingPaySecondsPerCredit:     2 * 3600.0 / 50.0, // 144 сек/оксар
		FindTargetUserShipsMin:         1000,
		FindTargetPlanetShipsMin:       100,
		FindCreditTargetUserShipsMin:   300_000,
		FindCreditTargetPlanetShipsMin: 10_000,
		SolarSatelliteTargetChance:     10,
		ChangeMissionChance:            60,
		FlyUnknownAttackChanceThursday: 90,
		FlyUnknownAttackChanceNormal:   50,
		FlyUnknownGiftChance:           5,
		FlyUnknownReplanChance:         10,
		FlyUnknownGrabChance:           10,
		HoldingAIRecheckChance:         1,
		BuyoutBaseOxsars:               100,
	}
}
