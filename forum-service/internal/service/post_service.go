package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Frozz164/forum-app_v2/forum-service/internal/domain"
	"github.com/Frozz164/forum-app_v2/forum-service/internal/repository"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type PostServiceImpl struct {
	repo   repository.PostRepository
	logger zerolog.Logger
}

type PostService interface {
	CreatePost(ctx context.Context, post *domain.Post) (*domain.Post, error)
	GetPost(ctx context.Context, id int64) (*domain.Post, error)
	GetAllPosts(ctx context.Context) ([]*domain.Post, error)
	GetPostsPaginated(ctx context.Context, offset, limit int) ([]*domain.Post, error)
	DeletePost(ctx context.Context, id, authorID int64) error
	GetPostsWithAuthors(ctx context.Context) ([]*domain.Post, error)
}

func NewPostService(repo repository.PostRepository) PostService {
	return &PostServiceImpl{
		repo:   repo,
		logger: log.With().Str("component", "post_service").Logger(),
	}
}

func (s *PostServiceImpl) CreatePost(ctx context.Context, post *domain.Post) (*domain.Post, error) {
	logger := s.logger.With().
		Str("method", "CreatePost").
		Str("title", post.Title).
		Int64("author_id", post.AuthorID).
		Logger()

	// Валидация
	post.Title = strings.TrimSpace(post.Title)
	post.Content = strings.TrimSpace(post.Content)

	if len(post.Title) < 3 || len(post.Title) > 100 {
		err := errors.New("title must be between 3-100 characters")
		logger.Warn().Err(err).Msg("Validation failed")
		return nil, err
	}
	if len(post.Content) < 10 {
		err := errors.New("content must be at least 10 characters")
		logger.Warn().Err(err).Msg("Validation failed")
		return nil, err
	}
	if post.AuthorID == 0 {
		err := errors.New("author ID is required")
		logger.Warn().Err(err).Msg("Validation failed")
		return nil, err
	}

	id, err := s.repo.Create(ctx, post)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create post in repository")
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	createdPost, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.Error().Err(err).
			Int64("post_id", id).
			Msg("Failed to fetch created post")
		return nil, fmt.Errorf("failed to fetch created post: %w", err)
	}

	logger.Info().
		Int64("post_id", id).
		Msg("Post created successfully")
	return createdPost, nil
}

func (s *PostServiceImpl) GetPost(ctx context.Context, id int64) (*domain.Post, error) {
	logger := s.logger.With().
		Str("method", "GetPost").
		Int64("post_id", id).
		Logger()

	if id <= 0 {
		err := errors.New("invalid post ID")
		logger.Warn().Err(err).Msg("Validation failed")
		return nil, err
	}

	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get post from repository")
		return nil, fmt.Errorf("failed to get post: %w", err)
	}
	if post == nil {
		logger.Debug().Msg("Post not found")
		return nil, errors.New("post not found")
	}

	logger.Debug().Msg("Post retrieved successfully")
	return post, nil
}

func (s *PostServiceImpl) GetAllPosts(ctx context.Context) ([]*domain.Post, error) {
	logger := s.logger.With().
		Str("method", "GetAllPosts").
		Logger()

	posts, err := s.repo.GetAll(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get posts from repository")
		return nil, fmt.Errorf("failed to get posts: %w", err)
	}

	logger.Debug().
		Int("post_count", len(posts)).
		Msg("Retrieved all posts successfully")
	return posts, nil
}

func (s *PostServiceImpl) GetPostsPaginated(ctx context.Context, offset, limit int) ([]*domain.Post, error) {
	logger := s.logger.With().
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

	posts, err := s.repo.GetPostsPaginated(ctx, offset, limit)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get paginated posts from repository")
		return nil, fmt.Errorf("failed to get paginated posts: %w", err)
	}

	logger.Debug().
		Int("post_count", len(posts)).
		Msg("Retrieved paginated posts successfully")
	return posts, nil
}

func (s *PostServiceImpl) DeletePost(ctx context.Context, id, authorID int64) error {
	logger := s.logger.With().
		Str("method", "DeletePost").
		Int64("post_id", id).
		Int64("author_id", authorID).
		Logger()

	if id <= 0 || authorID <= 0 {
		err := errors.New("invalid ID")
		logger.Warn().Err(err).Msg("Validation failed")
		return err
	}

	err := s.repo.Delete(ctx, id, authorID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to delete post in repository")
		return fmt.Errorf("failed to delete post: %w", err)
	}

	logger.Info().Msg("Post deleted successfully")
	return nil
}

func (s *PostServiceImpl) GetPostsWithAuthors(ctx context.Context) ([]*domain.Post, error) {
	logger := s.logger.With().
		Str("method", "GetPostsWithAuthors").
		Logger()

	posts, err := s.repo.GetPostsWithAuthors(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get posts with authors from repository")
		return nil, fmt.Errorf("failed to get posts with authors: %w", err)
	}

	logger.Debug().
		Int("post_count", len(posts)).
		Msg("Retrieved posts with authors successfully")
	return posts, nil
}
