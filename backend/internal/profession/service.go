// Package profession управляет профессией игрока.
//
// Профессия даёт виртуальные бонусы/штрафы к уровням зданий и исследований.
// Смена профессии: 1000 кр, мин. интервал 14 дней. Значение "none" означает
// отсутствие профессии (нет ни бонусов, ни штрафов).
package profession

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/config"
	"github.com/oxsar/nova/backend/internal/repo"
)

const (
	ChangeCost     = int64(1000) // кредитов за смену профессии
	ChangeInterval = 14 * 24 * time.Hour
	NoProfession   = "none"
)

var (
	ErrUnknownProfession = errors.New("profession: unknown profession key")
	ErrNotEnoughCredit   = errors.New("profession: not enough credit")
	ErrChangeTooSoon     = errors.New("profession: cannot change profession yet (14 day cooldown)")
)

type Service struct {
	db      repo.Exec
	catalog *config.Catalog
}

func NewService(db repo.Exec, cat *config.Catalog) *Service {
	return &Service{db: db, catalog: cat}
}

// CurrentInfo — текущая профессия и когда следующая смена будет доступна.
type CurrentInfo struct {
	Profession        string     `json:"profession"`
	Label             string     `json:"label"`
	NextChangeAllowed *time.Time `json:"next_change_allowed,omitempty"`
}

// List возвращает список всех профессий с их бонусами.
func (s *Service) List() []ProfessionDTO {
	out := make([]ProfessionDTO, 0, len(s.catalog.Professions.Professions))
	for key, spec := range s.catalog.Professions.Professions {
		out = append(out, ProfessionDTO{
			Key:   key,
			Label: spec.Label,
			Bonus: spec.Bonus,
			Malus: spec.Malus,
		})
	}
	return out
}

// Get возвращает текущую профессию пользователя.
func (s *Service) Get(ctx context.Context, userID string) (CurrentInfo, error) {
	var profession string
	var changedAt *time.Time
	err := s.db.Pool().QueryRow(ctx,
		`SELECT profession, profession_changed_at FROM users WHERE id=$1`, userID,
	).Scan(&profession, &changedAt)
	if err != nil {
		return CurrentInfo{}, err
	}

	info := CurrentInfo{Profession: profession}
	if profession != NoProfession {
		if spec, ok := s.catalog.Professions.Professions[profession]; ok {
			info.Label = spec.Label
		}
	}
	if changedAt != nil {
		next := changedAt.Add(ChangeInterval)
		info.NextChangeAllowed = &next
	}
	return info, nil
}

// Change меняет профессию пользователя. Списывает 1000 кр, проверяет
// интервал 14 дней, валидирует ключ профессии.
func (s *Service) Change(ctx context.Context, userID, professionKey string) error {
	if professionKey != NoProfession {
		if _, ok := s.catalog.Professions.Professions[professionKey]; !ok {
			return ErrUnknownProfession
		}
	}

	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var credit float64
		var changedAt *time.Time
		if err := tx.QueryRow(ctx,
			`SELECT credit, profession_changed_at FROM users WHERE id=$1 FOR UPDATE`,
			userID,
		).Scan(&credit, &changedAt); err != nil {
			return err
		}

		if changedAt != nil && time.Since(*changedAt) < ChangeInterval {
			return ErrChangeTooSoon
		}

		if credit < float64(ChangeCost) {
			return ErrNotEnoughCredit
		}

		now := time.Now().UTC()
		_, err := tx.Exec(ctx,
			`UPDATE users SET profession=$1, profession_changed_at=$2, credit=credit-$3 WHERE id=$4`,
			professionKey, now, ChangeCost, userID,
		)
		return err
	})
}

// BonusForUser возвращает карту смещений уровней для данного пользователя.
// Ключи: те же, что в buildings.yml и research.yml, плюс "gun", "shield_weapon",
// "shell_weapon", "ballistics", "masking".
// Возвращает nil если профессия не задана или "none".
func (s *Service) BonusForUser(ctx context.Context, userID string) (map[string]int, error) {
	var profession string
	if err := s.db.Pool().QueryRow(ctx,
		`SELECT profession FROM users WHERE id=$1`, userID,
	).Scan(&profession); err != nil {
		return nil, err
	}
	return BonusFromKey(s.catalog, profession), nil
}

// BonusFromKey вычисляет суммарные смещения (bonus+malus) для данного ключа профессии.
func BonusFromKey(cat *config.Catalog, professionKey string) map[string]int {
	if professionKey == NoProfession || professionKey == "" {
		return nil
	}
	spec, ok := cat.Professions.Professions[professionKey]
	if !ok {
		return nil
	}
	out := make(map[string]int, len(spec.Bonus)+len(spec.Malus))
	for k, v := range spec.Bonus {
		out[k] += v
	}
	for k, v := range spec.Malus {
		out[k] += v
	}
	return out
}

type ProfessionDTO struct {
	Key   string         `json:"key"`
	Label string         `json:"label"`
	Bonus map[string]int `json:"bonus,omitempty"`
	Malus map[string]int `json:"malus,omitempty"`
}
