// Package profession управляет профессией игрока.
//
// Профессия даёт виртуальные бонусы/штрафы к уровням зданий и исследований.
// Смена профессии: 1000 кр, мин. интервал 14 дней. Значение "none" означает
// отсутствие профессии (нет ни бонусов, ни штрафов).
//
// План 72.1.15: паритет с legacy `Profession.class.php`:
//   - umode-блок: нельзя менять в режиме отпуска (Logger::dieMessage('UMODE_ENABLED')).
//   - same-profession check: смена на ту же — no-op без списания.
//   - AutoMsg MSG_CREDIT_PROFESSION_CHANGED в папку MSG_FOLDER_CREDIT (8)
//     при успешном списании.
package profession

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/automsg"
	"oxsar/game-nova/internal/config"
	"oxsar/game-nova/internal/i18n"
	"oxsar/game-nova/internal/repo"
)

const (
	ChangeCost     = int64(1000) // кредитов за смену профессии
	ChangeInterval = 14 * 24 * time.Hour
	NoProfession   = "none"

	// MSG_FOLDER_CREDIT в legacy = 8 (config/consts.php:515).
	creditMessageFolder = 8
)

var (
	ErrUnknownProfession = errors.New("profession: unknown profession key")
	ErrNotEnoughCredit   = errors.New("profession: not enough credit")
	ErrChangeTooSoon     = errors.New("profession: cannot change profession yet (14 day cooldown)")
	ErrInVacation        = errors.New("profession: cannot change profession in vacation mode")
)

type Service struct {
	db      repo.Exec
	catalog *config.Catalog
	automsg *automsg.Service
	bundle  *i18n.Bundle
}

func NewService(db repo.Exec, cat *config.Catalog) *Service {
	return &Service{db: db, catalog: cat}
}

// WithAutoMsg подключает automsg-сервис для отправки уведомления о
// списании кредитов при смене профессии (legacy MSG_CREDIT_PROFESSION_CHANGED).
// Если не вызван — Change работает, но без AutoMsg (graceful degradation).
func (s *Service) WithAutoMsg(am *automsg.Service) *Service {
	s.automsg = am
	return s
}

// WithBundle подключает i18n-бандл для перевода текста AutoMsg на язык юзера.
func (s *Service) WithBundle(b *i18n.Bundle) *Service {
	s.bundle = b
	return s
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
			Key:         key,
			Label:       spec.Label,
			Description: spec.Description,
			Bonus:       spec.Bonus,
			Malus:       spec.Malus,
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
// интервал 14 дней, валидирует ключ профессии, блокирует смену в umode.
//
// План 72.1.15: 1:1 с legacy `Profession.class.php::changeProfession`:
//   - umode → ErrInVacation (legacy `Logger::dieMessage('UMODE_ENABLED')`).
//   - смена на ту же → no-op (legacy `if($profession != $id)`).
//   - после списания → AutoMsg MSG_CREDIT_PROFESSION_CHANGED.
func (s *Service) Change(ctx context.Context, userID, professionKey string) error {
	if professionKey != NoProfession {
		if _, ok := s.catalog.Professions.Professions[professionKey]; !ok {
			return ErrUnknownProfession
		}
	}

	var sentAutoMsg bool
	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var (
			umode      bool
			credit     float64
			currentKey string
			changedAt  *time.Time
		)
		if err := tx.QueryRow(ctx,
			`SELECT umode, credit, profession, profession_changed_at FROM users WHERE id=$1 FOR UPDATE`,
			userID,
		).Scan(&umode, &credit, &currentKey, &changedAt); err != nil {
			return err
		}

		// Legacy: if($umode){ Logger::dieMessage('UMODE_ENABLED'); }
		if umode {
			return ErrInVacation
		}

		// Legacy: if($profession != $id …) — same → no-op без списания.
		if currentKey == professionKey {
			return nil
		}

		if changedAt != nil && time.Since(*changedAt) < ChangeInterval {
			return ErrChangeTooSoon
		}

		if credit < float64(ChangeCost) {
			return ErrNotEnoughCredit
		}

		now := time.Now().UTC()
		if _, err := tx.Exec(ctx,
			`UPDATE users SET profession=$1, profession_changed_at=$2, credit=credit-$3 WHERE id=$4`,
			professionKey, now, ChangeCost, userID,
		); err != nil {
			return err
		}

		// Legacy MSG_CREDIT_PROFESSION_CHANGED — отправка в одной транзакции
		// чтобы списание и уведомление были атомарны.
		if s.automsg != nil && s.bundle != nil {
			lang := s.userLang(ctx, tx, userID)
			label := s.labelFor(professionKey, lang)
			vars := map[string]string{
				"credits":    fmt.Sprintf("%d", ChangeCost),
				"profession": label,
			}
			title := s.bundle.Tr(lang, "autoMessages", "creditProfessionChanged.title", vars)
			body := s.bundle.Tr(lang, "autoMessages", "creditProfessionChanged.body", vars)
			if err := s.automsg.SendDirect(ctx, tx, userID, creditMessageFolder, title, body); err != nil {
				return fmt.Errorf("profession.automsg: %w", err)
			}
			sentAutoMsg = true
		}
		return nil
	})
	_ = sentAutoMsg
	return err
}

// userLang читает язык пользователя из транзакции (чтобы прочитать в той же
// БД-видимости, в которой работает Change). Fallback ru при ошибке.
func (s *Service) userLang(ctx context.Context, tx pgx.Tx, userID string) i18n.Lang {
	var lang string
	_ = tx.QueryRow(ctx, `SELECT language FROM users WHERE id=$1`, userID).Scan(&lang)
	if lang == "" {
		return i18n.LangRu
	}
	return i18n.Lang(lang)
}

// labelFor возвращает локализованную метку профессии для AutoMsg.
// Для NoProfession — i18n-fallback из autoMessages.
func (s *Service) labelFor(key string, lang i18n.Lang) string {
	if key == NoProfession {
		return s.bundle.Tr(lang, "autoMessages", "creditProfessionChanged.noneLabel", nil)
	}
	if spec, ok := s.catalog.Professions.Professions[key]; ok {
		return spec.Label
	}
	return key
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
	Key         string         `json:"key"`
	Label       string         `json:"label"`
	Description string         `json:"description,omitempty"`
	Bonus       map[string]int `json:"bonus,omitempty"`
	Malus       map[string]int `json:"malus,omitempty"`
}
