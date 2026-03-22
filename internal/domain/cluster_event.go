package domain

import (
	"time"

	"github.com/google/uuid"
)

// EventSeverity는 이벤트 심각도를 나타냅니다.
type EventSeverity string

const (
	SeverityInfo    EventSeverity = "info"
	SeverityWarning EventSeverity = "warning"
	SeverityError   EventSeverity = "error"
)

// ClusterEvent는 클러스터 이벤트 도메인 모델입니다.
type ClusterEvent struct {
	ID        uuid.UUID     `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	ClusterID uuid.UUID     `json:"cluster_id" gorm:"type:uuid;not null;index"`
	Type      string        `json:"type" gorm:"size:50;not null"`
	Severity  EventSeverity `json:"severity" gorm:"size:20;not null;default:info"`
	Message   string        `json:"message" gorm:"not null"`
	Source    string        `json:"source,omitempty" gorm:"size:100"`
	Metadata  JSONMap       `json:"metadata,omitempty" gorm:"type:jsonb"`
	CreatedAt time.Time     `json:"created_at"`
}
