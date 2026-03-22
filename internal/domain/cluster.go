package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ClusterStatus는 클러스터 상태를 나타냅니다.
type ClusterStatus string

const (
	ClusterStatusPending      ClusterStatus = "pending"
	ClusterStatusProvisioning ClusterStatus = "provisioning"
	ClusterStatusRunning      ClusterStatus = "running"
	ClusterStatusUpdating     ClusterStatus = "updating"
	ClusterStatusFailed       ClusterStatus = "failed"
	ClusterStatusDeleting     ClusterStatus = "deleting"
	ClusterStatusDeleted      ClusterStatus = "deleted"
)

// StringSlice는 JSONB로 저장되는 문자열 슬라이스입니다.
type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *StringSlice) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan StringSlice: %v", value)
	}
	return json.Unmarshal(bytes, s)
}

// JSONMap은 JSONB로 저장되는 map입니다.
type JSONMap map[string]interface{}

func (m JSONMap) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *JSONMap) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan JSONMap: %v", value)
	}
	return json.Unmarshal(bytes, m)
}

// Cluster는 클러스터 도메인 모델입니다.
type Cluster struct {
	ID                    uuid.UUID     `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	TenantID              uuid.UUID     `json:"tenant_id" gorm:"type:uuid;not null;index"`
	Name                  string        `json:"name" gorm:"size:100;not null"`
	KubernetesVersion     string        `json:"kubernetes_version" gorm:"size:20;not null"`
	Status                ClusterStatus `json:"status" gorm:"size:20;not null;default:pending"`
	CPCount               int           `json:"cp_count" gorm:"not null"`
	WorkerCount           int           `json:"worker_count" gorm:"not null"`
	CPFlavor              string        `json:"cp_flavor" gorm:"size:100;not null"`
	WorkerFlavor          string        `json:"worker_flavor" gorm:"size:100;not null"`
	OSImage               string        `json:"os_image" gorm:"size:100;not null"`
	NetworkID             string        `json:"network_id" gorm:"size:64;not null"`
	SubnetID              string        `json:"subnet_id" gorm:"size:64;not null"`
	ExternalNetworkID     string        `json:"external_network_id" gorm:"size:64;not null"`
	SSHKeyName            string        `json:"ssh_key_name" gorm:"size:100;not null"`
	PodCIDR               string        `json:"pod_cidr" gorm:"size:50;not null;default:10.244.0.0/16"`
	ServiceCIDR           string        `json:"service_cidr" gorm:"size:50;not null;default:10.96.0.0/12"`
	CNIPlugin             string        `json:"cni_plugin" gorm:"size:20;not null;default:calico"`
	DNSNameservers        StringSlice   `json:"dns_nameservers" gorm:"type:jsonb;default:'[\"8.8.8.8\"]'"`
	APIEndpoint           string        `json:"api_endpoint,omitempty" gorm:"size:255"`
	CAPINamespace         string        `json:"capi_namespace" gorm:"size:100;not null"`
	CAPIManifestSnapshot  JSONMap       `json:"capi_manifest_snapshot,omitempty" gorm:"type:jsonb"`
	ErrorMessage          string        `json:"error_message,omitempty"`
	CreatedBy             *uuid.UUID    `json:"created_by,omitempty" gorm:"type:uuid"`
	CreatedAt             time.Time     `json:"created_at"`
	UpdatedAt             time.Time     `json:"updated_at"`
	DeletedAt             *time.Time    `json:"deleted_at,omitempty" gorm:"index"`

	// 연관 관계
	Tenant *Tenant       `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	Nodes  []Node        `json:"nodes,omitempty" gorm:"foreignKey:ClusterID"`
	Events []ClusterEvent `json:"events,omitempty" gorm:"foreignKey:ClusterID"`
}

// CanScale은 클러스터가 스케일링 가능한 상태인지 확인합니다.
func (c *Cluster) CanScale() bool {
	return c.Status == ClusterStatusRunning
}

// CanUpgrade는 클러스터가 업그레이드 가능한 상태인지 확인합니다.
func (c *Cluster) CanUpgrade() bool {
	return c.Status == ClusterStatusRunning
}

// CanDelete는 클러스터가 삭제 가능한 상태인지 확인합니다.
func (c *Cluster) CanDelete() bool {
	return c.Status != ClusterStatusDeleting && c.Status != ClusterStatusDeleted
}
