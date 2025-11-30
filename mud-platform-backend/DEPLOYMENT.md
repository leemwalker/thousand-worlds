# Deployment Guide - Thousand Worlds MUD Platform

## Quick Start (Development)

### Prerequisites
- Docker & Docker Compose
- Go 1.21+ (for local development)
- Node.js 18+ (for frontend)

### 1. Start Infrastructure Services

```bash
cd mud-platform-backend/deploy
docker-compose up -d
```

This starts:
- PostgreSQL with PostGIS
- MongoDB
- NATS JetStream
- Redis
- Ollama (LLM)

### 2. Run Database Migrations

```bash
cd mud-platform-backend
go run cmd/migrate/main.go
```

### 3. Start Game Server

**Option A: Local Development**
```bash
./scripts/dev.sh
```

**Option B: Docker**
```bash
docker-compose -f docker-compose.prod.yml up game-server
```

### 4. Pull LLM Model

```bash
docker exec mud_ollama ollama pull llama3.1:8b
```

---

## Production Deployment

### Environment Variables

Create a `.env` file:

```bash
# Database
DATABASE_URL=postgres://user:password@host:5432/mud_core?sslmode=require

# Redis
REDIS_ADDR=redis-host:6379

# NATS
NATS_URL=nats://nats-host:4222

# Ollama
OLLAMA_HOST=http://ollama-host:11434

# Security
JWT_SECRET=<generate-strong-secret>

# Server
PORT=8080
```

### Build Production Images

```bash
# Build game server
docker build -f Dockerfile.game-server -t thousand-worlds/game-server:latest .

# Tag for registry
docker tag thousand-worlds/game-server:latest registry.example.com/game-server:latest

# Push to registry
docker push registry.example.com/game-server:latest
```

### Deploy with Docker Compose

```bash
# Production deployment
docker-compose -f docker-compose.prod.yml up -d
```

### Deploy to Cloud (AWS/GCP/Azure)

**Using Docker:**
1. Push images to container registry (ECR, GCR, ACR)
2. Deploy to container service (ECS, Cloud Run, ACI)
3. Configure load balancer for WebSocket sticky sessions
4. Set up managed database (RDS, Cloud SQL, Azure Database)
5. Configure managed Redis (ElastiCache, Memorystore, Azure Cache)

**Using Kubernetes:**
(See kubernetes/ directory for manifests)

```bash
kubectl apply -f kubernetes/
```

---

## Health Checks

### Game Server
```bash
curl http://localhost:8080/health
```

### PostgreSQL
```bash
docker exec mud_postgis pg_isready -U admin
```

### Redis
```bash
docker exec mud_redis redis-cli ping
```

### NATS
```bash
curl http://localhost:8222/healthz
```

---

## Monitoring

### Prometheus Metrics
Available at: `http://localhost:9090/metrics`

### Grafana Dashboards
1. Access Grafana (if configured)
2. Import dashboards from `/deploy/grafana/dashboards/`

---

## Backup & Restore

### Database Backup
```bash
docker exec mud_postgis pg_dump -U admin mud_core > backup.sql
```

### Database Restore
```bash
docker exec -i mud_postgis psql -U admin mud_core < backup.sql
```

### Volume Backup
```bash
docker run --rm -v mud_postgres_data:/data -v $(pwd):/backup alpine tar czf /backup/postgres-backup.tar.gz /data
```

---

## Scaling

### Horizontal Scaling

1. **Game Server**: Multiple instances with load balancer
   - Configure WebSocket sticky sessions
   - Share state via Redis
   - Coordinate via NATS

2. **Database**: Read replicas for queries
   - Write to primary
   - Read from replicas
   - Use PgBouncer for connection pooling

3. **Redis**: Redis Cluster or Sentinel
   - High availability
   - Automatic failover

---

## Troubleshooting

### Common Issues

**Port Conflicts:**
```bash
# Check what's using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>
```

**Database Connection:**
```bash
# Test connection
docker exec mud_postgis psql -U admin -d mud_core -c "SELECT 1;"
```

**Redis Connection:**
```bash
# Test connection
docker exec mud_redis redis-cli ping
```

**Ollama Model Missing:**
```bash
# List models
docker exec mud_ollama ollama list

# Pull model
docker exec mud_ollama ollama pull llama3.1:8b
```

---

## Security Checklist

- [ ] Change default passwords in production
- [ ] Use strong JWT_SECRET
- [ ] Enable SSL/TLS for database connections
- [ ] Configure firewall rules
- [ ] Set up VPC/private network
- [ ] Enable database encryption at rest
- [ ] Configure CORS properly
- [ ] Implement rate limiting
- [ ] Set up monitoring and alerts
- [ ] Regular security updates
- [ ] Backup procedures tested

---

## Performance Tuning

### PostgreSQL
```sql
-- Optimize for spatial queries
CREATE INDEX CONCURRENTLY idx_characters_position_spatial 
ON characters USING GIST (position);

-- Connection pooling
ALTER SYSTEM SET max_connections = 200;
ALTER SYSTEM SET shared_buffers = '2GB';
```

### Redis
```bash
# In redis.conf
maxmemory 2gb
maxmemory-policy allkeys-lru
```

### NATS
```bash
# Increase JetStream storage
--store_dir /data --max_file_store 10GB
```

---

## Maintenance

### Update LLM Model
```bash
docker exec mud_ollama ollama pull llama3.1:8b
```

### Database Migrations
```bash
go run cmd/migrate/main.go
```

### Clear Redis Cache
```bash
docker exec mud_redis redis-cli FLUSHDB
```

### View Logs
```bash
# Game server
docker logs -f mud_game_server

# PostgreSQL
docker logs -f mud_postgis

# All services
docker-compose logs -f
```
