package helper

import (
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
func GenerateJWT(userID int64, secretKey string, expiresIn int) (string, error) {
	expirationTime := time.Now().Add(time.Duration(expiresIn) * time.Second)
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     expirationTime.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		log.Printf("Error signing token: %v", err)
		return "", err
	}
	return tokenString, nil
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
