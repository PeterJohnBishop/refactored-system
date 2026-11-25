package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/peterjohnbishop/refactored-system/kubesvc"
)

func main() {
	fmt.Println("refactored-system initializing...")

	// Initialize the Kubernetes client.
	clientset, err := kubesvc.NewKubesvcClient()
	if err != nil {
		log.Fatalf("Failed to connect to Kubernetes: %v", err)
	}

	version, _ := clientset.Discovery().ServerVersion()

	fmt.Println(version.GitVersion)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	snapshot, err := kubesvc.GetClusterSnapshot(ctx, clientset)
	if err != nil {
		log.Fatalf("Failed to retrieve cluster snapshot: %v", err)
	}

	fmt.Printf("Nodes: %d\n", len(snapshot.Nodes))
	fmt.Printf("Pods: %d\n", len(snapshot.Pods))
	fmt.Printf("Deployments: %d\n", len(snapshot.Deployments))
	fmt.Printf("Services: %d\n", len(snapshot.Services))
	fmt.Printf("ConfigMaps: %d\n", len(snapshot.ConfigMaps))
	fmt.Printf("PVCs: %d\n", len(snapshot.PVCs))
	fmt.Printf("PVs: %d\n", len(snapshot.PVs))
	fmt.Printf("Jobs: %d\n", len(snapshot.Jobs))
	fmt.Printf("CronJobs: %d\n", len(snapshot.CronJobs))
	fmt.Printf("Runtimes: %d\n", len(snapshot.NodeRuntimes))
	fmt.Printf("Endpoints: %d\n", len(snapshot.APIEndpoints))
	fmt.Printf("Endpoint Slices: %d\n", len(snapshot.EndpointSlices))

	if c := snapshot.Core; c != nil {
		fmt.Printf("API server pods: %d, etcd pods: %d, scheduler pods: %d, "+
			"controller-manager pods: %d, kube-proxy pods: %d, DNS pods: %d\n",
			len(c.APIServerPods), len(c.EtcdPods), len(c.SchedulerPods),
			len(c.ControllerManagerPods), len(c.KubeProxyPods), len(c.DNSPods))
		fmt.Printf("Readyz: %s\n", c.Health["/readyz"])
	}

	// Live metrics (requires metrics-server). Streams until Ctrl-C.
	mc, err := kubesvc.NewMetricsClient()
	if err != nil {
		log.Fatalf("Failed to create metrics client: %v", err)
	}

	if !kubesvc.MetricsAvailable(clientset) {
		log.Fatalf("Skipping live metrics: %v", kubesvc.ErrMetricsUnavailable)
	}

	sctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	for u := range kubesvc.StreamMetrics(sctx, mc, clientset, 15*time.Second) {
		if u.Err != nil {
			log.Printf("metrics error: %v", u.Err)
			continue
		}
		fmt.Printf("[%s] node samples: %d, pod samples: %d\n",
			u.Snapshot.Timestamp.Format(time.TimeOnly),
			len(u.Snapshot.Nodes), len(u.Snapshot.Pods))
	}
}
