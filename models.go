package main

// SecurityRow represents a single mapping for Service Accounts (Apps view).
type SecurityRow struct {
	SA          string `json:"sa"`
	Namespace   string `json:"namespace"`
	IAMRole     string `json:"iam_role"`
	BindingType string `json:"binding_type"` // e.g., "ClusterRoleBinding"
	BindingName string `json:"binding_name"`
	BindingYAML string `json:"binding_yaml"`
	RoleYAML    string `json:"role_yaml"`
	RoleName    string `json:"role_name"`
	RoleKind    string `json:"role_kind"` // "Role" or "ClusterRole"
	IsGlobal    bool   `json:"is_global"` // For showing the "Cluster-Wide" badge
}

// YamlBlock is a metadata wrapper for K8s manifests in the Group view.
type YamlBlock struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Data      string `json:"data"`
	Namespace string `json:"namespace"`
}

// GroupSecurityRow aggregates all permissions for a specific User Group.
type GroupSecurityRow struct {
	GroupName  string      `json:"group_name"`
	Roles      []string    `json:"roles"`
	Namespaces []string    `json:"namespaces"`
	AllYAMLs   []YamlBlock `json:"all_yamls"`
}

type CiliumPolicyRow struct {
	Name           string `json:"name"`
	Namespace      string `json:"namespace"` // Empty for Clusterwide
	Kind           string `json:"kind"`      // CiliumNetworkPolicy or CiliumClusterwideNetworkPolicy
	IsClusterWide  bool   `json:"is_cluster_wide"`
	TargetSelector string `json:"target_selector"`
	Type           string `json:"type"` // Ingress, Egress, or Both
	Yaml           string `json:"yaml"`
}

type IAMAuditRow struct {
	IAMRole   string      `json:"iam_role"`
	Type      string      `json:"type"` // Set to "iam" for UI badges
	AISummary string      `json:"ai_summary"`
	AllYAMLs  []YamlBlock `json:"all_yamls"`
}
