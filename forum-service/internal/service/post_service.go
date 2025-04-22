package service

import (
	"context"
	"errors"
	"strings"

	"github.com/Frozz164/forum-app_v2/forum-service/internal/domain"
	"github.com/Frozz164/forum-app_v2/forum-service/internal/repository"
)

type (
	PostService interface {
		CreatePost(ctx context.Context, post *domain.Post) (*domain.Post, error)
		GetPost(ctx context.Context, id int64) (*domain.Post, error)
		GetAllPosts(ctx context.Context) ([]*domain.Post, error)
		DeletePost(ctx context.Context, id, authorID int64) error
		GetPostsWithAuthors(ctx context.Context) (interface{}, interface{})
	}
)

type PostServiceImpl struct {
	repo repository.PostRepository
}

func NewPostService(repo repository.PostRepository) *PostServiceImpl {
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

	id, err := s.repo.Create(ctx, post)
	if err != nil {
		return nil, err
	}

	createdPost, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return createdPost, nil
}

func (s *PostServiceImpl) GetPost(ctx context.Context, id int64) (*domain.Post, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *PostServiceImpl) GetAllPosts(ctx context.Context) ([]*domain.Post, error) {
	return s.repo.GetAll(ctx)
}

func (s *PostServiceImpl) DeletePost(ctx context.Context, id, authorID int64) error {
	return s.repo.Delete(ctx, id, authorID)
}
func (s *PostServiceImpl) GetPostsWithAuthors(ctx context.Context) ([]*domain.Post, error) {
	return s.repo.GetPostsWithAuthors(ctx)
}
