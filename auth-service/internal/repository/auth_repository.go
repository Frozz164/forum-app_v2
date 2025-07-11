package repository

import (
	"context"

	"github.com/Frozz164/forum-app_v2/auth-service/internal/domain"
)

type AuthRepository interface {
	Create(ctx context.Context, user *domain.User) (int64, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	LoginByEmail(ctx context.Context, email, password string) (string, error)
	GetByEmail(ctx context.Context, email string) (interface{}, interface{})
}
