// Package models defines global data types
package models

import (
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	networkingv1 "k8s.io/api/networking/v1"
	storagev1 "k8s.io/api/storage/v1"
)

type ClusterSnapshot struct {
	NodeRuntimes []NodeRuntimeInfo

	Nodes           []corev1.Node
	Namespaces      []corev1.Namespace
	Pods            []corev1.Pod
	Services        []corev1.Service
	ConfigMaps      []corev1.ConfigMap
	PVCs            []corev1.PersistentVolumeClaim
	PVs             []corev1.PersistentVolume
	ServiceAccounts []corev1.ServiceAccount
	ResourceQuotas  []corev1.ResourceQuota

	Deployments  []appsv1.Deployment
	StatefulSets []appsv1.StatefulSet
	DaemonSets   []appsv1.DaemonSet

	Jobs     []batchv1.Job
	CronJobs []batchv1.CronJob

	Ingresses      []networkingv1.Ingress
	StorageClasses []storagev1.StorageClass

	APIEndpoints   []APIEndpoint
	EndpointSlices []discoveryv1.EndpointSlice

	EtcdPods []corev1.Pod
}
