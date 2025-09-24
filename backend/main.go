package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/bean-stalk-k8s/backend/handlers"
)

func main() {
	// Create a new handler
	handler, err := handlers.NewHandler()
	if err != nil {
		log.Fatalf("Failed to create handler: %v", err)
	}

	// Create a new router
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("/api/namespaces", handler.GetNamespaces)
	mux.HandleFunc("/api/pods", handler.GetPodMetrics)
	mux.HandleFunc("/api/pods/analysis", handler.GetHistoricalAnalysis)
	mux.HandleFunc("/api/pods/trends", handler.GetPodTrends)

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: handlers.EnableCORS(mux),
	}

	// Start server
	log.Printf("Starting server on port %s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
