package alien

import "time"

// HoldingAIPayload — typed-payload (R13) для KindAlienHoldingAI (Kind=80).
//
// JSON-shape совместима с `internal/alien.holdingPayload` (старый пакет
// nova): HALT-handler создаёт KindAlienHoldingAI с тем же json-marshal'ом,
// поэтому при переключении регистрации worker'а (план 66 Ф.4) новый
// handler читает payload из существующего HALT-цикла без ломки
// сериализации. Поля ControlTimes/PaidSumCredit/PaidCredit добавлены
// `omitempty` — отсутствуют в старом payload и интерпретируются как 0
// при первом тике.
//
// Семантика origin (`AlienAI.class.php:931-989`, `events.data`):
//
//   - control_times — счётчик AI-итераций (растёт на каждом тике).
//   - paid_credit — единоразовый платёж за продление (потребляется в
//     этом тике, после обнуляется в payload).
//   - paid_sum_credit — суммарный платёж за всё HOLDING (для
//     сообщения в HoldingHandler / отчётов).
//   - paid_times — количество продлений.
//
// HoldingEventID — UUID parent KindAlienHolding события (нужен и
// для проверки актуальности parent, и для продления fire_at platежом).
type HoldingAIPayload struct {
	PlanetID       string             `json:"planet_id"`
	UserID         string             `json:"user_id"`
	Tier           int                `json:"tier"`
	AlienFleet     []HoldingFleetUnit `json:"alien_fleet"`
	StartTime      time.Time          `json:"start_time"`
	HoldingEventID string             `json:"holding_event_id,omitempty"`
	ControlTimes   int                `json:"control_times,omitempty"`
	PaidCredit     int64              `json:"paid_credit,omitempty"`
	PaidSumCredit  int64              `json:"paid_sum_credit,omitempty"`
	PaidTimes      int                `json:"paid_times,omitempty"`

	// Snapshot ресурсов планеты, ранее захваченных пришельцами
	// (origin: parent_event["data"]["metal"/"silicon"/"hydrogen"]).
	// Используется в SubphaseUnloadAlienResources для расчёта подарка
	// игроку. nova HALT/HOLDING может не сохранять snapshot (план 15
	// упрощение, см. simplifications.md "[Alien AI] unloadAlienResources
	// — процент от текущих, не от захваченных") — тогда поля 0 и
	// unload даёт 0 ресурсов, как и в origin при пустом snapshot.
	Metal    int64 `json:"metal,omitempty"`
	Silicon  int64 `json:"silicon,omitempty"`
	Hydrogen int64 `json:"hydrogen,omitempty"`
}

// HoldingFleetUnit — компактный снимок alien-флота в payload
// (без Attack/Shield — restored из catalog при загрузке).
//
// JSON-теги совпадают с `internal/alien.fleetStack` (UnitID/Quantity).
type HoldingFleetUnit struct {
	UnitID   int   `json:"unit_id"`
	Quantity int64 `json:"quantity"`
}

// HoldingParentSnapshot — поля parent KindAlienHolding события, которые
// HOLDING_AI читает на каждом тике и потенциально модифицирует.
//
// Origin хранит `data.metal/silicon/hydrogen` как ресурсы, ранее
// захваченные на планете (для подарков из onUnloadAlienResoursesAI
// PHP:1058). В nova HALT/HOLDING сохраняет их в `holdingPayload` через
// resource-snapshot (см. план 15 этап 2). При первом тике эти поля
// могут быть 0 — тогда unload-action даёт 0 ресурсов, как и в origin
// при пустом snapshot.
type HoldingParentSnapshot struct {
	Metal    int64 `json:"metal"`
	Silicon  int64 `json:"silicon"`
	Hydrogen int64 `json:"hydrogen"`
}
