package helper

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// GeneratePassword ...
func GeneratePassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error generating password hash: %v", err)
		return "", err
	}
	return string(hashedPassword), nil
}

// ComparePasswords ...
func ComparePasswords(hashedPassword string, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		log.Printf("Error comparing passwords: %v", err)
		return fmt.Errorf("invalid credentials")
	}
	return nil
}

// ValidateToken ...
func ValidateToken(tokenString, secretKey string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		log.Printf("Error parsing token: %v", err)
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			return 0, fmt.Errorf("invalid user ID in token")
		}
		userID := int64(userIDFloat)
		return userID, nil
	}

	return 0, fmt.Errorf("invalid token")
}

// GenerateJWT ...
func GenerateJWT(userID int64, secretKey string, expiresIn string) (string, error) {
	if secretKey == "" {
		return "", errors.New("secret key cannot be empty")
	}

	expiresInSeconds, err := strconv.ParseInt(expiresIn, 10, 64)
	if err != nil {
		return "", fmt.Errorf("invalid expiresIn: %w", err)
	}
	if expiresInSeconds <= 0 {
		return "", errors.New("expiresIn must be positive")
	}

	expirationTime := time.Now().Add(time.Duration(expiresInSeconds) * time.Second)

	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     expirationTime.Unix(),
		"iss":     "forum-app",
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// GetEnv ...
func GetEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Convert string to int
func StringToInt(str string) (int, error) {
	num, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("cannot convert string to integer: %w", err)
	}
	return num, nil
}

var ErrInvalidToken = errors.New("invalid token")

type CustomClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func ValidateTokenWithClaims(tokenString, secretKey string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

func GenerateJWTWithClaims(userID int64, username, secretKey string, expiresIn int) (string, error) {
	claims := CustomClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiresIn) * time.Second)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}
