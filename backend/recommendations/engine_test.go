package recommendations

import (
	"testing"

	"github.com/bean-stalk-k8s/backend/k8s"
	"github.com/bean-stalk-k8s/backend/models"
)

// Helper function to create test config
func getTestConfig() models.RecommendationConfig {
	return models.RecommendationConfig{
		CPUBufferPercentage:      30.0,
		MemoryBufferPercentage:   20.0,
		MinCPURequest:           "10m",
		MaxCPURequest:           "4000m",
		MinMemoryRequest:        "16Mi",
		MaxMemoryRequest:        "8Gi",
		MaxScaleUpFactor:        3.0,
		MaxScaleDownFactor:      0.3,
		RemoveCPULimits:         true,
		AlignMemoryRequestsLimits: true,
	}
}

// TestGuaranteedQoS tests pods with Guaranteed QoS (requests = limits)
func TestGuaranteedQoS(t *testing.T) {
	engine := NewRecommendationEngine(getTestConfig())

	// Scenario: Well-utilized Guaranteed pod
	// CPU: 100m usage, 150m request/limit (66% utilization)
	// Memory: 256Mi usage, 384Mi request/limit (66% utilization)
	metric := k8s.PodMetric{
		Name:         "test-guaranteed-pod",
		Namespace:    "default",
		ContainerName: "app",
		CPUUsage:     0.100,  // 100m in cores
		CPURequest:   0.150,  // 150m in cores
		CPULimit:     0.150,  // 150m in cores
		MemoryUsage:  268435456, // 256Mi in bytes
		MemoryRequest: 402653184, // 384Mi in bytes
		MemoryLimit:  402653184, // 384Mi in bytes
	}

	rec, err := engine.generatePodRecommendation(metric)
	if err != nil {
		t.Fatalf("Failed to generate recommendation: %v", err)
	}

	// CPU: 100m * 1.30 = 130m
	expectedCPU := 0.130
	if rec.CPU.RecommendedRequestValue < metric.CPUUsage {
		t.Errorf("CPU recommendation (%.4f) is less than usage (%.4f)",
			rec.CPU.RecommendedRequestValue, metric.CPUUsage)
	}
	if rec.CPU.RecommendedRequestValue != expectedCPU {
		t.Logf("WARNING: Expected CPU %.4f, got %.4f", expectedCPU, rec.CPU.RecommendedRequestValue)
	}

	// Memory: 256Mi * 1.20 = 307Mi (322122547 bytes)
	if rec.Memory.RecommendedRequestValue < metric.MemoryUsage {
		t.Errorf("Memory recommendation (%.0f) is less than usage (%.0f)",
			rec.Memory.RecommendedRequestValue, metric.MemoryUsage)
	}

	t.Logf("Guaranteed QoS Test:")
	t.Logf("  CPU: %.4f cores -> %s (%.1f%% change)",
		metric.CPUUsage, *rec.CPU.RecommendedRequest, rec.CPU.PercentageChange)
	t.Logf("  Memory: %.0f bytes -> %s (%.1f%% change)",
		metric.MemoryUsage, *rec.Memory.RecommendedRequest, rec.Memory.PercentageChange)
}

// TestBestEffortQoS tests pods with Best-Effort QoS (no requests or limits)
func TestBestEffortQoS(t *testing.T) {
	engine := NewRecommendationEngine(getTestConfig())

	// Scenario: Best-Effort pod with actual usage
	// No requests or limits, but has usage
	// CPU: 50m usage, 0 request/limit
	// Memory: 128Mi usage, 0 request/limit
	metric := k8s.PodMetric{
		Name:         "test-besteffort-pod",
		Namespace:    "default",
		ContainerName: "app",
		CPUUsage:     0.050,  // 50m in cores
		CPURequest:   0,
		CPULimit:     0,
		MemoryUsage:  134217728, // 128Mi in bytes
		MemoryRequest: 0,
		MemoryLimit:  0,
	}

	rec, err := engine.generatePodRecommendation(metric)
	if err != nil {
		t.Fatalf("Failed to generate recommendation: %v", err)
	}

	// CPU: 50m * 1.30 = 65m
	if rec.CPU.RecommendedRequestValue < metric.CPUUsage {
		t.Errorf("CPU recommendation (%.4f) is less than usage (%.4f)",
			rec.CPU.RecommendedRequestValue, metric.CPUUsage)
	}
	if rec.CPU.RecommendedRequestValue == 0 {
		t.Errorf("CPU recommendation should not be 0 when usage is %.4f", metric.CPUUsage)
	}

	// Memory: 128Mi * 1.20 = 154Mi (161480089 bytes)
	if rec.Memory.RecommendedRequestValue < metric.MemoryUsage {
		t.Errorf("Memory recommendation (%.0f) is less than usage (%.0f)",
			rec.Memory.RecommendedRequestValue, metric.MemoryUsage)
	}
	if rec.Memory.RecommendedRequestValue == 0 {
		t.Errorf("Memory recommendation should not be 0 when usage is %.0f", metric.MemoryUsage)
	}

	// Check reasons include missing requests
	hasMissingRequests := false
	for _, reason := range rec.Reasons {
		if reason == models.ReasonMissingRequests {
			hasMissingRequests = true
			break
		}
	}
	if !hasMissingRequests {
		t.Error("Expected ReasonMissingRequests for Best-Effort pod")
	}

	t.Logf("Best-Effort QoS Test:")
	t.Logf("  CPU: %.4f cores -> %s", metric.CPUUsage, *rec.CPU.RecommendedRequest)
	t.Logf("  Memory: %.0f bytes -> %s", metric.MemoryUsage, *rec.Memory.RecommendedRequest)
}

// TestBurstableQoS tests pods with Burstable QoS (requests != limits or missing one)
func TestBurstableQoS(t *testing.T) {
	engine := NewRecommendationEngine(getTestConfig())

	// Scenario 1: Burstable with memory under-provisioned
	// CPU: 75m usage, 100m request, 200m limit
	// Memory: 512Mi usage, 256Mi request, 1Gi limit (over-utilized!)
	metric := k8s.PodMetric{
		Name:         "test-burstable-pod",
		Namespace:    "default",
		ContainerName: "app",
		CPUUsage:     0.075,  // 75m in cores
		CPURequest:   0.100,  // 100m in cores
		CPULimit:     0.200,  // 200m in cores
		MemoryUsage:  536870912, // 512Mi in bytes
		MemoryRequest: 268435456, // 256Mi in bytes
		MemoryLimit:  1073741824, // 1Gi in bytes
	}

	rec, err := engine.generatePodRecommendation(metric)
	if err != nil {
		t.Fatalf("Failed to generate recommendation: %v", err)
	}

	// CPU: 75m * 1.30 = 97.5m (should scale down from 100m to 97.5m)
	if rec.CPU.RecommendedRequestValue < metric.CPUUsage {
		t.Errorf("CPU recommendation (%.4f) is less than usage (%.4f)",
			rec.CPU.RecommendedRequestValue, metric.CPUUsage)
	}

	// Memory: 512Mi * 1.20 = 614Mi
	// BUT current request is 256Mi, max scale up is 256Mi * 3 = 768Mi
	// So should recommend 614Mi (within scale limits)
	if rec.Memory.RecommendedRequestValue < metric.MemoryUsage {
		t.Errorf("Memory recommendation (%.0f) is less than usage (%.0f)",
			rec.Memory.RecommendedRequestValue, metric.MemoryUsage)
	}

	// Check for memory over-utilization
	hasMemoryOverUtilized := false
	for _, reason := range rec.Reasons {
		if reason == models.ReasonMemoryOverUtilized {
			hasMemoryOverUtilized = true
			break
		}
	}
	if !hasMemoryOverUtilized {
		t.Error("Expected ReasonMemoryOverUtilized for over-utilized memory")
	}

	t.Logf("Burstable QoS Test (under-provisioned):")
	t.Logf("  CPU: %.4f cores -> %s (%.1f%% change)",
		metric.CPUUsage, *rec.CPU.RecommendedRequest, rec.CPU.PercentageChange)
	t.Logf("  Memory: %.0f bytes -> %s (%.1f%% change)",
		metric.MemoryUsage, *rec.Memory.RecommendedRequest, rec.Memory.PercentageChange)
}

// TestOverProvisionedPod tests the original bug scenario
func TestOverProvisionedPod(t *testing.T) {
	engine := NewRecommendationEngine(getTestConfig())

	// Scenario: Heavily over-provisioned pod (the bug scenario)
	// CPU: 50m usage, 500m request, 1000m limit (10% utilization)
	// Memory: 100Mi usage, 300Mi request, 600Mi limit (33% utilization)
	metric := k8s.PodMetric{
		Name:         "test-overprovisioned-pod",
		Namespace:    "default",
		ContainerName: "app",
		CPUUsage:     0.050,  // 50m in cores
		CPURequest:   0.500,  // 500m in cores
		CPULimit:     1.000,  // 1000m in cores
		MemoryUsage:  104857600, // 100Mi in bytes
		MemoryRequest: 314572800, // 300Mi in bytes
		MemoryLimit:  629145600, // 600Mi in bytes
	}

	rec, err := engine.generatePodRecommendation(metric)
	if err != nil {
		t.Fatalf("Failed to generate recommendation: %v", err)
	}

	// This is the critical test - recommendations must NEVER be below usage
	if rec.CPU.RecommendedRequestValue < metric.CPUUsage {
		t.Errorf("BUG FOUND! CPU recommendation (%.4f) is less than usage (%.4f)",
			rec.CPU.RecommendedRequestValue, metric.CPUUsage)
	}

	if rec.Memory.RecommendedRequestValue < metric.MemoryUsage {
		t.Errorf("BUG FOUND! Memory recommendation (%.0f) is less than usage (%.0f)",
			rec.Memory.RecommendedRequestValue, metric.MemoryUsage)
	}

	// With buffer calculation:
	// CPU: 50m * 1.30 = 65m
	// Memory: 100Mi * 1.20 = 120Mi
	// These should be the recommendations (or higher due to scale-down limits)

	t.Logf("Over-Provisioned Pod Test:")
	t.Logf("  CPU: %.4f cores (request: %.4f) -> %s (%.1f%% change)",
		metric.CPUUsage, metric.CPURequest, *rec.CPU.RecommendedRequest, rec.CPU.PercentageChange)
	t.Logf("  Memory: %.0f bytes (request: %.0f) -> %s (%.1f%% change)",
		metric.MemoryUsage, metric.MemoryRequest, *rec.Memory.RecommendedRequest, rec.Memory.PercentageChange)
}

// TestZeroUsagePod tests pods with zero usage
func TestZeroUsagePod(t *testing.T) {
	engine := NewRecommendationEngine(getTestConfig())

	// Scenario: Pod with requests but zero usage (idle or just started)
	metric := k8s.PodMetric{
		Name:         "test-zero-usage-pod",
		Namespace:    "default",
		ContainerName: "app",
		CPUUsage:     0,
		CPURequest:   0.100,
		CPULimit:     0.200,
		MemoryUsage:  0,
		MemoryRequest: 268435456, // 256Mi
		MemoryLimit:  536870912, // 512Mi
	}

	rec, err := engine.generatePodRecommendation(metric)
	if err != nil {
		t.Fatalf("Failed to generate recommendation: %v", err)
	}

	// With zero usage, should default to minimum configured values
	// CPU: MinCPURequest = 10m = 0.010 cores
	// Memory: MinMemoryRequest = 16Mi = 16777216 bytes

	if rec.CPU.RecommendedRequestValue == 0 {
		t.Error("CPU recommendation should not be 0, should use minimum")
	}

	if rec.Memory.RecommendedRequestValue == 0 {
		t.Error("Memory recommendation should not be 0, should use minimum")
	}

	t.Logf("Zero Usage Pod Test:")
	t.Logf("  CPU: 0 cores -> %s", *rec.CPU.RecommendedRequest)
	t.Logf("  Memory: 0 bytes -> %s", *rec.Memory.RecommendedRequest)
}

// TestScalingLimits tests that scaling limits are respected
func TestScalingLimits(t *testing.T) {
	engine := NewRecommendationEngine(getTestConfig())

	// Scenario: Pod that needs massive scale-up
	// Usage is 100m CPU but current request is only 10m (1000% usage!)
	// Scale-up limit is 3x, so max allowed is 30m
	metric := k8s.PodMetric{
		Name:         "test-scale-limit-pod",
		Namespace:    "default",
		ContainerName: "app",
		CPUUsage:     0.100,  // 100m
		CPURequest:   0.010,  // 10m (extremely under-provisioned)
		CPULimit:     0.020,  // 20m
		MemoryUsage:  536870912, // 512Mi
		MemoryRequest: 67108864,  // 64Mi (extremely under-provisioned)
		MemoryLimit:  134217728, // 128Mi
	}

	rec, err := engine.generatePodRecommendation(metric)
	if err != nil {
		t.Fatalf("Failed to generate recommendation: %v", err)
	}

	// CPU: 100m * 1.30 = 130m, but max allowed is 10m * 3 = 30m
	// Should be capped at 30m
	maxAllowedCPU := metric.CPURequest * 3.0
	if rec.CPU.RecommendedRequestValue > maxAllowedCPU {
		t.Errorf("CPU recommendation (%.4f) exceeds max allowed (%.4f)",
			rec.CPU.RecommendedRequestValue, maxAllowedCPU)
	}

	// Memory: 512Mi * 1.20 = 614Mi, but max allowed is 64Mi * 3 = 192Mi
	// Should be capped at 192Mi
	maxAllowedMemory := metric.MemoryRequest * 3.0
	if rec.Memory.RecommendedRequestValue > maxAllowedMemory {
		t.Errorf("Memory recommendation (%.0f) exceeds max allowed (%.0f)",
			rec.Memory.RecommendedRequestValue, maxAllowedMemory)
	}

	t.Logf("Scaling Limits Test:")
	t.Logf("  CPU: %.4f cores (request: %.4f, max allowed: %.4f) -> %s",
		metric.CPUUsage, metric.CPURequest, maxAllowedCPU, *rec.CPU.RecommendedRequest)
	t.Logf("  Memory: %.0f bytes (request: %.0f, max allowed: %.0f) -> %s",
		metric.MemoryUsage, metric.MemoryRequest, maxAllowedMemory, *rec.Memory.RecommendedRequest)
}
