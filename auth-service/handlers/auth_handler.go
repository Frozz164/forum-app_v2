package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/Frozz164/forum-app_v2/auth-service/config"
	"github.com/Frozz164/forum-app_v2/auth-service/internal/domain"
	"github.com/Frozz164/forum-app_v2/auth-service/internal/service"
	"github.com/Frozz164/forum-app_v2/auth-service/model"
	"github.com/Frozz164/forum-app_v2/auth-service/pkg/helper"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// AuthServiceHandler ...
type AuthServiceHandler struct {
	cfg         *config.Config
	authService service.AuthService
	validator   *validator.Validate
}

// NewAuthServiceHandler ...
func NewAuthServiceHandler(cfg *config.Config, authService service.AuthService) *AuthServiceHandler {
	return &AuthServiceHandler{
		cfg:         cfg,
		authService: authService,
		validator:   validator.New(),
	}
}

// Register ...
func (h *AuthServiceHandler) Register(c *gin.Context) {
	fmt.Println("Register endpoint called")

	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := h.authService.CreateUser(c.Request.Context(), req.Username, req.Password, req.Email)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	token, err := h.generateToken(userID, h.cfg.JWT.SecretKey, strconv.Itoa(h.cfg.JWT.ExpiresIn))
	if err != nil {
		log.Printf("Error generating token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	resp := model.RegisterResponse{
		UserID:      fmt.Sprint(userID),
		AccessToken: token,
	}

	c.JSON(http.StatusCreated, resp)
	fmt.Printf("User %s registered successfully\n", req.Username)
}

// Login ...
func (h *AuthServiceHandler) Login(c *gin.Context) {
	var req domain.User
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": token})
}

// Validate ...
func (h *AuthServiceHandler) Validate(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := h.authService.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_id": userID})
}

// generateToken helper function
func (h *AuthServiceHandler) generateToken(userID int64, secretKey string, expiresIn string) (string, error) {
	expiresInSeconds := h.cfg.JWT.ExpiresIn
	expiresInInt, err := strconv.Atoi(strconv.Itoa(expiresInSeconds))
	if err != nil {
		return "", fmt.Errorf("could not convert JWT_EXPIRES_IN to integer: %w", err)
	}

	token, err := helper.GenerateJWT(userID, secretKey, expiresInInt)
	if err != nil {
		log.Printf("Error generating JWT: %v", err)
		return "", fmt.Errorf("could not generate token: %w", err)
	}

	return token, nil
}
