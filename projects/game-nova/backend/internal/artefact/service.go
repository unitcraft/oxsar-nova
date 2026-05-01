package artefact

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/config"
	"oxsar/game-nova/internal/i18n"
	"oxsar/game-nova/internal/repo"
	"oxsar/game-nova/pkg/ids"
)

// Состояния артефакта (повторяют enum artefact_state из миграции).
const (
	StateHeld     = "held"
	StateDelayed  = "delayed"
	StateActive   = "active"
	StateExpired  = "expired"
	StateConsumed = "consumed"
)

// Ошибки доменного слоя.
var (
	ErrNotFound          = errors.New("artefact: not found")
	ErrNotOwner          = errors.New("artefact: not owned by user")
	ErrAlreadyActive     = errors.New("artefact: already active")
	ErrPlanetRequired    = errors.New("artefact: planet_id required for per-planet effect")
	ErrUnknownArtefact   = errors.New("artefact: unknown artefact id")
	ErrMaxStacksReached  = errors.New("artefact: max stacks already active")
)

// Record — одна запись в artefacts_user.
type Record struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	PlanetID    *string   `json:"planet_id,omitempty"`
	UnitID      int       `json:"unit_id"`
	State       string    `json:"state"`
	AcquiredAt  time.Time `json:"acquired_at"`
	ActivatedAt *time.Time `json:"activated_at,omitempty"`
	ExpireAt    *time.Time `json:"expire_at,omitempty"`
}

// AutoMsgSender — узкий интерфейс к automsg.SendDirect.
type AutoMsgSender interface {
	SendDirect(ctx context.Context, tx pgx.Tx, userID string, folder int, title, body string) error
}

// Service — доменный фасад над артефактами.
type Service struct {
	db      repo.Exec
	catalog *config.Catalog
	automsg AutoMsgSender
	bundle  *i18n.Bundle
}

func NewService(db repo.Exec, cat *config.Catalog) *Service {
	return &Service{db: db, catalog: cat}
}

// WithAutoMsg подключает сервис системных сообщений (опционально).
func (s *Service) WithAutoMsg(a AutoMsgSender) *Service {
	s.automsg = a
	return s
}

func (s *Service) WithBundle(b *i18n.Bundle) *Service {
	s.bundle = b
	return s
}

func (s *Service) tr(group, key string, vars map[string]string) string {
	if s.bundle == nil {
		return "[" + group + "." + key + "]"
	}
	return s.bundle.Tr(i18n.LangRu, group, key, vars)
}

// Grant кладёт артефакт в инвентарь пользователя.
// Используется админской выдачей и наградами экспедиций/боёв.
func (s *Service) Grant(ctx context.Context, userID string, unitID int, planetID *string) (Record, error) {
	spec, ok := s.lookupByID(unitID)
	if !ok {
		return Record{}, ErrUnknownArtefact
	}
	_ = spec

	rec := Record{
		ID: ids.New(), UserID: userID, PlanetID: planetID,
		UnitID: unitID, State: StateHeld, AcquiredAt: time.Now().UTC(),
	}
	_, err := s.db.Pool().Exec(ctx, `
		INSERT INTO artefacts_user (id, user_id, planet_id, unit_id, state, acquired_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, rec.ID, rec.UserID, rec.PlanetID, rec.UnitID, rec.State, rec.AcquiredAt)
	if err != nil {
		return Record{}, fmt.Errorf("insert artefact: %w", err)
	}
	return rec, nil
}

// Activate активирует артефакт игрока.
// Правила (§5.10.1, Artefact.class.php::activateArtefact):
//   1) Артефакт должен принадлежать userID и быть в state=held.
//   2) Для non-stackable: если у игрока есть активный той же спеки —
//      отказ (ErrNonStackable).
//   3) Для factor_planet нужен planet_id у записи.
//   4) Применяем FactorChange через applyChange(), обновляем state,
//      выставляем expire_at и вставляем event EVENT_ARTEFACT_EXPIRE.
//
// План 72.1.33 часть 2: для packed_building / packed_research артефактов
// делегирует в ActivatePackedBuilding / ActivatePackedResearch (special-case
// flow с обновлением building/research level из payload).
func (s *Service) Activate(ctx context.Context, userID, artefactID string) (Record, error) {
	// Спец-ветка для packed-артефактов — особая семантика активации.
	var preUnitID int
	_ = s.db.Pool().QueryRow(ctx,
		`SELECT unit_id FROM artefacts_user WHERE id = $1`, artefactID,
	).Scan(&preUnitID)
	switch preUnitID {
	case UnitPackedBuilding:
		// Целевая планета берётся из артефакта (planet_id), активация
		// на «другой планете» — отдельный endpoint /pack/.../activate
		// если потребуется.
		return s.ActivatePackedBuilding(ctx, userID, artefactID, "")
	case UnitPackedResearch:
		return s.ActivatePackedResearch(ctx, userID, artefactID)
	case UnitPackingBuilding, UnitPackingResearch:
		// Packing-артефакты не активируются напрямую — только через
		// PackBuilding/Research endpoint с указанием цели (building_id).
		return Record{}, fmt.Errorf("artefact: packing artefact must be used via /api/planets/.../buildings/.../pack endpoint")
	}

	var rec Record
	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		r, spec, err := s.loadOwned(ctx, tx, userID, artefactID)
		if err != nil {
			return err
		}
		if r.State != StateHeld && r.State != StateDelayed {
			return ErrAlreadyActive
		}

		if !spec.Stackable {
			var activeCount int
			if err := tx.QueryRow(ctx, `
				SELECT COUNT(*) FROM artefacts_user
				WHERE user_id = $1 AND unit_id = $2 AND state = $3
			`, userID, r.UnitID, StateActive).Scan(&activeCount); err != nil {
				return fmt.Errorf("check stackable: %w", err)
			}
			if activeCount > 0 {
				return ErrNonStackable
			}
		}

		if spec.MaxStacks > 0 {
			var activeCount int
			if err := tx.QueryRow(ctx, `
				SELECT COUNT(*) FROM artefacts_user
				WHERE user_id = $1 AND unit_id = $2 AND state = $3
			`, userID, r.UnitID, StateActive).Scan(&activeCount); err != nil {
				return fmt.Errorf("check max stacks: %w", err)
			}
			if activeCount >= spec.MaxStacks {
				return ErrMaxStacksReached
			}
		}

		// Если у артефакта есть delay — переводим в delayed и планируем
		// событие KindArtefactDelay (63). Сами эффекты применятся позже,
		// когда delay-событие сработает.
		if spec.DelaySeconds > 0 && r.State == StateHeld {
			now := time.Now().UTC()
			fireAt := now.Add(time.Duration(spec.DelaySeconds) * time.Second)
			if _, err := tx.Exec(ctx, `
				UPDATE artefacts_user SET state = $1, activated_at = $2 WHERE id = $3
			`, StateDelayed, now, r.ID); err != nil {
				return fmt.Errorf("set delayed: %w", err)
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO events (id, user_id, planet_id, kind, state, fire_at, payload)
				VALUES ($1, $2, $3, 63, 'wait', $4, $5)
			`, ids.New(), userID, r.PlanetID, fireAt,
				fmt.Sprintf(`{"artefact_id":"%s"}`, r.ID)); err != nil {
				return fmt.Errorf("insert delay event: %w", err)
			}
			r.State = StateDelayed
			r.ActivatedAt = &now
			rec = r
			return nil
		}

		change, err := computeChanges(spec, dirApply)
		if err != nil && !errors.Is(err, ErrUnsupported) {
			return err
		}
		if change != nil {
			if change.Scope == "planet" && r.PlanetID == nil {
				return ErrPlanetRequired
			}
			if err := applyChange(ctx, tx, *change, userID, r.PlanetID); err != nil {
				return err
			}
		}

		now := time.Now().UTC()
		var expire *time.Time
		if spec.LifetimeSeconds > 0 {
			t := now.Add(time.Duration(spec.LifetimeSeconds) * time.Second)
			expire = &t
		}
		if _, err := tx.Exec(ctx, `
			UPDATE artefacts_user SET state = $1, activated_at = $2, expire_at = $3
			WHERE id = $4
		`, StateActive, now, expire, r.ID); err != nil {
			return fmt.Errorf("update artefact state: %w", err)
		}

		if expire != nil {
			if _, err := tx.Exec(ctx, `
				INSERT INTO events (id, user_id, planet_id, kind, state, fire_at, payload)
				VALUES ($1, $2, $3, 60, 'wait', $4, $5)
			`, ids.New(), userID, r.PlanetID, *expire,
				fmt.Sprintf(`{"artefact_id":"%s"}`, r.ID)); err != nil {
				return fmt.Errorf("insert expire event: %w", err)
			}
		}

		r.State = StateActive
		r.ActivatedAt = &now
		r.ExpireAt = expire
		rec = r
		return nil
	})
	return rec, err
}

// Deactivate отменяет активный артефакт (revert-эффекта).
func (s *Service) Deactivate(ctx context.Context, userID, artefactID string) error {
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		r, spec, err := s.loadOwned(ctx, tx, userID, artefactID)
		if err != nil {
			return err
		}
		if r.State != StateActive {
			return nil // идемпотентность: уже не активен
		}
		change, err := computeChanges(spec, dirRevert)
		if err != nil && !errors.Is(err, ErrUnsupported) {
			return err
		}
		if change != nil {
			if err := applyChange(ctx, tx, *change, userID, r.PlanetID); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx,
			`UPDATE artefacts_user SET state = $1 WHERE id = $2`,
			StateExpired, r.ID); err != nil {
			return fmt.Errorf("update to expired: %w", err)
		}
		return nil
	})
}

// SlotsInfo — план 72.1.45: storage_slots/used_slots для UI шапки
// (legacy `Artefact::getStorageSlots`/`getUsedSlots`).
//
// storage_slots = research level UNIT_ARTEFACTS_TECH=111.
// used_slots    = COUNT(*) FROM artefacts_user WHERE state='active'.
//
// Возвращает (techLevel, usedSlots).
func (s *Service) SlotsInfo(ctx context.Context, userID string) (int, int, error) {
	const unitArtefactsTech = 111
	var techLevel int
	if err := s.db.Pool().QueryRow(ctx,
		`SELECT COALESCE(level, 0) FROM research WHERE user_id=$1 AND unit_id=$2`,
		userID, unitArtefactsTech,
	).Scan(&techLevel); err != nil {
		// no row OK.
		techLevel = 0
	}
	var usedSlots int
	if err := s.db.Pool().QueryRow(ctx,
		`SELECT COUNT(*) FROM artefacts_user WHERE user_id=$1 AND state='active'`,
		userID,
	).Scan(&usedSlots); err != nil {
		return 0, 0, fmt.Errorf("count active: %w", err)
	}
	return techLevel, usedSlots, nil
}

// ListUser возвращает инвентарь игрока.
func (s *Service) ListUser(ctx context.Context, userID string) ([]Record, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT id, user_id, planet_id, unit_id, state, acquired_at, activated_at, expire_at
		FROM artefacts_user
		WHERE user_id = $1 AND state <> 'consumed'
		ORDER BY acquired_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}
	defer rows.Close()

	var out []Record
	for rows.Next() {
		var r Record
		if err := rows.Scan(&r.ID, &r.UserID, &r.PlanetID, &r.UnitID,
			&r.State, &r.AcquiredAt, &r.ActivatedAt, &r.ExpireAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// loadOwned читает артефакт и проверяет владельца + наличие спеки.
func (s *Service) loadOwned(ctx context.Context, tx pgx.Tx, userID, id string) (Record, config.ArtefactSpec, error) {
	var r Record
	err := tx.QueryRow(ctx, `
		SELECT id, user_id, planet_id, unit_id, state, acquired_at, activated_at, expire_at
		FROM artefacts_user WHERE id = $1 FOR UPDATE
	`, id).Scan(&r.ID, &r.UserID, &r.PlanetID, &r.UnitID, &r.State,
		&r.AcquiredAt, &r.ActivatedAt, &r.ExpireAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Record{}, config.ArtefactSpec{}, ErrNotFound
		}
		return Record{}, config.ArtefactSpec{}, fmt.Errorf("select: %w", err)
	}
	if r.UserID != userID {
		return Record{}, config.ArtefactSpec{}, ErrNotOwner
	}
	spec, ok := s.lookupByID(r.UnitID)
	if !ok {
		return Record{}, config.ArtefactSpec{}, ErrUnknownArtefact
	}
	return r, spec, nil
}

func (s *Service) lookupByID(unitID int) (config.ArtefactSpec, bool) {
	for _, spec := range s.catalog.Artefacts.Artefacts {
		if spec.ID == unitID {
			return spec, true
		}
	}
	return config.ArtefactSpec{}, false
}

// applyChange исполняет один FactorChange в БД.
// Имена колонок НЕ санитизируются runtime — они белым списком
// провалидированы в effects.go::allowedField.
func applyChange(ctx context.Context, tx pgx.Tx, c FactorChange, userID string, planetID *string) error {
	switch c.Scope {
	case "user":
		var q string
		switch c.Op {
		case "set":
			q = fmt.Sprintf(`UPDATE users SET %s = $1 WHERE id = $2`, c.Field)
			_, err := tx.Exec(ctx, q, c.NewValue, userID)
			return err
		case "add":
			q = fmt.Sprintf(`UPDATE users SET %s = %s + $1 WHERE id = $2`, c.Field, c.Field)
			_, err := tx.Exec(ctx, q, c.Delta, userID)
			return err
		}
	case "planet":
		if planetID == nil {
			return ErrPlanetRequired
		}
		q := fmt.Sprintf(`UPDATE planets SET %s = %s + $1 WHERE id = $2`, c.Field, c.Field)
		_, err := tx.Exec(ctx, q, c.Delta, *planetID)
		return err
	case "all_planets":
		var q string
		switch c.Op {
		case "set":
			q = fmt.Sprintf(`UPDATE planets SET %s = $1 WHERE user_id = $2`, c.Field)
			_, err := tx.Exec(ctx, q, c.NewValue, userID)
			return err
		case "add":
			q = fmt.Sprintf(`UPDATE planets SET %s = %s + $1 WHERE user_id = $2`, c.Field, c.Field)
			_, err := tx.Exec(ctx, q, c.Delta, userID)
			return err
		}
	}
	return fmt.Errorf("artefact: unhandled scope/op %q/%q", c.Scope, c.Op)
}

// ActiveBattleModifiers возвращает итоговый боевой модификатор для userID
// из всех активных battle_bonus артефактов на момент вызова.
// Используется в attack.go для применения модификаторов атакующей стороне.
func (s *Service) ActiveBattleModifiers(ctx context.Context, tx pgx.Tx, userID string) (BattleModifier, error) {
	rows, err := tx.Query(ctx, `
		SELECT unit_id FROM artefacts_user
		WHERE user_id = $1 AND state = $2
	`, userID, StateActive)
	if err != nil {
		return BattleModifier{}, fmt.Errorf("query active artefacts: %w", err)
	}
	defer rows.Close()

	var specs []config.ArtefactSpec
	for rows.Next() {
		var unitID int
		if err := rows.Scan(&unitID); err != nil {
			return BattleModifier{}, err
		}
		spec, ok := s.lookupByID(unitID)
		if !ok {
			continue // неизвестный артефакт, пропускаем
		}
		if spec.Effect.Type == "battle_bonus" {
			specs = append(specs, spec)
		}
	}
	if err := rows.Err(); err != nil {
		return BattleModifier{}, err
	}
	return ComputeBattleModifier(specs), nil
}
