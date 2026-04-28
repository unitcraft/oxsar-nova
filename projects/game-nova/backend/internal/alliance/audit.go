// Audit-log альянса (план 67 Ф.2, U-013).
//
// Запись пишется на каждое значимое действие: создание альянса,
// изменение описаний, передача лидерства (Ф.3), изгнание,
// смена ранга, изменение дипстатуса, и т.д.
//
// Дизайн скопирован с admin_audit_log (план 14): action — символическое
// имя в snake_case, target_kind/target_id — на кого/что направлено,
// payload — JSONB с дополнительной информацией.
//
// По R10 nova однобазная (universe = отдельная БД), поэтому universe_id
// не пишется. Если nova станет мультитенантной — добавится отдельной миграцией.

package alliance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// Стандартные имена action (snake_case по R1). Используются и в коде,
// и в UI как ключи для перевода.
const (
	ActionAllianceCreated       = "alliance_created"
	ActionAllianceDisbanded     = "alliance_disbanded"
	ActionDescriptionChanged    = "description_changed"
	ActionMemberJoined          = "member_joined"
	ActionMemberLeft            = "member_left"
	ActionMemberKicked          = "member_kicked"
	ActionMemberRankAssigned    = "member_rank_assigned"
	ActionRankCreated           = "rank_created"
	ActionRankUpdated           = "rank_updated"
	ActionRankDeleted           = "rank_deleted"
	ActionApplicationApproved   = "application_approved"
	ActionApplicationRejected   = "application_rejected"
	ActionRelationProposed      = "relation_proposed"
	ActionRelationAccepted      = "relation_accepted"
	ActionRelationRejected      = "relation_rejected"
	ActionRelationCleared       = "relation_cleared"
	ActionLeadershipTransferred = "leadership_transferred"
	ActionOpenChanged           = "open_changed"
)

// Target kind — тип объекта действия.
const (
	TargetKindUser     = "user"
	TargetKindRank     = "rank"
	TargetKindRelation = "relation"
	TargetKindAlliance = "alliance"
)

// AuditEntry — запись для GET /api/alliances/{id}/audit.
type AuditEntry struct {
	ID         string          `json:"id"`
	AllianceID string          `json:"alliance_id"`
	ActorID    string          `json:"actor_id"` // "" если NULL (системное)
	ActorName  string          `json:"actor_name"`
	Action     string          `json:"action"`
	TargetKind string          `json:"target_kind"`
	TargetID   string          `json:"target_id"`
	Payload    json.RawMessage `json:"payload"`
	CreatedAt  time.Time       `json:"created_at"`
}

// writeAuditTx пишет запись в транзакции. Ошибки логируются и не
// пробрасываются — отказ аудита не должен ронять основную операцию.
//
// payload — произвольный сериализуемый объект; nil → '{}'.
func writeAuditTx(ctx context.Context, tx pgx.Tx, allianceID, actorID, action, targetKind, targetID string, payload any) {
	raw := mustMarshalAuditPayload(ctx, action, payload)
	var actor any
	if actorID != "" {
		actor = actorID
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO alliance_audit_log
			(alliance_id, actor_id, action, target_kind, target_id, payload)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, allianceID, actor, action, targetKind, targetID, raw); err != nil {
		slog.WarnContext(ctx, "alliance_audit_insert_failed",
			slog.String("alliance_id", allianceID),
			slog.String("action", action),
			slog.String("err", err.Error()))
	}
}

func mustMarshalAuditPayload(ctx context.Context, action string, payload any) []byte {
	if payload == nil {
		return []byte(`{}`)
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		slog.WarnContext(ctx, "alliance_audit_marshal_failed",
			slog.String("action", action),
			slog.String("err", err.Error()))
		return []byte(`{}`)
	}
	return raw
}

// ListAudit GET /api/alliances/{id}/audit?action=&actor_id=&limit=&offset=
//
// Доступ — любой член альянса. Лог внутригрупповой, без фильтра по
// permission (UI может прятать, но read разрешён всем участникам).
func (s *Service) ListAudit(ctx context.Context, requesterID, allianceID string, filters AuditFilters) ([]AuditEntry, error) {
	var memAlliance *string
	if err := s.db.Pool().QueryRow(ctx,
		`SELECT alliance_id FROM alliance_members WHERE user_id=$1`, requesterID).Scan(&memAlliance); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotMember
		}
		return nil, fmt.Errorf("audit list: check member: %w", err)
	}
	if memAlliance == nil || *memAlliance != allianceID {
		return nil, ErrNotMember
	}

	limit := filters.Limit
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	offset := filters.Offset
	if offset < 0 {
		offset = 0
	}

	q := `
		SELECT a.id, a.alliance_id, COALESCE(a.actor_id::text, '') AS actor_id,
		       COALESCE(u.username, '') AS actor_name,
		       a.action, a.target_kind, a.target_id, a.payload, a.created_at
		FROM alliance_audit_log a
		LEFT JOIN users u ON u.id = a.actor_id
		WHERE a.alliance_id = $1`
	args := []any{allianceID}
	addFilter := func(sql string, v any) {
		args = append(args, v)
		q += " AND " + strings.Replace(sql, "?", "$"+strconv.Itoa(len(args)), 1)
	}
	if filters.Action != "" {
		addFilter("a.action = ?", filters.Action)
	}
	if filters.ActorID != "" {
		addFilter("a.actor_id = ?", filters.ActorID)
	}
	q += " ORDER BY a.created_at DESC LIMIT $" + strconv.Itoa(len(args)+1) +
		" OFFSET $" + strconv.Itoa(len(args)+2)
	args = append(args, limit, offset)

	rows, err := s.db.Pool().Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("audit list: query: %w", err)
	}
	defer rows.Close()
	out := []AuditEntry{}
	for rows.Next() {
		var e AuditEntry
		var raw []byte
		if err := rows.Scan(&e.ID, &e.AllianceID, &e.ActorID, &e.ActorName,
			&e.Action, &e.TargetKind, &e.TargetID, &raw, &e.CreatedAt); err != nil {
			return nil, err
		}
		if len(raw) > 0 {
			e.Payload = raw
		} else {
			e.Payload = json.RawMessage("{}")
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// AuditFilters — параметры фильтрации для ListAudit.
type AuditFilters struct {
	Action  string
	ActorID string
	Limit   int
	Offset  int
}
