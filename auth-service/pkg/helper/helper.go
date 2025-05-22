package helper

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	// Настройка логгера для пакета helper
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func maskSensitive(data string) string {
	if len(data) < 6 {
		return "***"
	}
	return data[:3] + "***" + data[len(data)-3:]
}

// GeneratePassword хеширует пароль с логированием
func GeneratePassword(password string) (string, error) {
	log.Debug().Msg("Starting password generation")
	start := time.Now()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().
			Err(err).
			Str("operation", "password_hash").
			Dur("duration", time.Since(start)).
			Msg("Failed to generate password hash")
		return "", err
	}

	log.Info().
		Dur("duration", time.Since(start)).
		Msg("Password hashed successfully")
	return string(hashedPassword), nil
}

// ComparePasswords сравнивает хеш пароля с логированием
func ComparePasswords(hashedPassword string, password string) error {
	log.Debug().
		Str("hash_prefix", maskSensitive(hashedPassword)).
		Msg("Starting password comparison")

	start := time.Now()
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		log.Warn().
			Err(err).
			Str("operation", "password_compare").
			Dur("duration", time.Since(start)).
			Msg("Password comparison failed")
		return fmt.Errorf("invalid credentials")
	}

	log.Debug().
		Dur("duration", time.Since(start)).
		Msg("Password comparison successful")
	return nil
}
func ValidateToken(tokenString, secretKey string) (int64, error) {
	log.Debug().
		Str("token_prefix", maskSensitive(tokenString)).
		Msg("Starting token validation")

	start := time.Now()
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Error().
				Str("alg", token.Header["alg"].(string)).
				Msg("Unexpected signing method")
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		log.Error().
			Err(err).
			Str("operation", "token_validation").
			Dur("duration", time.Since(start)).
			Msg("Token parsing failed")
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			log.Error().
				Interface("claims", claims).
				Msg("Invalid user_id in token claims")
			return 0, fmt.Errorf("invalid user ID in token")
		}

		log.Info().
			Int64("user_id", int64(userIDFloat)).
			Dur("duration", time.Since(start)).
			Msg("Token validated successfully")
		return int64(userIDFloat), nil
	}

	log.Error().
		Dur("duration", time.Since(start)).
		Msg("Invalid token claims")
	return 0, fmt.Errorf("invalid token")
}

// GenerateJWT генерирует JWT токен с логированием
func GenerateJWT(userID int64, secretKey string, expiresIn string) (string, error) {
	log.Debug().
		Int64("user_id", userID).
		Str("expires_in", expiresIn).
		Msg("Starting JWT generation")

	start := time.Now()
	expiresInSeconds, err := strconv.ParseInt(expiresIn, 10, 64)
	if err != nil {
		log.Error().
			Err(err).
			Str("expires_in", expiresIn).
			Msg("Failed to parse expires_in")
		return "", fmt.Errorf("invalid expires_in value")
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Second * time.Duration(expiresInSeconds)).Unix(),
		"iss":     "auth-service",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		log.Error().
			Err(err).
			Int64("user_id", userID).
			Dur("duration", time.Since(start)).
			Msg("Failed to sign JWT token")
		return "", err
	}

	log.Info().
		Int64("user_id", userID).
		Str("token_prefix", maskSensitive(tokenString)).
		Dur("duration", time.Since(start)).
		Msg("JWT generated successfully")
	return tokenString, nil
}

// ValidateTokenWithClaims валидирует токен с claims и логированием
func ValidateTokenWithClaims(tokenString, secretKey string) (*CustomClaims, error) {
	log.Debug().
		Str("token_prefix", maskSensitive(tokenString)).
		Msg("Starting token validation with claims")

	start := time.Now()
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Error().
				Str("alg", token.Header["alg"].(string)).
				Msg("Unexpected signing method in token")
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		log.Error().
			Err(err).
			Str("operation", "token_validation").
			Dur("duration", time.Since(start)).
			Msg("Token parsing failed")
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		log.Info().
			Int64("user_id", claims.UserID).
			Str("username", claims.Username).
			Dur("duration", time.Since(start)).
			Msg("Token with claims validated successfully")
		return claims, nil
	}

	log.Error().
		Dur("duration", time.Since(start)).
		Msg("Invalid token claims")
	return nil, ErrInvalidToken
}

// GenerateJWTWithClaims генерирует JWT с claims и логированием
func GenerateJWTWithClaims(userID int64, username, secretKey string, expiresIn int) (string, error) {
	log.Debug().
		Int64("user_id", userID).
		Str("username", username).
		Int("expires_in_sec", expiresIn).
		Msg("Starting JWT generation with claims")

	start := time.Now()
	claims := CustomClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiresIn) * time.Second)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		log.Error().
			Err(err).
			Int64("user_id", userID).
			Dur("duration", time.Since(start)).
			Msg("Failed to sign JWT token with claims")
		return "", err
	}

	log.Info().
		Int64("user_id", userID).
		Str("username", username).
		Str("token_prefix", maskSensitive(tokenString)).
		Dur("duration", time.Since(start)).
		Msg("JWT with claims generated successfully")
	return tokenString, nil
}

func GetEnv(key string, defaultValue string) string {
	return os.Getenv(key)
}

func StringToInt(str string) (int, error) {
	return strconv.Atoi(str)
}

var ErrInvalidToken = errors.New("invalid token")

type CustomClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}
