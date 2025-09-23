#!/bin/bash

# Deploy Pod Metrics Dashboard to Kind cluster
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Check if kind cluster exists
if ! kind get clusters | grep -q "kind"; then
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
if ! kind load docker-image pod-metrics-backend:$VERSION; then
    echo -e "${RED}❌ Failed to load versioned backend image into Kind${NC}"
    exit 1
fi
if ! kind load docker-image pod-metrics-backend:latest; then
    echo -e "${RED}❌ Failed to load latest backend image into Kind${NC}"
    exit 1
fi

echo -e "${YELLOW}Loading frontend image (v$VERSION)...${NC}"
if ! kind load docker-image pod-metrics-frontend:$VERSION; then
    echo -e "${RED}❌ Failed to load versioned frontend image into Kind${NC}"
    exit 1
fi  
if ! kind load docker-image pod-metrics-frontend:latest; then
    echo -e "${RED}❌ Failed to load latest frontend image into Kind${NC}"
    exit 1
fi

# Verify images are loaded in Kind
echo "🔍 Verifying images in Kind cluster..."
docker exec -it kind-control-plane crictl images | grep pod-metrics || echo "Warning: Images may not be loaded correctly"

# Install metrics-server if not already installed
echo "📊 Installing metrics-server..."
## kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
# helm repo add metrics-server https://kubernetes-sigs.github.io/metrics-server/
# helm repo update
# helm upgrade --install metrics-server metrics-server/metrics-server --namespace kube-system --create-namespace
helm upgrade --install metrics-server metrics-server/metrics-server \
  --namespace kube-system \
  --create-namespace \
  --set "args={--secure-port=10250,--kubelet-insecure-tls}"

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

# Update deployment files with version
echo -e "${BLUE}🔧 Updating deployment files with version $VERSION...${NC}"

# Create versioned deployment files
sed "s|pod-metrics-backend:latest|pod-metrics-backend:$VERSION|g" k8s/backend-deployment.yaml > /tmp/backend-deployment-versioned.yaml
sed "s|pod-metrics-frontend:latest|pod-metrics-frontend:$VERSION|g" k8s/frontend-deployment.yaml > /tmp/frontend-deployment-versioned.yaml

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
