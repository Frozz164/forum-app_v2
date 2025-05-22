package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Frozz164/forum-app_v2/forum-service/internal/domain"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type PostRepositoryImpl struct {
	db     *sql.DB
	logger zerolog.Logger
}
type PostRepository interface {
	Create(ctx context.Context, post *domain.Post) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Post, error)
	GetAll(ctx context.Context) ([]*domain.Post, error)
	Delete(ctx context.Context, id, authorID int64) error
	GetPostsWithAuthors(ctx context.Context) ([]*domain.Post, error)
	GetPostsPaginated(ctx context.Context, offset, limit int) ([]*domain.Post, error)
}

func NewPostRepository(db *sql.DB) PostRepository {
	return &PostRepositoryImpl{
		db:     db,
		logger: log.With().Str("component", "post_repository").Logger(),
	}
}

func (r *PostRepositoryImpl) Create(ctx context.Context, post *domain.Post) (int64, error) {
	logger := r.logger.With().
		Str("method", "Create").
		Str("title", post.Title).
		Int64("author_id", post.AuthorID).
		Logger()

	query := `
		INSERT INTO posts (title, content, author_id, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(ctx, query,
		post.Title,
		post.Content,
		post.AuthorID,
		time.Now(),
	).Scan(&id)

	if err != nil {
		logger.Error().Err(err).
			Str("content_prefix", truncateString(post.Content, 20)).
			Msg("Failed to create post")
		return 0, fmt.Errorf("failed to create post: %w", err)
	}

	logger.Info().
		Int64("post_id", id).
		Msg("Post created successfully")
	return id, nil
}

func (r *PostRepositoryImpl) GetByID(ctx context.Context, id int64) (*domain.Post, error) {
	logger := r.logger.With().
		Str("method", "GetByID").
		Int64("post_id", id).
		Logger()

	query := `
		SELECT id, title, content, author_id, created_at
		FROM posts
		WHERE id = $1
	`

	var post domain.Post
	var createdAt time.Time

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.AuthorID,
		&createdAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Debug().Msg("Post not found")
			return nil, nil
		}
		logger.Error().Err(err).Msg("Failed to get post")
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	post.CreatedAt = createdAt.Format(time.RFC3339)
	logger.Debug().Msg("Post retrieved successfully")
	return &post, nil
}

func (r *PostRepositoryImpl) GetAll(ctx context.Context) ([]*domain.Post, error) {
	logger := r.logger.With().
		Str("method", "GetAll").
		Logger()

	query := `
		SELECT id, title, content, author_id, created_at
		FROM posts
		ORDER BY created_at DESC
	`

	posts, err := r.queryPosts(ctx, query)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get all posts")
		return nil, err
	}

	logger.Debug().
		Int("post_count", len(posts)).
		Msg("Successfully retrieved all posts")
	return posts, nil
}

func (r *PostRepositoryImpl) GetPostsWithAuthors(ctx context.Context) ([]*domain.Post, error) {
	logger := r.logger.With().
		Str("method", "GetPostsWithAuthors").
		Logger()

	query := `
		SELECT p.id, p.title, p.content, p.author_id, p.created_at, u.username as author
		FROM posts p
		JOIN users u ON p.author_id = u.id
		ORDER BY p.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to query posts with authors")
		return nil, fmt.Errorf("failed to query posts: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Warn().Err(err).Msg("Failed to close rows")
		}
	}()

	var posts []*domain.Post
	for rows.Next() {
		var post domain.Post
		var createdAt time.Time

		if err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.AuthorID,
			&createdAt,
			&post.Author,
		); err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}

		post.CreatedAt = createdAt.Format(time.RFC3339)
		posts = append(posts, &post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	logger.Debug().
		Int("post_count", len(posts)).
		Msg("Successfully retrieved posts with authors")
	return posts, nil
}

func (r *PostRepositoryImpl) GetPostsPaginated(ctx context.Context, offset, limit int) ([]*domain.Post, error) {
	logger := r.logger.With().
		Str("method", "GetPostsPaginated").
		Int("offset", offset).
		Int("limit", limit).
		Logger()

	if limit <= 0 {
		limit = 10
		logger.Debug().Msg("Using default limit value")
	}
	if offset < 0 {
		offset = 0
		logger.Debug().Msg("Using default offset value")
	}

	query := `
		SELECT id, title, content, author_id, created_at
		FROM posts
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	posts, err := r.queryPosts(ctx, query, limit, offset)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get paginated posts")
		return nil, err
	}

	logger.Debug().
		Int("post_count", len(posts)).
		Msg("Successfully retrieved paginated posts")
	return posts, nil
}

func (r *PostRepositoryImpl) Delete(ctx context.Context, id, authorID int64) error {
	logger := r.logger.With().
		Str("method", "Delete").
		Int64("post_id", id).
		Int64("author_id", authorID).
		Logger()

	query := `
		DELETE FROM posts
		WHERE id = $1 AND author_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, id, authorID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to delete post")
		return fmt.Errorf("failed to delete post: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to check rows affected")
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		logger.Warn().Msg("Post not found or not authorized")
		return fmt.Errorf("post not found or not authorized")
	}

	logger.Info().Msg("Post deleted successfully")
	return nil
}

func (r *PostRepositoryImpl) queryPosts(ctx context.Context, query string, args ...interface{}) ([]*domain.Post, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query posts: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Warn().Err(err).Msg("Failed to close rows")
		}
	}()

	var posts []*domain.Post
	for rows.Next() {
		var post domain.Post
		var createdAt time.Time

		if err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.AuthorID,
			&createdAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}

		post.CreatedAt = createdAt.Format(time.RFC3339)
		posts = append(posts, &post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return posts, nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
