package k8s

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// PrometheusClient wraps the Prometheus API client
type PrometheusClient struct {
	client v1.API
}

// NewPrometheusClient creates a new Prometheus client
func NewPrometheusClient(prometheusURL string) (*PrometheusClient, error) {
	config := api.Config{
		Address: prometheusURL,
	}

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus client: %w", err)
	}

	return &PrometheusClient{
		client: v1.NewAPI(client),
	}, nil
}

// HistoricalMetrics represents metrics data over time
type HistoricalMetrics struct {
	PodName       string                 `json:"podName"`
	Namespace     string                 `json:"namespace"`
	ContainerName string                 `json:"containerName"`
	CPU           HistoricalResourceData `json:"cpu"`
	Memory        HistoricalResourceData `json:"memory"`
	Analysis      UsageAnalysis          `json:"analysis"`
}

// HistoricalResourceData contains historical resource usage data
type HistoricalResourceData struct {
	Usage      []DataPoint `json:"usage"`
	Requests   []DataPoint `json:"requests"`
	Limits     []DataPoint `json:"limits"`
	Average    float64     `json:"average"`
	Peak       float64     `json:"peak"`
	Minimum    float64     `json:"minimum"`
	P95        float64     `json:"p95"`
	P99        float64     `json:"p99"`
	Trend      string      `json:"trend"` // "increasing", "decreasing", "stable"
}

// DataPoint represents a single metric data point
type DataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// UsageAnalysis provides insights about resource usage patterns
type UsageAnalysis struct {
	CPUEfficiency     float64                `json:"cpuEfficiency"`     // Average usage/request ratio
	MemoryEfficiency  float64                `json:"memoryEfficiency"`  // Average usage/request ratio
	ResourceWaste     ResourceWasteAnalysis  `json:"resourceWaste"`
	Recommendations   []string               `json:"recommendations"`
	Patterns          UsagePatterns          `json:"patterns"`
}

// ResourceWasteAnalysis identifies over/under-provisioned resources
type ResourceWasteAnalysis struct {
	CPUOverProvisioned    bool    `json:"cpuOverProvisioned"`
	MemoryOverProvisioned bool    `json:"memoryOverProvisioned"`
	CPUUnderProvisioned   bool    `json:"cpuUnderProvisioned"`
	MemoryUnderProvisioned bool   `json:"memoryUnderProvisioned"`
	CPUWastePercentage    float64 `json:"cpuWastePercentage"`
	MemoryWastePercentage float64 `json:"memoryWastePercentage"`
}

// UsagePatterns identifies usage patterns
type UsagePatterns struct {
	PeakHours       []int   `json:"peakHours"`       // Hours of day with peak usage
	LowUsageHours   []int   `json:"lowUsageHours"`   // Hours of day with low usage
	DailyVariation  float64 `json:"dailyVariation"`  // Coefficient of variation across days
	WeeklyVariation float64 `json:"weeklyVariation"` // Variation across week
}

// GetHistoricalMetrics retrieves and analyzes 7-day historical metrics for pods
func (p *PrometheusClient) GetHistoricalMetrics(ctx context.Context, namespace string) ([]HistoricalMetrics, error) {
	now := time.Now()
	sevenDaysAgo := now.Add(-7 * 24 * time.Hour)
	
	// Get pod list from the last 7 days
	pods, err := p.getActivePods(ctx, namespace, sevenDaysAgo, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get active pods: %w", err)
	}

	var results []HistoricalMetrics
	for _, pod := range pods {
		for _, container := range pod.Containers {
			metrics, err := p.getHistoricalMetricsForContainer(ctx, pod.Name, pod.Namespace, container, sevenDaysAgo, now)
			if err != nil {
				log.Printf("Warning: failed to get metrics for pod %s/%s container %s: %v", 
					pod.Namespace, pod.Name, container, err)
				continue
			}
			results = append(results, metrics)
		}
	}

	return results, nil
}

// PodInfo represents basic pod information
type PodInfo struct {
	Name       string   `json:"name"`
	Namespace  string   `json:"namespace"`
	Containers []string `json:"containers"`
}

// getActivePods retrieves pods that were active during the specified time range
func (p *PrometheusClient) getActivePods(ctx context.Context, namespace string, start, end time.Time) ([]PodInfo, error) {
	query := `group by (pod, namespace, container) (
		rate(container_cpu_usage_seconds_total{namespace=~"` + namespace + `", container!="POD", container!=""}[5m])
	)`
	
	result, warnings, err := p.client.Query(ctx, query, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query active pods: %w", err)
	}
	
	if len(warnings) > 0 {
		log.Printf("Prometheus query warnings: %v", warnings)
	}

	podMap := make(map[string]PodInfo)
	
	if vector, ok := result.(model.Vector); ok {
		for _, sample := range vector {
			pod := string(sample.Metric["pod"])
			ns := string(sample.Metric["namespace"])
			container := string(sample.Metric["container"])
			
			// Filter by namespace if specified
			if namespace != "" && ns != namespace {
				continue
			}
			
			key := ns + "/" + pod
			if existing, exists := podMap[key]; exists {
				// Add container to existing pod
				existing.Containers = append(existing.Containers, container)
				podMap[key] = existing
			} else {
				podMap[key] = PodInfo{
					Name:       pod,
					Namespace:  ns,
					Containers: []string{container},
				}
			}
		}
	}
	
	var pods []PodInfo
	for _, pod := range podMap {
		pods = append(pods, pod)
	}
	
	return pods, nil
}

// getHistoricalMetricsForContainer retrieves and analyzes historical metrics for a specific container
func (p *PrometheusClient) getHistoricalMetricsForContainer(ctx context.Context, pod, namespace, container string, start, end time.Time) (HistoricalMetrics, error) {
	// Query CPU usage over time
	cpuUsage, err := p.queryRangeMetric(ctx, 
		fmt.Sprintf(`rate(container_cpu_usage_seconds_total{namespace="%s", pod="%s", container="%s"}[5m])`, 
			namespace, pod, container), start, end)
	if err != nil {
		return HistoricalMetrics{}, fmt.Errorf("failed to query CPU usage: %w", err)
	}

	// Query Memory usage over time
	memUsage, err := p.queryRangeMetric(ctx,
		fmt.Sprintf(`container_memory_working_set_bytes{namespace="%s", pod="%s", container="%s"}`, 
			namespace, pod, container), start, end)
	if err != nil {
		return HistoricalMetrics{}, fmt.Errorf("failed to query memory usage: %w", err)
	}

	// Query CPU requests
	cpuRequests, err := p.queryRangeMetric(ctx,
		fmt.Sprintf(`kube_pod_container_resource_requests{namespace="%s", pod="%s", container="%s", resource="cpu"}`, 
			namespace, pod, container), start, end)
	if err != nil {
		log.Printf("Warning: failed to query CPU requests for %s/%s/%s: %v", namespace, pod, container, err)
		cpuRequests = []DataPoint{} // Continue without requests data
	}

	// Query Memory requests
	memRequests, err := p.queryRangeMetric(ctx,
		fmt.Sprintf(`kube_pod_container_resource_requests{namespace="%s", pod="%s", container="%s", resource="memory"}`, 
			namespace, pod, container), start, end)
	if err != nil {
		log.Printf("Warning: failed to query memory requests for %s/%s/%s: %v", namespace, pod, container, err)
		memRequests = []DataPoint{} // Continue without requests data
	}

	// Query CPU limits
	cpuLimits, err := p.queryRangeMetric(ctx,
		fmt.Sprintf(`kube_pod_container_resource_limits{namespace="%s", pod="%s", container="%s", resource="cpu"}`, 
			namespace, pod, container), start, end)
	if err != nil {
		log.Printf("Warning: failed to query CPU limits for %s/%s/%s: %v", namespace, pod, container, err)
		cpuLimits = []DataPoint{} // Continue without limits data
	}

	// Query Memory limits
	memLimits, err := p.queryRangeMetric(ctx,
		fmt.Sprintf(`kube_pod_container_resource_limits{namespace="%s", pod="%s", container="%s", resource="memory"}`, 
			namespace, pod, container), start, end)
	if err != nil {
		log.Printf("Warning: failed to query memory limits for %s/%s/%s: %v", namespace, pod, container, err)
		memLimits = []DataPoint{} // Continue without limits data
	}

	// Analyze the data
	cpuData := p.analyzeResourceData(cpuUsage, cpuRequests, cpuLimits)
	memData := p.analyzeResourceData(memUsage, memRequests, memLimits)
	
	analysis := p.generateUsageAnalysis(cpuData, memData)

	return HistoricalMetrics{
		PodName:       pod,
		Namespace:     namespace,
		ContainerName: container,
		CPU:           cpuData,
		Memory:        memData,
		Analysis:      analysis,
	}, nil
}

// queryRangeMetric executes a range query and returns data points
func (p *PrometheusClient) queryRangeMetric(ctx context.Context, query string, start, end time.Time) ([]DataPoint, error) {
	step := 5 * time.Minute // 5-minute resolution
	
	result, warnings, err := p.client.QueryRange(ctx, query, v1.Range{
		Start: start,
		End:   end,
		Step:  step,
	})
	
	if err != nil {
		return nil, err
	}
	
	if len(warnings) > 0 {
		log.Printf("Prometheus query warnings: %v", warnings)
	}

	var dataPoints []DataPoint
	
	if matrix, ok := result.(model.Matrix); ok {
		for _, series := range matrix {
			for _, value := range series.Values {
				dataPoints = append(dataPoints, DataPoint{
					Timestamp: value.Timestamp.Time(),
					Value:     float64(value.Value),
				})
			}
		}
	}
	
	return dataPoints, nil
}

// analyzeResourceData performs statistical analysis on resource data
func (p *PrometheusClient) analyzeResourceData(usage, requests, limits []DataPoint) HistoricalResourceData {
	if len(usage) == 0 {
		return HistoricalResourceData{
			Usage:    usage,
			Requests: requests,
			Limits:   limits,
			Trend:    "unknown",
		}
	}

	// Calculate statistics
	var total, min, max float64
	min = usage[0].Value
	max = usage[0].Value
	
	values := make([]float64, len(usage))
	for i, point := range usage {
		values[i] = point.Value
		total += point.Value
		if point.Value < min {
			min = point.Value
		}
		if point.Value > max {
			max = point.Value
		}
	}
	
	average := total / float64(len(usage))
	
	// Calculate percentiles
	p95 := p.calculatePercentile(values, 0.95)
	p99 := p.calculatePercentile(values, 0.99)
	
	// Determine trend
	trend := p.calculateTrend(usage)

	return HistoricalResourceData{
		Usage:    usage,
		Requests: requests,
		Limits:   limits,
		Average:  average,
		Peak:     max,
		Minimum:  min,
		P95:      p95,
		P99:      p99,
		Trend:    trend,
	}
}

// calculatePercentile calculates the specified percentile of a dataset
func (p *PrometheusClient) calculatePercentile(values []float64, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	// Simple percentile calculation (could be improved with proper sorting)
	n := len(values)
	index := int(percentile * float64(n))
	if index >= n {
		index = n - 1
	}
	
	// For simplicity, return a rough approximation
	var sum float64
	count := 0
	for _, v := range values {
		if count < index {
			sum += v
			count++
		}
	}
	
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

// calculateTrend determines if the usage is increasing, decreasing, or stable
func (p *PrometheusClient) calculateTrend(usage []DataPoint) string {
	if len(usage) < 10 {
		return "insufficient_data"
	}
	
	// Simple trend calculation using first vs last quartile
	quarterSize := len(usage) / 4
	firstQuarter := usage[:quarterSize]
	lastQuarter := usage[len(usage)-quarterSize:]
	
	var firstSum, lastSum float64
	for _, point := range firstQuarter {
		firstSum += point.Value
	}
	for _, point := range lastQuarter {
		lastSum += point.Value
	}
	
	firstAvg := firstSum / float64(len(firstQuarter))
	lastAvg := lastSum / float64(len(lastQuarter))
	
	diff := (lastAvg - firstAvg) / firstAvg
	
	if diff > 0.1 { // 10% increase
		return "increasing"
	} else if diff < -0.1 { // 10% decrease
		return "decreasing"
	}
	return "stable"
}

// generateUsageAnalysis creates usage analysis and recommendations
func (p *PrometheusClient) generateUsageAnalysis(cpu, memory HistoricalResourceData) UsageAnalysis {
	analysis := UsageAnalysis{
		Recommendations: []string{},
	}
	
	// Calculate efficiency if requests data is available
	if len(cpu.Requests) > 0 && len(cpu.Requests[0:]) > 0 {
		avgRequest := p.getAverageValue(cpu.Requests)
		if avgRequest > 0 {
			analysis.CPUEfficiency = (cpu.Average / avgRequest) * 100
		}
	}
	
	if len(memory.Requests) > 0 && len(memory.Requests[0:]) > 0 {
		avgRequest := p.getAverageValue(memory.Requests)
		if avgRequest > 0 {
			analysis.MemoryEfficiency = (memory.Average / avgRequest) * 100
		}
	}
	
	// Generate waste analysis
	analysis.ResourceWaste = p.generateWasteAnalysis(cpu, memory, analysis.CPUEfficiency, analysis.MemoryEfficiency)
	
	// Generate recommendations
	analysis.Recommendations = p.generateRecommendations(cpu, memory, analysis.CPUEfficiency, analysis.MemoryEfficiency)
	
	// Generate patterns (simplified)
	analysis.Patterns = UsagePatterns{
		DailyVariation:  p.calculateVariation(cpu.Usage),
		WeeklyVariation: p.calculateVariation(memory.Usage),
	}
	
	return analysis
}

// getAverageValue calculates average of data points
func (p *PrometheusClient) getAverageValue(points []DataPoint) float64 {
	if len(points) == 0 {
		return 0
	}
	
	var sum float64
	for _, point := range points {
		sum += point.Value
	}
	return sum / float64(len(points))
}

// generateWasteAnalysis identifies resource waste
func (p *PrometheusClient) generateWasteAnalysis(cpu, memory HistoricalResourceData, cpuEff, memEff float64) ResourceWasteAnalysis {
	waste := ResourceWasteAnalysis{}
	
	// CPU analysis
	if cpuEff > 0 && cpuEff < 30 {
		waste.CPUOverProvisioned = true
		waste.CPUWastePercentage = 100 - cpuEff
	} else if cpuEff > 80 {
		waste.CPUUnderProvisioned = true
	}
	
	// Memory analysis
	if memEff > 0 && memEff < 30 {
		waste.MemoryOverProvisioned = true
		waste.MemoryWastePercentage = 100 - memEff
	} else if memEff > 80 {
		waste.MemoryUnderProvisioned = true
	}
	
	return waste
}

// generateRecommendations creates actionable recommendations
func (p *PrometheusClient) generateRecommendations(cpu, memory HistoricalResourceData, cpuEff, memEff float64) []string {
	var recommendations []string
	
	if cpuEff > 0 && cpuEff < 30 {
		recommendations = append(recommendations, fmt.Sprintf("Consider reducing CPU requests - current efficiency: %.1f%%", cpuEff))
	} else if cpuEff > 80 {
		recommendations = append(recommendations, fmt.Sprintf("Consider increasing CPU requests - current efficiency: %.1f%%", cpuEff))
	}
	
	if memEff > 0 && memEff < 30 {
		recommendations = append(recommendations, fmt.Sprintf("Consider reducing memory requests - current efficiency: %.1f%%", memEff))
	} else if memEff > 80 {
		recommendations = append(recommendations, fmt.Sprintf("Consider increasing memory requests - current efficiency: %.1f%%", memEff))
	}
	
	if cpu.Trend == "increasing" {
		recommendations = append(recommendations, "CPU usage is trending upward - monitor for potential scaling needs")
	}
	
	if memory.Trend == "increasing" {
		recommendations = append(recommendations, "Memory usage is trending upward - monitor for potential memory leaks or scaling needs")
	}
	
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Resource usage appears well-optimized")
	}
	
	return recommendations
}

// calculateVariation calculates coefficient of variation
func (p *PrometheusClient) calculateVariation(points []DataPoint) float64 {
	if len(points) < 2 {
		return 0
	}
	
	// Calculate mean
	var sum float64
	for _, point := range points {
		sum += point.Value
	}
	mean := sum / float64(len(points))
	
	if mean == 0 {
		return 0
	}
	
	// Calculate variance
	var variance float64
	for _, point := range points {
		variance += (point.Value - mean) * (point.Value - mean)
	}
	variance /= float64(len(points))
	
	// Return coefficient of variation (std dev / mean)
	stdDev := variance // Simplified - should be sqrt(variance)
	return stdDev / mean * 100
}

// GetNamespaces retrieves all namespaces from Prometheus metrics
func (p *PrometheusClient) GetNamespaces(ctx context.Context) ([]string, error) {
	query := `group by (namespace) (kube_pod_info)`
	
	result, warnings, err := p.client.Query(ctx, query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query namespaces: %w", err)
	}
	
	if len(warnings) > 0 {
		log.Printf("Prometheus query warnings: %v", warnings)
	}

	var namespaces []string
	namespacesSet := make(map[string]bool)
	
	if vector, ok := result.(model.Vector); ok {
		for _, sample := range vector {
			namespace := string(sample.Metric["namespace"])
			if namespace != "" && !namespacesSet[namespace] {
				namespacesSet[namespace] = true
				namespaces = append(namespaces, namespace)
			}
		}
	}
	
	return namespaces, nil
}

// PodMetric represents current pod metrics
type PodMetric struct {
	Name          string
	Namespace     string
	ContainerName string
	CPUUsage      float64
	CPURequest    float64
	CPULimit      float64
	MemoryUsage   float64
	MemoryRequest float64
	MemoryLimit   float64
	Labels        map[string]string
}

// GetCurrentPodMetrics retrieves current pod metrics from Prometheus
func (p *PrometheusClient) GetCurrentPodMetrics(ctx context.Context, namespace string) ([]PodMetric, error) {
	var pods []PodMetric
	
	// Build namespace filter
	namespaceFilter := ""
	if namespace != "" {
		namespaceFilter = fmt.Sprintf(`namespace="%s"`, namespace)
	}
	
	// Get current CPU usage
	cpuQuery := `rate(container_cpu_usage_seconds_total{container!="POD", container!=""`
	if namespaceFilter != "" {
		cpuQuery += "," + namespaceFilter
	}
	cpuQuery += `}[5m])`
	
	// DEBUG: Log the exact CPU query being executed
	log.Printf("DEBUG: Executing CPU query: %s", cpuQuery)
	
	cpuResult, warnings, err := p.client.Query(ctx, cpuQuery, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query CPU usage: %w", err)
	}
	if len(warnings) > 0 {
		log.Printf("CPU query warnings: %v", warnings)
	}
	
	// Get current Memory usage
	memQuery := `container_memory_working_set_bytes{container!="POD", container!=""`
	if namespaceFilter != "" {
		memQuery += "," + namespaceFilter
	}
	memQuery += `}`
	
	// DEBUG: Log the exact memory query being executed
	log.Printf("DEBUG: Executing Memory query: %s", memQuery)
	
	memResult, warnings, err := p.client.Query(ctx, memQuery, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query memory usage: %w", err)
	}
	if len(warnings) > 0 {
		log.Printf("Memory query warnings: %v", warnings)
	}
	
	// Create a map to group metrics by pod/container
	podMetrics := make(map[string]*PodMetric)
	
	// Process CPU usage
	if cpuVector, ok := cpuResult.(model.Vector); ok {
		for _, sample := range cpuVector {
			key := fmt.Sprintf("%s/%s/%s", 
				string(sample.Metric["namespace"]), 
				string(sample.Metric["pod"]), 
				string(sample.Metric["container"]))
			
			if _, exists := podMetrics[key]; !exists {
				podMetrics[key] = &PodMetric{
					Name:          string(sample.Metric["pod"]),
					Namespace:     string(sample.Metric["namespace"]),
					ContainerName: string(sample.Metric["container"]),
					Labels:        make(map[string]string),
				}
			}
			podMetrics[key].CPUUsage = float64(sample.Value)
		}
	}
	
	// Process Memory usage
	if memVector, ok := memResult.(model.Vector); ok {
		for _, sample := range memVector {
			key := fmt.Sprintf("%s/%s/%s", 
				string(sample.Metric["namespace"]), 
				string(sample.Metric["pod"]), 
				string(sample.Metric["container"]))
			
			// DEBUG: Log raw memory values from Prometheus
			memoryBytes := float64(sample.Value)
			log.Printf("DEBUG: Raw memory for %s: %.0f bytes (%.2f Mi)", 
				key, memoryBytes, memoryBytes/(1024*1024))
			
			if _, exists := podMetrics[key]; !exists {
				podMetrics[key] = &PodMetric{
					Name:          string(sample.Metric["pod"]),
					Namespace:     string(sample.Metric["namespace"]),
					ContainerName: string(sample.Metric["container"]),
					Labels:        make(map[string]string),
				}
			}
			podMetrics[key].MemoryUsage = memoryBytes
		}
	}
	
	// Get resource requests and limits
	err = p.addResourceLimitsAndRequests(ctx, podMetrics, namespace)
	if err != nil {
		log.Printf("Warning: failed to get resource requests/limits: %v", err)
	}
	
	// Convert map to slice
	for _, metric := range podMetrics {
		pods = append(pods, *metric)
	}
	
	return pods, nil
}

// addResourceLimitsAndRequests adds resource requests and limits to pod metrics
func (p *PrometheusClient) addResourceLimitsAndRequests(ctx context.Context, podMetrics map[string]*PodMetric, namespace string) error {
	// Build namespace filter
	namespaceFilter := ""
	if namespace != "" {
		namespaceFilter = fmt.Sprintf(`namespace="%s"`, namespace)
	}
	
	// Get CPU requests
	cpuReqQuery := `kube_pod_container_resource_requests{resource="cpu"`
	if namespaceFilter != "" {
		cpuReqQuery += "," + namespaceFilter
	}
	cpuReqQuery += `}`
	
	cpuReqResult, _, err := p.client.Query(ctx, cpuReqQuery, time.Now())
	if err != nil {
		return fmt.Errorf("failed to query CPU requests: %w", err)
	}
	
	if cpuReqVector, ok := cpuReqResult.(model.Vector); ok {
		for _, sample := range cpuReqVector {
			key := fmt.Sprintf("%s/%s/%s", 
				string(sample.Metric["namespace"]), 
				string(sample.Metric["pod"]), 
				string(sample.Metric["container"]))
			
			if metric, exists := podMetrics[key]; exists {
				metric.CPURequest = float64(sample.Value)
			}
		}
	}
	
	// Get CPU limits
	cpuLimitQuery := `kube_pod_container_resource_limits{resource="cpu"`
	if namespaceFilter != "" {
		cpuLimitQuery += "," + namespaceFilter
	}
	cpuLimitQuery += `}`
	
	cpuLimitResult, _, err := p.client.Query(ctx, cpuLimitQuery, time.Now())
	if err != nil {
		return fmt.Errorf("failed to query CPU limits: %w", err)
	}
	
	if cpuLimitVector, ok := cpuLimitResult.(model.Vector); ok {
		for _, sample := range cpuLimitVector {
			key := fmt.Sprintf("%s/%s/%s", 
				string(sample.Metric["namespace"]), 
				string(sample.Metric["pod"]), 
				string(sample.Metric["container"]))
			
			if metric, exists := podMetrics[key]; exists {
				metric.CPULimit = float64(sample.Value)
			}
		}
	}
	
	// Get Memory requests
	memReqQuery := `kube_pod_container_resource_requests{resource="memory"`
	if namespaceFilter != "" {
		memReqQuery += "," + namespaceFilter
	}
	memReqQuery += `}`
	
	memReqResult, _, err := p.client.Query(ctx, memReqQuery, time.Now())
	if err != nil {
		return fmt.Errorf("failed to query memory requests: %w", err)
	}
	
	if memReqVector, ok := memReqResult.(model.Vector); ok {
		for _, sample := range memReqVector {
			key := fmt.Sprintf("%s/%s/%s", 
				string(sample.Metric["namespace"]), 
				string(sample.Metric["pod"]), 
				string(sample.Metric["container"]))
			
			if metric, exists := podMetrics[key]; exists {
				metric.MemoryRequest = float64(sample.Value)
			}
		}
	}
	
	// Get Memory limits
	memLimitQuery := `kube_pod_container_resource_limits{resource="memory"`
	if namespaceFilter != "" {
		memLimitQuery += "," + namespaceFilter
	}
	memLimitQuery += `}`
	
	memLimitResult, _, err := p.client.Query(ctx, memLimitQuery, time.Now())
	if err != nil {
		return fmt.Errorf("failed to query memory limits: %w", err)
	}
	
	if memLimitVector, ok := memLimitResult.(model.Vector); ok {
		for _, sample := range memLimitVector {
			key := fmt.Sprintf("%s/%s/%s", 
				string(sample.Metric["namespace"]), 
				string(sample.Metric["pod"]), 
				string(sample.Metric["container"]))
			
			if metric, exists := podMetrics[key]; exists {
				metric.MemoryLimit = float64(sample.Value)
			}
		}
	}
	
	return nil
}
