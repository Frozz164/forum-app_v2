package service

import (
	"context"
	"fmt"
	"github.com/Frozz164/forum-app_v2/auth-service/config"
	"github.com/Frozz164/forum-app_v2/auth-service/internal/domain"
	"github.com/Frozz164/forum-app_v2/auth-service/internal/repository"
	"github.com/Frozz164/forum-app_v2/auth-service/pkg/helper"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"strconv"
)

type AuthServiceImpl struct {
	authRepository repository.AuthRepository
	cfg            *config.Config
	logger         zerolog.Logger
}

func NewAuthServiceImpl(authRepository *repository.AuthRepositoryImpl, cfg *config.Config) AuthService {
	return &AuthServiceImpl{
		authRepository: authRepository,
		cfg:            cfg,
		logger:         log.With().Str("component", "auth_service").Logger(),
	}
}

func (s *AuthServiceImpl) LoginByEmail(ctx context.Context, email string, password string) (string, error) {
	s.logger.Info().Str("email", email).Msg("LoginByEmail called")
	// Реализация остается прежней
	panic("implement me")
}

func (s *AuthServiceImpl) CreateUser(ctx context.Context, username, password, email string) (int64, error) {
	s.logger.Info().
		Str("username", username).
		Str("email", email).
		Msg("CreateUser called")

	hashedPassword, err := helper.GeneratePassword(password)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error generating password hash")
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.User{
		Username: username,
		Password: hashedPassword,
		Email:    email,
	}

	id, err := s.authRepository.Create(ctx, user)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error creating user in repository")
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info().Int64("user_id", id).Msg("User created successfully")
	return id, nil
}

func (s *AuthServiceImpl) Login(ctx context.Context, username, password string) (string, error) {
	s.logger.Info().Str("username", username).Msg("Login called")

	user, err := s.authRepository.GetByUsername(ctx, username)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error getting user by username")
		return "", fmt.Errorf("failed to get user by username: %w", err)
	}

	if user == nil {
		s.logger.Warn().Msg("User not found - invalid credentials")
		return "", fmt.Errorf("invalid credentials")
	}

	err = helper.ComparePasswords(user.Password, password)
	if err != nil {
		s.logger.Warn().Msg("Password mismatch - invalid credentials")
		return "", fmt.Errorf("invalid credentials")
	}

	token, err := helper.GenerateJWT(user.ID, s.cfg.JWT.SecretKey, strconv.Itoa(s.cfg.JWT.ExpiresIn))
	if err != nil {
		s.logger.Error().Err(err).Msg("Error generating JWT")
		return "", fmt.Errorf("failed to generate JWT: %w", err)
	}

	s.logger.Info().Msg("Login successful")
	return token, nil
}

func (s *AuthServiceImpl) ValidateToken(token string) (int64, error) {
	s.logger.Info().Msg("ValidateToken called")

	userID, err := helper.ValidateToken(token, s.cfg.JWT.SecretKey)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error validating token")
		return 0, fmt.Errorf("invalid token: %w", err)
	}

	s.logger.Info().Int64("user_id", userID).Msg("Token validated successfully")
	return userID, nil
}
