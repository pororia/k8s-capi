package domain

import (
	"time"

	"github.com/google/uuid"
)

// AuditResult는 감사 로그 결과를 나타냅니다.
type AuditResult string

const (
	AuditResultSuccess AuditResult = "success"
	AuditResultFailure AuditResult = "failure"
)

// AuditLog는 감사 로그 도메인 모델입니다.
type AuditLog struct {
	ID             uuid.UUID   `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	TenantID       *uuid.UUID  `json:"tenant_id,omitempty" gorm:"type:uuid;index"`
	UserID         *uuid.UUID  `json:"user_id,omitempty" gorm:"type:uuid;index"`
	Action         string      `json:"action" gorm:"size:50;not null;index"`
	ResourceType   string      `json:"resource_type" gorm:"size:50;not null;index"`
	ResourceID     string      `json:"resource_id,omitempty" gorm:"size:100"`
	ResourceName   string      `json:"resource_name,omitempty" gorm:"size:200"`
	RequestBody    JSONMap     `json:"request_body,omitempty" gorm:"type:jsonb"`
	ResponseStatus int         `json:"response_status,omitempty"`
	Result         AuditResult `json:"result" gorm:"size:20;not null;default:success"`
	ErrorMessage   string      `json:"error_message,omitempty"`
	IPAddress      string      `json:"ip_address,omitempty" gorm:"size:45"`
	UserAgent      string      `json:"user_agent,omitempty" gorm:"size:500"`
	DurationMs     int         `json:"duration_ms,omitempty"`
	CreatedAt      time.Time   `json:"created_at;index"`
}
