# Deploying to Nobara Linux

This guide details how to deploy the Thousand Worlds application stack on a Nobara Linux machine with an NVIDIA GPU.

## System Requirements
- **OS**: Nobara Linux (or Fedora-based equivalent)
- **RAM**: 32GB+ (Postgres tuned for 4GB shared buffers)
- **GPU**: NVIDIA GPU (RTX 480 or newer recommended for AI features)
- **Drivers**: Proprietary NVIDIA Drivers installed

## Prerequisites

### 1. Install Docker & Nvidia Container Toolkit
Nobara usually simplifies this, but ensure you have the necessary packages:
```bash
sudo dnf install docker docker-compose-plugin nvidia-container-toolkit
sudo systemctl enable --now docker
```

Configure Docker to use Nvidia runtime if not automatic:
```bash
sudo nvidia-ctk runtime configure --runtime=docker
sudo systemctl restart docker
```

### 2. Firewall
Open necessary ports if accessing remotely:
- **3000**: Frontend
- **8080**: Backend API
```bash
sudo firewall-cmd --add-port=3000/tcp --permanent
sudo firewall-cmd --add-port=8080/tcp --permanent
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
   chmod +x deploy_nobara.sh
   ./deploy_nobara.sh
   ```

   The script will:
   - Check for Docker and Nvidia support.
   - Create a `.env` file from template if missing.
   - Build and start all services using `docker-compose.prod.yml`.

3. **Verify Status**:
   ```bash
   docker compose -f docker-compose.prod.yml ps
   docker compose -f docker-compose.prod.yml logs -f
   ```

## Troubleshooting
- **Ollama GPU Error**: If Ollama complains about no GPU, ensure `nvidia-smi` works on the host and the container toolkit is properly configured.
- **Permission Denied**: Ensure your user is in the `docker` group (`sudo usermod -aG docker $USER`).
