package galaxy

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SystemView — что видит игрок при открытии экрана галактики.
// Pos 1..15; если клетка пуста, в PlanetName=nil, Moon=false, Owner=nil.
type SystemView struct {
	Galaxy  int        `json:"galaxy"`
	System  int        `json:"system"`
	Cells   []CellView `json:"cells"`
}

// CellView — одна клетка (позиция). Объединяет планету (если есть),
// признаки луны и обломков. Это «денормализованное» представление
// для UI — под капотом читается join'ом planets × galaxy_cells.
type CellView struct {
	Position      int     `json:"position"`
	PlanetName    *string `json:"planet_name,omitempty"`
	HasPlanet     bool    `json:"has_planet"`
	HasMoon       bool    `json:"has_moon"`
	MoonName      *string `json:"moon_name,omitempty"`
	OwnerUsername *string `json:"owner_username,omitempty"`
	OwnerID       *string `json:"owner_id,omitempty"`
	OwnerRank     *int    `json:"owner_rank,omitempty"`
	DebrisMetal   int64   `json:"debris_metal"`
	DebrisSilicon int64   `json:"debris_silicon"`
}

// Repository — чтение галактической сетки.
type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// ReadSystem возвращает все 15 позиций указанной системы.
//
// Читает planets (по is_moon=false и =true) + galaxy_cells в одном
// запросе, формирует пустые клетки для отсутствующих позиций. Если
// таблица galaxy_cells для этой системы пуста — это не ошибка, просто
// debris = 0.
//
// Координаты валидируются ВЫЗЫВАЮЩИМ (service-слой) — репо ожидает
// уже валидный вход.
func (r *Repository) ReadSystem(ctx context.Context, galaxyNum, systemNum int) (SystemView, error) {
	out := SystemView{Galaxy: galaxyNum, System: systemNum}

	// Собираем планеты и луны на позициях.
	type planetRow struct {
		Position  int
		IsMoon    bool
		Name      string
		OwnerID   *string
		OwnerName *string
		OwnerRank *int
	}
	planetRows := map[int]planetRow{} // планеты (is_moon=false)
	moonRows := map[int]planetRow{}   // луны на тех же координатах

	rows, err := r.pool.Query(ctx, `
		SELECT p.position, p.is_moon, p.name, p.user_id, u.username,
		       CASE WHEN u.id IS NULL THEN NULL
		            ELSE (SELECT COUNT(*)+1 FROM users u2 WHERE u2.points > u.points AND u2.umode=false)::int
		       END AS rank
		FROM planets p
		LEFT JOIN users u ON u.id = p.user_id
		WHERE p.galaxy = $1 AND p.system = $2 AND p.destroyed_at IS NULL
	`, galaxyNum, systemNum)
	if err != nil {
		return out, fmt.Errorf("read planets: %w", err)
	}
	for rows.Next() {
		var pr planetRow
		if err := rows.Scan(&pr.Position, &pr.IsMoon, &pr.Name, &pr.OwnerID, &pr.OwnerName, &pr.OwnerRank); err != nil {
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

	// Заполняем все 15 позиций (пусть пустых).
	out.Cells = make([]CellView, 0, 15)
	for pos := 1; pos <= 15; pos++ {
		cell := CellView{Position: pos}
		if p, ok := planetRows[pos]; ok {
			name := p.Name
			cell.PlanetName = &name
			cell.HasPlanet = true
			cell.OwnerID = p.OwnerID
			cell.OwnerUsername = p.OwnerName
			cell.OwnerRank = p.OwnerRank
		}
		if m, ok := moonRows[pos]; ok {
			name := m.Name
			cell.MoonName = &name
			cell.HasMoon = true
		}
		if d, ok := debris[pos]; ok {
			cell.DebrisMetal = d.m
			cell.DebrisSilicon = d.s
		}
		out.Cells = append(out.Cells, cell)
	}

	_ = time.Now // оставлено на будущее (cache TTL в редисе, когда
	// экран галактики начнёт читать миллион раз в минуту).
	return out, nil
}
