// План 72.1.33 часть 2 — упаковка здания/исследования в артефакт.
//
// Legacy `Artefact::canPackBuilding` + `BuildingInfo::packCurrentConstruction`:
// если у игрока на этой планете есть `held` packing-артефакт того же
// типа (323=building / 324=research), `pack`-операция:
//   1. Понижает уровень здания/исследования до `level - added` (минус
//      уровни, ранее «полученные» через активацию packed-артефактов).
//      Если added=level — ничего не понижается, артефакт consumed
//      без эффекта (legacy `if(level > 0)`).
//   2. Удаляет packing-артефакт (`state=consumed`).
//   3. Создаёт новый `held` packed-артефакт (321=packed_building /
//      322=packed_research) с payload `{construction_id, level}`.
//
// Активация packed-артефакта обрабатывается в `Service.Activate` через
// доп. ветку (см. activate_packed.go).

package artefact

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/pkg/ids"
)

// Legacy ID артефактов (consts.php 126-129).
const (
	UnitPackingBuilding = 323
	UnitPackingResearch = 324
	UnitPackedBuilding  = 321
	UnitPackedResearch  = 322
)

// Effect.Type строки (configs/artefacts.yml).
const (
	EffectPackBuilding    = "pack_building"
	EffectPackResearch    = "pack_research"
	EffectPackedBuilding  = "packed_building"
	EffectPackedResearch  = "packed_research"
)

// Ошибки pack-операции.
var (
	ErrPackingArtefactNotFound = errors.New("artefact: no held packing artefact for this planet/research")
	ErrNothingToPack           = errors.New("artefact: building/research has level=0 or all levels are 'added'")
	ErrPackBuildingMismatch    = errors.New("artefact: pack-building artefact requires planet_id (use moon_construction artefacts on moons)")
)

// PackBuilding упаковывает здание игрока в packed-артефакт.
//
// Алгоритм (legacy `BuildingInfo::packCurrentConstruction`):
//  1. Найти `held` packing-артефакт пользователя на этой планете
//     (state=held, unit_id=323).
//  2. Прочитать buildings.level. Origin не имеет колонки `added` —
//     для совместимости с legacy логикой `level - added` мы используем
//     `level` напрямую (added всегда = 0 в nova до создания механизма
//     packed-bonus). При активации packed-арта мы будем добавлять
//     уровни обратно в виртуальный счётчик; пока considered = level.
//  3. UPDATE buildings.level = level-1 (упаковываем 1 уровень).
//  4. Создать packed-артефакт unit_id=321 с
//     payload={construction_id: unitID, level: 1}.
//  5. Mark packing-артефакт как consumed.
func (s *Service) PackBuilding(ctx context.Context, userID, planetID string, buildingUnitID int) (Record, error) {
	var rec Record
	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// 1. Проверка ownership планеты.
		var ownerID string
		if err := tx.QueryRow(ctx,
			`SELECT user_id FROM planets WHERE id = $1`, planetID,
		).Scan(&ownerID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("planet owner: %w", err)
		}
		if ownerID != userID {
			return ErrNotOwner
		}

		// 2. Текущий уровень здания.
		var curLevel int
		err := tx.QueryRow(ctx,
			`SELECT level FROM buildings WHERE planet_id=$1 AND unit_id=$2`,
			planetID, buildingUnitID,
		).Scan(&curLevel)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("building level: %w", err)
		}
		if curLevel <= 0 {
			return ErrNothingToPack
		}

		// 3. Найти held packing-артефакт на этой планете.
		var packingArtID string
		err = tx.QueryRow(ctx, `
			SELECT id FROM artefacts_user
			WHERE user_id=$1 AND planet_id=$2 AND unit_id=$3 AND state=$4
			LIMIT 1
		`, userID, planetID, UnitPackingBuilding, StateHeld).Scan(&packingArtID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrPackingArtefactNotFound
			}
			return fmt.Errorf("find packing artefact: %w", err)
		}

		// 4. Понижаем уровень здания на 1.
		if _, err := tx.Exec(ctx, `
			UPDATE buildings SET level = level - 1
			WHERE planet_id=$1 AND unit_id=$2
		`, planetID, buildingUnitID); err != nil {
			return fmt.Errorf("decrement building level: %w", err)
		}

		// 5. Если уровень стал 0 — освобождаем поле планеты (зеркало
		//    HandleBuildConstruction / HandleDemolishConstruction).
		if curLevel-1 == 0 {
			if _, err := tx.Exec(ctx,
				`UPDATE planets SET used_fields = GREATEST(used_fields - 1, 0) WHERE id = $1`,
				planetID); err != nil {
				return fmt.Errorf("dec used_fields: %w", err)
			}
		}

		// 6. Mark packing artefact as consumed.
		if _, err := tx.Exec(ctx,
			`UPDATE artefacts_user SET state=$1 WHERE id=$2`,
			StateConsumed, packingArtID); err != nil {
			return fmt.Errorf("consume packing artefact: %w", err)
		}

		// 7. Создать packed-артефакт с payload.
		packedID := ids.New()
		payload := fmt.Sprintf(`{"construction_id":%d,"level":1}`, buildingUnitID)
		now := nowUTC()
		if _, err := tx.Exec(ctx, `
			INSERT INTO artefacts_user (id, user_id, planet_id, unit_id, state, acquired_at, payload)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, packedID, userID, planetID, UnitPackedBuilding, StateHeld, now, payload); err != nil {
			return fmt.Errorf("insert packed artefact: %w", err)
		}

		rec = Record{
			ID: packedID, UserID: userID, PlanetID: &planetID,
			UnitID: UnitPackedBuilding, State: StateHeld, AcquiredAt: now,
		}
		return nil
	})
	return rec, err
}

// PackResearch — упаковка исследования игрока. Аналог PackBuilding,
// но для research-таблицы (без planet_id, привязка к user_id).
//
// Legacy `BuildingInfo::packCurrentResearch`:
//  1. Найти `held` packing-research артефакт у игрока (state=held,
//     unit_id=324). Привязка к planet_id требуется (legacy строка
//     packCurrentResearch берёт `NS::getPlanet()` — текущая планета
//     должна совпадать с planet_id артефакта).
//  2. UPDATE research.level = level - 1.
//  3. Mark packing-арт as consumed.
//  4. Создать packed-research-арт (unit_id=322).
func (s *Service) PackResearch(ctx context.Context, userID, planetID string, researchUnitID int) (Record, error) {
	var rec Record
	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// 1. Проверка ownership планеты (как контекст активации).
		var ownerID string
		if err := tx.QueryRow(ctx,
			`SELECT user_id FROM planets WHERE id = $1`, planetID,
		).Scan(&ownerID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("planet owner: %w", err)
		}
		if ownerID != userID {
			return ErrNotOwner
		}

		// 2. Текущий уровень исследования.
		var curLevel int
		err := tx.QueryRow(ctx,
			`SELECT level FROM research WHERE user_id=$1 AND unit_id=$2`,
			userID, researchUnitID,
		).Scan(&curLevel)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("research level: %w", err)
		}
		if curLevel <= 0 {
			return ErrNothingToPack
		}

		// 3. Найти held packing-research артефакт на этой планете.
		var packingArtID string
		err = tx.QueryRow(ctx, `
			SELECT id FROM artefacts_user
			WHERE user_id=$1 AND planet_id=$2 AND unit_id=$3 AND state=$4
			LIMIT 1
		`, userID, planetID, UnitPackingResearch, StateHeld).Scan(&packingArtID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrPackingArtefactNotFound
			}
			return fmt.Errorf("find packing-research artefact: %w", err)
		}

		// 4. Понижаем уровень.
		if _, err := tx.Exec(ctx, `
			UPDATE research SET level = level - 1
			WHERE user_id=$1 AND unit_id=$2
		`, userID, researchUnitID); err != nil {
			return fmt.Errorf("decrement research level: %w", err)
		}

		// 5. Mark packing-арт as consumed.
		if _, err := tx.Exec(ctx,
			`UPDATE artefacts_user SET state=$1 WHERE id=$2`,
			StateConsumed, packingArtID); err != nil {
			return fmt.Errorf("consume packing-research artefact: %w", err)
		}

		// 6. Создать packed-research-арт.
		packedID := ids.New()
		payload := fmt.Sprintf(`{"construction_id":%d,"level":1}`, researchUnitID)
		now := nowUTC()
		if _, err := tx.Exec(ctx, `
			INSERT INTO artefacts_user (id, user_id, planet_id, unit_id, state, acquired_at, payload)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, packedID, userID, planetID, UnitPackedResearch, StateHeld, now, payload); err != nil {
			return fmt.Errorf("insert packed-research artefact: %w", err)
		}

		rec = Record{
			ID: packedID, UserID: userID, PlanetID: &planetID,
			UnitID: UnitPackedResearch, State: StateHeld, AcquiredAt: now,
		}
		return nil
	})
	return rec, err
}

// nowUTC — обёртка для тестов (можно подменить).
func nowUTC() time.Time { return time.Now().UTC() }
