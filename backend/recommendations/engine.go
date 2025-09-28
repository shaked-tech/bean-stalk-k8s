package recommendations

import (
	"fmt"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bean-stalk-k8s/backend/k8s"
	"github.com/bean-stalk-k8s/backend/models"
)

// RecommendationEngine generates resource recommendations based on current usage analysis
type RecommendationEngine struct {
	config models.RecommendationConfig
}

// NewRecommendationEngine creates a new recommendation engine with the given configuration
func NewRecommendationEngine(config models.RecommendationConfig) *RecommendationEngine {
	return &RecommendationEngine{
		config: config,
	}
}

// GenerateRecommendations analyzes current metrics and generates resource recommendations
func (e *RecommendationEngine) GenerateRecommendations(currentMetrics []k8s.PodMetric) (models.RecommendationsResponse, error) {
	log.Printf("INFO: Starting recommendation generation for %d pods based on current usage", len(currentMetrics))
	
	var recommendations []models.PodResourceRecommendation
	
	// Generate recommendations for each pod based on current usage
	for _, current := range currentMetrics {
		recommendation, err := e.generatePodRecommendation(current)
		if err != nil {
			log.Printf("WARN: Failed to generate recommendation for pod %s/%s: %v", current.Namespace, current.Name, err)
			continue
		}
		
		recommendations = append(recommendations, recommendation)
	}
	
	// Generate summary
	summary := e.generateSummary(recommendations)
	
	// Create response
	response := models.RecommendationsResponse{
		Recommendations:   recommendations,
		Summary:          summary,
		GeneratedAt:      time.Now(),
		AnalysisWindow:    "current usage",
		TargetUtilization: e.config.TargetCPUUtilization,
	}
	
	log.Printf("INFO: Generated %d recommendations with %d high priority items", len(recommendations), summary.HighPriorityRecommendations)
	
	return response, nil
}

// generatePodRecommendation creates a recommendation for a single pod based on current usage
func (e *RecommendationEngine) generatePodRecommendation(current k8s.PodMetric) (models.PodResourceRecommendation, error) {
	
	// Initialize recommendation structure
	recommendation := models.PodResourceRecommendation{
		PodName:       current.Name,
		Namespace:     current.Namespace,
		ContainerName: current.ContainerName,
		LastUpdated:   time.Now(),
		ApplicableFrom: time.Now(),
		HistoricalAvgCPU: current.CPUUsage,    // Current usage as "historical"
		HistoricalAvgMemory: current.MemoryUsage, // Current usage as "historical"
	}
	
	var reasons []models.RecommendationReason
	
	// Generate CPU recommendation
	cpuRecommendation, cpuReasons := e.generateCPURecommendationFromCurrent(current.CPUUsage, current.CPURequest, current.CPULimit)
	recommendation.CPU = cpuRecommendation
	reasons = append(reasons, cpuReasons...)
	
	// Generate Memory recommendation
	memoryRecommendation, memoryReasons := e.generateMemoryRecommendationFromCurrent(current.MemoryUsage, current.MemoryRequest, current.MemoryLimit)
	recommendation.Memory = memoryRecommendation
	reasons = append(reasons, memoryReasons...)
	
	// Calculate overall metrics based on current data
	recommendation.Reasons = reasons
	recommendation.ConfidenceScore = 95.0 // High confidence with current data
	recommendation.DataQuality = "excellent" // Current data is always excellent quality
	recommendation.Priority = e.calculatePriority(reasons, recommendation.ConfidenceScore)
	recommendation.RiskLevel = e.assessRiskLevel(cpuRecommendation, memoryRecommendation)
	recommendation.EstimatedSavings = e.calculateEstimatedSavings(cpuRecommendation, memoryRecommendation)
	
	return recommendation, nil
}

// generateCPURecommendationFromCurrent creates CPU-specific recommendations based on current usage
func (e *RecommendationEngine) generateCPURecommendationFromCurrent(currentUsage, currentRequest, currentLimit float64) (models.ResourceRecommendation, []models.RecommendationReason) {
	var reasons []models.RecommendationReason
	
	log.Printf("DEBUG: CPU analysis - Usage: %.4f, Request: %.4f, Limit: %.4f", currentUsage, currentRequest, currentLimit)
	
	// Calculate target request to achieve desired utilization (currentUsage / targetUtilization)
	var targetRequest float64
	if currentUsage > 0 {
		targetRequest = currentUsage / (e.config.TargetCPUUtilization / 100.0)
	} else {
		// If no usage data, use minimum CPU as safe default
		targetRequest = e.parseResourceValue(e.config.MinCPURequest, "cpu")
	}
	
	// Apply safety bounds
	minCPU := e.parseResourceValue(e.config.MinCPURequest, "cpu")
	maxCPU := e.parseResourceValue(e.config.MaxCPURequest, "cpu")
	
	// Apply scaling limits if current request exists
	if currentRequest > 0 {
		maxAllowedIncrease := currentRequest * e.config.MaxScaleUpFactor
		maxAllowedDecrease := currentRequest * e.config.MaxScaleDownFactor
		
		if targetRequest > maxAllowedIncrease {
			log.Printf("DEBUG: CPU target %.4f limited by max scale up %.4f", targetRequest, maxAllowedIncrease)
			targetRequest = maxAllowedIncrease
		} else if targetRequest < maxAllowedDecrease {
			log.Printf("DEBUG: CPU target %.4f limited by max scale down %.4f", targetRequest, maxAllowedDecrease)
			targetRequest = maxAllowedDecrease
		}
	}
	
	// Ensure within absolute bounds
	if targetRequest < minCPU {
		targetRequest = minCPU
	} else if targetRequest > maxCPU {
		targetRequest = maxCPU
	}
	
	log.Printf("DEBUG: CPU final target request: %.4f", targetRequest)
	
	// Determine current utilization
	currentUtilization := 0.0
	if currentRequest > 0 && currentUsage > 0 {
		currentUtilization = (currentUsage / currentRequest) * 100
	}
	
	// Determine change type and reasons
	var changeType string
	percentageChange := 0.0
	
	if currentRequest == 0 {
		reasons = append(reasons, models.ReasonMissingRequests)
		changeType = "increase"
		percentageChange = 100.0 // New request
	} else {
		percentageChange = ((targetRequest - currentRequest) / currentRequest) * 100
		
		// Check if current utilization is within optimal range (60-75%)
		if currentUtilization >= 60.0 && currentUtilization <= 75.0 {
			changeType = "no_change"
			reasons = append(reasons, models.ReasonWellOptimized)
		} else if currentUtilization > 75.0 {
			changeType = "increase"
			reasons = append(reasons, models.ReasonCPUOverUtilized)
		} else if currentUtilization < 60.0 && currentUtilization > 0 {
			changeType = "decrease"
			reasons = append(reasons, models.ReasonCPUUnderUtilized)
		} else {
			// Handle edge case where utilization is 0 or negative
			changeType = "increase"
			reasons = append(reasons, models.ReasonCPUUnderUtilized)
		}
	}
	
	// Check for CPU limits (should be removed)
	var recommendedLimit *string
	if currentLimit > 0 {
		reasons = append(reasons, models.ReasonCPULimitPresent)
		// Always recommend removing CPU limits
		recommendedLimit = nil
	}
	
	// Format recommended request
	recommendedRequestStr := e.formatCPUValue(targetRequest)
	
	recommendation := models.ResourceRecommendation{
		CurrentRequest:          e.formatCPUValue(currentRequest),
		CurrentLimit:            e.formatCPUValue(currentLimit),
		CurrentUsage:            e.formatCPUValue(currentUsage),
		RecommendedRequest:      &recommendedRequestStr,
		RecommendedLimit:        recommendedLimit,
		CurrentUtilization:      currentUtilization,
		TargetUtilization:       e.config.TargetCPUUtilization,
		CurrentRequestValue:     currentRequest,
		CurrentLimitValue:       currentLimit,
		CurrentUsageValue:       currentUsage,
		RecommendedRequestValue: targetRequest,
		RecommendedLimitValue:   nil, // Always nil for CPU
		ResourceChange:          changeType,
		PercentageChange:        percentageChange,
	}
	
	return recommendation, reasons
}

// generateMemoryRecommendationFromCurrent creates memory-specific recommendations based on current usage
func (e *RecommendationEngine) generateMemoryRecommendationFromCurrent(currentUsage, currentRequest, currentLimit float64) (models.ResourceRecommendation, []models.RecommendationReason) {
	var reasons []models.RecommendationReason
	
	log.Printf("DEBUG: Memory analysis - Usage: %.0f bytes (%.0f Mi), Request: %.0f bytes (%.0f Mi), Limit: %.0f bytes (%.0f Mi)", 
		currentUsage, currentUsage/(1024*1024), currentRequest, currentRequest/(1024*1024), currentLimit, currentLimit/(1024*1024))
	
	// SIMPLIFIED CALCULATION: currentUsage / targetUtilization
	var targetRequest float64
	if currentUsage > 0 {
		// Simple formula: usage / target percentage = recommended request
		targetRequest = currentUsage / (e.config.TargetMemoryUtilization / 100.0)
		log.Printf("STEP 1 - Memory target calculated: %.0f bytes (%.0f Mi) for %.0f%% target utilization", 
			targetRequest, targetRequest/(1024*1024), e.config.TargetMemoryUtilization)
		
		// Apply basic minimum bounds (never go below usage + 10% buffer)
		minMemory := currentUsage * 1.1 // 10% buffer above actual usage
		configMinMemory := e.parseResourceValue(e.config.MinMemoryRequest, "memory")
		if configMinMemory > minMemory {
			minMemory = configMinMemory
		}
		
		log.Printf("STEP 2 - Before bounds check: targetRequest=%.0f, minMemory=%.0f, configMin=%.0f", targetRequest, minMemory, configMinMemory)
		
		if targetRequest < minMemory {
			log.Printf("STEP 3 - Memory target adjusted from %.0f to minimum %.0f bytes", targetRequest, minMemory)
			targetRequest = minMemory
		}
		
		// Apply maximum bounds
		maxMemory := e.parseResourceValue(e.config.MaxMemoryRequest, "memory")
		log.Printf("STEP 4 - Before max check: targetRequest=%.0f, maxMemory=%.0f", targetRequest, maxMemory)
		if targetRequest > maxMemory {
			log.Printf("STEP 5 - Memory target adjusted from %.0f to maximum %.0f bytes", targetRequest, maxMemory)
			targetRequest = maxMemory
		}
		
		log.Printf("STEP 6 - After bounds: targetRequest=%.0f bytes (%.0f Mi)", targetRequest, targetRequest/(1024*1024))
		
		// CRITICAL FIX: Apply scaling limits if current request exists (same as CPU function)
		if currentRequest > 0 {
			maxAllowedIncrease := currentRequest * e.config.MaxScaleUpFactor
			maxAllowedDecrease := currentRequest * e.config.MaxScaleDownFactor
			
			log.Printf("STEP 7 - Scaling limits: current=%.0f, maxIncrease=%.0f, maxDecrease=%.0f", currentRequest, maxAllowedIncrease, maxAllowedDecrease)
			
			if targetRequest > maxAllowedIncrease {
				log.Printf("STEP 8 - Memory target %.0f LIMITED by max scale up %.0f", targetRequest, maxAllowedIncrease)
				targetRequest = maxAllowedIncrease
			} else if targetRequest < maxAllowedDecrease {
				log.Printf("STEP 8 - Memory target %.0f LIMITED by max scale down %.0f", targetRequest, maxAllowedDecrease)
				targetRequest = maxAllowedDecrease
			}
		}
		
		log.Printf("STEP 9 - After scaling limits: targetRequest=%.0f bytes (%.0f Mi)", targetRequest, targetRequest/(1024*1024))
		
	} else {
		// No usage data - use configured minimum
		targetRequest = e.parseResourceValue(e.config.MinMemoryRequest, "memory")
		log.Printf("STEP 1 - No memory usage data, using default minimum: %.0f bytes (%.0f Mi)", targetRequest, targetRequest/(1024*1024))
	}
	
	log.Printf("DEBUG: Memory final target request: %.0f bytes (%.0f Mi)", targetRequest, targetRequest/(1024*1024))
	
	// For memory: requests should equal limits
	targetLimit := targetRequest
	
	// Calculate current utilization
	currentUtilization := 0.0
	if currentRequest > 0 && currentUsage > 0 {
		currentUtilization = (currentUsage / currentRequest) * 100
	}
	
	// Determine change type and reasons - BUT DON'T RESET TARGET REQUEST!
	var changeType string
	percentageChange := 0.0
	
	log.Printf("CRITICAL DEBUG: Before percentage calculation - currentRequest: %.0f bytes (%.0f Mi), targetRequest: %.0f bytes (%.0f Mi)", currentRequest, currentRequest/(1024*1024), targetRequest, targetRequest/(1024*1024))
	
	if currentRequest == 0 {
		reasons = append(reasons, models.ReasonMissingRequests)
		changeType = "increase"
		// For pods with no current request, show the actual recommended amount as percentage of a baseline
		// Use a meaningful baseline (like 1Mi = 1048576 bytes) to calculate percentage
		baselineRequest := float64(1048576) // 1Mi baseline for percentage calculation
		percentageChange = (targetRequest / baselineRequest) * 100
		log.Printf("DEBUG: Pod has no current memory request, using calculated target: %.0f bytes (%.0f Mi), %% change: %.1f%%", targetRequest, targetRequest/(1024*1024), percentageChange)
	} else {
		// Normal percentage calculation for pods with current requests
		percentageChange = ((targetRequest - currentRequest) / currentRequest) * 100
		log.Printf("CRITICAL DEBUG: Normal percentage calculation: (%.0f - %.0f) / %.0f * 100 = %.1f%%", targetRequest, currentRequest, currentRequest, percentageChange)
		
		// Simple utilization check for 60-75% target range
		if currentUtilization >= 60.0 && currentUtilization <= 75.0 {
			changeType = "no_change"
			reasons = append(reasons, models.ReasonWellOptimized)
		} else if currentUtilization > 75.0 {
			changeType = "increase"
			reasons = append(reasons, models.ReasonMemoryOverUtilized)
		} else if currentUtilization < 60.0 && currentUtilization > 0 {
			changeType = "decrease"
			reasons = append(reasons, models.ReasonMemoryUnderUtilized)
		} else {
			changeType = "increase"
			reasons = append(reasons, models.ReasonMemoryUnderUtilized)
		}
	}
	
	log.Printf("CRITICAL DEBUG: After changeType logic - percentageChange: %.1f%%, changeType: %s", percentageChange, changeType)
	
	// Check for memory misalignment (request != limit)
	if e.config.AlignMemoryRequestsLimits && currentRequest > 0 && currentLimit > 0 {
		if math.Abs(currentRequest-currentLimit) > currentRequest*0.01 {
			reasons = append(reasons, models.ReasonMemoryMisaligned)
		}
	}
	
	// Check for missing limits
	if currentLimit == 0 && currentRequest > 0 {
		reasons = append(reasons, models.ReasonMissingLimits)
	}
	
	// CRITICAL BUG FIX: Ensure targetRequest is not 0 when we have usage data
	if currentUsage > 0 && targetRequest == 0 {
		// Recalculate if somehow targetRequest got reset
		targetRequest = currentUsage / (e.config.TargetMemoryUtilization / 100.0)
		targetLimit = targetRequest
		log.Printf("ERROR: Target request was 0 but should be %.0f bytes (%.0f Mi) - FIXED!", targetRequest, targetRequest/(1024*1024))
	}
	
	log.Printf("DEBUG: About to format memory - targetRequest: %.0f bytes (%.0f Mi)", targetRequest, targetRequest/(1024*1024))
	
	// Format recommended values
	recommendedRequestStr := e.formatMemoryValue(targetRequest)
	recommendedLimitStr := e.formatMemoryValue(targetLimit)
	
	log.Printf("DEBUG: Formatted memory recommendation: %s (from %.0f bytes)", recommendedRequestStr, targetRequest)
	
	recommendation := models.ResourceRecommendation{
		CurrentRequest:          e.formatMemoryValue(currentRequest),
		CurrentLimit:            e.formatMemoryValue(currentLimit),
		CurrentUsage:            e.formatMemoryValue(currentUsage),
		RecommendedRequest:      &recommendedRequestStr,
		RecommendedLimit:        &recommendedLimitStr,
		CurrentUtilization:      currentUtilization,
		TargetUtilization:       e.config.TargetMemoryUtilization,
		CurrentRequestValue:     currentRequest,
		CurrentLimitValue:       currentLimit,
		CurrentUsageValue:       currentUsage,
		RecommendedRequestValue: targetRequest,
		RecommendedLimitValue:   &targetLimit,
		ResourceChange:          changeType,
		PercentageChange:        percentageChange,
	}
	
	return recommendation, reasons
}

// Helper functions

// parseResourceValue converts resource string to float64 in base units
func (e *RecommendationEngine) parseResourceValue(value string, resourceType string) float64 {
	if value == "" || value == "0" {
		return 0
	}
	
	// Remove whitespace
	value = strings.TrimSpace(value)
	
	if resourceType == "cpu" {
		return e.parseCPUValue(value)
	} else if resourceType == "memory" {
		return e.parseMemoryValue(value)
	}
	
	return 0
}

// parseCPUValue converts CPU string to cores (float64)
func (e *RecommendationEngine) parseCPUValue(cpu string) float64 {
	if cpu == "" || cpu == "0" {
		return 0
	}
	
	// Handle millicores (e.g., "100m")
	if strings.HasSuffix(cpu, "m") {
		milliStr := strings.TrimSuffix(cpu, "m")
		if milli, err := strconv.ParseFloat(milliStr, 64); err == nil {
			return milli / 1000.0 // Convert millicores to cores
		}
	}
	
	// Handle cores directly (e.g., "1.5")
	if cores, err := strconv.ParseFloat(cpu, 64); err == nil {
		return cores
	}
	
	return 0
}

// parseMemoryValue converts memory string to bytes (float64)
func (e *RecommendationEngine) parseMemoryValue(memory string) float64 {
	if memory == "" || memory == "0" {
		return 0
	}
	
	// Regular expression to parse memory values
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*([KMGT]?i?)B?$`)
	matches := re.FindStringSubmatch(strings.ToUpper(memory))
	
	if len(matches) != 3 {
		// Try without 'B' suffix
		re2 := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*([KMGT]?i?)$`)
		matches = re2.FindStringSubmatch(strings.ToUpper(memory))
		if len(matches) != 3 {
			return 0
		}
	}
	
	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0
	}
	
	unit := matches[2]
	multiplier := 1.0
	
	switch unit {
	case "", "B":
		multiplier = 1
	case "K", "KB":
		multiplier = 1000
	case "KI", "KIB":
		multiplier = 1024
	case "M", "MB":
		multiplier = 1000 * 1000
	case "MI", "MIB":
		multiplier = 1024 * 1024
	case "G", "GB":
		multiplier = 1000 * 1000 * 1000
	case "GI", "GIB":
		multiplier = 1024 * 1024 * 1024
	case "T", "TB":
		multiplier = 1000 * 1000 * 1000 * 1000
	case "TI", "TIB":
		multiplier = 1024 * 1024 * 1024 * 1024
	}
	
	return value * multiplier
}

// formatCPUValue converts cores (float64) to CPU string
func (e *RecommendationEngine) formatCPUValue(cores float64) string {
	if cores == 0 {
		return "0m"
	}
	
	// Convert to millicores and format
	millicores := cores * 1000
	if millicores < 1 {
		return fmt.Sprintf("%.1fm", millicores)
	}
	return fmt.Sprintf("%.0fm", millicores)
}

// formatMemoryValue converts bytes (float64) to memory string
func (e *RecommendationEngine) formatMemoryValue(bytes float64) string {
	if bytes == 0 {
		return "0Mi"
	}
	
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	
	if bytes >= GB {
		return fmt.Sprintf("%.1fGi", bytes/GB)
	} else if bytes >= MB {
		return fmt.Sprintf("%.0fMi", bytes/MB)
	} else if bytes >= KB {
		return fmt.Sprintf("%.0fKi", bytes/KB)
	}
	return fmt.Sprintf("%.0fB", bytes)
}

// calculateConfidenceScore determines confidence in the recommendation
func (e *RecommendationEngine) calculateConfidenceScore(historical k8s.HistoricalMetrics) float64 {
	score := 100.0
	
	// Reduce confidence based on data sparsity
	cpuDataPoints := len(historical.CPU.Usage)
	memoryDataPoints := len(historical.Memory.Usage)
	
	if cpuDataPoints < e.config.MinDataPointsReq {
		score -= float64(e.config.MinDataPointsReq-cpuDataPoints) * 2
	}
	if memoryDataPoints < e.config.MinDataPointsReq {
		score -= float64(e.config.MinDataPointsReq-memoryDataPoints) * 2
	}
	
	// Reduce confidence for high variance
	if historical.CPU.Peak > 0 && historical.CPU.Average > 0 {
		cpuVariance := (historical.CPU.Peak - historical.CPU.Average) / historical.CPU.Average
		if cpuVariance > 2.0 { // More than 200% variance
			score -= 20
		}
	}
	
	if historical.Memory.Peak > 0 && historical.Memory.Average > 0 {
		memoryVariance := (historical.Memory.Peak - historical.Memory.Average) / historical.Memory.Average
		if memoryVariance > 1.0 { // More than 100% variance
			score -= 15
		}
	}
	
	// Ensure score is within bounds
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	
	return score
}

// assessDataQuality determines the quality of historical data
func (e *RecommendationEngine) assessDataQuality(historical k8s.HistoricalMetrics) string {
	cpuDataPoints := len(historical.CPU.Usage)
	memoryDataPoints := len(historical.Memory.Usage)
	
	minPoints := cpuDataPoints
	if memoryDataPoints < minPoints {
		minPoints = memoryDataPoints
	}
	
	expectedPoints := e.config.AnalysisDays * 24 * 4 // Assuming 15min intervals
	coverage := float64(minPoints) / float64(expectedPoints)
	
	if coverage >= 0.8 {
		return "excellent"
	} else if coverage >= 0.6 {
		return "good"
	} else if coverage >= 0.4 {
		return "fair"
	}
	return "poor"
}

// calculatePriority determines the priority level of the recommendation
func (e *RecommendationEngine) calculatePriority(reasons []models.RecommendationReason, confidence float64) string {
	highPriorityReasons := []models.RecommendationReason{
		models.ReasonOverUtilized,
		models.ReasonMissingRequests,
		models.ReasonCPULimitPresent,
	}
	
	mediumPriorityReasons := []models.RecommendationReason{
		models.ReasonUnderUtilized,
		models.ReasonMemoryMisaligned,
		models.ReasonMissingLimits,
	}
	
	hasHighPriority := false
	hasMediumPriority := false
	
	for _, reason := range reasons {
		for _, highReason := range highPriorityReasons {
			if reason == highReason {
				hasHighPriority = true
				break
			}
		}
		for _, mediumReason := range mediumPriorityReasons {
			if reason == mediumReason {
				hasMediumPriority = true
				break
			}
		}
	}
	
	// High confidence boosts priority
	if hasHighPriority && confidence >= 70 {
		return "high"
	} else if (hasHighPriority || hasMediumPriority) && confidence >= 50 {
		return "medium"
	}
	
	return "low"
}

// assessRiskLevel determines the risk of applying recommendations
func (e *RecommendationEngine) assessRiskLevel(cpu, memory models.ResourceRecommendation) string {
	// High risk if significant increases
	if cpu.PercentageChange > 100 || memory.PercentageChange > 100 {
		return "high"
	}
	
	// High risk if removing limits on heavily utilized resources
	if cpu.ResourceChange == "remove_limit" && cpu.CurrentUtilization > 80 {
		return "high"
	}
	
	// Medium risk for moderate changes
	if math.Abs(cpu.PercentageChange) > 50 || math.Abs(memory.PercentageChange) > 50 {
		return "medium"
	}
	
	return "low"
}

// calculateEstimatedSavings estimates potential cost/resource savings
func (e *RecommendationEngine) calculateEstimatedSavings(cpu, memory models.ResourceRecommendation) string {
	// This is a simplified calculation - in reality, you'd want to factor in
	// actual cloud provider pricing
	
	cpuChange := cpu.RecommendedRequestValue - cpu.CurrentRequestValue
	memoryChange := memory.RecommendedRequestValue - memory.CurrentRequestValue
	
	if cpuChange < 0 && memoryChange < 0 {
		return "high" // Both resources decreasing
	} else if cpuChange < 0 || memoryChange < 0 {
		return "medium" // One resource decreasing
	} else if math.Abs(cpuChange) < 0.1 && math.Abs(memoryChange) < 100*1024*1024 {
		return "none" // Minimal changes
	}
	
	return "low"
}

// generateSummary creates an aggregate summary of all recommendations
func (e *RecommendationEngine) generateSummary(recommendations []models.PodResourceRecommendation) models.RecommendationsSummary {
	summary := models.RecommendationsSummary{
		TotalPodsAnalyzed: len(recommendations),
	}
	
	var totalCPUIncrease, totalCPUDecrease, totalMemoryChange float64
	
	for _, rec := range recommendations {
		// Count optimization needs
		needsOptimization := false
		for _, reason := range rec.Reasons {
			if reason != models.ReasonWellOptimized {
				needsOptimization = true
				break
			}
		}
		
		if needsOptimization {
			summary.PodsNeedingOptimization++
		} else {
			summary.PodsWellOptimized++
		}
		
		// Count by priority
		switch rec.Priority {
		case "high":
			summary.HighPriorityRecommendations++
		case "medium":
			summary.MediumPriorityRecommendations++
		case "low":
			summary.LowPriorityRecommendations++
		}
		
		// Count by category
		for _, reason := range rec.Reasons {
			switch reason {
			case models.ReasonOverUtilized:
				summary.OverUtilizedPods++
			case models.ReasonUnderUtilized:
				summary.UnderUtilizedPods++
			case models.ReasonCPULimitPresent:
				summary.PodsWithCPULimits++
			case models.ReasonMemoryMisaligned:
				summary.MemoryMisalignedPods++
			case models.ReasonMissingRequests:
				summary.PodsMissingRequests++
			}
		}
		
		// Calculate resource changes
		cpuChange := rec.CPU.RecommendedRequestValue - rec.CPU.CurrentRequestValue
		memoryChange := rec.Memory.RecommendedRequestValue - rec.Memory.CurrentRequestValue
		
		if cpuChange > 0 {
			totalCPUIncrease += cpuChange
		} else {
			totalCPUDecrease += math.Abs(cpuChange)
		}
		
		totalMemoryChange += memoryChange // Can be positive or negative
	}
	
	// Format totals
	summary.TotalCPURequestIncrease = e.formatCPUValue(totalCPUIncrease)
	summary.TotalCPURequestDecrease = e.formatCPUValue(totalCPUDecrease)
	if totalMemoryChange >= 0 {
		summary.TotalMemoryRequestChange = "+" + e.formatMemoryValue(totalMemoryChange)
	} else {
		summary.TotalMemoryRequestChange = e.formatMemoryValue(totalMemoryChange)
	}
	
	// Estimate savings
	if totalCPUDecrease > totalCPUIncrease && totalMemoryChange < 0 {
		summary.EstimatedCostSavings = "high"
	} else if totalCPUDecrease > 0 || totalMemoryChange < 0 {
		summary.EstimatedCostSavings = "medium"
	} else {
		summary.EstimatedCostSavings = "low"
	}
	
	summary.EstimatedResourceSavings = summary.EstimatedCostSavings
	
	return summary
}
