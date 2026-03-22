package k8s

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

// ClusterEvent는 Watcher가 감지한 클러스터 상태 변경 이벤트입니다.
type ClusterEvent struct {
	ClusterName string
	Namespace   string
	Kind        string
	EventType   string // "ADDED", "MODIFIED", "DELETED"
	Object      map[string]interface{}
}

// ClusterWatcher는 CAPI 리소스를 Informer로 감시합니다.
type ClusterWatcher struct {
	dynamicClient dynamic.Interface
	factory       dynamicinformer.DynamicSharedInformerFactory
	EventChan     chan ClusterEvent
	logger        *zap.Logger
}

// NewClusterWatcher는 새로운 ClusterWatcher를 생성합니다.
func NewClusterWatcher(client dynamic.Interface, logger *zap.Logger) *ClusterWatcher {
	return &ClusterWatcher{
		dynamicClient: client,
		EventChan:     make(chan ClusterEvent, 100),
		logger:        logger,
	}
}

// Start는 CAPI 리소스에 대한 Informer를 시작합니다.
func (w *ClusterWatcher) Start(ctx context.Context) error {
	w.factory = dynamicinformer.NewDynamicSharedInformerFactory(w.dynamicClient, 30*time.Second)

	watchResources := []struct {
		gvr  schema.GroupVersionResource
		kind string
	}{
		{gvrCluster, "Cluster"},
		{schema.GroupVersionResource{Group: "cluster.x-k8s.io", Version: "v1beta1", Resource: "machines"}, "Machine"},
		{gvrMachineDeployment, "MachineDeployment"},
		{gvrKCP, "KubeadmControlPlane"},
		{gvrOpenStackCluster, "OpenStackCluster"},
	}

	for _, r := range watchResources {
		informer := w.factory.ForResource(r.gvr).Informer()
		kind := r.kind

		informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				w.handleEvent("ADDED", kind, obj)
			},
			UpdateFunc: func(_, newObj interface{}) {
				w.handleEvent("MODIFIED", kind, newObj)
			},
			DeleteFunc: func(obj interface{}) {
				w.handleEvent("DELETED", kind, obj)
			},
		})
	}

	w.factory.Start(ctx.Done())

	// 캐시 동기화 대기
	w.factory.WaitForCacheSync(ctx.Done())

	w.logger.Info("cluster watcher started")
	return nil
}

func (w *ClusterWatcher) handleEvent(eventType, kind string, obj interface{}) {
	type namedObject interface {
		GetName() string
		GetNamespace() string
	}
	unstr, ok := obj.(namedObject)
	if !ok {
		return
	}

	evt := ClusterEvent{
		ClusterName: unstr.GetName(),
		Namespace:   unstr.GetNamespace(),
		Kind:        kind,
		EventType:   eventType,
	}

	select {
	case w.EventChan <- evt:
	default:
		w.logger.Warn("cluster event channel is full, dropping event",
			zap.String("kind", kind),
			zap.String("name", unstr.GetName()))
	}
}

// Stop은 Watcher를 중지합니다.
func (w *ClusterWatcher) Stop() {
	if w.factory != nil {
		w.factory.Shutdown()
	}
	close(w.EventChan)
}

// GetClusterPhase는 CAPI Cluster 오브젝트의 phase를 반환합니다.
func GetClusterPhase(obj map[string]interface{}) string {
	phase, _, _ := getNestedString(obj, "status", "phase")
	return phase
}

func getNestedString(obj map[string]interface{}, fields ...string) (string, bool, error) {
	current := obj
	for i, field := range fields {
		v, ok := current[field]
		if !ok {
			return "", false, nil
		}
		if i == len(fields)-1 {
			s, ok := v.(string)
			return s, ok, nil
		}
		current, ok = v.(map[string]interface{})
		if !ok {
			return "", false, fmt.Errorf("field %s is not a map", field)
		}
	}
	return "", false, nil
}
