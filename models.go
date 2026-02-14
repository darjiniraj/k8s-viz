package main

import rbacv1 "k8s.io/api/rbac/v1"

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

// IAMRBACDetail represents a single RBAC attachment chain for a subject.
type IAMRBACDetail struct {
	BindingKind      string              `json:"binding_kind"`
	BindingName      string              `json:"binding_name"`
	BindingNamespace string              `json:"binding_namespace,omitempty"`
	BindingYAML      string              `json:"binding_yaml"`
	RoleKind         string              `json:"role_kind"`
	RoleName         string              `json:"role_name"`
	RoleNamespace    string              `json:"role_namespace,omitempty"`
	RoleYAML         string              `json:"role_yaml"`
	Rules            []rbacv1.PolicyRule `json:"rules"`
}

// IAMRBACMapRow is the UI-oriented shape for IAM -> K8s RBAC mapping.
type IAMRBACMapRow struct {
	IAMPrincipal       string          `json:"iam_principal"`
	AttachmentType     string          `json:"attachment_type"` // AccessEntry, PodIdentity, IRSA
	AccessPolicies     []string        `json:"access_policies,omitempty"`
	K8sSubject         string          `json:"k8s_subject"`
	RBACDetails        []IAMRBACDetail `json:"rbac_details"`
	SummaryPlaceholder string          `json:"summary_placeholder"`
}
