# Web SSH Service - Quick Deployment Guide

## Quick Start (5 minutes)

### Prerequisites
- Docker and Docker Compose installed
- At least 2GB RAM available

### Steps

1. **Clone and navigate to the project**
```bash
git clone <your-repo-url>
cd UpworkNotifier
```

2. **Start the services**
```bash
docker-compose up -d
```

3. **Access the application**
- Open browser: http://localhost
- Login with default credentials:
  - Email: `admin@example.com`
  - Password: `admin123`

4. **Create your first SSH connection**
- Click "New Connection"
- Fill in your SSH server details
- Click "Connect"

That's it! 🎉

## Architecture Overview

```
[Browser] ←WebSocket→ [Frontend:80] → [Backend:8080] ←SSH→ [Your SSH Server]
                           ↓              ↓
                    [PostgreSQL:5432] [Redis:6379]
```

## Key Features

✅ Browser-based SSH terminal (xterm.js)  
✅ Password & Private Key authentication  
✅ Host key verification  
✅ Session timeout & management  
✅ Encrypted credential storage  
✅ Multi-user support  
✅ Role-based access control  
✅ Rate limiting  
✅ Audit logging  

## Configuration

### Security Settings (IMPORTANT!)

Before production deployment, update these in `docker-compose.yml`:

```yaml
environment:
  - JWT_SECRET=your-random-secret-here-change-this
  - ENCRYPTION_KEY=your-32-char-encryption-key-here!
```

Generate secure keys:
```bash
# JWT Secret (any length)
openssl rand -base64 32

# Encryption Key (32+ chars)
openssl rand -base64 32
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| JWT_SECRET | change-me | JWT signing key |
| ENCRYPTION_KEY | change-me | Credential encryption key (32+ chars) |
| SSH_MAX_CONCURRENT_SESSIONS | 1000 | Max concurrent SSH sessions |
| SSH_SESSION_IDLE_TIMEOUT | 30m | Auto-disconnect idle sessions |
| RATE_LIMIT_PER_MINUTE | 100 | API rate limit per user |

## Production Deployment

### 1. Using Docker Compose (Recommended)

```bash
# Update docker-compose.yml with production values
vim docker-compose.yml

# Start services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### 2. Manual Deployment

#### Backend (Go)
```bash
cd backend
cp .env.example .env
# Edit .env with your values
go build -o webssh ./cmd/server
./webssh
```

#### Frontend (React)
```bash
cd frontend
cp .env.example .env
npm install
npm run build
# Serve build/ with nginx or any web server
```

#### Database
```bash
psql -U postgres -d webssh -f database/schema.sql
```

## Scaling

For 10,000 concurrent sessions:
- 10 backend instances (1k sessions each)
- 1 PostgreSQL master + 2 replicas
- 2 Redis clusters (sessions + rate limiting)
- Load balancer (NGINX/HAProxy)

See `README_WEBSSH.md` for detailed scaling architecture.

## Troubleshooting

### Services won't start
```bash
docker-compose ps
docker-compose logs backend
docker-compose logs postgres
```

### Can't connect to SSH server
- Verify SSH server is reachable from backend container
- Check firewall rules
- Test with: `docker-compose exec backend sh -c "nc -zv your-server 22"`

### WebSocket connection fails
- Check CORS settings in backend
- Verify reverse proxy WebSocket config
- Check browser console for errors

### Database connection errors
```bash
docker-compose exec postgres psql -U postgres -d webssh -c "SELECT 1"
```

## Default Credentials

**⚠️ CHANGE IMMEDIATELY IN PRODUCTION**

- Email: `admin@example.com`
- Password: `admin123`

## Backup

### Database
```bash
docker-compose exec postgres pg_dump -U postgres webssh > backup.sql
```

### Restore
```bash
docker-compose exec -T postgres psql -U postgres webssh < backup.sql
```

## Monitoring

Health check endpoint:
```bash
curl http://localhost:8080/health
```

View active sessions:
```sql
SELECT * FROM ssh_sessions WHERE status = 'active';
```

## Security Checklist

- [ ] Changed default admin password
- [ ] Set strong JWT_SECRET
- [ ] Set strong ENCRYPTION_KEY (32+ chars)
- [ ] Enabled PostgreSQL SSL
- [ ] Configured firewall rules
- [ ] Set up HTTPS/TLS
- [ ] Enabled audit logging
- [ ] Regular backups configured
- [ ] Monitoring and alerts set up

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/api/v1/auth/register` | POST | Register new user |
| `/api/v1/auth/login` | POST | Login |
| `/api/v1/connections` | GET | List connections |
| `/api/v1/connections` | POST | Save connection |
| `/api/v1/connections` | DELETE | Delete connection |
| `/api/v1/ssh/connect` | WS | WebSocket SSH connection |

## Support

- Documentation: `README_WEBSSH.md`
- Architecture: `ARCHITECTURE.md`
- Issues: GitHub Issues

## License

MIT
