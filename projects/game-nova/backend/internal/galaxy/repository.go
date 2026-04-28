package galaxy

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SystemView — что видит игрок при открытии экрана галактики.
type SystemView struct {
	Galaxy int        `json:"galaxy"`
	System int        `json:"system"`
	Cells  []CellView `json:"cells"`
}

// CellView — одна клетка (позиция). Объединяет планету (если есть),
// признаки луны и обломков.
type CellView struct {
	Position      int     `json:"position"`
	PlanetName    *string `json:"planet_name,omitempty"`
	PlanetID      *string `json:"planet_id,omitempty"`
	PlanetType    *string `json:"planet_type,omitempty"`
	HasPlanet     bool    `json:"has_planet"`
	HasMoon       bool    `json:"has_moon"`
	MoonName      *string `json:"moon_name,omitempty"`
	MoonDiameter  *int    `json:"moon_diameter,omitempty"`
	MoonTempMin   *int    `json:"moon_temp_min,omitempty"`
	MoonTempMax   *int    `json:"moon_temp_max,omitempty"`
	OwnerUsername *string `json:"owner_username,omitempty"`
	OwnerID       *string `json:"owner_id,omitempty"`
	OwnerRank     *int    `json:"owner_rank,omitempty"`
	OwnerLastSeen *string `json:"owner_last_seen,omitempty"` // ISO-8601
	OwnerVacation bool    `json:"owner_vacation,omitempty"`
	OwnerBanned   bool    `json:"owner_banned,omitempty"`
	AllianceTag   *string `json:"alliance_tag,omitempty"`
	// Отношение альянса владельца клетки к альянсу viewer'а: nap | war | ally.
	// pending-отношения не показываются (не active).
	Relation      *string `json:"relation,omitempty"`
	// IsFriend = true, если владелец клетки в списке друзей viewer'а.
	IsFriend      bool    `json:"is_friend,omitempty"`
	DebrisMetal   int64   `json:"debris_metal"`
	DebrisSilicon int64   `json:"debris_silicon"`
}

// Repository — чтение галактической сетки.
type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// ReadSystem возвращает все 16 позиций указанной системы. viewerUserID —
// текущий пользователь, для вычисления отношения альянсов.
func (r *Repository) ReadSystem(ctx context.Context, galaxyNum, systemNum int, viewerUserID string) (SystemView, error) {
	out := SystemView{Galaxy: galaxyNum, System: systemNum}

	// Находим alliance_id viewer'а (если он в альянсе).
	var viewerAllianceID *string
	if viewerUserID != "" {
		_ = r.pool.QueryRow(ctx, `SELECT alliance_id FROM users WHERE id = $1`, viewerUserID).Scan(&viewerAllianceID)
	}

	type planetRow struct {
		Position      int
		IsMoon        bool
		ID            string
		Name          string
		Diameter      int
		TempMin       int
		TempMax       int
		PlanetType    *string
		OwnerID       *string
		OwnerName     *string
		OwnerRank     *int
		OwnerLastSeen *time.Time
		OwnerVacation bool
		OwnerBanned   bool
		AllianceTag   *string
		AllianceID    *string
	}
	planetRows := map[int]planetRow{}
	moonRows := map[int]planetRow{}

	rows, err := r.pool.Query(ctx, `
		SELECT
			p.position, p.is_moon, p.id, p.name, p.diameter, p.temperature_min, p.temperature_max,
			p.planet_type, p.user_id,
			u.username,
			CASE WHEN u.id IS NULL THEN NULL
			     ELSE (SELECT COUNT(*)+1 FROM users u2 WHERE u2.points > u.points AND u2.umode=false AND u2.is_observer=false)::int
			END AS rank,
			u.last_seen,
			COALESCE(u.umode, false) AS vacation,
			(u.banned_at IS NOT NULL) AS banned,
			al.tag AS alliance_tag,
			al.id  AS alliance_id
		FROM planets p
		LEFT JOIN users u ON u.id = p.user_id
		LEFT JOIN alliances al ON al.id = u.alliance_id
		WHERE p.galaxy = $1 AND p.system = $2 AND p.destroyed_at IS NULL
	`, galaxyNum, systemNum)
	if err != nil {
		return out, fmt.Errorf("read planets: %w", err)
	}
	for rows.Next() {
		var pr planetRow
		if err := rows.Scan(
			&pr.Position, &pr.IsMoon, &pr.ID, &pr.Name, &pr.Diameter, &pr.TempMin, &pr.TempMax,
			&pr.PlanetType, &pr.OwnerID,
			&pr.OwnerName, &pr.OwnerRank,
			&pr.OwnerLastSeen, &pr.OwnerVacation, &pr.OwnerBanned,
			&pr.AllianceTag, &pr.AllianceID,
		); err != nil {
			rows.Close()
			return out, fmt.Errorf("scan planet: %w", err)
		}
		if pr.IsMoon {
			moonRows[pr.Position] = pr
		} else {
			planetRows[pr.Position] = pr
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return out, fmt.Errorf("rows err: %w", err)
	}

	// Обломки.
	debris := map[int]struct{ m, s int64 }{}
	dRows, err := r.pool.Query(ctx, `
		SELECT position, debris_metal, debris_silicon
		FROM galaxy_cells
		WHERE galaxy = $1 AND system = $2
	`, galaxyNum, systemNum)
	if err != nil {
		return out, fmt.Errorf("read debris: %w", err)
	}
	for dRows.Next() {
		var pos int
		var m, s int64
		if err := dRows.Scan(&pos, &m, &s); err != nil {
			dRows.Close()
			return out, fmt.Errorf("scan debris: %w", err)
		}
		debris[pos] = struct{ m, s int64 }{m, s}
	}
	dRows.Close()
	if err := dRows.Err(); err != nil {
		return out, fmt.Errorf("debris rows: %w", err)
	}

	// Отношения альянсов: map[targetAllianceID]relation (active only).
	relations := map[string]string{}
	if viewerAllianceID != nil {
		rRows, err := r.pool.Query(ctx, `
			SELECT target_alliance_id, relation FROM alliance_relationships
			WHERE alliance_id = $1 AND status = 'active'
		`, *viewerAllianceID)
		if err == nil {
			for rRows.Next() {
				var tgt, rel string
				if err := rRows.Scan(&tgt, &rel); err == nil {
					relations[tgt] = rel
				}
			}
			rRows.Close()
		}
	}

	// Друзья viewer'а: set[friend_id].
	friends := map[string]struct{}{}
	if viewerUserID != "" {
		fRows, err := r.pool.Query(ctx, `SELECT friend_id FROM friends WHERE user_id = $1`, viewerUserID)
		if err == nil {
			for fRows.Next() {
				var fid string
				if err := fRows.Scan(&fid); err == nil {
					friends[fid] = struct{}{}
				}
			}
			fRows.Close()
		}
	}

	out.Cells = make([]CellView, 0, 16)
	for pos := 1; pos <= 16; pos++ {
		cell := CellView{Position: pos}
		if p, ok := planetRows[pos]; ok {
			name := p.Name
			cell.PlanetName = &name
			cell.PlanetID = &p.ID
			cell.PlanetType = p.PlanetType
			cell.HasPlanet = true
			cell.OwnerID = p.OwnerID
			cell.OwnerUsername = p.OwnerName
			cell.OwnerRank = p.OwnerRank
			cell.OwnerVacation = p.OwnerVacation
			cell.OwnerBanned = p.OwnerBanned
			cell.AllianceTag = p.AllianceTag
			if p.AllianceID != nil {
				if rel, ok := relations[*p.AllianceID]; ok {
					r := rel
					cell.Relation = &r
				}
			}
			if p.OwnerID != nil {
				if _, ok := friends[*p.OwnerID]; ok {
					cell.IsFriend = true
				}
			}
			if p.OwnerLastSeen != nil {
				s := p.OwnerLastSeen.UTC().Format(time.RFC3339)
				cell.OwnerLastSeen = &s
			}
		}
		if m, ok := moonRows[pos]; ok {
			name := m.Name
			cell.MoonName = &name
			cell.HasMoon = true
			cell.MoonDiameter = &m.Diameter
			cell.MoonTempMin = &m.TempMin
			cell.MoonTempMax = &m.TempMax
		}
		if d, ok := debris[pos]; ok {
			cell.DebrisMetal = d.m
			cell.DebrisSilicon = d.s
		}
		out.Cells = append(out.Cells, cell)
	}

	return out, nil
}
