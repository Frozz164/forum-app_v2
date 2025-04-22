package handlers

import (
	_ "fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/Frozz164/forum-app_v2/auth-service/config"
	_ "github.com/Frozz164/forum-app_v2/auth-service/internal/domain"
	"github.com/Frozz164/forum-app_v2/auth-service/internal/service"
	"github.com/Frozz164/forum-app_v2/auth-service/model"
	"github.com/Frozz164/forum-app_v2/auth-service/pkg/helper"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AuthServiceHandler struct {
	cfg         *config.Config
	authService service.AuthService
	validator   *validator.Validate
}

func NewAuthServiceHandler(cfg *config.Config, authService service.AuthService) *AuthServiceHandler {
	v := validator.New()
	_ = v.RegisterValidation("username", validateUsername)
	return &AuthServiceHandler{
		cfg:         cfg,
		authService: authService,
		validator:   v,
	}
}

func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	if len(username) < 3 || len(username) > 20 {
		return false
	}
	return regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(username)
}

func (h *AuthServiceHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Дополнительная валидация
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверка email
	if !strings.Contains(req.Email, "@") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	userID, err := h.authService.CreateUser(c.Request.Context(), req.Username, req.Password, req.Email)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	token, err := h.generateToken(userID)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user_id":      userID,
		"access_token": token,
		"username":     req.Username,
	})
}

func (h *AuthServiceHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	var token string
	var err error

	if req.Username != "" {
		token, err = h.authService.Login(c.Request.Context(), req.Username, req.Password)
	} else if req.Email != "" {
		// Реализуйте метод LoginByEmail в сервисе
		token, err = h.authService.LoginByEmail(c.Request.Context(), req.Email, req.Password)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username or email required"})
		return
	}

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   h.cfg.JWT.ExpiresIn,
	})
}

func (h *AuthServiceHandler) generateToken(userID int64) (string, error) {
	return helper.GenerateJWT(userID, h.cfg.JWT.SecretKey, strconv.Itoa(h.cfg.JWT.ExpiresIn))
}
func (h *AuthServiceHandler) Validate(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
		return
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
		return
	}

	token := tokenParts[1]
	userID, err := h.authService.ValidateToken(token)
	if err != nil {
		log.Printf("Token validation failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":  userID,
		"is_valid": true,
	})
}
