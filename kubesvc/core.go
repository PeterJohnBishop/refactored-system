package kubesvc

import (
	"context"
	"fmt"

	"github.com/peterjohnbishop/refactored-system/models"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const systemNamespace = "kube-system"

// GetCoreComponents gathers control plane and core add-on details.
// Individual lookups that fail do not abort the whole call: partial
// data is returned alongside the first error encountered.
func GetCoreComponents(
	ctx context.Context,
	client *kubernetes.Clientset,
) (*models.CoreComponents, error) {
	core := &models.CoreComponents{}
	var firstErr error
	note := func(what string, err error) {
		if err != nil && firstErr == nil {
			firstErr = fmt.Errorf("%s: %w", what, err)
		}
	}

	v, err := client.Discovery().ServerVersion()
	note("server version", err)
	core.ServerVersion = v

	core.ComponentStatuses, err = ListComponentStatuses(ctx, client)
	note("component statuses", err)

	core.Health = GetAPIServerHealth(ctx, client)

	// Control plane pods are labelled by "component" (kubeadm and
	// most distros) in kube-system; add-ons use k8s-app.
	pods := []struct {
		dst      *[]corev1.Pod
		selector string
	}{
		{&core.APIServerPods, "component=kube-apiserver"},
		{&core.EtcdPods, "component=etcd"},
		{&core.SchedulerPods, "component=kube-scheduler"},
		{&core.ControllerManagerPods, "component=kube-controller-manager"},
		{&core.KubeProxyPods, "k8s-app=kube-proxy"},
		{&core.DNSPods, "k8s-app=kube-dns"},
	}
	for _, p := range pods {
		*p.dst, err = listSystemPods(ctx, client, p.selector)
		note(p.selector, err)
	}

	return core, firstErr
}

// ListComponentStatuses returns the (deprecated) componentstatuses.
func ListComponentStatuses(
	ctx context.Context,
	client *kubernetes.Clientset,
) ([]corev1.ComponentStatus, error) {
	list, err := client.CoreV1().
		ComponentStatuses().
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

// GetAPIServerHealth queries the API server health endpoints and
// returns the body of each, or the error text if the call failed.
func GetAPIServerHealth(
	ctx context.Context,
	client *kubernetes.Clientset,
) map[string]string {
	out := map[string]string{}

	for _, path := range []string{"/livez", "/readyz", "/healthz"} {
		body, err := client.Discovery().RESTClient().
			Get().
			AbsPath(path).
			DoRaw(ctx)
		if err != nil {
			out[path] = "error: " + err.Error()
			continue
		}
		out[path] = string(body)
	}

	return out
}

func listSystemPods(
	ctx context.Context,
	client *kubernetes.Clientset,
	selector string,
) ([]corev1.Pod, error) {
	pods, err := client.CoreV1().
		Pods(systemNamespace).
		List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, err
	}

	return pods.Items, nil
}
