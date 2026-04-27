// Package report — пользовательские жалобы (user_reports), 149-ФЗ.
//
// План 56: ownership перенесён из game-nova в portal. Жалоба относится
// к глобальному identity-аккаунту, не к конкретной вселенной — поэтому
// единый реестр на portal-backend, доступный из всех вселенных
// (game-nova, game-origin, будущие) через POST /api/reports.
//
// Поток: игрок жмёт «Пожаловаться» в UI → POST /api/reports → запись
// со status='new' → модератор смотрит в admin-frontend (через admin-bff,
// см. план 53) → POST /api/admin/reports/{id}/resolve переводит в
// 'resolved' или 'rejected' с пометкой о принятом действии.
//
// Сама санкция (warn/mute/ban/rename) выполняется отдельным admin-API
// (план 14 / identity-admin). Здесь только трекинг жалобы.
package report

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maxCommentLen = 1000
	maxReasonLen  = 64
	defaultLimit  = 50
	maxLimit      = 200
)

// Допустимые target_type. Должны совпадать с CHECK в миграции 0002.
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

// Service — application-сервис жалоб. Хранилище — pgxpool.Pool
// (portal-БД, таблица user_reports, миграция 0002_user_reports.sql).
type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service { return &Service{pool: pool} }

// Report — DTO для UI/админки.
type Report struct {
	ID             string     `json:"id"`
	ReporterID     string     `json:"reporter_id"`
	TargetType     string     `json:"target_type"`
	TargetID       string     `json:"target_id"`
	Reason         string     `json:"reason"`
	Comment        string     `json:"comment"`
	Status         string     `json:"status"`
	ResolvedBy     *string    `json:"resolved_by,omitempty"`
	ResolutionNote string     `json:"resolution_note,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
}

// CreateInput — входные данные жалобы (нормализуются в Create).
type CreateInput struct {
	TargetType string
	TargetID   string
	Reason     string
	Comment    string
}

// Create регистрирует новую жалобу. reporterID — текущий пользователь
// (из JWT context). Защита от самовыстрелов: жалоба на self target_id
// отклоняется.
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
	if in.TargetType == "user" && in.TargetID == reporterID {
		return Report{}, ErrSelfReport
	}

	id := uuid.NewString()
	now := time.Now().UTC()
	if _, err := s.pool.Exec(ctx, `
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

// List возвращает жалобы для админки. status — фильтр по статусу
// (пустая строка = все). limit ограничен maxLimit; 0 → defaultLimit.
//
// В отличие от game-nova-версии портал НЕ делает JOIN на users —
// у portal-БД нет своей таблицы users. Reporter/resolver username
// при необходимости подтягиваются на стороне admin-bff/frontend
// через identity API.
func (s *Service) List(ctx context.Context, status string, limit int) ([]Report, error) {
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	q := `
		SELECT id, reporter_id, target_type, target_id, reason, comment, status,
		       resolved_by, COALESCE(resolution_note, ''), created_at, resolved_at
		FROM user_reports
	`
	args := []any{}
	if status != "" {
		q += " WHERE status = $1"
		args = append(args, status)
	}
	q += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", len(args)+1)
	args = append(args, limit)

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query reports: %w", err)
	}
	defer rows.Close()
	var out []Report
	for rows.Next() {
		var r Report
		var resolvedBy *string
		if err := rows.Scan(
			&r.ID, &r.ReporterID, &r.TargetType, &r.TargetID,
			&r.Reason, &r.Comment, &r.Status,
			&resolvedBy, &r.ResolutionNote, &r.CreatedAt, &r.ResolvedAt,
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

// Resolve переводит жалобу в 'resolved' или 'rejected'. moderatorID —
// id админа из JWT-context. note — текстовая пометка о принятом
// действии (например, «бан 7 дней»). Резолюция идемпотентна по
// результату — повторная резолюция уже закрытой жалобы возвращает
// ErrAlreadyClosed.
func (s *Service) Resolve(ctx context.Context, reportID, moderatorID, status, note string) error {
	if status != "resolved" && status != "rejected" {
		return fmt.Errorf("report: status must be resolved/rejected")
	}
	if utf8.RuneCountInString(note) > maxCommentLen {
		return ErrTooLong
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE user_reports
		SET status=$1, resolved_by=$2, resolution_note=$3, resolved_at=now()
		WHERE id=$4 AND status='new'
	`, status, moderatorID, note, reportID)
	if err != nil {
		return fmt.Errorf("update report: %w", err)
	}
	if tag.RowsAffected() == 0 {
		// Либо нет такого id, либо уже не 'new'. Без второго запроса
		// различить нельзя — возвращаем «ожидаемый» AlreadyClosed.
		return ErrAlreadyClosed
	}
	return nil
}
