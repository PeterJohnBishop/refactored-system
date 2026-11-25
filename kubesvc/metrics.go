package kubesvc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/peterjohnbishop/refactored-system/models"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

// ErrMetricsUnavailable means the cluster doesn't serve
// metrics.k8s.io, usually because metrics-server isn't installed.
var ErrMetricsUnavailable = errors.New(
	"metrics.k8s.io API not available; install metrics-server " +
		"(kubectl apply -f https://github.com/kubernetes-sigs/" +
		"metrics-server/releases/latest/download/components.yaml)",
)

// MetricsAvailable reports whether the metrics API is served.
func MetricsAvailable(client *kubernetes.Clientset) bool {
	_, err := client.Discovery().
		ServerResourcesForGroupVersion("metrics.k8s.io/v1beta1")

	return err == nil
}

// NewMetricsClient creates a metrics.k8s.io client from the active
// kubeconfig context. The cluster must run metrics-server.
func NewMetricsClient() (*metricsclient.Clientset, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, err
	}

	return metricsclient.NewForConfig(config)
}

// GetNodeMetrics fetches current usage for all nodes. Allocatable
// values are filled in from the node objects when client is non-nil.
func GetNodeMetrics(
	ctx context.Context,
	mc *metricsclient.Clientset,
	client *kubernetes.Clientset,
) ([]models.NodeMetrics, error) {
	list, err := mc.MetricsV1beta1().
		NodeMetricses().
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	alloc := map[string]corev1.ResourceList{}
	if client != nil {
		if nodes, err := ListNodes(ctx, client); err == nil {
			for _, n := range nodes {
				alloc[n.Name] = n.Status.Allocatable
			}
		}
	}

	out := make([]models.NodeMetrics, 0, len(list.Items))
	for _, m := range list.Items {
		nm := models.NodeMetrics{
			Name:      m.Name,
			Timestamp: m.Timestamp.Time,
			Window:    m.Window.Duration,
			CPU:       m.Usage.Cpu().MilliValue(),
			Memory:    m.Usage.Memory().Value(),
		}
		if a, ok := alloc[m.Name]; ok {
			nm.CPUAllocatable = a.Cpu().MilliValue()
			nm.MemoryAllocatable = a.Memory().Value()
		}
		out = append(out, nm)
	}

	return out, nil
}

// GetPodMetrics fetches current usage for pods in a namespace
// (metav1.NamespaceAll for every namespace).
func GetPodMetrics(
	ctx context.Context,
	mc *metricsclient.Clientset,
	namespace string,
) ([]models.PodMetrics, error) {
	list, err := mc.MetricsV1beta1().
		PodMetricses(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	out := make([]models.PodMetrics, 0, len(list.Items))
	for _, m := range list.Items {
		out = append(out, convertPodMetrics(m))
	}

	return out, nil
}

func convertPodMetrics(m metricsv1beta1.PodMetrics) models.PodMetrics {
	pm := models.PodMetrics{
		Namespace: m.Namespace,
		Name:      m.Name,
		Timestamp: m.Timestamp.Time,
		Window:    m.Window.Duration,
	}
	for _, c := range m.Containers {
		cpu := c.Usage.Cpu().MilliValue()
		mem := c.Usage.Memory().Value()
		pm.Containers = append(pm.Containers, models.ContainerMetrics{
			Name: c.Name, CPU: cpu, Memory: mem,
		})
		pm.CPU += cpu
		pm.Memory += mem
	}

	return pm
}

// GetMetricsSnapshot samples node and pod metrics once.
func GetMetricsSnapshot(
	ctx context.Context,
	mc *metricsclient.Clientset,
	client *kubernetes.Clientset,
) (*models.MetricsSnapshot, error) {
	nodes, err := GetNodeMetrics(ctx, mc, client)
	if err != nil {
		return nil, wrapMetricsErr(err)
	}

	pods, err := GetPodMetrics(ctx, mc, metav1.NamespaceAll)
	if err != nil {
		return nil, wrapMetricsErr(err)
	}

	return &models.MetricsSnapshot{
		Timestamp: time.Now(),
		Nodes:     nodes,
		Pods:      pods,
	}, nil
}

func wrapMetricsErr(err error) error {
	if apierrors.IsNotFound(err) {
		return fmt.Errorf("%w: %v", ErrMetricsUnavailable, err)
	}

	return err
}

// StreamMetrics samples metrics every interval and sends each result
// on the returned channel until ctx is cancelled, then closes it.
// metrics-server only refreshes roughly every 15s, so intervals
// below that return repeated samples. A failed sample is delivered
// as an update with Err set and streaming continues, except for
// ErrMetricsUnavailable, which is sent once before the channel closes.
func StreamMetrics(
	ctx context.Context,
	mc *metricsclient.Clientset,
	client *kubernetes.Clientset,
	interval time.Duration,
) <-chan models.MetricsUpdate {
	if interval <= 0 {
		interval = 15 * time.Second
	}

	out := make(chan models.MetricsUpdate)

	go func() {
		defer close(out)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			snap, err := GetMetricsSnapshot(ctx, mc, client)
			select {
			case out <- models.MetricsUpdate{Snapshot: snap, Err: err}:
			case <-ctx.Done():
				return
			}

			// Retrying can't fix a missing metrics API.
			if errors.Is(err, ErrMetricsUnavailable) {
				return
			}

			select {
			case <-ticker.C:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out
}

// WatchPods streams live pod changes via a shared informer. The
// initial list is delivered as ADDED events. The channel closes when
// ctx is cancelled. Events are dropped only if ctx is done.
func WatchPods(
	ctx context.Context,
	client *kubernetes.Clientset,
	namespace string,
) (<-chan models.PodEvent, error) {
	factory := informers.NewSharedInformerFactoryWithOptions(
		client, 0, informers.WithNamespace(namespace),
	)
	informer := factory.Core().V1().Pods().Informer()

	out := make(chan models.PodEvent, 256)
	send := func(t models.ResourceEventType, obj any) {
		if tomb, ok := obj.(cache.DeletedFinalStateUnknown); ok {
			obj = tomb.Obj
		}
		pod, ok := obj.(*corev1.Pod)
		if !ok {
			return
		}
		select {
		case out <- models.PodEvent{Type: t, Pod: pod}:
		case <-ctx.Done():
		}
	}

	_, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(o any) { send(models.ResourceAdded, o) },
		UpdateFunc: func(_, o any) { send(models.ResourceModified, o) },
		DeleteFunc: func(o any) { send(models.ResourceDeleted, o) },
	})
	if err != nil {
		return nil, err
	}

	go func() {
		factory.Start(ctx.Done())
		<-ctx.Done()
		factory.Shutdown()
		close(out)
	}()

	return out, nil
}

// WatchNodes streams live node changes via a shared informer.
func WatchNodes(
	ctx context.Context,
	client *kubernetes.Clientset,
) (<-chan models.NodeEvent, error) {
	factory := informers.NewSharedInformerFactory(client, 0)
	informer := factory.Core().V1().Nodes().Informer()

	out := make(chan models.NodeEvent, 64)
	send := func(t models.ResourceEventType, obj any) {
		if tomb, ok := obj.(cache.DeletedFinalStateUnknown); ok {
			obj = tomb.Obj
		}
		node, ok := obj.(*corev1.Node)
		if !ok {
			return
		}
		select {
		case out <- models.NodeEvent{Type: t, Node: node}:
		case <-ctx.Done():
		}
	}

	_, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(o any) { send(models.ResourceAdded, o) },
		UpdateFunc: func(_, o any) { send(models.ResourceModified, o) },
		DeleteFunc: func(o any) { send(models.ResourceDeleted, o) },
	})
	if err != nil {
		return nil, err
	}

	go func() {
		factory.Start(ctx.Done())
		<-ctx.Done()
		factory.Shutdown()
		close(out)
	}()

	return out, nil
}
