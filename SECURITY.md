# Thousand Worlds - Security & Secrets Management

This document explains how to configure credentials and secrets for the Thousand Worlds game.

## Quick Start

1. **Copy the environment template:**
   ```bash
   cp .env.example .env
   ```

2.  **Generate secure secrets:**
   ```bash
   # Generate JWT secret (32+ characters)
   echo "JWT_SECRET=$(openssl rand -hex 32)" >> .env
   
   # Generate database password
   echo "POSTGRES_PASSWORD=$(openssl rand -hex 16)" >> .env
   
   # Generate MongoDB password  
   echo "MONGO_PASSWORD=$(openssl rand -hex 16)" >> .env
   ```

3. **Configure CORS (optional for production):**
   ```bash
   # TODO_SECURITY: Update when you have a production domain
   echo "CORS_ALLOWED_ORIGINS=http://localhost:5173" >> .env
   ```

4. **Validate your configuration:**
   ```bash
   cd mud-platform-backend
   ./scripts/validate-env.sh
   ```

5. **Launch the application:**
   ```bash
   ./launch.sh
   ```

## Environment Variables

### Required Secrets

These MUST be set for the application to start:

| Variable | Description | Generate With |
|----------|-------------|---------------|
| `JWT_SECRET` | Secret key for JWT token signing | `openssl rand -hex 32` |
| `POSTGRES_PASSWORD` | PostgreSQL database password | `openssl rand -hex 16` |
| `MONGO_PASSWORD` | MongoDB password | `openssl rand -hex 16` |

### Optional Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `CORS_ALLOWED_ORIGINS` | `http://localhost:5173` | Comma-separated list of allowed origins |
| `POSTGRES_USER` | `admin` | PostgreSQL username |
| `POSTGRES_DB` | `mud_core` | PostgreSQL database name |
| `PORT` | `8080` | Backend server port |
| `REDIS_ADDR` | `localhost:6379` | Redis connection address |

## Security Best Practices

### ✅ DO:
- Generate unique, strong secrets for each environment
- Use `.env` files for local development
- Use secret management systems (Vault, AWS Secrets Manager) in production
- Rotate secrets regularly (every 90 days minimum)
- Never commit `.env` files to git
- Use SSL/TLS for database connections in production

### ❌ DON'T:
- Use the default/example secrets in production
- Share secrets via email or chat
- Commit secrets to version control
- Use the same secrets across environments
- Hard-code secrets in application code
- Use weak or short secrets (< 32 characters for JWT)

## CORS Configuration

### Development
For local development with multiple devices:
```bash
CORS_ALLOWED_ORIGINS=http://localhost:5173,http://192.168.0.0/16
```

### Production
```bash
# TODO_SECURITY: Replace with your actual production domains
# This placeholder should be updated when deploying to production
CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
```

> **⚠️ WARNING:**  Never use `*` (wildcard) for `CORS_ALLOWED_ORIGINS` in production!

## Deployment

### Pre-Deployment Checklist
Before deploying to production:

1. Run the validation script:
   ```bash
   ./mud-platform-backend/scripts/validate-env.sh
   ```

2. Verify no hardcoded credentials:
   ```bash
   git grep -n "password123\|secret.*=.*[\"'][^$]" --and --not -e ".md" --not -e "template"
   ```

3. Confirm secrets are strong:
   - JWT_SECRET: ≥ 32 characters
   - Passwords: ≥ 12 characters
   - All randomly generated (not dictionary words)

4. Update CORS configuration for production domains

5. Enable SSL/TLS for database connections:
   ```bash
   DATABASE_URL="postgres://user:pass@host:5432/db?sslmode=require"
   ```

## Troubleshooting

### Application fails to start with "JWT_SECRET must be set"
- Ensure you've created a `.env` file with all required secrets
- Verify the `.env` file is being loaded (check `launch.sh` output)
- Try running with explicit export: `export JWT_SECRET=...`

### Docker Compose fails with environment variable errors
- Make sure `.env` file is in the project root or `mud-platform-backend/` directory
- Verify no syntax errors in `.env` (no spaces around `=`)
- Check Docker Compose version supports variable substitution

### "CORS policy" errors in browser console
- Add your frontend URL to `CORS_ALLOWED_ORIGINS`
- For mobile testing, include your local network IP range
- Restart the backend after changing CORS configuration

## Secret Rotation

To rotate secrets:

1. Generate new secrets:
   ```bash
   NEW_JWT_SECRET=$(openssl rand -hex 32)
   NEW_POSTGRES_PASSWORD=$(openssl rand -hex 16)
   ```

2. Update deployment:
   - For Docker: Update environment variables and restart containers
   - For bare metal: Update `.env` and restart services

3. Update backups to use new credentials

4. Revoke old credentials after confirming new ones work

## Additional Resources

- [OWASP Secrets Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html)
- [Docker Secrets Documentation](https://docs.docker.com/engine/swarm/secrets/)
- See `.env.template` for full list of configuration options
