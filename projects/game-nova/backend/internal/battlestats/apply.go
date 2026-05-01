// Package battlestats: ApplyBattleResult — порт Java Participant.java:
// 924-987 (план 72.1.1).
//
// После реального боя (НЕ симулятор) каждому участнику-юзеру:
//   - INSERT в user_experience (idempotent через UNIQUE
//     (battle_id, user_id, is_atter)).
//   - users.e_points  += experience.
//   - users.be_points += experience.
//   - users.battles   += 1.
//   - users.points    -= lost_points.  (GREATEST(0, ...))
//   - users.u_points  -= lost_points.  (GREATEST(0, ...))
//   - users.u_count   -= lost_units.   (GREATEST(0, ...))
//
// Ограничения:
//   - alien-стороны (IsAliens=true) пропускаются — это NPC.
//   - Юзер появляется в reported-set только один раз на бой
//     (legacy useridReported). Если игрок и в Attackers, и в
//     Defenders (теоретически возможно через ACS), второй проход
//     skip'ается.
//   - Idempotency: повторный вызов вернёт ErrAlreadyApplied для уже
//     обработанных юзеров; первый раз отработает нормально, второй —
//     скипнет всех (ON CONFLICT DO NOTHING).
package battlestats

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/battle"
)

// ErrAlreadyApplied — все юзеры этого боя уже получили опыт ранее.
// Возвращается ApplyBattleResult если ни один INSERT не прошёл из-за
// дубликата (battle_id, user_id, is_atter). Не critical-ошибка —
// caller обычно её игнорирует (на стороне event-loop это normal flow
// при re-process).
var ErrAlreadyApplied = errors.New("battlestats: already applied for this battle")

// ApplyBattleResult — главная функция (см. doc-комментарий пакета).
//
// tx обязателен: все 4 SQL'а (insert log + 2 update users + повтор
// для defender) в одной транзакции с battle_reports / списанием
// потерь / выдачей добычи.
//
// battleID — uuid отчёта в battle_reports. Пустая строка допускается
// для сценариев без battle_reports (alien-рейды, экспедиционные бои с
// NPC) — тогда `user_experience.battle_id = NULL`. В этом случае
// idempotency-индекс UNIQUE (battle_id, user_id, is_atter) НЕ
// защищает от дубликатов (NULL != NULL в SQL UNIQUE), и
// ответственность за exactly-once-обработку ложится на caller
// (обычно event-loop с exactly-once-семантикой). Trade-off зафиксирован
// в docs/simplifications.md «battle exp w/o battle_reports».
func ApplyBattleResult(
	ctx context.Context,
	tx pgx.Tx,
	report battle.Report,
	battleID string,
) error {
	// useridReported — порт Java Set<Integer> useridReported. Один
	// и тот же юзер не получает дважды опыт за один бой.
	reported := make(map[string]bool)

	any := false
	for _, side := range report.Attackers {
		ok, err := applySide(ctx, tx, battleID, side, true, report.AttackerExp, reported)
		if err != nil {
			return err
		}
		any = any || ok
	}
	for _, side := range report.Defenders {
		ok, err := applySide(ctx, tx, battleID, side, false, report.DefenderExp, reported)
		if err != nil {
			return err
		}
		any = any || ok
	}

	if !any {
		return ErrAlreadyApplied
	}
	return nil
}

// applySide — обработать одну SideResult. Возвращает (applied, err):
// applied=true если хотя бы один INSERT/UPDATE прошёл, false — если
// сторона была пропущена (alien / уже reported / дубликат в логе).
func applySide(
	ctx context.Context,
	tx pgx.Tx,
	battleID string,
	side battle.SideResult,
	isAtter bool,
	experience int,
	reported map[string]bool,
) (bool, error) {
	if side.IsAliens {
		return false, nil
	}
	if side.UserID == "" {
		return false, nil
	}
	if reported[side.UserID] {
		return false, nil
	}
	reported[side.UserID] = true

	// 1. Лог опыта. ON CONFLICT гарантирует idempotency только для
	// случаев когда battleID непустой (NULL UNIQUE не блокирует
	// дубликаты — см. doc ApplyBattleResult).
	var battleIDArg interface{}
	if battleID != "" {
		battleIDArg = battleID
	} else {
		battleIDArg = nil
	}
	tag, err := tx.Exec(ctx, `
		INSERT INTO user_experience (battle_id, user_id, is_atter, experience)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (battle_id, user_id, is_atter) DO NOTHING
	`, battleIDArg, side.UserID, isAtter, experience)
	if err != nil {
		return false, fmt.Errorf("battlestats: insert user_experience: %w", err)
	}
	if tag.RowsAffected() == 0 {
		// Уже применили этот бой к этому юзеру — пропускаем UPDATE'ы тоже.
		return false, nil
	}

	// 2. Прирост опыта/боёв.
	if _, err := tx.Exec(ctx, `
		UPDATE users
		SET e_points  = e_points  + $1,
		    be_points = be_points + $1,
		    battles   = battles   + 1
		WHERE id = $2
	`, experience, side.UserID); err != nil {
		return false, fmt.Errorf("battlestats: update users (exp): %w", err)
	}

	// 3. Списание потерь. GREATEST(0, ...) — не уходим в минус
	// (порт Java Participant.java GREATEST(0, points - lostPoints)).
	if _, err := tx.Exec(ctx, `
		UPDATE users
		SET points   = GREATEST(0, points   - $1),
		    u_points = GREATEST(0, u_points - $1),
		    u_count  = GREATEST(0, u_count  - $2)
		WHERE id = $3
	`, side.LostPoints, side.LostUnits, side.UserID); err != nil {
		return false, fmt.Errorf("battlestats: update users (losses): %w", err)
	}

	return true, nil
}
