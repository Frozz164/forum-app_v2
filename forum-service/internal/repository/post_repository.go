package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/Frozz164/forum-app_v2/forum-service/internal/domain"
)

type PostRepository interface {
	Create(ctx context.Context, post *domain.Post) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Post, error)
	GetAll(ctx context.Context) ([]*domain.Post, error)
	Delete(ctx context.Context, id, authorID int64) error
	GetPostsWithAuthors(ctx context.Context) ([]*domain.Post, error)
}

type PostRepositoryImpl struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) PostRepository {
	return &PostRepositoryImpl{db: db}
}

func (r *PostRepositoryImpl) Create(ctx context.Context, post *domain.Post) (int64, error) {
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
		log.Printf("Error creating post: %v", err)
		return 0, fmt.Errorf("failed to create post: %w", err)
	}

	return id, nil
}

func (r *PostRepositoryImpl) GetByID(ctx context.Context, id int64) (*domain.Post, error) {
	query := `
		SELECT id, title, content, author_id, created_at
		FROM posts
		WHERE id = $1
	`

	var post domain.Post
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.AuthorID,
		&post.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Printf("Error getting post by ID: %v", err)
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	return &post, nil
}

func (r *PostRepositoryImpl) GetAll(ctx context.Context) ([]*domain.Post, error) {
	query := `
		SELECT id, title, content, author_id, created_at
		FROM posts
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts: %w", err)
	}
	defer rows.Close()

	var posts []*domain.Post
	for rows.Next() {
		var post domain.Post
		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.AuthorID,
			&post.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, &post)
	}

	return posts, nil
}

func (r *PostRepositoryImpl) Delete(ctx context.Context, id, authorID int64) error {
	query := `
		DELETE FROM posts
		WHERE id = $1 AND author_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, id, authorID)
	if err != nil {
		log.Printf("Error deleting post: %v", err)
		return fmt.Errorf("failed to delete post: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("post not found or you're not the author")
	}

	return nil

}
func (r *PostRepositoryImpl) GetPostsWithAuthors(ctx context.Context) ([]*domain.Post, error) {
	query := `
        SELECT p.id, p.title, p.content, p.author_id, p.created_at, u.username
        FROM posts p
        JOIN users u ON p.author_id = u.id
        ORDER BY p.created_at DESC
    `

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts: %w", err)
	}
	defer rows.Close()

	var posts []*domain.Post
	for rows.Next() {
		var post domain.Post
		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.AuthorID,
			&post.CreatedAt,
			&post.Author,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, &post)
	}

	return posts, nil
}
