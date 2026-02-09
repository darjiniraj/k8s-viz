package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"k8s-viz/internal/cache"
	"log"
	"net/http"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// This tells Go to look into the ui/dist folder created by Vite
//
//go:embed ui/dist/*
var uiAssets embed.FS
var appCache = cache.New()

func main() {
	// 1. Initialize BOTH clients
	clientset, dynClient, err := initK8sClients()
	if err != nil {
		log.Fatalf("Fatal: Could not connect to Kubernetes: %v", err)
	}

	service := &K8sService{Clientset: clientset}

	mux := http.NewServeMux()

	// --- ROUTES ---
	mux.HandleFunc("/api/table", getSecurityDataHandler(service))
	mux.HandleFunc("/api/groups", getGroupSecurityDataHandler(service))

	// NEW CILIUM ROUTE: Pass the dynClient here
	mux.HandleFunc("/api/cilium", getCiliumPoliciesHandler(dynClient))

	// --- FRONTEND ---
	distFiles, err := fs.Sub(uiAssets, "ui/dist")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("/", http.FileServer(http.FS(distFiles)))

	fmt.Println("🚀 K8s-GUARD running at: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func getCiliumPoliciesHandler(dynClient dynamic.Interface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cacheKey := "cilium_data"
		forceRefresh := r.URL.Query().Get("refresh") == "true"

		// 1. Cache logic
		if !forceRefresh {
			if cachedData, found := appCache.Get(cacheKey); found {
				respondWithJSON(w, cachedData)
				return
			}
		}

		// 2. Call your cilium_service.go logic
		policies, err := GetCiliumPolicies(dynClient)
		if err != nil {
			log.Printf("Cilium Error: %v", err)
			respondWithJSON(w, []CiliumPolicyRow{}) // Return empty list if error
			return
		}

		// 3. Cache and respond
		appCache.Set(cacheKey, policies, 10*time.Minute)
		respondWithJSON(w, policies)
	}
}

// getSecurityDataHandler processes the Apps/Service Account view.
func getSecurityDataHandler(s *K8sService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Cache Check
		cacheKey := "sa_data"
		forceRefresh := r.URL.Query().Get("refresh") == "true"

		if !forceRefresh {
			if cachedData, found := appCache.Get(cacheKey); found {
				w.Header().Set("X-Cache", "HIT")
				respondWithJSON(w, cachedData)
				return
			}
		}

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

					// FIX: Use the raw sub.Namespace so the UI filter works.
					// We add a separate 'IsGlobal' flag for the UI to show the badge.
					rows = append(rows, SecurityRow{
						SA:          sub.Name,
						Namespace:   sub.Namespace, // No more " (Global)" suffix here
						IAMRole:     iam,
						BindingType: "ClusterRoleBinding",
						BindingName: crb.Name,
						BindingYAML: bY,
						RoleYAML:    rY,
						RoleName:    crb.RoleRef.Name,
						RoleKind:    "ClusterRole", // Add this field to your SecurityRow struct
						IsGlobal:    true,          // Add this field to your SecurityRow struct
					})
				}
			}
		}
		// 2. Before responding, Save to Cache (e.g., 10 minutes)
		appCache.Set(cacheKey, rows, 10*time.Minute)

		w.Header().Set("X-Cache", "MISS")
		respondWithJSON(w, rows)
	}
}

// getGroupSecurityDataHandler processes the User Groups view.
func getGroupSecurityDataHandler(s *K8sService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Cache Check
		cacheKey := "group_data"
		forceRefresh := r.URL.Query().Get("refresh") == "true"

		if !forceRefresh {
			if cachedData, found := appCache.Get(cacheKey); found {
				w.Header().Set("X-Cache", "HIT")
				respondWithJSON(w, cachedData)
				return
			}
		}
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
		// 2. Store result in cache
		appCache.Set(cacheKey, res, 10*time.Minute)

		w.Header().Set("X-Cache", "MISS")
		respondWithJSON(w, res)
	}
}

func initK8sClients() (*kubernetes.Clientset, dynamic.Interface, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, nil, err
	}

	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	dc, err := dynamic.NewForConfig(config) // Create the dynamic client
	if err != nil {
		return nil, nil, err
	}

	return cs, dc, nil
}

// Helper: Standardized JSON Response
func respondWithJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}
