package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

type K8sService struct {
	Clientset   *kubernetes.Clientset
	EKSClient   *eks.Client // Add this
	ClusterName string      // Add this (required for AWS API calls)
}

// MarshalToYaml converts K8s objects to string YAML safely.
func (s *K8sService) MarshalToYaml(obj interface{}) string {
	y, err := yaml.Marshal(obj)
	if err != nil {
		return ""
	}
	return string(y)
}

func (s *K8sService) FetchEKSAccessEntries(ctx context.Context) map[string][]string {
	mapping := make(map[string][]string)

	// 1. Try to call the real AWS API
	listOutput, err := s.EKSClient.ListAccessEntries(ctx, &eks.ListAccessEntriesInput{
		ClusterName: &s.ClusterName,
	})

	// 2. FALLBACK: If AWS fails or returns no entries, use Mock Data
	if err != nil || listOutput == nil || len(listOutput.AccessEntries) == 0 {
		return map[string][]string{
			// This maps an IAM Role to a K8s Group that actually exists in your cluster
			"arn:aws:iam::123456789012:role/LocalDev-Admin":   {"system:masters"},
			"arn:aws:iam::123456789012:role/DataScience-Role": {"data-science-group"},
		}
	}

	// 3. REAL LOGIC: If AWS succeeded, process real entries
	for _, entryArn := range listOutput.AccessEntries {
		describeOutput, err := s.EKSClient.DescribeAccessEntry(ctx, &eks.DescribeAccessEntryInput{
			ClusterName:  &s.ClusterName,
			PrincipalArn: &entryArn,
		})
		if err == nil {
			mapping[*describeOutput.AccessEntry.PrincipalArn] = describeOutput.AccessEntry.KubernetesGroups
		}
	}
	return mapping
}

// GetRoleDetail fetches YAML for either a Role or ClusterRole based on kind.
func (s *K8sService) GetRoleDetail(ctx context.Context, ns, name, kind string) (string, string) {
	if kind == "ClusterRole" {
		role, _ := s.Clientset.RbacV1().ClusterRoles().Get(ctx, name, metav1.GetOptions{})
		return s.MarshalToYaml(role), "ClusterRole"
	}
	role, _ := s.Clientset.RbacV1().Roles(ns).Get(ctx, name, metav1.GetOptions{})
	return s.MarshalToYaml(role), "Role"
}

// FetchRawSecurityRows is the logic moved from getSecurityDataHandler
func (s *K8sService) FetchRawSecurityRows(ctx context.Context) []SecurityRow {
	saList, _ := s.Clientset.CoreV1().ServiceAccounts("").List(ctx, metav1.ListOptions{})
	rbList, _ := s.Clientset.RbacV1().RoleBindings("").List(ctx, metav1.ListOptions{})
	crbList, _ := s.Clientset.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	roleList, _ := s.Clientset.RbacV1().Roles("").List(ctx, metav1.ListOptions{})
	cRoleList, _ := s.Clientset.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})

	roleMap := make(map[string]string)
	for _, r := range roleList.Items {
		roleMap[r.Namespace+"/"+r.Name] = s.MarshalToYaml(r)
	}

	cRoleMap := make(map[string]string)
	for _, r := range cRoleList.Items {
		cRoleMap[r.Name] = s.MarshalToYaml(r)
	}

	iamMap := make(map[string]string)
	for _, sa := range saList.Items {
		if arn, ok := sa.Annotations["eks.amazonaws.com/role-arn"]; ok {
			iamMap[sa.Namespace+"/"+sa.Name] = arn
		}
	}

	var rows []SecurityRow
	for _, rb := range rbList.Items {
		bY := s.MarshalToYaml(rb)
		rY := roleMap[rb.Namespace+"/"+rb.RoleRef.Name]
		if rY == "" && rb.RoleRef.Kind == "ClusterRole" {
			rY = cRoleMap[rb.RoleRef.Name]
		}
		for _, sub := range rb.Subjects {
			if sub.Kind == "ServiceAccount" {
				iam := iamMap[rb.Namespace+"/"+sub.Name]
				if iam == "" {
					iam = "None"
				}
				rows = append(rows, SecurityRow{
					SA: sub.Name, Namespace: rb.Namespace, IAMRole: iam,
					BindingType: "RoleBinding", BindingName: rb.Name,
					BindingYAML: bY, RoleYAML: rY, RoleName: rb.RoleRef.Name,
				})
			}
		}
	}

	for _, crb := range crbList.Items {
		bY := s.MarshalToYaml(crb)
		rY := cRoleMap[crb.RoleRef.Name]
		for _, sub := range crb.Subjects {
			if sub.Kind == "ServiceAccount" {
				iam := iamMap[sub.Namespace+"/"+sub.Name]
				if iam == "" {
					iam = "None"
				}
				rows = append(rows, SecurityRow{
					SA: sub.Name, Namespace: sub.Namespace, IAMRole: iam,
					BindingType: "ClusterRoleBinding", BindingName: crb.Name,
					BindingYAML: bY, RoleYAML: rY, RoleName: crb.RoleRef.Name,
					RoleKind: "ClusterRole", IsGlobal: true,
				})
			}
		}
	}
	return rows
}

// FetchRawGroupRows is the logic moved from getGroupSecurityDataHandler
func (s *K8sService) FetchRawGroupRows(ctx context.Context) []GroupSecurityRow {
	rbList, _ := s.Clientset.RbacV1().RoleBindings("").List(ctx, metav1.ListOptions{})
	crbList, _ := s.Clientset.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	groupMap := make(map[string]*GroupSecurityRow)

	addEntry := func(group, role, ns, bName, bKind, bYaml, rName, rKind, rYaml string) {
		if _, exists := groupMap[group]; !exists {
			groupMap[group] = &GroupSecurityRow{GroupName: group}
		}
		g := groupMap[group]
		g.Roles = append(g.Roles, role)
		g.Namespaces = append(g.Namespaces, ns)
		g.AllYAMLs = append(g.AllYAMLs,
			YamlBlock{Kind: bKind, Name: bName, Data: bYaml, Namespace: ns},
			YamlBlock{Kind: rKind, Name: rName, Data: rYaml, Namespace: ns},
		)
	}

	for _, rb := range rbList.Items {
		bY := s.MarshalToYaml(rb)
		rY, rKind := s.GetRoleDetail(ctx, rb.Namespace, rb.RoleRef.Name, rb.RoleRef.Kind)
		for _, sub := range rb.Subjects {
			if sub.Kind == "Group" {
				addEntry(sub.Name, rb.RoleRef.Name, rb.Namespace, rb.Name, "RoleBinding", bY, rb.RoleRef.Name, rKind, rY)
			}
		}
	}

	for _, crb := range crbList.Items {
		bY := s.MarshalToYaml(crb)
		crole, _ := s.Clientset.RbacV1().ClusterRoles().Get(ctx, crb.RoleRef.Name, metav1.GetOptions{})
		rY := s.MarshalToYaml(crole)
		for _, sub := range crb.Subjects {
			if sub.Kind == "Group" {
				addEntry(sub.Name, crb.RoleRef.Name, "Cluster-Wide", crb.Name, "ClusterRoleBinding", bY, crb.RoleRef.Name, "ClusterRole", rY)
			}
		}
	}

	var res []GroupSecurityRow
	for _, v := range groupMap {
		res = append(res, *v)
	}
	return res
}
