# k8s-viz 🚀

Kubernetes RBAC visualizer that maps Service Accounts and User Groups to their IAM roles and permissions.

## Features
- **Bulk Fetching**: Optimized to query the K8s API in 5 parallel calls.
- **IAM Mapping**: Automatically detects EKS IAM role annotations.
- **YAML Inspector**: View live manifests for Roles and Bindings directly in the UI.

## Getting Started
1. Ensure your `kubeconfig` is pointing to a cluster.
2. Run the server:
   ```bash
   go run .