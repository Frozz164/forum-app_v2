package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Frozz164/forum-app_v2/auth-service/internal/domain"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type AuthRepositoryImpl struct {
	db     *sql.DB
	logger zerolog.Logger
}

func (r *AuthRepositoryImpl) LoginByEmail(ctx context.Context, email, password string) (string, error) {
	r.logger.Info().Str("email", email).Msg("LoginByEmail called")
	// Реализация остается прежней
	panic("implement me")
}

func (r *AuthRepositoryImpl) GetByEmail(ctx context.Context, email string) (interface{}, interface{}) {
	r.logger.Info().Str("email", email).Msg("GetByEmail called")
	// Реализация остается прежней
	panic("implement me")
}

func NewAuthRepositoryImpl(db *sql.DB) *AuthRepositoryImpl {
	return &AuthRepositoryImpl{
		db:     db,
		logger: log.With().Str("component", "auth_repository").Logger(),
	}
}

func (r *AuthRepositoryImpl) Create(ctx context.Context, user *domain.User) (int64, error) {
	r.logger.Info().
		Str("username", user.Username).
		Str("email", user.Email).
		Msg("Create user called")

	query := `
		INSERT INTO users (username, password, email)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(ctx, query, user.Username, user.Password, user.Email).Scan(&id)
	if err != nil {
		r.logger.Error().Err(err).Msg("Error creating user in database")
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	r.logger.Info().Int64("user_id", id).Msg("User created in database")
	return id, nil
}

func (r *AuthRepositoryImpl) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	r.logger.Info().Str("username", username).Msg("GetByUsername called")

	query := `
		SELECT id, username, password, email
		FROM users
		WHERE username = $1
	`

	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(&user.ID, &user.Username, &user.Password, &user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Info().Msg("User not found by username")
			return nil, nil
		}
		r.logger.Error().Err(err).Msg("Error getting user by username")
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	r.logger.Info().Int64("user_id", user.ID).Msg("User found by username")
	return user, nil
}

func (r *AuthRepositoryImpl) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	r.logger.Info().Int64("user_id", id).Msg("GetByID called")

	query := `
		SELECT id, username, password, email
		FROM users
		WHERE id = $1
	`

	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Username, &user.Password, &user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Info().Msg("User not found by ID")
			return nil, nil
		}
		r.logger.Error().Err(err).Msg("Error getting user by ID")
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	r.logger.Info().Int64("user_id", user.ID).Msg("User found by ID")
	return user, nil
}
