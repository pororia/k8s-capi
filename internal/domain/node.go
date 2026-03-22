package domain

import (
	"time"

	"github.com/google/uuid"
)

// NodeRole은 노드 역할을 나타냅니다.
type NodeRole string

const (
	NodeRoleControlPlane NodeRole = "control_plane"
	NodeRoleWorker       NodeRole = "worker"
)

// NodeStatus는 노드 상태를 나타냅니다.
type NodeStatus string

const (
	NodeStatusProvisioning NodeStatus = "provisioning"
	NodeStatusRunning      NodeStatus = "running"
	NodeStatusFailed       NodeStatus = "failed"
	NodeStatusDeleting     NodeStatus = "deleting"
)

// Node는 노드 도메인 모델입니다.
type Node struct {
	ID                uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	ClusterID         uuid.UUID  `json:"cluster_id" gorm:"type:uuid;not null;index"`
	Name              string     `json:"name" gorm:"size:100;not null"`
	Role              NodeRole   `json:"role" gorm:"size:20;not null"`
	Status            NodeStatus `json:"status" gorm:"size:20;not null;default:provisioning"`
	MachineID         string     `json:"machine_id,omitempty" gorm:"size:100"`
	OpenStackServerID string     `json:"openstack_server_id,omitempty" gorm:"size:64"`
	PrivateIP         string     `json:"private_ip,omitempty" gorm:"size:45"`
	FloatingIP        string     `json:"floating_ip,omitempty" gorm:"size:45"`
	Flavor            string     `json:"flavor" gorm:"size:100;not null"`
	Ready             bool       `json:"ready" gorm:"not null;default:false"`
	KubernetesVersion string     `json:"kubernetes_version,omitempty" gorm:"size:20"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}
