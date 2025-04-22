package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/Frozz164/forum-app_v2/auth-service/internal/domain"
)

// AuthRepositoryImpl ...
type AuthRepositoryImpl struct {
	db *sql.DB
}

func (r *AuthRepositoryImpl) LoginByEmail(ctx context.Context, email, password string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (r *AuthRepositoryImpl) GetByEmail(ctx context.Context, email string) (interface{}, interface{}) {
	//TODO implement me
	panic("implement me")
}

// NewAuthRepositoryImpl ...
func NewAuthRepositoryImpl(db *sql.DB) *AuthRepositoryImpl {
	return &AuthRepositoryImpl{db: db}
}

// Create ...
func (r *AuthRepositoryImpl) Create(ctx context.Context, user *domain.User) (int64, error) {
	query := `
		INSERT INTO users (username, password, email)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(ctx, query, user.Username, user.Password, user.Email).Scan(&id)
	if err != nil {
		log.Printf("Error creating user in database: %v", err)
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	log.Printf("User created with ID: %d", id)
	return id, nil
}

// GetByUsername ...
func (r *AuthRepositoryImpl) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, username, password, email
		FROM users
		WHERE username = $1
	`

	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(&user.ID, &user.Username, &user.Password, &user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		log.Printf("Error getting user by username: %v", err)
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return user, nil
}

// GetByID ...
func (r *AuthRepositoryImpl) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `
		SELECT id, username, password, email
		FROM users
		WHERE id = $1
	`

	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Username, &user.Password, &user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		log.Printf("Error getting user by ID: %v", err)
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}
