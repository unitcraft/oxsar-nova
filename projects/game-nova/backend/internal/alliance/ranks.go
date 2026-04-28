// Ranks — гранулярные права рангов альянса (план 67 Ф.2, D-014, U-005).
//
// CRUD-операции над таблицей alliance_ranks (миграция 0074) и
// привязка членов к рангам через alliance_members.rank_id.
//
// Все CRUD-операции защищены permission can_manage_ranks (или builtin
// owner). Member.rank_id — nullable: если NULL, fallback-логика
// builtin "owner"/"member" (см. permissions.go).

package alliance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
)

// Rank — кастомный ранг альянса.
type Rank struct {
	ID          string          `json:"id"`
	AllianceID  string          `json:"alliance_id"`
	Name        string          `json:"name"`
	Position    int             `json:"position"`
	Permissions map[string]bool `json:"permissions"`
	CreatedAt   time.Time       `json:"created_at"`
}

// ListRanks GET /api/alliances/{id}/ranks. Доступ — любой член.
func (s *Service) ListRanks(ctx context.Context, requesterID, allianceID string) ([]Rank, error) {
	var memAlliance *string
	if err := s.db.Pool().QueryRow(ctx,
		`SELECT alliance_id FROM alliance_members WHERE user_id=$1`, requesterID).Scan(&memAlliance); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotMember
		}
		return nil, fmt.Errorf("ranks list: check member: %w", err)
	}
	if memAlliance == nil || *memAlliance != allianceID {
		return nil, ErrNotMember
	}

	rows, err := s.db.Pool().Query(ctx, `
		SELECT id, alliance_id, name, position, permissions, created_at
		FROM alliance_ranks WHERE alliance_id=$1
		ORDER BY position ASC, name ASC
	`, allianceID)
	if err != nil {
		return nil, fmt.Errorf("ranks list: query: %w", err)
	}
	defer rows.Close()
	out := []Rank{}
	for rows.Next() {
		var r Rank
		var raw []byte
		if err := rows.Scan(&r.ID, &r.AllianceID, &r.Name, &r.Position, &raw, &r.CreatedAt); err != nil {
			return nil, err
		}
		r.Permissions = decodePermissions(raw)
		out = append(out, r)
	}
	return out, rows.Err()
}

// CreateRank POST /api/alliances/{id}/ranks. Право: PermManageRanks.
//
// permissions — карта ключ→bool с подмножеством AllPermissions; неизвестные
// ключи → ErrInvalidPermission. Отсутствующие ключи трактуются как false.
func (s *Service) CreateRank(ctx context.Context, requesterID, allianceID, name string, position int, perms map[string]bool) (Rank, error) {
	out := Rank{}
	if err := validateRankName(name); err != nil {
		return out, err
	}
	if err := validatePermissions(perms); err != nil {
		return out, err
	}

	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		mem, err := LoadMembership(ctx, tx, requesterID, allianceID)
		if err != nil {
			return err
		}
		if mem == nil {
			return ErrNotMember
		}
		can, err := Has(ctx, tx, mem, PermManageRanks)
		if err != nil {
			return err
		}
		if !can {
			return ErrForbidden
		}

		raw, err := json.Marshal(perms)
		if err != nil {
			return fmt.Errorf("create rank: marshal perms: %w", err)
		}
		err = tx.QueryRow(ctx, `
			INSERT INTO alliance_ranks (alliance_id, name, position, permissions)
			VALUES ($1, $2, $3, $4)
			RETURNING id, created_at
		`, allianceID, name, position, raw).Scan(&out.ID, &out.CreatedAt)
		if err != nil {
			if isDupKey(err) {
				return ErrRankNameTaken
			}
			return fmt.Errorf("create rank: %w", err)
		}
		out.AllianceID = allianceID
		out.Name = name
		out.Position = position
		out.Permissions = perms
		writeAuditTx(ctx, tx, allianceID, requesterID, ActionRankCreated,
			TargetKindRank, out.ID, map[string]any{"name": name})
		return nil
	})
	return out, err
}

// UpdateRank PATCH /api/alliances/{id}/ranks/{rank_id}.
type UpdateRankInput struct {
	Name        *string
	Position    *int
	Permissions *map[string]bool
}

func (s *Service) UpdateRank(ctx context.Context, requesterID, allianceID, rankID string, in UpdateRankInput) error {
	if in.Name != nil {
		if err := validateRankName(*in.Name); err != nil {
			return err
		}
	}
	if in.Permissions != nil {
		if err := validatePermissions(*in.Permissions); err != nil {
			return err
		}
	}

	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		mem, err := LoadMembership(ctx, tx, requesterID, allianceID)
		if err != nil {
			return err
		}
		if mem == nil {
			return ErrNotMember
		}
		can, err := Has(ctx, tx, mem, PermManageRanks)
		if err != nil {
			return err
		}
		if !can {
			return ErrForbidden
		}

		// Удостоверимся, что ранг принадлежит этому альянсу.
		var ownerAlliance string
		if err := tx.QueryRow(ctx,
			`SELECT alliance_id FROM alliance_ranks WHERE id=$1`, rankID).Scan(&ownerAlliance); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrRankNotFound
			}
			return fmt.Errorf("update rank: read: %w", err)
		}
		if ownerAlliance != allianceID {
			return ErrRankNotFound
		}

		set := []string{}
		args := []any{rankID}
		if in.Name != nil {
			args = append(args, *in.Name)
			set = append(set, "name=$2")
		}
		if in.Position != nil {
			args = append(args, *in.Position)
			set = append(set, fmt.Sprintf("position=$%d", len(args)))
		}
		if in.Permissions != nil {
			raw, err := json.Marshal(*in.Permissions)
			if err != nil {
				return fmt.Errorf("update rank: marshal: %w", err)
			}
			args = append(args, raw)
			set = append(set, fmt.Sprintf("permissions=$%d", len(args)))
		}
		if len(set) == 0 {
			return nil
		}
		q := "UPDATE alliance_ranks SET " + joinComma(set) + " WHERE id=$1"
		if _, err := tx.Exec(ctx, q, args...); err != nil {
			if isDupKey(err) {
				return ErrRankNameTaken
			}
			return fmt.Errorf("update rank: exec: %w", err)
		}
		writeAuditTx(ctx, tx, allianceID, requesterID, ActionRankUpdated,
			TargetKindRank, rankID, nil)
		return nil
	})
}

// DeleteRank DELETE /api/alliances/{id}/ranks/{rank_id}. Право: PermManageRanks.
//
// alliance_members.rank_id с FK ON DELETE SET NULL — члены, у которых был
// этот ранг, откатятся к builtin-роли (rank='member', нет прав).
func (s *Service) DeleteRank(ctx context.Context, requesterID, allianceID, rankID string) error {
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		mem, err := LoadMembership(ctx, tx, requesterID, allianceID)
		if err != nil {
			return err
		}
		if mem == nil {
			return ErrNotMember
		}
		can, err := Has(ctx, tx, mem, PermManageRanks)
		if err != nil {
			return err
		}
		if !can {
			return ErrForbidden
		}

		tag, err := tx.Exec(ctx,
			`DELETE FROM alliance_ranks WHERE id=$1 AND alliance_id=$2`, rankID, allianceID)
		if err != nil {
			return fmt.Errorf("delete rank: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return ErrRankNotFound
		}
		writeAuditTx(ctx, tx, allianceID, requesterID, ActionRankDeleted,
			TargetKindRank, rankID, nil)
		return nil
	})
}

// AssignMemberRank PATCH /api/alliances/{id}/members/{user_id}/rank-id
// Body: {"rank_id": "..." | null}
//
// Привязывает кастомный ранг к участнику. rankID="" — отвязать
// (вернуться к builtin). Право: PermManageRanks.
//
// Owner'у привязать ранг можно (это не меняет его builtin='owner' →
// все права остаются true). Это корректно: PermManageRanks не значит
// «понизить владельца».
func (s *Service) AssignMemberRank(ctx context.Context, requesterID, allianceID, memberUserID, rankID string) error {
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		mem, err := LoadMembership(ctx, tx, requesterID, allianceID)
		if err != nil {
			return err
		}
		if mem == nil {
			return ErrNotMember
		}
		can, err := Has(ctx, tx, mem, PermManageRanks)
		if err != nil {
			return err
		}
		if !can {
			return ErrForbidden
		}

		var rankArg any
		if rankID != "" {
			// Проверим что ранг этого альянса.
			var ownerAlliance string
			if err := tx.QueryRow(ctx,
				`SELECT alliance_id FROM alliance_ranks WHERE id=$1`, rankID).Scan(&ownerAlliance); err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return ErrRankNotFound
				}
				return fmt.Errorf("assign rank: read: %w", err)
			}
			if ownerAlliance != allianceID {
				return ErrRankNotFound
			}
			rankArg = rankID
		}

		tag, err := tx.Exec(ctx, `
			UPDATE alliance_members SET rank_id=$1 WHERE alliance_id=$2 AND user_id=$3
		`, rankArg, allianceID, memberUserID)
		if err != nil {
			return fmt.Errorf("assign rank: exec: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return ErrMemberNotFound
		}

		writeAuditTx(ctx, tx, allianceID, requesterID, ActionMemberRankAssigned,
			TargetKindUser, memberUserID, map[string]any{"rank_id": rankID})
		return nil
	})
}

func validateRankName(name string) error {
	n := utf8.RuneCountInString(name)
	if n < 1 || n > 32 {
		return ErrRankNameInvalid
	}
	return nil
}

func validatePermissions(perms map[string]bool) error {
	for k := range perms {
		if !IsValidPermission(k) {
			return fmt.Errorf("%w: %q", ErrInvalidPermission, k)
		}
	}
	return nil
}

func decodePermissions(raw []byte) map[string]bool {
	out := map[string]bool{}
	if len(raw) == 0 {
		return out
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return out
	}
	for k, v := range m {
		if b, ok := v.(bool); ok {
			out[k] = b
		}
	}
	return out
}

func joinComma(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for i := 1; i < len(parts); i++ {
		out += ", " + parts[i]
	}
	return out
}
