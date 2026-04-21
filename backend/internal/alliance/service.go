// Package alliance — создание и управление альянсами (MVP M6).
//
// Ограничения:
//   - Один игрок — один альянс (PK alliance_members.user_id).
//   - Только owner может распустить альянс (DELETE).
//   - Kick-механика, ранги, заявки, отношения (NAP/WAR/ALLY) — M6+.
package alliance

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/repo"
)

type Service struct {
	db repo.Exec
}

func NewService(db repo.Exec) *Service { return &Service{db: db} }

var (
	ErrNotFound       = errors.New("alliance: not found")
	ErrAlreadyMember  = errors.New("alliance: already in an alliance")
	ErrNotMember      = errors.New("alliance: not a member")
	ErrNotOwner       = errors.New("alliance: not the owner")
	ErrTagTaken       = errors.New("alliance: tag already taken")
	ErrNameTaken      = errors.New("alliance: name already taken")
	ErrInvalidTag     = errors.New("alliance: tag must be 3–5 latin letters/digits")
	ErrCannotLeaveOwn = errors.New("alliance: owner must transfer or disband before leaving")
)

// Alliance — полная запись для UI.
type Alliance struct {
	ID          string    `json:"id"`
	Tag         string    `json:"tag"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     string    `json:"owner_id"`
	OwnerName   string    `json:"owner_name"`
	MemberCount int       `json:"member_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// Member — элемент списка участников.
type Member struct {
	UserID   string    `json:"user_id"`
	Username string    `json:"username"`
	Rank     string    `json:"rank"`
	JoinedAt time.Time `json:"joined_at"`
}

// List возвращает первые N альянсов, сортировка по числу участников.
func (s *Service) List(ctx context.Context, limit int) ([]Alliance, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := s.db.Pool().Query(ctx, `
		SELECT a.id, a.tag, a.name, a.description, a.owner_id,
		       COALESCE(u.username,'') AS owner_name,
		       COUNT(m.user_id)        AS member_count,
		       a.created_at
		FROM alliances a
		LEFT JOIN users u ON u.id = a.owner_id
		LEFT JOIN alliance_members m ON m.alliance_id = a.id
		GROUP BY a.id, u.username
		ORDER BY member_count DESC, a.created_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("alliances list: %w", err)
	}
	defer rows.Close()
	var out []Alliance
	for rows.Next() {
		var al Alliance
		if err := rows.Scan(&al.ID, &al.Tag, &al.Name, &al.Description,
			&al.OwnerID, &al.OwnerName, &al.MemberCount, &al.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, al)
	}
	return out, rows.Err()
}

// Get возвращает альянс по ID вместе со списком участников.
func (s *Service) Get(ctx context.Context, id string) (Alliance, []Member, error) {
	var al Alliance
	err := s.db.Pool().QueryRow(ctx, `
		SELECT a.id, a.tag, a.name, a.description, a.owner_id,
		       COALESCE(u.username,'') AS owner_name,
		       (SELECT COUNT(*) FROM alliance_members WHERE alliance_id=a.id),
		       a.created_at
		FROM alliances a
		LEFT JOIN users u ON u.id = a.owner_id
		WHERE a.id = $1
	`, id).Scan(&al.ID, &al.Tag, &al.Name, &al.Description,
		&al.OwnerID, &al.OwnerName, &al.MemberCount, &al.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Alliance{}, nil, ErrNotFound
		}
		return Alliance{}, nil, fmt.Errorf("get alliance: %w", err)
	}

	rows, err := s.db.Pool().Query(ctx, `
		SELECT m.user_id, COALESCE(u.username,''), m.rank, m.joined_at
		FROM alliance_members m
		JOIN users u ON u.id = m.user_id
		WHERE m.alliance_id = $1
		ORDER BY
		  CASE m.rank WHEN 'owner' THEN 0 ELSE 1 END,
		  m.joined_at ASC
	`, id)
	if err != nil {
		return Alliance{}, nil, fmt.Errorf("get members: %w", err)
	}
	defer rows.Close()
	var members []Member
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.UserID, &m.Username, &m.Rank, &m.JoinedAt); err != nil {
			return Alliance{}, nil, err
		}
		members = append(members, m)
	}
	return al, members, rows.Err()
}

// Create создаёт новый альянс и делает создателя owner'ом.
func (s *Service) Create(ctx context.Context, ownerID, tag, name, description string) (Alliance, error) {
	tag = strings.TrimSpace(tag)
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	if err := validateTag(tag); err != nil {
		return Alliance{}, err
	}
	if n := utf8.RuneCountInString(name); n < 3 || n > 64 {
		return Alliance{}, fmt.Errorf("alliance: name must be 3–64 characters")
	}
	if utf8.RuneCountInString(description) > 2000 {
		description = string([]rune(description)[:2000])
	}

	var out Alliance
	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Проверяем что пользователь не в альянсе.
		var existing *string
		if err := tx.QueryRow(ctx,
			`SELECT alliance_id FROM users WHERE id=$1`, ownerID).Scan(&existing); err != nil {
			return fmt.Errorf("check user: %w", err)
		}
		if existing != nil {
			return ErrAlreadyMember
		}

		var id string
		err := tx.QueryRow(ctx, `
			INSERT INTO alliances (tag, name, description, owner_id)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, tag, name, description, ownerID).Scan(&id)
		if err != nil {
			if isDupKey(err) {
				if strings.Contains(err.Error(), "alliances_tag_key") {
					return ErrTagTaken
				}
				return ErrNameTaken
			}
			return fmt.Errorf("insert alliance: %w", err)
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO alliance_members (alliance_id, user_id, rank)
			VALUES ($1, $2, 'owner')
		`, id, ownerID); err != nil {
			return fmt.Errorf("insert member: %w", err)
		}

		if _, err := tx.Exec(ctx,
			`UPDATE users SET alliance_id=$1 WHERE id=$2`, id, ownerID); err != nil {
			return fmt.Errorf("update user: %w", err)
		}

		var ownerName string
		_ = tx.QueryRow(ctx, `SELECT username FROM users WHERE id=$1`, ownerID).Scan(&ownerName)
		out = Alliance{
			ID: id, Tag: tag, Name: name, Description: description,
			OwnerID: ownerID, OwnerName: ownerName, MemberCount: 1,
		}
		return nil
	})
	return out, err
}

// Join добавляет пользователя в альянс.
func (s *Service) Join(ctx context.Context, userID, allianceID string) error {
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var existing *string
		if err := tx.QueryRow(ctx,
			`SELECT alliance_id FROM users WHERE id=$1`, userID).Scan(&existing); err != nil {
			return fmt.Errorf("check user: %w", err)
		}
		if existing != nil {
			return ErrAlreadyMember
		}
		// Проверяем что альянс существует.
		var ownerID string
		err := tx.QueryRow(ctx, `SELECT owner_id FROM alliances WHERE id=$1`, allianceID).Scan(&ownerID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("check alliance: %w", err)
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO alliance_members (alliance_id, user_id, rank)
			VALUES ($1, $2, 'member')
		`, allianceID, userID); err != nil {
			return fmt.Errorf("insert member: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`UPDATE users SET alliance_id=$1 WHERE id=$2`, allianceID, userID); err != nil {
			return fmt.Errorf("update user: %w", err)
		}
		return nil
	})
}

// Leave удаляет пользователя из альянса. Owner не может выйти
// (должен сначала передать права или распустить альянс).
func (s *Service) Leave(ctx context.Context, userID string) error {
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var rank string
		var allianceID string
		err := tx.QueryRow(ctx,
			`SELECT alliance_id, rank FROM alliance_members WHERE user_id=$1`,
			userID).Scan(&allianceID, &rank)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotMember
			}
			return fmt.Errorf("check member: %w", err)
		}
		if rank == "owner" {
			return ErrCannotLeaveOwn
		}

		if _, err := tx.Exec(ctx,
			`DELETE FROM alliance_members WHERE user_id=$1`, userID); err != nil {
			return fmt.Errorf("delete member: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`UPDATE users SET alliance_id=NULL WHERE id=$1`, userID); err != nil {
			return fmt.Errorf("update user: %w", err)
		}
		return nil
	})
}

// Disband распускает альянс. Только owner.
func (s *Service) Disband(ctx context.Context, userID, allianceID string) error {
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var ownerID string
		err := tx.QueryRow(ctx,
			`SELECT owner_id FROM alliances WHERE id=$1`, allianceID).Scan(&ownerID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("check alliance: %w", err)
		}
		if ownerID != userID {
			return ErrNotOwner
		}
		// ON DELETE CASCADE уберёт alliance_members.
		// ON DELETE SET NULL уберёт users.alliance_id.
		if _, err := tx.Exec(ctx,
			`DELETE FROM alliances WHERE id=$1`, allianceID); err != nil {
			return fmt.Errorf("disband: %w", err)
		}
		return nil
	})
}

// MyAlliance возвращает альянс текущего пользователя (если есть).
func (s *Service) MyAlliance(ctx context.Context, userID string) (*Alliance, []Member, error) {
	var allianceID *string
	if err := s.db.Pool().QueryRow(ctx,
		`SELECT alliance_id FROM users WHERE id=$1`, userID).Scan(&allianceID); err != nil {
		return nil, nil, fmt.Errorf("my alliance: %w", err)
	}
	if allianceID == nil {
		return nil, nil, nil
	}
	al, members, err := s.Get(ctx, *allianceID)
	if err != nil {
		return nil, nil, err
	}
	return &al, members, nil
}

func validateTag(tag string) error {
	n := utf8.RuneCountInString(tag)
	if n < 3 || n > 5 {
		return ErrInvalidTag
	}
	for _, r := range tag {
		if !isAlphanumASCII(r) {
			return ErrInvalidTag
		}
	}
	return nil
}

func isAlphanumASCII(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
}

func isDupKey(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate key")
}
