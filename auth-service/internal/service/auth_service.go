package service

import (
	"context"
)

type AuthService interface {
	CreateUser(ctx context.Context, username, password, email string) (int64, error)
	Login(ctx context.Context, username, password string) (string, error)
	ValidateToken(token string) (int64, error)
	LoginByEmail(ctx context.Context, email string, password string) (string, error)
}
