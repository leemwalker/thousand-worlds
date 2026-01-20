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
docker build --no-cache -t tw-frontend/frontend:latest -f tw-frontend/Dockerfile tw-frontend

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

# 2a. Generate SSL Certificates if missing
if ! kubectl -n mud-world get secret nginx-certs > /dev/null 2>&1; then
    echo "SSL Certificate Secret (nginx-certs) not found. Generating..."
    if [[ -f "./generate_certs.sh" ]]; then
        chmod +x ./generate_certs.sh
        ./generate_certs.sh
    else
        echo "WARNING: ./generate_certs.sh not found. Skipping certificate generation."
    fi
else
    echo "SSL Certificate Secret (nginx-certs) exists. Skipping generation."
fi

# 2b. Apply Patches to Resolve Port Conflicts (CRITICAL)
echo "Applying System Patches..."

# 1. Agones Allocator Conflict: Moves Agones from 443 to 4443
# This allows Nginx Ingress to bind to 443.
if kubectl -n agones-system get svc agones-allocator > /dev/null 2>&1; then
    echo "Patching Agones Allocator port to 4443..."
    kubectl -n agones-system patch service agones-allocator --type='json' -p='[{"op": "replace", "path": "/spec/ports/0/port", "value": 4443}]' || true
fi

# 2. Traefik Conflict: K3s restores Traefik on restart. We move it to 81/444.
# This prevents it from fighting Nginx for 80/443.
if kubectl -n kube-system get svc traefik > /dev/null 2>&1; then
    echo "Patching K3s Traefik ports to 81/444..."
    kubectl -n kube-system patch service traefik --type='json' -p='[{"op": "replace", "path": "/spec/ports/0/port", "value": 81}, {"op": "replace", "path": "/spec/ports/1/port", "value": 444}]' || true
fi

# 3. Restart Nginx LB Pod if it was stuck pending
# This ensures it retries binding if it failed earlier.
kubectl -n kube-system delete pod -l app=svclb-ingress-nginx-controller --ignore-not-found > /dev/null 2>&1 || true


# Apply all manifests in order (00-10)
echo "Applying manifests from tw-backend/deploy/k8s/..."
kubectl apply -f tw-backend/deploy/k8s/

# 3. Force Rollout (Required because we use 'latest' tag and local images)
echo "Restarting deployments to pick up new images..."
kubectl -n mud-world rollout restart deployment/game-server
kubectl -n mud-world rollout restart deployment/frontend

# 4. Status
echo "Deployment applied. Waiting for rollouts..."
kubectl -n mud-world rollout status deployment/mud-postgis --timeout=60s || echo "Postgres rollout pending..."
kubectl -n mud-world get fleet world-simulation-fleet
kubectl -n mud-world rollout status deployment/game-server --timeout=60s || echo "Game Server rollout pending..."

echo "=== Deployment Complete ==="
echo "Frontend available at: http://10.0.0.17:8080 (via Nginx Gateway)"
echo "Check pods: kubectl -n mud-world get pods"
echo "Check ingress: kubectl -n mud-world get ingress"

