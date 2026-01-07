#!/bin/bash
set -euo pipefail

echo "=== Thousand Worlds Kubernetes Deployment ==="

# --- KUBECONFIG SETUP ---
# Use user-local kubeconfig if available, otherwise fall back to system location
USER_KUBECONFIG="$HOME/.kube/config"
SYSTEM_KUBECONFIG="/etc/rancher/k3s/k3s.yaml"

if [[ -f "$USER_KUBECONFIG" ]]; then
    export KUBECONFIG="$USER_KUBECONFIG"
    echo "Using kubeconfig: $USER_KUBECONFIG"
elif [[ -r "$SYSTEM_KUBECONFIG" ]]; then
    export KUBECONFIG="$SYSTEM_KUBECONFIG"
    echo "Using kubeconfig: $SYSTEM_KUBECONFIG"
else
    echo "ERROR: No accessible kubeconfig found."
    echo ""
    echo "Run the following ONE-TIME SETUP to create a user-accessible kubeconfig:"
    echo "  sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config"
    echo "  sudo chown \$(id -u):\$(id -g) ~/.kube/config"
    echo "  chmod 600 ~/.kube/config"
    echo "  sed -i 's/127.0.0.1/localhost/g' ~/.kube/config"
    echo ""
    exit 1
fi

# 1. Build Docker Images
echo "Building Core Physics Docker Image..."
docker build -t tw-backend/core-physics:latest -f tw-backend/Dockerfile.core-physics tw-backend

echo "Building Game Server Docker Image..."
docker build -t tw-backend/game-server:latest -f tw-backend/Dockerfile.game-server tw-backend

echo "Building Frontend Docker Image..."
docker build -t tw-frontend/frontend:latest -f tw-frontend/Dockerfile tw-frontend

# 1b. Import Images to K3s (REQUIRED for K3s)
# The user must be in the 'k3s' group OR run this via k3s subcommand
echo "Importing images to K3s (containerd)..."

# Check if user can access containerd socket directly
K3S_SOCKET="/run/k3s/containerd/containerd.sock"
if [[ -r "$K3S_SOCKET" ]] && [[ -w "$K3S_SOCKET" ]]; then
    # Direct access available (user is in k3s group or socket is world-accessible)
    docker save tw-backend/core-physics:latest | k3s ctr images import -
    docker save tw-backend/game-server:latest | k3s ctr images import -
    docker save tw-frontend/frontend:latest | k3s ctr images import -
else
    echo "Containerd socket requires elevated access. Using sudo for image import..."
    docker save tw-backend/core-physics:latest | sudo k3s ctr images import -
    docker save tw-backend/game-server:latest | sudo k3s ctr images import -
    docker save tw-frontend/frontend:latest | sudo k3s ctr images import -
fi

echo "Configuring K3s permissions..."

# 2. Apply Kubernetes Manifests in Order
echo "Applying Kubernetes Manifests..."

# Create Namespace if it doesn't exist
kubectl create namespace mud-world --dry-run=client -o yaml | kubectl apply -f -

# --- INGRESS CONTROLLER SETUP ---
# Check if nginx ingress controller is already installed
if ! kubectl get namespace ingress-nginx &>/dev/null; then
    echo "Installing Nginx Ingress Controller (bare metal)..."
    kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/baremetal/deploy.yaml
    
    echo "Waiting for Ingress Controller to be ready..."
    kubectl -n ingress-nginx wait --for=condition=available deployment/ingress-nginx-controller --timeout=120s || echo "Ingress controller taking longer than expected..."
    
    # Patch for bare metal: enable hostNetwork so Ingress listens on host's port 80
    echo "Patching Ingress Controller for host network access (port 80)..."
    kubectl patch deployment ingress-nginx-controller -n ingress-nginx --type='json' \
        -p='[{"op": "add", "path": "/spec/template/spec/hostNetwork", "value": true}]'
    
    # Restart to apply hostNetwork change
    kubectl -n ingress-nginx rollout restart deployment/ingress-nginx-controller
    kubectl -n ingress-nginx rollout status deployment/ingress-nginx-controller --timeout=60s || echo "Ingress restart pending..."
else
    echo "Nginx Ingress Controller already installed."
fi

# Infrastructure (NATS, Postgres, Redis, MinIO, Ollama)
kubectl apply -f tw-backend/deploy/k8s/00-infrastructure.yaml

# Config & Secrets
kubectl apply -f tw-backend/deploy/k8s/01-config-and-secrets.yaml

# Services (Headless & LoadBalanced)
kubectl apply -f tw-backend/deploy/k8s/02-service.yaml

# Game Server (Backend API)
kubectl apply -f tw-backend/deploy/k8s/03-game-server.yaml

# Core Physics StatefulSet
kubectl apply -f tw-backend/deploy/k8s/04-statefulset.yaml

# Frontend
kubectl apply -f tw-backend/deploy/k8s/05-frontend.yaml

# Ingress (routes traffic to frontend and game-server)
kubectl apply -f tw-backend/deploy/k8s/09-ingress.yaml

# 3. Force Rollout (Required because we use 'latest' tag and local images)
echo "Restarting deployments to pick up new images..."
kubectl -n mud-world rollout restart deployment/game-server
kubectl -n mud-world rollout restart deployment/frontend
kubectl -n mud-world rollout restart statefulset/world-simulation

# 4. Status
echo "Deployment applied. Waiting for rollouts..."
kubectl -n mud-world rollout status deployment/mud-postgis --timeout=60s || echo "Postgres rollout pending..."
kubectl -n mud-world rollout status statefulset/world-simulation --timeout=60s || echo "Physics rollout pending..."
kubectl -n mud-world rollout status deployment/game-server --timeout=60s || echo "Game Server rollout pending..."

echo "=== Deployment Complete ==="
echo "Frontend available at: http://10.0.0.17 (port 80 via Ingress)"
echo "Check pods: kubectl -n mud-world get pods"
echo "Check ingress: kubectl -n mud-world get ingress"

