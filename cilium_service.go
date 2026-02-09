package main

import (
	"context"
	"fmt"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
)

// GVR definitions for Cilium
var (
	cnpGVR  = schema.GroupVersionResource{Group: "cilium.io", Version: "v2", Resource: "ciliumnetworkpolicies"}
	ccnpGVR = schema.GroupVersionResource{Group: "cilium.io", Version: "v2", Resource: "ciliumclusterwidenetworkpolicies"}
)

func GetCiliumPolicies(dynClient dynamic.Interface) ([]CiliumPolicyRow, error) {
	var policies []CiliumPolicyRow

	// 1. Fetch Namespaced Policies (CNP)
	cnps, err := dynClient.Resource(cnpGVR).List(context.TODO(), metav1.ListOptions{})
	if err == nil {
		for _, item := range cnps.Items {
			policies = append(policies, parseCiliumObject(item, false))
		}
	}

	// 2. Fetch Clusterwide Policies (CCNP)
	ccnps, err := dynClient.Resource(ccnpGVR).List(context.TODO(), metav1.ListOptions{})
	if err == nil {
		for _, item := range ccnps.Items {
			policies = append(policies, parseCiliumObject(item, true))
		}
	}

	// Sort by name for a stable UI
	sort.Slice(policies, func(i, j int) bool {
		return policies[i].Name < policies[j].Name
	})

	return policies, nil
}

func parseCiliumObject(obj unstructured.Unstructured, isGlobal bool) CiliumPolicyRow {
	yamlData, _ := yaml.Marshal(obj.Object)
	
	spec, ok := obj.Object["spec"].(map[string]interface{})
	selectorText := "All Endpoints"
	policyType := "L3/L4"

	if ok {
		// Parse Selector
		if sel, ok := spec["endpointSelector"].(map[string]interface{}); ok {
			if labels, ok := sel["matchLabels"].(map[string]interface{}); ok {
				selectorText = ""
				for k, v := range labels {
					selectorText += fmt.Sprintf("%s=%v ", k, v)
				}
			}
		}

		// Parse Traffic Type
		_, hasIngress := spec["ingress"]
		_, hasEgress := spec["egress"]
		if hasIngress && hasEgress {
			policyType = "Ingress + Egress"
		} else if hasIngress {
			policyType = "Ingress Only"
		} else if hasEgress {
			policyType = "Egress Only"
		}
	}

	return CiliumPolicyRow{
		Name:           obj.GetName(),
		Namespace:      obj.GetNamespace(),
		Kind:           obj.GetKind(),
		IsClusterWide:  isGlobal,
		TargetSelector: selectorText,
		Type:           policyType,
		Yaml:           string(yamlData),
	}
}