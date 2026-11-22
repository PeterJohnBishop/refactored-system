package main

import (
	"context"
	"fmt"
	"log"
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
}
