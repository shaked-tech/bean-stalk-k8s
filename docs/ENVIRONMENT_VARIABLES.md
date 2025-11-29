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

### METRICS_VICTORIAMETRICS_URL
**Default:** `http://victoria-metrics-victoria-metrics-cluster-vmselect.pod-metrics-dashboard.svc.cluster.local:8481/select/0/prometheus`  
**Description:** URL for connecting to VictoriaMetrics/VAgent server.

**Examples:**
```bash
# Custom VictoriaMetrics URL
METRICS_VICTORIAMETRICS_URL=http://victoria-metrics.monitoring.svc.cluster.local:8481/select/0/prometheus

# External VictoriaMetrics
METRICS_VICTORIAMETRICS_URL=https://vmagent.example.com/prometheus
```

### METRICS_VICTORIAMETRICS_USERNAME
**Default:** `` (empty - authentication disabled)  
**Description:** Optional username for Basic Authentication with VictoriaMetrics.

**Examples:**
```bash
# Enable authentication
METRICS_VICTORIAMETRICS_USERNAME=admin
METRICS_VICTORIAMETRICS_PASSWORD=secret123
```

### METRICS_VICTORIAMETRICS_PASSWORD
**Default:** `` (empty - authentication disabled)  
**Description:** Optional password for Basic Authentication with VictoriaMetrics. Must be used with METRICS_VICTORIAMETRICS_USERNAME.

**Examples:**
```bash
# Enable authentication
METRICS_VICTORIAMETRICS_USERNAME=admin
METRICS_VICTORIAMETRICS_PASSWORD=secret123
```

### METRICS_PROMETHEUS_USERNAME
**Default:** `` (empty - authentication disabled)  
**Description:** Optional username for Basic Authentication with Prometheus.

**Examples:**
```bash
# Enable authentication
METRICS_PROMETHEUS_USERNAME=admin
METRICS_PROMETHEUS_PASSWORD=secret456
```

### METRICS_PROMETHEUS_PASSWORD
**Default:** `` (empty - authentication disabled)  
**Description:** Optional password for Basic Authentication with Prometheus. Must be used with METRICS_PROMETHEUS_USERNAME.

**Examples:**
```bash
# Enable authentication
METRICS_PROMETHEUS_USERNAME=admin
METRICS_PROMETHEUS_PASSWORD=secret456
```

### METRICS_INSECURE_SKIP_VERIFY
**Default:** `false` (certificate verification enabled - secure by default)  
**Description:** Skip TLS certificate verification when connecting to metrics backends. When set to `true`, the backend will not verify SSL/TLS certificates, allowing connections to servers with self-signed or untrusted certificates.

⚠️ **Security Warning:** This option should only be used in development/testing environments or trusted internal networks. Using this in production reduces connection security and exposes you to man-in-the-middle attacks.

**Examples:**
```bash
# Disable certificate verification (for development/testing only)
METRICS_INSECURE_SKIP_VERIFY=true

# Enable certificate verification (default - recommended for production)
METRICS_INSECURE_SKIP_VERIFY=false
```

**When to use:**
- ✅ Development environments with self-signed certificates
- ✅ Testing environments
- ✅ Trusted internal networks with self-signed CA
- ❌ Production environments with external services
- ❌ Any environment where security is critical

**Logging:**
When `METRICS_INSECURE_SKIP_VERIFY=true`, the backend will log prominent security warnings on startup:
```
⚠️  WARNING: TLS certificate verification is DISABLED
⚠️  WARNING: This should only be used in development/testing environments
⚠️  WARNING: Connection security is reduced - use at your own risk
```

### METRICS_K8S_CLUSTER
**Default:** `` (empty - no cluster filtering)  
**Description:** Optional filter to query metrics from a specific Kubernetes cluster when using VictoriaMetrics with multi-cluster monitoring. This adds a `k8s_cluster` label filter to all queries. Only applies when using VictoriaMetrics backend.

**Examples:**
```bash
# Query metrics from a specific cluster
METRICS_K8S_CLUSTER=i1-k8s-mgmt

# Query from production cluster
METRICS_K8S_CLUSTER=prod-cluster-01

# No filtering (default - queries all clusters)
METRICS_K8S_CLUSTER=
```

**When to use:**
- ✅ Multi-cluster VictoriaMetrics deployments with aggregated metrics
- ✅ When you need to isolate metrics from a specific cluster
- ✅ In environments with the `k8s_cluster` label in your metrics
- ❌ Single cluster deployments (not needed)
- ❌ When using Prometheus backend (has no effect)

**Example URL with cluster filter:**
```
https://victoria-metrics.dyn:8428/api/v1/query?query=container_cpu_usage_seconds_total{k8s_cluster="i1-k8s-mgmt"}
```

**Logging:**
When `METRICS_K8S_CLUSTER` is set, the backend will log:
```
INFO: VictoriaMetrics client configured with k8s_cluster filter: i1-k8s-mgmt
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
