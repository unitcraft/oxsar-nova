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
	"sort"
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
	// NoProfession — устаревшее значение в БД для юзеров, у которых
	// профессия ещё не выбрана. Семантически эквивалентно UniversalKey
	// (ни бонусов, ни штрафов). План 72.1.58: Get() возвращает
	// UniversalKey даже если в БД лежит `none` — для совместимости UI.
	NoProfession = "none"
	// UniversalKey — ключ опции «Универсал» из configs/professions.yml
	// (план 72.1.58, legacy `Профессия Универсал`).
	UniversalKey = "universal"

	// MSG_FOLDER_CREDIT в legacy = 8 (config/consts.php:515).
	creditMessageFolder = 8
)

var (
	ErrUnknownProfession = errors.New("profession: unknown profession key")
	ErrNotEnoughCredit   = errors.New("profession: not enough credit")
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
// План 72.1.47: добавлены ChangeCost (0 после cooldown, иначе ChangeCost=1000)
// и DaysRemain (legacy `getProfessionChangeDaysRemain()`) для UI.
// План 72.1.59: убрано Label — фронт переводит ключ через i18n
// (group `profession`, key `<profession>Label`).
type CurrentInfo struct {
	Profession        string     `json:"profession"`
	NextChangeAllowed *time.Time `json:"next_change_allowed,omitempty"`
	ChangeCost        int64      `json:"change_cost"`
	DaysRemain        int        `json:"days_remain"`
}

// List возвращает список всех профессий с их эффектами, отсортированный
// по SortOrder из YAML (legacy: Универсал → Шахтёр → Атакёр → Защитник
// → Танк). План 72.1.58.
//
// План 72.1.59: эффекты теперь ordered slice []EffectEntry (а не
// bonus/malus map'ы), порядок 1:1 с legacy
// `consts.php:$GLOBALS["PROFESSIONS"]` tech_special. Frontend
// рендерит as-is без сортировки.
func (s *Service) List() []ProfessionDTO {
	out := make([]ProfessionDTO, 0, len(s.catalog.Professions.Professions))
	for key, spec := range s.catalog.Professions.Professions {
		// Клонируем slice (immutable view).
		effects := make([]EffectDTO, 0, len(spec.Effects))
		for _, e := range spec.Effects {
			effects = append(effects, EffectDTO{Key: e.Key, Value: e.Value})
		}
		out = append(out, ProfessionDTO{
			Key:       key,
			SortOrder: spec.SortOrder,
			Effects:   effects,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].SortOrder != out[j].SortOrder {
			return out[i].SortOrder < out[j].SortOrder
		}
		return out[i].Key < out[j].Key
	})
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

	// План 72.1.58: legacy-значение БД `none` (юзер не выбрал)
	// семантически эквивалентно `universal`. Для UI мапим в universal,
	// чтобы radio-button «Универсал» был отмечен как активный.
	displayKey := profession
	if displayKey == NoProfession {
		displayKey = UniversalKey
	}
	info := CurrentInfo{Profession: displayKey}
	if changedAt != nil {
		next := changedAt.Add(ChangeInterval)
		info.NextChangeAllowed = &next
		// План 72.1.47: legacy `getProfessionChangeCost`. Если ещё в
		// cooldown — cost=1000 + days_remain > 0; иначе оба = 0.
		if time.Since(*changedAt) < ChangeInterval {
			info.ChangeCost = ChangeCost
			info.DaysRemain = int((ChangeInterval - time.Since(*changedAt)) / (24 * time.Hour))
			if info.DaysRemain < 1 {
				info.DaysRemain = 1
			}
		}
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

		// План 72.1.47: legacy `getProfessionChangeCost()` (NS.class.php:2170)
		// возвращает 0 если `now - prof_time >= MIN_DAYS`, иначе COST. Раньше
		// мы трактовали это как «нельзя менять до cooldown», но legacy
		// разрешает менять В ЛЮБОЕ ВРЕМЯ — просто внутри cooldown берёт 1000 cr.
		var effectiveCost int64 = 0
		if changedAt != nil && time.Since(*changedAt) < ChangeInterval {
			effectiveCost = ChangeCost
		}

		if credit < float64(effectiveCost) {
			return ErrNotEnoughCredit
		}

		now := time.Now().UTC()
		if _, err := tx.Exec(ctx,
			`UPDATE users SET profession=$1, profession_changed_at=$2, credit=credit-$3 WHERE id=$4`,
			professionKey, now, effectiveCost, userID,
		); err != nil {
			return err
		}

		// Legacy MSG_CREDIT_PROFESSION_CHANGED — отправка в одной транзакции
		// чтобы списание и уведомление были атомарны.
		// План 72.1.47: AutoMsg credit отправляется только когда списались
		// деньги (effectiveCost > 0). Если cooldown прошёл и cost=0 —
		// legacy не шлёт `MSG_CREDIT_*`.
		if effectiveCost > 0 && s.automsg != nil && s.bundle != nil {
			lang := s.userLang(ctx, tx, userID)
			label := s.labelFor(professionKey, lang)
			vars := map[string]string{
				"credits":    fmt.Sprintf("%d", effectiveCost),
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
// План 72.1.59: имя берётся из i18n group `profession` ключ
// `<key>Label` (раньше был spec.Label из YAML — игнорировал lang).
// Для NoProfession — i18n-fallback из autoMessages.
// Если bundle не подключён (graceful degradation) — возвращаем key.
func (s *Service) labelFor(key string, lang i18n.Lang) string {
	if s.bundle == nil {
		return key
	}
	if key == NoProfession {
		return s.bundle.Tr(lang, "autoMessages", "creditProfessionChanged.noneLabel", nil)
	}
	if _, ok := s.catalog.Professions.Professions[key]; ok {
		return s.bundle.Tr(lang, "profession", key+"Label", nil)
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

// BonusFromKey вычисляет суммарные смещения уровней для данного
// ключа профессии. План 72.1.59: spec теперь содержит ordered
// []EffectEntry — суммируем в map для использования в расчёте
// производства/боя (там порядок не важен, важна сумма).
func BonusFromKey(cat *config.Catalog, professionKey string) map[string]int {
	// План 72.1.58: universal эквивалентен NoProfession (нет эффектов).
	if professionKey == NoProfession || professionKey == UniversalKey || professionKey == "" {
		return nil
	}
	spec, ok := cat.Professions.Professions[professionKey]
	if !ok {
		return nil
	}
	out := make(map[string]int, len(spec.Effects))
	for _, e := range spec.Effects {
		out[e.Key] += e.Value
	}
	return out
}

// ProfessionDTO — балансовый ответ /api/professions. План 72.1.59:
// label/description отсутствуют — фронт переводит ключ через i18n
// (group `profession`, key `<key>Label` / `<key>Desc`).
// Effects — ordered list (порядок 1:1 с legacy
// `consts.php:$GLOBALS["PROFESSIONS"]` tech_special).
type ProfessionDTO struct {
	Key string `json:"key"`
	// SortOrder — порядок сортировки для UI (план 72.1.58).
	SortOrder int         `json:"sort_order,omitempty"`
	Effects   []EffectDTO `json:"effects"`
}

// EffectDTO — один эффект профессии в API-ответе.
type EffectDTO struct {
	Key   string `json:"key"`
	Value int    `json:"value"`
}
