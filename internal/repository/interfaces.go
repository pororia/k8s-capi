package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/k8s-cmp/k8s-cmp/internal/domain"
	"github.com/k8s-cmp/k8s-cmp/pkg/pagination"
)

// UserRepository는 사용자 데이터 접근 인터페이스입니다.
type UserRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByTenantID(ctx context.Context, tenantID uuid.UUID, page *pagination.Params) ([]*domain.User, int64, error)
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateLastLogin(ctx context.Context, id uuid.UUID, t time.Time) error
}

// TenantRepository는 테넌트 데이터 접근 인터페이스입니다.
type TenantRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
	FindByName(ctx context.Context, name string) (*domain.Tenant, error)
	FindAll(ctx context.Context, page *pagination.Params) ([]*domain.Tenant, int64, error)
	Create(ctx context.Context, tenant *domain.Tenant) error
	Update(ctx context.Context, tenant *domain.Tenant) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountClusters(ctx context.Context, tenantID uuid.UUID) (int64, error)
}

// ClusterRepository는 클러스터 데이터 접근 인터페이스입니다.
type ClusterRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Cluster, error)
	FindByTenantID(ctx context.Context, tenantID uuid.UUID, filter ClusterFilter, page *pagination.Params) ([]*domain.Cluster, int64, error)
	FindAll(ctx context.Context, filter ClusterFilter, page *pagination.Params) ([]*domain.Cluster, int64, error)
	Create(ctx context.Context, cluster *domain.Cluster) error
	Update(ctx context.Context, cluster *domain.Cluster) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ClusterStatus, errMsg string) error
	UpdateAPIEndpoint(ctx context.Context, id uuid.UUID, endpoint string) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// NodeRepository는 노드 데이터 접근 인터페이스입니다.
type NodeRepository interface {
	FindByClusterID(ctx context.Context, clusterID uuid.UUID) ([]*domain.Node, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Node, error)
	Upsert(ctx context.Context, node *domain.Node) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByClusterID(ctx context.Context, clusterID uuid.UUID) error
}

// ClusterEventRepository는 클러스터 이벤트 데이터 접근 인터페이스입니다.
type ClusterEventRepository interface {
	FindByClusterID(ctx context.Context, clusterID uuid.UUID, page *pagination.Params) ([]*domain.ClusterEvent, int64, error)
	Create(ctx context.Context, event *domain.ClusterEvent) error
}

// AuditLogRepository는 감사 로그 데이터 접근 인터페이스입니다.
type AuditLogRepository interface {
	Create(ctx context.Context, log *domain.AuditLog) error
	FindAll(ctx context.Context, filter AuditLogFilter, page *pagination.Params) ([]*domain.AuditLog, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.AuditLog, error)
}

// ClusterFilter는 클러스터 목록 조회 필터입니다.
type ClusterFilter struct {
	TenantID  *uuid.UUID
	Status    []domain.ClusterStatus
	Search    string
	SortBy    string
	SortOrder string
}

// AuditLogFilter는 감사 로그 조회 필터입니다.
type AuditLogFilter struct {
	TenantID     *uuid.UUID
	UserID       *uuid.UUID
	Action       string
	ResourceType string
	Result       string
	Search       string
	From         *time.Time
	To           *time.Time
}
