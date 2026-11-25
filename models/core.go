package models

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/version"
)

// CoreComponents describes the control plane and core add-ons.
type CoreComponents struct {
	ServerVersion *version.Info

	// ComponentStatuses is deprecated upstream and empty on many
	// managed clusters; Health below is the reliable signal.
	ComponentStatuses []corev1.ComponentStatus

	// Health maps API server health endpoints (/livez, /readyz,
	// /healthz) to their response body, or the error text.
	Health map[string]string

	APIServerPods         []corev1.Pod
	EtcdPods              []corev1.Pod
	SchedulerPods         []corev1.Pod
	ControllerManagerPods []corev1.Pod
	KubeProxyPods         []corev1.Pod
	DNSPods               []corev1.Pod
}
