package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS는 CORS 정책을 설정하는 미들웨어입니다.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowedOriginsSet := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowedOriginsSet[o] = true
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if allowedOriginsSet[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Tenant-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
