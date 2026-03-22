package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/k8s-cmp/k8s-cmp/internal/domain"
	apperrors "github.com/k8s-cmp/k8s-cmp/pkg/errors"
	"github.com/k8s-cmp/k8s-cmp/pkg/response"
)

// RequireRole은 최소 역할을 요구하는 미들웨어입니다.
func RequireRole(minRole domain.UserRole) gin.HandlerFunc {
	roleLevel := map[domain.UserRole]int{
		domain.RoleViewer:      1,
		domain.RoleOperator:    2,
		domain.RoleTenantAdmin: 3,
		domain.RoleSuperAdmin:  4,
	}

	return func(c *gin.Context) {
		roleStr, exists := c.Get(ContextKeyRole)
		if !exists {
			response.Abort(c, apperrors.ErrUnauthorized("role not found in context"))
			return
		}

		role := domain.UserRole(roleStr.(string))
		if roleLevel[role] < roleLevel[minRole] {
			response.Abort(c, apperrors.ErrForbidden("insufficient permissions"))
			return
		}

		c.Next()
	}
}

// SuperAdminOnly는 Super Admin만 접근 가능한 미들웨어입니다.
func SuperAdminOnly() gin.HandlerFunc {
	return RequireRole(domain.RoleSuperAdmin)
}

// OperatorOrAbove는 Operator 이상만 접근 가능한 미들웨어입니다.
func OperatorOrAbove() gin.HandlerFunc {
	return RequireRole(domain.RoleOperator)
}

// TenantAdminOrAbove는 TenantAdmin 이상만 접근 가능한 미들웨어입니다.
func TenantAdminOrAbove() gin.HandlerFunc {
	return RequireRole(domain.RoleTenantAdmin)
}
