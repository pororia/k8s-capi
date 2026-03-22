package openstack

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/k8s-cmp/k8s-cmp/internal/domain"
)

// ClientFactory는 테넌트별 OpenStack 클라이언트를 생성합니다.
type ClientFactory struct{}

// NewClientFactory는 새로운 ClientFactory를 생성합니다.
func NewClientFactory() *ClientFactory {
	return &ClientFactory{}
}

// NewProvider는 테넌트의 자격증명으로 gophercloud ProviderClient를 생성합니다.
func (f *ClientFactory) NewProvider(ctx context.Context, tenant *domain.Tenant, creds *domain.OSCredentialsJSON) (*gophercloud.ProviderClient, error) {
	opts := gophercloud.AuthOptions{
		IdentityEndpoint: tenant.OSAuthURL,
		Username:         creds.Username,
		Password:         creds.Password,
		DomainName:       creds.UserDomainName,
		TenantID:         tenant.OSProjectID,
	}

	provider, err := openstack.AuthenticatedClient(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with OpenStack: %w", err)
	}

	return provider, nil
}

// ComputeService는 Nova API 래퍼입니다.
type ComputeService struct {
	client *gophercloud.ServiceClient
}

// NewComputeService는 Compute Service를 생성합니다.
func NewComputeService(provider *gophercloud.ProviderClient) (*ComputeService, error) {
	client, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{})
	if err != nil {
		return nil, fmt.Errorf("failed to create compute client: %w", err)
	}
	return &ComputeService{client: client}, nil
}

// NetworkService는 Neutron API 래퍼입니다.
type NetworkService struct {
	client *gophercloud.ServiceClient
}

// NewNetworkService는 Network Service를 생성합니다.
func NewNetworkService(provider *gophercloud.ProviderClient) (*NetworkService, error) {
	client, err := openstack.NewNetworkV2(provider, gophercloud.EndpointOpts{})
	if err != nil {
		return nil, fmt.Errorf("failed to create network client: %w", err)
	}
	return &NetworkService{client: client}, nil
}

// ImageService는 Glance API 래퍼입니다.
type ImageService struct {
	client *gophercloud.ServiceClient
}

// NewImageService는 Image Service를 생성합니다.
func NewImageService(provider *gophercloud.ProviderClient) (*ImageService, error) {
	client, err := openstack.NewImageV2(provider, gophercloud.EndpointOpts{})
	if err != nil {
		return nil, fmt.Errorf("failed to create image client: %w", err)
	}
	return &ImageService{client: client}, nil
}
