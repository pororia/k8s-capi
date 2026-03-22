package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/k8s-cmp/k8s-cmp/internal/domain"
	"github.com/k8s-cmp/k8s-cmp/internal/k8s"
	"github.com/k8s-cmp/k8s-cmp/internal/repository"
	apperrors "github.com/k8s-cmp/k8s-cmp/pkg/errors"
	"github.com/k8s-cmp/k8s-cmp/pkg/pagination"
	"go.uber.org/zap"
)

// CreateClusterRequest는 클러스터 생성 요청입니다.
type CreateClusterRequest struct {
	ClusterName       string   `json:"cluster_name" validate:"required,cluster_name"`
	KubernetesVersion string   `json:"kubernetes_version" validate:"required,k8s_version"`
	ControlPlaneCount int      `json:"control_plane_count" validate:"required,cp_replicas"`
	WorkerCount       int      `json:"worker_count" validate:"required,min=1,max=100"`
	CPFlavor          string   `json:"control_plane_flavor" validate:"required"`
	WorkerFlavor      string   `json:"worker_flavor" validate:"required"`
	OSImage           string   `json:"os_image" validate:"required"`
	NetworkID         string   `json:"network_id" validate:"required"`
	SubnetID          string   `json:"subnet_id" validate:"required"`
	ExternalNetworkID string   `json:"external_network_id" validate:"required"`
	SSHKeyName        string   `json:"ssh_key_name" validate:"required"`
	PodCIDR           string   `json:"pod_cidr"`
	ServiceCIDR       string   `json:"service_cidr"`
	CNIPlugin         string   `json:"cni_plugin"`
	DNSNameservers    []string `json:"dns_nameservers"`
}

// ScaleParams는 스케일링 파라미터입니다.
type ScaleParams struct {
	ControlPlaneCount *int `json:"control_plane_count" validate:"omitempty,cp_replicas"`
	WorkerCount       *int `json:"worker_count" validate:"omitempty,min=1,max=100"`
}

// UpgradeParams는 업그레이드 파라미터입니다.
type UpgradeParams struct {
	TargetVersion string `json:"target_version" validate:"required,k8s_version"`
}

// ClusterListFilter는 클러스터 목록 조회 필터입니다.
type ClusterListFilter struct {
	Status    []string
	Search    string
	SortBy    string
	SortOrder string
}

// ClusterService는 클러스터 비즈니스 로직을 담당합니다.
type ClusterService struct {
	clusterRepo   repository.ClusterRepository
	tenantRepo    repository.TenantRepository
	eventRepo     repository.ClusterEventRepository
	k8sManager    *k8s.ClusterManager
	manifestGen   *k8s.ManifestGenerator
	tenantSvc     *TenantService
	logger        *zap.Logger
}

// NewClusterService는 새로운 ClusterService를 생성합니다.
func NewClusterService(
	clusterRepo repository.ClusterRepository,
	tenantRepo repository.TenantRepository,
	eventRepo repository.ClusterEventRepository,
	k8sManager *k8s.ClusterManager,
	manifestGen *k8s.ManifestGenerator,
	tenantSvc *TenantService,
	logger *zap.Logger,
) *ClusterService {
	return &ClusterService{
		clusterRepo: clusterRepo,
		tenantRepo:  tenantRepo,
		eventRepo:   eventRepo,
		k8sManager:  k8sManager,
		manifestGen: manifestGen,
		tenantSvc:   tenantSvc,
		logger:      logger,
	}
}

// List는 클러스터 목록을 반환합니다.
func (s *ClusterService) List(ctx context.Context, tenantID uuid.UUID, filter ClusterListFilter, page *pagination.Params) ([]*domain.Cluster, int64, error) {
	repoFilter := repository.ClusterFilter{
		Search:    filter.Search,
		SortBy:    filter.SortBy,
		SortOrder: filter.SortOrder,
	}
	for _, st := range filter.Status {
		repoFilter.Status = append(repoFilter.Status, domain.ClusterStatus(st))
	}
	return s.clusterRepo.FindByTenantID(ctx, tenantID, repoFilter, page)
}

// GetByID는 특정 클러스터를 반환합니다.
func (s *ClusterService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Cluster, error) {
	return s.clusterRepo.FindByID(ctx, id)
}

// Create는 새로운 클러스터를 생성합니다.
func (s *ClusterService) Create(ctx context.Context, tenantID uuid.UUID, createdBy uuid.UUID, req *CreateClusterRequest) (*domain.Cluster, error) {
	// 테넌트 쿼터 확인
	tenant, err := s.tenantRepo.FindByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	clusterCount, err := s.tenantRepo.CountClusters(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if int(clusterCount) >= tenant.MaxClusters {
		return nil, apperrors.ErrQuotaExceeded(fmt.Sprintf("tenant cluster quota exceeded (%d/%d)", clusterCount, tenant.MaxClusters))
	}

	// 노드 수 쿼터 확인
	totalNodes := req.ControlPlaneCount + req.WorkerCount
	if totalNodes > tenant.MaxNodesPerCluster {
		return nil, apperrors.ErrQuotaExceeded(fmt.Sprintf("node count exceeds tenant limit (%d/%d)", totalNodes, tenant.MaxNodesPerCluster))
	}

	// 기본값 설정
	podCIDR := req.PodCIDR
	if podCIDR == "" {
		podCIDR = "10.244.0.0/16"
	}
	serviceCIDR := req.ServiceCIDR
	if serviceCIDR == "" {
		serviceCIDR = "10.96.0.0/12"
	}
	cniPlugin := req.CNIPlugin
	if cniPlugin == "" {
		cniPlugin = "calico"
	}
	dnsNameservers := req.DNSNameservers
	if len(dnsNameservers) == 0 {
		dnsNameservers = []string{"8.8.8.8"}
	}

	// CAPI 네임스페이스: 테넌트 ID 기반
	namespace := fmt.Sprintf("cmp-%s", tenantID.String()[:8])

	cluster := &domain.Cluster{
		TenantID:          tenantID,
		Name:              req.ClusterName,
		KubernetesVersion: req.KubernetesVersion,
		Status:            domain.ClusterStatusPending,
		CPCount:           req.ControlPlaneCount,
		WorkerCount:       req.WorkerCount,
		CPFlavor:          req.CPFlavor,
		WorkerFlavor:      req.WorkerFlavor,
		OSImage:           req.OSImage,
		NetworkID:         req.NetworkID,
		SubnetID:          req.SubnetID,
		ExternalNetworkID: req.ExternalNetworkID,
		SSHKeyName:        req.SSHKeyName,
		PodCIDR:           podCIDR,
		ServiceCIDR:       serviceCIDR,
		CNIPlugin:         cniPlugin,
		DNSNameservers:    domain.StringSlice(dnsNameservers),
		CAPINamespace:     namespace,
		CreatedBy:         &createdBy,
	}

	if err := s.clusterRepo.Create(ctx, cluster); err != nil {
		return nil, err
	}

	// OpenStack 자격증명 복호화
	creds, err := s.tenantSvc.DecryptCredentials(tenant)
	if err != nil {
		_ = s.clusterRepo.UpdateStatus(ctx, cluster.ID, domain.ClusterStatusFailed, "failed to decrypt credentials")
		return nil, err
	}

	// CAPI 매니페스트 생성 및 Apply
	params := k8s.ClusterCreateParams{
		ClusterName:       req.ClusterName,
		Namespace:         namespace,
		KubernetesVersion: req.KubernetesVersion,
		ControlPlaneCount: req.ControlPlaneCount,
		WorkerCount:       req.WorkerCount,
		CPFlavor:          req.CPFlavor,
		WorkerFlavor:      req.WorkerFlavor,
		OSImage:           req.OSImage,
		NetworkID:         req.NetworkID,
		SubnetID:          req.SubnetID,
		ExternalNetworkID: req.ExternalNetworkID,
		SSHKeyName:        req.SSHKeyName,
		PodCIDR:           podCIDR,
		ServiceCIDR:       serviceCIDR,
		CNIPlugin:         cniPlugin,
		DNSNameservers:    dnsNameservers,
		CloudName:         "openstack",
		AuthURL:           tenant.OSAuthURL,
		ProjectID:         tenant.OSProjectID,
		DomainName:        tenant.OSDomainName,
		Username:          creds.Username,
		Password:          creds.Password,
		UserDomainName:    creds.UserDomainName,
	}

	manifests, err := s.manifestGen.GenerateManifests(params)
	if err != nil {
		_ = s.clusterRepo.UpdateStatus(ctx, cluster.ID, domain.ClusterStatusFailed, err.Error())
		return nil, apperrors.ErrInternal(fmt.Errorf("failed to generate manifests: %w", err))
	}

	if err := s.k8sManager.ApplyManifests(ctx, manifests); err != nil {
		_ = s.clusterRepo.UpdateStatus(ctx, cluster.ID, domain.ClusterStatusFailed, err.Error())
		return nil, apperrors.ErrInternal(fmt.Errorf("failed to apply manifests: %w", err))
	}

	// 상태를 Provisioning으로 업데이트
	_ = s.clusterRepo.UpdateStatus(ctx, cluster.ID, domain.ClusterStatusProvisioning, "")

	s.logger.Info("cluster creation started",
		zap.String("cluster_id", cluster.ID.String()),
		zap.String("name", cluster.Name),
		zap.String("tenant_id", tenantID.String()))

	cluster.Status = domain.ClusterStatusProvisioning
	return cluster, nil
}

// Delete는 클러스터를 삭제합니다.
func (s *ClusterService) Delete(ctx context.Context, id uuid.UUID) error {
	cluster, err := s.clusterRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if !cluster.CanDelete() {
		return apperrors.ErrInvalidStatus(fmt.Sprintf("cannot delete cluster in status: %s", cluster.Status))
	}

	// CAPI 리소스 삭제
	if err := s.k8sManager.DeleteCluster(ctx, cluster.CAPINamespace, cluster.Name); err != nil {
		s.logger.Error("failed to delete CAPI resources", zap.Error(err))
	}

	// 상태를 Deleting으로 업데이트
	return s.clusterRepo.UpdateStatus(ctx, id, domain.ClusterStatusDeleting, "")
}

// Scale은 클러스터의 노드 수를 조정합니다.
func (s *ClusterService) Scale(ctx context.Context, id uuid.UUID, params ScaleParams) error {
	cluster, err := s.clusterRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if !cluster.CanScale() {
		return apperrors.ErrInvalidStatus(fmt.Sprintf("cannot scale cluster in status: %s", cluster.Status))
	}

	if err := s.k8sManager.ScaleCluster(ctx, cluster.CAPINamespace, cluster.Name, params.ControlPlaneCount, params.WorkerCount); err != nil { //nolint:typecheck
		return apperrors.ErrInternal(fmt.Errorf("failed to scale cluster: %w", err))
	}

	return s.clusterRepo.UpdateStatus(ctx, id, domain.ClusterStatusUpdating, "")
}

// Upgrade는 클러스터의 Kubernetes 버전을 업그레이드합니다.
func (s *ClusterService) Upgrade(ctx context.Context, id uuid.UUID, params UpgradeParams) error {
	cluster, err := s.clusterRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if !cluster.CanUpgrade() {
		return apperrors.ErrInvalidStatus(fmt.Sprintf("cannot upgrade cluster in status: %s", cluster.Status))
	}

	if err := s.k8sManager.UpgradeCluster(ctx, cluster.CAPINamespace, cluster.Name, params.TargetVersion); err != nil {
		return apperrors.ErrInternal(fmt.Errorf("failed to upgrade cluster: %w", err))
	}

	return s.clusterRepo.UpdateStatus(ctx, id, domain.ClusterStatusUpdating, "")
}

// GetKubeconfig는 클러스터의 kubeconfig를 반환합니다.
func (s *ClusterService) GetKubeconfig(ctx context.Context, id uuid.UUID) ([]byte, error) {
	cluster, err := s.clusterRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if cluster.Status != domain.ClusterStatusRunning {
		return nil, apperrors.ErrInvalidStatus("kubeconfig is only available for running clusters")
	}

	return s.k8sManager.GetKubeconfig(ctx, cluster.CAPINamespace, cluster.Name)
}

// GetEvents는 클러스터 이벤트 목록을 반환합니다.
func (s *ClusterService) GetEvents(ctx context.Context, clusterID uuid.UUID, page *pagination.Params) ([]*domain.ClusterEvent, int64, error) {
	return s.eventRepo.FindByClusterID(ctx, clusterID, page)
}
