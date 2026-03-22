package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger는 구조화된 HTTP 요청 로깅 미들웨어입니다.
func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		fields := []zap.Field{
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
		}

		if query != "" {
			fields = append(fields, zap.String("query", query))
		}
		if userID, exists := c.Get(ContextKeyUserID); exists {
			fields = append(fields, zap.Any("user_id", userID))
		}
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		switch {
		case status >= 500:
			logger.Error("server error", fields...)
		case status >= 400:
			logger.Warn("client error", fields...)
		default:
			logger.Info("request", fields...)
		}
	}
}
