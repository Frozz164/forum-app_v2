package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Frozz164/forum-app_v2/forum-service/internal/domain"
	"github.com/Frozz164/forum-app_v2/forum-service/internal/repository"
)

type PostService interface {
	CreatePost(ctx context.Context, post *domain.Post) (*domain.Post, error)
	GetPost(ctx context.Context, id int64) (*domain.Post, error)
	GetAllPosts(ctx context.Context) ([]*domain.Post, error)
	GetPostsPaginated(ctx context.Context, offset, limit int) ([]*domain.Post, error)
	DeletePost(ctx context.Context, id, authorID int64) error
	GetPostsWithAuthors(ctx context.Context) ([]*domain.Post, error)
}

type PostServiceImpl struct {
	repo repository.PostRepository
}

func NewPostService(repo repository.PostRepository) PostService {
	return &PostServiceImpl{repo: repo}
}

func (s *PostServiceImpl) CreatePost(ctx context.Context, post *domain.Post) (*domain.Post, error) {
	// Валидация
	post.Title = strings.TrimSpace(post.Title)
	post.Content = strings.TrimSpace(post.Content)

	if len(post.Title) < 3 || len(post.Title) > 100 {
		return nil, errors.New("title must be between 3-100 characters")
	}
	if len(post.Content) < 10 {
		return nil, errors.New("content must be at least 10 characters")
	}
	if post.AuthorID == 0 {
		return nil, errors.New("author ID is required")
	}

	id, err := s.repo.Create(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *PostServiceImpl) GetPost(ctx context.Context, id int64) (*domain.Post, error) {
	if id <= 0 {
		return nil, errors.New("invalid post ID")
	}

	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}
	if post == nil {
		return nil, errors.New("post not found")
	}

	return post, nil
}

func (s *PostServiceImpl) GetAllPosts(ctx context.Context) ([]*domain.Post, error) {
	posts, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts: %w", err)
	}

	return posts, nil
}

func (s *PostServiceImpl) GetPostsPaginated(ctx context.Context, offset, limit int) ([]*domain.Post, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.GetPostsPaginated(ctx, offset, limit)
}

func (s *PostServiceImpl) DeletePost(ctx context.Context, id, authorID int64) error {
	if id <= 0 || authorID <= 0 {
		return errors.New("invalid ID")
	}

	return s.repo.Delete(ctx, id, authorID)
}

func (s *PostServiceImpl) GetPostsWithAuthors(ctx context.Context) ([]*domain.Post, error) {
	posts, err := s.repo.GetPostsWithAuthors(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts with authors: %w", err)
	}

	return posts, nil
}
