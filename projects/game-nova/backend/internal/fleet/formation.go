// Package fleet — formation: post-factum ACS-группы (план 72.1.48,
// legacy `Mission.class.php::formation` action).
//
// Семантика гибрида (отличается от legacy 1:1):
//   - Legacy мутирует event'у kind: SINGLE→ALLIANCE. Nova не мутирует
//     events (immutable), вместо этого создаёт `acs_groups` запись для
//     уже летящего ATTACK_SINGLE флота. Сам event остаётся, но fleets
//     получает acs_group_id и при handler'е боя группируется как ACS.
//   - Invitation: legacy `formation_invitation`. Nova `acs_invitations`
//     с PK (acs_group_id, user_id), accepted_at NULL = pending.
//     Приглашённый видит инвайт через GET /api/acs/invitations и
//     может присоединиться через стандартный Send с acs_group_id.
//   - Relation gating: invite разрешён только если у пригласившего и
//     приглашённого есть `alliance_relationships` с relation IN
//     ('ally','nap'), т.е. дип. отношения между их альянсами.

package fleet

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/event"
	"oxsar/game-nova/pkg/ids"
)

var (
	ErrFleetNotPromotable = errors.New("formation: fleet not in attack_single state")
	ErrAlreadyPromoted    = errors.New("formation: fleet already promoted to ACS")
	ErrNotLeader          = errors.New("formation: only ACS leader can invite")
	ErrInviteeNotFound    = errors.New("formation: invitee user not found")
	ErrNoRelation         = errors.New("formation: no diplomatic relation (ally/nap) with invitee's alliance")
	ErrSelfInvite         = errors.New("formation: cannot invite yourself")
	ErrInvalidName        = errors.New("formation: name must be 1..128 chars")
	ErrInvitationNotFound = errors.New("formation: invitation not found")
)

// ACSGroup — публичный DTO для UI.
type ACSGroup struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	LeaderUserID  string `json:"leader_user_id"`
	LeaderFleetID string `json:"leader_fleet_id"`
	CreatedAt     string `json:"created_at"`
}

// Invitation — pending или accepted приглашение.
type Invitation struct {
	ACSGroupID    string  `json:"acs_group_id"`
	GroupName     string  `json:"group_name"`
	LeaderName    string  `json:"leader_name"`
	InvitedBy     string  `json:"invited_by"`
	InvitedAt     string  `json:"invited_at"`
	AcceptedAt    *string `json:"accepted_at,omitempty"`
}

// PromoteToACS — конверсия ATTACK_SINGLE флота в ACS-группу.
// Создаёт `acs_groups` запись и проставляет fleets.acs_group_id.
// Возвращает группу для UI.
func (s *TransportService) PromoteToACS(ctx context.Context, userID, fleetID, name string) (ACSGroup, error) {
	name = strings.TrimSpace(name)
	if len(name) == 0 || len(name) > 128 {
		return ACSGroup{}, ErrInvalidName
	}
	var out ACSGroup
	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var (
			ownerID                          string
			currentACSGroupID                *string
			fleetMission                     int
			dstGalaxy, dstSystem, dstPos     int
			dstIsMoon                        bool
			arriveAt                         interface{} // timestamptz
		)
		err := tx.QueryRow(ctx, `
			SELECT owner_user_id, mission, acs_group_id,
			       dst_galaxy, dst_system, dst_position, dst_is_moon, arrive_at
			FROM fleets WHERE id=$1 FOR UPDATE
		`, fleetID).Scan(&ownerID, &fleetMission, &currentACSGroupID,
			&dstGalaxy, &dstSystem, &dstPos, &dstIsMoon, &arriveAt)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrFleetNotFound
			}
			return fmt.Errorf("formation: select fleet: %w", err)
		}
		if ownerID != userID {
			return ErrPlanetOwnership
		}
		// Только ATTACK_SINGLE и destroy-варианты можно promote.
		// ACS уже создан — нечего promote'ить.
		if currentACSGroupID != nil && *currentACSGroupID != "" {
			return ErrAlreadyPromoted
		}
		k := event.Kind(fleetMission)
		if k != event.KindAttackSingle && k != event.KindAttackDestroyBuilding && k != event.KindAttackDestroyMoon {
			return ErrFleetNotPromotable
		}
		// Найдём event-id флота (ATTACK_SINGLE с этим fleet_id в payload).
		// Это бест-эффорт — leader_event_id записываем для UI.
		var leaderEventID *string
		_ = tx.QueryRow(ctx, `
			SELECT id FROM events
			WHERE kind=$1 AND payload->>'fleet_id'=$2 AND state='wait'
			LIMIT 1
		`, fleetMission, fleetID).Scan(&leaderEventID)

		groupID := ids.New()
		// План 72.1.48: acs_groups (см. миграцию 0021_acs.sql) требует
		// target+arrive_at, миграция 0094 добавляет name+leader.
		if _, err := tx.Exec(ctx, `
			INSERT INTO acs_groups (id, target_galaxy, target_system, target_position,
			                        target_is_moon, arrive_at,
			                        name, leader_user_id, leader_fleet_id, leader_event_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, groupID, dstGalaxy, dstSystem, dstPos, dstIsMoon, arriveAt,
			name, userID, fleetID, leaderEventID); err != nil {
			return fmt.Errorf("insert acs_groups: %w", err)
		}
		// Обновляем fleets.acs_group_id и mission на ACS-вариант.
		newMission := int(event.KindAttackAlliance)
		switch k {
		case event.KindAttackDestroyBuilding:
			newMission = int(event.KindAttackAllianceDestroyBuilding)
		case event.KindAttackDestroyMoon:
			newMission = int(event.KindAttackAllianceDestroyMoon)
		}
		if _, err := tx.Exec(ctx,
			`UPDATE fleets SET acs_group_id=$1, mission=$2 WHERE id=$3`,
			groupID, newMission, fleetID); err != nil {
			return fmt.Errorf("update fleet: %w", err)
		}
		// Также добавляем leader-userId в acs_invitations как accepted —
		// для единообразия списков (он сам участник).
		if _, err := tx.Exec(ctx, `
			INSERT INTO acs_invitations (acs_group_id, user_id, invited_by, accepted_at)
			VALUES ($1, $2, $2, now())
			ON CONFLICT DO NOTHING
		`, groupID, userID); err != nil {
			return fmt.Errorf("insert leader invitation: %w", err)
		}
		var createdAt string
		if err := tx.QueryRow(ctx,
			`SELECT created_at::text FROM acs_groups WHERE id=$1`, groupID,
		).Scan(&createdAt); err != nil {
			return fmt.Errorf("read created_at: %w", err)
		}
		out = ACSGroup{
			ID: groupID, Name: name, LeaderUserID: userID,
			LeaderFleetID: fleetID, CreatedAt: createdAt,
		}
		return nil
	})
	return out, err
}

// InviteToFormation — пригласить юзера в ACS-группу. Только лидер
// может приглашать. Требуется ally/nap-relation между альянсами.
// Если у пригласителя или приглашённого нет альянса → запрещено
// (legacy: NS::getRelation(...) возвращает empty → false).
func (s *TransportService) InviteToFormation(ctx context.Context, leaderID, acsGroupID, inviteeUsername string) error {
	inviteeUsername = strings.TrimSpace(inviteeUsername)
	if inviteeUsername == "" {
		return ErrInviteeNotFound
	}
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// 1. Group exists и лидер — мы.
		var groupLeader string
		err := tx.QueryRow(ctx,
			`SELECT leader_user_id FROM acs_groups WHERE id=$1`, acsGroupID,
		).Scan(&groupLeader)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrInvitationNotFound
			}
			return fmt.Errorf("formation: select group: %w", err)
		}
		if groupLeader != leaderID {
			return ErrNotLeader
		}
		// 2. Найти приглашаемого.
		var inviteeID string
		err = tx.QueryRow(ctx,
			`SELECT id FROM users WHERE LOWER(username)=LOWER($1) AND deleted_at IS NULL`,
			inviteeUsername,
		).Scan(&inviteeID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrInviteeNotFound
			}
			return fmt.Errorf("formation: select invitee: %w", err)
		}
		if inviteeID == leaderID {
			return ErrSelfInvite
		}
		// 3. Relation check: оба в альянсах + есть ally/nap.
		var hasRelation bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM alliance_members lm
				JOIN alliance_members im ON im.user_id = $2
				JOIN alliance_relationships ar
				  ON ar.alliance_id = lm.alliance_id
				 AND ar.target_alliance_id = im.alliance_id
				WHERE lm.user_id = $1
				  AND ar.relation IN ('ally','nap')
			)
		`, leaderID, inviteeID).Scan(&hasRelation); err != nil {
			return fmt.Errorf("formation: relation check: %w", err)
		}
		if !hasRelation {
			return ErrNoRelation
		}
		// 4. Insert invitation (idempotent: ON CONFLICT DO NOTHING).
		if _, err := tx.Exec(ctx, `
			INSERT INTO acs_invitations (acs_group_id, user_id, invited_by)
			VALUES ($1, $2, $3)
			ON CONFLICT (acs_group_id, user_id) DO NOTHING
		`, acsGroupID, inviteeID, leaderID); err != nil {
			return fmt.Errorf("formation: insert invitation: %w", err)
		}
		return nil
	})
}

// ListInvitations — список приглашений для текущего юзера.
// Включает pending (accepted_at IS NULL) и accepted; UI фильтрует.
func (s *TransportService) ListInvitations(ctx context.Context, userID string) ([]Invitation, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT i.acs_group_id, g.name,
		       COALESCE(lu.username, ''), COALESCE(iu.username, ''),
		       i.invited_at::text, i.accepted_at::text
		  FROM acs_invitations i
		  JOIN acs_groups g ON g.id = i.acs_group_id
		  LEFT JOIN users lu ON lu.id = g.leader_user_id
		  LEFT JOIN users iu ON iu.id = i.invited_by
		 WHERE i.user_id = $1
		 ORDER BY i.invited_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("formation: list invitations: %w", err)
	}
	defer rows.Close()
	out := make([]Invitation, 0)
	for rows.Next() {
		var inv Invitation
		var acceptedAt *string
		if err := rows.Scan(&inv.ACSGroupID, &inv.GroupName,
			&inv.LeaderName, &inv.InvitedBy, &inv.InvitedAt, &acceptedAt,
		); err != nil {
			return nil, err
		}
		inv.AcceptedAt = acceptedAt
		out = append(out, inv)
	}
	return out, rows.Err()
}

// AcceptInvitation — принять приглашение. После accept юзер может
// `Send` с указанием acs_group_id (resolveACSGroup проверит, что у
// него есть accepted-row).
func (s *TransportService) AcceptInvitation(ctx context.Context, userID, acsGroupID string) error {
	res, err := s.db.Pool().Exec(ctx, `
		UPDATE acs_invitations
		   SET accepted_at = now()
		 WHERE acs_group_id = $1 AND user_id = $2 AND accepted_at IS NULL
	`, acsGroupID, userID)
	if err != nil {
		return fmt.Errorf("formation: accept: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrInvitationNotFound
	}
	return nil
}
