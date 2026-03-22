package domain

import (
	"time"

	"github.com/google/uuid"
)

// TenantStatus는 테넌트 상태를 나타냅니다.
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusDeleted   TenantStatus = "deleted"
)

// Tenant는 테넌트 도메인 모델입니다.
type Tenant struct {
	ID                 uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name               string       `json:"name" gorm:"size:100;uniqueIndex;not null"`
	Description        string       `json:"description"`
	OSAuthURL          string       `json:"os_auth_url" gorm:"size:255;not null"`
	OSProjectID        string       `json:"os_project_id" gorm:"size:64;not null"`
	OSProjectName      string       `json:"os_project_name" gorm:"size:100;not null"`
	OSDomainName       string       `json:"os_domain_name" gorm:"size:100;not null;default:Default"`
	OSCredentials      string       `json:"-" gorm:"not null"` // AES-256-GCM 암호화된 JSON
	MaxClusters        int          `json:"max_clusters" gorm:"not null;default:10"`
	MaxNodesPerCluster int          `json:"max_nodes_per_cluster" gorm:"not null;default:20"`
	Status             TenantStatus `json:"status" gorm:"size:20;not null;default:active"`
	CreatedAt          time.Time    `json:"created_at"`
	UpdatedAt          time.Time    `json:"updated_at"`
	DeletedAt          *time.Time   `json:"deleted_at,omitempty" gorm:"index"`

	// 연관 관계 (로드 시 사용)
	Clusters []Cluster `json:"clusters,omitempty" gorm:"foreignKey:TenantID"`
	Users    []User    `json:"users,omitempty" gorm:"foreignKey:TenantID"`
}

// OSCredentialsJSON은 복호화된 OpenStack 자격증명 구조입니다.
type OSCredentialsJSON struct {
	Username       string `json:"username"`
	Password       string `json:"password"`
	UserDomainName string `json:"user_domain_name"`
}

// IsActive는 테넌트가 활성 상태인지 확인합니다.
func (t *Tenant) IsActive() bool {
	return t.Status == TenantStatusActive && t.DeletedAt == nil
}
