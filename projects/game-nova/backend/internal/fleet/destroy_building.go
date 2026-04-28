package fleet

// План 65 Ф.3-Ф.4 (D-037): Building Destruction.
//
// Атаки KindAttackDestroyBuilding=26 (single) и
// KindAttackAllianceDestroyBuilding=29 (ACS) — после обычного боя
// при победе атакующего понижается уровень здания на планете-цели
// на 1 (с удалением здания при level==1→0).
//
// Семантика origin (Assault.class.php:599-651):
//   * target_building выбирается клиентом-инициатором миссии (передаётся
//     в payload как TargetBuildingID), либо случайно из подходящих зданий
//     планеты — getRandomTargetBuilding() в legacy.
//   * После боя, если target_destroyed (атакующий победил) и target ≠
//     луна — здание понижается: level-1 при level>1, удаление при level=1.
//   * Сообщение защитнику об утерянном уровне: MSG_BUILDING_DESTROYED.
//
// Решения для nova:
//
//   * **Выбор target_building**: если payload.TargetBuildingID > 0 —
//     используем; иначе выбираем случайно из buildings.unit_id планеты,
//     исключая UNIT_EXCHANGE=107 и UNIT_NANO_FACTORY=7 (origin-фильтр,
//     legacy consts.php:317-327). Сужения «у атакующего должно быть
//     здание сравнимого уровня» (Assault.class.php:253-272) — НЕ применяем
//     в nova: это легаси-эвристика устаревшего балансировочного компромисса
//     (DESTROY_BUILD_RESULT_MIN_OFFS_LEVEL), без неё миссия становится
//     более прямолинейной — что соответствует упрощённому подходу nova.
//     **Сознательное отклонение от legacy**, фиксируется в simplifications.md.
//
//   * **Условие срабатывания**: только при rep.Winner=="attackers" и
//     target ≠ луна (зеркалит legacy). Если планета — луна, ничего не
//     делаем (для лун есть отдельный KindAttackDestroyMoon).
//
//   * **Идемпотентность**: уровень читается с FOR UPDATE; если уже понижен
//     ниже текущего видимого payload-состояния (rare race), вторая попытка
//     no-op. Поскольку handler выполняется в той же транзакции, что и
//     applyDefenderLosses/finalizeAttack, повторная обработка события
//     невозможна без отдельной advisory-блокировки — событие закрывается
//     worker'ом единожды (FOR UPDATE SKIP LOCKED, см. план 09).
//
//   * **Очки**: нет необходимости вручную снижать `users.points` —
//     score derived state в nova (план 23, ScoreRecalcAll). Снос здания
//     отражается в очках при следующей пересборке через decorator
//     `withScore` после handler'а (worker/main.go).
//
//   * **used_fields**: при удалении здания (level=1→0) освобождаем 1
//     поле планеты (зеркалит HandleBuildConstruction/HandleDemolishConstruction).
//
//   * **Сообщения**: одно сообщение защитнику (folder=2, как в legacy
//     MSG_BUILDING_DESTROYED), одно атакующему/лидеру ACS — для UI-маркера
//     успеха миссии. i18n-ключи: assaultReport.buildingDestroyed* и
//     assaultReport.enemyBuildingDestroyed*.
//
//   * **Audit (R3)**: структурированный slog с полями event_id, planet_id,
//     unit_id, level_from, level_to, attacker_user_id, defender_user_id.
//
//   * **Метрики (R8)**: автоматически на уровне worker'а
//     (`oxsar_events_processed{kind="26|29"}`).

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/pkg/rng"
)

// Origin/legacy: UNIT_EXCHANGE и UNIT_NANO_FACTORY исключаются из
// «случайно выбранных целей сноса» (consts.php:93,24). Список замкнут
// на legacy origin — для nova-only вселенных список потенциально иной,
// но пока origin — единственный потребитель этих Kind'ов, держим
// зеркало legacy.
const (
	unitExchange     = 107
	unitNanoFactory  = 7
)

// tryDestroyBuilding — общая ветка для KindAttackDestroyBuilding и
// KindAttackAllianceDestroyBuilding. Понижает уровень здания на 1 при
// победе атакующего и не-лунной цели. Возвращает (unit_id, levelFrom,
// levelTo, true) при успешном понижении; (0,0,0,false) если ничего не
// сделано (не выбрано здание, проигрыш атакующего, луна, и т.п.).
//
// Параметры:
//   - planetID: цель.
//   - isMoon: если true — handler ничего не делает (для лун —
//     KindAttackDestroyMoon).
//   - winner: rep.Winner; здание ломается только при "attackers".
//   - explicitTargetUnitID: payload.TargetBuildingID; 0 → выбрать случайно.
//   - battleSeed: rep.Seed, используется как seed детерминированного
//     случайного выбора (если explicit не задан).
func tryDestroyBuilding(ctx context.Context, tx pgx.Tx,
	planetID string, isMoon bool, winner string,
	explicitTargetUnitID int, battleSeed uint64) (int, int, int, bool, error) {

	if isMoon {
		return 0, 0, 0, false, nil
	}
	if winner != "attackers" {
		return 0, 0, 0, false, nil
	}

	// 1. Определяем target_unit_id.
	targetUnitID := explicitTargetUnitID
	if targetUnitID <= 0 {
		// Случайный выбор: все здания планеты с level>0, кроме исключений.
		rows, err := tx.Query(ctx, `
			SELECT unit_id FROM buildings
			WHERE planet_id=$1 AND level > 0
			  AND unit_id NOT IN ($2, $3)
			ORDER BY unit_id
		`, planetID, unitExchange, unitNanoFactory)
		if err != nil {
			return 0, 0, 0, false, fmt.Errorf("destroy_building: select candidates: %w", err)
		}
		var ids []int
		for rows.Next() {
			var id int
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return 0, 0, 0, false, err
			}
			ids = append(ids, id)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return 0, 0, 0, false, err
		}
		if len(ids) == 0 {
			// На планете нет подходящих зданий — handler no-op.
			return 0, 0, 0, false, nil
		}
		// Детерминированный выбор по seed боя — repro отчётов.
		r := rng.New(battleSeed ^ 0xB1D0DE578C0FFEE)
		targetUnitID = ids[r.IntN(len(ids))]
	}

	// 2. Читаем текущий level с lock'ом.
	var curLevel int
	err := tx.QueryRow(ctx,
		`SELECT level FROM buildings WHERE planet_id=$1 AND unit_id=$2 FOR UPDATE`,
		planetID, targetUnitID,
	).Scan(&curLevel)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Здание уже снесено / не построено. No-op.
			return 0, 0, 0, false, nil
		}
		return 0, 0, 0, false, fmt.Errorf("destroy_building: select level: %w", err)
	}
	if curLevel <= 0 {
		return 0, 0, 0, false, nil
	}

	newLevel := curLevel - 1
	if _, err := tx.Exec(ctx,
		`UPDATE buildings SET level=$1 WHERE planet_id=$2 AND unit_id=$3`,
		newLevel, planetID, targetUnitID,
	); err != nil {
		return 0, 0, 0, false, fmt.Errorf("destroy_building: update level: %w", err)
	}

	// 3. Если здание полностью снесено — освободить поле планеты
	// (зеркало HandleDemolishConstruction).
	if newLevel == 0 {
		if _, err := tx.Exec(ctx,
			`UPDATE planets SET used_fields = GREATEST(used_fields - 1, 0) WHERE id=$1`,
			planetID); err != nil {
			return 0, 0, 0, false, fmt.Errorf("destroy_building: dec used_fields: %w", err)
		}
	}

	return targetUnitID, curLevel, newLevel, true, nil
}

// sendBuildingDestroyedMessages — два сообщения: защитнику и атакующему.
// Использует уже существующую sendMoonMessage (см. moon_destruction.go) —
// семантика «folder=2, from=NULL» одинакова. i18n-ключи:
// assaultReport.buildingDestroyed{Subject,Body} —
// assaultReport.enemyBuildingDestroyed{Subject,Body}.
func sendBuildingDestroyedMessages(ctx context.Context, tx pgx.Tx,
	tr func(component, key string, vars map[string]string) string,
	defenderUserID, attackerUserID string,
	unitID, levelFrom, levelTo int, eventID string) error {

	vars := map[string]string{
		"unit_id":    strconv.Itoa(unitID),
		"level_from": strconv.Itoa(levelFrom),
		"level_to":   strconv.Itoa(levelTo),
	}
	if defenderUserID != "" {
		if err := sendMoonMessage(ctx, tx, defenderUserID,
			tr("assaultReport", "buildingDestroyedSubject", vars),
			tr("assaultReport", "buildingDestroyedBody", vars)); err != nil {
			return err
		}
	}
	if attackerUserID != "" && attackerUserID != defenderUserID {
		if err := sendMoonMessage(ctx, tx, attackerUserID,
			tr("assaultReport", "enemyBuildingDestroyedSubject", vars),
			tr("assaultReport", "enemyBuildingDestroyedBody", vars)); err != nil {
			return err
		}
	}
	slog.InfoContext(ctx, "event_destroy_building_applied",
		slog.String("event_id", eventID),
		slog.Int("unit_id", unitID),
		slog.Int("level_from", levelFrom),
		slog.Int("level_to", levelTo),
		slog.String("defender_user_id", defenderUserID),
		slog.String("attacker_user_id", attackerUserID))
	return nil
}

