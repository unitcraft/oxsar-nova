// Package requirements проверяет, выполнены ли предусловия для
// постройки/исследования/корабля (§5.1 ТЗ, REQ_BUILDING, REQ_RESEARCH
// из oxsar2/consts.php).
//
// Реализовано как отдельный пакет, чтобы research/shipyard/defense не
// дублировали логику.
package requirements

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/config"
)

// ErrNotMet — требования не выполнены. Оборачивает конкретный unmet-
// requirement, чтобы handler мог вернуть осмысленный текст.
type ErrNotMet struct {
	Kind  string // building | research
	Key   string
	Need  int
	Have  int
}

func (e *ErrNotMet) Error() string {
	return fmt.Sprintf("requirement not met: %s %s level %d (have %d)", e.Kind, e.Key, e.Need, e.Have)
}

// IsNotMet helper для handler'ов.
func IsNotMet(err error) bool {
	var e *ErrNotMet
	return errors.As(err, &e)
}

// Checker держит ссылку на каталог и умеет проверять требования.
// Запросы к БД идут через переданный tx — это важно: предусловия
// проверяются в той же транзакции, что и списание ресурсов / постановка
// в очередь, иначе race между «есть уровень 5» и «уровень 5 уже запрошен».
type Checker struct {
	cat *config.Catalog
}

func New(cat *config.Catalog) *Checker { return &Checker{cat: cat} }

// Check проверяет все требования для targetKey.
// userID — владелец исследований; planetID — для building-требований.
// Если для targetKey требований нет в конфиге — считается, что нет
// предусловий (return nil). Это сознательно: новый юнит без
// requirements работает с первого дня без правок кода.
func (c *Checker) Check(ctx context.Context, tx pgx.Tx, targetKey, userID, planetID string) error {
	reqs, ok := c.cat.Requirements.Requirements[targetKey]
	if !ok {
		return nil
	}
	for _, r := range reqs {
		switch r.Kind {
		case "building":
			have, err := buildingLevel(ctx, tx, planetID, c.lookupBuildingID(r.Key))
			if err != nil {
				return err
			}
			if have < r.Level {
				return &ErrNotMet{Kind: "building", Key: r.Key, Need: r.Level, Have: have}
			}
		case "research":
			have, err := researchLevel(ctx, tx, userID, c.lookupResearchID(r.Key))
			if err != nil {
				return err
			}
			if have < r.Level {
				return &ErrNotMet{Kind: "research", Key: r.Key, Need: r.Level, Have: have}
			}
		default:
			return fmt.Errorf("requirements: unknown kind %q for target %q", r.Kind, targetKey)
		}
	}
	return nil
}

// UnmetItem описывает одно невыполненное требование для отображения в UI.
type UnmetItem struct {
	Kind     string `json:"kind"`
	Key      string `json:"key"`
	Required int    `json:"required"`
	Current  int    `json:"current"`
}

// UnmetForTarget возвращает список невыполненных требований для targetKey.
// Если требований нет или все выполнены — возвращает nil.
// Использует пул (не транзакцию), т.к. только для чтения и не в составе мутации.
func (c *Checker) UnmetForTarget(ctx context.Context, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, targetKey, userID, planetID string) ([]UnmetItem, error) {
	reqs, ok := c.cat.Requirements.Requirements[targetKey]
	if !ok {
		return nil, nil
	}
	var out []UnmetItem
	for _, r := range reqs {
		switch r.Kind {
		case "building":
			var lvl int
			err := db.QueryRow(ctx,
				`SELECT COALESCE(level, 0) FROM buildings WHERE planet_id = $1 AND unit_id = $2`,
				planetID, c.lookupBuildingID(r.Key),
			).Scan(&lvl)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("requirements: %w", err)
			}
			if lvl < r.Level {
				out = append(out, UnmetItem{Kind: "building", Key: r.Key, Required: r.Level, Current: lvl})
			}
		case "research":
			var lvl int
			err := db.QueryRow(ctx,
				`SELECT COALESCE(level, 0) FROM research WHERE user_id = $1 AND unit_id = $2`,
				userID, c.lookupResearchID(r.Key),
			).Scan(&lvl)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("requirements: %w", err)
			}
			if lvl < r.Level {
				out = append(out, UnmetItem{Kind: "research", Key: r.Key, Required: r.Level, Current: lvl})
			}
		}
	}
	return out, nil
}

func (c *Checker) lookupBuildingID(key string) int {
	if spec, ok := c.cat.Buildings.Buildings[key]; ok {
		return spec.ID
	}
	return 0
}

func (c *Checker) lookupResearchID(key string) int {
	if spec, ok := c.cat.Research.Research[key]; ok {
		return spec.ID
	}
	return 0
}

func buildingLevel(ctx context.Context, tx pgx.Tx, planetID string, unitID int) (int, error) {
	if unitID == 0 {
		return 0, nil
	}
	var lvl int
	err := tx.QueryRow(ctx,
		`SELECT level FROM buildings WHERE planet_id = $1 AND unit_id = $2`,
		planetID, unitID,
	).Scan(&lvl)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("building level: %w", err)
	}
	return lvl, nil
}

func researchLevel(ctx context.Context, tx pgx.Tx, userID string, unitID int) (int, error) {
	if unitID == 0 {
		return 0, nil
	}
	var lvl int
	err := tx.QueryRow(ctx,
		`SELECT level FROM research WHERE user_id = $1 AND unit_id = $2`,
		userID, unitID,
	).Scan(&lvl)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("research level: %w", err)
	}
	return lvl, nil
}
