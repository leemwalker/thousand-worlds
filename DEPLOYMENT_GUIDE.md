# Deployment Workflows

## Overview
We have standardized our deployment process on Kubernetes.

### Kubernetes Deployment (Standard)
**Script:** `./deploy_k8s.sh`

This is the primary script used for building and deploying the application.

**What it does:**
1.  **Builds Docker Images**: Builds `core-physics`, `game-server`, and `frontend` images.
2.  **Imports to K3s**: Imports the built images into the K3s containerd runtime.
3.  **Applies Manifests**: Applies all Kubernetes manifests from `tw-backend/deploy/k8s/`.
4.  **Rollout Restart**: Restarts deployments to ensure new images are picked up.

### Legacy "Update Build" (On Server)
**Script:** `./update_build.sh`

This script runs on the **server**. Use this if you want to deploy exactly what is in the Git repository (e.g., from a different machine or for a stable "release").

**Optimizations Made:**
- **Removed aggressive cache clearing**: The script no longer runs `docker builder prune -af` or uses `--no-cache`. This significantly speeds up builds.
- **Improved Frontend caching**: Added `.dockerignore` to `tw-frontend` to prevent `node_modules` from invalidating the build context.

## Troubleshooting

### "Old States Impacting Tests"
If you find that old database state is causing issues, do **not** verify by clearing build cache. Instead, use the game's reset commands:
```bash
# Reset world state
kubectl -n mud-world exec -it deployment/game-server -- ./game-server reset
# Or via the API/URL
curl http://10.0.0.17:8080/api/reset
```
Or manually restart specific infrastructure containers with volume clearing if absolutely necessary (destructive):
```bash
# K8s equivalent would be deleting the PVCs and restarting statefulsets/deployments
kubectl -n mud-world delete pvc --all
kubectl -n mud-world delete pods --all
```
