package main

import (
	_ "errors"
	_ "fmt"
	"github.com/Frozz164/forum-app_v2/auth-service/config"
	"github.com/Frozz164/forum-app_v2/auth-service/handlers"
	"github.com/Frozz164/forum-app_v2/auth-service/internal/migrations"
	"github.com/Frozz164/forum-app_v2/auth-service/internal/repository"
	"github.com/Frozz164/forum-app_v2/auth-service/internal/service"
	"github.com/Frozz164/forum-app_v2/auth-service/pkg/helper"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"strconv"
	"time"
)

func init() {
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func maskToken(token string) string {
	const visibleChars = 6
	if len(token) <= visibleChars*2 {
		return "***"
	}
	return token[:visibleChars] + "***" + token[len(token)-visibleChars:]
}

func main() {
	log.Info().Msg("Starting auth service initialization")

	cfg := config.Load()
	log.Info().
		Str("port", cfg.Port).
		Str("db_host", cfg.Database.Host).
		Msg("Configuration loaded")

	token, err := helper.GenerateJWT(1, cfg.JWT.SecretKey, strconv.Itoa(cfg.JWT.ExpiresIn))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to generate test token")
	}
	log.Info().
		Str("user_id", "1").
		Str("token", maskToken(token)).
		Msg("Test token generated (check at https://jwt.io)")

	db, err := cfg.Database.Connect()
	if err != nil {
		log.Fatal().
			Err(err).
			Str("host", cfg.Database.Host).
			Str("dbname", cfg.Database.Name).
			Msg("Failed to connect to database")
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error().Err(err).Msg("Error closing database connection")
		}
	}()

	log.Info().Msg("Running database migrations...")
	if err := migrations.MigrateDB(db); err != nil {
		log.Fatal().Err(err).Msg("Database migrations failed")
	}
	log.Info().Msg("Migrations completed successfully")

	authRepo := repository.NewAuthRepositoryImpl(db)
	authService := service.NewAuthServiceImpl(authRepo, cfg)
	authHandler := handlers.NewAuthServiceHandler(cfg, authService)

	router := gin.New()
	router.Use(ginLoggerMiddleware())
	router.Use(gin.Recovery())

	// Настройка CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	api := router.Group("/api/v1")
	{
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)
		api.GET("/validate", authHandler.Validate)
	}

	router.Static("/static", "../web")
	router.StaticFile("/", "../web/index.html")

	addr := ":" + cfg.Port
	log.Info().Str("address", addr).Msg("Starting auth service")
	if err := router.Run(addr); err != nil {
		log.Fatal().Err(err).Msg("Server startup failed")
	}
}

func ginLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		c.Next()

		status := c.Writer.Status()
		logEvent := log.Info().
			Str("method", c.Request.Method).
			Str("path", path).
			Str("query", raw).
			Int("status", status).
			Str("client_ip", c.ClientIP()).
			Dur("latency", time.Since(start))

		// Логируем ошибки
		if len(c.Errors) > 0 {
			errorsSlice := make([]error, len(c.Errors))
			for i, ginErr := range c.Errors {
				errorsSlice[i] = ginErr.Err
			}

			logEvent = log.Error().
				Errs(
					"errors",
					errorsSlice,
				)
		} else if status >= 500 {
			logEvent = log.Error()
		} else if status >= 400 {
			logEvent = log.Warn()
		}

		logEvent.Msg("HTTP request")
	}
}
