package models

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
