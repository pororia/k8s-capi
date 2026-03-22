package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/k8s-cmp/k8s-cmp/internal/service"
	apperrors "github.com/k8s-cmp/k8s-cmp/pkg/errors"
	"github.com/k8s-cmp/k8s-cmp/pkg/response"
)

const (
	ContextKeyUserID   = "user_id"
	ContextKeyTenantID = "tenant_id"
	ContextKeyRole     = "role"
	ContextKeyClaims   = "claims"
)

// Auth는 JWT 액세스 토큰을 검증하는 미들웨어입니다.
func Auth(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Abort(c, apperrors.ErrUnauthorized("authorization header required"))
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Abort(c, apperrors.ErrUnauthorized("invalid authorization header format"))
			return
		}

		claims, err := authSvc.ValidateAccessToken(parts[1])
		if err != nil {
			response.Abort(c, apperrors.ErrUnauthorized("invalid or expired token"))
			return
		}

		c.Set(ContextKeyClaims, claims)
		c.Set(ContextKeyUserID, claims.UserID.String())
		c.Set(ContextKeyRole, string(claims.Role))
		if claims.TenantID != nil {
			c.Set(ContextKeyTenantID, claims.TenantID.String())
		}

		c.Next()
	}
}
