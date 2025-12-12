# Deploying to Fedora 42 Server

Deploy the Thousand Worlds application stack on Fedora 42 Server.

## System Requirements
- **OS**: Fedora 42 Server
- **CPU**: Intel Core i7-6700K @ 4.0GHz (or equivalent)
- **RAM**: 32GB+ (Postgres tuned for 4GB shared buffers)
- **Storage**: 450GB+ SSD recommended
- **GPU**: AMD Radeon RX 480 (CPU-only for LLM inference)

## Prerequisites

### 1. Install Docker & Docker Compose
```bash
# Add Docker repository
sudo dnf config-manager --add-repo https://download.docker.com/linux/fedora/docker-ce.repo

# Install Docker
sudo dnf install docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Enable and start Docker
sudo systemctl enable --now docker

# Add user to docker group (log out and back in after)
sudo usermod -aG docker $USER
```

### 2. Configure Firewall
Open necessary ports for remote access:
```bash
# Frontend
sudo firewall-cmd --add-port=3000/tcp --permanent

# Backend API
sudo firewall-cmd --add-port=8080/tcp --permanent

# Reload firewall
sudo firewall-cmd --reload
```

## Deployment

1. **Clone the repository**:
   ```bash
   git clone <repo_url>
   cd thousand-worlds/tw-backend
   ```

2. **Run the deployment script**:
   ```bash
   cd deploy
   chmod +x deploy_fedora.sh
   ./deploy_fedora.sh
   ```

   The script will:
   - Check for Docker installation
   - Create a `.env` file from template if missing
   - Build and start all services using `docker-compose.prod.yml`

3. **Verify Status**:
   ```bash
   docker compose -f docker-compose.prod.yml ps
   docker compose -f docker-compose.prod.yml logs -f
   ```

## GPU Acceleration (Optional)

> [!NOTE]
> The AMD Radeon RX 480 (Polaris) has limited ROCm support. Ollama runs in CPU-only mode by default for maximum compatibility.

If you upgrade to an AMD GPU with official ROCm support (RDNA2+), you can enable GPU acceleration:

```bash
# Install ROCm
sudo dnf install rocm-hip-runtime rocm-smi

# Verify GPU detection
rocm-smi

# Update docker-compose.prod.yml to enable GPU passthrough
```

## Troubleshooting

- **Permission Denied**: Ensure your user is in the `docker` group and re-login
- **Port Conflicts**: Check if ports 3000, 5432, 8080 are already in use
- **Slow LLM**: Expected behavior with CPU-only inference; consider smaller models
