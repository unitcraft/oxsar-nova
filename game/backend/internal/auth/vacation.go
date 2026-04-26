package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// VacationMinInterval — минимальный интервал между отпусками
	// (LAST_TIME_ON_VACATION_DISABLE = 20 дней, Functions.inc.php:548).
	VacationMinInterval = 20 * 24 * time.Hour

	// VacationMinDuration — минимальная продолжительность отпуска
	// (spec §18.8, umode_min = now+48h). Раньше выйти нельзя.
	VacationMinDuration = 48 * time.Hour

	// VacationAutoDisable — если отпуск длится дольше и last_seen
	// старше этого срока, воркер автоматически выключает его
	// (VACATION_DISABLE_TIME = 30 дней, Preferences.class.php).
	VacationAutoDisable = 30 * 24 * time.Hour
)

var (
	ErrVacationAlreadyActive  = errors.New("vacation: already active")
	ErrVacationNotActive      = errors.New("vacation: not active")
	ErrVacationIntervalNotMet = errors.New("vacation: min interval not met (20 days)")
	ErrVacationTooEarly       = errors.New("vacation: minimum 48h not passed")
	ErrVacationBlocked        = errors.New("vacation: blocked by pending events (build/fleet/research)")
)

// VacationService управляет режимом отпуска игрока.
type VacationService struct {
	db *pgxpool.Pool
}

func NewVacationService(db *pgxpool.Pool) *VacationService {
	return &VacationService{db: db}
}

// vacationBlockingKinds — kinds событий, которые блокируют включение
// отпуска. Соответствует VACATION_BLOCKING_EVENTS в legacy
// (Preferences.class.php). Если у игрока есть хотя бы одно wait-событие
// из этого списка — включить отпуск нельзя.
var vacationBlockingKinds = []int{
	1, 2, 3, 4, 5, // BUILD_CONSTRUCTION, DEMOLISH, RESEARCH, BUILD_FLEET, BUILD_DEFENSE
	6, 7, 8, 9, 10, 11, 12, // POSITION, TRANSPORT, COLONIZE, RECYCLING, ATTACK_SINGLE, SPY, ATTACK_ALLIANCE
	13, 14, 15, 16, 17, // HALT, MOON_DESTRUCTION, EXPEDITION, ROCKET_ATTACK, HOLDING
	28,     // STARGATE_TRANSPORT
	50, 51, // REPAIR, DISASSEMBLE
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
	// Проверяем, нет ли активных событий, блокирующих отпуск.
	var pendingCount int
	err = s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM events
		WHERE user_id = $1 AND state = 'wait' AND kind = ANY($2)
	`, userID, vacationBlockingKinds).Scan(&pendingCount)
	if err != nil {
		return fmt.Errorf("check blocking events: %w", err)
	}
	if pendingCount > 0 {
		return ErrVacationBlocked
	}
	_, err = s.db.Exec(ctx,
		`UPDATE users SET vacation_since=NOW() WHERE id=$1`, userID)
	return err
}

// UnsetVacation выключает режим отпуска для userID. Можно выйти
// только если прошло ≥48 часов с включения (spec §18.8).
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
	if time.Since(*vacSince) < VacationMinDuration {
		return ErrVacationTooEarly
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
