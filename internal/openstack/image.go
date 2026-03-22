package openstack

import (
	"context"
	"fmt"
	"strings"

	"github.com/gophercloud/gophercloud/v2"
)

// Image는 OpenStack Image 정보입니다.
type Image struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	K8sVersion string `json:"k8s_version,omitempty"`
}

// imageResult는 이미지 API 응답 구조입니다.
type imageResult struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Status     string                 `json:"status"`
	Properties map[string]interface{} `json:"properties"`
}

// ListK8sImages는 Kubernetes 호환 이미지 목록을 반환합니다.
// 이미지 이름에 "kube-v" 패턴이 포함된 이미지만 반환합니다.
func (s *ImageService) ListK8sImages(ctx context.Context) ([]*Image, error) {
	var result []*Image

	// gophercloud v2 image API는 직접 HTTP 요청 사용
	url := s.client.ServiceURL("v2", "images") + "?status=active"
	var resp struct {
		Images []imageResult `json:"images"`
	}

	_, err := s.client.Get(ctx, url, &resp, &gophercloud.RequestOpts{})
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	for _, img := range resp.Images {
		if !isK8sImage(img.Name, img.Properties) {
			continue
		}
		image := &Image{
			ID:     img.ID,
			Name:   img.Name,
			Status: img.Status,
		}
		if v, ok := img.Properties["k8s_version"]; ok {
			if vs, ok := v.(string); ok {
				image.K8sVersion = vs
			}
		}
		result = append(result, image)
	}

	return result, nil
}

func isK8sImage(name string, props map[string]interface{}) bool {
	if strings.Contains(name, "kube-v") {
		return true
	}
	if _, ok := props["k8s_version"]; ok {
		return true
	}
	return false
}
