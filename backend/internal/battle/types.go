// Package battle реализует боевой движок.
//
// ПОРТ Java-движка d:\Sources\oxsar2-java (ADR-0002). Пакет содержит
// чистую функцию Calculate(input) -> report, без обращений к БД
// (I/O — ответственность вызывающего слоя). Полная реализация идёт
// в M4 (§16 ТЗ).
//
// Требования к паритету с Java-движком см. §5.7 и §14.4 ТЗ.
package battle

// Input — полный вход в боевой расчёт. Всё, что нужно движку, идёт
// сюда; никаких скрытых зависимостей.
type Input struct {
	Seed      uint64                `json:"seed"`
	Rounds    int                   `json:"rounds,omitempty"`
	Attackers []Side                `json:"attackers"`
	Defenders []Side                `json:"defenders"`
	Rapidfire map[int]map[int]int   `json:"rapidfire,omitempty"`
	IsMoon    bool                  `json:"is_moon,omitempty"`
}

// Side — одна сторона боя (один игрок или ACS-участник).
type Side struct {
	UserID        string `json:"user_id"`
	Username      string `json:"username,omitempty"`
	IsAliens      bool   `json:"is_aliens,omitempty"`
	Tech          Tech   `json:"tech,omitempty"`
	Units         []Unit `json:"units"`
	PrimaryTarget int    `json:"primary_target,omitempty"`
}

// Tech — уровни технологий игрока, влияющие на юнитов этой стороны.
type Tech struct {
	Gun        int `json:"gun,omitempty"`
	Shield     int `json:"shield,omitempty"`
	Shell      int `json:"shell,omitempty"`
	Laser      int `json:"laser,omitempty"`
	Ion        int `json:"ion,omitempty"`
	Plasma     int `json:"plasma,omitempty"`
	Ballistics int `json:"ballistics,omitempty"`
	Masking    int `json:"masking,omitempty"`
}

// Unit — юнит конкретной стороны. Поля соответствуют oxsar2-java Units.java.
type Unit struct {
	UnitID       int        `json:"unit_id"`
	Mode         int        `json:"mode,omitempty"`
	Quantity     int64      `json:"quantity"`
	Damaged      int64      `json:"damaged,omitempty"`
	ShellPercent float64    `json:"shell_percent,omitempty"`
	Front        int        `json:"front,omitempty"`
	Attack       [3]float64 `json:"attack"`
	Shield       [3]float64 `json:"shield,omitempty"`
	Shell        float64    `json:"shell"`
	Name         string     `json:"name,omitempty"`
	Cost         UnitCost   `json:"cost,omitempty"`
}

type UnitCost struct {
	Metal    int64 `json:"metal,omitempty"`
	Silicon  int64 `json:"silicon,omitempty"`
	Hydrogen int64 `json:"hydrogen,omitempty"`
}

// Report — результат боя.
type Report struct {
	Seed          uint64       `json:"seed"`
	Rounds        int          `json:"rounds"`
	Winner        string       `json:"winner"`
	RoundsTrace   []RoundTrace `json:"rounds_trace,omitempty"`
	Attackers     []SideResult `json:"attackers,omitempty"`
	Defenders     []SideResult `json:"defenders,omitempty"`
	DebrisMetal   int64        `json:"debris_metal,omitempty"`
	DebrisSilicon int64        `json:"debris_silicon,omitempty"`
	MoonChance    float64      `json:"moon_chance,omitempty"`
	MoonCreated   bool         `json:"moon_created,omitempty"`
}

type RoundTrace struct {
	Index          int   `json:"index"`
	AttackersAlive int64 `json:"attackers_alive"`
	DefendersAlive int64 `json:"defenders_alive"`
}

type SideResult struct {
	UserID       string       `json:"user_id"`
	Username     string       `json:"username,omitempty"`
	LostMetal    int64        `json:"lost_metal"`
	LostSilicon  int64        `json:"lost_silicon"`
	LostHydrogen int64        `json:"lost_hydrogen"`
	Units        []UnitResult `json:"units"`
}

type UnitResult struct {
	UnitID          int     `json:"unit_id"`
	QuantityStart   int64   `json:"quantity_start"`
	QuantityEnd     int64   `json:"quantity_end"`
	DamagedEnd      int64   `json:"damaged_end,omitempty"`
	ShellPercentEnd float64 `json:"shell_percent_end,omitempty"`
}
