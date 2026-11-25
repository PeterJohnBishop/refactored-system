# refactored-system

Nodes, Pods, deployments, events, etc. - k8s.io/client-go

Cluster resource metrics - k8s.io/metrics

Live updates client-go/informers


## Live metrics setup

Live metrics (`StreamMetrics`, `GetMetricsSnapshot`) need metrics-server in the cluster.

1. Install metrics-server:

   ```sh
   kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
   ```

2. Local/dev clusters (kind, etc.) only: the kubelet certificates have no IP SANs, so metrics-server fails to scrape nodes and never becomes ready. Patch it to skip kubelet TLS verification:

   ```sh
   kubectl -n kube-system patch deployment metrics-server --type=json \
     -p='[{"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--kubelet-insecure-tls"}]'
   ```

   Don't use this flag in production; fix the kubelet serving certificates instead.

Check it's working with `kubectl top nodes`. It can take about 30 seconds after the patch.
