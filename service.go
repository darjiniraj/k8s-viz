package main

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

type K8sService struct {
	Clientset *kubernetes.Clientset
}

// MarshalToYaml converts K8s objects to string YAML safely.
func (s *K8sService) MarshalToYaml(obj interface{}) string {
	y, err := yaml.Marshal(obj)
	if err != nil {
		return ""
	}
	return string(y)
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
