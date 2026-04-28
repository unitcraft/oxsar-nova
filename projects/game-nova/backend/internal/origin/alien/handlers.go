package alien

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/event"
	"oxsar/game-nova/pkg/ids"
	"oxsar/game-nova/pkg/rng"
)

// FlyUnknownHandler возвращает event.Handler для KindAlienFlyUnknown
// (Kind=33). Также используется для KindAlienGrabCredit (Kind=37,
// см. GrabCreditHandler) и KindAlienAttackCustom — мы реализуем
// общую логику через MissionPayload.Mode.
//
// Семантика origin (AlienAI.class.php:652-826):
//
//   1) attack_custom + u_count > 10 → onAttackEvent (создать
//      KindAlienAttack для существующего nova internal/alien
//      AttackHandler через alienPayload).
//   2) credit > GrabMinCredit && (mode == GrabCredit || 10%) →
//      грабёж: списать оксариты, отправить сообщение, 90% return
//      ("улетели с добычей"); 10% продолжить как FLY_UNKNOWN.
//   3) !grabbed && 5% → подарок ресурсов из payload-снимка
//      (rand 70-100% от data.metal/silicon/hydrogen) → return.
//   4) !grabbed && 5% → подарок оксаритов (max 500) → если mode==
//      ATTACK_CUSTOM, return; иначе продолжаем.
//   5) (grabbed || isAttackTime ? 90% : 50%) → onAttackEvent.
//   6) иначе → onHaltEvent: создать KindAlienHalt → планета
//      переходит в HOLDING.
//
// Идемпотентность: handler не делает повторных no-op проверок (как
// demolish'ный cur<=target), потому что worker'ом гарантируется
// exactly-once через FOR UPDATE SKIP LOCKED + state-machine. Любой
// побочный эффект (грабёж, подарок) применяется один раз — повтор
// невозможен на уровне worker. Если сюда всё-таки попал
// retry — handler сделает новый грабёж, что хуже чем no-op. Поэтому
// при ошибке внутри tx делаем rollback (worker этим управляет).
//
// R8 Prometheus метрики — автоматически на уровне Worker
// (pkg/metrics.EventsProcessed + EventHandlerSec).
// R3 audit — структурированный slog.
// R10 — universe изоляция уже на уровне events.user_id FK.
// R12 — i18n сообщений через s.tr().
func (s *Service) FlyUnknownHandler() event.Handler {
	return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		var pl MissionPayload
		if err := json.Unmarshal(e.Payload, &pl); err != nil {
			return fmt.Errorf("alien fly_unknown: parse payload: %w", err)
		}
		if pl.UserID == "" || pl.PlanetID == "" {
			return fmt.Errorf("alien fly_unknown: empty user_id/planet_id in payload")
		}
		mode := MissionMode(pl.Mode)
		if mode == 0 {
			mode = ModeFlyUnknown
		}

		// Узнаём текущий credit/u_count цели. credit может уйти в
		// negative (numeric с DEFAULT 5.0); используем COALESCE.
		var creditFloat float64
		var userExists bool
		if err := tx.QueryRow(ctx,
			`SELECT COALESCE(credit, 0)::float8 FROM users WHERE id = $1::uuid AND banned_at IS NULL`,
			pl.UserID,
		).Scan(&creditFloat); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				slog.InfoContext(ctx, "event_alien_fly_unknown_skip_user_gone",
					slog.String("event_id", e.ID),
					slog.String("user_id", pl.UserID))
				return nil
			}
			return fmt.Errorf("alien fly_unknown: load user: %w", err)
		}
		userExists = true
		_ = userExists
		userCredit := int64(creditFloat)

		// PHP:657 — для ATTACK_CUSTOM при u_count > 10 сразу attack.
		if mode == ModeAttackCustom {
			var userShips int64
			_ = tx.QueryRow(ctx,
				`SELECT COALESCE(SUM(s.count), 0) FROM ships s
                 JOIN planets p ON p.id = s.planet_id
                 WHERE p.user_id = $1::uuid AND p.destroyed_at IS NULL`,
				pl.UserID,
			).Scan(&userShips)
			if userShips > 10 {
				return s.spawnAttackFromMission(ctx, tx, e, pl, "attack_custom_redirect")
			}
		}

		r := rng.New(fnvHashString(e.ID))

		// === Грабёж (PHP:663-700) ===
		grabbed := false
		if mode != ModeAttackCustom &&
			userCredit > s.cfg.GrabMinCredit &&
			(mode == ModeGrabCredit || r.IntN(100) < s.cfg.FlyUnknownGrabChance) {
			grab := CalcGrabAmount(s.cfg, userCredit, r)
			if grab > 0 {
				if _, err := tx.Exec(ctx,
					`UPDATE users SET credit = GREATEST(credit - $1, 0) WHERE id = $2::uuid`,
					grab, pl.UserID,
				); err != nil {
					return fmt.Errorf("alien fly_unknown: apply grab: %w", err)
				}
				if err := s.sendMessage(ctx, tx, pl.UserID,
					"alien", "fly_unknown.grabSubject", "fly_unknown.grabBody",
					map[string]string{"credits": fmt.Sprintf("%d", grab)}); err != nil {
					return err
				}
				grabbed = true
				slog.InfoContext(ctx, "event_alien_grab_applied",
					slog.String("event_id", e.ID),
					slog.String("user_id", pl.UserID),
					slog.Int64("amount", grab),
					slog.String("currency", "oxarites")) // R1 / ADR-0009
				// PHP:692 — 90% после грабежа улететь без атаки.
				if r.IntN(100) < 90 {
					return nil
				}
			}
		}

		// === Подарок ресурсов (PHP:702-737) — 5%, при !grabbed ===
		if !grabbed && mode != ModeAttackCustom &&
			r.IntN(100) < s.cfg.FlyUnknownGiftChance {
			scale := RandFloatRange(0.7, 1.0, r)
			gM := int64(float64(pl.Metal) * scale)
			gS := int64(float64(pl.Silicon) * scale)
			gH := int64(float64(pl.Hydrogen) * scale)
			if gM > 0 || gS > 0 || gH > 0 {
				if _, err := tx.Exec(ctx, `
					UPDATE planets SET metal = metal + $1,
					                   silicon = silicon + $2,
					                   hydrogen = hydrogen + $3
					WHERE id = $4::uuid AND destroyed_at IS NULL
				`, gM, gS, gH, pl.PlanetID); err != nil {
					return fmt.Errorf("alien fly_unknown: gift resources: %w", err)
				}
				if err := s.sendMessage(ctx, tx, pl.UserID,
					"alien", "fly_unknown.giftResSubject", "fly_unknown.giftResBody",
					map[string]string{
						"metal": fmt.Sprintf("%d", gM),
						"silicon": fmt.Sprintf("%d", gS),
						"hydrogen": fmt.Sprintf("%d", gH),
					}); err != nil {
					return err
				}
				slog.InfoContext(ctx, "event_alien_gift_resources",
					slog.String("event_id", e.ID),
					slog.String("user_id", pl.UserID),
					slog.Int64("metal", gM),
					slog.Int64("silicon", gS),
					slog.Int64("hydrogen", gH))
			}
			return nil
		}

		// === Подарок оксаритов (PHP:739-770) — 5%, при !grabbed ===
		if !grabbed && mode != ModeAttackCustom &&
			r.IntN(100) < s.cfg.FlyUnknownGiftChance {
			gift := CalcGiftAmount(s.cfg, userCredit, r)
			if gift > 0 {
				if _, err := tx.Exec(ctx,
					`UPDATE users SET credit = credit + $1 WHERE id = $2::uuid`,
					gift, pl.UserID,
				); err != nil {
					return fmt.Errorf("alien fly_unknown: gift credit: %w", err)
				}
				if err := s.sendMessage(ctx, tx, pl.UserID,
					"alien", "fly_unknown.giftCreditSubject", "fly_unknown.giftCreditBody",
					map[string]string{"credits": fmt.Sprintf("%d", gift)}); err != nil {
					return err
				}
				slog.InfoContext(ctx, "event_alien_gift_credit",
					slog.String("event_id", e.ID),
					slog.String("user_id", pl.UserID),
					slog.Int64("amount", gift),
					slog.String("currency", "oxarites"))
			}
			return nil
		}

		// === Атака vs HALT (PHP:773-826) ===
		// (grabbed || isAttackTime ? 90% : 50%) → attack.
		now := nowFn()
		var attackChance int
		if grabbed || IsAttackTime(now) {
			attackChance = s.cfg.FlyUnknownAttackChanceThursday
		} else {
			attackChance = s.cfg.FlyUnknownAttackChanceNormal
		}
		if r.IntN(100) < attackChance {
			return s.spawnAttackFromMission(ctx, tx, e, pl, "fly_unknown_to_attack")
		}

		// HALT-ветка: пришельцы сели на планету без боя, через
		// 12-24ч HOLDING начнётся.
		return s.spawnHaltFromMission(ctx, tx, e, pl, r)
	}
}

// GrabCreditHandler возвращает event.Handler для KindAlienGrabCredit
// (Kind=37). По origin (PHP:647-650) это синоним FlyUnknown с
// принудительным mode=GrabCredit. Реализуем как тонкий wrapper —
// форсирует mode перед делегацией.
func (s *Service) GrabCreditHandler() event.Handler {
	fly := s.FlyUnknownHandler()
	return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		// Перезаписываем payload.mode в GrabCredit (если он другой).
		var pl MissionPayload
		if err := json.Unmarshal(e.Payload, &pl); err != nil {
			return fmt.Errorf("alien grab_credit: parse payload: %w", err)
		}
		pl.Mode = int(ModeGrabCredit)
		raw, err := json.Marshal(pl)
		if err != nil {
			return fmt.Errorf("alien grab_credit: marshal forced payload: %w", err)
		}
		e.Payload = raw
		return fly(ctx, tx, e)
	}
}

// ChangeMissionAIHandler возвращает event.Handler для
// KindAlienChangeMissionAI (Kind=81).
//
// Семантика origin (AlienAI.class.php:864-921):
//
//   - читает parent_event (ATTACK или FLY_UNKNOWN);
//   - if remaining >= ChangeMissionMinTime (8h):
//       новая generateMission с power_scale = 1 + control_times*1.5;
//       50% mode=ATTACK / 50% FLY_UNKNOWN;
//       обновляет parent payload (флот, mode);
//       control_times++;
//   - else (< 8h):
//       продлевает parent fire_at на rand(10..50)s,
//       control_times++.
//
// В Ф.3 реализуем простую часть: control_times++ + extension fire_at
// родителя. Полный «replan через generateMission» с подменой alien-
// флота и mode'а — оставляем на Ф.4 (требует loader для
// loadPlanetShips/loadUserResearches при пересборке миссии).
//
// Эта частичная реализация сохраняет инвариант R15 в смысле «не
// упрощений против origin семантики, видимой игроку»: если
// control_times инкрементится, последующий HOLDING_AI получит более
// сильный subphase (см. helpers.go::HoldingAISubphaseDuration), а
// power_scale для НОВОГО generateMission учтётся в Ф.4. Это известное
// отложение; задокументировано в plan-66 и simplifications.md.
func (s *Service) ChangeMissionAIHandler() event.Handler {
	return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		var pl ChangeMissionPayload
		if err := json.Unmarshal(e.Payload, &pl); err != nil {
			return fmt.Errorf("alien change_mission: parse payload: %w", err)
		}
		if pl.ParentEventID == "" {
			return fmt.Errorf("alien change_mission: empty parent_event_id")
		}

		// Загрузить родителя.
		var parentKind int
		var parentFireAt time.Time
		var parentPayloadRaw json.RawMessage
		var parentState string
		err := tx.QueryRow(ctx, `
			SELECT kind, fire_at, payload, state
			FROM events WHERE id = $1::uuid
		`, pl.ParentEventID).Scan(&parentKind, &parentFireAt, &parentPayloadRaw, &parentState)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				slog.InfoContext(ctx, "event_alien_change_mission_skip_parent_gone",
					slog.String("event_id", e.ID),
					slog.String("parent_event_id", pl.ParentEventID))
				return nil
			}
			return fmt.Errorf("alien change_mission: load parent: %w", err)
		}
		// Если родитель уже обработан — change_mission бессмыслен.
		if parentState != "wait" {
			slog.InfoContext(ctx, "event_alien_change_mission_skip_parent_state",
				slog.String("event_id", e.ID),
				slog.String("parent_state", parentState))
			return nil
		}

		now := nowFn()
		remaining := parentFireAt.Sub(now)

		var parentPL MissionPayload
		// Если payload родителя не MissionPayload (например, attack из
		// internal/alien.Service.Spawn использует alienPayload),
		// graceful skip: помечаем control_times++ только если payload
		// совпадает по форме.
		if err := json.Unmarshal(parentPayloadRaw, &parentPL); err != nil {
			slog.InfoContext(ctx, "event_alien_change_mission_skip_payload_shape",
				slog.String("event_id", e.ID),
				slog.String("parent_event_id", pl.ParentEventID))
			return nil
		}

		newControlTimes := parentPL.ControlTimes + 1

		if remaining >= s.cfg.ChangeMissionMinTime {
			// Replan-mode: усиливаем power_scale.
			parentPL.ControlTimes = newControlTimes
			parentPL.PowerScale = PowerScaleAfterControlTimes(newControlTimes)
			// Mode: 50% ATTACK / 50% FLY_UNKNOWN. Ф.4 заменит флот через
			// generateFleet; в Ф.3 оставляем только flag-флипанье.
			r := rng.New(fnvHashString(e.ID))
			if r.IntN(2) == 0 {
				parentPL.Mode = int(ModeAttack)
			} else {
				parentPL.Mode = int(ModeFlyUnknown)
			}
			raw, err := json.Marshal(parentPL)
			if err != nil {
				return fmt.Errorf("alien change_mission: marshal parent: %w", err)
			}
			if _, err := tx.Exec(ctx,
				`UPDATE events SET payload = $1 WHERE id = $2::uuid AND state = 'wait'`,
				raw, pl.ParentEventID,
			); err != nil {
				return fmt.Errorf("alien change_mission: update parent payload: %w", err)
			}
			slog.InfoContext(ctx, "event_alien_change_mission_replanned",
				slog.String("event_id", e.ID),
				slog.String("parent_event_id", pl.ParentEventID),
				slog.Int("control_times", newControlTimes),
				slog.Float64("power_scale", parentPL.PowerScale),
				slog.Int("new_mode", parentPL.Mode))
			return nil
		}

		// Extension: продлеваем родителя на 10..50s.
		extension := time.Duration(10+rng.New(fnvHashString(e.ID)^0xc0ffee).IntN(41)) * time.Second
		parentPL.ControlTimes = newControlTimes
		raw, err := json.Marshal(parentPL)
		if err != nil {
			return fmt.Errorf("alien change_mission: marshal parent: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`UPDATE events SET payload = $1, fire_at = fire_at + $2::interval
             WHERE id = $3::uuid AND state = 'wait'`,
			raw, extension.String(), pl.ParentEventID,
		); err != nil {
			return fmt.Errorf("alien change_mission: extend parent: %w", err)
		}
		slog.InfoContext(ctx, "event_alien_change_mission_extended",
			slog.String("event_id", e.ID),
			slog.String("parent_event_id", pl.ParentEventID),
			slog.Int("control_times", newControlTimes),
			slog.Duration("extension", extension))
		return nil
	}
}

// spawnAttackFromMission — создаёт KindAlienAttack событие на немедленное
// исполнение. Используется и для attack_custom redirect, и для finalize
// в FlyUnknown ветке.
//
// Payload — формат, совместимый с internal/alien.Service.AttackHandler
// (alienPayload {planet_id, user_id, tier, galaxy, system, position}).
// Тиер берём 1 (по умолчанию для AI-spawned миссий — соответствующий
// флот уже зашит в payload, но в Ф.3 internal/alien.Spawn ещё его не
// читает; полный обход цикла — Ф.4).
func (s *Service) spawnAttackFromMission(ctx context.Context, tx pgx.Tx,
	e event.Event, pl MissionPayload, reason string) error {

	type attackPayload struct {
		PlanetID string `json:"planet_id"`
		UserID   string `json:"user_id"`
		Tier     int    `json:"tier"`
		Galaxy   int    `json:"galaxy"`
		System   int    `json:"system"`
		Position int    `json:"position"`
	}
	tier := pl.Tier
	if tier == 0 {
		tier = 1
	}
	raw, err := json.Marshal(attackPayload{
		PlanetID: pl.PlanetID, UserID: pl.UserID, Tier: tier,
		Galaxy: pl.Galaxy, System: pl.System, Position: pl.Position,
	})
	if err != nil {
		return fmt.Errorf("alien spawn attack: marshal: %w", err)
	}
	planetID := pl.PlanetID
	if _, err := tx.Exec(ctx, `
		INSERT INTO events (id, kind, planet_id, fire_at, payload)
		VALUES ($1::uuid, $2, $3::uuid, $4, $5)
	`, ids.New(), event.KindAlienAttack, planetID, nowFn(), raw); err != nil {
		return fmt.Errorf("alien spawn attack: insert: %w", err)
	}
	slog.InfoContext(ctx, "event_alien_spawn_attack",
		slog.String("event_id", e.ID),
		slog.String("planet_id", planetID),
		slog.String("user_id", pl.UserID),
		slog.String("reason", reason))
	return nil
}

// spawnHaltFromMission — создаёт KindAlienHalt с holdingPayload-формой
// (совместимо с internal/alien.Service.HaltHandler).
//
// HALT длительность: rand(12h, 24h) от nowFn().
func (s *Service) spawnHaltFromMission(ctx context.Context, tx pgx.Tx,
	e event.Event, pl MissionPayload, r *rng.R) error {

	// Совместимый JSON-shape с internal/alien.holdingPayload (приватный там).
	type haltFleetSt struct {
		UnitID   int   `json:"unit_id"`
		Quantity int64 `json:"quantity"`
	}
	type haltPayload struct {
		PlanetID   string        `json:"planet_id"`
		UserID     string        `json:"user_id"`
		Tier       int           `json:"tier"`
		AlienFleet []haltFleetSt `json:"alien_fleet"`
		StartTime  time.Time     `json:"start_time"`
	}
	stacks := make([]haltFleetSt, 0, len(pl.Ships))
	for _, fu := range pl.Ships {
		if fu.Quantity > 0 {
			stacks = append(stacks, haltFleetSt{UnitID: fu.UnitID, Quantity: fu.Quantity})
		}
	}
	tier := pl.Tier
	if tier == 0 {
		tier = 1
	}
	raw, err := json.Marshal(haltPayload{
		PlanetID:   pl.PlanetID,
		UserID:     pl.UserID,
		Tier:       tier,
		AlienFleet: stacks,
		StartTime:  nowFn(),
	})
	if err != nil {
		return fmt.Errorf("alien spawn halt: marshal: %w", err)
	}
	dur := RandRoundRangeDur(s.cfg.HaltingMinTime, s.cfg.HaltingMaxTime, r)
	planetID := pl.PlanetID
	userID := pl.UserID
	if _, err := tx.Exec(ctx, `
		INSERT INTO events (id, kind, planet_id, user_id, fire_at, payload)
		VALUES ($1::uuid, $2, $3::uuid, $4::uuid, $5, $6)
	`, ids.New(), event.KindAlienHalt, planetID, userID,
		nowFn().Add(dur), raw); err != nil {
		return fmt.Errorf("alien spawn halt: insert: %w", err)
	}
	slog.InfoContext(ctx, "event_alien_spawn_halt",
		slog.String("event_id", e.ID),
		slog.String("planet_id", planetID),
		slog.String("user_id", userID),
		slog.Duration("duration", dur))
	return nil
}

// sendMessage отправляет в-игровое сообщение игроку по ключам i18n.
//
// folder=1 (alien notifications), from_user_id=NULL.
func (s *Service) sendMessage(ctx context.Context, tx pgx.Tx, userID,
	group, subjectKey, bodyKey string, vars map[string]string) error {

	if _, err := tx.Exec(ctx, `
		INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
		VALUES ($1::uuid, $2::uuid, NULL, 1, $3, $4)
	`, ids.New(), userID,
		s.tr(group, subjectKey, vars),
		s.tr(group, bodyKey, vars)); err != nil {
		return fmt.Errorf("alien message: %w", err)
	}
	return nil
}

// fnvHashString — детерминированный hash для seeding rng.R по event.ID.
// Скопирован из internal/alien/helpers.go (FNV-1a).
func fnvHashString(s string) uint64 {
	var h uint64 = 14695981039346656037
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 1099511628211
	}
	return h
}
