package cmd

import (
	"github.com/Frozz164/forum-app_v2/forum-service/internal/domain"
	"github.com/Frozz164/forum-app_v2/forum-service/internal/handler"
	"github.com/Frozz164/forum-app_v2/forum-service/internal/repository"
	"github.com/Frozz164/forum-app_v2/forum-service/internal/service"
	"github.com/Frozz164/forum-app_v2/forum-service/pkg/middleware"
	"github.com/Frozz164/forum-app_v2/forum-service/pkg/websocket"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http"
)

func main() {
	// Инициализация БД
	db, err := gorm.Open(sqlite.Open("forum.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Автомиграция
	db.AutoMigrate(&domain.Post{}, &domain.Message{})

	// Инициализация слоев
	postRepo := repository.NewPostRepository(db)
	postService := service.NewPostService(postRepo)
	postHandler := handler.NewPostHandler(postService)

	chatPool := websocket.NewPool()
	go chatPool.Start()

	// Настройка роутера
	r := gin.Default()

	// Статические файлы
	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("web/templates/*")

	// API для постов
	api := r.Group("/api")
	{
		api.Use(middleware.AuthMiddleware())
		api.POST("/posts", postHandler.CreatePost)
		api.GET("/posts", postHandler.GetAllPosts)
		api.DELETE("/posts/:id", postHandler.DeletePost)
	}

	// WebSocket endpoint
	r.GET("/ws", func(c *gin.Context) {
		handler.WebsocketHandler(c, chatPool)
	})

	// HTML страницы
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.GET("/chat", func(c *gin.Context) {
		c.HTML(http.StatusOK, "chat.html", nil)
	})

	r.Run(":8080")
}
