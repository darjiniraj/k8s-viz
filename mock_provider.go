package main

import rbacv1 "k8s.io/api/rbac/v1"

type MockProvider struct{}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (m *MockProvider) SecurityRows() []SecurityRow {
	return []SecurityRow{
		{
			SA:          "payments-api",
			Namespace:   "payments",
			IAMRole:     "arn:aws:iam::111122223333:role/eks-irsa-payments-api",
			BindingType: "RoleBinding",
			BindingName: "payments-api-reader-binding",
			BindingYAML: `apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: payments-api-reader-binding
  namespace: payments
subjects:
- kind: ServiceAccount
  name: payments-api
  namespace: payments
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: payments-api-reader`,
			RoleYAML: `apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: payments-api-reader
  namespace: payments
rules:
- apiGroups: [""]
  resources: ["pods", "services", "configmaps"]
  verbs: ["get", "list", "watch"]`,
			RoleName: "payments-api-reader",
			RoleKind: "Role",
		},
		{
			SA:          "inventory-sync",
			Namespace:   "ops",
			IAMRole:     "arn:aws:iam::111122223333:role/eks-podid-inventory-sync",
			BindingType: "ClusterRoleBinding",
			BindingName: "inventory-sync-view",
			BindingYAML: `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: inventory-sync-view
subjects:
- kind: ServiceAccount
  name: inventory-sync
  namespace: ops
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: view`,
			RoleYAML: `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: view
rules:
- apiGroups: [""]
  resources: ["pods", "services", "endpoints", "persistentvolumeclaims"]
  verbs: ["get", "list", "watch"]`,
			RoleName: "view",
			RoleKind: "ClusterRole",
			IsGlobal: true,
		},
		{
			SA:          "metrics-agent",
			Namespace:   "monitoring",
			IAMRole:     "None",
			BindingType: "RoleBinding",
			BindingName: "metrics-agent-reader",
			BindingYAML: `apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: metrics-agent-reader
  namespace: monitoring
subjects:
- kind: ServiceAccount
  name: metrics-agent
  namespace: monitoring
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: metrics-read`,
			RoleYAML: `apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: metrics-read
  namespace: monitoring
rules:
- apiGroups: [""]
  resources: ["pods", "nodes", "events"]
  verbs: ["get", "list", "watch"]`,
			RoleName: "metrics-read",
			RoleKind: "Role",
		},
	}
}

func (m *MockProvider) GroupRows() []GroupSecurityRow {
	return []GroupSecurityRow{
		{
			GroupName:  "eks:platform-admins",
			Roles:      []string{"cluster-admin"},
			Namespaces: []string{"Cluster-Wide"},
			AllYAMLs: []YamlBlock{
				{
					Kind:      "ClusterRoleBinding",
					Name:      "platform-admins-cluster-admin",
					Namespace: "Cluster-Wide",
					Data: `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: platform-admins-cluster-admin
subjects:
- kind: Group
  name: eks:platform-admins
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin`,
				},
				{
					Kind:      "ClusterRole",
					Name:      "cluster-admin",
					Namespace: "Cluster-Wide",
					Data: `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cluster-admin
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["*"]`,
				},
			},
		},
		{
			GroupName:  "eks:readonly-auditors",
			Roles:      []string{"view"},
			Namespaces: []string{"payments", "ops"},
			AllYAMLs: []YamlBlock{
				{
					Kind:      "RoleBinding",
					Name:      "auditors-view",
					Namespace: "payments",
					Data: `apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: auditors-view
  namespace: payments
subjects:
- kind: Group
  name: eks:readonly-auditors
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: view`,
				},
				{
					Kind:      "ClusterRole",
					Name:      "view",
					Namespace: "Cluster-Wide",
					Data: `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: view
rules:
- apiGroups: [""]
  resources: ["pods", "services", "configmaps"]
  verbs: ["get", "list", "watch"]`,
				},
			},
		},
	}
}

func (m *MockProvider) CiliumRows() []CiliumPolicyRow {
	return []CiliumPolicyRow{
		{
			Name:           "allow-payments-to-db",
			Namespace:      "payments",
			Kind:           "CiliumNetworkPolicy",
			IsClusterWide:  false,
			TargetSelector: "app=payments-api",
			Type:           "Egress Only",
			Yaml: `apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: allow-payments-to-db
  namespace: payments
spec:
  endpointSelector:
    matchLabels:
      app: payments-api
  egress:
  - toEndpoints:
    - matchLabels:
        app: postgres`,
		},
		{
			Name:           "deny-all-except-dns",
			Namespace:      "",
			Kind:           "CiliumClusterwideNetworkPolicy",
			IsClusterWide:  true,
			TargetSelector: "All Endpoints",
			Type:           "Ingress + Egress",
			Yaml: `apiVersion: cilium.io/v2
kind: CiliumClusterwideNetworkPolicy
metadata:
  name: deny-all-except-dns
spec:
  endpointSelector: {}
  egress:
  - toEndpoints:
    - matchLabels:
        k8s:io.kubernetes.pod.namespace: kube-system`,
		},
	}
}

func (m *MockProvider) IAMAuditRows() []IAMRBACMapRow {
	viewRules := []rbacv1.PolicyRule{
		{
			Verbs:     []string{"get", "list", "watch"},
			APIGroups: []string{""},
			Resources: []string{"pods", "services", "configmaps"},
		},
	}
	adminRules := []rbacv1.PolicyRule{
		{
			Verbs:     []string{"*"},
			APIGroups: []string{"*"},
			Resources: []string{"*"},
		},
	}

	return []IAMRBACMapRow{
		{
			IAMPrincipal:   "arn:aws:iam::111122223333:role/eks-accessentry-platform-admins",
			AttachmentType: "AccessEntry",
			AccessPolicies: []string{"AmazonEKSClusterAdminPolicy"},
			K8sSubject:     "eks:platform-admins",
			RBACDetails: []IAMRBACDetail{
				{
					BindingKind: "ClusterRoleBinding",
					BindingName: "platform-admins-cluster-admin",
					BindingYAML: `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: platform-admins-cluster-admin
subjects:
- kind: Group
  name: eks:platform-admins
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin`,
					RoleKind: "ClusterRole",
					RoleName: "cluster-admin",
					RoleYAML: `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cluster-admin
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["*"]`,
					Rules: adminRules,
				},
			},
			SummaryPlaceholder: "Platform admins can perform any action across all API groups and resources.",
		},
		{
			IAMPrincipal:   "arn:aws:iam::111122223333:role/eks-podid-inventory-sync",
			AttachmentType: "PodIdentity",
			K8sSubject:     "ops/inventory-sync",
			RBACDetails: []IAMRBACDetail{
				{
					BindingKind: "ClusterRoleBinding",
					BindingName: "inventory-sync-view",
					BindingYAML: `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: inventory-sync-view
subjects:
- kind: ServiceAccount
  name: inventory-sync
  namespace: ops
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: view`,
					RoleKind: "ClusterRole",
					RoleName: "view",
					RoleYAML: `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: view
rules:
- apiGroups: [""]
  resources: ["pods","services","configmaps"]
  verbs: ["get","list","watch"]`,
					Rules: viewRules,
				},
			},
			SummaryPlaceholder: "Inventory sync can read workload resources cluster-wide through the view cluster role.",
		},
		{
			IAMPrincipal:   "arn:aws:iam::111122223333:role/eks-irsa-payments-api",
			AttachmentType: "IRSA",
			K8sSubject:     "payments/payments-api",
			RBACDetails: []IAMRBACDetail{
				{
					BindingKind:      "RoleBinding",
					BindingName:      "payments-api-reader-binding",
					BindingNamespace: "payments",
					BindingYAML: `apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: payments-api-reader-binding
  namespace: payments
subjects:
- kind: ServiceAccount
  name: payments-api
  namespace: payments
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: payments-api-reader`,
					RoleKind:      "Role",
					RoleName:      "payments-api-reader",
					RoleNamespace: "payments",
					RoleYAML: `apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: payments-api-reader
  namespace: payments
  annotations:
    description: "Read-only access for payments service account"
rules:
- apiGroups: [""]
  resources: ["pods","services","configmaps"]
  verbs: ["get","list","watch"]`,
					Rules: viewRules,
				},
			},
			SummaryPlaceholder: "Payments API has read-only namespace-scoped access in the payments namespace.",
		},
	}
}
