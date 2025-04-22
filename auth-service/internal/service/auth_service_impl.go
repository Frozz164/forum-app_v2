package service

import (
	"context"
	"fmt"
	"github.com/Frozz164/forum-app_v2/auth-service/config"
	"github.com/Frozz164/forum-app_v2/auth-service/internal/domain"
	"github.com/Frozz164/forum-app_v2/auth-service/internal/repository"
	"github.com/Frozz164/forum-app_v2/auth-service/pkg/helper"
	"log"
)

// AuthServiceImpl ...
type AuthServiceImpl struct {
	authRepository repository.AuthRepository
	cfg            *config.Config
}

// NewAuthServiceImpl ...
func NewAuthServiceImpl(authRepository repository.AuthRepository, cfg *config.Config) AuthService {
	return &AuthServiceImpl{authRepository: authRepository, cfg: cfg}
}

// CreateUser ...
func (s *AuthServiceImpl) CreateUser(ctx context.Context, username, password, email string) (int64, error) {
	hashedPassword, err := helper.GeneratePassword(password)
	if err != nil {
		log.Printf("Error generating password hash: %v", err)
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.User{
		Username: username,
		Password: hashedPassword,
		Email:    email,
	}

	id, err := s.authRepository.Create(ctx, user)
	if err != nil {
		log.Printf("Error creating user in repository: %v", err)
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	log.Printf("User created with ID: %d", id)
	return id, nil
}

// Login ...
func (s *AuthServiceImpl) Login(ctx context.Context, username, password string) (string, error) {
	user, err := s.authRepository.GetByUsername(ctx, username)
	if err != nil {
		log.Printf("Error getting user by username: %v", err)
		return "", fmt.Errorf("failed to get user by username: %w", err)
	}

	if user == nil {
		return "", fmt.Errorf("invalid credentials")
	}

	err = helper.ComparePasswords(user.Password, password)
	if err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	token, err := helper.GenerateJWT(user.ID, s.cfg.JWT.SecretKey, s.cfg.JWT.ExpiresIn)
	if err != nil {
		log.Printf("Error generating JWT: %v", err)
		return "", fmt.Errorf("failed to generate JWT: %w", err)
	}

	return token, nil
}

// ValidateToken ...
func (s *AuthServiceImpl) ValidateToken(token string) (int64, error) {
	userID, err := helper.ValidateToken(token, s.cfg.JWT.SecretKey)
	if err != nil {
		log.Printf("Error validating token: %v", err)
		return 0, fmt.Errorf("invalid token: %w", err)
	}

	return userID, nil
}
