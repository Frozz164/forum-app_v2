package handler

import (
	"net/http"
	"strconv"

	"github.com/Frozz164/forum-app_v2/forum-service/internal/domain"
	"github.com/Frozz164/forum-app_v2/forum-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type PostHandler struct {
	service service.PostService
	logger  zerolog.Logger
}

func NewPostHandler(service service.PostService) *PostHandler {
	return &PostHandler{
		service: service,
		logger:  log.With().Str("component", "post_handler").Logger(),
	}
}

func (h *PostHandler) CreatePost(c *gin.Context) {
	logger := h.logger.With().Str("method", "CreatePost").Logger()

	// Получаем ID автора из JWT
	authorID, exists := c.Get("userID")
	if !exists {
		logger.Warn().Msg("Unauthorized attempt to create post")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var post domain.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		logger.Warn().Err(err).Msg("Invalid request format")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post.AuthorID = authorID.(int64)
	logger = logger.With().Int64("author_id", post.AuthorID).Str("title", post.Title).Logger()

	createdPost, err := h.service.CreatePost(c.Request.Context(), &post)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create post")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info().Int64("post_id", createdPost.ID).Msg("Post created successfully")
	c.JSON(http.StatusCreated, createdPost)
}

func (h *PostHandler) GetPost(c *gin.Context) {
	logger := h.logger.With().Str("method", "GetPost").Logger()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Warn().Err(err).Str("post_id_param", c.Param("id")).Msg("Invalid post ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	logger = logger.With().Int64("post_id", id).Logger()
	post, err := h.service.GetPost(c.Request.Context(), id)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get post")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if post == nil {
		logger.Debug().Msg("Post not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}

	logger.Debug().Msg("Post retrieved successfully")
	c.JSON(http.StatusOK, post)
}

func (h *PostHandler) GetAllPosts(c *gin.Context) {
	logger := h.logger.With().Str("method", "GetAllPosts").Logger()

	posts, err := h.service.GetAllPosts(c.Request.Context())
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get all posts")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Debug().Int("post_count", len(posts)).Msg("Retrieved all posts")
	c.JSON(http.StatusOK, posts)
}

func (h *PostHandler) DeletePost(c *gin.Context) {
	logger := h.logger.With().Str("method", "DeletePost").Logger()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Warn().Err(err).Str("post_id_param", c.Param("id")).Msg("Invalid post ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	authorID, exists := c.Get("userID")
	if !exists {
		logger.Warn().Msg("Unauthorized attempt to delete post")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	logger = logger.With().Int64("post_id", id).Int64("author_id", authorID.(int64)).Logger()
	err = h.service.DeletePost(c.Request.Context(), id, authorID.(int64))
	if err != nil {
		logger.Error().Err(err).Msg("Failed to delete post")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info().Msg("Post deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "post deleted successfully"})
}
