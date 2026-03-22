package openstack

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/subnets"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/security/groups"
	"github.com/gophercloud/gophercloud/v2/pagination"
)

// Network는 OpenStack Network 정보입니다.
type Network struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	External bool   `json:"external"`
	Status   string `json:"status"`
}

// Subnet은 OpenStack Subnet 정보입니다.
type Subnet struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	NetworkID string `json:"network_id"`
	CIDR      string `json:"cidr"`
	IPVersion int    `json:"ip_version"`
}

// SecurityGroup은 OpenStack Security Group 정보입니다.
type SecurityGroup struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ListNetworks는 네트워크 목록을 반환합니다.
func (s *NetworkService) ListNetworks(ctx context.Context) ([]*Network, error) {
	var result []*Network

	pager := networks.List(s.client, networks.ListOpts{})
	err := pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		nets, err := networks.ExtractNetworks(page)
		if err != nil {
			return false, fmt.Errorf("failed to extract networks: %w", err)
		}
		for _, n := range nets {
			result = append(result, &Network{
				ID:     n.ID,
				Name:   n.Name,
				Status: n.Status,
			})
		}
		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list networks: %w", err)
	}

	return result, nil
}

// ListExternalNetworks는 외부 네트워크 목록을 반환합니다.
func (s *NetworkService) ListExternalNetworks(ctx context.Context) ([]*Network, error) {
	var result []*Network

	pager := networks.List(s.client, networks.ListOpts{})
	err := pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		nets, err := networks.ExtractNetworks(page)
		if err != nil {
			return false, fmt.Errorf("failed to extract external networks: %w", err)
		}
		for _, n := range nets {
			result = append(result, &Network{
				ID:     n.ID,
				Name:   n.Name,
				Status: n.Status,
			})
		}
		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list external networks: %w", err)
	}

	return result, nil
}

// ListSubnets는 특정 네트워크의 서브넷 목록을 반환합니다.
func (s *NetworkService) ListSubnets(ctx context.Context, networkID string) ([]*Subnet, error) {
	var result []*Subnet

	listOpts := subnets.ListOpts{}
	if networkID != "" {
		listOpts.NetworkID = networkID
	}

	pager := subnets.List(s.client, listOpts)
	err := pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		subs, err := subnets.ExtractSubnets(page)
		if err != nil {
			return false, fmt.Errorf("failed to extract subnets: %w", err)
		}
		for _, sub := range subs {
			result = append(result, &Subnet{
				ID:        sub.ID,
				Name:      sub.Name,
				NetworkID: sub.NetworkID,
				CIDR:      sub.CIDR,
				IPVersion: sub.IPVersion,
			})
		}
		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list subnets: %w", err)
	}

	return result, nil
}

// ListSecurityGroups는 보안 그룹 목록을 반환합니다.
func (s *NetworkService) ListSecurityGroups(ctx context.Context) ([]*SecurityGroup, error) {
	var result []*SecurityGroup

	pager := groups.List(s.client, groups.ListOpts{})
	err := pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		sgs, err := groups.ExtractGroups(page)
		if err != nil {
			return false, fmt.Errorf("failed to extract security groups: %w", err)
		}
		for _, sg := range sgs {
			result = append(result, &SecurityGroup{
				ID:          sg.ID,
				Name:        sg.Name,
				Description: sg.Description,
			})
		}
		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list security groups: %w", err)
	}

	return result, nil
}
