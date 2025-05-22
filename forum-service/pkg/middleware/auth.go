package middleware

import (
	"fmt"
	"github.com/Frozz164/forum-app_v2/auth-service/pkg/helper"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"strings"
)

func AuthMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := log.With().
			Str("middleware", "AuthMiddleware").
			Str("path", c.Request.URL.Path).
			Logger()

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn().Msg("Authorization header is missing")
			c.AbortWithStatusJSON(401, gin.H{"error": "Authorization header required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Warn().
				Str("header", authHeader).
				Msg("Invalid authorization header format")
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid authorization header format"})
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			logger.Warn().Msg("Empty token string")
			c.AbortWithStatusJSON(401, gin.H{"error": "Token required"})
			return
		}

		logger = logger.With().
			Str("token_prefix", tokenString[:min(10, len(tokenString))]).
			Logger()

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				logger.Warn().
					Str("alg", fmt.Sprintf("%v", token.Header["alg"])).
					Msg("Unexpected signing method")
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secretKey), nil
		})

		if err != nil {
			logger.Warn().
				Err(err).
				Msg("Token validation failed")
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
			return
		}

		if !token.Valid {
			logger.Warn().Msg("Invalid token provided")
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
			return
		}

		logger.Debug().Msg("Token validated successfully")
		c.Next()
	}
}

func AuthWebSocketMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := log.With().
			Str("middleware", "AuthWebSocketMiddleware").
			Str("path", c.Request.URL.Path).
			Logger()

		// 1. Получаем токен из query параметров
		token := c.Query("token")
		if token == "" {
			// 2. Пробуем получить из заголовков (для совместимости)
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				logger.Debug().Msg("Token is missing in both query and header")
				c.AbortWithStatusJSON(401, gin.H{"error": "token required"})
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				logger.Warn().
					Str("header", authHeader).
					Msg("Invalid authorization header format")
				c.AbortWithStatusJSON(401, gin.H{"error": "invalid auth format"})
				return
			}
			token = parts[1]
		}

		// 3. Валидация токена
		claims, err := helper.ValidateTokenWithClaims(token, secretKey)
		if err != nil {
			logger.Warn().
				Err(err).
				Str("token_prefix", token[:min(10, len(token))]).
				Msg("WebSocket token validation failed")
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}

		// 4. Сохраняем данные в контекст
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)

		logger.Info().
			Str("username", claims.Username).
			Int64("user_id", claims.UserID).
			Msg("WebSocket authentication successful")

		c.Next()
	}
}
