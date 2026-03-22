package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/k8s-cmp/k8s-cmp/internal/domain"
	"github.com/k8s-cmp/k8s-cmp/internal/repository"
	"go.uber.org/zap"
)

// auditActions는 감사 로그를 기록할 경로 패턴과 액션 매핑입니다.
var auditActions = map[string]string{
	"POST /api/v1/auth/login":     "auth.login",
	"POST /api/v1/auth/logout":    "auth.logout",
	"POST /api/v1/tenants":        "tenant.create",
	"PUT /api/v1/tenants":         "tenant.update",
	"DELETE /api/v1/tenants":      "tenant.delete",
	"POST /api/v1/clusters":       "cluster.create",
	"DELETE /api/v1/clusters":     "cluster.delete",
	"POST /api/v1/clusters/scale": "cluster.scale",
}

// Audit는 주요 작업을 감사 로그에 기록하는 미들웨어입니다.
func Audit(auditRepo repository.AuditLogRepository, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 요청 바디 읽기 (읽은 후 다시 설정)
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		c.Next()

		// 감사 대상 액션 확인
		action := detectAction(c)
		if action == "" {
			return
		}

		duration := int(time.Since(start).Milliseconds())
		status := c.Writer.Status()
		result := domain.AuditResultSuccess
		if status >= 400 {
			result = domain.AuditResultFailure
		}

		log := &domain.AuditLog{
			Action:         action,
			ResourceType:   extractResourceType(action),
			ResponseStatus: status,
			Result:         result,
			IPAddress:      c.ClientIP(),
			UserAgent:      c.Request.UserAgent(),
			DurationMs:     duration,
		}

		// 사용자 정보
		if userIDStr, exists := c.Get(ContextKeyUserID); exists {
			if id, err := uuid.Parse(userIDStr.(string)); err == nil {
				log.UserID = &id
			}
		}
		if tenantIDStr, exists := c.Get(ContextKeyTenantID); exists {
			if id, err := uuid.Parse(tenantIDStr.(string)); err == nil {
				log.TenantID = &id
			}
		}

		// 요청 바디 마스킹
		if len(bodyBytes) > 0 {
			var bodyMap map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &bodyMap); err == nil {
				maskSensitiveFields(bodyMap)
				log.RequestBody = domain.JSONMap(bodyMap)
			}
		}

		// 비동기로 저장
		go func() {
			if err := auditRepo.Create(c.Request.Context(), log); err != nil {
				logger.Error("failed to save audit log", zap.Error(err))
			}
		}()
	}
}

func detectAction(c *gin.Context) string {
	method := c.Request.Method
	path := c.FullPath()
	key := method + " " + path

	// 정확한 매핑 확인
	if action, ok := auditActions[key]; ok {
		return action
	}

	// 패턴 매칭 (ID 포함 경로)
	for pattern, action := range auditActions {
		if strings.HasPrefix(key, pattern) {
			return action
		}
	}

	return ""
}

func extractResourceType(action string) string {
	parts := strings.SplitN(action, ".", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}

func maskSensitiveFields(m map[string]interface{}) {
	sensitiveKeys := []string{"password", "os_credentials", "credentials", "token", "secret"}
	for _, key := range sensitiveKeys {
		if _, ok := m[key]; ok {
			m[key] = "***"
		}
	}
}
