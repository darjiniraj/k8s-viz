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

	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eks"
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

	// Initialize AWS Config
	awsCfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1")) // Change to your region
	if err != nil {
		log.Printf("Warning: Could not load AWS config (EKS Access Entries will be empty): %v", err)
	}
	eksClient := eks.NewFromConfig(awsCfg)

	// Update your service struct initialization
	service := &K8sService{
		Clientset:   clientset,
		EKSClient:   eksClient,
		ClusterName: "your-cluster-name",
	}

	mux := http.NewServeMux()

	// --- ROUTES ---
	mux.HandleFunc("/api/table", getSecurityDataHandler(service))
	mux.HandleFunc("/api/groups", getGroupSecurityDataHandler(service))

	// NEW CILIUM ROUTE: Pass the dynClient here
	mux.HandleFunc("/api/cilium", getCiliumPoliciesHandler(dynClient))
	mux.HandleFunc("/api/iam-audit", getIAMAuditHandler(service))
	// --- FRONTEND ---
	distFiles, err := fs.Sub(uiAssets, "ui/dist")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("/", http.FileServer(http.FS(distFiles)))

	fmt.Println("🚀 K8s-GUARD running at: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func getIAMAuditHandler(s *K8sService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cacheKey := "iam_audit_data"

		// 1. Cache Check
		if cachedData, found := appCache.Get(cacheKey); found {
			respondWithJSON(w, cachedData)
			return
		}

		// 2. Use our refactored Service methods to get the raw data
		saData := s.FetchRawSecurityRows(r.Context())
		groupData := s.FetchRawGroupRows(r.Context())
		// 3. Get Access Entry Mappings (Real or Mock)
		accessEntries := s.FetchEKSAccessEntries(r.Context())

		// 3. Use the IAM package to aggregate it
		auditData := Aggregate(saData, groupData, accessEntries)

		// 4. Cache and Respond
		appCache.Set(cacheKey, auditData, 10*time.Minute)
		respondWithJSON(w, auditData)
	}
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
		cacheKey := "sa_data"
		if cachedData, found := appCache.Get(cacheKey); found && r.URL.Query().Get("refresh") != "true" {
			w.Header().Set("X-Cache", "HIT")
			respondWithJSON(w, cachedData)
			return
		}

		// CALL THE NEW SERVICE METHOD
		rows := s.FetchRawSecurityRows(r.Context())

		appCache.Set(cacheKey, rows, 10*time.Minute)
		w.Header().Set("X-Cache", "MISS")
		respondWithJSON(w, rows)
	}
}

// getGroupSecurityDataHandler processes the User Groups view.
func getGroupSecurityDataHandler(s *K8sService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cacheKey := "group_data"
		if cachedData, found := appCache.Get(cacheKey); found && r.URL.Query().Get("refresh") != "true" {
			w.Header().Set("X-Cache", "HIT")
			respondWithJSON(w, cachedData)
			return
		}

		// CALL THE NEW SERVICE METHOD
		res := s.FetchRawGroupRows(r.Context())

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
