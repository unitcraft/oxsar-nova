package alien

import "time"

// MissionMode — режим alien-миссии. Совпадает по семантике с PHP
// constants EVENT_ALIEN_* (origin: consts.php:440-445).
type MissionMode int

const (
	ModeFlyUnknown   MissionMode = 33 // EVENT_ALIEN_FLY_UNKNOWN
	ModeHolding      MissionMode = 34
	ModeAttack       MissionMode = 35
	ModeHalt         MissionMode = 36
	ModeGrabCredit   MissionMode = 37
	ModeAttackCustom MissionMode = 38 // admin-only spawn
	ModeHoldingAI    MissionMode = 80
	ModeChangeMission MissionMode = 81
)

// FleetUnit — один тип юнита в составе alien-флота.
//
// Семантически 1-в-1 с записью PHP $fleet[$id] из generateFleet:
// {id, quantity, damaged, shell_percent, name}.
type FleetUnit struct {
	UnitID       int    `json:"unit_id"`
	Name         string `json:"name,omitempty"`
	Quantity     int64  `json:"quantity"`
	Damaged      int64  `json:"damaged,omitempty"`
	ShellPercent int    `json:"shell_percent"` // 0..100
}

// Fleet — состав alien-флота (или target-флота для расчётов).
type Fleet []FleetUnit

// Mission — миссия пришельцев на цель. Не event-payload как таковой,
// а внутренняя структура AI до записи в events.payload. R13 typed
// payload для KindAlienAttack/Holding/FlyUnknown/* строится в Ф.3
// поверх Mission.
type Mission struct {
	Mode        MissionMode `json:"mode"`
	UserID      string      `json:"user_id"`      // владелец цели
	PlanetID    string      `json:"planet_id"`    // цель
	AlienFleet  Fleet       `json:"alien_fleet"`
	FlightTime  time.Duration `json:"flight_time"`
	HoldingTime time.Duration `json:"holding_time"`

	// Resource snapshot — backstory для подарков/грабежа,
	// в origin лежит в events.data.metal/silicon/hydrogen.
	Metal    int64 `json:"metal"`
	Silicon  int64 `json:"silicon"`
	Hydrogen int64 `json:"hydrogen"`

	// ControlTimes — счётчик «итераций AI», увеличивается на
	// CHANGE_MISSION_AI и HOLDING_AI (PHP $event[data][control_times]).
	ControlTimes int `json:"control_times"`

	// PowerScale — множитель силы при generateFleet (1.5..2.0 в четверг,
	// 0.9..1.1 в обычные дни; растёт через CHANGE_MISSION_AI).
	PowerScale float64 `json:"power_scale"`

	// AlienActor — флаг для UI/логов, что событие создал AI.
	AlienActor bool `json:"alien_actor"`

	// AddTech — модифицированные tech-уровни цели после shuffleKeyValues
	// (используются Assault как «фактические» уровни в момент боя).
	// Ключ — tech ID (см. economy.IDTech*).
	AddTech map[int]int `json:"add_tech,omitempty"`
}

// HoldingState — состояние HOLDING-события. R13 typed payload.
//
// В Ф.3 будет сериализоваться в events.payload как JSON; в Ф.1
// нужна для сигнатур helper'ов.
type HoldingState struct {
	EventID      string    `json:"event_id"`
	PlanetID     string    `json:"planet_id"`
	UserID       string    `json:"user_id"`
	StartAt      time.Time `json:"start_at"`
	HoldsUntilAt time.Time `json:"holds_until_at"` // R1: имя поля БД
	MaxRealEndAt time.Time `json:"max_real_end_at"`
	AlienFleet   Fleet     `json:"alien_fleet"`
	ControlTimes int       `json:"control_times"`
	// PaidHard — суммарно потрачено оксаров на продление (R1: hard
	// валюта, ADR-0009; имя поля исторически paid_credit в PHP).
	PaidHard int64 `json:"paid_hard"`
	PaidTimes int  `json:"paid_times"`
}

// PlanetSnapshot — минимальное представление цели в момент выбора
// (loadPlanetShips + ресурсы). Используется generateMission и
// generateFleet.
type PlanetSnapshot struct {
	UserID   string
	PlanetID string
	Galaxy   int
	System   int
	Position int
	Metal    int64
	Silicon  int64
	Hydrogen int64
	Ships    Fleet // unit_id → quantity (без UNIT_SOLAR_SATELLITE; PHP:269)
}

// TechProfile — уровни исследований игрока (loadUserResearches).
// Ключ — economy.IDTech* константа.
type TechProfile map[int]int

// Clone возвращает копию TechProfile (для shuffleKeyValues, чтобы
// не мутировать caller'ову карту — в Go это важно, в отличие от
// PHP с pass-by-reference).
func (t TechProfile) Clone() TechProfile {
	out := make(TechProfile, len(t))
	for k, v := range t {
		out[k] = v
	}
	return out
}
