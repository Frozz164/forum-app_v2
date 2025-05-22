package main

import (
	"github.com/Frozz164/forum-app_v2/forum-service/config"
	"github.com/Frozz164/forum-app_v2/forum-service/internal/handler"
	"github.com/Frozz164/forum-app_v2/forum-service/internal/migrations"
	"github.com/Frozz164/forum-app_v2/forum-service/internal/repository"
	"github.com/Frozz164/forum-app_v2/forum-service/internal/service"
	"github.com/Frozz164/forum-app_v2/forum-service/pkg/middleware"
	"github.com/Frozz164/forum-app_v2/forum-service/pkg/websocket"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

func initLogger() {
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	log.Logger = log.Output(output).With().Caller().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func main() {
	initLogger()
	log.Info().Msg("Starting forum service")

	cfg := config.Load()
	log.Debug().Interface("config", cfg).Msg("Configuration loaded")

	// Database connection
	db, err := cfg.Database.Connect()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close database connection")
		}
	}()

	// Run migrations
	log.Info().Msg("Running database migrations")
	if err := migrations.MigrateDB(db); err != nil {
		log.Fatal().Err(err).Msg("Failed to migrate database")
	}

	// Initialize layers
	log.Info().Msg("Initializing application layers")
	postRepo := repository.NewPostRepository(db)
	chatRepo := repository.NewChatRepository(db)

	postService := service.NewPostService(postRepo)
	chatService := service.NewChatService(chatRepo)

	pool := websocket.NewPool(chatService)
	go pool.Start()

	postHandler := handler.NewPostHandler(postService)
	chatHandler := handler.NewChatHandler(chatService, pool, cfg.JWT.SecretKey)

	// Gin setup
	router := gin.Default()
	router.Use(middleware.GinLogger())

	// CORS configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowWebSockets:  true,
		MaxAge:           12 * time.Hour,
	}))

	router.Static("/static", "../web")
	router.StaticFile("/", "../web/index.html")

	// Public routes
	router.GET("/api/posts", postHandler.GetAllPosts)
	router.GET("/api/posts/:id", postHandler.GetPost)
	router.GET("/ws", middleware.AuthWebSocketMiddleware(cfg.JWT.SecretKey), chatHandler.WebsocketHandler)

	// Protected routes
	authGroup := router.Group("/api")
	authGroup.Use(middleware.AuthMiddleware(cfg.JWT.SecretKey))
	{
		authGroup.POST("/posts", postHandler.CreatePost)
		authGroup.DELETE("/posts/:id", postHandler.DeletePost)
	}

	// Start server
	log.Info().Str("port", cfg.Port).Msg("Starting HTTP server")
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}
