package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"github.com/bean-stalk-k8s/backend/k8s"
	"github.com/bean-stalk-k8s/backend/models"
)

// Handler contains only the Prometheus client for unified data access
type Handler struct {
	prometheusClient *k8s.PrometheusClient
}

// NewHandler creates a new Handler with only Prometheus client for unified data access
func NewHandler() (*Handler, error) {
	// Initialize Prometheus client
	prometheusURL := os.Getenv("PROMETHEUS_URL")
	if prometheusURL == "" {
		prometheusURL = "http://prometheus-stack-kube-prom-prometheus.pod-metrics-dashboard.svc.cluster.local:9090"
	}

	prometheusClient, err := k8s.NewPrometheusClient(prometheusURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus client: %w", err)
	}

	return &Handler{
		prometheusClient: prometheusClient,
	}, nil
}

// GetNamespaces returns a list of all namespaces from Prometheus
func (h *Handler) GetNamespaces(w http.ResponseWriter, r *http.Request) {
	if h.prometheusClient == nil {
		http.Error(w, "Service unavailable - Prometheus client not initialized", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	namespaces, err := h.prometheusClient.GetNamespaces(ctx)
	if err != nil {
		log.Printf("Error getting namespaces from Prometheus: %v", err)
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

// GetPodMetrics returns current metrics for all pods from Prometheus
func (h *Handler) GetPodMetrics(w http.ResponseWriter, r *http.Request) {
	if h.prometheusClient == nil {
		http.Error(w, "Service unavailable - Prometheus client not initialized", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	// Get namespace from query parameter
	namespace := r.URL.Query().Get("namespace")

	prometheusMetrics, err := h.prometheusClient.GetCurrentPodMetrics(ctx, namespace)
	if err != nil {
		log.Printf("Error getting pod metrics from Prometheus: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert Prometheus metrics to models format
	var pods []models.PodMetrics
	for _, metric := range prometheusMetrics {
		podMetric := convertPrometheusToModelMetric(metric)
		pods = append(pods, podMetric)
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

// GetHistoricalAnalysis returns 7-day historical analysis for pods
func (h *Handler) GetHistoricalAnalysis(w http.ResponseWriter, r *http.Request) {
	if h.prometheusClient == nil {
		http.Error(w, "Historical analysis not available - Prometheus client not initialized", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Get namespace from query parameter
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		namespace = ".*" // All namespaces
	}

	historicalData, err := h.prometheusClient.GetHistoricalMetrics(ctx, namespace)
	if err != nil {
		log.Printf("Error getting historical metrics: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Convert k8s types to models types
	var modelMetrics []models.HistoricalMetrics
	for _, hm := range historicalData {
		modelMetrics = append(modelMetrics, models.HistoricalMetrics{
			PodName:       hm.PodName,
			Namespace:     hm.Namespace,
			ContainerName: hm.ContainerName,
			CPU: models.HistoricalResourceData{
				Usage:    convertDataPoints(hm.CPU.Usage),
				Requests: convertDataPoints(hm.CPU.Requests),
				Limits:   convertDataPoints(hm.CPU.Limits),
				Average:  hm.CPU.Average,
				Peak:     hm.CPU.Peak,
				Minimum:  hm.CPU.Minimum,
				P95:      hm.CPU.P95,
				P99:      hm.CPU.P99,
				Trend:    hm.CPU.Trend,
			},
			Memory: models.HistoricalResourceData{
				Usage:    convertDataPoints(hm.Memory.Usage),
				Requests: convertDataPoints(hm.Memory.Requests),
				Limits:   convertDataPoints(hm.Memory.Limits),
				Average:  hm.Memory.Average,
				Peak:     hm.Memory.Peak,
				Minimum:  hm.Memory.Minimum,
				P95:      hm.Memory.P95,
				P99:      hm.Memory.P99,
				Trend:    hm.Memory.Trend,
			},
			Analysis: models.UsageAnalysis{
				CPUEfficiency:    hm.Analysis.CPUEfficiency,
				MemoryEfficiency: hm.Analysis.MemoryEfficiency,
				ResourceWaste: models.ResourceWasteAnalysis{
					CPUOverProvisioned:     hm.Analysis.ResourceWaste.CPUOverProvisioned,
					MemoryOverProvisioned:  hm.Analysis.ResourceWaste.MemoryOverProvisioned,
					CPUUnderProvisioned:    hm.Analysis.ResourceWaste.CPUUnderProvisioned,
					MemoryUnderProvisioned: hm.Analysis.ResourceWaste.MemoryUnderProvisioned,
					CPUWastePercentage:     hm.Analysis.ResourceWaste.CPUWastePercentage,
					MemoryWastePercentage:  hm.Analysis.ResourceWaste.MemoryWastePercentage,
				},
				Recommendations: hm.Analysis.Recommendations,
				Patterns: models.UsagePatterns{
					PeakHours:       hm.Analysis.Patterns.PeakHours,
					LowUsageHours:   hm.Analysis.Patterns.LowUsageHours,
					DailyVariation:  hm.Analysis.Patterns.DailyVariation,
					WeeklyVariation: hm.Analysis.Patterns.WeeklyVariation,
				},
			},
		})
	}

	// Create response
	response := models.HistoricalAnalysisList{
		HistoricalMetrics: modelMetrics,
		GeneratedAt:      time.Now(),
		TimeRange: models.TimeRange{
			Start: time.Now().Add(-7 * 24 * time.Hour),
			End:   time.Now(),
		},
		Summary: generateAnalysisSummary(modelMetrics),
	}

	// Write response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetPodTrends returns trend analysis for a specific pod
func (h *Handler) GetPodTrends(w http.ResponseWriter, r *http.Request) {
	if h.prometheusClient == nil {
		http.Error(w, "Trend analysis not available - Prometheus client not initialized", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	// Get parameters
	namespace := r.URL.Query().Get("namespace")
	podName := r.URL.Query().Get("pod")
	days := r.URL.Query().Get("days")
	
	if namespace == "" || podName == "" {
		http.Error(w, "namespace and pod parameters are required", http.StatusBadRequest)
		return
	}

	// Default to 7 days if not specified
	daysInt := 7
	if days != "" {
		if d, err := time.ParseDuration(days + "d"); err == nil {
			daysInt = int(d.Hours() / 24)
		}
	}

	// Get historical data for the specific pod
	historicalData, err := h.prometheusClient.GetHistoricalMetrics(ctx, namespace)
	if err != nil {
		log.Printf("Error getting pod trends: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert and filter for the specific pod
	var podTrends []models.HistoricalMetrics
	for _, hm := range historicalData {
		if hm.PodName == podName && hm.Namespace == namespace {
			// Convert to models type
			modelMetric := models.HistoricalMetrics{
				PodName:       hm.PodName,
				Namespace:     hm.Namespace,
				ContainerName: hm.ContainerName,
				CPU: models.HistoricalResourceData{
					Usage:    convertDataPoints(hm.CPU.Usage),
					Requests: convertDataPoints(hm.CPU.Requests),
					Limits:   convertDataPoints(hm.CPU.Limits),
					Average:  hm.CPU.Average,
					Peak:     hm.CPU.Peak,
					Minimum:  hm.CPU.Minimum,
					P95:      hm.CPU.P95,
					P99:      hm.CPU.P99,
					Trend:    hm.CPU.Trend,
				},
				Memory: models.HistoricalResourceData{
					Usage:    convertDataPoints(hm.Memory.Usage),
					Requests: convertDataPoints(hm.Memory.Requests),
					Limits:   convertDataPoints(hm.Memory.Limits),
					Average:  hm.Memory.Average,
					Peak:     hm.Memory.Peak,
					Minimum:  hm.Memory.Minimum,
					P95:      hm.Memory.P95,
					P99:      hm.Memory.P99,
					Trend:    hm.Memory.Trend,
				},
				Analysis: models.UsageAnalysis{
					CPUEfficiency:    hm.Analysis.CPUEfficiency,
					MemoryEfficiency: hm.Analysis.MemoryEfficiency,
					ResourceWaste: models.ResourceWasteAnalysis{
						CPUOverProvisioned:     hm.Analysis.ResourceWaste.CPUOverProvisioned,
						MemoryOverProvisioned:  hm.Analysis.ResourceWaste.MemoryOverProvisioned,
						CPUUnderProvisioned:    hm.Analysis.ResourceWaste.CPUUnderProvisioned,
						MemoryUnderProvisioned: hm.Analysis.ResourceWaste.MemoryUnderProvisioned,
						CPUWastePercentage:     hm.Analysis.ResourceWaste.CPUWastePercentage,
						MemoryWastePercentage:  hm.Analysis.ResourceWaste.MemoryWastePercentage,
					},
					Recommendations: hm.Analysis.Recommendations,
					Patterns: models.UsagePatterns{
						PeakHours:       hm.Analysis.Patterns.PeakHours,
						LowUsageHours:   hm.Analysis.Patterns.LowUsageHours,
						DailyVariation:  hm.Analysis.Patterns.DailyVariation,
						WeeklyVariation: hm.Analysis.Patterns.WeeklyVariation,
					},
				},
			}
			podTrends = append(podTrends, modelMetric)
		}
	}

	if len(podTrends) == 0 {
		http.Error(w, "No trend data found for the specified pod", http.StatusNotFound)
		return
	}

	// Generate summary
	summary := generatePodTrendSummary(podTrends)

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Create response
	response := models.PodTrendAnalysis{
		PodName:      podName,
		Namespace:    namespace,
		Containers:   podTrends,
		DaysAnalyzed: daysInt,
		GeneratedAt:  time.Now(),
		Summary:      summary,
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
	
	prometheusStatus := "unavailable"
	if h.prometheusClient != nil {
		prometheusStatus = "available"
	}
	
	response := map[string]interface{}{
		"status":           "healthy",
		"timestamp":        time.Now().Format(time.RFC3339),
		"prometheusClient": prometheusStatus,
		"features": map[string]bool{
			"realTimeMetrics":    true,
			"historicalAnalysis": h.prometheusClient != nil,
			"trendAnalysis":      h.prometheusClient != nil,
		},
	}
	
	json.NewEncoder(w).Encode(response)
}

// Helper function to convert k8s DataPoints to models DataPoints
func convertDataPoints(k8sPoints []k8s.DataPoint) []models.DataPoint {
	var modelPoints []models.DataPoint
	for _, point := range k8sPoints {
		modelPoints = append(modelPoints, models.DataPoint{
			Timestamp: point.Timestamp,
			Value:     point.Value,
		})
	}
	return modelPoints
}

// Helper function to convert Prometheus PodMetric to models PodMetrics
func convertPrometheusToModelMetric(metric k8s.PodMetric) models.PodMetrics {
	// Format values
	cpuUsageStr := formatCPU(metric.CPUUsage)
	cpuRequestStr := formatCPU(metric.CPURequest)
	cpuLimitStr := formatCPU(metric.CPULimit)
	
	memUsageStr := formatMemory(metric.MemoryUsage)
	memRequestStr := formatMemory(metric.MemoryRequest)
	memLimitStr := formatMemory(metric.MemoryLimit)
	
	// Calculate percentages
	var cpuRequestPercentage, cpuLimitPercentage float64
	var memRequestPercentage, memLimitPercentage float64
	
	if metric.CPURequest > 0 {
		cpuRequestPercentage = (metric.CPUUsage / metric.CPURequest) * 100
	}
	if metric.CPULimit > 0 {
		cpuLimitPercentage = (metric.CPUUsage / metric.CPULimit) * 100
	}
	if metric.MemoryRequest > 0 {
		memRequestPercentage = (metric.MemoryUsage / metric.MemoryRequest) * 100
	}
	if metric.MemoryLimit > 0 {
		memLimitPercentage = (metric.MemoryUsage / metric.MemoryLimit) * 100
	}
	
	return models.PodMetrics{
		Name:          metric.Name,
		Namespace:     metric.Namespace,
		ContainerName: metric.ContainerName,
		CPU: models.ResourceMetrics{
			Usage:             cpuUsageStr,
			Request:           cpuRequestStr,
			Limit:             cpuLimitStr,
			UsageValue:        metric.CPUUsage,
			RequestValue:      metric.CPURequest,
			LimitValue:        metric.CPULimit,
			RequestPercentage: cpuRequestPercentage,
			LimitPercentage:   cpuLimitPercentage,
		},
		Memory: models.ResourceMetrics{
			Usage:             memUsageStr,
			Request:           memRequestStr,
			Limit:             memLimitStr,
			UsageValue:        metric.MemoryUsage,
			RequestValue:      metric.MemoryRequest,
			LimitValue:        metric.MemoryLimit,
			RequestPercentage: memRequestPercentage,
			LimitPercentage:   memLimitPercentage,
		},
		Labels: metric.Labels,
	}
}

// Helper function to format CPU values (cores to millicores)
func formatCPU(cpuCores float64) string {
	if cpuCores == 0 {
		return "0m"
	}
	// Convert cores to millicores
	millicores := cpuCores * 1000
	if millicores < 1 {
		return fmt.Sprintf("%.1fm", millicores)
	}
	return fmt.Sprintf("%.0fm", millicores)
}

// Helper function to format memory values (bytes to human readable)
func formatMemory(bytes float64) string {
	// DEBUG: Log memory conversion
	log.Printf("DEBUG: formatMemory input: %.0f bytes", bytes)
	
	if bytes == 0 {
		return "0Mi"
	}
	
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	
	var result string
	if bytes >= GB {
		result = fmt.Sprintf("%.1fGi", bytes/GB)
	} else if bytes >= MB {
		result = fmt.Sprintf("%.0fMi", bytes/MB)
	} else if bytes >= KB {
		result = fmt.Sprintf("%.0fKi", bytes/KB)
	} else {
		result = fmt.Sprintf("%.0fB", bytes)
	}
	
	// DEBUG: Log conversion result
	log.Printf("DEBUG: formatMemory output: %s (%.2f Mi)", result, bytes/MB)
	
	return result
}

// Helper function to generate analysis summary
func generateAnalysisSummary(metrics []models.HistoricalMetrics) models.AnalysisSummary {
	if len(metrics) == 0 {
		return models.AnalysisSummary{}
	}

	var totalEfficiency float64
	var overProvisioned, underProvisioned, wellOptimized int
	var totalRecommendations int
	recommendationCount := make(map[string]int)

	for _, metric := range metrics {
		// Count efficiency
		avgEfficiency := (metric.Analysis.CPUEfficiency + metric.Analysis.MemoryEfficiency) / 2
		totalEfficiency += avgEfficiency

		// Categorize based on resource waste analysis
		if metric.Analysis.ResourceWaste.CPUOverProvisioned || metric.Analysis.ResourceWaste.MemoryOverProvisioned {
			overProvisioned++
		} else if metric.Analysis.ResourceWaste.CPUUnderProvisioned || metric.Analysis.ResourceWaste.MemoryUnderProvisioned {
			underProvisioned++
		} else {
			wellOptimized++
		}

		// Count recommendations
		totalRecommendations += len(metric.Analysis.Recommendations)
		for _, rec := range metric.Analysis.Recommendations {
			recommendationCount[rec]++
		}
	}

	// Find most common recommendation
	var mostCommon string
	var maxCount int
	for rec, count := range recommendationCount {
		if count > maxCount {
			maxCount = count
			mostCommon = rec
		}
	}

	return models.AnalysisSummary{
		TotalPodsAnalyzed:        len(metrics),
		OverProvisionedPods:      overProvisioned,
		UnderProvisionedPods:     underProvisioned,
		WellOptimizedPods:        wellOptimized,
		AverageEfficiency:        totalEfficiency / float64(len(metrics)),
		TotalRecommendations:     totalRecommendations,
		MostCommonRecommendation: mostCommon,
	}
}

// Helper function to generate pod trend summary
func generatePodTrendSummary(containers []models.HistoricalMetrics) models.PodTrendSummary {
	if len(containers) == 0 {
		return models.PodTrendSummary{
			OverallTrend: "unknown",
			RiskLevel:    "unknown",
		}
	}

	// Analyze trends across all containers
	var increasingCount, decreasingCount, stableCount int
	var allRecommendations []string
	var highEfficiencyCount, lowEfficiencyCount int

	for _, container := range containers {
		// Count trend types
		switch container.CPU.Trend {
		case "increasing":
			increasingCount++
		case "decreasing":
			decreasingCount++
		case "stable":
			stableCount++
		}

		// Collect recommendations
		allRecommendations = append(allRecommendations, container.Analysis.Recommendations...)

		// Check efficiency levels
		avgEff := (container.Analysis.CPUEfficiency + container.Analysis.MemoryEfficiency) / 2
		if avgEff > 70 {
			highEfficiencyCount++
		} else if avgEff < 30 {
			lowEfficiencyCount++
		}
	}

	// Determine overall trend
	var overallTrend string
	totalContainers := len(containers)
	if increasingCount > totalContainers/2 {
		overallTrend = "increasing"
	} else if decreasingCount > totalContainers/2 {
		overallTrend = "decreasing"
	} else {
		overallTrend = "stable"
	}

	// Determine risk level
	var riskLevel string
	if lowEfficiencyCount > totalContainers/2 || increasingCount > totalContainers/2 {
		riskLevel = "high"
	} else if lowEfficiencyCount > 0 || increasingCount > 0 {
		riskLevel = "medium"
	} else {
		riskLevel = "low"
	}

	// Remove duplicate recommendations
	uniqueRecommendations := make(map[string]bool)
	var finalRecommendations []string
	for _, rec := range allRecommendations {
		if !uniqueRecommendations[rec] {
			uniqueRecommendations[rec] = true
			finalRecommendations = append(finalRecommendations, rec)
		}
	}

	// Calculate next review date based on risk level
	var nextReview time.Time
	switch riskLevel {
	case "high":
		nextReview = time.Now().Add(3 * 24 * time.Hour) // 3 days
	case "medium":
		nextReview = time.Now().Add(7 * 24 * time.Hour) // 1 week
	default:
		nextReview = time.Now().Add(30 * 24 * time.Hour) // 1 month
	}

	return models.PodTrendSummary{
		OverallTrend:            overallTrend,
		ResourceRecommendations: finalRecommendations,
		RiskLevel:               riskLevel,
		NextReviewDate:          nextReview,
	}
}

// GetPodSummary returns summary statistics including low and high usage pods
func (h *Handler) GetPodSummary(w http.ResponseWriter, r *http.Request) {
	if h.prometheusClient == nil {
		http.Error(w, "Service unavailable - Prometheus client not initialized", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	// Get namespace from query parameter
	namespace := r.URL.Query().Get("namespace")

	prometheusMetrics, err := h.prometheusClient.GetCurrentPodMetrics(ctx, namespace)
	if err != nil {
		log.Printf("Error getting pod metrics from Prometheus: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert Prometheus metrics to models format
	var pods []models.PodMetrics
	for _, metric := range prometheusMetrics {
		podMetric := convertPrometheusToModelMetric(metric)
		pods = append(pods, podMetric)
	}

	// Calculate summary statistics
	totalPods := len(pods)
	var totalCPUUsage, totalMemoryUsage float64
	var highCPUPods, highMemoryPods int
	var lowCPUPods, lowMemoryPods int

	for _, pod := range pods {
		// Add to totals for averages
		totalCPUUsage += pod.CPU.RequestPercentage
		totalMemoryUsage += pod.Memory.RequestPercentage

		// Count high usage pods (>80%)
		if pod.CPU.RequestPercentage > 80 {
			highCPUPods++
		}
		if pod.Memory.RequestPercentage > 80 {
			highMemoryPods++
		}

		// Count low usage pods (<40%)
		if pod.CPU.RequestPercentage < 40 && pod.CPU.RequestPercentage > 0 {
			lowCPUPods++
		}
		if pod.Memory.RequestPercentage < 40 && pod.Memory.RequestPercentage > 0 {
			lowMemoryPods++
		}
	}

	// Calculate averages
	var averageCPUUsage, averageMemoryUsage float64
	if totalPods > 0 {
		averageCPUUsage = totalCPUUsage / float64(totalPods)
		averageMemoryUsage = totalMemoryUsage / float64(totalPods)
	}

	// Create response
	response := models.PodSummaryResponse{
		TotalPods:          totalPods,
		AverageCPUUsage:    averageCPUUsage,
		AverageMemoryUsage: averageMemoryUsage,
		HighCPUPods:        highCPUPods,
		HighMemoryPods:     highMemoryPods,
		LowCPUPods:         lowCPUPods,
		LowMemoryPods:      lowMemoryPods,
		GeneratedAt:        time.Now(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Write response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
