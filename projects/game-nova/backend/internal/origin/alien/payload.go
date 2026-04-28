package alien

import "encoding/json"

// MissionPayload — typed-payload (R13) для KindAlienFlyUnknown,
// KindAlienGrabCredit, KindAlienAttackCustom, KindAlienAttack
// (когда событие создано через AlienAI.generateMission).
//
// Семантика origin (events.data в PHP):
//   - mode = EVENT_ALIEN_*  (PHP int constant)
//   - ships — флот пришельцев (id → quantity / damaged / shell_percent)
//   - metal/silicon/hydrogen — снимок ресурсов цели на момент создания
//   - control_times — счётчик AI-итераций (растёт через CHANGE_MISSION_AI)
//   - power_scale — множитель силы для будущих generateFleet
//   - alien_actor — флаг (всегда true для AI-созданных)
//   - add_tech_<id> — модифицированные tech-уровни (после shuffle+
//     weakening), используются Assault как «фактические» уровни цели
//
// В nova ID полей snake_case (как везде в payload-структурах nova).
type MissionPayload struct {
	// Mode — EVENT_ALIEN_* (33/34/35/36/37/38/80/81). Из MissionMode.
	Mode int `json:"mode"`
	// Tier — из существующего alienPayload (для совместимости с
	// internal/alien.Service.AttackHandler в Ф.4 — отдельный Kind=35
	// payload пока остаётся другим).
	Tier int `json:"tier,omitempty"`

	UserID   string `json:"user_id"`
	PlanetID string `json:"planet_id"`
	Galaxy   int    `json:"galaxy"`
	System   int    `json:"system"`
	Position int    `json:"position"`

	// Ships — alien-флот (UnitID → данные). Сохраняем как Fleet
	// (List of FleetUnit) для упорядоченности и определённости JSON.
	Ships Fleet `json:"ships"`

	// Snapshot ресурсов цели на момент создания миссии (origin: events.data).
	Metal    int64 `json:"metal"`
	Silicon  int64 `json:"silicon"`
	Hydrogen int64 `json:"hydrogen"`

	ControlTimes int     `json:"control_times"`
	PowerScale   float64 `json:"power_scale"`
	AlienActor   bool    `json:"alien_actor"`

	// AddTech — tech_id → видимый AI уровень (после ShuffleAllAlienTechGroups
	// + ApplyShuffledTechWeakening). Передаётся в Assault при бою.
	AddTech map[int]int `json:"add_tech,omitempty"`
}

// MarshalPayload — JSON-marshal с детерминированным порядком ключей
// (стандартный encoding/json уже даёт sort'd order для map'ов начиная
// с Go 1.12). Helper только для согласования сигнатуры с другими
// payload'ами nova.
func (p *MissionPayload) MarshalPayload() ([]byte, error) {
	return json.Marshal(p)
}

// ChangeMissionPayload — typed-payload для KindAlienChangeMissionAI.
//
// Семантика origin (AlienAI.class.php:864-921):
//   - control_times — счётчик AI-итераций
//   - alien_actor — true
//   - parent_event — UUID родительского ATTACK/FLY_UNKNOWN
//
// В отличие от origin, где `parent_eventid` — отдельная колонка events,
// в nova nova/events не имеет parent_eventid (см. divergence-log D-024).
// Поэтому переносим ID в payload.
type ChangeMissionPayload struct {
	ParentEventID string `json:"parent_event_id"`
	UserID        string `json:"user_id"`
	PlanetID      string `json:"planet_id"`
	ControlTimes  int    `json:"control_times"`
	AlienActor    bool   `json:"alien_actor"`
}
