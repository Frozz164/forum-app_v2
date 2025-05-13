package middleware

import (
	"github.com/Frozz164/forum-app_v2/auth-service/pkg/helper"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strings"
)

func AuthMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			return
		}

		token := parts[1]
		claims, err := helper.ValidateTokenWithClaims(token, secretKey)
		if err != nil {
			log.Printf("Auth error: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
func AuthWebSocketMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем токен из query параметров
		token := c.Query("token")
		if token == "" {
			// Или из заголовков
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			token = parts[1]
		}

		// Валидируем токен
		claims, err := helper.ValidateTokenWithClaims(token, secretKey)
		if err != nil {
			log.Printf("WebSocket auth error: %v", err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Сохраняем данные пользователя в контексте
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
