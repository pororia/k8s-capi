package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/k8s-cmp/k8s-cmp/internal/openstack"
	"github.com/k8s-cmp/k8s-cmp/internal/repository"
	apperrors "github.com/k8s-cmp/k8s-cmp/pkg/errors"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// OpenStackService는 OpenStack 리소스 조회 서비스입니다.
type OpenStackService struct {
	tenantRepo repository.TenantRepository
	tenantSvc  *TenantService
	osFactory  *openstack.ClientFactory
	redis      *redis.Client
	logger     *zap.Logger
}

// NewOpenStackService는 새로운 OpenStackService를 생성합니다.
func NewOpenStackService(
	tenantRepo repository.TenantRepository,
	tenantSvc *TenantService,
	osFactory *openstack.ClientFactory,
	redis *redis.Client,
	logger *zap.Logger,
) *OpenStackService {
	return &OpenStackService{
		tenantRepo: tenantRepo,
		tenantSvc:  tenantSvc,
		osFactory:  osFactory,
		redis:      redis,
		logger:     logger,
	}
}

// GetFlavors는 OpenStack Flavor 목록을 반환합니다.
func (s *OpenStackService) GetFlavors(ctx context.Context, tenantID uuid.UUID) ([]*openstack.Flavor, error) {
	client, err := s.getComputeClient(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return client.ListFlavors(ctx)
}

// GetImages는 K8s 호환 이미지 목록을 반환합니다.
func (s *OpenStackService) GetImages(ctx context.Context, tenantID uuid.UUID) ([]*openstack.Image, error) {
	client, err := s.getImageClient(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return client.ListK8sImages(ctx)
}

// GetNetworks는 네트워크 목록을 반환합니다.
func (s *OpenStackService) GetNetworks(ctx context.Context, tenantID uuid.UUID) ([]*openstack.Network, error) {
	client, err := s.getNetworkClient(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return client.ListNetworks(ctx)
}

// GetSubnets는 특정 네트워크의 서브넷 목록을 반환합니다.
func (s *OpenStackService) GetSubnets(ctx context.Context, tenantID uuid.UUID, networkID string) ([]*openstack.Subnet, error) {
	client, err := s.getNetworkClient(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return client.ListSubnets(ctx, networkID)
}

// GetSecurityGroups는 보안 그룹 목록을 반환합니다.
func (s *OpenStackService) GetSecurityGroups(ctx context.Context, tenantID uuid.UUID) ([]*openstack.SecurityGroup, error) {
	client, err := s.getNetworkClient(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return client.ListSecurityGroups(ctx)
}

// GetKeypairs는 SSH 키페어 목록을 반환합니다.
func (s *OpenStackService) GetKeypairs(ctx context.Context, tenantID uuid.UUID) ([]*openstack.Keypair, error) {
	client, err := s.getComputeClient(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return client.ListKeypairs(ctx)
}

// GetExternalNetworks는 외부 네트워크 목록을 반환합니다.
func (s *OpenStackService) GetExternalNetworks(ctx context.Context, tenantID uuid.UUID) ([]*openstack.Network, error) {
	client, err := s.getNetworkClient(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return client.ListExternalNetworks(ctx)
}

// GetQuotas는 테넌트의 OpenStack 리소스 쿼터를 반환합니다.
func (s *OpenStackService) GetQuotas(ctx context.Context, tenantID uuid.UUID) (*openstack.Quota, error) {
	tenant, err := s.tenantRepo.FindByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	creds, err := s.tenantSvc.DecryptCredentials(tenant)
	if err != nil {
		return nil, err
	}

	provider, err := s.osFactory.NewProvider(ctx, tenant, creds)
	if err != nil {
		return nil, apperrors.ErrInternal(fmt.Errorf("failed to create OpenStack client: %w", err))
	}

	return openstack.NewQuotaService(provider).GetQuota(ctx, tenant.OSProjectID)
}

func (s *OpenStackService) getComputeClient(ctx context.Context, tenantID uuid.UUID) (*openstack.ComputeService, error) {
	tenant, err := s.tenantRepo.FindByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	creds, err := s.tenantSvc.DecryptCredentials(tenant)
	if err != nil {
		return nil, err
	}
	provider, err := s.osFactory.NewProvider(ctx, tenant, creds)
	if err != nil {
		return nil, apperrors.ErrInternal(err)
	}
	return openstack.NewComputeService(provider)
}

func (s *OpenStackService) getNetworkClient(ctx context.Context, tenantID uuid.UUID) (*openstack.NetworkService, error) {
	tenant, err := s.tenantRepo.FindByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	creds, err := s.tenantSvc.DecryptCredentials(tenant)
	if err != nil {
		return nil, err
	}
	provider, err := s.osFactory.NewProvider(ctx, tenant, creds)
	if err != nil {
		return nil, apperrors.ErrInternal(err)
	}
	return openstack.NewNetworkService(provider)
}

func (s *OpenStackService) getImageClient(ctx context.Context, tenantID uuid.UUID) (*openstack.ImageService, error) {
	tenant, err := s.tenantRepo.FindByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	creds, err := s.tenantSvc.DecryptCredentials(tenant)
	if err != nil {
		return nil, err
	}
	provider, err := s.osFactory.NewProvider(ctx, tenant, creds)
	if err != nil {
		return nil, apperrors.ErrInternal(err)
	}
	return openstack.NewImageService(provider)
}
