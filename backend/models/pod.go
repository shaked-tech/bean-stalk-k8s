package models

import (
	"time"
)

// PodMetrics represents resource usage and limits for a single pod
type PodMetrics struct {
	Name          string            `json:"name"`
	Namespace     string            `json:"namespace"`
	ContainerName string            `json:"containerName"`
	CPU           ResourceMetrics   `json:"cpu"`
	Memory        ResourceMetrics   `json:"memory"`
	Labels        map[string]string `json:"labels,omitempty"`
}

// ResourceMetrics represents resource usage, requests, and limits
type ResourceMetrics struct {
	Usage      string  `json:"usage"`
	Request    string  `json:"request"`
	Limit      string  `json:"limit"`
	UsageValue float64 `json:"usageValue"`
	RequestValue float64 `json:"requestValue"`
	LimitValue float64 `json:"limitValue"`
	// Percentage of request that's being used (usage/request * 100)
	RequestPercentage float64 `json:"requestPercentage"`
	// Percentage of limit that's being used (usage/limit * 100)
	LimitPercentage float64 `json:"limitPercentage,omitempty"`
}

// NamespaceList represents a list of available namespaces
type NamespaceList struct {
	Namespaces []string `json:"namespaces"`
}

// PodMetricsList represents a list of pod metrics
type PodMetricsList struct {
	Pods []PodMetrics `json:"pods"`
}

// TimeRange represents a time range for historical data
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// DataPoint represents a single metric data point
type DataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
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

// UsagePatterns identifies usage patterns
type UsagePatterns struct {
	PeakHours       []int   `json:"peakHours"`       // Hours of day with peak usage
	LowUsageHours   []int   `json:"lowUsageHours"`   // Hours of day with low usage
	DailyVariation  float64 `json:"dailyVariation"`  // Coefficient of variation across days
	WeeklyVariation float64 `json:"weeklyVariation"` // Variation across week
}

// ResourceWasteAnalysis identifies over/under-provisioned resources
type ResourceWasteAnalysis struct {
	CPUOverProvisioned     bool    `json:"cpuOverProvisioned"`
	MemoryOverProvisioned  bool    `json:"memoryOverProvisioned"`
	CPUUnderProvisioned    bool    `json:"cpuUnderProvisioned"`
	MemoryUnderProvisioned bool    `json:"memoryUnderProvisioned"`
	CPUWastePercentage     float64 `json:"cpuWastePercentage"`
	MemoryWastePercentage  float64 `json:"memoryWastePercentage"`
}

// UsageAnalysis provides insights about resource usage patterns
type UsageAnalysis struct {
	CPUEfficiency     float64               `json:"cpuEfficiency"`     // Average usage/request ratio
	MemoryEfficiency  float64               `json:"memoryEfficiency"`  // Average usage/request ratio
	ResourceWaste     ResourceWasteAnalysis `json:"resourceWaste"`
	Recommendations   []string              `json:"recommendations"`
	Patterns          UsagePatterns         `json:"patterns"`
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

// HistoricalAnalysisList represents the response for historical analysis
type HistoricalAnalysisList struct {
	HistoricalMetrics []HistoricalMetrics `json:"historicalMetrics"`
	GeneratedAt       time.Time           `json:"generatedAt"`
	TimeRange         TimeRange           `json:"timeRange"`
	Summary           AnalysisSummary     `json:"summary"`
}

// AnalysisSummary provides aggregate insights across all analyzed pods
type AnalysisSummary struct {
	TotalPodsAnalyzed        int     `json:"totalPodsAnalyzed"`
	OverProvisionedPods      int     `json:"overProvisionedPods"`
	UnderProvisionedPods     int     `json:"underProvisionedPods"`
	WellOptimizedPods        int     `json:"wellOptimizedPods"`
	AverageEfficiency        float64 `json:"averageEfficiency"`
	TotalRecommendations     int     `json:"totalRecommendations"`
	MostCommonRecommendation string  `json:"mostCommonRecommendation"`
}

// PodTrendAnalysis represents detailed trend analysis for a specific pod
type PodTrendAnalysis struct {
	PodName      string              `json:"podName"`
	Namespace    string              `json:"namespace"`
	Containers   []HistoricalMetrics `json:"containers"`
	DaysAnalyzed int                 `json:"daysAnalyzed"`
	GeneratedAt  time.Time           `json:"generatedAt"`
	Summary      PodTrendSummary     `json:"summary"`
}

// PodTrendSummary provides summary insights for pod trend analysis
type PodTrendSummary struct {
	OverallTrend            string    `json:"overallTrend"`
	ResourceRecommendations []string  `json:"resourceRecommendations"`
	RiskLevel               string    `json:"riskLevel"` // low, medium, high
	NextReviewDate          time.Time `json:"nextReviewDate"`
}
