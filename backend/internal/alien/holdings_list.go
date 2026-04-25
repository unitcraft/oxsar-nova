package alien

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/oxsar/nova/backend/internal/event"
	"github.com/oxsar/nova/backend/internal/repo"
)

// HoldingInfo — карточка активного HOLDING на планете игрока для UI
// (план 15 этап 4).
type HoldingInfo struct {
	EventID     string         `json:"event_id"`
	PlanetID    string         `json:"planet_id"`
	PlanetName  string         `json:"planet_name"`
	Galaxy      int            `json:"galaxy"`
	System      int            `json:"system"`
	Position    int            `json:"position"`
	Tier        int            `json:"tier"`
	StartTime   time.Time      `json:"start_time"`
	EndsAt      time.Time      `json:"ends_at"`         // fire_at события HOLDING
	PaidCredit  int64          `json:"paid_credit"`
	PaidTimes   int            `json:"paid_times"`
	MaxEndsAt   time.Time      `json:"max_ends_at"`     // start + 15 дней (cap)
	AlienFleet  []FleetStack   `json:"alien_fleet"`
}

// FleetStack — стек alien-флота для UI.
type FleetStack struct {
	UnitID   int   `json:"unit_id"`
	Quantity int64 `json:"quantity"`
}

// ListMyHoldings — все активные HOLDING на планетах игрока. Используется
// FleetScreen / OverviewScreen / отдельным виджетом, чтобы игрок видел
// захват своей планеты пришельцами.
func ListMyHoldings(ctx context.Context, db repo.Exec, userID string) ([]HoldingInfo, error) {
	rows, err := db.Pool().Query(ctx, `
		SELECT e.id, e.payload, e.fire_at,
		       p.id, p.name, p.galaxy, p.system, p.position
		FROM events e
		JOIN planets p ON p.id = e.planet_id
		WHERE e.kind = $1 AND e.state = 'wait'
		  AND p.user_id = $2 AND p.destroyed_at IS NULL
		ORDER BY e.fire_at ASC
	`, int(event.KindAlienHolding), userID)
	if err != nil {
		return nil, fmt.Errorf("alien list holdings: %w", err)
	}
	defer rows.Close()

	var out []HoldingInfo
	for rows.Next() {
		var (
			eventID                 string
			payload                 []byte
			fireAt                  time.Time
			planetID, planetName    string
			g, sys, pos             int
		)
		if err := rows.Scan(&eventID, &payload, &fireAt,
			&planetID, &planetName, &g, &sys, &pos); err != nil {
			return nil, err
		}
		var hp holdingPayload
		if err := json.Unmarshal(payload, &hp); err != nil {
			continue
		}
		fleet := make([]FleetStack, 0, len(hp.AlienFleet))
		for _, fs := range hp.AlienFleet {
			fleet = append(fleet, FleetStack{UnitID: fs.UnitID, Quantity: fs.Quantity})
		}
		out = append(out, HoldingInfo{
			EventID:    eventID,
			PlanetID:   planetID,
			PlanetName: planetName,
			Galaxy:     g,
			System:     sys,
			Position:   pos,
			Tier:       hp.Tier,
			StartTime:  hp.StartTime,
			EndsAt:     fireAt,
			PaidCredit: hp.PaidCredit,
			PaidTimes:  hp.PaidTimes,
			MaxEndsAt:  hp.StartTime.Add(15 * 24 * time.Hour),
			AlienFleet: fleet,
		})
	}
	return out, rows.Err()
}
