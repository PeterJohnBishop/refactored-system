package kubesvc

import (
	"context"
	"strings"

	"github.com/peterjohnbishop/refactored-system/models"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
)

func ListAPIEndpoints(
	client *kubernetes.Clientset,
) ([]models.APIEndpoint, error) {
	_, resourceLists, err := client.Discovery().
		ServerGroupsAndResources()

	if err != nil && len(resourceLists) == 0 {
		return nil, err
	}

	var endpoints []models.APIEndpoint

	for _, list := range resourceLists {
		gv, parseErr := schema.ParseGroupVersion(
			list.GroupVersion,
		)
		if parseErr != nil {
			return nil, parseErr
		}

		prefix := "/apis/" + list.GroupVersion
		if gv.Group == "" {
			prefix = "/api/" + gv.Version
		}

		for _, resource := range list.APIResources {
			// Skip subresources such as pods/log.
			if strings.Contains(resource.Name, "/") {
				continue
			}

			path := prefix

			if resource.Namespaced {
				path += "/namespaces/{namespace}"
			}

			path += "/" + resource.Name

			endpoints = append(endpoints, models.APIEndpoint{
				Kind:       resource.Kind,
				Path:       path,
				Namespaced: resource.Namespaced,
				Verbs:      []string(resource.Verbs),
			})
		}
	}

	// May return partial results alongside a
	// discovery error from an unavailable API group.
	return endpoints, err
}

func ListEndpointSlices(
	ctx context.Context,
	client *kubernetes.Clientset,
) ([]discoveryv1.EndpointSlice, error) {
	slices, err := client.DiscoveryV1().
		EndpointSlices(metav1.NamespaceAll).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return slices.Items, nil
}
