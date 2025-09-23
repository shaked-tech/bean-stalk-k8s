package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/bean-stalk-k8s/backend/k8s"
	"github.com/bean-stalk-k8s/backend/models"
)

// Handler contains the Kubernetes client and handler functions
type Handler struct {
	client *k8s.Client
}

// NewHandler creates a new Handler with a Kubernetes client
func NewHandler() (*Handler, error) {
	client, err := k8s.NewClient()
	if err != nil {
		return nil, err
	}

	return &Handler{
		client: client,
	}, nil
}

// GetNamespaces returns a list of all namespaces
func (h *Handler) GetNamespaces(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	namespaces, err := h.client.GetNamespaces(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	
	// Create response
	response := models.NamespaceList{
		Namespaces: namespaces,
	}

	// Write response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetPodMetrics returns metrics for all pods
func (h *Handler) GetPodMetrics(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Get namespace from query parameter
	namespace := r.URL.Query().Get("namespace")

	pods, err := h.client.GetPodMetrics(ctx, namespace)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Create response
	response := models.PodMetricsList{
		Pods: pods,
	}

	// Write response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Health returns a simple health check response
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	response := map[string]string{
		"status": "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	json.NewEncoder(w).Encode(response)
}

// EnableCORS is a middleware that sets CORS headers
func EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// If this is a preflight request, respond with 200 OK
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
