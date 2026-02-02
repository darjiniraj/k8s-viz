// package main

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"net/http"

// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/client-go/kubernetes"
// 	"k8s.io/client-go/tools/clientcmd"
// 	"sigs.k8s.io/yaml"
// )

// type SecurityRow struct {
// 	SA          string `json:"sa"`
// 	Namespace   string `json:"namespace"`
// 	IAMRole     string `json:"iam_role"`
// 	BindingType string `json:"binding_type"`
// 	BindingName string `json:"binding_name"`
// 	BindingYAML string `json:"binding_yaml"`
// 	RoleYAML    string `json:"role_yaml"`
// 	RoleName    string `json:"role"`
// }

// type YamlBlock struct {
// 	Kind      string `json:"kind"`
// 	Name      string `json:"name"`
// 	Data      string `json:"data"`
// 	Namespace string `json:"namespace"`
// }

// type GroupSecurityRow struct {
// 	GroupName  string      `json:"group_name"`
// 	Roles      []string    `json:"roles"`
// 	Namespaces []string    `json:"namespaces"`
// 	AllYAMLs   []YamlBlock `json:"all_yamls"` // Now a slice of structs
// }

// func main() {
// 	http.HandleFunc("/api/table", getSecurityData)
// 	http.HandleFunc("/api/groups", getGroupSecurityData)
// 	http.Handle("/", http.FileServer(http.Dir(".")))
// 	fmt.Println("🚀 High-Speed Server: http://localhost:8080")
// 	log.Fatal(http.ListenAndServe(":8080", nil))
// }

// func getSecurityData(w http.ResponseWriter, r *http.Request) {
// 	rules := clientcmd.NewDefaultClientConfigLoadingRules()
// 	config, _ := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{}).ClientConfig()
// 	clientset, err := kubernetes.NewForConfig(config)
// 	if err != nil {
// 		log.Printf("Config Error: %v", err)
// 		return
// 	}

// 	// 1. BULK FETCH (Get everything in 5 calls total)
// 	ctx := context.TODO()
// 	saList, _ := clientset.CoreV1().ServiceAccounts("").List(ctx, metav1.ListOptions{})
// 	rbList, _ := clientset.RbacV1().RoleBindings("").List(ctx, metav1.ListOptions{})
// 	crbList, _ := clientset.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
// 	roleList, _ := clientset.RbacV1().Roles("").List(ctx, metav1.ListOptions{})
// 	cRoleList, _ := clientset.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})

// 	// 2. BUILD LOOKUP MAPS (Cache)
// 	roleMap := make(map[string]string)
// 	for _, r := range roleList.Items {
// 		y, _ := yaml.Marshal(r)
// 		roleMap[r.Namespace+"/"+r.Name] = string(y)
// 	}

// 	cRoleMap := make(map[string]string)
// 	for _, r := range cRoleList.Items {
// 		y, _ := yaml.Marshal(r)
// 		cRoleMap[r.Name] = string(y)
// 	}

// 	iamMap := make(map[string]string)
// 	for _, sa := range saList.Items {
// 		if arn, ok := sa.Annotations["eks.amazonaws.com/role-arn"]; ok {
// 			iamMap[sa.Namespace+"/"+sa.Name] = arn
// 		}
// 	}

// 	var rows []SecurityRow

// 	// 3. PROCESS ROLE BINDINGS
// 	for _, rb := range rbList.Items {
// 		bY, _ := yaml.Marshal(rb)
// 		// Attempt to get the Role YAML from our cached map
// 		rY := roleMap[rb.Namespace+"/"+rb.RoleRef.Name]

// 		// Fallback: If it's a ClusterRole being used in a RoleBinding
// 		if rY == "" && rb.RoleRef.Kind == "ClusterRole" {
// 			rY = cRoleMap[rb.RoleRef.Name]
// 		}

// 		for _, sub := range rb.Subjects {
// 			if sub.Kind == "ServiceAccount" {
// 				iam := iamMap[rb.Namespace+"/"+sub.Name]
// 				if iam == "" {
// 					iam = "None"
// 				}

// 				rows = append(rows, SecurityRow{
// 					SA:          sub.Name,
// 					Namespace:   rb.Namespace,
// 					IAMRole:     iam,
// 					BindingType: "RoleBinding",
// 					BindingName: rb.Name, // Populating the binding name
// 					BindingYAML: string(bY),
// 					RoleYAML:    rY,
// 					RoleName:    rb.RoleRef.Name,
// 				})
// 			}
// 		}
// 	}

// 	// 4. PROCESS CLUSTER ROLE BINDINGS
// 	// 4. PROCESS CLUSTER ROLE BINDINGS
// 	for _, crb := range crbList.Items {
// 		bY, _ := yaml.Marshal(crb)
// 		rY := cRoleMap[crb.RoleRef.Name]
// 		for _, sub := range crb.Subjects {
// 			if sub.Kind == "ServiceAccount" {
// 				// Note: Use the subject's own namespace to look up IAM
// 				iam := iamMap[sub.Namespace+"/"+sub.Name]
// 				if iam == "" {
// 					iam = "None"
// 				}

// 				rows = append(rows, SecurityRow{
// 					SA:          sub.Name,
// 					Namespace:   sub.Namespace + " (Global)",
// 					IAMRole:     iam,
// 					BindingType: "ClusterRoleBinding",
// 					BindingName: crb.Name, // Populating the binding name
// 					BindingYAML: string(bY),
// 					RoleYAML:    rY,
// 					RoleName:    crb.RoleRef.Name,
// 				})
// 			}
// 		}
// 	}

// 	log.Printf("✅ Processed %d security paths", len(rows))
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(rows)
// }

// func getGroupSecurityData(w http.ResponseWriter, r *http.Request) {
// 	clientset := getK8sClient()
// 	ctx := context.TODO()

// 	rbList, _ := clientset.RbacV1().RoleBindings("").List(ctx, metav1.ListOptions{})
// 	crbList, _ := clientset.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})

// 	groupMap := make(map[string]*GroupSecurityRow)
// 	addEntry := func(group, role, ns, bName, bKind, bYaml, rName, rKind, rYaml string) {
// 		if _, exists := groupMap[group]; !exists {
// 			groupMap[group] = &GroupSecurityRow{GroupName: group}
// 		}
// 		g := groupMap[group]
// 		g.Roles = append(g.Roles, role)
// 		g.Namespaces = append(g.Namespaces, ns)

// 		// Add the Binding and the Role as distinct blocks with metadata
// 		// NEW: We now pass the 'ns' variable into the Namespace field
// 		g.AllYAMLs = append(g.AllYAMLs,
// 			YamlBlock{
// 				Kind:      bKind,
// 				Name:      bName,
// 				Data:      bYaml,
// 				Namespace: ns, // <--- Add this
// 			},
// 			YamlBlock{
// 				Kind:      rKind,
// 				Name:      rName,
// 				Data:      rYaml,
// 				Namespace: ns, // <--- Add this
// 			},
// 		)
// 	}

// 	// for _, rb := range rbList.Items {
// 	// 	bY, _ := yaml.Marshal(rb)
// 	// 	role, _ := clientset.RbacV1().Roles(rb.Namespace).Get(ctx, rb.RoleRef.Name, metav1.GetOptions{})
// 	// 	rY, _ := yaml.Marshal(role)
// 	// 	for _, sub := range rb.Subjects {
// 	// 		if sub.Kind == "Group" {
// 	// 			addEntry(sub.Name, rb.RoleRef.Name, rb.Namespace,
// 	// 				rb.Name, "RoleBinding", string(bY),
// 	// 				rb.RoleRef.Name, "Role", string(rY))
// 	// 		}
// 	// 	}
// 	// }
// 	for _, rb := range rbList.Items {
// 		bY, _ := yaml.Marshal(rb)
// 		var rY []byte
// 		var rKind string

// 		// CHECK: Is this binding pointing to a Role or a ClusterRole?
// 		if rb.RoleRef.Kind == "ClusterRole" {
// 			role, _ := clientset.RbacV1().ClusterRoles().Get(ctx, rb.RoleRef.Name, metav1.GetOptions{})
// 			rY, _ = yaml.Marshal(role)
// 			rKind = "ClusterRole"
// 		} else {
// 			role, _ := clientset.RbacV1().Roles(rb.Namespace).Get(ctx, rb.RoleRef.Name, metav1.GetOptions{})
// 			rY, _ = yaml.Marshal(role)
// 			rKind = "Role"
// 		}

// 		for _, sub := range rb.Subjects {
// 			if sub.Kind == "Group" {
// 				addEntry(sub.Name, rb.RoleRef.Name, rb.Namespace,
// 					rb.Name, "RoleBinding", string(bY),
// 					rb.RoleRef.Name, rKind, string(rY))
// 			}
// 		}
// 	}

// 	for _, crb := range crbList.Items {
// 		bY, _ := yaml.Marshal(crb)
// 		// Fetch the ClusterRole
// 		crole, _ := clientset.RbacV1().ClusterRoles().Get(ctx, crb.RoleRef.Name, metav1.GetOptions{})
// 		rY, _ := yaml.Marshal(crole)

// 		for _, sub := range crb.Subjects {
// 			if sub.Kind == "Group" {
// 				// Updated to match the new addEntry(group, role, ns, bName, bKind, bYaml, rName, rKind, rYaml)
// 				addEntry(
// 					sub.Name,             // Group Name
// 					crb.RoleRef.Name,     // Role Name
// 					"Cluster-Wide",       // Namespace
// 					crb.Name,             // Binding Name
// 					"ClusterRoleBinding", // Binding Kind
// 					string(bY),           // Binding YAML
// 					crb.RoleRef.Name,     // Role Name
// 					"ClusterRole",        // Role Kind
// 					string(rY),           // Role YAML
// 				)
// 			}
// 		}
// 	}

// 	var res []GroupSecurityRow
// 	for _, v := range groupMap {
// 		res = append(res, *v)
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(res)
// }

// func getK8sClient() *kubernetes.Clientset {
// 	rules := clientcmd.NewDefaultClientConfigLoadingRules()
// 	config, _ := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{}).ClientConfig()
// 	c, _ := kubernetes.NewForConfig(config)
// 	return c
// }

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	clientset, err := initK8sClient()
	if err != nil {
		log.Fatalf("Fatal: Could not connect to Kubernetes: %v", err)
	}

	service := &K8sService{Clientset: clientset}

	// Routes
	http.HandleFunc("/api/table", getSecurityDataHandler(service))
	http.HandleFunc("/api/groups", getGroupSecurityDataHandler(service))
	http.Handle("/", http.FileServer(http.Dir(".")))

	fmt.Println("🚀 K8s-GUARD Server: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// getSecurityDataHandler processes the Apps/Service Account view.
func getSecurityDataHandler(s *K8sService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// 1. Bulk Fetch (Optimized for speed)
		saList, _ := s.Clientset.CoreV1().ServiceAccounts("").List(ctx, metav1.ListOptions{})
		rbList, _ := s.Clientset.RbacV1().RoleBindings("").List(ctx, metav1.ListOptions{})
		crbList, _ := s.Clientset.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
		roleList, _ := s.Clientset.RbacV1().Roles("").List(ctx, metav1.ListOptions{})
		cRoleList, _ := s.Clientset.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})

		// 2. Build Lookup Maps
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

		// 3. Process RoleBindings
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

		// 4. Process ClusterRoleBindings
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
						SA: sub.Name, Namespace: sub.Namespace + " (Global)", IAMRole: iam,
						BindingType: "ClusterRoleBinding", BindingName: crb.Name,
						BindingYAML: bY, RoleYAML: rY, RoleName: crb.RoleRef.Name,
					})
				}
			}
		}

		respondWithJSON(w, rows)
	}
}

// getGroupSecurityDataHandler processes the User Groups view.
func getGroupSecurityDataHandler(s *K8sService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
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

		// Process RoleBindings for Groups
		for _, rb := range rbList.Items {
			bY := s.MarshalToYaml(rb)
			rY, rKind := s.GetRoleDetail(ctx, rb.Namespace, rb.RoleRef.Name, rb.RoleRef.Kind)

			for _, sub := range rb.Subjects {
				if sub.Kind == "Group" {
					addEntry(sub.Name, rb.RoleRef.Name, rb.Namespace, rb.Name, "RoleBinding", bY, rb.RoleRef.Name, rKind, rY)
				}
			}
		}

		// Process ClusterRoleBindings for Groups
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
		respondWithJSON(w, res)
	}
}

// Helper: Setup K8s Connection
func initK8sClient() (*kubernetes.Clientset, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

// Helper: Standardized JSON Response
func respondWithJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}
