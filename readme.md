# Kubernetes Pod Metrics Dashboard

A comprehensive web application that provides both real-time and historical visualization of Kubernetes pod resource usage. Features include 7-day historical analysis, trend detection, resource optimization recommendations, and advanced monitoring through integrated Prometheus and Grafana stack.

## 🚀 Features

### Real-time Monitoring
- **Real-time Metrics**: Live pod CPU and memory usage data
- **Resource Comparison**: Visual comparison of usage vs requests/limits
- **Namespace Filtering**: Filter pods by specific namespaces
- **Multi-column Sorting**: Sort by name, namespace, CPU, memory, and percentages
- **Progress Bars**: Visual indicators for resource utilization
- **Summary Statistics**: Overview cards showing total pods, averages, and high-usage alerts
- **Responsive Design**: Material-UI components for a modern interface

### Historical Analysis & Intelligence
- **7-Day Historical Analysis**: Deep insights into resource usage patterns
- **Trend Detection**: Identifies increasing, decreasing, or stable usage trends
- **Resource Efficiency Analysis**: Calculates actual vs requested resource utilization
- **Smart Recommendations**: AI-powered suggestions for resource optimization
- **Waste Detection**: Identifies over-provisioned and under-provisioned resources
- **Statistical Analysis**: P95/P99 percentiles, averages, peaks, and minimums
- **Risk Assessment**: Categorizes pods by operational risk (low/medium/high)

### Advanced Monitoring Stack
- **Dual Metrics Backend Support**: Choose between Prometheus or VictoriaMetrics (vmagent)
- **Prometheus Integration**: Industry-standard time-series metrics collection  
- **VictoriaMetrics Integration**: High-performance, cost-effective alternative to Prometheus
- **Grafana Dashboards**: Professional monitoring dashboards with historical views
- **Custom Dashboards**: Pre-built dashboards for pod resource analysis
- **Alerting Ready**: Foundation for setting up resource usage alerts

## 🏗️ Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   React Frontend│    │   Go Backend     │    │ Kubernetes API  │
│   (nginx)       │───▶│   REST API       │───▶│ metrics-server  │
│   Port 80       │    │   Port 8080      │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

### Components

- **Backend (Go)**: 
  - REST API server using Gin framework
  - Kubernetes client for metrics-server API calls
  - CORS enabled for frontend communication
  - Health check endpoint

- **Frontend (React + TypeScript)**:
  - Material-UI components
  - Real-time data fetching with error handling
  - Responsive table with sorting and filtering
  - nginx reverse proxy for API calls

## 📋 Prerequisites

### Basic Requirements
- **Docker** and **Docker Compose**
- **Kubernetes cluster** with metrics-server installed
- **kubectl** configured to access your cluster
- **Kind** (for local testing) - optional
- **Node.js** 18+ (for local development)
- **Go** 1.21+ (for local development)

### Monitoring Stack Prerequisites
For historical analysis and advanced monitoring features, the following Helm repositories are required:

```bash
# Add required Helm repositories
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add metrics-server https://kubernetes-sigs.github.io/metrics-server/
helm repo update
```

#### Required Components:
1. **Prometheus Stack** - For historical metrics collection and analysis
   - **kube-prometheus-stack** Helm chart (includes Prometheus, Grafana, AlertManager)
   - Automatically installed by deployment script
   - Requires ~2GB memory and 15Gi storage for 7-day retention

2. **Metrics Server** - For real-time metrics
   - **metrics-server** Helm chart
   - Automatically installed by deployment script
   - Required for both real-time and historical analysis

3. **Grafana** - For advanced dashboard visualization
   - Included in kube-prometheus-stack
   - Pre-configured with custom pod metrics dashboards
   - Admin credentials: admin/admin

## 🚀 Quick Start

### Option 1: Local Development with Docker Compose

```bash
# Clone the repository
git clone <repository-url>
cd bean-stalk-k8s

# Start both services
docker-compose up --build

# Access the dashboard
open http://localhost:3000
```

### Option 2: Deploy to Kind (Local Kubernetes)

```bash
# Create Kind cluster with metrics-server support
kind create cluster --config=kind-config.yaml

# Deploy everything with one script
./deploy-to-kind.sh

# Port forward to access the dashboard
kubectl port-forward -n pod-metrics-dashboard service/pod-metrics-frontend-service 3000:80

# Access the dashboard
open http://localhost:3000
```

### Option 3: Deploy to Existing Kubernetes Cluster

```bash
# Ensure metrics-server is installed
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# Or using Helm
helm repo add metrics-server https://kubernetes-sigs.github.io/metrics-server/
helm repo update
helm install metrics-server metrics-server/metrics-server

# Build and push images (replace with your registry)
docker build -t your-registry/pod-metrics-backend:latest ./backend
docker build -t your-registry/pod-metrics-frontend:latest ./frontend
docker push your-registry/pod-metrics-backend:latest
docker push your-registry/pod-metrics-frontend:latest

# Update image names in k8s manifests, then deploy
kubectl apply -f k8s/
```

## 🔧 Metrics Backend Configuration

The application now supports dual metrics backends - choose between **Prometheus** or **VictoriaMetrics** (vmagent) based on your requirements.

### Backend Selection

Configure the metrics backend using environment variables:

```bash
# Use Prometheus (default)
export METRICS_BACKEND=prometheus
export PROMETHEUS_URL=http://prometheus-stack-kube-prom-prometheus.pod-metrics-dashboard.svc.cluster.local:9090

# Use VictoriaMetrics
export METRICS_BACKEND=vmagent
export VMAGENT_URL=http://vmselect-victoria-metrics.pod-metrics-dashboard.svc.cluster.local:8481
```

### Prometheus vs VictoriaMetrics Comparison

| Feature | Prometheus | VictoriaMetrics |
|---------|------------|-----------------|
| **Memory Usage** | Higher | Up to 10x lower |
| **Storage Efficiency** | Standard | Up to 10x compression |
| **Query Performance** | Good | Faster for large datasets |
| **PromQL Compatibility** | Native | 100% compatible |
| **Setup Complexity** | Simple | Slightly more complex |
| **Maturity** | Very mature | Growing rapidly |

### Deploy with Prometheus (Default)

```bash
# Deploy with existing Prometheus setup
./deploy-to-kind.sh

# The script will automatically install:
# - kube-prometheus-stack (Prometheus + Grafana)
# - metrics-server
# - Pod Metrics Dashboard (configured for Prometheus)
```

### Deploy with VictoriaMetrics

```bash
# Step 1: Deploy VictoriaMetrics stack
helm repo add vm https://victoriametrics.github.io/helm-charts/
helm repo update

# Install VictoriaMetrics cluster
helm install victoria-metrics vm/victoria-metrics-cluster \
  --namespace pod-metrics-dashboard \
  --create-namespace \
  --values victoriametrics-values.yaml

# Step 2: Deploy metrics-server
helm install metrics-server metrics-server/metrics-server \
  --namespace kube-system \
  --set args="{--cert-dir=/tmp,--secure-port=4443,--kubelet-insecure-tls,--kubelet-preferred-address-types=InternalIP\,ExternalIP\,Hostname}"

# Step 3: Deploy Pod Metrics Dashboard with VictoriaMetrics backend
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/rbac.yaml

# Create ConfigMap with VictoriaMetrics configuration
kubectl create configmap pod-metrics-config \
  --namespace pod-metrics-dashboard \
  --from-literal=METRICS_BACKEND=vmagent \
  --from-literal=VMAGENT_URL=http://vmselect-victoria-metrics.pod-metrics-dashboard.svc.cluster.local:8481

# Deploy backend with VictoriaMetrics configuration
kubectl set env deployment/pod-metrics-backend \
  --namespace pod-metrics-dashboard \
  METRICS_BACKEND=vmagent \
  VMAGENT_URL=http://vmselect-victoria-metrics.pod-metrics-dashboard.svc.cluster.local:8481

kubectl apply -f k8s/backend-deployment.yaml
kubectl apply -f k8s/frontend-deployment.yaml
```

### Switching Between Backends

You can switch between backends at runtime by updating environment variables:

```bash
# Switch to VictoriaMetrics
kubectl set env deployment/pod-metrics-backend \
  --namespace pod-metrics-dashboard \
  METRICS_BACKEND=vmagent \
  VMAGENT_URL=http://vmselect-victoria-metrics.pod-metrics-dashboard.svc.cluster.local:8481

# Switch to Prometheus
kubectl set env deployment/pod-metrics-backend \
  --namespace pod-metrics-dashboard \
  METRICS_BACKEND=prometheus \
  PROMETHEUS_URL=http://prometheus-stack-kube-prom-prometheus.pod-metrics-dashboard.svc.cluster.local:9090

# Restart the backend to apply changes
kubectl rollout restart deployment/pod-metrics-backend --namespace pod-metrics-dashboard
```

### Monitoring Stack Access (VictoriaMetrics)

When using VictoriaMetrics, access the monitoring interfaces:

**VMSelect** (Query interface, similar to Prometheus):
```bash
kubectl port-forward -n pod-metrics-dashboard service/vmselect-victoria-metrics 8481:8481
```
→ http://localhost:8481/select/0/prometheus/

**vmagent** (Metrics collection status):
```bash
kubectl port-forward -n pod-metrics-dashboard service/vmagent-victoria-metrics 8429:8429
```
→ http://localhost:8429/

### Configuration Files

- **prometheus-values.yaml**: Prometheus stack configuration
- **victoriametrics-values.yaml**: VictoriaMetrics stack configuration
- Both optimized for Kind cluster with 7-day retention

## 🔧 Development Setup

### Backend Development

```bash
cd backend

# Install dependencies
go mod download

# Run locally with Prometheus (default)
go run main.go

# Run locally with VictoriaMetrics
export METRICS_BACKEND=vmagent
export VMAGENT_URL=http://localhost:8481
go run main.go

# Build
go build -o bin/main main.go
```

### Frontend Development

```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm start

# Build for production
npm run build
```

## 📡 API Endpoints

### Real-time Metrics APIs
| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/namespaces` | List all namespaces |
| `GET` | `/api/pods` | Get current pod metrics |
| `GET` | `/api/pods?namespace=<name>` | Get pod metrics for specific namespace |
| `GET` | `/health` | Health check with feature availability |

### Historical Analysis APIs
| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/pods/analysis` | Get 7-day historical analysis for all pods |
| `GET` | `/api/pods/analysis?namespace=<name>` | Get 7-day analysis for specific namespace |
| `GET` | `/api/pods/trends?namespace=<ns>&pod=<name>` | Get detailed trend analysis for specific pod |

### Monitoring Stack Access
After deployment, access the monitoring interfaces:

**Pod Metrics Dashboard** (Enhanced with historical analysis):
```bash
kubectl port-forward -n pod-metrics-dashboard service/pod-metrics-frontend-service 3000:8080
```
→ http://localhost:3000

**Grafana** (Professional monitoring dashboards):
```bash
kubectl port-forward -n pod-metrics-dashboard service/prometheus-stack-grafana 3001:80
```
→ http://localhost:3001
- **Username**: admin
- **Password**: admin

**Prometheus** (Raw metrics and query interface):
```bash
kubectl port-forward -n pod-metrics-dashboard service/prometheus-stack-kube-prom-prometheus 9090:9090
```
→ http://localhost:9090

### Response Examples

**GET /api/namespaces**
```json
{
  "namespaces": ["default", "kube-system", "pod-metrics-dashboard"]
}
```

**GET /api/pods**
```json
{
  "pods": [
    {
      "name": "my-app-7d4b8b5f4c-xyz12",
      "namespace": "default",
      "containerName": "my-app",
      "cpu": {
        "usage": "15m",
        "request": "100m",
        "limit": "200m",
        "usageValue": 15,
        "requestValue": 100,
        "limitValue": 200,
        "requestPercentage": 15.0,
        "limitPercentage": 7.5
      },
      "memory": {
        "usage": "128Mi",
        "request": "256Mi",
        "limit": "512Mi",
        "usageValue": 134217728,
        "requestValue": 268435456,
        "limitValue": 536870912,
        "requestPercentage": 50.0,
        "limitPercentage": 25.0
      },
      "labels": {
        "app": "my-app",
        "version": "v1.0.0"
      }
    }
  ]
}
```

## 🐳 Docker Images

### Building Images

```bash
# Backend
docker build -t pod-metrics-backend:latest ./backend

# Frontend  
docker build -t pod-metrics-frontend:latest ./frontend
```

### Image Details

- **Backend**: Multi-stage Go build with minimal Alpine runtime
- **Frontend**: Multi-stage Node.js build with nginx serving static files
- Both images are optimized for size and security

## ☸️ Kubernetes Deployment

The application includes comprehensive Kubernetes manifests:

- **Namespace**: Isolated environment for the application
- **RBAC**: Service account with minimal required permissions
- **Backend Deployment**: Go API server with health checks
- **Frontend Deployment**: nginx serving React app with API proxy
- **Services**: Internal communication and external access
- **ConfigMap**: nginx configuration for API proxy

### RBAC Permissions

The backend service account has minimal permissions:
- Read access to pods and namespaces
- Access to metrics.k8s.io API for pod metrics

## 🔍 Monitoring and Troubleshooting

### Viewing Logs

```bash
# Backend logs
kubectl logs -n pod-metrics-dashboard -l app=pod-metrics-backend

# Frontend logs
kubectl logs -n pod-metrics-dashboard -l app=pod-metrics-frontend

# Follow logs in real-time
kubectl logs -n pod-metrics-dashboard -l app=pod-metrics-backend -f
```

### Common Issues

1. **Metrics Server Not Found**
   ```bash
   # Check if metrics-server is running
   kubectl get deployment metrics-server -n kube-system
   
   # If not installed, install it
   kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
   ```

2. **RBAC Permission Errors**
   ```bash
   # Check if service account has proper permissions
   kubectl auth can-i get pods --as=system:serviceaccount:pod-metrics-dashboard:pod-metrics-backend
   ```

3. **Frontend Can't Reach Backend**
   - Check if nginx proxy configuration is correct
   - Verify backend service is running and accessible

### Health Checks

```bash
# Check backend health
kubectl port-forward -n pod-metrics-dashboard service/pod-metrics-backend-service 8080:8080
curl http://localhost:8080/health

# Check frontend
kubectl port-forward -n pod-metrics-dashboard service/pod-metrics-frontend-service 3000:80
curl http://localhost:3000
```

## 🧹 Cleanup

### Remove from Kubernetes

```bash
# Delete the entire namespace (removes everything)
kubectl delete namespace pod-metrics-dashboard

# Or delete individual resources
kubectl delete -f k8s/
```

### Remove Kind Cluster

```bash
kind delete cluster
```

### Stop Docker Compose

```bash
docker-compose down
```

## 📊 Screenshots

The dashboard provides:
- Summary cards with total pods, average usage, and high-usage alerts
- Sortable table with all pod metrics
- Visual progress bars for resource utilization
- Namespace filtering dropdown
- Real-time refresh capability

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test with Kind cluster
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License.
