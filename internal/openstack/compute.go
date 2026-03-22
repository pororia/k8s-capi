package openstack

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/keypairs"
	"github.com/gophercloud/gophercloud/v2/pagination"
)

// Flavor는 OpenStack Flavor 정보입니다.
type Flavor struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	VCPUs int    `json:"vcpus"`
	RAM   int    `json:"ram"`   // MB
	Disk  int    `json:"disk"`  // GB
}

// Keypair는 OpenStack Keypair 정보입니다.
type Keypair struct {
	Name        string `json:"name"`
	Fingerprint string `json:"fingerprint"`
}

// ListFlavors는 Flavor 목록을 반환합니다.
func (s *ComputeService) ListFlavors(ctx context.Context) ([]*Flavor, error) {
	var result []*Flavor

	listOpts := flavors.ListOpts{
		AccessType: flavors.PublicAccess,
	}

	pager := flavors.ListDetail(s.client, listOpts)
	err := pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		fs, err := flavors.ExtractFlavors(page)
		if err != nil {
			return false, fmt.Errorf("failed to extract flavors: %w", err)
		}
		for _, f := range fs {
			result = append(result, &Flavor{
				ID:    f.ID,
				Name:  f.Name,
				VCPUs: f.VCPUs,
				RAM:   f.RAM,
				Disk:  f.Disk,
			})
		}
		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list flavors: %w", err)
	}

	return result, nil
}

// ListKeypairs는 Keypair 목록을 반환합니다.
func (s *ComputeService) ListKeypairs(ctx context.Context) ([]*Keypair, error) {
	var result []*Keypair

	pager := keypairs.List(s.client, keypairs.ListOpts{})
	err := pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		kps, err := keypairs.ExtractKeyPairs(page)
		if err != nil {
			return false, fmt.Errorf("failed to extract keypairs: %w", err)
		}
		for _, kp := range kps {
			result = append(result, &Keypair{
				Name:        kp.Name,
				Fingerprint: kp.Fingerprint,
			})
		}
		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list keypairs: %w", err)
	}

	return result, nil
}
