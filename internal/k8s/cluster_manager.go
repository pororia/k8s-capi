package k8s

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var (
	gvrCluster              = schema.GroupVersionResource{Group: "cluster.x-k8s.io", Version: "v1beta1", Resource: "clusters"}
	gvrOpenStackCluster     = schema.GroupVersionResource{Group: "infrastructure.cluster.x-k8s.io", Version: "v1beta1", Resource: "openstackclusters"}
	gvrKCP                  = schema.GroupVersionResource{Group: "controlplane.cluster.x-k8s.io", Version: "v1beta1", Resource: "kubeadmcontrolplanes"}
	gvrMachineDeployment    = schema.GroupVersionResource{Group: "cluster.x-k8s.io", Version: "v1beta1", Resource: "machinedeployments"}
	gvrOpenStackMachineTemplate = schema.GroupVersionResource{Group: "infrastructure.cluster.x-k8s.io", Version: "v1beta1", Resource: "openstackmachinetemplates"}
	gvrSecret               = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
)

// ClusterManager는 CAPI 리소스를 관리합니다.
type ClusterManager struct {
	client dynamic.Interface
}

// NewClusterManager는 새로운 ClusterManager를 생성합니다.
func NewClusterManager(client dynamic.Interface) *ClusterManager {
	return &ClusterManager{client: client}
}

// ApplyManifests는 매니페스트를 Management Cluster에 적용합니다.
func (m *ClusterManager) ApplyManifests(ctx context.Context, manifests []unstructured.Unstructured) error {
	for _, obj := range manifests {
		gvr, err := getGVR(&obj)
		if err != nil {
			return fmt.Errorf("failed to get GVR for %s/%s: %w", obj.GetKind(), obj.GetName(), err)
		}

		namespace := obj.GetNamespace()
		name := obj.GetName()

		// 네임스페이스가 없으면 먼저 생성
		if namespace != "" {
			if err := m.ensureNamespace(ctx, namespace); err != nil {
				return fmt.Errorf("failed to ensure namespace %s: %w", namespace, err)
			}
		}

		// 기존 리소스가 있으면 업데이트, 없으면 생성
		existing, err := m.client.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				if _, err := m.client.Resource(gvr).Namespace(namespace).Create(ctx, &obj, metav1.CreateOptions{}); err != nil {
					return fmt.Errorf("failed to create %s/%s: %w", obj.GetKind(), name, err)
				}
				continue
			}
			return fmt.Errorf("failed to get %s/%s: %w", obj.GetKind(), name, err)
		}

		obj.SetResourceVersion(existing.GetResourceVersion())
		if _, err := m.client.Resource(gvr).Namespace(namespace).Update(ctx, &obj, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("failed to update %s/%s: %w", obj.GetKind(), name, err)
		}
	}
	return nil
}

// DeleteCluster는 CAPI 클러스터 리소스를 삭제합니다.
// Cluster 오브젝트 삭제 시 CAPI가 연관 리소스를 함께 삭제합니다.
func (m *ClusterManager) DeleteCluster(ctx context.Context, namespace, clusterName string) error {
	err := m.client.Resource(gvrCluster).Namespace(namespace).Delete(ctx, clusterName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete cluster %s: %w", clusterName, err)
	}
	return nil
}

// ScaleCluster는 클러스터의 노드 수를 조정합니다.
func (m *ClusterManager) ScaleCluster(ctx context.Context, namespace, clusterName string, cpCount, workerCount *int) error {
	if cpCount != nil {
		kcp, err := m.client.Resource(gvrKCP).Namespace(namespace).Get(ctx, clusterName+"-control-plane", metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get KubeadmControlPlane: %w", err)
		}
		if err := unstructured.SetNestedField(kcp.Object, int64(*cpCount), "spec", "replicas"); err != nil {
			return fmt.Errorf("failed to set CP replicas: %w", err)
		}
		if _, err := m.client.Resource(gvrKCP).Namespace(namespace).Update(ctx, kcp, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("failed to update KubeadmControlPlane: %w", err)
		}
	}

	if workerCount != nil {
		md, err := m.client.Resource(gvrMachineDeployment).Namespace(namespace).Get(ctx, clusterName+"-workers", metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get MachineDeployment: %w", err)
		}
		if err := unstructured.SetNestedField(md.Object, int64(*workerCount), "spec", "replicas"); err != nil {
			return fmt.Errorf("failed to set worker replicas: %w", err)
		}
		if _, err := m.client.Resource(gvrMachineDeployment).Namespace(namespace).Update(ctx, md, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("failed to update MachineDeployment: %w", err)
		}
	}

	return nil
}

// UpgradeCluster는 클러스터의 Kubernetes 버전을 업그레이드합니다.
func (m *ClusterManager) UpgradeCluster(ctx context.Context, namespace, clusterName, targetVersion string) error {
	// KubeadmControlPlane 버전 업데이트
	kcp, err := m.client.Resource(gvrKCP).Namespace(namespace).Get(ctx, clusterName+"-control-plane", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get KubeadmControlPlane: %w", err)
	}
	if err := unstructured.SetNestedField(kcp.Object, targetVersion, "spec", "version"); err != nil {
		return fmt.Errorf("failed to set KCP version: %w", err)
	}
	if _, err := m.client.Resource(gvrKCP).Namespace(namespace).Update(ctx, kcp, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("failed to update KubeadmControlPlane version: %w", err)
	}

	// MachineDeployment 버전 업데이트
	md, err := m.client.Resource(gvrMachineDeployment).Namespace(namespace).Get(ctx, clusterName+"-workers", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get MachineDeployment: %w", err)
	}
	if err := unstructured.SetNestedField(md.Object, targetVersion, "spec", "template", "spec", "version"); err != nil {
		return fmt.Errorf("failed to set MachineDeployment version: %w", err)
	}
	if _, err := m.client.Resource(gvrMachineDeployment).Namespace(namespace).Update(ctx, md, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("failed to update MachineDeployment version: %w", err)
	}

	return nil
}

// GetKubeconfig는 CAPI가 생성한 Secret에서 kubeconfig를 추출합니다.
func (m *ClusterManager) GetKubeconfig(ctx context.Context, namespace, clusterName string) ([]byte, error) {
	secretName := clusterName + "-kubeconfig"
	secret, err := m.client.Resource(gvrSecret).Namespace(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig secret: %w", err)
	}

	data, found, err := unstructured.NestedMap(secret.Object, "data")
	if err != nil || !found {
		return nil, fmt.Errorf("kubeconfig secret has no data")
	}

	kubeconfig, ok := data["value"].(string)
	if !ok {
		return nil, fmt.Errorf("kubeconfig secret value is not a string")
	}

	return []byte(kubeconfig), nil
}

func (m *ClusterManager) ensureNamespace(ctx context.Context, name string) error {
	gvrNamespace := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}
	_, err := m.client.Resource(gvrNamespace).Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		return nil
	}
	if !errors.IsNotFound(err) {
		return err
	}

	ns := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata":   map[string]interface{}{"name": name},
		},
	}
	_, err = m.client.Resource(gvrNamespace).Create(ctx, ns, metav1.CreateOptions{})
	return err
}

// getGVR는 unstructured 오브젝트의 GVR을 반환합니다.
func getGVR(obj *unstructured.Unstructured) (schema.GroupVersionResource, error) {
	apiVersion := obj.GetAPIVersion()
	kind := obj.GetKind()

	gvrMap := map[string]schema.GroupVersionResource{
		"v1/Secret":                                                                          gvrSecret,
		"cluster.x-k8s.io/v1beta1/Cluster":                                                 gvrCluster,
		"infrastructure.cluster.x-k8s.io/v1beta1/OpenStackCluster":                         gvrOpenStackCluster,
		"controlplane.cluster.x-k8s.io/v1beta1/KubeadmControlPlane":                        gvrKCP,
		"cluster.x-k8s.io/v1beta1/MachineDeployment":                                       gvrMachineDeployment,
		"infrastructure.cluster.x-k8s.io/v1beta1/OpenStackMachineTemplate":                  gvrOpenStackMachineTemplate,
		"bootstrap.cluster.x-k8s.io/v1beta1/KubeadmConfigTemplate": {
			Group: "bootstrap.cluster.x-k8s.io", Version: "v1beta1", Resource: "kubeadmconfigtemplates",
		},
	}

	key := apiVersion + "/" + kind
	gvr, ok := gvrMap[key]
	if !ok {
		return schema.GroupVersionResource{}, fmt.Errorf("unknown resource kind: %s/%s", apiVersion, kind)
	}
	return gvr, nil
}
