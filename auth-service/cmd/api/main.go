package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Frozz164/forum-app_v2/auth-service/config"
	"github.com/Frozz164/forum-app_v2/auth-service/handlers"
	"github.com/Frozz164/forum-app_v2/auth-service/internal/migrations"
	"github.com/Frozz164/forum-app_v2/auth-service/internal/repository"
	"github.com/Frozz164/forum-app_v2/auth-service/internal/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	// Загрузка конфигурации
	cfg := config.Load()

	// Инициализация подключения к базе данных
	db, err := cfg.Database.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	// Применение миграций
	err = migrations.MigrateDB(db)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migrations applied successfully")

	// Инициализация слоев приложения
	authRepo := repository.NewAuthRepositoryImpl(db)
	authService := service.NewAuthServiceImpl(authRepo, cfg)
	authHandler := handlers.NewAuthServiceHandler(cfg, authService)

	// Настройка роутера Gin
	router := gin.Default()

	// Настройка CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8081"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Маршруты API
	api := router.Group("/api/v1")
	{
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)
		api.GET("/validate", authHandler.Validate)
	}

	// Обслуживание статических файлов (если нужно)
	router.Static("/static", "../web")          // CSS/JS/Images
	router.StaticFile("/", "../web/index.html") // Главная страница

	// Запуск сервера
	port := cfg.Port
	if port == "" {
		port = "8080" // Порт по умолчанию
	}
	addr := fmt.Sprintf(":%s", port)

	log.Printf("Starting auth service on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	api.OPTIONS("/register", func(c *gin.Context) {
		c.Status(200)
	})
}
