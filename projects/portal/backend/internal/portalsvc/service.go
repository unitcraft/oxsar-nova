// Package portalsvc реализует логику портала: новости и система предложений.
package portalsvc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"oxsar/portal/internal/repo"
	"oxsar/portal/pkg/ids"
)

var (
	ErrNotFound            = errors.New("portalsvc: not found")
	ErrForbidden           = errors.New("portalsvc: forbidden")
	ErrTooManyProposals    = errors.New("portalsvc: too many active proposals (max 5)")
	ErrInsufficientCredits = errors.New("portalsvc: insufficient credits")
)

const maxActiveProposals = 5

// NewsItem — новость.
type NewsItem struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	AuthorID  string    `json:"author_id"`
	Published bool      `json:"published"`
	Pinned    bool      `json:"pinned"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FeedbackPost — предложение игрока.
type FeedbackPost struct {
	ID         string    `json:"id"`
	AuthorID   string    `json:"author_id"`
	AuthorName string    `json:"author_name"`
	Title      string    `json:"title"`
	Body       string    `json:"body"`
	Status     string    `json:"status"` // pending|approved|rejected|implemented
	VoteCount  int64     `json:"vote_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// FeedbackComment — комментарий к предложению.
type FeedbackComment struct {
	ID         string     `json:"id"`
	PostID     string     `json:"post_id"`
	ParentID   *string    `json:"parent_id,omitempty"`
	AuthorID   string     `json:"author_id"`
	AuthorName string     `json:"author_name"`
	Body       string     `json:"body"`
	CreatedAt  time.Time  `json:"created_at"`
	EditedAt   *time.Time `json:"edited_at,omitempty"`
}

// Service — основной сервис портала.
type Service struct {
	db *repo.PG
}

// New создаёт Service.
func New(pool *pgxpool.Pool) *Service {
	return &Service{db: repo.New(pool)}
}

// --- News ---

// ListNews возвращает опубликованные новости, сначала закреплённые.
func (s *Service) ListNews(ctx context.Context, limit, offset int) ([]NewsItem, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	rows, err := s.db.Pool().Query(ctx, `
		SELECT id, title, body, author_id, published, pinned, created_at, updated_at
		FROM news
		WHERE published = true
		ORDER BY pinned DESC, created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list news: %w", err)
	}
	defer rows.Close()
	return scanNews(rows)
}

// CreateNews создаёт новость (только для модераторов/администраторов).
func (s *Service) CreateNews(ctx context.Context, authorID, title, body string, published bool) (NewsItem, error) {
	id := ids.New()
	_, err := s.db.Pool().Exec(ctx, `
		INSERT INTO news (id, title, body, author_id, published, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, now(), now())
	`, id, title, body, authorID, published)
	if err != nil {
		return NewsItem{}, fmt.Errorf("create news: %w", err)
	}
	return s.GetNews(ctx, id)
}

// GetNews возвращает новость по ID.
func (s *Service) GetNews(ctx context.Context, id string) (NewsItem, error) {
	var n NewsItem
	err := s.db.Pool().QueryRow(ctx, `
		SELECT id, title, body, author_id, published, pinned, created_at, updated_at
		FROM news WHERE id = $1
	`, id).Scan(&n.ID, &n.Title, &n.Body, &n.AuthorID, &n.Published, &n.Pinned, &n.CreatedAt, &n.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return NewsItem{}, ErrNotFound
	}
	if err != nil {
		return NewsItem{}, fmt.Errorf("get news: %w", err)
	}
	return n, nil
}

// --- Feedback ---

// ListFeedback возвращает одобренные предложения, сортировка по голосам.
func (s *Service) ListFeedback(ctx context.Context, status string, limit, offset int) ([]FeedbackPost, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	if status == "" {
		status = "approved"
	}
	rows, err := s.db.Pool().Query(ctx, `
		SELECT id, author_id, author_name, title, body, status, vote_count, created_at, updated_at
		FROM feedback_posts
		WHERE status = $1
		ORDER BY vote_count DESC, created_at DESC
		LIMIT $2 OFFSET $3
	`, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list feedback: %w", err)
	}
	defer rows.Close()
	return scanPosts(rows)
}

// CreateFeedback создаёт новое предложение (идёт на модерацию).
func (s *Service) CreateFeedback(ctx context.Context, authorID, authorName, title, body string) (FeedbackPost, error) {
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)

	// Проверяем лимит активных предложений
	var activeCount int
	if err := s.db.Pool().QueryRow(ctx, `
		SELECT count(*) FROM feedback_posts
		WHERE author_id = $1 AND status IN ('pending','approved')
	`, authorID).Scan(&activeCount); err != nil {
		return FeedbackPost{}, fmt.Errorf("count proposals: %w", err)
	}
	if activeCount >= maxActiveProposals {
		return FeedbackPost{}, ErrTooManyProposals
	}

	id := ids.New()
	_, err := s.db.Pool().Exec(ctx, `
		INSERT INTO feedback_posts (id, author_id, author_name, title, body, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'pending', now(), now())
	`, id, authorID, authorName, title, body)
	if err != nil {
		return FeedbackPost{}, fmt.Errorf("create feedback: %w", err)
	}
	return s.GetFeedback(ctx, id)
}

// GetFeedback возвращает предложение по ID.
func (s *Service) GetFeedback(ctx context.Context, id string) (FeedbackPost, error) {
	var p FeedbackPost
	err := s.db.Pool().QueryRow(ctx, `
		SELECT id, author_id, author_name, title, body, status, vote_count, created_at, updated_at
		FROM feedback_posts WHERE id = $1
	`, id).Scan(&p.ID, &p.AuthorID, &p.AuthorName, &p.Title, &p.Body, &p.Status, &p.VoteCount, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return FeedbackPost{}, ErrNotFound
	}
	if err != nil {
		return FeedbackPost{}, fmt.Errorf("get feedback: %w", err)
	}
	return p, nil
}

// ModerateFeedback меняет статус предложения (pending→approved|rejected, approved→implemented).
func (s *Service) ModerateFeedback(ctx context.Context, id, newStatus string) error {
	tag, err := s.db.Pool().Exec(ctx, `
		UPDATE feedback_posts SET status = $1, updated_at = now() WHERE id = $2
	`, newStatus, id)
	if err != nil {
		return fmt.Errorf("moderate feedback: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// VoteFeedback добавляет голос за предложение (вызывается после списания кредитов через billing-service).
func (s *Service) VoteFeedback(ctx context.Context, postID, userID string, creditsSpent int64) error {
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO feedback_votes (post_id, user_id, credits_spent, created_at)
			VALUES ($1, $2, $3, now())
		`, postID, userID, creditsSpent)
		if err != nil {
			return fmt.Errorf("insert vote: %w", err)
		}
		_, err = tx.Exec(ctx, `
			UPDATE feedback_posts SET vote_count = vote_count + 1, updated_at = now()
			WHERE id = $1 AND status = 'approved'
		`, postID)
		return err
	})
}

// ListComments возвращает комментарии к предложению (не удалённые).
func (s *Service) ListComments(ctx context.Context, postID string) ([]FeedbackComment, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT id, post_id, parent_id, author_id, author_name, body, created_at, edited_at
		FROM feedback_comments
		WHERE post_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC
	`, postID)
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	defer rows.Close()
	return scanComments(rows)
}

// AddComment добавляет комментарий к предложению.
func (s *Service) AddComment(ctx context.Context, postID string, parentID *string, authorID, authorName, body string) (FeedbackComment, error) {
	id := ids.New()
	_, err := s.db.Pool().Exec(ctx, `
		INSERT INTO feedback_comments (id, post_id, parent_id, author_id, author_name, body, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, now())
	`, id, postID, parentID, authorID, authorName, body)
	if err != nil {
		return FeedbackComment{}, fmt.Errorf("add comment: %w", err)
	}
	var c FeedbackComment
	err = s.db.Pool().QueryRow(ctx, `
		SELECT id, post_id, parent_id, author_id, author_name, body, created_at, edited_at
		FROM feedback_comments WHERE id = $1
	`, id).Scan(&c.ID, &c.PostID, &c.ParentID, &c.AuthorID, &c.AuthorName, &c.Body, &c.CreatedAt, &c.EditedAt)
	if err != nil {
		return FeedbackComment{}, fmt.Errorf("get comment: %w", err)
	}
	return c, nil
}

// --- scan helpers ---

func scanNews(rows pgx.Rows) ([]NewsItem, error) {
	var result []NewsItem
	for rows.Next() {
		var n NewsItem
		if err := rows.Scan(&n.ID, &n.Title, &n.Body, &n.AuthorID, &n.Published, &n.Pinned, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, n)
	}
	return result, rows.Err()
}

func scanPosts(rows pgx.Rows) ([]FeedbackPost, error) {
	var result []FeedbackPost
	for rows.Next() {
		var p FeedbackPost
		if err := rows.Scan(&p.ID, &p.AuthorID, &p.AuthorName, &p.Title, &p.Body, &p.Status, &p.VoteCount, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

func scanComments(rows pgx.Rows) ([]FeedbackComment, error) {
	var result []FeedbackComment
	for rows.Next() {
		var c FeedbackComment
		if err := rows.Scan(&c.ID, &c.PostID, &c.ParentID, &c.AuthorID, &c.AuthorName, &c.Body, &c.CreatedAt, &c.EditedAt); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}
