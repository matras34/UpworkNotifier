# Web SSH Service - Production-Ready SaaS Platform

A production-ready SaaS Web SSH service that allows users to connect to their SSH servers through a browser-based terminal.

## Features

### Functional
- ✅ Browser-based terminal using xterm.js
- ✅ WebSocket-based real-time communication
- ✅ SSH connection with password and private key authentication
- ✅ Host key verification (manual approval required)
- ✅ Session timeout and idle connection cleanup
- ✅ Comprehensive audit logging
- ✅ Multi-tenant architecture
- ✅ Role-based access control (RBAC)
- ✅ Saved connection management

### Security
- ✅ KMS-like encryption for credentials
- ✅ No plaintext credential storage
- ✅ Brute force protection
- ✅ Rate limiting on all endpoints
- ✅ SSH session isolation
- ✅ Secure WebSocket handling
- ✅ Command injection prevention
- ✅ Resource limits per session

## Tech Stack

- **Backend**: Go 1.21+
- **Frontend**: React 18 + xterm.js
- **Database**: PostgreSQL 14+
- **Cache**: Redis 7+
- **Deployment**: Docker + Docker Compose

## Project Structure

```
.
├── backend/
│   ├── cmd/server/         # Application entry point
│   ├── internal/
│   │   ├── api/           # HTTP handlers and routing
│   │   ├── auth/          # Authentication and RBAC
│   │   ├── cache/         # Redis client
│   │   ├── config/        # Configuration management
│   │   ├── crypto/        # KMS encryption
│   │   ├── db/            # Database models and queries
│   │   ├── ssh/           # SSH client and session management
│   │   └── websocket/     # WebSocket hub and clients
│   └── pkg/logger/        # Structured logging
├── frontend/
│   ├── src/
│   │   ├── components/    # React components
│   │   └── services/      # API services
│   └── public/
├── database/
│   └── schema.sql         # PostgreSQL schema
└── docker-compose.yml     # Docker orchestration

```

## Quick Start

### Prerequisites

- Docker 20.10+
- Docker Compose 2.0+
- (Optional) Go 1.21+ for local development
- (Optional) Node.js 18+ for local development

### Deployment with Docker Compose

1. **Clone the repository**
```bash
git clone <repository-url>
cd UpworkNotifier
```

2. **Configure environment variables**
```bash
# Backend
cp backend/.env.example backend/.env
# Edit backend/.env and set secure values for:
# - JWT_SECRET
# - ENCRYPTION_KEY (must be 32+ characters)

# Frontend
cp frontend/.env.example frontend/.env
```

3. **Start all services**
```bash
docker-compose up -d
```

4. **Initialize database** (auto-initialized on first start)
The database schema is automatically applied via `docker-entrypoint-initdb.d`.

5. **Access the application**
- Frontend: http://localhost
- Backend API: http://localhost:8080
- Default credentials: `admin@example.com` / `admin123`

6. **Verify services**
```bash
docker-compose ps
docker-compose logs -f backend
```

### Manual Deployment

#### Backend

```bash
cd backend

# Install dependencies
go mod download

# Set environment variables
cp .env.example .env
# Edit .env

# Run
go run cmd/server/main.go
```

#### Frontend

```bash
cd frontend

# Install dependencies
npm install

# Development mode
npm start

# Production build
npm run build
```

#### Database Setup

```bash
# Connect to PostgreSQL
psql -h localhost -U postgres -d webssh

# Run schema
\i database/schema.sql
```

## API Documentation

### Authentication

**POST** `/api/v1/auth/register`
```json
{
  "email": "user@example.com",
  "password": "securepassword"
}
```

**POST** `/api/v1/auth/login`
```json
{
  "email": "user@example.com",
  "password": "securepassword"
}
```
Response:
```json
{
  "token": "jwt-token",
  "user": { "id": "...", "email": "...", "role": "user" }
}
```

### SSH Connections

**GET** `/api/v1/connections`
Headers: `Authorization: Bearer <token>`

**POST** `/api/v1/connections`
```json
{
  "name": "My Server",
  "host": "192.168.1.100",
  "port": 22,
  "username": "root",
  "auth_type": "password",
  "password": "secret"
}
```

**DELETE** `/api/v1/connections?id=<connection-id>`

### WebSocket SSH Connection

**WS** `/api/v1/ssh/connect`
Query params: `token=<jwt-token>`

Message format:
```json
{
  "type": "data|resize|hostkey|error|close|connected",
  "data": "..."
}
```

## Configuration

### Backend Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| SERVER_PORT | 8080 | HTTP server port |
| DB_HOST | localhost | PostgreSQL host |
| DB_PORT | 5432 | PostgreSQL port |
| DB_USER | postgres | Database user |
| DB_PASSWORD | postgres | Database password |
| DB_NAME | webssh | Database name |
| REDIS_HOST | localhost | Redis host |
| REDIS_PORT | 6379 | Redis port |
| JWT_SECRET | - | JWT signing key (required) |
| ENCRYPTION_KEY | - | Master encryption key (32+ chars) |
| SSH_MAX_CONCURRENT_SESSIONS | 1000 | Max concurrent SSH sessions |
| SSH_SESSION_IDLE_TIMEOUT | 30m | Idle timeout |
| SSH_SESSION_MAX_DURATION | 4h | Maximum session duration |
| RATE_LIMIT_PER_MINUTE | 100 | API rate limit per user |

## Scaling to 10,000+ Concurrent Sessions

### Architecture for Scale

```
                    [Load Balancer (NGINX/HAProxy)]
                                |
        ┌──────────────┬───────────────┬──────────────┐
        ▼              ▼               ▼              ▼
   [Backend 1]   [Backend 2]   [Backend 3]  ... [Backend 10]
   (1k sessions) (1k sessions) (1k sessions)   (1k sessions)
        │              │               │              │
        └──────────────┴───────────────┴──────────────┘
                                |
                ┌───────────────┼───────────────┐
                ▼               ▼               ▼
         [PostgreSQL]    [Redis Cluster]  [Redis Cluster]
          + Replicas     (Sessions)       (Rate Limiting)
```

### Scaling Strategy

#### 1. Horizontal Scaling
- **10 backend instances** @ 1,000 sessions each
- Each instance: 4 vCPU, 8GB RAM
- Session affinity via load balancer (sticky sessions)

#### 2. Database Optimization
- **PostgreSQL**:
  - Master-replica setup (1 master + 2 read replicas)
  - Connection pooling: max 100 per instance
  - Partition `audit_logs` table by month
  - Indexes on `user_id`, `created_at`, `host`
  
```sql
-- Example partitioning
CREATE TABLE audit_logs_2024_01 PARTITION OF audit_logs
FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
```

#### 3. Redis Optimization
- **Separate Redis clusters**:
  - Cluster 1: Session data (persistent, AOF enabled)
  - Cluster 2: Rate limiting (ephemeral, no persistence)
- Redis Cluster mode for sharding (3 masters + 3 replicas)

#### 4. Resource Management Per Instance
```go
MaxConcurrentSessions: 1000
MaxBufferSizePerSession: 10MB (10 * 1024 * 1024)
SessionIdleTimeout: 30min
SessionMaxDuration: 4hours
TCPKeepalive: 60s
```

#### 5. Network Optimization
- WebSocket compression enabled
- Binary frame protocol for efficiency
- Connection pooling to common SSH destinations
- TCP keepalive tuning: 60s intervals

### Infrastructure Requirements

**For 10,000 concurrent sessions:**

| Component | Count | Specs | Notes |
|-----------|-------|-------|-------|
| Backend | 10 | 4 vCPU, 8GB RAM | 1k sessions each |
| PostgreSQL Master | 1 | 8 vCPU, 16GB RAM | Primary writes |
| PostgreSQL Replica | 2 | 4 vCPU, 8GB RAM | Read queries |
| Redis Session | 3 | 4 vCPU, 8GB RAM | Session data cluster |
| Redis Rate Limit | 3 | 2 vCPU, 4GB RAM | Rate limiting cluster |
| Load Balancer | 2 | 2 vCPU, 4GB RAM | HA pair |

**Estimated costs (AWS):**
- ~$800-1200/month for 10k concurrent sessions

### Performance Bottlenecks

#### 1. Network I/O (Primary Bottleneck)
- **Impact**: Each SSH session generates network traffic
- **Solution**: 
  - Multiple backend instances with dedicated network bandwidth
  - Optimize buffer sizes (8KB-16KB per session)
  - Use connection pooling for frequently accessed hosts

#### 2. SSH Connection Overhead
- **Impact**: Each SSH handshake takes 100-500ms
- **Solution**:
  - Connection pooling for common destinations
  - Keep-alive connections to reduce handshakes
  - Parallel connection establishment

#### 3. Database Write Load
- **Impact**: Audit logs generate heavy write load
- **Solution**:
  - Async batch inserts (every 5 seconds)
  - Partition tables by date
  - Use read replicas for historical queries

#### 4. Memory Consumption
- **Impact**: 10MB buffer per session = 100GB for 10k sessions
- **Solution**:
  - Limit buffer sizes strictly
  - Use memory-efficient data structures
  - Implement graceful degradation

#### 5. Redis Performance
- **Impact**: Rate limit checks on every request
- **Solution**:
  - Local in-memory cache (1-second TTL)
  - Separate Redis clusters by use case
  - Use Redis pipelining

### Monitoring

```bash
# Prometheus metrics endpoints
GET /metrics

# Key metrics to monitor:
- active_ssh_sessions
- websocket_connections
- database_connection_pool_usage
- redis_command_latency
- http_request_duration
- session_creation_rate
- session_error_rate
```

## Adding Ephemeral SSH Certificates

### Architecture Changes

```
                  [User Request]
                       |
                       ▼
              [Backend API + CA]
                       |
                       ▼
          [Generate Short-Lived Cert]
             (TTL: 1 hour, signed by CA)
                       |
                       ▼
         [Use Cert for SSH Connection]
                       |
                       ▼
         [Auto-Expiration, No Storage]
```

### Implementation Steps

#### 1. Add Certificate Authority Service

```go
// backend/internal/ca/ca.go
package ca

import (
    "crypto/rand"
    "crypto/rsa"
    "golang.org/x/crypto/ssh"
    "time"
)

type CertificateAuthority struct {
    signer ssh.Signer
}

func NewCA(privateKeyPath string) (*CertificateAuthority, error) {
    // Load CA private key
    privateKey, err := loadPrivateKey(privateKeyPath)
    if err != nil {
        return nil, err
    }

    signer, err := ssh.NewSignerFromKey(privateKey)
    if err != nil {
        return nil, err
    }

    return &CertificateAuthority{signer: signer}, nil
}

func (ca *CertificateAuthority) IssueCertificate(username string, principals []string, ttl time.Duration) (*ssh.Certificate, error) {
    // Generate user key pair
    privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
    if err != nil {
        return nil, err
    }

    publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
    if err != nil {
        return nil, err
    }

    // Create certificate
    cert := &ssh.Certificate{
        Key:             publicKey,
        CertType:        ssh.UserCert,
        KeyId:           username,
        ValidPrincipals: principals,
        ValidAfter:      uint64(time.Now().Unix()),
        ValidBefore:     uint64(time.Now().Add(ttl).Unix()),
        Permissions: ssh.Permissions{
            Extensions: map[string]string{
                "permit-pty": "",
            },
        },
    }

    // Sign certificate
    if err := cert.SignCert(rand.Reader, ca.signer); err != nil {
        return nil, err
    }

    return cert, nil
}
```

#### 2. API Endpoint for Certificate Issuance

```go
// Add to backend/internal/api/handlers/ssh.go
func (h *SSHHandler) IssueCertificate(w http.ResponseWriter, r *http.Request) {
    claims := middleware.GetUserFromContext(r.Context())
    if claims == nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    var req struct {
        Host       string   `json:"host"`
        Username   string   `json:"username"`
        Principals []string `json:"principals"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // Issue certificate (1 hour TTL)
    cert, privateKey, err := h.ca.IssueCertificate(
        req.Username,
        req.Principals,
        1*time.Hour,
    )
    if err != nil {
        http.Error(w, "Failed to issue certificate", http.StatusInternalServerError)
        return
    }

    // Return certificate and private key
    json.NewEncoder(w).Encode(map[string]string{
        "certificate": string(ssh.MarshalAuthorizedKey(cert)),
        "private_key": encodePrivateKey(privateKey),
        "expires_at":  time.Now().Add(1 * time.Hour).Format(time.RFC3339),
    })
}
```

#### 3. Update SSH Client to Use Certificates

```go
// backend/internal/ssh/client.go - Add certificate auth
func NewSSHClientWithCert(host string, port int, username string, cert *ssh.Certificate, privateKey *rsa.PrivateKey, hostKeyCallback ssh.HostKeyCallback) (*SSHClient, error) {
    signer, err := ssh.NewSignerFromKey(privateKey)
    if err != nil {
        return nil, err
    }

    certSigner, err := ssh.NewCertSigner(cert, signer)
    if err != nil {
        return nil, err
    }

    config := &ssh.ClientConfig{
        User: username,
        Auth: []ssh.AuthMethod{
            ssh.PublicKeys(certSigner),
        },
        HostKeyCallback: hostKeyCallback,
        Timeout:         10 * time.Second,
    }

    // ... rest of connection logic
}
```

#### 4. Benefits

✅ **No credential storage**: Certificates are ephemeral
✅ **Automatic expiration**: 1-hour TTL, no manual revocation needed
✅ **Centralized access control**: CA controls who gets certificates
✅ **Audit trail**: All certificate issuances logged
✅ **Revocation support**: Can revoke CA or implement CRL

#### 5. Server-Side Configuration

SSH servers need to trust the CA:

```bash
# On SSH server
# /etc/ssh/sshd_config
TrustedUserCAKeys /etc/ssh/ca.pub

# Restart SSH
sudo systemctl restart sshd
```

## Security Best Practices

1. **Credentials**
   - Never use default passwords in production
   - Rotate JWT secrets regularly
   - Use strong encryption keys (32+ characters, random)

2. **Network**
   - Use TLS/SSL for all connections (HTTPS/WSS)
   - Configure firewall rules (allow only necessary ports)
   - Use VPN for database access

3. **Database**
   - Enable SSL mode for PostgreSQL
   - Use strong passwords
   - Regular backups

4. **Monitoring**
   - Enable audit logging
   - Monitor failed login attempts
   - Set up alerts for anomalies

5. **Updates**
   - Keep dependencies up to date
   - Regular security patches
   - Vulnerability scanning

## Development

### Backend

```bash
cd backend

# Run tests
go test ./...

# Format code
go fmt ./...

# Lint
golangci-lint run
```

### Frontend

```bash
cd frontend

# Run tests
npm test

# Lint
npm run lint

# Build
npm run build
```

## Troubleshooting

### Backend won't start
```bash
# Check logs
docker-compose logs backend

# Verify database connection
docker-compose exec postgres psql -U postgres -d webssh -c "SELECT 1"

# Verify Redis connection
docker-compose exec redis redis-cli ping
```

### WebSocket connection fails
- Check CORS configuration in backend
- Verify WebSocket upgrade headers in nginx
- Check browser console for errors

### SSH connection fails
- Verify SSH server is accessible from backend
- Check host key verification
- Verify credentials are correct

## License

MIT

## Support

For issues and questions, please open a GitHub issue.
