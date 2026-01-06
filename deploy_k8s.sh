#!/bin/bash
set -euo pipefail

echo "=== Thousand Worlds Kubernetes Deployment ==="

# 1. Build Docker Images
echo "Building Core Physics Docker Image..."
docker build -t tw-backend/core-physics:latest -f tw-backend/Dockerfile.core-physics tw-backend

echo "Building Game Server Docker Image..."
docker build -t tw-backend/game-server:latest -f tw-backend/Dockerfile.game-server tw-backend

echo "Building Frontend Docker Image..."
docker build -t tw-frontend/frontend:latest -f tw-frontend/Dockerfile tw-frontend

# 1b. Import Images to K3s (REQUIRED for K3s)
echo "Importing images to K3s (containerd)..."
# We need to save from docker and import to k3s ctr
# Using 'sudo' might be needed depending on user permissions, but assuming script runs as user who can sudo or access k3s
docker save tw-backend/core-physics:latest | k3s ctr images import -
docker save tw-backend/game-server:latest | k3s ctr images import -
docker save tw-frontend/frontend:latest | k3s ctr images import -

# 1c. Setup Kubernetes Config
echo "Configuring K3s permissions..."
# Assumes /etc/rancher/k3s/k3s.yaml is already readable (manual setup required once)
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

# 2. Apply Kubernetes Manifests in Order
echo "Applying Kubernetes Manifests..."

# Create Namespace if it doesn't exist
kubectl create namespace mud-world --dry-run=client -o yaml | kubectl apply -f -

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
echo "Frontend available at: http://localhost:30000 (if using NodePort)"
echo "Check pods: kubectl -n mud-world get pods"
