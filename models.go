package main

// SecurityRow represents a single mapping for Service Accounts (Apps view).
type SecurityRow struct {
	SA          string `json:"sa"`
	Namespace   string `json:"namespace"`
	IAMRole     string `json:"iam_role"`
	BindingType string `json:"binding_type"`
	BindingName string `json:"binding_name"`
	BindingYAML string `json:"binding_yaml"`
	RoleYAML    string `json:"role_yaml"`
	RoleName    string `json:"role"`
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
