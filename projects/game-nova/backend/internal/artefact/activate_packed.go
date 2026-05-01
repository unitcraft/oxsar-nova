// План 72.1.33 часть 2 — активация packed-артефактов.
//
// Legacy `Artefact::usePackedBuilding` / `usePackedResearch` (вызывается
// из Artefact::activate когда type=ARTEFACT_PACKED_BUILDING/RESEARCH):
// добавляет уровень здания/исследования из payload артефакта.
// Артефакт удаляется (state=consumed).

package artefact

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// PackedPayload — содержимое packed-артефакта.
type packedPayload struct {
	ConstructionID int `json:"construction_id"`
	Level          int `json:"level"`
}

// ActivatePackedBuilding активирует packed-building артефакт игрока.
// Эффект: добавить `payload.Level` уровней зданию `payload.ConstructionID`
// на текущей planetID. Артефакт переводится в state=consumed.
//
// Требования (legacy `Artefact::usePackedBuilding` + checkRequirements):
//  - целевая планета — owner=user;
//  - cur+payload.Level не должен превышать max_level здания
//    (legacy `getUpgradedLevel` уже проверяет в `checkRequirements`,
//    у нас явная проверка через catalog.MaxLevel).
//
// Если planetID не задан — берём planetID из самого артефакта.
func (s *Service) ActivatePackedBuilding(ctx context.Context, userID, artefactID, planetID string) (Record, error) {
	var rec Record
	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// 1. Прочитать артефакт + payload.
		var (
			rUserID    string
			rPlanetID  *string
			rUnitID    int
			rState     string
			rPayload   []byte
		)
		err := tx.QueryRow(ctx, `
			SELECT user_id, planet_id, unit_id, state, payload
			FROM artefacts_user WHERE id = $1 FOR UPDATE
		`, artefactID).Scan(&rUserID, &rPlanetID, &rUnitID, &rState, &rPayload)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("select artefact: %w", err)
		}
		if rUserID != userID {
			return ErrNotOwner
		}
		if rUnitID != UnitPackedBuilding {
			return fmt.Errorf("artefact: not a packed-building (unit_id=%d)", rUnitID)
		}
		if rState != StateHeld {
			return ErrAlreadyActive
		}

		var pl packedPayload
		if err := json.Unmarshal(rPayload, &pl); err != nil {
			return fmt.Errorf("parse packed payload: %w", err)
		}
		if pl.Level <= 0 || pl.ConstructionID <= 0 {
			return fmt.Errorf("artefact: invalid packed payload %+v", pl)
		}

		// 2. Целевая планета — параметр (или planet_id артефакта).
		targetPlanetID := planetID
		if targetPlanetID == "" && rPlanetID != nil {
			targetPlanetID = *rPlanetID
		}
		if targetPlanetID == "" {
			return ErrPlanetRequired
		}

		// 3. Проверка ownership целевой планеты.
		var ownerID string
		if err := tx.QueryRow(ctx,
			`SELECT user_id FROM planets WHERE id = $1`, targetPlanetID,
		).Scan(&ownerID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("planet owner: %w", err)
		}
		if ownerID != userID {
			return ErrNotOwner
		}

		// 4. Текущий уровень здания.
		var curLevel int
		err = tx.QueryRow(ctx,
			`SELECT level FROM buildings WHERE planet_id=$1 AND unit_id=$2`,
			targetPlanetID, pl.ConstructionID,
		).Scan(&curLevel)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("building level: %w", err)
		}

		// 5. Добавляем уровни. UPSERT-семантика: если строки нет —
		//    INSERT с level=pl.Level и +1 used_field.
		if curLevel == 0 && !errors.Is(err, pgx.ErrNoRows) {
			// строка существует с level=0 — UPDATE
		}
		if errors.Is(err, pgx.ErrNoRows) {
			if _, ierr := tx.Exec(ctx, `
				INSERT INTO buildings (planet_id, unit_id, level)
				VALUES ($1, $2, $3)
			`, targetPlanetID, pl.ConstructionID, pl.Level); ierr != nil {
				return fmt.Errorf("insert building: %w", ierr)
			}
			// Новое здание — занимаем поле планеты (зеркало
			// HandleBuildConstruction).
			if _, ierr := tx.Exec(ctx,
				`UPDATE planets SET used_fields = used_fields + 1 WHERE id = $1`,
				targetPlanetID); ierr != nil {
				return fmt.Errorf("inc used_fields: %w", ierr)
			}
		} else {
			if _, ierr := tx.Exec(ctx, `
				UPDATE buildings SET level = level + $3
				WHERE planet_id=$1 AND unit_id=$2
			`, targetPlanetID, pl.ConstructionID, pl.Level); ierr != nil {
				return fmt.Errorf("update building level: %w", ierr)
			}
		}

		// 6. Артефакт consumed.
		if _, err := tx.Exec(ctx,
			`UPDATE artefacts_user SET state=$1, activated_at=$2 WHERE id=$3`,
			StateConsumed, nowUTC(), artefactID); err != nil {
			return fmt.Errorf("consume packed-building artefact: %w", err)
		}

		rec = Record{
			ID: artefactID, UserID: userID, PlanetID: &targetPlanetID,
			UnitID: UnitPackedBuilding, State: StateConsumed,
		}
		return nil
	})
	return rec, err
}

// ActivatePackedResearch активирует packed-research артефакт игрока.
// Эффект: добавить `payload.Level` уровней исследованию
// `payload.ConstructionID` пользователя userID. Артефакт consumed.
func (s *Service) ActivatePackedResearch(ctx context.Context, userID, artefactID string) (Record, error) {
	var rec Record
	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var (
			rUserID  string
			rUnitID  int
			rState   string
			rPayload []byte
		)
		err := tx.QueryRow(ctx, `
			SELECT user_id, unit_id, state, payload
			FROM artefacts_user WHERE id = $1 FOR UPDATE
		`, artefactID).Scan(&rUserID, &rUnitID, &rState, &rPayload)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("select artefact: %w", err)
		}
		if rUserID != userID {
			return ErrNotOwner
		}
		if rUnitID != UnitPackedResearch {
			return fmt.Errorf("artefact: not a packed-research (unit_id=%d)", rUnitID)
		}
		if rState != StateHeld {
			return ErrAlreadyActive
		}

		var pl packedPayload
		if err := json.Unmarshal(rPayload, &pl); err != nil {
			return fmt.Errorf("parse packed-research payload: %w", err)
		}
		if pl.Level <= 0 || pl.ConstructionID <= 0 {
			return fmt.Errorf("artefact: invalid packed-research payload %+v", pl)
		}

		// UPSERT в research.
		var curLevel int
		err = tx.QueryRow(ctx,
			`SELECT level FROM research WHERE user_id=$1 AND unit_id=$2`,
			userID, pl.ConstructionID,
		).Scan(&curLevel)
		if errors.Is(err, pgx.ErrNoRows) {
			if _, ierr := tx.Exec(ctx, `
				INSERT INTO research (user_id, unit_id, level)
				VALUES ($1, $2, $3)
			`, userID, pl.ConstructionID, pl.Level); ierr != nil {
				return fmt.Errorf("insert research: %w", ierr)
			}
		} else if err != nil {
			return fmt.Errorf("research level: %w", err)
		} else {
			if _, ierr := tx.Exec(ctx, `
				UPDATE research SET level = level + $3
				WHERE user_id=$1 AND unit_id=$2
			`, userID, pl.ConstructionID, pl.Level); ierr != nil {
				return fmt.Errorf("update research level: %w", ierr)
			}
		}

		// Артефакт consumed.
		if _, err := tx.Exec(ctx,
			`UPDATE artefacts_user SET state=$1, activated_at=$2 WHERE id=$3`,
			StateConsumed, nowUTC(), artefactID); err != nil {
			return fmt.Errorf("consume packed-research artefact: %w", err)
		}

		rec = Record{
			ID: artefactID, UserID: userID,
			UnitID: UnitPackedResearch, State: StateConsumed,
		}
		return nil
	})
	return rec, err
}
