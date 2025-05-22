package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		timestamp := time.Now()
		latency := timestamp.Sub(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		logger := log.With().
			Int("status", status).
			Str("method", method).
			Str("path", path).
			Str("query", raw).
			Str("ip", clientIP).
			Dur("latency", latency).
			Logger()

		switch {
		case status >= 500:
			logger.Error().
				Str("error", errorMessage).
				Msg("Server error")
		case status >= 400:
			logger.Warn().
				Str("error", errorMessage).
				Msg("Client error")
		default:
			logger.Info().Msg("Request processed")
		}
	}
}
