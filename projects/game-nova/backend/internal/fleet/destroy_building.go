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

// destroyBuildResultMinOffsLevel — legacy DESTROY_BUILD_RESULT_MIN_OFFS_LEVEL
// (consts.php). Значение в legacy `consts.dm.local.php`/`consts.php` обычно 0:
// исключаем здание защитника, если после понижения его уровень станет
// ниже уровня соответствующего здания у атакующего. План 72.1.56 B7
// 1:1 с legacy `Assault.class.php:261-263`.
const destroyBuildResultMinOffsLevel = 0

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
//   - attackerUserIDs: список user_id всех атакующих (1 для single,
//     N для ACS); используется для эвристики
//     DESTROY_BUILD_RESULT_MIN_OFFS_LEVEL — здание исключается, если
//     `defender.level - 1 < max_attacker_level + threshold` или ни у
//     кого из атакующих этого здания нет (план 72.1.56 B7, legacy
//     `Assault.class.php:253-272`).
//   - battleSeed: rep.Seed, используется как seed детерминированного
//     случайного выбора (если explicit не задан).
func tryDestroyBuilding(ctx context.Context, tx pgx.Tx,
	planetID string, isMoon bool, winner string,
	explicitTargetUnitID int, attackerUserIDs []string, battleSeed uint64) (int, int, int, bool, error) {

	if isMoon {
		return 0, 0, 0, false, nil
	}
	if winner != "attackers" {
		return 0, 0, 0, false, nil
	}

	// 1. Определяем target_unit_id.
	targetUnitID := explicitTargetUnitID
	if targetUnitID <= 0 {
		candidates, err := selectRandomDestroyCandidates(ctx, tx, planetID, attackerUserIDs)
		if err != nil {
			return 0, 0, 0, false, err
		}
		if len(candidates) == 0 {
			// На планете нет подходящих зданий — handler no-op.
			return 0, 0, 0, false, nil
		}
		// Детерминированный выбор по seed боя — repro отчётов.
		r := rng.New(battleSeed ^ 0xB1D0DE578C0FFEE)
		targetUnitID = candidates[r.IntN(len(candidates))]
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

// selectRandomDestroyCandidates возвращает список unit_id зданий
// планеты-цели, доступных для случайного выбора при destroy_building
// миссии. Зеркало legacy `Assault.class.php:223-310::getRandomTargetBuilding`:
//
//  1. Берём все здания цели с level>0, кроме UNIT_EXCHANGE и
//     UNIT_NANO_FACTORY (origin-фильтр).
//  2. Для каждого attacker_user_id читаем MAX(level) каждого здания
//     по всем его планетам. Здание защитника **исключается**, если
//     у этого attacker'а есть это здание И
//     `defender.level - 1 < max_attacker_level + threshold`. Иными
//     словами — атакующий должен иметь сравнимое здание чтобы понимать
//     цель сноса (legacy балансовый компромисс).
//  3. Здание, которое **ни одним** attacker'ом не покрыто, тоже
//     исключается («unchecked» в legacy).
//
// Если attackerUserIDs пуст (нет известных атакующих, например,
// alien-attack handler не передал) — возвращаем кандидатов без
// эвристики (legacy fallback: эвристика отключена).
func selectRandomDestroyCandidates(ctx context.Context, tx pgx.Tx,
	planetID string, attackerUserIDs []string) ([]int, error) {

	rows, err := tx.Query(ctx, `
		SELECT unit_id, level FROM buildings
		WHERE planet_id=$1 AND level > 0
		  AND unit_id NOT IN ($2, $3)
		ORDER BY unit_id
	`, planetID, unitExchange, unitNanoFactory)
	if err != nil {
		return nil, fmt.Errorf("destroy_building: select candidates: %w", err)
	}
	defenderBuilds := map[int]int{}
	for rows.Next() {
		var id, lvl int
		if err := rows.Scan(&id, &lvl); err != nil {
			rows.Close()
			return nil, err
		}
		defenderBuilds[id] = lvl
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(defenderBuilds) == 0 {
		return nil, nil
	}

	// Без атакующих — fallback на старое поведение (без эвристики).
	if len(attackerUserIDs) == 0 {
		return filterDestroyCandidates(defenderBuilds, nil), nil
	}

	// Для всех attacker'ов: MAX(level) каждого здания по всем планетам.
	atkRows, err := tx.Query(ctx, `
		SELECT b.unit_id, MAX(b.level) AS max_level
		FROM buildings b
		JOIN planets p ON p.id = b.planet_id
		WHERE p.user_id = ANY($1) AND p.destroyed_at IS NULL
		GROUP BY b.unit_id
	`, attackerUserIDs)
	if err != nil {
		return nil, fmt.Errorf("destroy_building: select attacker buildings: %w", err)
	}
	attackerMaxLevels := map[int]int{}
	for atkRows.Next() {
		var unitID, maxLvl int
		if err := atkRows.Scan(&unitID, &maxLvl); err != nil {
			atkRows.Close()
			return nil, err
		}
		attackerMaxLevels[unitID] = maxLvl
	}
	atkRows.Close()
	if err := atkRows.Err(); err != nil {
		return nil, err
	}

	return filterDestroyCandidates(defenderBuilds, attackerMaxLevels), nil
}

// filterDestroyCandidates — pure-функция legacy-эвристики
// `Assault.class.php:253-281`. Вынесена для unit-тестируемости
// без БД (план 72.1.56 B7).
//
// Если attackerMaxLevels=nil — эвристика выключена, возвращаем все
// defender_builds (legacy fallback при отсутствии attacker'ов).
//
// Иначе:
//   - Для каждого defender_build: если у атакующих НЕТ этого здания
//     («unchecked» в legacy) → исключаем.
//   - Если есть и `defender.level - 1 < attacker.max_level + threshold`
//     → исключаем («new_result_level < min_result_level»).
//
// Результат отсортирован по unit_id для repro.
func filterDestroyCandidates(defenderBuilds, attackerMaxLevels map[int]int) []int {
	out := make([]int, 0, len(defenderBuilds))
	for id, defLvl := range defenderBuilds {
		if attackerMaxLevels == nil {
			out = append(out, id)
			continue
		}
		atkLvl, ok := attackerMaxLevels[id]
		if !ok {
			// «unchecked» — исключаем.
			continue
		}
		minResultLevel := atkLvl + destroyBuildResultMinOffsLevel
		newResultLevel := defLvl - 1
		if newResultLevel < minResultLevel {
			continue
		}
		out = append(out, id)
	}
	sortIntsAsc(out)
	return out
}

func sortIntsAsc(a []int) {
	for i := 1; i < len(a); i++ {
		v := a[i]
		j := i - 1
		for j >= 0 && a[j] > v {
			a[j+1] = a[j]
			j--
		}
		a[j+1] = v
	}
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

