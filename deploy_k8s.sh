#!/bin/bash
set -euo pipefail

echo "=== Thousand Worlds Kubernetes Deployment ==="

# 1. Build the Physics Image
echo "Building Core Physics Docker Image..."
# We tag it specifically so the local K8s (if using shared docker daemon like MicroK8s or Docker Desktop) can find it
docker build -t tw-backend/core-physics:latest -f tw-backend/Dockerfile.core-physics tw-backend

# 2. Apply Kubernetes Manifests in Order
echo "Applying Kubernetes Manifests..."

# Infrastructure (NATS, Postgres, etc.)
kubectl apply -f tw-backend/deploy/k8s/00-infrastructure.yaml

# Config & Secrets
kubectl apply -f tw-backend/deploy/k8s/01-config-and-secrets.yaml

# Services
kubectl apply -f tw-backend/deploy/k8s/02-service.yaml

# Core Physics StatefulSet
kubectl apply -f tw-backend/deploy/k8s/04-statefulset.yaml

# 3. Status
echo "Deployment applied. Waiting for rollout..."
kubectl -n mud-world rollout status statefulset/world-simulation --timeout=60s || echo "Rollout still in progress..."

echo "=== Deployment Complete ==="
echo "Check status with: kubectl -n mud-world get pods"
