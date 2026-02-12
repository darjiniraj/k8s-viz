# k8s-wiz

Kubernetes + IAM security visualizer for:
- Service Accounts
- User Groups
- Cilium Policies
- IAM to Kubernetes RBAC mapping

## Run (Live Mode)
Requires valid `kubeconfig` (and AWS creds for IAM audit endpoint).

```bash
go run .
```

## Run (Demo Mode)
No Kubernetes cluster and no AWS account required.

```bash
DEMO_MODE=true go run .
```

On Windows PowerShell:

```powershell
$env:DEMO_MODE="true"
go run .
```

When Demo Mode is enabled, backend serves mock datasets for:
- `/api/table`
- `/api/groups`
- `/api/cilium`
- `/api/iam-audit`

The response includes header `X-App-Mode: DEMO`.

