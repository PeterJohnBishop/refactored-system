// Package kubesvc hydrates cluster metrics
package kubesvc

import (
	"context"
	"fmt"

	"github.com/peterjohnbishop/refactored-system/models"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// NewKubesvcClient initializes and returns a Kubernetes clientset.
func NewKubesvcClient() (*kubernetes.Clientset, error) {
	// Load kubeconfig using the active kubectl context.
	rules := clientcmd.NewDefaultClientConfigLoadingRules()

	config, err := clientcmd.
		NewNonInteractiveDeferredLoadingClientConfig(
			rules,
			&clientcmd.ConfigOverrides{},
		).
		ClientConfig()
	if err != nil {
		return nil, err
	}

	// Create the Kubernetes client.
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

// GetClusterSnapshot retrieves all supported Kubernetes resources.
func GetClusterSnapshot(
	ctx context.Context,
	client *kubernetes.Clientset,
) (*models.ClusterSnapshot, error) {
	snapshot := &models.ClusterSnapshot{}

	// Nodes
	nodes, err := ListNodes(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}
	snapshot.Nodes = nodes

	// Namespaces
	namespaces, err := ListNamespaces(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}
	snapshot.Namespaces = namespaces

	// Pods
	pods, err := ListPods(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}
	snapshot.Pods = pods

	// Deployments
	deployments, err := ListDeployments(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}
	snapshot.Deployments = deployments

	// Services
	services, err := ListServices(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}
	snapshot.Services = services

	// StatefulSets
	statefulsets, err := ListStatefulSets(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to list statefulsets: %w", err)
	}
	snapshot.StatefulSets = statefulsets

	// DaemonSets
	daemonsets, err := ListDaemonSets(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to list daemonsets: %w", err)
	}
	snapshot.DaemonSets = daemonsets

	// ConfigMaps
	configmaps, err := ListConfigMaps(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to list configmaps: %w", err)
	}
	snapshot.ConfigMaps = configmaps

	// PersistentVolumeClaims
	pvcs, err := ListPVCs(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to list PVCs: %w", err)
	}
	snapshot.PVCs = pvcs

	// PersistentVolumes
	pvs, err := ListPVs(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to list PVs: %w", err)
	}
	snapshot.PVs = pvs

	// ServiceAccounts
	serviceAccounts, err := ListServiceAccounts(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to list service accounts: %w", err)
	}
	snapshot.ServiceAccounts = serviceAccounts

	// ResourceQuotas
	resourceQuotas, err := ListResourceQuotas(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to list resource quotas: %w", err)
	}
	snapshot.ResourceQuotas = resourceQuotas

	// Jobs
	jobs, err := ListJobs(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}
	snapshot.Jobs = jobs

	// CronJobs
	cronjobs, err := ListCronJobs(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to list cronjobs: %w", err)
	}
	snapshot.CronJobs = cronjobs

	// Ingresses
	ingresses, err := ListIngresses(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to list ingresses: %w", err)
	}
	snapshot.Ingresses = ingresses

	// StorageClasses
	storageClasses, err := ListStorageClasses(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to list storage classes: %w", err)
	}
	snapshot.StorageClasses = storageClasses

	return snapshot, nil
}

// ListNodes retrieves all nodes in the cluster.
func ListNodes(
	ctx context.Context,
	client *kubernetes.Clientset,
) ([]corev1.Node, error) {
	nodes, err := client.CoreV1().
		Nodes().
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return nodes.Items, nil
}

// ListNamespaces retrieves all namespaces.
func ListNamespaces(
	ctx context.Context,
	client *kubernetes.Clientset,
) ([]corev1.Namespace, error) {
	namespaces, err := client.CoreV1().
		Namespaces().
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return namespaces.Items, nil
}

// ListPods retrieves pods across all namespaces.
func ListPods(
	ctx context.Context,
	client *kubernetes.Clientset,
) ([]corev1.Pod, error) {
	pods, err := client.CoreV1().
		Pods(metav1.NamespaceAll).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return pods.Items, nil
}

// ListDeployments retrieves deployments across all namespaces.
func ListDeployments(
	ctx context.Context,
	client *kubernetes.Clientset,
) ([]appsv1.Deployment, error) {
	deployments, err := client.AppsV1().
		Deployments(metav1.NamespaceAll).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return deployments.Items, nil
}

// ListServices retrieves services across all namespaces.
func ListServices(
	ctx context.Context,
	client *kubernetes.Clientset,
) ([]corev1.Service, error) {
	services, err := client.CoreV1().
		Services(metav1.NamespaceAll).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return services.Items, nil
}

// ListStatefulSets retrieves StatefulSets across all namespaces.
func ListStatefulSets(
	ctx context.Context,
	client *kubernetes.Clientset,
) ([]appsv1.StatefulSet, error) {
	statefulsets, err := client.AppsV1().
		StatefulSets(metav1.NamespaceAll).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return statefulsets.Items, nil
}

// ListDaemonSets retrieves DaemonSets across all namespaces.
func ListDaemonSets(
	ctx context.Context,
	client *kubernetes.Clientset,
) ([]appsv1.DaemonSet, error) {
	daemonsets, err := client.AppsV1().
		DaemonSets(metav1.NamespaceAll).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return daemonsets.Items, nil
}

// ListConfigMaps retrieves ConfigMaps across all namespaces.
func ListConfigMaps(
	ctx context.Context,
	client *kubernetes.Clientset,
) ([]corev1.ConfigMap, error) {
	configmaps, err := client.CoreV1().
		ConfigMaps(metav1.NamespaceAll).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return configmaps.Items, nil
}

// ListSecrets retrieves Secrets across all namespaces.
func ListSecrets(
	ctx context.Context,
	client *kubernetes.Clientset,
) ([]corev1.Secret, error) {
	secrets, err := client.CoreV1().
		Secrets(metav1.NamespaceAll).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return secrets.Items, nil
}

// ListPVCs retrieves PersistentVolumeClaims across all namespaces.
func ListPVCs(
	ctx context.Context,
	client *kubernetes.Clientset,
) ([]corev1.PersistentVolumeClaim, error) {
	pvcs, err := client.CoreV1().
		PersistentVolumeClaims(metav1.NamespaceAll).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return pvcs.Items, nil
}

// ListPVs retrieves cluster-wide PersistentVolumes.
func ListPVs(
	ctx context.Context,
	client *kubernetes.Clientset,
) ([]corev1.PersistentVolume, error) {
	pvs, err := client.CoreV1().
		PersistentVolumes().
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return pvs.Items, nil
}

// ListServiceAccounts retrieves ServiceAccounts.
func ListServiceAccounts(
	ctx context.Context,
	client *kubernetes.Clientset,
) ([]corev1.ServiceAccount, error) {
	accounts, err := client.CoreV1().
		ServiceAccounts(metav1.NamespaceAll).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return accounts.Items, nil
}

// ListResourceQuotas retrieves ResourceQuotas.
func ListResourceQuotas(
	ctx context.Context,
	client *kubernetes.Clientset,
) ([]corev1.ResourceQuota, error) {
	quotas, err := client.CoreV1().
		ResourceQuotas(metav1.NamespaceAll).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return quotas.Items, nil
}

// ListJobs retrieves Jobs across all namespaces.
func ListJobs(
	ctx context.Context,
	client *kubernetes.Clientset,
) ([]batchv1.Job, error) {
	jobs, err := client.BatchV1().
		Jobs(metav1.NamespaceAll).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return jobs.Items, nil
}

// ListCronJobs retrieves CronJobs across all namespaces.
func ListCronJobs(
	ctx context.Context,
	client *kubernetes.Clientset,
) ([]batchv1.CronJob, error) {
	cronjobs, err := client.BatchV1().
		CronJobs(metav1.NamespaceAll).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return cronjobs.Items, nil
}

// ListIngresses retrieves Ingress resources.
func ListIngresses(
	ctx context.Context,
	client *kubernetes.Clientset,
) ([]networkingv1.Ingress, error) {
	ingresses, err := client.NetworkingV1().
		Ingresses(metav1.NamespaceAll).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return ingresses.Items, nil
}

// ListStorageClasses retrieves StorageClasses.
func ListStorageClasses(
	ctx context.Context,
	client *kubernetes.Clientset,
) ([]storagev1.StorageClass, error) {
	classes, err := client.StorageV1().
		StorageClasses().
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return classes.Items, nil
}
