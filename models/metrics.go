package models

import (
	"time"

	corev1 "k8s.io/api/core/v1"
)

// NodeMetrics is a point-in-time usage sample for a node.
type NodeMetrics struct {
	Name      string
	Timestamp time.Time
	Window    time.Duration
	CPU       int64 // millicores used
	Memory    int64 // bytes used

	// Capacity/allocatable are filled when available, for percentages.
	CPUAllocatable    int64
	MemoryAllocatable int64
}

// ContainerMetrics is usage for a single container.
type ContainerMetrics struct {
	Name   string
	CPU    int64 // millicores
	Memory int64 // bytes
}

// PodMetrics is a point-in-time usage sample for a pod.
type PodMetrics struct {
	Namespace  string
	Name       string
	Timestamp  time.Time
	Window     time.Duration
	Containers []ContainerMetrics
	CPU        int64 // millicores, summed
	Memory     int64 // bytes, summed
}

// MetricsSnapshot is one sample of the whole cluster.
type MetricsSnapshot struct {
	Timestamp time.Time
	Nodes     []NodeMetrics
	Pods      []PodMetrics
}

// MetricsUpdate is delivered by StreamMetrics. Exactly one of
// Snapshot or Err is set.
type MetricsUpdate struct {
	Snapshot *MetricsSnapshot
	Err      error
}

// ResourceEventType mirrors watch event types.
type ResourceEventType string

const (
	ResourceAdded    ResourceEventType = "ADDED"
	ResourceModified ResourceEventType = "MODIFIED"
	ResourceDeleted  ResourceEventType = "DELETED"
)

// PodEvent is a live informer-driven pod change.
type PodEvent struct {
	Type ResourceEventType
	Pod  *corev1.Pod
}

// NodeEvent is a live informer-driven node change.
type NodeEvent struct {
	Type ResourceEventType
	Node *corev1.Node
}
