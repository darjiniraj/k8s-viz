# k8s-wiz UI

Vue 3 + Vite frontend for k8s-wiz.

## Run

```bash
npm install
npm run dev
```

## Feature Flags (K8s Resources)

Set these `VITE_` env vars to control which sub-tabs are shown under **K8s Resources**:

- `VITE_ENABLE_SERVICE_ACCOUNT_VIEW` (`true` by default)
- `VITE_ENABLE_USER_GROUP_VIEW` (`true` by default)
- `VITE_ENABLE_CILIUM_VIEW` (`true` by default)

Examples:

```bash
VITE_ENABLE_CILIUM_VIEW=false npm run dev
```

PowerShell:

```powershell
$env:VITE_ENABLE_CILIUM_VIEW="false"
npm run dev
```

