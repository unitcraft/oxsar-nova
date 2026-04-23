package auth

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// VacationMinInterval — минимальный интервал между отпусками (LAST_TIME_ON_VACATION_DISABLE).
	VacationMinInterval = 20 * 24 * time.Hour
)

var (
	ErrVacationAlreadyActive  = errors.New("vacation: already active")
	ErrVacationNotActive      = errors.New("vacation: not active")
	ErrVacationIntervalNotMet = errors.New("vacation: min interval not met (20 days)")
)

// VacationService управляет режимом отпуска игрока.
type VacationService struct {
	db *pgxpool.Pool
}

func NewVacationService(db *pgxpool.Pool) *VacationService {
	return &VacationService{db: db}
}

// SetVacation включает режим отпуска для userID.
func (s *VacationService) SetVacation(ctx context.Context, userID string) error {
	var vacSince *time.Time
	var lastEnd *time.Time
	err := s.db.QueryRow(ctx,
		`SELECT vacation_since, vacation_last_end FROM users WHERE id=$1`,
		userID).Scan(&vacSince, &lastEnd)
	if err != nil {
		return err
	}
	if vacSince != nil {
		return ErrVacationAlreadyActive
	}
	if lastEnd != nil && time.Since(*lastEnd) < VacationMinInterval {
		return ErrVacationIntervalNotMet
	}
	_, err = s.db.Exec(ctx,
		`UPDATE users SET vacation_since=NOW() WHERE id=$1`, userID)
	return err
}

// UnsetVacation выключает режим отпуска для userID.
func (s *VacationService) UnsetVacation(ctx context.Context, userID string) error {
	var vacSince *time.Time
	err := s.db.QueryRow(ctx,
		`SELECT vacation_since FROM users WHERE id=$1`,
		userID).Scan(&vacSince)
	if err != nil {
		return err
	}
	if vacSince == nil {
		return ErrVacationNotActive
	}
	_, err = s.db.Exec(ctx,
		`UPDATE users SET vacation_since=NULL, vacation_last_end=NOW() WHERE id=$1`, userID)
	return err
}

// IsOnVacation возвращает true если пользователь userID сейчас в отпуске.
func IsOnVacation(ctx context.Context, db *pgxpool.Pool, userID string) (bool, error) {
	var vacSince *time.Time
	err := db.QueryRow(ctx,
		`SELECT vacation_since FROM users WHERE id=$1`, userID).Scan(&vacSince)
	if err != nil {
		return false, err
	}
	return vacSince != nil, nil
}
