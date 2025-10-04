# Resource Recommendations API

This document describes the new resource recommendations feature added to the Pod Metrics Dashboard backend.

## Overview

The recommendation system analyzes historical pod resource usage patterns to provide intelligent recommendations for optimizing Kubernetes resource requests and limits. The system implements best practices for Kubernetes resource management:

- **CPU Strategy**: Sets CPU requests to achieve 70% target utilization, removes CPU limits to prevent throttling
- **Memory Strategy**: Aligns memory requests and limits (requests = limits) to achieve 70% target utilization
- **Historical Analysis**: Uses 7-day historical data with P95 percentile for memory and average for CPU

## API Endpoint

### GET /api/pods/recommendations

Returns resource optimization recommendations for pods based on historical usage analysis.

**Query Parameters:**
- `namespace` (optional): Filter recommendations by specific namespace. If not provided, returns recommendations for all namespaces.

**Example Requests:**
```bash
# Get recommendations for all namespaces
curl http://localhost:8080/api/pods/recommendations

# Get recommendations for specific namespace
curl http://localhost:8080/api/pods/recommendations?namespace=production
```

**Response Format:**
```json
{
  "recommendations": [
    {
      "podName": "web-app-7d4b8b5f4c-xyz12",
      "namespace": "production",
      "containerName": "web-app",
      "cpu": {
        "currentRequest": "100m",
        "currentLimit": "200m",
        "currentUsage": "45m",
        "recommendedRequest": "64m",
        "recommendedLimit": null,
        "currentUtilization": 45.0,
        "targetUtilization": 70.0,
        "resourceChange": "remove_limit",
        "percentageChange": -36.0
      },
      "memory": {
        "currentRequest": "256Mi",
        "currentLimit": "512Mi",
        "currentUsage": "180Mi",
        "recommendedRequest": "257Mi",
        "recommendedLimit": "257Mi",
        "currentUtilization": 70.3,
        "targetUtilization": 70.0,
        "resourceChange": "decrease",
        "percentageChange": 0.4
      },
      "reasons": ["cpu_limit_present", "memory_misaligned"],
      "priority": "high",
      "confidenceScore": 87.5,
      "estimatedSavings": "medium",
      "riskLevel": "low",
      "basedOnDays": 7,
      "dataQuality": "excellent"
    }
  ],
  "summary": {
    "totalPodsAnalyzed": 25,
    "podsNeedingOptimization": 18,
    "podsWellOptimized": 7,
    "totalCpuRequestIncrease": "2500m",
    "totalCpuRequestDecrease": "1800m",
    "totalMemoryRequestChange": "+2.1Gi",
    "estimatedCostSavings": "medium",
    "highPriorityRecommendations": 12,
    "mediumPriorityRecommendations": 6,
    "lowPriorityRecommendations": 7,
    "overUtilizedPods": 8,
    "underUtilizedPods": 10,
    "podsWithCpuLimits": 15,
    "memoryMisalignedPods": 12,
    "podsMissingRequests": 3
  },
  "generatedAt": "2025-01-15T14:30:25Z",
  "timeRange": {
    "start": "2025-01-08T14:30:25Z",
    "end": "2025-01-15T14:30:25Z"
  },
  "analysisWindow": "7 days",
  "targetUtilization": 70.0
}
```

## Recommendation Logic

### CPU Recommendations

1. **Target Calculation**: Uses historical average CPU usage to calculate target request
   ```
   Target CPU Request = Average Usage / (Target Utilization / 100)
   ```

2. **Limit Removal**: Recommends removing CPU limits to prevent throttling
   - CPU limits are generally harmful in Kubernetes
   - Allows pods to burst when cluster resources are available

3. **Safety Bounds**:
   - Minimum: 10m (configurable via `RECOMMENDATIONS_MIN_CPU_REQUEST`)
   - Maximum: 4000m (configurable via `RECOMMENDATIONS_MAX_CPU_REQUEST`)
   - Max scale up: 300% of current request
   - Max scale down: 70% reduction

### Memory Recommendations

1. **Target Calculation**: Uses historical P95 memory usage (95th percentile)
   ```
   Target Memory Request = P95 Usage / (Target Utilization / 100)
   Target Memory Limit = Target Memory Request
   ```

2. **Request/Limit Alignment**: Memory requests always equal limits
   - Prevents OOMKills due to burstable memory behavior
   - Ensures predictable memory allocation

3. **Safety Bounds**:
   - Minimum: 64Mi (configurable)
   - Maximum: 8Gi (configurable)
   - Same scaling limits as CPU

### Priority Levels

- **High Priority**: Over-utilized pods, missing requests, CPU limits present
- **Medium Priority**: Under-utilized pods, memory misaligned, missing limits  
- **Low Priority**: Well-optimized pods, minor adjustments

### Confidence Scoring

Confidence is calculated based on:
- **Data Completeness**: Number of available data points
- **Usage Variance**: Stability of resource usage patterns
- **Time Coverage**: How much of the analysis window has data

Score ranges: 0-100, where 100 is highest confidence.

## Configuration

Configure the recommendation engine using environment variables:

```bash
# Target utilization percentages (default: 70%)
RECOMMENDATIONS_TARGET_CPU_UTILIZATION=70.0
RECOMMENDATIONS_TARGET_MEMORY_UTILIZATION=70.0

# Resource bounds
RECOMMENDATIONS_MIN_CPU_REQUEST=10m
RECOMMENDATIONS_MAX_CPU_REQUEST=4000m
RECOMMENDATIONS_MIN_MEMORY_REQUEST=64Mi
RECOMMENDATIONS_MAX_MEMORY_REQUEST=8Gi

# Feature flags
RECOMMENDATIONS_REMOVE_CPU_LIMITS=true
RECOMMENDATIONS_ALIGN_MEMORY_REQUESTS_LIMITS=true
```

## Integration Examples

### Using with kubectl

You can use the recommendations to generate Kubernetes manifests:

```bash
# Get recommendations and process with jq
curl -s http://localhost:8080/api/pods/recommendations | \
  jq -r '.recommendations[] | select(.priority == "high") |
    "Pod: \(.podName) - CPU: \(.cpu.recommendedRequest // "remove"), Memory: \(.memory.recommendedRequest)"'
```

### Automation Scripts

The API is designed to integrate with GitOps workflows and automation tools. The structured JSON response includes all necessary information for:

- Generating Kubernetes resource manifests
- Creating pull requests with recommended changes
- Monitoring and alerting on optimization opportunities
- Cost analysis and resource planning

## Error Handling

The API returns appropriate HTTP status codes:

- `200 OK`: Recommendations generated successfully
- `503 Service Unavailable`: Metrics client not initialized
- `500 Internal Server Error`: Failed to retrieve metrics or generate recommendations

## Performance Considerations

- **Timeout**: 45 seconds for recommendation generation
- **Historical Data**: Requires 7 days of metrics history for optimal results
- **Memory Usage**: Processes all pod metrics in memory - suitable for clusters with thousands of pods
- **Caching**: Consider implementing caching for frequently accessed namespaces

## Limitations

1. **New Pods**: Limited recommendations for pods with insufficient historical data
2. **Seasonal Patterns**: 7-day analysis may not capture weekly or monthly patterns
3. **Batch Workloads**: May not be optimal for workloads with irregular usage patterns
4. **Resource Constraints**: Does not consider cluster-wide resource availability

## Best Practices

1. **Gradual Implementation**: Apply recommendations incrementally, starting with high-confidence, low-risk changes
2. **Monitoring**: Monitor pod performance after applying recommendations
3. **Review Cycle**: Run recommendations weekly and review changes before applying
4. **Testing**: Test recommendations in staging environments first
5. **Backup**: Keep track of original resource settings before changes
