// Package identitysvc — RBAC: динамическая модель ролей и permissions.
//
// План 52 Ф.2: identity-service владеет таблицами roles, permissions,
// role_permissions, user_roles, audit_role_changes. Все остальные сервисы
// читают роли/permissions ТОЛЬКО из JWT (локальная валидация через JWKS).
//
// API:
//   - ListRoles, GetRolePermissions
//   - GrantUserRole, RevokeUserRole
//   - ListUserRoles, GetUserPermissions (плоский set из всех ролей юзера)
//   - LogAuditChange (immutable history)
//
// Все mutating-операции пишут запись в audit_role_changes.

package identitysvc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Role — справочная запись из таблицы roles.
type Role struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
}

// Permission — справочная запись из таблицы permissions.
type Permission struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// UserRoleAssignment — запись из user_roles + JOIN roles для имени.
type UserRoleAssignment struct {
	UserID    uuid.UUID  `json:"user_id"`
	RoleID    int        `json:"role_id"`
	RoleName  string     `json:"role_name"`
	GrantedBy *uuid.UUID `json:"granted_by"`
	GrantedAt time.Time  `json:"granted_at"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// AuditRoleChange — запись из audit_role_changes (immutable log).
type AuditRoleChange struct {
	ID        int64      `json:"id"`
	ActorID   uuid.UUID  `json:"actor_id"`
	TargetID  uuid.UUID  `json:"target_id"`
	RoleName  string     `json:"role_name"`
	Action    string     `json:"action"` // 'grant' | 'revoke'
	Reason    string     `json:"reason"`
	IPAddress *net.IP    `json:"ip_address"`
	UserAgent string     `json:"user_agent"`
	CreatedAt time.Time  `json:"created_at"`
}

// RBACService — операции над ролями и permissions.
type RBACService struct {
	pool *pgxpool.Pool
}

// NewRBACService создаёт сервис поверх pgxpool.
func NewRBACService(pool *pgxpool.Pool) *RBACService {
	return &RBACService{pool: pool}
}

// Errors

var (
	ErrRoleNotFound       = errors.New("role not found")
	ErrUserNotFound       = errors.New("user not found")
	ErrSystemRoleReadOnly = errors.New("system role is read-only")
	ErrAlreadyGranted     = errors.New("role already granted")
	ErrNotGranted         = errors.New("role not currently granted")
)

// ListRoles возвращает все роли (system + custom).
func (s *RBACService) ListRoles(ctx context.Context) ([]Role, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, name, COALESCE(description,''), is_system, created_at
		   FROM roles ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("query roles: %w", err)
	}
	defer rows.Close()

	out := make([]Role, 0, 8)
	for rows.Next() {
		var r Role
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.IsSystem, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// GetRolePermissions возвращает permissions конкретной роли.
func (s *RBACService) GetRolePermissions(ctx context.Context, roleID int) ([]Permission, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT p.id, p.name, COALESCE(p.description,''), p.created_at
		  FROM permissions p
		  JOIN role_permissions rp ON rp.permission_id = p.id
		 WHERE rp.role_id = $1
		 ORDER BY p.name`, roleID)
	if err != nil {
		return nil, fmt.Errorf("query role permissions: %w", err)
	}
	defer rows.Close()

	out := make([]Permission, 0, 8)
	for rows.Next() {
		var p Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ListUserRoles возвращает текущие активные роли юзера (с учётом expires_at).
func (s *RBACService) ListUserRoles(ctx context.Context, userID uuid.UUID) ([]UserRoleAssignment, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT ur.user_id, ur.role_id, r.name, ur.granted_by, ur.granted_at, ur.expires_at
		  FROM user_roles ur
		  JOIN roles r ON r.id = ur.role_id
		 WHERE ur.user_id = $1
		   AND (ur.expires_at IS NULL OR ur.expires_at > now())
		 ORDER BY r.name`, userID)
	if err != nil {
		return nil, fmt.Errorf("query user roles: %w", err)
	}
	defer rows.Close()

	out := make([]UserRoleAssignment, 0, 4)
	for rows.Next() {
		var a UserRoleAssignment
		if err := rows.Scan(&a.UserID, &a.RoleID, &a.RoleName, &a.GrantedBy,
			&a.GrantedAt, &a.ExpiresAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// GetUserPermissions возвращает плоский набор permissions из всех ролей юзера.
// Дедупликация — на стороне SQL через DISTINCT.
func (s *RBACService) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT p.name
		  FROM permissions p
		  JOIN role_permissions rp ON rp.permission_id = p.id
		  JOIN user_roles ur       ON ur.role_id = rp.role_id
		 WHERE ur.user_id = $1
		   AND (ur.expires_at IS NULL OR ur.expires_at > now())
		 ORDER BY p.name`, userID)
	if err != nil {
		return nil, fmt.Errorf("query user permissions: %w", err)
	}
	defer rows.Close()

	out := make([]string, 0, 16)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

// GetUserRoleNames — лёгкий вариант ListUserRoles, возвращает только имена ролей.
// Используется для встраивания в JWT claims (roles[]).
func (s *RBACService) GetUserRoleNames(ctx context.Context, userID uuid.UUID) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT r.name
		  FROM user_roles ur
		  JOIN roles r ON r.id = ur.role_id
		 WHERE ur.user_id = $1
		   AND (ur.expires_at IS NULL OR ur.expires_at > now())
		 ORDER BY r.name`, userID)
	if err != nil {
		return nil, fmt.Errorf("query user role names: %w", err)
	}
	defer rows.Close()

	out := make([]string, 0, 4)
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// GrantOptions — опциональные параметры для GrantUserRole.
type GrantOptions struct {
	ExpiresAt *time.Time
	Reason    string
	IPAddress *net.IP
	UserAgent string
}

// GrantUserRole выдаёт юзеру роль. Транзакция: INSERT user_roles + INSERT
// audit_role_changes — атомарно.
//
// actorID — uuid того кто грантит (superadmin); reason обязателен (не должен
// быть пустой строкой, политика audit-trail).
func (s *RBACService) GrantUserRole(
	ctx context.Context,
	actorID uuid.UUID,
	targetID uuid.UUID,
	roleName string,
	opts GrantOptions,
) error {
	if opts.Reason == "" {
		return errors.New("reason is required for grant")
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var roleID int
	err = tx.QueryRow(ctx, `SELECT id FROM roles WHERE name = $1`, roleName).Scan(&roleID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrRoleNotFound
	}
	if err != nil {
		return fmt.Errorf("lookup role: %w", err)
	}

	// Проверка, что target user существует.
	var dummy uuid.UUID
	err = tx.QueryRow(ctx, `SELECT id FROM users WHERE id = $1 AND deleted_at IS NULL`, targetID).Scan(&dummy)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrUserNotFound
	}
	if err != nil {
		return fmt.Errorf("lookup user: %w", err)
	}

	// INSERT с ON CONFLICT — идемпотентность (повторный grant обновляет expires_at).
	tag, err := tx.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id, granted_by, granted_at, expires_at)
		VALUES ($1, $2, $3, now(), $4)
		ON CONFLICT (user_id, role_id)
		DO UPDATE SET granted_by = $3, granted_at = now(), expires_at = $4`,
		targetID, roleID, actorID, opts.ExpiresAt)
	if err != nil {
		return fmt.Errorf("insert user_roles: %w", err)
	}
	_ = tag

	// Audit log.
	_, err = tx.Exec(ctx, `
		INSERT INTO audit_role_changes
			(actor_id, target_id, role_name, action, reason, ip_address, user_agent)
		VALUES ($1, $2, $3, 'grant', $4, $5, $6)`,
		actorID, targetID, roleName, opts.Reason, opts.IPAddress, opts.UserAgent)
	if err != nil {
		return fmt.Errorf("insert audit: %w", err)
	}

	return tx.Commit(ctx)
}

// RevokeUserRole снимает роль с юзера. Транзакция: DELETE user_roles + INSERT
// audit_role_changes.
func (s *RBACService) RevokeUserRole(
	ctx context.Context,
	actorID uuid.UUID,
	targetID uuid.UUID,
	roleName string,
	reason string,
	ipAddress *net.IP,
	userAgent string,
) error {
	if reason == "" {
		return errors.New("reason is required for revoke")
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var roleID int
	err = tx.QueryRow(ctx, `SELECT id FROM roles WHERE name = $1`, roleName).Scan(&roleID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrRoleNotFound
	}
	if err != nil {
		return fmt.Errorf("lookup role: %w", err)
	}

	tag, err := tx.Exec(ctx,
		`DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`,
		targetID, roleID)
	if err != nil {
		return fmt.Errorf("delete user_roles: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotGranted
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO audit_role_changes
			(actor_id, target_id, role_name, action, reason, ip_address, user_agent)
		VALUES ($1, $2, $3, 'revoke', $4, $5, $6)`,
		actorID, targetID, roleName, reason, ipAddress, userAgent)
	if err != nil {
		return fmt.Errorf("insert audit: %w", err)
	}

	return tx.Commit(ctx)
}

// AuditQuery — параметры выборки audit-log с пагинацией.
type AuditQuery struct {
	ActorID  *uuid.UUID
	TargetID *uuid.UUID
	Action   string // 'grant' | 'revoke' | ""
	Since    *time.Time
	Until    *time.Time
	Limit    int
	Offset   int
}

// QueryAuditChanges — пагинированная выборка из audit_role_changes
// с опциональными фильтрами. Сортировка по created_at DESC.
func (s *RBACService) QueryAuditChanges(ctx context.Context, q AuditQuery) ([]AuditRoleChange, error) {
	if q.Limit <= 0 || q.Limit > 200 {
		q.Limit = 50
	}

	args := []any{q.Limit, q.Offset}
	where := "WHERE 1=1"
	if q.ActorID != nil {
		args = append(args, *q.ActorID)
		where += fmt.Sprintf(" AND actor_id = $%d", len(args))
	}
	if q.TargetID != nil {
		args = append(args, *q.TargetID)
		where += fmt.Sprintf(" AND target_id = $%d", len(args))
	}
	if q.Action != "" {
		args = append(args, q.Action)
		where += fmt.Sprintf(" AND action = $%d", len(args))
	}
	if q.Since != nil {
		args = append(args, *q.Since)
		where += fmt.Sprintf(" AND created_at >= $%d", len(args))
	}
	if q.Until != nil {
		args = append(args, *q.Until)
		where += fmt.Sprintf(" AND created_at <= $%d", len(args))
	}

	sql := fmt.Sprintf(`
		SELECT id, actor_id, target_id, role_name, action,
		       COALESCE(reason,''), ip_address, COALESCE(user_agent,''), created_at
		  FROM audit_role_changes
		  %s
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`, where)

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query audit: %w", err)
	}
	defer rows.Close()

	out := make([]AuditRoleChange, 0, q.Limit)
	for rows.Next() {
		var a AuditRoleChange
		var ipAddr pgxIP
		if err := rows.Scan(&a.ID, &a.ActorID, &a.TargetID, &a.RoleName, &a.Action,
			&a.Reason, &ipAddr, &a.UserAgent, &a.CreatedAt); err != nil {
			return nil, err
		}
		a.IPAddress = ipAddr.IP
		out = append(out, a)
	}
	return out, rows.Err()
}

// pgxIP — тонкий wrapper для NULL-able net.IP в Scan.
type pgxIP struct {
	IP *net.IP
}

func (p *pgxIP) Scan(src any) error {
	if src == nil {
		p.IP = nil
		return nil
	}
	switch v := src.(type) {
	case string:
		ip := net.ParseIP(v)
		p.IP = &ip
	case []byte:
		ip := net.ParseIP(string(v))
		p.IP = &ip
	}
	return nil
}
