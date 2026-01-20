#!/bin/bash
# =============================================================================
# Thousand Worlds - Full K3s Cluster Rebuild Script
# =============================================================================
# This script performs a COMPLETE reset of the K3s cluster and redeploys
# all services. WARNING: This will DELETE ALL DATA.
#
# Usage: ./rebuild_k8s.sh
# =============================================================================

set -euo pipefail

echo "=============================================="
echo "  THOUSAND WORLDS - FULL K3S CLUSTER REBUILD  "
echo "=============================================="
echo ""
echo "⚠️  WARNING: This will DELETE ALL DATA including:"
echo "    - All databases (tw-postgis, tw-timescaledb)"
echo "    - All saved worlds and simulations"
echo "    - All user accounts and characters"
echo ""
read -p "Are you sure you want to continue? (yes/no): " confirm
if [[ "$confirm" != "yes" ]]; then
    echo "Aborted."
    exit 0
fi

# --- KUBECONFIG SETUP ---
USER_KUBECONFIG="$HOME/.kube/config"
SYSTEM_KUBECONFIG="/etc/rancher/k3s/k3s.yaml"

if [[ -f "$USER_KUBECONFIG" ]]; then
    export KUBECONFIG="$USER_KUBECONFIG"
elif [[ -r "$SYSTEM_KUBECONFIG" ]]; then
    export KUBECONFIG="$SYSTEM_KUBECONFIG"
else
    echo "ERROR: No accessible kubeconfig found."
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# =============================================================================
# PHASE 1: Clean Shutdown
# =============================================================================
echo ""
echo "=== PHASE 1: Clean Shutdown ==="

echo "Deleting application namespaces (this may take a minute)..."
kubectl delete namespace tw-world --ignore-not-found --timeout=60s || true
kubectl delete namespace tw-ingress --ignore-not-found --timeout=60s || true

echo "Waiting for namespace deletion..."
while kubectl get namespace tw-world &>/dev/null 2>&1; do
    echo "  Waiting for tw-world namespace to terminate..."
    sleep 2
done
while kubectl get namespace tw-ingress &>/dev/null 2>&1; do
    echo "  Waiting for tw-ingress namespace to terminate..."
    sleep 2
done

echo "✓ Namespaces deleted"

# =============================================================================
# PHASE 2: K3s Full Reset (Optional but recommended)
# =============================================================================
echo ""
echo "=== PHASE 2: K3s Reset ==="

read -p "Perform full K3s reset? (Recommended for networking issues) (yes/no): " do_reset
if [[ "$do_reset" == "yes" ]]; then
    echo "Stopping K3s..."
    sudo /usr/local/bin/k3s-killall.sh || true
    
    echo "Cleaning up containerd state..."
    sudo rm -rf /var/lib/rancher/k3s/agent/containerd/io.containerd.* || true
    
    echo "Restarting K3s service..."
    sudo systemctl restart k3s
    
    echo "Waiting for K3s to be ready..."
    sleep 10
    
    # Wait for K3s API to be available
    until kubectl get nodes &>/dev/null 2>&1; do
        echo "  Waiting for K3s API..."
        sleep 3
    done
    
    echo "✓ K3s restarted successfully"
    
    # Re-export kubeconfig after restart
    if [[ -f "$USER_KUBECONFIG" ]]; then
        export KUBECONFIG="$USER_KUBECONFIG"
    fi
else
    echo "Skipping K3s reset."
fi

# =============================================================================
# PHASE 3: Port Conflict Resolution
# =============================================================================
echo ""
echo "=== PHASE 3: Resolving Port Conflicts ==="

# Wait for Traefik to come back up after K3s restart
echo "Waiting for Traefik to be available..."
sleep 5

# Patch Traefik to use non-standard ports (K3s brings it back on restart)
if kubectl -n kube-system get svc traefik &>/dev/null 2>&1; then
    echo "Patching Traefik to ports 81/444..."
    kubectl -n kube-system patch service traefik --type='json' \
        -p='[{"op": "replace", "path": "/spec/ports/0/port", "value": 81}, {"op": "replace", "path": "/spec/ports/1/port", "value": 444}]' || true
fi

# Patch Agones Allocator if present
if kubectl -n agones-system get svc agones-allocator &>/dev/null 2>&1; then
    echo "Patching Agones Allocator to port 4443..."
    kubectl -n agones-system patch service agones-allocator --type='json' \
        -p='[{"op": "replace", "path": "/spec/ports/0/port", "value": 4443}]' || true
fi

echo "✓ Port conflicts resolved"

# =============================================================================
# PHASE 4: Build Docker Images
# =============================================================================
echo ""
echo "=== PHASE 4: Building Docker Images ==="

echo "Building Core Physics..."
docker build -t tw-backend/core-physics:latest -f tw-backend/Dockerfile.core-physics tw-backend

echo "Building Game Server..."
docker build -t tw-backend/game-server:latest -f tw-backend/Dockerfile.game-server tw-backend

echo "Building Frontend..."
docker build --no-cache -t tw-frontend/frontend:latest -f tw-frontend/Dockerfile tw-frontend

echo "✓ Docker images built"

# =============================================================================
# PHASE 5: Import Images to K3s
# =============================================================================
echo ""
echo "=== PHASE 5: Importing Images to K3s ==="

K3S_SOCKET="/run/k3s/containerd/containerd.sock"
if [[ -r "$K3S_SOCKET" ]] && [[ -w "$K3S_SOCKET" ]]; then
    docker save tw-backend/core-physics:latest | k3s ctr images import -
    docker save tw-backend/game-server:latest | k3s ctr images import -
    docker save tw-frontend/frontend:latest | k3s ctr images import -
else
    echo "Using sudo for image import..."
    docker save tw-backend/core-physics:latest | sudo k3s ctr images import -
    docker save tw-backend/game-server:latest | sudo k3s ctr images import -
    docker save tw-frontend/frontend:latest | sudo k3s ctr images import -
fi

echo "✓ Images imported to K3s"

# =============================================================================
# PHASE 6: Create Namespaces and SSL Certificates
# =============================================================================
echo ""
echo "=== PHASE 6: Creating Namespaces and SSL Certificates ==="

# Create namespaces
kubectl create namespace tw-world --dry-run=client -o yaml | kubectl apply -f -
kubectl create namespace tw-ingress --dry-run=client -o yaml | kubectl apply -f -

# Generate SSL certificates
echo "Generating SSL certificates..."
chmod +x ./generate_certs.sh
./generate_certs.sh

# CRITICAL FIX: Copy certs to tw-ingress namespace for Nginx Ingress Controller
# The Ingress resource is in tw-world but Nginx controller is in tw-ingress
echo "Copying SSL certificates to tw-ingress namespace..."
kubectl get secret nginx-certs -n tw-world -o yaml | \
    sed 's/namespace: tw-world/namespace: tw-ingress/' | \
    kubectl apply -f -

echo "✓ Namespaces and certificates ready"

# =============================================================================
# PHASE 7: Apply Kubernetes Manifests
# =============================================================================
echo ""
echo "=== PHASE 7: Applying Kubernetes Manifests ==="

kubectl apply -f tw-backend/deploy/k8s/

echo "✓ Manifests applied"

# =============================================================================
# PHASE 8: Wait for Critical Pods
# =============================================================================
echo ""
echo "=== PHASE 8: Waiting for Pods ==="

echo "Waiting for PostGIS..."
kubectl -n tw-world wait --for=condition=ready pod -l app=tw-postgis --timeout=120s || echo "PostGIS still starting..."

echo "Waiting for NATS..."
kubectl -n tw-world wait --for=condition=ready pod -l app=tw-nats --timeout=60s || echo "NATS still starting..."

echo "Waiting for DragonflyDB..."
kubectl -n tw-world wait --for=condition=ready pod -l app=tw-dragonfly --timeout=60s || echo "Dragonfly still starting..."

echo "Waiting for Game Server..."
kubectl -n tw-world rollout status deployment/game-server --timeout=120s || echo "Game Server still starting..."

echo "Waiting for Frontend..."
kubectl -n tw-world rollout status deployment/frontend --timeout=120s || echo "Frontend still starting..."

echo "Waiting for Nginx Ingress Controller..."
kubectl -n tw-ingress wait --for=condition=ready pod -l app.kubernetes.io/name=ingress-nginx --timeout=120s || echo "Ingress still starting..."

# =============================================================================
# PHASE 9: Run Database Migrations
# =============================================================================
echo ""
echo "=== PHASE 9: Running Database Migrations ==="

# Wait a moment for PostGIS to be fully ready for connections
sleep 5

# Run migrations by exec into the game-server pod
GAME_SERVER_POD=$(kubectl -n tw-world get pod -l app=game-server -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)

if [[ -n "$GAME_SERVER_POD" ]]; then
    echo "Running migrations via game-server pod: $GAME_SERVER_POD"
    # The game-server typically runs migrations on startup, but we can trigger manually if there's a migrate command
    echo "Note: Game server runs migrations automatically on startup."
else
    echo "Warning: Could not find game-server pod for migrations."
    echo "You may need to run migrations manually: cd tw-backend && make migrate"
fi

echo "✓ Database initialization in progress (migrations run on app startup)"

# =============================================================================
# PHASE 10: Verification
# =============================================================================
echo ""
echo "=== PHASE 10: Verification ==="

echo ""
echo "Pod Status:"
kubectl -n tw-world get pods -o wide

echo ""
echo "Ingress Status:"
kubectl -n tw-world get ingress

echo ""
echo "Nginx Ingress Controller Status:"
kubectl -n tw-ingress get pods

echo ""
echo "Services:"
kubectl -n tw-world get svc

# Test connectivity
echo ""
echo "Testing HTTPS connectivity..."
sleep 3
if curl -k -s --connect-timeout 5 https://10.0.0.17 > /dev/null 2>&1; then
    echo "✓ HTTPS://10.0.0.17 is accessible!"
else
    echo "⚠ HTTPS://10.0.0.17 not yet accessible (pods may still be starting)"
    echo "  Try: curl -k https://10.0.0.17"
fi

# =============================================================================
# Complete
# =============================================================================
echo ""
echo "=============================================="
echo "  REBUILD COMPLETE"
echo "=============================================="
echo ""
echo "Access the application at: https://10.0.0.17"
echo ""
echo "Useful commands:"
echo "  kubectl -n tw-world get pods              # Check pod status"
echo "  kubectl -n tw-world logs deploy/game-server  # View game server logs"
echo "  kubectl -n tw-world logs deploy/frontend     # View frontend logs"
echo "  kubectl -n tw-ingress logs deploy/ingress-nginx-controller  # Ingress logs"
echo ""
