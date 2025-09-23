package k8s

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bean-stalk-k8s/backend/models"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// GetPodMetrics returns metrics for all pods in the specified namespace
// If namespace is empty, it returns metrics for pods in all namespaces
func (c *Client) GetPodMetrics(ctx context.Context, namespace string) ([]models.PodMetrics, error) {
	// Get pods
	var podList *corev1.PodList
	var err error

	if namespace == "" {
		podList, err = c.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	} else {
		podList, err = c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}

	// Initialize with empty slice to ensure we never return nil
	podMetricsList := make([]models.PodMetrics, 0)

	// Create a map to store pod resource requests and limits
	podResources := make(map[string]map[string]map[string]resource.Quantity)

	// Populate pod resources map
	for _, pod := range podList.Items {
		if pod.Status.Phase != corev1.PodRunning {
			continue
		}

		podKey := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
		podResources[podKey] = make(map[string]map[string]resource.Quantity)

		for _, container := range pod.Spec.Containers {
			containerResources := make(map[string]resource.Quantity)

			// Get requests
			if container.Resources.Requests != nil {
				if cpuRequest, ok := container.Resources.Requests[corev1.ResourceCPU]; ok {
					containerResources["cpuRequest"] = cpuRequest
				}
				if memoryRequest, ok := container.Resources.Requests[corev1.ResourceMemory]; ok {
					containerResources["memoryRequest"] = memoryRequest
				}
			}

			// Get limits
			if container.Resources.Limits != nil {
				if cpuLimit, ok := container.Resources.Limits[corev1.ResourceCPU]; ok {
					containerResources["cpuLimit"] = cpuLimit
				}
				if memoryLimit, ok := container.Resources.Limits[corev1.ResourceMemory]; ok {
					containerResources["memoryLimit"] = memoryLimit
				}
			}

			podResources[podKey][container.Name] = containerResources
		}
	}

	// Get metrics for all namespaces or specific namespace
	var metricsItems []metav1.ObjectMeta
	var containerMetrics map[string]map[string]map[string]resource.Quantity

	if namespace == "" {
		metrics, err := c.metricsClient.MetricsV1beta1().PodMetricses("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get pod metrics: %v", err)
		}

		// Process metrics
		containerMetrics = make(map[string]map[string]map[string]resource.Quantity)
		for _, m := range metrics.Items {
			podKey := fmt.Sprintf("%s/%s", m.Namespace, m.Name)
			containerMetrics[podKey] = make(map[string]map[string]resource.Quantity)

			for _, container := range m.Containers {
				containerMetrics[podKey][container.Name] = map[string]resource.Quantity{
					"cpu":    container.Usage[corev1.ResourceCPU],
					"memory": container.Usage[corev1.ResourceMemory],
				}
			}
		}

		// Get all pods' metadata
		for _, pod := range podList.Items {
			if pod.Status.Phase == corev1.PodRunning {
				metricsItems = append(metricsItems, pod.ObjectMeta)
			}
		}
	} else {
		metrics, err := c.metricsClient.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get pod metrics: %v", err)
		}

		// Process metrics
		containerMetrics = make(map[string]map[string]map[string]resource.Quantity)
		for _, m := range metrics.Items {
			podKey := fmt.Sprintf("%s/%s", m.Namespace, m.Name)
			containerMetrics[podKey] = make(map[string]map[string]resource.Quantity)

			for _, container := range m.Containers {
				containerMetrics[podKey][container.Name] = map[string]resource.Quantity{
					"cpu":    container.Usage[corev1.ResourceCPU],
					"memory": container.Usage[corev1.ResourceMemory],
				}
			}
		}

		// Get specific namespace pods' metadata
		for _, pod := range podList.Items {
			if pod.Status.Phase == corev1.PodRunning && pod.Namespace == namespace {
				metricsItems = append(metricsItems, pod.ObjectMeta)
			}
		}
	}

	// Combine metrics with resource requests and limits
	for _, podMeta := range metricsItems {
		podKey := fmt.Sprintf("%s/%s", podMeta.Namespace, podMeta.Name)
		
		// Skip if we don't have metrics for this pod
		if _, ok := containerMetrics[podKey]; !ok {
			continue
		}

		// Process each container
		for containerName, metrics := range containerMetrics[podKey] {
			// Skip if we don't have resource info for this container
			if _, ok := podResources[podKey][containerName]; !ok {
				continue
			}

			cpuUsage := metrics["cpu"]
			memoryUsage := metrics["memory"]
			
			// Check if resources are set
			containerRes := podResources[podKey][containerName]
			_, hasCPURequest := containerRes["cpuRequest"]
			_, hasCPULimit := containerRes["cpuLimit"]
			_, hasMemoryRequest := containerRes["memoryRequest"]
			_, hasMemoryLimit := containerRes["memoryLimit"]

			podMetric := models.PodMetrics{
				Name:          podMeta.Name,
				Namespace:     podMeta.Namespace,
				ContainerName: containerName,
				Labels:        podMeta.Labels,
				CPU:           createResourceMetrics(cpuUsage, containerRes["cpuRequest"], containerRes["cpuLimit"], true, hasCPURequest, hasCPULimit),
				Memory:        createResourceMetrics(memoryUsage, containerRes["memoryRequest"], containerRes["memoryLimit"], false, hasMemoryRequest, hasMemoryLimit),
			}

			podMetricsList = append(podMetricsList, podMetric)
		}
	}

	return podMetricsList, nil
}

// createResourceMetrics creates a ResourceMetrics struct from usage, request, and limit values
func createResourceMetrics(usage, request, limit resource.Quantity, isCPU bool, hasRequest, hasLimit bool) models.ResourceMetrics {
	// Convert values for calculations
	usageValue, _ := strconv.ParseFloat(parseQuantity(usage, isCPU), 64)
	requestValue, _ := strconv.ParseFloat(parseQuantity(request, isCPU), 64)
	limitValue, _ := strconv.ParseFloat(parseQuantity(limit, isCPU), 64)

	// Format string values consistently (use Mi for memory, cores for CPU)
	var usageStr, requestStr, limitStr string
	if isCPU {
		usageStr = FormatResourceValue(usageValue, true)
		
		// Use '-' if request/limit is not set, otherwise format the value
		if hasRequest {
			requestStr = FormatResourceValue(requestValue, true)
		} else {
			requestStr = "-"
			requestValue = 0 // Ensure value is 0 for unset resources
		}
		
		if hasLimit {
			limitStr = FormatResourceValue(limitValue, true)
		} else {
			limitStr = "-"
			limitValue = 0 // Ensure value is 0 for unset resources
		}
	} else {
		// For memory, always use Mi format
		usageStr = fmt.Sprintf("%.2f Mi", usageValue)
		
		if hasRequest {
			requestStr = fmt.Sprintf("%.2f Mi", requestValue)
		} else {
			requestStr = "-"
			requestValue = 0
		}
		
		if hasLimit {
			limitStr = fmt.Sprintf("%.2f Mi", limitValue)
		} else {
			limitStr = "-"
			limitValue = 0
		}
	}

	result := models.ResourceMetrics{
		Usage:       usageStr,
		Request:     requestStr,
		Limit:       limitStr,
		UsageValue:  usageValue,
		RequestValue: requestValue,
		LimitValue:  limitValue,
	}

	// Calculate percentages only if request/limit are set
	if hasRequest && requestValue > 0 {
		result.RequestPercentage = (usageValue / requestValue) * 100
	}

	if hasLimit && limitValue > 0 {
		result.LimitPercentage = (usageValue / limitValue) * 100
	}

	return result
}

// parseQuantity converts a resource.Quantity to a string representation of its value
// For CPU, it returns the value in cores
// For memory, it returns the value in Mi
func parseQuantity(quantity resource.Quantity, isCPU bool) string {
	if quantity.IsZero() {
		return "0"
	}

	if isCPU {
		// CPU values can be in different formats (cores, millicores, nanocores)
		// Get the value in nanocores first
		nanoValue := quantity.ScaledValue(resource.Nano)
		
		// Convert nanocores to cores
		cores := float64(nanoValue) / 1000000000.0
		
		return fmt.Sprintf("%.3f", cores)
	} else {
		// Memory - convert to Mi
		bytes := quantity.Value()
		if bytes == 0 {
			return "0"
		}
		
		mi := float64(bytes) / (1024 * 1024)
		return fmt.Sprintf("%.2f", mi)
	}
}

// FormatResourceValue formats a resource value for display
func FormatResourceValue(value float64, isCPU bool) string {
	if isCPU {
		// Convert cores to millicores
		millicores := value * 1000
		
		// For very small values, show minimum of 0.1 m
		if millicores < 0.1 && millicores > 0 {
			return "0.1 m"
		}
		
		// For values less than 1000 millicores, show as millicores with 1 decimal
		if millicores < 1000 {
			return fmt.Sprintf("%.1f m", millicores)
		}
		
		// For larger values, show as cores with 1 decimal
		return fmt.Sprintf("%.1f", value)
	} else {
		// Memory
		if value > 1024 {
			return fmt.Sprintf("%.2f Gi", value/1024)
		}
		return fmt.Sprintf("%.2f Mi", value)
	}
}
