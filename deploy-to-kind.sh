#!/bin/bash

# Deploy Pod Metrics Dashboard to Kind cluster with configurable monitoring stack
set -e

CLUSTER_NAME="metrics"

# Configuration: Choose metrics backend (prometheus or vmagent)
METRICS_BACKEND="${METRICS_BACKEND:-prometheus}"  # Default to prometheus

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Pod Metrics Dashboard with Historical Analysis${NC}"
echo -e "${BLUE}=================================================${NC}"
echo ""
echo -e "${BLUE}Prerequisites Check:${NC}"
echo -e "✓ Kubernetes cluster (Kind)"
echo -e "✓ Docker for building images"
echo -e "✓ Helm 3.x for installing monitoring stack"
echo -e "✓ kubectl configured for cluster access"
echo ""

# Get version from parameter or VERSION file
VERSION_FILE="VERSION"
VERSION=""

if [[ $# -gt 0 ]]; then
    VERSION=$1
    echo -e "${BLUE}📋 Using specified version: ${GREEN}$VERSION${NC}"
else
    if [[ -f $VERSION_FILE ]]; then
        VERSION=$(cat $VERSION_FILE | tr -d '\n')
        echo -e "${BLUE}📋 Using version from VERSION file: ${GREEN}$VERSION${NC}"
    else
        echo -e "${RED}❌ No version specified and VERSION file not found${NC}"
        echo "Usage: $0 [version]"
        echo "Example: $0 0.1.0"
        exit 1
    fi
fi

# Validate version format
if [[ ! $VERSION =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "${RED}❌ Invalid version format: $VERSION${NC}"
    echo -e "${YELLOW}Expected format: MAJOR.MINOR.PATCH (e.g., 0.1.2)${NC}"
    exit 1
fi

echo -e "${BLUE}🚀 Deploying Pod Metrics Dashboard to Kind (v$VERSION)...${NC}"
echo -e "${BLUE}📊 Metrics Backend: ${GREEN}$METRICS_BACKEND${NC}"

# Check if kind cluster exists
if ! kind get clusters | grep -q $CLUSTER_NAME; then
    echo "❌ Kind cluster not found. Please create a Kind cluster first:"
    echo "   kind create cluster --config=kind-config.yaml"
    exit 1
fi

# Build Docker images concurrently
echo -e "${BLUE}📦 Building Docker images concurrently with version ${GREEN}$VERSION${NC}..."

# Function to build backend image
build_backend() {
    echo -e "${YELLOW}Building backend...${NC}"
    if docker build -t pod-metrics-backend:$VERSION -t pod-metrics-backend:latest ./backend; then
        echo -e "${GREEN}✅ Backend build completed${NC}"
        return 0
    else
        echo -e "${RED}❌ Failed to build backend image${NC}"
        return 1
    fi
}

# Function to build frontend image
build_frontend() {
    echo -e "${YELLOW}Building frontend...${NC}"
    if docker build -t pod-metrics-frontend:$VERSION -t pod-metrics-frontend:latest ./frontend; then
        echo -e "${GREEN}✅ Frontend build completed${NC}"
        return 0
    else
        echo -e "${RED}❌ Failed to build frontend image${NC}"
        return 1
    fi
}

# Start both builds in background
echo -e "${BLUE}🚀 Starting concurrent builds...${NC}"
build_backend &
BACKEND_PID=$!

build_frontend &
FRONTEND_PID=$!

# Wait for both builds to complete and capture exit codes
echo -e "${BLUE}⏳ Waiting for builds to complete...${NC}"
wait $BACKEND_PID
BACKEND_EXIT=$?

wait $FRONTEND_PID
FRONTEND_EXIT=$?

# Check if both builds succeeded
if [ $BACKEND_EXIT -ne 0 ]; then
    echo -e "${RED}❌ Backend build failed${NC}"
    exit 1
fi

if [ $FRONTEND_EXIT -ne 0 ]; then
    echo -e "${RED}❌ Frontend build failed${NC}"
    exit 1
fi

echo -e "${GREEN}✅ All Docker images built successfully${NC}"

# Verify images were built
echo "🔍 Verifying built images..."
docker images | grep pod-metrics

# Load images into Kind cluster
echo -e "${BLUE}📥 Loading images into Kind cluster...${NC}"
echo -e "${YELLOW}Loading backend image (v$VERSION)...${NC}"
if ! kind load --name $CLUSTER_NAME docker-image pod-metrics-backend:$VERSION; then
    echo -e "${RED}❌ Failed to load versioned backend image into Kind${NC}"
    exit 1
fi
if ! kind load --name $CLUSTER_NAME docker-image pod-metrics-backend:latest; then
    echo -e "${RED}❌ Failed to load latest backend image into Kind${NC}"
    exit 1
fi

echo -e "${YELLOW}Loading frontend image (v$VERSION)...${NC}"
if ! kind load --name $CLUSTER_NAME docker-image pod-metrics-frontend:$VERSION; then
    echo -e "${RED}❌ Failed to load versioned frontend image into Kind${NC}"
    exit 1
fi  
if ! kind load --name $CLUSTER_NAME docker-image pod-metrics-frontend:latest; then
    echo -e "${RED}❌ Failed to load latest frontend image into Kind${NC}"
    exit 1
fi

# Verify images are loaded in Kind
echo "🔍 Verifying images in Kind cluster..."
docker exec kind-control-plane crictl images | grep pod-metrics || echo "Warning: Images may not be loaded correctly"

echo ""
echo -e "${BLUE}📦 Setting up Monitoring Stack Prerequisites${NC}"
echo -e "${BLUE}===========================================${NC}"

# Add required Helm repositories based on metrics backend
echo -e "${BLUE}📋 Adding required Helm repositories...${NC}"

if [ "$METRICS_BACKEND" = "prometheus" ]; then
    echo -e "${YELLOW}Adding prometheus-community repository for Prometheus/Grafana stack...${NC}"
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts >/dev/null 2>&1 || true
elif [ "$METRICS_BACKEND" = "vmagent" ]; then
    echo -e "${YELLOW}Adding VictoriaMetrics repository for vmagent/VictoriaMetrics stack...${NC}"
    helm repo add vm https://victoriametrics.github.io/helm-charts >/dev/null 2>&1 || true
    echo -e "${YELLOW}Adding prometheus-community repository for kube-state-metrics...${NC}"
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts >/dev/null 2>&1 || true
fi

echo -e "${YELLOW}Adding metrics-server repository for real-time metrics...${NC}"
helm repo add metrics-server https://kubernetes-sigs.github.io/metrics-server/ >/dev/null 2>&1 || true

echo -e "${YELLOW}Updating Helm repositories...${NC}"
helm repo update >/dev/null 2>&1

echo -e "${GREEN}✅ Helm repositories configured successfully${NC}"

# Install metrics-server for real-time metrics
echo ""
echo -e "${BLUE}📊 Installing Metrics Server (Real-time metrics)${NC}"
echo -e "   ${YELLOW}Purpose:${NC} Provides current CPU/Memory usage for pods"
echo -e "   ${YELLOW}Chart:${NC} metrics-server/metrics-server"
helm upgrade --install metrics-server metrics-server/metrics-server \
  --namespace kube-system \
  --create-namespace \
  --set "args={--secure-port=10250,--kubelet-insecure-tls}"

# Install monitoring stack based on configured backend
echo ""
if [ "$METRICS_BACKEND" = "prometheus" ]; then
    echo -e "${BLUE}📈 Installing Prometheus Stack (Historical analysis)${NC}"
    echo -e "   ${YELLOW}Purpose:${NC} Collects and stores 7+ days of metrics for historical analysis"
    echo -e "   ${YELLOW}Chart:${NC} prometheus-community/kube-prometheus-stack"
    echo -e "   ${YELLOW}Includes:${NC} Prometheus, Grafana, AlertManager, Node Exporters"
    echo -e "   ${YELLOW}Storage:${NC} 15Gi for 7-day retention"
    echo -e "   ${YELLOW}Config:${NC} prometheus-values.yaml"
    helm upgrade --install prometheus-stack prometheus-community/kube-prometheus-stack \
      --namespace pod-metrics-dashboard \
      --create-namespace \
      --values prometheus-values.yaml

    echo "⏳ Waiting for Prometheus stack to be ready..."
    kubectl wait --for=condition=ready --timeout=300s pod -l app.kubernetes.io/name=prometheus -n pod-metrics-dashboard || echo "Warning: Prometheus may still be starting"
    kubectl wait --for=condition=ready --timeout=300s pod -l app.kubernetes.io/name=grafana -n pod-metrics-dashboard || echo "Warning: Grafana may still be starting"

elif [ "$METRICS_BACKEND" = "vmagent" ]; then
    echo -e "${BLUE}📈 Installing VictoriaMetrics Stack (vmagent + VictoriaMetrics)${NC}"
    echo -e "   ${YELLOW}Purpose:${NC} Lightweight metrics collection with high-performance storage"
    echo -e "   ${YELLOW}Chart:${NC} vm/victoria-metrics-cluster"
    echo -e "   ${YELLOW}Includes:${NC} vminsert, vmselect, vmstorage"
    echo -e "   ${YELLOW}Config:${NC} victoriametrics-values.yaml"
    
    # Install VictoriaMetrics Cluster
    helm upgrade --install victoria-metrics vm/victoria-metrics-cluster \
      --namespace pod-metrics-dashboard \
      --create-namespace \
      --values victoriametrics-values.yaml

    echo "⏳ Waiting for VictoriaMetrics cluster to be ready..."
    kubectl wait --for=condition=ready --timeout=300s pod -l app.kubernetes.io/name=vminsert -n pod-metrics-dashboard || echo "Warning: vminsert may still be starting"
    kubectl wait --for=condition=ready --timeout=300s pod -l app.kubernetes.io/name=vmselect -n pod-metrics-dashboard || echo "Warning: vmselect may still be starting"
    kubectl wait --for=condition=ready --timeout=300s pod -l app.kubernetes.io/name=vmstorage -n pod-metrics-dashboard || echo "Warning: vmstorage may still be starting"

    # Install kube-state-metrics for resource requests/limits
    echo -e "${BLUE}📊 Installing kube-state-metrics (Resource metadata)${NC}"
    echo -e "   ${YELLOW}Purpose:${NC} Provides resource requests/limits data"
    echo -e "   ${YELLOW}Chart:${NC} prometheus-community/kube-state-metrics"
    helm upgrade --install kube-state-metrics prometheus-community/kube-state-metrics \
      --namespace pod-metrics-dashboard \
      --create-namespace

    # Install vmagent using Helm
    echo -e "${BLUE}🔧 Installing vmagent (Metrics collector)${NC}"
    echo -e "   ${YELLOW}Purpose:${NC} Lightweight Kubernetes metrics scraping agent"
    echo -e "   ${YELLOW}Chart:${NC} vm/victoria-metrics-agent"
    echo -e "   ${YELLOW}Config:${NC} vmagent-values.yaml"
    helm upgrade --install vmagent vm/victoria-metrics-agent \
      --namespace pod-metrics-dashboard \
      --create-namespace \
      --values vmagent-values.yaml

    echo "⏳ Waiting for vmagent to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/vmagent-victoria-metrics-agent -n pod-metrics-dashboard || echo "Warning: vmagent may still be starting"
    kubectl wait --for=condition=available --timeout=300s deployment/kube-state-metrics -n pod-metrics-dashboard || echo "Warning: kube-state-metrics may still be starting"

else
    echo -e "${RED}❌ Unsupported metrics backend: $METRICS_BACKEND${NC}"
    echo -e "${YELLOW}Supported backends: prometheus, vmagent${NC}"
    exit 1
fi

# # Patch metrics-server for Kind (disable TLS verification)
# kubectl patch deployment metrics-server -n kube-system --patch '
# {
#   "spec": {
#     "template": {
#       "spec": {
#         "containers": [
#           {
#             "name": "metrics-server",
#             "args": [
#               "--cert-dir=/tmp",
#               "--secure-port=10250",
#               "--kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname",
#               "--kubelet-use-node-status-port",
#               "--metric-resolution=15s",
#               "--kubelet-insecure-tls"
#             ]
#           }
#         ]
#       }
#     }
#   }
# }'

# Wait for metrics-server to be ready
echo "⏳ Waiting for metrics-server to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/metrics-server -n kube-system

# Update deployment files with version and metrics backend configuration
echo -e "${BLUE}🔧 Updating deployment files with version $VERSION...${NC}"

# Create versioned deployment files with metrics backend configuration
sed "s|pod-metrics-backend:latest|pod-metrics-backend:$VERSION|g" k8s/backend-deployment.yaml > /tmp/backend-deployment-versioned.yaml
sed "s|pod-metrics-frontend:latest|pod-metrics-frontend:$VERSION|g" k8s/frontend-deployment.yaml > /tmp/frontend-deployment-versioned.yaml

# Configure backend environment variables based on metrics backend
if [ "$METRICS_BACKEND" = "vmagent" ]; then
    echo -e "${BLUE}🔧 Configuring backend for vmagent + VictoriaMetrics...${NC}"
    # Add vmagent environment variables to backend deployment
    sed -i '/env:/a\
          - name: METRICS_BACKEND\
            value: "vmagent"\
          - name: VMAGENT_URL\
            value: "http://victoria-metrics-victoria-metrics-cluster-vmselect.pod-metrics-dashboard.svc.cluster.local:8481/select/0/prometheus"' \
    /tmp/backend-deployment-versioned.yaml
else
    echo -e "${BLUE}🔧 Configuring backend for Prometheus...${NC}"
    # Backend defaults to prometheus configuration (no changes needed)
fi

# Deploy the application
echo -e "${BLUE}🔧 Deploying application manifests (v$VERSION)...${NC}"
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/rbac.yaml
kubectl apply -f /tmp/backend-deployment-versioned.yaml
kubectl apply -f /tmp/frontend-deployment-versioned.yaml

# Clean up temporary files
rm -f /tmp/backend-deployment-versioned.yaml /tmp/frontend-deployment-versioned.yaml

# Wait for deployments to be ready
echo "⏳ Waiting for deployments to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/pod-metrics-backend -n pod-metrics-dashboard
kubectl wait --for=condition=available --timeout=300s deployment/pod-metrics-frontend -n pod-metrics-dashboard

# Get the frontend service port
echo "🌐 Getting service information..."
kubectl get services -n pod-metrics-dashboard

echo ""
echo -e "${GREEN}✅ Deployment complete! (Version: $VERSION)${NC}"
echo ""
echo -e "${BLUE}📊 Deployed images:${NC}"
echo -e "  ${YELLOW}Backend:${NC}  pod-metrics-backend:$VERSION"
echo -e "  ${YELLOW}Frontend:${NC} pod-metrics-frontend:$VERSION"
echo ""
echo -e "${BLUE}🌐 To access the dashboard:${NC}"
echo "1. Port forward the frontend service:"
echo "   kubectl port-forward -n pod-metrics-dashboard service/pod-metrics-frontend-service 3000:8080"
echo ""
echo "2. Open your browser to: http://localhost:3000"
echo ""

if [ "$METRICS_BACKEND" = "prometheus" ]; then
    echo -e "${BLUE}📈 To access Grafana monitoring:${NC}"
    echo "1. Port forward the Grafana service:"
    echo "   kubectl port-forward -n pod-metrics-dashboard service/prometheus-stack-grafana 3001:80"
    echo ""
    echo "2. Open your browser to: http://localhost:3001"
    echo "   Username: admin"
    echo "   Password: pod-metrics-admin"
    echo ""
elif [ "$METRICS_BACKEND" = "vmagent" ]; then
    echo -e "${BLUE}📈 VictoriaMetrics Query Interface:${NC}"
    echo "1. Port forward the vmselect service:"
    echo "   kubectl port-forward -n pod-metrics-dashboard service/victoria-metrics-victoria-metrics-cluster-vmselect 8481:8481"
    echo ""
    echo "2. Access VictoriaMetrics UI: http://localhost:8481/select/0/vmui"
    echo ""
fi
echo ""
echo -e "${BLUE}📋 To view logs:${NC}"
echo "   kubectl logs -n pod-metrics-dashboard -l app=pod-metrics-backend"
echo "   kubectl logs -n pod-metrics-dashboard -l app=pod-metrics-frontend"
echo ""
echo -e "${BLUE}🧹 To clean up:${NC}"
echo "   kubectl delete namespace pod-metrics-dashboard"
echo ""
echo -e "${BLUE}🔧 Version management:${NC}"
echo "   ./version.sh show          # Show current version"
echo "   ./version.sh patch --deploy # Increment patch and deploy"
echo "   ./version.sh minor --deploy # Increment minor and deploy"
echo ""
echo -e "${BLUE}⚙️  Metrics Backend Selection:${NC}"
echo -e "${YELLOW}To deploy with Prometheus (default):${NC}"
echo "   ./deploy-to-kind.sh"
echo "   # OR explicitly:"
echo "   METRICS_BACKEND=prometheus ./deploy-to-kind.sh"
echo ""
echo -e "${YELLOW}To deploy with vmagent + VictoriaMetrics:${NC}"
echo "   METRICS_BACKEND=vmagent ./deploy-to-kind.sh"
echo ""
echo -e "${BLUE}🔧 Monitoring Stack Troubleshooting:${NC}"
echo -e "${YELLOW}If Prometheus/Grafana are not accessible:${NC}"
echo "1. Check if pods are running:"
echo "   kubectl get pods -n pod-metrics-dashboard | grep prometheus"
echo "   kubectl get pods -n pod-metrics-dashboard | grep grafana"
echo ""
echo "2. Check service status:"
echo "   kubectl get services -n pod-metrics-dashboard"
echo ""
echo "3. View logs for troubleshooting:"
echo "   kubectl logs -n pod-metrics-dashboard deployment/prometheus-stack-grafana"
echo "   kubectl logs -n pod-metrics-dashboard deployment/prometheus-stack-kube-prom-operator"
echo ""
echo -e "${YELLOW}If historical analysis endpoints return errors:${NC}"
echo "4. Verify Prometheus connectivity from backend:"
echo "   kubectl port-forward -n pod-metrics-dashboard service/prometheus-stack-kube-prom-prometheus 9090:9090"
echo "   curl http://localhost:9090/api/v1/query?query=up"
echo ""
echo -e "${GREEN}🎉 Deployment complete with full monitoring stack!${NC}"
