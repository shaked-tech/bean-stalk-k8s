# Kubernetes Pod Metrics Dashboard

A comprehensive web application that provides real-time visualization of Kubernetes pod resource usage, displaying CPU and memory metrics collected from the Kubernetes metrics-server. Features include namespace filtering, sorting capabilities, and visual progress bars showing resource utilization against requests and limits.

## 🚀 Features

- **Real-time Metrics**: Live pod CPU and memory usage data
- **Resource Comparison**: Visual comparison of usage vs requests/limits
- **Namespace Filtering**: Filter pods by specific namespaces
- **Multi-column Sorting**: Sort by name, namespace, CPU, memory, and percentages
- **Progress Bars**: Visual indicators for resource utilization
- **Summary Statistics**: Overview cards showing total pods, averages, and high-usage alerts
- **Responsive Design**: Material-UI components for a modern interface

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

- **Docker** and **Docker Compose**
- **Kubernetes cluster** with metrics-server installed
- **kubectl** configured to access your cluster
- **Kind** (for local testing) - optional
- **Node.js** 18+ (for local development)
- **Go** 1.21+ (for local development)

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

## 🔧 Development Setup

### Backend Development

```bash
cd backend

# Install dependencies
go mod download

# Run locally (requires kubeconfig)
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

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/namespaces` | List all namespaces |
| `GET` | `/api/pods` | Get all pod metrics |
| `GET` | `/api/pods?namespace=<name>` | Get pod metrics for specific namespace |
| `GET` | `/health` | Health check endpoint |

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
