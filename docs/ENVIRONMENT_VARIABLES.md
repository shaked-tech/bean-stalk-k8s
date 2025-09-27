# Environment Variables Configuration

This document describes all available environment variables for configuring the bean-stalk-k8s backend metrics collection.

## Core Configuration

### METRICS_BACKEND
**Default:** `vmagent`  
**Options:** `prometheus`, `vmagent`, `victoriametrics`  
**Description:** Selects which metrics backend to use for data collection.

**Examples:**
```bash
# Use Prometheus
METRICS_BACKEND=prometheus

# Use VictoriaMetrics/VAgent (default)
METRICS_BACKEND=vmagent
```

## Connection URLs

### METRICS_PROMETHEUS_URL
**Default:** `http://prometheus-stack-kube-prom-prometheus.pod-metrics-dashboard.svc.cluster.local:9090`  
**Description:** URL for connecting to Prometheus server.

**Examples:**
```bash
# Custom Prometheus URL
METRICS_PROMETHEUS_URL=http://my-prometheus.monitoring.svc.cluster.local:9090

# External Prometheus
METRICS_PROMETHEUS_URL=https://prometheus.example.com
```

### METRICS_VMAGENT_URL
**Default:** `http://victoria-metrics-victoria-metrics-cluster-vmselect.pod-metrics-dashboard.svc.cluster.local.:8481/select/0/prometheus`  
**Description:** URL for connecting to VictoriaMetrics/VAgent server.

**Examples:**
```bash
# Custom VictoriaMetrics URL
METRICS_VMAGENT_URL=http://victoria-metrics.monitoring.svc.cluster.local:8481/select/0/prometheus

# External VictoriaMetrics
METRICS_VMAGENT_URL=https://vmagent.example.com/prometheus
```

## Legacy Support (Backward Compatibility)

### PROMETHEUS_URL
**Description:** Legacy environment variable for Prometheus URL. Use `METRICS_PROMETHEUS_URL` instead.

### VMAGENT_URL
**Description:** Legacy environment variable for VAgent URL. Use `METRICS_VMAGENT_URL` instead.

## Advanced Configuration

### METRICS_TIMEOUT
**Default:** `30s`  
**Description:** Default timeout for metrics queries.

**Examples:**
```bash
# Longer timeout for complex queries
METRICS_TIMEOUT=60s

# Shorter timeout for simple queries
METRICS_TIMEOUT=15s
```

### METRICS_RETRY_ATTEMPTS
**Default:** `3`  
**Description:** Number of retry attempts for failed metrics queries.

**Examples:**
```bash
# More retries for unreliable networks
METRICS_RETRY_ATTEMPTS=5

# No retries
METRICS_RETRY_ATTEMPTS=0
```

## Feature Flags

### METRICS_ENABLE_CACHING
**Default:** `false`  
**Description:** Enable/disable metrics response caching.

**Examples:**
```bash
# Enable caching
METRICS_ENABLE_CACHING=true

# Disable caching (default)
METRICS_ENABLE_CACHING=false
```

### METRICS_ENABLE_HISTORICAL
**Default:** `true`  
**Description:** Enable/disable historical metrics analysis features.

**Examples:**
```bash
# Disable historical analysis
METRICS_ENABLE_HISTORICAL=false

# Enable historical analysis (default)
METRICS_ENABLE_HISTORICAL=true
```

### METRICS_ENABLE_TREND
**Default:** `true`  
**Description:** Enable/disable trend analysis features.

**Examples:**
```bash
# Disable trend analysis
METRICS_ENABLE_TREND=false

# Enable trend analysis (default)
METRICS_ENABLE_TREND=true
```

## Environment Variable Priority

The backend reads configuration in the following order (highest to lowest priority):

1. **New environment variables** (e.g., `METRICS_PROMETHEUS_URL`)
2. **Legacy environment variables** (e.g., `PROMETHEUS_URL`)  
3. **Default values**

## Kubernetes Configuration Examples

### Using ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: metrics-env-config
  namespace: pod-metrics-dashboard
data:
  METRICS_BACKEND: "prometheus"
  METRICS_PROMETHEUS_URL: "http://my-prometheus.monitoring.svc.cluster.local:9090"
  METRICS_TIMEOUT: "45s"
  METRICS_RETRY_ATTEMPTS: "5"
  METRICS_ENABLE_CACHING: "true"
```

### Using Deployment Environment Variables

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pod-metrics-backend
spec:
  template:
    spec:
      containers:
      - name: backend
        image: pod-metrics-backend:latest
        env:
        - name: METRICS_BACKEND
          value: "prometheus"
        - name: METRICS_PROMETHEUS_URL
          value: "http://my-prometheus.monitoring.svc.cluster.local:9090"
        envFrom:
        - configMapRef:
            name: metrics-env-config
```

## Docker/Docker Compose Examples

### Docker Run
```bash
docker run -e METRICS_BACKEND=prometheus \
           -e METRICS_PROMETHEUS_URL=http://prometheus:9090 \
           -e METRICS_TIMEOUT=60s \
           pod-metrics-backend:latest
```

### Docker Compose
```yaml
version: '3.8'
services:
  backend:
    image: pod-metrics-backend:latest
    environment:
      - METRICS_BACKEND=prometheus
      - METRICS_PROMETHEUS_URL=http://prometheus:9090
      - METRICS_TIMEOUT=60s
      - METRICS_ENABLE_CACHING=true
```

## Validation and Logging

The backend validates environment variables on startup and logs:

- ✅ Successfully loaded configuration values
- ⚠️ Warnings for invalid values (falls back to defaults)
- 🔄 Which legacy variables are being used
- 📊 Final configuration summary

**Example startup logs:**
```
INFO: Environment override - Backend: prometheus
INFO: Environment override - Prometheus URL: http://my-prometheus:9090
INFO: Metrics configuration loaded:
  - Backend: prometheus
  - URL: http://my-prometheus:9090
  - Timeout: 30s
  - Retry Attempts: 3
  - Features: Caching=false, Historical=true, Trend=true
```

## Troubleshooting

### Common Issues

1. **Invalid boolean values**: Use `true`/`false`, not `yes`/`no` or `1`/`0`
2. **Invalid timeout format**: Use Go duration format (`30s`, `1m`, `1h30m`)
3. **URL format**: Ensure URLs include protocol (`http://` or `https://`)

### Health Check Endpoint

Visit `/health` to verify current configuration:

```json
{
  "status": "healthy",
  "metricsClient": "available",
  "metricsBackend": "prometheus",
  "features": {
    "realTimeMetrics": true,
    "historicalAnalysis": true,
    "trendAnalysis": true
  }
}
```

## Migration Guide

### From Legacy Variables

**Before:**
```bash
PROMETHEUS_URL=http://prometheus:9090
VMAGENT_URL=http://vmagent:8481/prometheus
```

**After:**
```bash
METRICS_BACKEND=prometheus
METRICS_PROMETHEUS_URL=http://prometheus:9090
METRICS_VMAGENT_URL=http://vmagent:8481/prometheus
```

**Note:** Legacy variables are still supported but deprecated.
