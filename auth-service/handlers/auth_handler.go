package handlers

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/Frozz164/forum-app_v2/auth-service/config"
	"github.com/Frozz164/forum-app_v2/auth-service/internal/service"
	"github.com/Frozz164/forum-app_v2/auth-service/model"
	"github.com/Frozz164/forum-app_v2/auth-service/pkg/helper"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type AuthServiceHandler struct {
	cfg         *config.Config
	authService service.AuthService
	validator   *validator.Validate
	logger      zerolog.Logger
}

func NewAuthServiceHandler(cfg *config.Config, authService service.AuthService) *AuthServiceHandler {
	v := validator.New()
	_ = v.RegisterValidation("username", validateUsername)
	return &AuthServiceHandler{
		cfg:         cfg,
		authService: authService,
		validator:   v,
		logger:      log.With().Str("component", "auth_handler").Logger(),
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
	logger := h.logger.With().Str("method", "Register").Logger()

	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn().Err(err).Msg("Invalid request format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	logger = logger.With().
		Str("username", req.Username).
		Str("email", req.Email).
		Logger()

	// Дополнительная валидация
	if err := h.validator.Struct(req); err != nil {
		logger.Warn().Err(err).Msg("Validation failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверка email
	if !strings.Contains(req.Email, "@") {
		logger.Warn().Msg("Invalid email format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	userID, err := h.authService.CreateUser(c.Request.Context(), req.Username, req.Password, req.Email)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	token, err := h.generateToken(userID)
	if err != nil {
		logger.Error().
			Err(err).
			Int64("user_id", userID).
			Msg("Failed to generate token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	logger.Info().
		Int64("user_id", userID).
		Msg("User registered successfully")

	c.JSON(http.StatusCreated, gin.H{
		"user_id":      userID,
		"access_token": token,
		"username":     req.Username,
	})
}

func (h *AuthServiceHandler) Login(c *gin.Context) {
	logger := h.logger.With().Str("method", "Login").Logger()

	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn().Err(err).Msg("Invalid request format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	logger = logger.With().
		Str("username", req.Username).
		Str("email", req.Email).
		Logger()

	var token string
	var err error

	if req.Username != "" {
		token, err = h.authService.Login(c.Request.Context(), req.Username, req.Password)
	} else if req.Email != "" {
		token, err = h.authService.LoginByEmail(c.Request.Context(), req.Email, req.Password)
	} else {
		logger.Warn().Msg("Username or email required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username or email required"})
		return
	}

	if err != nil {
		logger.Warn().Err(err).Msg("Invalid credentials")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	logger.Info().Msg("User logged in successfully")

	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   h.cfg.JWT.ExpiresIn,
	})
}

func (h *AuthServiceHandler) generateToken(userID int64) (string, error) {
	h.logger.Debug().
		Int64("user_id", userID).
		Msg("Generating JWT token")
	return helper.GenerateJWT(userID, h.cfg.JWT.SecretKey, strconv.Itoa(h.cfg.JWT.ExpiresIn))
}

func (h *AuthServiceHandler) Validate(c *gin.Context) {
	logger := h.logger.With().Str("method", "Validate").Logger()

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		logger.Warn().Msg("Authorization header is missing")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
		return
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		logger.Warn().Str("header", authHeader).Msg("Invalid authorization format")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
		return
	}

	token := tokenParts[1]
	logger = logger.With().Str("token_prefix", token[:6]).Logger()

	userID, err := h.authService.ValidateToken(token)
	if err != nil {
		logger.Error().Err(err).Msg("Token validation failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	logger.Info().
		Int64("user_id", userID).
		Msg("Token validated successfully")

	c.JSON(http.StatusOK, gin.H{
		"user_id":  userID,
		"is_valid": true,
	})
}
