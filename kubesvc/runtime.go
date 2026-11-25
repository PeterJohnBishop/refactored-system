package kubesvc

import (
	"context"

	"github.com/peterjohnbishop/refactored-system/models"
	"k8s.io/client-go/kubernetes"
)

// ListNodeRuntimes returns runtime information for each node.
func ListNodeRuntimes(
	ctx context.Context,
	client *kubernetes.Clientset,
) ([]models.NodeRuntimeInfo, error) {
	nodes, err := ListNodes(ctx, client)
	if err != nil {
		return nil, err
	}

	result := make([]models.NodeRuntimeInfo, 0, len(nodes))

	for _, node := range nodes {
		info := node.Status.NodeInfo

		result = append(result, models.NodeRuntimeInfo{
			NodeName:         node.Name,
			ContainerRuntime: info.ContainerRuntimeVersion,
			KubeletVersion:   info.KubeletVersion,
			OSImage:          info.OSImage,
			OperatingSystem:  info.OperatingSystem,
			Architecture:     info.Architecture,
			KernelVersion:    info.KernelVersion,
		})
	}

	return result, nil
}
