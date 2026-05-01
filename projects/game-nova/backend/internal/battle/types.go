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
	Seed      uint64              `json:"seed"`
	Rounds    int                 `json:"rounds,omitempty"`
	NumSim    int                 `json:"num_sim,omitempty"` // 0/1 = один бой; 2..20 = multi-run статистика
	Attackers []Side              `json:"attackers"`
	Defenders []Side              `json:"defenders"`
	Rapidfire map[int]map[int]int `json:"rapidfire,omitempty"`
	// IsMoon — цель боя — луна (для UI-рендера заголовка раунда).
	// Не путать с HasPlanet — это отдельный флаг для опыта.
	IsMoon bool `json:"is_moon,omitempty"`
	// HasPlanet — есть ли у боя планетарная цель (= в legacy planetid != 0).
	// true — атака на планету или луну (Java: bpc *= 1).
	// false — бой в открытом космосе без planet'а (Java: bpc *= 0.5).
	// План 87 / BA-007: до фикса использовался IsMoon как proxy, но
	// семантика противоположна и обычные атаки планет получали ×0.5
	// опыта.
	HasPlanet bool `json:"has_planet,omitempty"`
}

// Side — одна сторона боя (один игрок или ACS-участник).
type Side struct {
	UserID        string `json:"user_id"`
	Username      string `json:"username,omitempty"`
	IsAliens      bool   `json:"is_aliens,omitempty"`
	Tech          Tech   `json:"tech,omitempty"`
	Units         []Unit `json:"units"`
	PrimaryTarget int    `json:"primary_target,omitempty"`

	// Позиция планеты атакующего/защитника. Используется в заголовке
	// каждого раунда: «Атакующий <username> [g:s:p] (луна)» — порт Java
	// Participant.getGalaxy/getSystem/getPosition/getIsMoon (план 72.1
	// ч.20.11.9). Для симулятора фронт может не передавать — рендерится
	// «[0:0:0]».
	Galaxy   int  `json:"galaxy,omitempty"`
	System   int  `json:"system,omitempty"`
	Position int  `json:"position,omitempty"`
	IsMoon   bool `json:"is_moon,omitempty"`
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
	Attack       float64    `json:"attack"`
	Shield       float64    `json:"shield,omitempty"`
	Shell        float64    `json:"shell"`
	Name         string     `json:"name,omitempty"`
	Cost         UnitCost   `json:"cost,omitempty"`
}

type UnitCost struct {
	Metal    int64 `json:"metal,omitempty"`
	Silicon  int64 `json:"silicon,omitempty"`
	Hydrogen int64 `json:"hydrogen,omitempty"`
}

// SimStats — агрегат по num_sim прогонам (план 72.1 ч.20.11.7).
// Pixel-perfect клон сводки legacy simulator.tpl: победа атакующего/
// обороняющегося/ничья в %, среднее число раундов, средний шанс луны,
// средние потери металла/кремния/водорода и опыта обеих сторон,
// средние обломки на орбите, время симуляции (общее и одной итерации).
type SimStats struct {
	NumSim   int `json:"num_sim"`

	// Доли исходов в %.
	AttackerWinPct float64 `json:"attacker_win_pct"`
	DefenderWinPct float64 `json:"defender_win_pct"`
	DrawPct        float64 `json:"draw_pct"`

	AvgRounds     float64 `json:"avg_rounds"`
	AvgMoonChance float64 `json:"avg_moon_chance"` // %

	AttackerLostMetal    float64 `json:"attacker_lost_metal"`
	AttackerLostSilicon  float64 `json:"attacker_lost_silicon"`
	AttackerLostHydrogen float64 `json:"attacker_lost_hydrogen"`
	DefenderLostMetal    float64 `json:"defender_lost_metal"`
	DefenderLostSilicon  float64 `json:"defender_lost_silicon"`
	DefenderLostHydrogen float64 `json:"defender_lost_hydrogen"`

	DebrisMetal   float64 `json:"debris_metal"`
	DebrisSilicon float64 `json:"debris_silicon"`

	AttackerExp float64 `json:"attacker_exp"`
	DefenderExp float64 `json:"defender_exp"`

	GenTimeAll float64 `json:"gen_time_all"` // секунды всего
	GenTime    float64 `json:"gen_time"`     // секунды на одну симуляцию
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

	// AttackerExp / DefenderExp — очки опыта (Java attackerExperience /
	// defenderExperience). atan-based формула, см. computeExperience в
	// engine.go (порт Java Assault.java:817-847).
	AttackerExp int `json:"attacker_exp,omitempty"`
	DefenderExp int `json:"defender_exp,omitempty"`

	// HaulMetal / HaulSilicon / HaulHydrogen — трофеи атакующего при
	// победе (Java haulMetal/Silicon/Hydrogen). Для симулятора всегда 0
	// (нет реальной планеты-цели), для реального боя через миссию —
	// заполняется fleet/attack handler'ом после Calculate.
	HaulMetal    int64 `json:"haul_metal,omitempty"`
	HaulSilicon  int64 `json:"haul_silicon,omitempty"`
	HaulHydrogen int64 `json:"haul_hydrogen,omitempty"`
}

// RoundTrace — полная статистика раунда боя, пixel-perfect клон
// oxsar2-java/Assault.java rendering (план 72.1 ч.20.11.4).
type RoundTrace struct {
	Index          int       `json:"index"`
	AttackersAlive int64     `json:"attackers_alive"`
	DefendersAlive int64     `json:"defenders_alive"`
	AttackerSide   RoundSide `json:"attacker_side"`
	DefenderSide   RoundSide `json:"defender_side"`
}

// RoundSide — статистика одной стороны (атакующих или защитников)
// в одном раунде.
type RoundSide struct {
	// Username + позиция планеты для заголовка раунда (Java:
	// «{ATTACKER} <user> [g:s:p] (луна)»).
	Username string `json:"username,omitempty"`
	Galaxy   int    `json:"galaxy,omitempty"`
	System   int    `json:"system,omitempty"`
	Position int    `json:"position,omitempty"`
	IsMoon   bool   `json:"is_moon,omitempty"`

	// Tech-power процентами (для отображения в стиле Java
	// "GUN_POWER: 50%": привязано к gun_level * 10).
	GunPowerPct    float64 `json:"gun_power_pct"`
	ShieldPowerPct float64 `json:"shield_power_pct"`
	ArmoringPct    float64 `json:"armoring_pct"`
	BallisticsLvl  int     `json:"ballistics_lvl"`
	MaskingLvl     int     `json:"masking_lvl"`

	// Агрегаты «Fight» таблицы (Java: attackerShots / attackerPower).
	Shots          int64   `json:"shots"`
	Power          float64 `json:"power"`
	ShieldAbsorbed float64 `json:"shield_absorbed"`
	ShellDestroyed float64 `json:"shell_destroyed"`
	UnitsDestroyed int64   `json:"units_destroyed"`

	// Per-unit snapshot после finishTurn этого раунда (Java startTurn*).
	Units []RoundUnit `json:"units"`
}

// RoundUnit — статус юнита на начало раунда (после регенерации
// предыдущего раунда). Это «Type / Quantity / Guns / Shields / Shells
// / Front / Ballistics / Masking / Survival%» из Java printParticipant.
type RoundUnit struct {
	UnitID                int     `json:"unit_id"`
	Name                  string  `json:"name,omitempty"`
	StartTurnQuantity     int64   `json:"start_turn_quantity"`
	StartTurnQuantityDiff int64   `json:"start_turn_quantity_diff"` // отрицательное = потери прошлого раунда
	StartTurnDamaged      int64   `json:"start_turn_damaged"`
	DamagedShellPercent   int     `json:"damaged_shell_percent"`
	Attack                float64 `json:"attack"`
	Shield                float64 `json:"shield"`
	Shell                 float64 `json:"shell"`
	Front                 int     `json:"front"`
	BallisticsLevel       int     `json:"ballistics_level,omitempty"`
	MaskingLevel          int     `json:"masking_level,omitempty"`
	StartBattleQuantity   int64   `json:"start_battle_quantity"`
	AlivePercent          int     `json:"alive_percent"`
}

type SideResult struct {
	UserID       string       `json:"user_id"`
	Username     string       `json:"username,omitempty"`
	IsAliens     bool         `json:"is_aliens,omitempty"`
	LostMetal    int64        `json:"lost_metal"`
	LostSilicon  int64        `json:"lost_silicon"`
	LostHydrogen int64        `json:"lost_hydrogen"`
	// LostPoints — суммарные очки потерянных юнитов (план 72.1.1):
	//   Σ qty_lost × (cost.metal + cost.silicon + cost.hydrogen) / 1000 × 2
	// Используется в `users.points -= LostPoints` после реального боя
	// (порт Java Participant.java:766, Units.java:113-115).
	LostPoints float64 `json:"lost_points"`
	// LostUnits — суммарное число потерянных юнитов (Σ qty_lost).
	// Используется в `users.u_count -= LostUnits` (Java
	// Participant.java:759-761).
	LostUnits int64        `json:"lost_units"`
	Units     []UnitResult `json:"units"`
}

type UnitResult struct {
	UnitID          int     `json:"unit_id"`
	QuantityStart   int64   `json:"quantity_start"`
	QuantityEnd     int64   `json:"quantity_end"`
	DamagedEnd      int64   `json:"damaged_end,omitempty"`
	ShellPercentEnd float64 `json:"shell_percent_end,omitempty"`
}
