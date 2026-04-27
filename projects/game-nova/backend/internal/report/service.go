// Package report — жалобы пользователей на UGC-нарушения (план 46 Ф.3, 149-ФЗ).
//
// Поток: игрок жмёт "Пожаловаться" в UI → POST /api/reports → запись
// со status='new' → модератор смотрит в админке → POST
// /api/admin/reports/{id}/resolve переводит в 'resolved' или 'rejected'
// с пометкой о принятом действии.
//
// Само действие (warn/mute/ban/rename) выполняется отдельным админ-API
// (план 14). Здесь только трекинг жалобы.
package report

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/repo"
	"oxsar/game-nova/pkg/ids"
)

const (
	maxCommentLen = 1000
	maxReasonLen  = 64
)

// Допустимые target_type. Должны совпадать с CHECK в миграции 0069.
var allowedTargetTypes = map[string]struct{}{
	"user":     {},
	"alliance": {},
	"chat_msg": {},
	"planet":   {},
}

var (
	ErrInvalidTarget = errors.New("report: invalid target_type")
	ErrEmptyReason   = errors.New("report: reason required")
	ErrTooLong       = errors.New("report: comment too long")
	ErrNotFound      = errors.New("report: not found")
	ErrAlreadyClosed = errors.New("report: already resolved")
	ErrSelfReport    = errors.New("report: cannot report yourself")
)

type Service struct {
	db repo.Exec
}

func NewService(db repo.Exec) *Service { return &Service{db: db} }

// Report — запись для UI.
type Report struct {
	ID             string     `json:"id"`
	ReporterID     string     `json:"reporter_id"`
	ReporterName   string     `json:"reporter_name,omitempty"`
	TargetType     string     `json:"target_type"`
	TargetID       string     `json:"target_id"`
	Reason         string     `json:"reason"`
	Comment        string     `json:"comment"`
	Status         string     `json:"status"`
	ResolvedBy     *string    `json:"resolved_by,omitempty"`
	ResolverName   string     `json:"resolver_name,omitempty"`
	ResolutionNote string     `json:"resolution_note,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
}

// CreateInput — входные данные жалобы.
type CreateInput struct {
	TargetType string
	TargetID   string
	Reason     string
	Comment    string
}

// Create регистрирует новую жалобу. reporter_id — текущий пользователь.
func (s *Service) Create(ctx context.Context, reporterID string, in CreateInput) (Report, error) {
	if _, ok := allowedTargetTypes[in.TargetType]; !ok {
		return Report{}, ErrInvalidTarget
	}
	in.TargetID = strings.TrimSpace(in.TargetID)
	in.Reason = strings.TrimSpace(in.Reason)
	in.Comment = strings.TrimSpace(in.Comment)
	if in.TargetID == "" {
		return Report{}, ErrInvalidTarget
	}
	if in.Reason == "" || utf8.RuneCountInString(in.Reason) > maxReasonLen {
		return Report{}, ErrEmptyReason
	}
	if utf8.RuneCountInString(in.Comment) > maxCommentLen {
		return Report{}, ErrTooLong
	}
	// Защита от самовыстрелов: жалоба на себя/на свой альянс — запрет
	// на уровне сервиса (target_type=user → сравнение target_id напрямую).
	if in.TargetType == "user" && in.TargetID == reporterID {
		return Report{}, ErrSelfReport
	}

	id := ids.New()
	now := time.Now().UTC()
	if _, err := s.db.Pool().Exec(ctx, `
		INSERT INTO user_reports
			(id, reporter_id, target_type, target_id, reason, comment, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'new', $7)
	`, id, reporterID, in.TargetType, in.TargetID, in.Reason, in.Comment, now); err != nil {
		return Report{}, fmt.Errorf("insert report: %w", err)
	}
	return Report{
		ID:         id,
		ReporterID: reporterID,
		TargetType: in.TargetType,
		TargetID:   in.TargetID,
		Reason:     in.Reason,
		Comment:    in.Comment,
		Status:     "new",
		CreatedAt:  now,
	}, nil
}

// List возвращает жалобы для админки. status — фильтр (пустая строка =
// все). limit ограничен 200 (если 0 — 50 по умолчанию).
func (s *Service) List(ctx context.Context, status string, limit int) ([]Report, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	q := `
		SELECT r.id, r.reporter_id, COALESCE(ur.username,''),
		       r.target_type, r.target_id, r.reason, r.comment, r.status,
		       r.resolved_by, COALESCE(uv.username,''),
		       COALESCE(r.resolution_note,''), r.created_at, r.resolved_at
		FROM user_reports r
		LEFT JOIN users ur ON ur.id = r.reporter_id
		LEFT JOIN users uv ON uv.id = r.resolved_by
	`
	args := []any{}
	if status != "" {
		q += " WHERE r.status = $1"
		args = append(args, status)
	}
	q += " ORDER BY r.created_at DESC LIMIT $" + fmt.Sprint(len(args)+1)
	args = append(args, limit)

	rows, err := s.db.Pool().Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query reports: %w", err)
	}
	defer rows.Close()
	var out []Report
	for rows.Next() {
		var r Report
		var resolvedBy *string
		if err := rows.Scan(
			&r.ID, &r.ReporterID, &r.ReporterName,
			&r.TargetType, &r.TargetID, &r.Reason, &r.Comment, &r.Status,
			&resolvedBy, &r.ResolverName,
			&r.ResolutionNote, &r.CreatedAt, &r.ResolvedAt,
		); err != nil {
			return nil, err
		}
		if resolvedBy != nil {
			r.ResolvedBy = resolvedBy
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// Resolve переводит жалобу в 'resolved' или 'rejected'. moderator_id —
// id админа. note — текстовая пометка о принятом действии.
func (s *Service) Resolve(ctx context.Context, reportID, moderatorID, status, note string) error {
	if status != "resolved" && status != "rejected" {
		return fmt.Errorf("report: status must be resolved/rejected")
	}
	if utf8.RuneCountInString(note) > maxCommentLen {
		return ErrTooLong
	}
	tag, err := s.db.Pool().Exec(ctx, `
		UPDATE user_reports
		SET status=$1, resolved_by=$2, resolution_note=$3, resolved_at=now()
		WHERE id=$4 AND status='new'
	`, status, moderatorID, note, reportID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("update report: %w", err)
	}
	if tag.RowsAffected() == 0 {
		// Либо нет такого id, либо уже не 'new'. Различить без второго
		// запроса нельзя — возвращаем «ожидаемый» AlreadyClosed.
		return ErrAlreadyClosed
	}
	return nil
}
