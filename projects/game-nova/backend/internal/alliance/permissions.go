// Permissions — гранулярные права рангов альянса (план 67 Ф.2).
//
// Модель: таблица alliance_ranks (миграция 0074) хранит permissions
// в JSONB-колонке с булевыми ключами в snake_case (R1). Член альянса
// может ссылаться на ранг через alliance_members.rank_id (FK,
// nullable). Если rank_id = NULL — fallback на builtin-роли:
//   - alliance_members.rank='owner'  → все права true,
//   - alliance_members.rank='member' → все права false.
//
// Проверка прав вызывается явными guard-helpers из service.go;
// HTTP-middleware не используется, потому что для большинства
// действий нужно сначала прочитать членство (alliance_id из
// alliance_members), а это происходит в самом сервисе.
//
// R0: добавление новой функциональности — старые члены без rank_id
// продолжают работать через builtin-fallback.

package alliance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/repo"
)

// Permission — символическое имя права (snake_case по R1).
type Permission string

const (
	PermInvite             Permission = "can_invite"
	PermKick               Permission = "can_kick"
	PermSendGlobalMail     Permission = "can_send_global_mail"
	PermManageDiplomacy    Permission = "can_manage_diplomacy"
	PermChangeDescription  Permission = "can_change_description"
	PermProposeRelations   Permission = "can_propose_relations"
	PermManageRanks        Permission = "can_manage_ranks"
)

// AllPermissions — упорядоченный список всех ключей. Используется
// при создании рангов и в API/UI для отображения чек-боксов.
var AllPermissions = []Permission{
	PermInvite,
	PermKick,
	PermSendGlobalMail,
	PermManageDiplomacy,
	PermChangeDescription,
	PermProposeRelations,
	PermManageRanks,
}

// ErrForbidden — у роли нет требуемого права.
var ErrForbidden = errors.New("alliance: permission denied")

// Membership — снимок членства пользователя в альянсе для проверки прав.
type Membership struct {
	UserID     string
	AllianceID string
	BuiltinRank string // 'owner' | 'member'
	RankID     *string
}

// LoadMembership читает alliance_members + alliance_id для пользователя.
// Возвращает nil-результат если пользователь не в альянсе или не в указанном.
// Если allianceID="" — возвращает любой альянс пользователя.
func LoadMembership(ctx context.Context, q pgx.Tx, userID, allianceID string) (*Membership, error) {
	m := &Membership{UserID: userID}
	var (
		row pgx.Row
	)
	if allianceID == "" {
		row = q.QueryRow(ctx, `
			SELECT alliance_id, rank, rank_id
			FROM alliance_members WHERE user_id=$1
		`, userID)
	} else {
		row = q.QueryRow(ctx, `
			SELECT alliance_id, rank, rank_id
			FROM alliance_members WHERE user_id=$1 AND alliance_id=$2
		`, userID, allianceID)
	}
	if err := row.Scan(&m.AllianceID, &m.BuiltinRank, &m.RankID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("load membership: %w", err)
	}
	return m, nil
}

// Has проверяет наличие permission у membership.
//
// Логика:
//   - BuiltinRank == "owner"      → всегда true (owner имеет все права).
//   - RankID != nil               → читаем permissions JSONB у соотв. ранга.
//   - иначе (member без rank_id)  → false (старая модель: рядовой не имеет прав).
//
// Отсутствие ключа в JSONB трактуется как false (R1).
func Has(ctx context.Context, q pgx.Tx, m *Membership, perm Permission) (bool, error) {
	if m == nil {
		return false, nil
	}
	if m.BuiltinRank == "owner" {
		return true, nil
	}
	if m.RankID == nil {
		return false, nil
	}
	var raw []byte
	err := q.QueryRow(ctx, `SELECT permissions FROM alliance_ranks WHERE id=$1`, *m.RankID).Scan(&raw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("read rank permissions: %w", err)
	}
	return permissionInJSON(raw, perm), nil
}

// HasViaPool — версия для случаев без активной транзакции.
func HasViaPool(ctx context.Context, db repo.Exec, m *Membership, perm Permission) (bool, error) {
	if m == nil {
		return false, nil
	}
	if m.BuiltinRank == "owner" {
		return true, nil
	}
	if m.RankID == nil {
		return false, nil
	}
	var raw []byte
	err := db.Pool().QueryRow(ctx, `SELECT permissions FROM alliance_ranks WHERE id=$1`, *m.RankID).Scan(&raw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("read rank permissions: %w", err)
	}
	return permissionInJSON(raw, perm), nil
}

// permissionInJSON — true если в raw JSONB ключ perm установлен в true.
// Невалидный JSON или отсутствие ключа → false.
func permissionInJSON(raw []byte, perm Permission) bool {
	if len(raw) == 0 {
		return false
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return false
	}
	v, ok := m[string(perm)]
	if !ok {
		return false
	}
	b, ok := v.(bool)
	return ok && b
}

// IsValidPermission — true если key — известное право.
func IsValidPermission(key string) bool {
	for _, p := range AllPermissions {
		if string(p) == key {
			return true
		}
	}
	return false
}
