// Package score — пересчёт и хранение очков игрока.
//
// Формулы идентичны legacy PointRenewer.class.php (oxsar2):
//
//	b_points = 0.0005 × Σ cost(building_level)  (по всем планетам)
//	r_points = 0.001  × Σ cost(research_level)
//	u_points = 0.002  × Σ count × cost(unit)    (ships + defense на всех планетах)
//	a_points = Σ achievement.points              (начислены при разблокировке)
//	points   = b_points + r_points + u_points
//
// RecalcUser вызывается в тех же транзакциях, что и мутации зданий/
// исследований/кораблей, или из экономического воркера — не нужна
// отдельная фоновая задача.
//
// Top возвращает N лучших для лидерборда (GET /api/highscore).
package score

import (
	"context"
	"fmt"
	"math"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/config"
	"github.com/oxsar/nova/backend/internal/repo"
)

// Service — пересчёт очков.
type Service struct {
	db   repo.Exec
	cat  *config.Catalog
	kBld float64 // коэффициент за постройки
	kRes float64 // коэффициент за исследования
	kUnt float64 // коэффициент за флот/оборону
}

// NewService создаёт Service с коэффициентами из конфига.
func NewService(db repo.Exec, cat *config.Catalog) *Service {
	return NewServiceWithCoeffs(db, cat, config.PointsCoefficients{
		Building: 0.00005,
		Research: 0.0005,
		Unit:     0.002,
	})
}

// NewServiceWithCoeffs создаёт Service с явными коэффициентами из конфига.
func NewServiceWithCoeffs(db repo.Exec, cat *config.Catalog, k config.PointsCoefficients) *Service {
	return &Service{db: db, cat: cat, kBld: k.Building, kRes: k.Research, kUnt: k.Unit}
}

// Entry — строка лидерборда.
type Entry struct {
	Rank        int     `json:"rank"`
	UserID      string  `json:"user_id"`
	Username    string  `json:"username"`
	AllianceTag *string `json:"alliance_tag,omitempty"`
	Points      float64 `json:"points"`
	BPoints     float64 `json:"b_points"`
	RPoints     float64 `json:"r_points"`
	UPoints     float64 `json:"u_points"`
	APoints     float64 `json:"a_points"`
	EPoints     float64 `json:"e_points"`
	// Координаты главной (первой созданной) планеты игрока — для клика из рейтинга.
	HomeGalaxy   *int `json:"home_galaxy,omitempty"`
	HomeSystem   *int `json:"home_system,omitempty"`
	HomePosition *int `json:"home_position,omitempty"`
}

// RecalcUser пересчитывает все компоненты очков userID и атомарно
// обновляет users. Если tx != nil — использует его (нет своей транзакции);
// иначе открывает новую.
func (s *Service) RecalcUser(ctx context.Context, userID string) error {
	bPoints, err := s.calcBuildings(ctx, userID)
	if err != nil {
		return fmt.Errorf("score.buildings: %w", err)
	}
	rPoints, err := s.calcResearch(ctx, userID)
	if err != nil {
		return fmt.Errorf("score.research: %w", err)
	}
	uPoints, err := s.calcUnits(ctx, userID)
	if err != nil {
		return fmt.Errorf("score.units: %w", err)
	}
	aPoints, err := s.calcAchievements(ctx, userID)
	if err != nil {
		return fmt.Errorf("score.achievements: %w", err)
	}
	total := bPoints + rPoints + uPoints
	_, err = s.db.Pool().Exec(ctx, `
		UPDATE users
		SET b_points=$2, r_points=$3, u_points=$4, a_points=$5, points=$6
		WHERE id=$1
	`, userID, roundPts(bPoints), roundPts(rPoints),
		roundPts(uPoints), roundPts(aPoints), roundPts(total))
	if err != nil {
		return fmt.Errorf("score.update: %w", err)
	}
	return nil
}

// Top возвращает топ-N для лидерборда. scoreType: "total"|"b"|"r"|"u"|"a".
// Неизвестный тип трактуется как "total".
func (s *Service) Top(ctx context.Context, scoreType string, limit int) ([]Entry, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	col := columnFor(scoreType)
	rows, err := s.db.Pool().Query(ctx, fmt.Sprintf(`
		SELECT u.id, u.username, a.tag,
		       u.points, u.b_points, u.r_points, u.u_points, u.a_points, u.e_points,
		       hp.galaxy, hp.system, hp.position
		FROM users u
		LEFT JOIN alliances a ON a.id = u.alliance_id
		LEFT JOIN LATERAL (
			SELECT galaxy, system, position FROM planets
			WHERE user_id = u.id AND destroyed_at IS NULL AND is_moon = false
			ORDER BY created_at ASC LIMIT 1
		) hp ON true
		WHERE u.umode = false
		ORDER BY u.%s DESC
		LIMIT $1
	`, col), limit)
	if err != nil {
		return nil, fmt.Errorf("score.top: %w", err)
	}
	defer rows.Close()

	var out []Entry
	rank := 1
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.UserID, &e.Username, &e.AllianceTag,
			&e.Points, &e.BPoints, &e.RPoints, &e.UPoints, &e.APoints, &e.EPoints,
			&e.HomeGalaxy, &e.HomeSystem, &e.HomePosition); err != nil {
			return nil, fmt.Errorf("score.scan: %w", err)
		}
		e.Rank = rank
		rank++
		out = append(out, e)
	}
	return out, rows.Err()
}

// AllianceEntry — строка рейтинга альянсов.
type AllianceEntry struct {
	Rank   int     `json:"rank"`
	Tag    string  `json:"tag"`
	Name   string  `json:"name"`
	Points float64 `json:"points"`
	Count  int     `json:"count"`
}

// VacationEntry — запись игрока в режиме отпуска.
type VacationEntry struct {
	Rank          int     `json:"rank"`
	UserID        string  `json:"user_id"`
	Username      string  `json:"username"`
	AllianceTag   *string `json:"alliance_tag,omitempty"`
	Points        float64 `json:"points"`
	VacationSince string  `json:"vacation_since"`
}

// TopAlliances возвращает топ альянсов по суммарным очкам членов.
func (s *Service) TopAlliances(ctx context.Context, limit int) ([]AllianceEntry, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	rows, err := s.db.Pool().Query(ctx, `
		SELECT a.tag, a.name, COUNT(u.id), SUM(u.points)
		FROM alliances a
		JOIN users u ON u.alliance_id = a.id AND u.umode = false
		GROUP BY a.id, a.tag, a.name
		ORDER BY SUM(u.points) DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("score.top_alliances: %w", err)
	}
	defer rows.Close()

	var out []AllianceEntry
	rank := 1
	for rows.Next() {
		var e AllianceEntry
		if err := rows.Scan(&e.Tag, &e.Name, &e.Count, &e.Points); err != nil {
			return nil, fmt.Errorf("score.alliance_scan: %w", err)
		}
		e.Rank = rank
		rank++
		out = append(out, e)
	}
	return out, rows.Err()
}

// VacationPlayers возвращает список игроков в режиме отпуска.
func (s *Service) VacationPlayers(ctx context.Context) ([]VacationEntry, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT u.id, u.username, a.tag, u.points,
		       ROW_NUMBER() OVER (ORDER BY u.points DESC) AS rank,
		       u.vacation_since
		FROM users u
		LEFT JOIN alliances a ON a.id = u.alliance_id
		WHERE u.vacation_since IS NOT NULL
		ORDER BY u.points DESC
		LIMIT 200
	`)
	if err != nil {
		return nil, fmt.Errorf("score.vacation_players: %w", err)
	}
	defer rows.Close()

	var out []VacationEntry
	for rows.Next() {
		var e VacationEntry
		var vs string
		if err := rows.Scan(&e.UserID, &e.Username, &e.AllianceTag, &e.Points, &e.Rank, &vs); err != nil {
			return nil, fmt.Errorf("score.vacation_scan: %w", err)
		}
		e.VacationSince = vs
		out = append(out, e)
	}
	return out, rows.Err()
}

// PlayerRank возвращает позицию userID в рейтинге.
func (s *Service) PlayerRank(ctx context.Context, userID, scoreType string) (int, error) {
	col := columnFor(scoreType)
	var pts float64
	if err := s.db.Pool().QueryRow(ctx,
		fmt.Sprintf(`SELECT %s FROM users WHERE id=$1`, col), userID).
		Scan(&pts); err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("score.player_pts: %w", err)
	}
	var rank int
	if err := s.db.Pool().QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*)+1 FROM users WHERE %s > $1 AND umode=false
	`, col), pts).Scan(&rank); err != nil {
		return 0, fmt.Errorf("score.player_rank: %w", err)
	}
	return rank, nil
}

// PlayerScore возвращает очки userID по типу scoreType.
func (s *Service) PlayerScore(ctx context.Context, userID, scoreType string) (float64, error) {
	col := columnFor(scoreType)
	var pts float64
	if err := s.db.Pool().QueryRow(ctx,
		fmt.Sprintf(`SELECT %s FROM users WHERE id=$1`, col), userID).
		Scan(&pts); err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("score.player_score: %w", err)
	}
	return pts, nil
}

// --- внутренние вычисления ---

func (s *Service) calcBuildings(ctx context.Context, userID string) (float64, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT b.unit_id, b.level
		FROM buildings b
		JOIN planets p ON p.id = b.planet_id
		WHERE p.user_id = $1 AND p.destroyed_at IS NULL
	`, userID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	// Индекс по ID для быстрого поиска.
	byID := make(map[int]config.BuildingSpec, len(s.cat.Buildings.Buildings))
	for _, sp := range s.cat.Buildings.Buildings {
		byID[sp.ID] = sp
	}

	var total float64
	for rows.Next() {
		var unitID, level int
		if err := rows.Scan(&unitID, &level); err != nil {
			return 0, err
		}
		spec, ok := byID[unitID]
		if !ok {
			continue
		}
		total += s.kBld * sumGeomCost(spec.CostBase, spec.CostFactor, level)
	}
	return total, rows.Err()
}

// sumGeomCost — сумма cost_base * factor^(i-1) для i=1..level, O(1).
// Формула: если factor=1, то cost_base_sum * level;
// иначе cost_base_sum * (factor^level - 1) / (factor - 1).
func sumGeomCost(cb config.ResCost, factor float64, level int) float64 {
	if level <= 0 {
		return 0
	}
	base := float64(cb.Metal + cb.Silicon + cb.Hydrogen)
	if factor == 1.0 || factor <= 0 {
		return base * float64(level)
	}
	return base * (math.Pow(factor, float64(level)) - 1) / (factor - 1)
}

func (s *Service) calcResearch(ctx context.Context, userID string) (float64, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT unit_id, level FROM research WHERE user_id=$1
	`, userID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	byID := make(map[int]config.ResearchSpec, len(s.cat.Research.Research))
	for _, sp := range s.cat.Research.Research {
		byID[sp.ID] = sp
	}

	var total float64
	for rows.Next() {
		var unitID, level int
		if err := rows.Scan(&unitID, &level); err != nil {
			return 0, err
		}
		spec, ok := byID[unitID]
		if !ok {
			continue
		}
		total += s.kRes * sumGeomCost(spec.CostBase, spec.CostFactor, level)
	}
	return total, rows.Err()
}

func (s *Service) calcUnits(ctx context.Context, userID string) (float64, error) {
	shipByID := make(map[int]config.ShipSpec, len(s.cat.Ships.Ships))
	for _, sp := range s.cat.Ships.Ships {
		shipByID[sp.ID] = sp
	}
	defByID := make(map[int]config.DefenseSpec, len(s.cat.Defense.Defense))
	for _, sp := range s.cat.Defense.Defense {
		defByID[sp.ID] = sp
	}

	rows, err := s.db.Pool().Query(ctx, `
		SELECT sh.unit_id, sh.count
		FROM ships sh
		JOIN planets p ON p.id = sh.planet_id
		WHERE p.user_id=$1 AND p.destroyed_at IS NULL AND sh.count > 0
	`, userID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var total float64
	for rows.Next() {
		var unitID int
		var count int64
		if err := rows.Scan(&unitID, &count); err != nil {
			return 0, err
		}
		if sp, ok := shipByID[unitID]; ok {
			total += s.kUnt * float64(sp.Cost.Metal+sp.Cost.Silicon+sp.Cost.Hydrogen) * float64(count)
		}
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	rows2, err := s.db.Pool().Query(ctx, `
		SELECT d.unit_id, d.count
		FROM defense d
		JOIN planets p ON p.id = d.planet_id
		WHERE p.user_id=$1 AND p.destroyed_at IS NULL AND d.count > 0
	`, userID)
	if err != nil {
		return 0, err
	}
	defer rows2.Close()

	for rows2.Next() {
		var unitID int
		var count int64
		if err := rows2.Scan(&unitID, &count); err != nil {
			return 0, err
		}
		if sp, ok := defByID[unitID]; ok {
			total += s.kUnt * float64(sp.Cost.Metal+sp.Cost.Silicon+sp.Cost.Hydrogen) * float64(count)
		}
	}
	return total, rows2.Err()
}

func (s *Service) calcAchievements(ctx context.Context, userID string) (float64, error) {
	var sum float64
	err := s.db.Pool().QueryRow(ctx, `
		SELECT COALESCE(SUM(points),0) FROM achievements WHERE user_id=$1
	`, userID).Scan(&sum)
	if err != nil {
		return 0, err
	}
	return sum, nil
}

// RecalcAll пересчитывает очки всех активных игроков (umode=false).
//
// DEPRECATED с плана 09 Ф.5.2: использовать RecalcAllEvent() и
// KindScoreRecalcAll (ежедневный event-based пересчёт). Метод оставлен
// для /admin/score/recalc on-demand и тестов.
// Ошибки отдельных игроков логируются, но не останавливают цикл.
func (s *Service) RecalcAll(ctx context.Context, log interface {
	WarnContext(context.Context, string, ...any)
}) error {
	rows, err := s.db.Pool().Query(ctx, `SELECT id FROM users WHERE umode=false`)
	if err != nil {
		return fmt.Errorf("score.recalc_all: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("score.recalc_all scan: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("score.recalc_all rows: %w", err)
	}

	for _, id := range ids {
		if err := s.RecalcUser(ctx, id); err != nil {
			log.WarnContext(ctx, "score_recalc_user_failed",
				"user_id", id, "err", err.Error())
		}
	}
	return nil
}

func columnFor(t string) string {
	switch t {
	case "b":
		return "b_points"
	case "r":
		return "r_points"
	case "u":
		return "u_points"
	case "a":
		return "a_points"
	case "e":
		return "e_points"
	default:
		return "points"
	}
}

func roundPts(v float64) float64 {
	return math.Round(v*100) / 100
}
