package openstack

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/quotasets"
)

// Quota는 테넌트의 리소스 쿼터 정보입니다.
type Quota struct {
	Compute ComputeQuota `json:"compute"`
}

// ComputeQuota는 Nova 쿼터 정보입니다.
type ComputeQuota struct {
	Instances   QuotaDetail `json:"instances"`
	Cores       QuotaDetail `json:"cores"`
	RAM         QuotaDetail `json:"ram"`
	KeyPairs    QuotaDetail `json:"key_pairs"`
	FloatingIPs QuotaDetail `json:"floating_ips"`
}

// QuotaDetail은 쿼터 상세 (한도/사용량)입니다.
type QuotaDetail struct {
	Limit int `json:"limit"`
	Used  int `json:"used"`
}

// QuotaService는 쿼터 조회 서비스입니다.
type QuotaService struct {
	provider *gophercloud.ProviderClient
}

// NewQuotaService는 새로운 QuotaService를 생성합니다.
func NewQuotaService(provider *gophercloud.ProviderClient) *QuotaService {
	return &QuotaService{provider: provider}
}

// GetQuota는 프로젝트의 쿼터 정보를 반환합니다.
func (s *QuotaService) GetQuota(ctx context.Context, projectID string) (*Quota, error) {
	computeClient, err := openstack.NewComputeV2(s.provider, gophercloud.EndpointOpts{})
	if err != nil {
		return nil, fmt.Errorf("failed to create compute client: %w", err)
	}

	qs, err := quotasets.GetDetail(ctx, computeClient, projectID).Extract()
	if err != nil {
		return nil, fmt.Errorf("failed to get compute quotas: %w", err)
	}

	return &Quota{
		Compute: ComputeQuota{
			Instances: QuotaDetail{Limit: qs.Instances.Limit, Used: qs.Instances.InUse},
			Cores:     QuotaDetail{Limit: qs.Cores.Limit, Used: qs.Cores.InUse},
			RAM:       QuotaDetail{Limit: qs.RAM.Limit, Used: qs.RAM.InUse},
		},
	}, nil
}
