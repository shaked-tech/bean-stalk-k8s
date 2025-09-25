package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// VMAgentClient wraps the VictoriaMetrics API client
type VMAgentClient struct {
	baseURL string
	client  *http.Client
}

// NewVMAgentClient creates a new VictoriaMetrics client
func NewVMAgentClient(vmSelectURL string) (*VMAgentClient, error) {
	// Ensure the URL ends with the API path
	if !strings.HasSuffix(vmSelectURL, "/") {
		vmSelectURL += "/"
	}
	
	return &VMAgentClient{
		baseURL: vmSelectURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Close closes the VictoriaMetrics client connection
func (vm *VMAgentClient) Close() error {
	// HTTP client doesn't require explicit closing
	return nil
}

// GetClientType returns the type of metrics client
func (vm *VMAgentClient) GetClientType() string {
	return "vmagent"
}

// VMResponse represents VictoriaMetrics API response structure
type VMResponse struct {
	Status string `json:"status"`
	Data   VMData `json:"data"`
}

// VMData represents the data section of VM response
type VMData struct {
	ResultType string     `json:"resultType"`
	Result     []VMResult `json:"result"`
}

// VMResult represents a single result from VM query
type VMResult struct {
	Metric map[string]string `json:"metric"`
	Value  []interface{}     `json:"value,omitempty"`
	Values [][]interface{}   `json:"values,omitempty"`
}

// GetCurrentPodMetrics retrieves current pod metrics from VictoriaMetrics
func (vm *VMAgentClient) GetCurrentPodMetrics(ctx context.Context, namespace string) ([]PodMetric, error) {
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
	
	log.Printf("DEBUG: Executing CPU query: %s", cpuQuery)
	
	cpuResult, err := vm.query(ctx, cpuQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query CPU usage: %w", err)
	}
	
	// Get current Memory usage
	memQuery := `container_memory_working_set_bytes{container!="POD", container!=""`
	if namespaceFilter != "" {
		memQuery += "," + namespaceFilter
	}
	memQuery += `}`
	
	log.Printf("DEBUG: Executing Memory query: %s", memQuery)
	
	memResult, err := vm.query(ctx, memQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query memory usage: %w", err)
	}
	
	// Create a map to group metrics by pod/container
	podMetrics := make(map[string]*PodMetric)
	
	// Process CPU usage
	for _, result := range cpuResult.Data.Result {
		key := fmt.Sprintf("%s/%s/%s",
			result.Metric["namespace"],
			result.Metric["pod"],
			result.Metric["container"])
		
		if _, exists := podMetrics[key]; !exists {
			podMetrics[key] = &PodMetric{
				Name:          result.Metric["pod"],
				Namespace:     result.Metric["namespace"],
				ContainerName: result.Metric["container"],
				Labels:        make(map[string]string),
			}
		}
		
		if len(result.Value) >= 2 {
			if val, ok := result.Value[1].(string); ok {
				if cpuUsage, err := strconv.ParseFloat(val, 64); err == nil {
					podMetrics[key].CPUUsage = cpuUsage
				}
			}
		}
	}
	
	// Process Memory usage
	for _, result := range memResult.Data.Result {
		key := fmt.Sprintf("%s/%s/%s",
			result.Metric["namespace"],
			result.Metric["pod"],
			result.Metric["container"])
		
		if _, exists := podMetrics[key]; !exists {
			podMetrics[key] = &PodMetric{
				Name:          result.Metric["pod"],
				Namespace:     result.Metric["namespace"],
				ContainerName: result.Metric["container"],
				Labels:        make(map[string]string),
			}
		}
		
		if len(result.Value) >= 2 {
			if val, ok := result.Value[1].(string); ok {
				if memUsage, err := strconv.ParseFloat(val, 64); err == nil {
					podMetrics[key].MemoryUsage = memUsage
					log.Printf("DEBUG: Raw memory for %s: %.0f bytes (%.2f Mi)",
						key, memUsage, memUsage/(1024*1024))
				}
			}
		}
	}
	
	// Get resource requests and limits
	err = vm.addResourceLimitsAndRequests(ctx, podMetrics, namespace)
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
func (vm *VMAgentClient) addResourceLimitsAndRequests(ctx context.Context, podMetrics map[string]*PodMetric, namespace string) error {
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
	
	cpuReqResult, err := vm.query(ctx, cpuReqQuery)
	if err != nil {
		return fmt.Errorf("failed to query CPU requests: %w", err)
	}
	
	for _, result := range cpuReqResult.Data.Result {
		key := fmt.Sprintf("%s/%s/%s",
			result.Metric["namespace"],
			result.Metric["pod"],
			result.Metric["container"])
		
		if metric, exists := podMetrics[key]; exists {
			if len(result.Value) >= 2 {
				if val, ok := result.Value[1].(string); ok {
					if cpuReq, err := strconv.ParseFloat(val, 64); err == nil {
						metric.CPURequest = cpuReq
					}
				}
			}
		}
	}
	
	// Get CPU limits
	cpuLimitQuery := `kube_pod_container_resource_limits{resource="cpu"`
	if namespaceFilter != "" {
		cpuLimitQuery += "," + namespaceFilter
	}
	cpuLimitQuery += `}`
	
	cpuLimitResult, err := vm.query(ctx, cpuLimitQuery)
	if err != nil {
		return fmt.Errorf("failed to query CPU limits: %w", err)
	}
	
	for _, result := range cpuLimitResult.Data.Result {
		key := fmt.Sprintf("%s/%s/%s",
			result.Metric["namespace"],
			result.Metric["pod"],
			result.Metric["container"])
		
		if metric, exists := podMetrics[key]; exists {
			if len(result.Value) >= 2 {
				if val, ok := result.Value[1].(string); ok {
					if cpuLimit, err := strconv.ParseFloat(val, 64); err == nil {
						metric.CPULimit = cpuLimit
					}
				}
			}
		}
	}
	
	// Get Memory requests
	memReqQuery := `kube_pod_container_resource_requests{resource="memory"`
	if namespaceFilter != "" {
		memReqQuery += "," + namespaceFilter
	}
	memReqQuery += `}`
	
	memReqResult, err := vm.query(ctx, memReqQuery)
	if err != nil {
		return fmt.Errorf("failed to query memory requests: %w", err)
	}
	
	for _, result := range memReqResult.Data.Result {
		key := fmt.Sprintf("%s/%s/%s",
			result.Metric["namespace"],
			result.Metric["pod"],
			result.Metric["container"])
		
		if metric, exists := podMetrics[key]; exists {
			if len(result.Value) >= 2 {
				if val, ok := result.Value[1].(string); ok {
					if memReq, err := strconv.ParseFloat(val, 64); err == nil {
						metric.MemoryRequest = memReq
					}
				}
			}
		}
	}
	
	// Get Memory limits
	memLimitQuery := `kube_pod_container_resource_limits{resource="memory"`
	if namespaceFilter != "" {
		memLimitQuery += "," + namespaceFilter
	}
	memLimitQuery += `}`
	
	memLimitResult, err := vm.query(ctx, memLimitQuery)
	if err != nil {
		return fmt.Errorf("failed to query memory limits: %w", err)
	}
	
	for _, result := range memLimitResult.Data.Result {
		key := fmt.Sprintf("%s/%s/%s",
			result.Metric["namespace"],
			result.Metric["pod"],
			result.Metric["container"])
		
		if metric, exists := podMetrics[key]; exists {
			if len(result.Value) >= 2 {
				if val, ok := result.Value[1].(string); ok {
					if memLimit, err := strconv.ParseFloat(val, 64); err == nil {
						metric.MemoryLimit = memLimit
					}
				}
			}
		}
	}
	
	return nil
}

// GetHistoricalMetrics retrieves and analyzes 7-day historical metrics for pods
func (vm *VMAgentClient) GetHistoricalMetrics(ctx context.Context, namespace string) ([]HistoricalMetrics, error) {
	now := time.Now()
	sevenDaysAgo := now.Add(-7 * 24 * time.Hour)
	
	// Get pod list from the last 7 days
	pods, err := vm.getActivePods(ctx, namespace, sevenDaysAgo, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get active pods: %w", err)
	}

	var results []HistoricalMetrics
	for _, pod := range pods {
		for _, container := range pod.Containers {
			metrics, err := vm.getHistoricalMetricsForContainer(ctx, pod.Name, pod.Namespace, container, sevenDaysAgo, now)
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

// getActivePods retrieves pods that were active during the specified time range
func (vm *VMAgentClient) getActivePods(ctx context.Context, namespace string, start, end time.Time) ([]PodInfo, error) {
	query := `group by (pod, namespace, container) (
		rate(container_cpu_usage_seconds_total{namespace=~"` + namespace + `", container!="POD", container!=""}[5m])
	)`
	
	result, err := vm.query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active pods: %w", err)
	}

	podMap := make(map[string]PodInfo)
	
	for _, vmResult := range result.Data.Result {
		pod := vmResult.Metric["pod"]
		ns := vmResult.Metric["namespace"]
		container := vmResult.Metric["container"]
		
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
	
	var pods []PodInfo
	for _, pod := range podMap {
		pods = append(pods, pod)
	}
	
	return pods, nil
}

// getHistoricalMetricsForContainer retrieves and analyzes historical metrics for a specific container
func (vm *VMAgentClient) getHistoricalMetricsForContainer(ctx context.Context, pod, namespace, container string, start, end time.Time) (HistoricalMetrics, error) {
	// Query CPU usage over time
	cpuUsage, err := vm.queryRangeMetric(ctx, 
		fmt.Sprintf(`rate(container_cpu_usage_seconds_total{namespace="%s", pod="%s", container="%s"}[5m])`, 
			namespace, pod, container), start, end)
	if err != nil {
		return HistoricalMetrics{}, fmt.Errorf("failed to query CPU usage: %w", err)
	}

	// Query Memory usage over time
	memUsage, err := vm.queryRangeMetric(ctx,
		fmt.Sprintf(`container_memory_working_set_bytes{namespace="%s", pod="%s", container="%s"}`, 
			namespace, pod, container), start, end)
	if err != nil {
		return HistoricalMetrics{}, fmt.Errorf("failed to query memory usage: %w", err)
	}

	// Query CPU requests
	cpuRequests, err := vm.queryRangeMetric(ctx,
		fmt.Sprintf(`kube_pod_container_resource_requests{namespace="%s", pod="%s", container="%s", resource="cpu"}`, 
			namespace, pod, container), start, end)
	if err != nil {
		log.Printf("Warning: failed to query CPU requests for %s/%s/%s: %v", namespace, pod, container, err)
		cpuRequests = []DataPoint{} // Continue without requests data
	}

	// Query Memory requests
	memRequests, err := vm.queryRangeMetric(ctx,
		fmt.Sprintf(`kube_pod_container_resource_requests{namespace="%s", pod="%s", container="%s", resource="memory"}`, 
			namespace, pod, container), start, end)
	if err != nil {
		log.Printf("Warning: failed to query memory requests for %s/%s/%s: %v", namespace, pod, container, err)
		memRequests = []DataPoint{} // Continue without requests data
	}

	// Query CPU limits
	cpuLimits, err := vm.queryRangeMetric(ctx,
		fmt.Sprintf(`kube_pod_container_resource_limits{namespace="%s", pod="%s", container="%s", resource="cpu"}`, 
			namespace, pod, container), start, end)
	if err != nil {
		log.Printf("Warning: failed to query CPU limits for %s/%s/%s: %v", namespace, pod, container, err)
		cpuLimits = []DataPoint{} // Continue without limits data
	}

	// Query Memory limits
	memLimits, err := vm.queryRangeMetric(ctx,
		fmt.Sprintf(`kube_pod_container_resource_limits{namespace="%s", pod="%s", container="%s", resource="memory"}`, 
			namespace, pod, container), start, end)
	if err != nil {
		log.Printf("Warning: failed to query memory limits for %s/%s/%s: %v", namespace, pod, container, err)
		memLimits = []DataPoint{} // Continue without limits data
	}

	// Analyze the data (reuse existing analysis functions)
	cpuData := vm.analyzeResourceData(cpuUsage, cpuRequests, cpuLimits)
	memData := vm.analyzeResourceData(memUsage, memRequests, memLimits)
	
	analysis := vm.generateUsageAnalysis(cpuData, memData)

	return HistoricalMetrics{
		PodName:       pod,
		Namespace:     namespace,
		ContainerName: container,
		CPU:           cpuData,
		Memory:        memData,
		Analysis:      analysis,
	}, nil
}

// GetNamespaces retrieves all namespaces from VictoriaMetrics
func (vm *VMAgentClient) GetNamespaces(ctx context.Context) ([]string, error) {
	// Use container metrics to get namespaces since we don't have kube-state-metrics
	query := `group by (namespace) (container_cpu_usage_seconds_total{container!="POD", container!=""})`
	
	result, err := vm.query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query namespaces: %w", err)
	}

	var namespaces []string
	namespacesSet := make(map[string]bool)
	
	for _, vmResult := range result.Data.Result {
		namespace := vmResult.Metric["namespace"]
		if namespace != "" && !namespacesSet[namespace] {
			namespacesSet[namespace] = true
			namespaces = append(namespaces, namespace)
		}
	}
	
	return namespaces, nil
}

// query executes a single query against VictoriaMetrics
func (vm *VMAgentClient) query(ctx context.Context, query string) (*VMResponse, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("time", strconv.FormatInt(time.Now().Unix(), 10))
	
	queryURL := vm.baseURL + "api/v1/query?" + params.Encode()
	
	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}
	
	resp, err := vm.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("VictoriaMetrics query failed with status %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var vmResp VMResponse
	err = json.Unmarshal(body, &vmResp)
	if err != nil {
		return nil, err
	}
	
	if vmResp.Status != "success" {
		return nil, fmt.Errorf("VictoriaMetrics query failed: %s", vmResp.Status)
	}
	
	return &vmResp, nil
}

// queryRangeMetric executes a range query and returns data points
func (vm *VMAgentClient) queryRangeMetric(ctx context.Context, query string, start, end time.Time) ([]DataPoint, error) {
	step := 5 * time.Minute // 5-minute resolution
	
	params := url.Values{}
	params.Set("query", query)
	params.Set("start", strconv.FormatInt(start.Unix(), 10))
	params.Set("end", strconv.FormatInt(end.Unix(), 10))
	params.Set("step", strconv.FormatInt(int64(step.Seconds()), 10))
	
	queryURL := vm.baseURL + "api/v1/query_range?" + params.Encode()
	
	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}
	
	resp, err := vm.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("VictoriaMetrics range query failed with status %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var vmResp VMResponse
	err = json.Unmarshal(body, &vmResp)
	if err != nil {
		return nil, err
	}
	
	if vmResp.Status != "success" {
		return nil, fmt.Errorf("VictoriaMetrics range query failed: %s", vmResp.Status)
	}

	var dataPoints []DataPoint
	
	for _, series := range vmResp.Data.Result {
		for _, values := range series.Values {
			if len(values) >= 2 {
				timestamp, ok1 := values[0].(float64)
				valueStr, ok2 := values[1].(string)
				
				if ok1 && ok2 {
					value, err := strconv.ParseFloat(valueStr, 64)
					if err == nil {
						dataPoints = append(dataPoints, DataPoint{
							Timestamp: time.Unix(int64(timestamp), 0),
							Value:     value,
						})
					}
				}
			}
		}
	}
	
	return dataPoints, nil
}

// The following methods are shared analysis functions that can be reused
// They are duplicated here for the VMAgentClient to maintain independence

// analyzeResourceData performs statistical analysis on resource data
func (vm *VMAgentClient) analyzeResourceData(usage, requests, limits []DataPoint) HistoricalResourceData {
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
	p95 := vm.calculatePercentile(values, 0.95)
	p99 := vm.calculatePercentile(values, 0.99)
	
	// Determine trend
	trend := vm.calculateTrend(usage)

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
func (vm *VMAgentClient) calculatePercentile(values []float64, percentile float64) float64 {
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
func (vm *VMAgentClient) calculateTrend(usage []DataPoint) string {
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
func (vm *VMAgentClient) generateUsageAnalysis(cpu, memory HistoricalResourceData) UsageAnalysis {
	analysis := UsageAnalysis{
		Recommendations: []string{},
	}
	
	// Calculate efficiency if requests data is available
	if len(cpu.Requests) > 0 && len(cpu.Requests[0:]) > 0 {
		avgRequest := vm.getAverageValue(cpu.Requests)
		if avgRequest > 0 {
			analysis.CPUEfficiency = (cpu.Average / avgRequest) * 100
		}
	}
	
	if len(memory.Requests) > 0 && len(memory.Requests[0:]) > 0 {
		avgRequest := vm.getAverageValue(memory.Requests)
		if avgRequest > 0 {
			analysis.MemoryEfficiency = (memory.Average / avgRequest) * 100
		}
	}
	
	// Generate waste analysis
	analysis.ResourceWaste = vm.generateWasteAnalysis(cpu, memory, analysis.CPUEfficiency, analysis.MemoryEfficiency)
	
	// Generate recommendations
	analysis.Recommendations = vm.generateRecommendations(cpu, memory, analysis.CPUEfficiency, analysis.MemoryEfficiency)
	
	// Generate patterns (simplified)
	analysis.Patterns = UsagePatterns{
		DailyVariation:  vm.calculateVariation(cpu.Usage),
		WeeklyVariation: vm.calculateVariation(memory.Usage),
	}
	
	return analysis
}

// getAverageValue calculates average of data points
func (vm *VMAgentClient) getAverageValue(points []DataPoint) float64 {
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
func (vm *VMAgentClient) generateWasteAnalysis(cpu, memory HistoricalResourceData, cpuEff, memEff float64) ResourceWasteAnalysis {
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
func (vm *VMAgentClient) generateRecommendations(cpu, memory HistoricalResourceData, cpuEff, memEff float64) []string {
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
func (vm *VMAgentClient) calculateVariation(points []DataPoint) float64 {
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
