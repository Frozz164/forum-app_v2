package main

import (
	"github.com/Frozz164/forum-app_v2/forum-service/pkg/middleware"
	"log"
	"time"

	"github.com/Frozz164/forum-app_v2/forum-service/config"
	"github.com/Frozz164/forum-app_v2/forum-service/internal/migrations"
	"github.com/Frozz164/forum-app_v2/forum-service/internal/repository"
	"github.com/Frozz164/forum-app_v2/forum-service/internal/service"
	"github.com/Frozz164/forum-app_v2/forum-service/pkg/websocket"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()

	// Initialize database connection
	db, err := cfg.Database.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Migrate the database
	err = migrations.MigrateDB(db)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize repositories
	postRepo := repository.NewPostRepository(db)
	chatRepo := repository.NewChatRepository(db)

	// Initialize services
	postService := service.NewPostService(postRepo)
	chatService := service.NewChatService(chatRepo)

	// Initialize WebSocket pool
	pool := websocket.NewPool()
	go pool.Start()

	// Initialize handlers
	postHandler := handler.NewPostHandler(postService)
	chatHandler := handler.NewChatHandler(chatService, pool)

	// Gin setup
	router := gin.Default()

	// CORS configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Public routes
	router.GET("/api/posts", postHandler.GetAllPosts)
	router.GET("/api/posts/:id", postHandler.GetPost)
	router.GET("/ws", chatHandler.WebsocketHandler)

	// Authenticated routes
	authGroup := router.Group("/api")
	authGroup.Use(middleware.AuthMiddleware(cfg.JWT.SecretKey))
	{
		authGroup.POST("/posts", postHandler.CreatePost)
		authGroup.DELETE("/posts/:id", postHandler.DeletePost)
	}

	// Start server
	port := cfg.Port
	if port == "" {
		port = "8081" // Different port from auth service
	}
	log.Printf("Forum service listening on :%s", port)
	router.Run(":" + port)
}
