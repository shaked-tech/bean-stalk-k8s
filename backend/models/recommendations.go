package models

import (
	"time"
)

// RecommendationReason explains why a recommendation was made
type RecommendationReason string

const (
	// Generic utilization reasons (kept for backward compatibility)
	ReasonOverUtilized     RecommendationReason = "over_utilized"
	ReasonUnderUtilized    RecommendationReason = "under_utilized"
	
	// Specific CPU utilization reasons
	ReasonCPUOverUtilized  RecommendationReason = "cpu_over_utilized"
	ReasonCPUUnderUtilized RecommendationReason = "cpu_under_utilized"
	
	// Specific Memory utilization reasons
	ReasonMemoryOverUtilized  RecommendationReason = "memory_over_utilized"
	ReasonMemoryUnderUtilized RecommendationReason = "memory_under_utilized"
	
	// Other reasons
	ReasonWellOptimized    RecommendationReason = "well_optimized"
	ReasonMissingRequests  RecommendationReason = "missing_requests"
	ReasonMissingLimits    RecommendationReason = "missing_limits"
	ReasonCPULimitPresent  RecommendationReason = "cpu_limit_present"
	ReasonMemoryMisaligned RecommendationReason = "memory_misaligned"
)

// ResourceRecommendation contains recommended resource settings
type ResourceRecommendation struct {
	// Current values
	CurrentRequest string  `json:"currentRequest"`
	CurrentLimit   string  `json:"currentLimit"`
	CurrentUsage   string  `json:"currentUsage"`
	
	// Recommended values
	RecommendedRequest *string `json:"recommendedRequest"`
	RecommendedLimit   *string `json:"recommendedLimit,omitempty"`
	
	// Analysis
	CurrentUtilization float64 `json:"currentUtilization"`
	TargetUtilization  float64 `json:"targetUtilization"`
	
	// Numerical values for calculations (in cores for CPU, bytes for memory)
	CurrentRequestValue      float64 `json:"currentRequestValue"`
	CurrentLimitValue        float64 `json:"currentLimitValue"`
	CurrentUsageValue        float64 `json:"currentUsageValue"`
	RecommendedRequestValue  float64 `json:"recommendedRequestValue"`
	RecommendedLimitValue    *float64 `json:"recommendedLimitValue,omitempty"`
	
	// Impact
	ResourceChange   string  `json:"resourceChange"` // "increase", "decrease", "no_change", "remove_limit"
	PercentageChange float64 `json:"percentageChange"`
}

// PodResourceRecommendation contains recommendations for a single pod
type PodResourceRecommendation struct {
	PodName       string                 `json:"podName"`
	Namespace     string                 `json:"namespace"`
	ContainerName string                 `json:"containerName"`
	
	// Resource recommendations
	CPU    ResourceRecommendation `json:"cpu"`
	Memory ResourceRecommendation `json:"memory"`
	
	// Overall analysis
	Reasons          []RecommendationReason `json:"reasons"`
	Priority         string                 `json:"priority"`         // "high", "medium", "low"
	ConfidenceScore  float64                `json:"confidenceScore"`  // 0-100
	EstimatedSavings string                 `json:"estimatedSavings"`
	RiskLevel        string                 `json:"riskLevel"`        // "low", "medium", "high"
	
	// Metadata
	BasedOnDays     int       `json:"basedOnDays"`
	DataQuality     string    `json:"dataQuality"`     // "excellent", "good", "fair", "poor"
	LastUpdated     time.Time `json:"lastUpdated"`
	ApplicableFrom  time.Time `json:"applicableFrom"`
	
	// Historical context
	HistoricalP95CPU    float64 `json:"historicalP95Cpu"`
	HistoricalP95Memory float64 `json:"historicalP95Memory"`
	HistoricalAvgCPU    float64 `json:"historicalAvgCpu"`
	HistoricalAvgMemory float64 `json:"historicalAvgMemory"`
}

// RecommendationsSummary provides aggregate insights
type RecommendationsSummary struct {
	TotalPodsAnalyzed        int     `json:"totalPodsAnalyzed"`
	PodsNeedingOptimization  int     `json:"podsNeedingOptimization"`
	PodsWellOptimized        int     `json:"podsWellOptimized"`
	
	// Resource impact
	TotalCPURequestIncrease  string  `json:"totalCpuRequestIncrease"`
	TotalCPURequestDecrease  string  `json:"totalCpuRequestDecrease"`
	TotalMemoryRequestChange string  `json:"totalMemoryRequestChange"`
	
	// Savings estimates
	EstimatedCostSavings     string  `json:"estimatedCostSavings"`
	EstimatedResourceSavings string  `json:"estimatedResourceSavings"`
	
	// Distribution
	HighPriorityRecommendations    int `json:"highPriorityRecommendations"`
	MediumPriorityRecommendations  int `json:"mediumPriorityRecommendations"`
	LowPriorityRecommendations     int `json:"lowPriorityRecommendations"`
	
	// Categories
	OverUtilizedPods       int `json:"overUtilizedPods"`
	UnderUtilizedPods      int `json:"underUtilizedPods"`
	PodsWithCPULimits      int `json:"podsWithCpuLimits"`
	MemoryMisalignedPods   int `json:"memoryMisalignedPods"`
	PodsMissingRequests    int `json:"podsMissingRequests"`
}

// RecommendationsResponse represents the API response for recommendations
type RecommendationsResponse struct {
	Recommendations []PodResourceRecommendation `json:"recommendations"`
	Summary         RecommendationsSummary      `json:"summary"`
	GeneratedAt     time.Time                   `json:"generatedAt"`
	AnalysisWindow  string                      `json:"analysisWindow"`  // "current usage", "real-time", etc.
	TargetUtilization float64                   `json:"targetUtilization"` // 70.0
}

// RecommendationConfig contains configuration for the recommendation engine
type RecommendationConfig struct {
	TargetCPUUtilization    float64 `json:"targetCpuUtilization"`    // Default: 70%
	TargetMemoryUtilization float64 `json:"targetMemoryUtilization"` // Default: 70%
	
	// Safety bounds
	MinCPURequest       string `json:"minCpuRequest"`       // e.g., "10m"
	MaxCPURequest       string `json:"maxCpuRequest"`       // e.g., "4000m"
	MinMemoryRequest    string `json:"minMemoryRequest"`    // e.g., "64Mi"
	MaxMemoryRequest    string `json:"maxMemoryRequest"`    // e.g., "8Gi"
	
	// Scaling limits
	MaxScaleUpFactor    float64 `json:"maxScaleUpFactor"`    // e.g., 3.0 (300% max increase)
	MaxScaleDownFactor  float64 `json:"maxScaleDownFactor"`  // e.g., 0.3 (70% max decrease)
	
	// Analysis parameters
	AnalysisDays        int     `json:"analysisDays"`        // Default: 7
	MinDataPointsReq    int     `json:"minDataPointsReq"`    // Minimum data points required
	ConfidenceThreshold float64 `json:"confidenceThreshold"` // Minimum confidence for recommendations
	
	// Feature flags
	RemoveCPULimits          bool `json:"removeCpuLimits"`          // Default: true
	AlignMemoryRequestsLimits bool `json:"alignMemoryRequestsLimits"` // Default: true
}

// DefaultRecommendationConfig returns sensible default configuration
func DefaultRecommendationConfig() RecommendationConfig {
	return RecommendationConfig{
		TargetCPUUtilization:     70.0,
		TargetMemoryUtilization:  70.0,
		MinCPURequest:           "10m",
		MaxCPURequest:           "4000m",
		MinMemoryRequest:        "16Mi", // FIXED: Lowered from 64Mi to allow proper downscaling
		MaxMemoryRequest:        "8Gi",
		MaxScaleUpFactor:        3.0,
		MaxScaleDownFactor:      0.3,
		AnalysisDays:           7,
		MinDataPointsReq:       10,
		ConfidenceThreshold:    60.0,
		RemoveCPULimits:        true,
		AlignMemoryRequestsLimits: true,
	}
}
