# Deployment Workflows

## Overview
We have optimized the deployment process to reduce iteration time.

### Optimized "Fast Deploy" (Recommended)
**Script:** `./fast_deploy.sh`

This script uses **rsync** to synchronize your local code to the server and then executes `docker compose` remotely via SSH.

**Prerequisites:**
1.  **SSH Keys**: You must have passwordless SSH access to the server.
    ```bash
    ssh-copy-id walker@10.0.0.17
    ```
2.  **Rsync**: Standard on macOS/Linux.

**Benefits:**
- **No Local Docker Required**: Works even if Docker Desktop is broken or not installed.
- **Fast**: Only transfers changed source files.
- **Reliable**: Builds internally on the server using the server's environment.

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
docker compose -f tw-backend/docker-compose.prod.yml exec game-server ./game-server reset
# Or via the API/URL
curl http://10.0.0.17:3000/api/reset
```
Or manually restart specific infrastructure containers with volume clearing if absolutely necessary (destructive):
```bash
docker compose -f tw-backend/docker-compose.prod.yml down -v mongo postgis
```
