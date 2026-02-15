# Web SSH Service Architecture

## System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         Load Balancer (NGINX)                    │
│                      (SSL Termination, Rate Limit)               │
└─────────────────────┬───────────────────────────────────────────┘
                      │
         ┌────────────┴────────────┐
         │                         │
┌────────▼────────┐       ┌────────▼────────┐
│  Backend API    │       │  Backend API    │  (Horizontal Scaling)
│  (Go Server)    │       │  (Go Server)    │
│  - REST API     │       │  - REST API     │
│  - WebSocket    │       │  - WebSocket    │
│  - SSH Handler  │       │  - SSH Handler  │
└────────┬────────┘       └────────┬────────┘
         │                         │
         └────────────┬────────────┘
                      │
         ┌────────────┴─────────────┐
         │                          │
┌────────▼────────┐       ┌─────────▼────────┐
│   PostgreSQL    │       │      Redis       │
│   - Users       │       │   - Sessions     │
│   - Connections │       │   - Rate Limit   │
│   - Audit Logs  │       │   - Cache        │
└─────────────────┘       └──────────────────┘
```

## Component Architecture

### 1. Frontend (React + xterm.js)
- **Terminal Component**: xterm.js integration
- **Connection Form**: Host, port, credentials input
- **Session Manager**: Active connections display
- **Host Key Verification**: Fingerprint confirmation UI
- **Auth**: JWT-based authentication

### 2. Backend (Go)
```
backend/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── api/
│   │   ├── handlers/            # HTTP/WebSocket handlers
│   │   ├── middleware/          # Auth, rate limiting, logging
│   │   └── router.go            # Route definitions
│   ├── ssh/
│   │   ├── client.go            # SSH client wrapper
│   │   ├── session.go           # SSH session management
│   │   └── hostkey.go           # Host key verification
│   ├── websocket/
│   │   ├── hub.go               # WebSocket connection hub
│   │   └── client.go            # WebSocket client handler
│   ├── auth/
│   │   ├── jwt.go               # JWT token handling
│   │   ├── rbac.go              # Role-based access control
│   │   └── middleware.go        # Auth middleware
│   ├── crypto/
│   │   └── kms.go               # Key encryption/decryption
│   ├── db/
│   │   ├── postgres.go          # PostgreSQL connection
│   │   ├── models/              # Data models
│   │   └── migrations/          # SQL migrations
│   ├── cache/
│   │   └── redis.go             # Redis operations
│   └── config/
│       └── config.go            # Configuration management
└── pkg/
    └── logger/                  # Structured logging
```

### 3. Database Schema (PostgreSQL)

**users**
- id (PK)
- email (unique)
- password_hash
- role (admin, user)
- organization_id (FK)
- created_at, updated_at

**organizations**
- id (PK)
- name
- max_concurrent_sessions
- created_at

**ssh_connections**
- id (PK)
- user_id (FK)
- name
- host
- port
- username
- auth_type (password/key)
- encrypted_password
- encrypted_private_key
- created_at, updated_at

**ssh_sessions**
- id (PK)
- user_id (FK)
- connection_id (FK)
- client_ip
- started_at
- ended_at
- status (active/closed)

**audit_logs**
- id (PK)
- user_id (FK)
- action
- resource_type
- resource_id
- ip_address
- timestamp
- details (JSONB)

**host_keys**
- id (PK)
- connection_id (FK)
- host
- port
- algorithm
- fingerprint
- verified_at
- verified_by_user_id (FK)

### 4. Security Layers

#### Encryption (KMS-like)
```go
// Master key from environment (rotatable)
// Individual keys encrypted with AES-256-GCM
// Salt and nonce stored per record
```

#### Rate Limiting
```
- 10 connections/minute per user
- 100 API requests/minute per user
- Redis-based token bucket
```

#### Session Isolation
```
- Each SSH session in goroutine
- Resource limits per session:
  - Max 10MB buffer
  - 30min idle timeout
  - Max 4 hours total
```

#### WebSocket Security
```
- JWT authentication required
- Origin validation
- Message size limits (1MB)
- Command injection prevention
```

## Data Flow

### SSH Connection Establishment
```
1. User submits connection details via UI
2. Backend validates credentials and RBAC
3. Backend retrieves encrypted credentials from DB
4. Backend decrypts credentials using KMS
5. Backend initiates SSH connection
6. SSH server presents host key
7. Backend checks if host key is known
8. If new: Send fingerprint to frontend for approval
9. User approves -> Store in DB, continue connection
10. WebSocket tunnel established
11. Terminal I/O proxied through WebSocket
12. Session logged to DB
```

### Session Lifecycle
```
WebSocket Connection
    ↓
JWT Validation
    ↓
Rate Limit Check
    ↓
Create SSH Session
    ↓
Authenticate to SSH Server
    ↓
Host Key Verification
    ↓
Proxy Terminal I/O
    ↓
Monitor Activity (Timeout Check)
    ↓
Close Session (Cleanup)
    ↓
Audit Log
```

## Scaling Strategy

### For 10,000 Concurrent Sessions

#### 1. Horizontal Scaling
- 10 backend instances @ 1,000 sessions each
- Each instance: 4 CPU cores, 8GB RAM
- Session affinity via load balancer

#### 2. Database Optimization
- Connection pooling (max 100 per instance)
- Read replicas for audit logs
- Partitioning audit_logs by month
- Indexes on user_id, created_at, host

#### 3. Redis Optimization
- Separate Redis clusters:
  - Session data (persistent)
  - Rate limiting (ephemeral)
- Redis Cluster for sharding

#### 4. Resource Management
```go
// Per instance limits
MaxConcurrentSessions: 1000
MaxBufferSizePerSession: 10MB
SessionIdleTimeout: 30min
SessionMaxDuration: 4hours
```

#### 5. Network Optimization
- WebSocket compression
- Binary frame protocol
- Connection pooling to common SSH hosts
- TCP keepalive tuning

### Bottlenecks

1. **Network I/O**: Primary bottleneck
   - Solution: Multiple backend instances, optimize buffers

2. **SSH Connection Overhead**: Each SSH handshake takes time
   - Solution: Connection pooling, warm connections

3. **Database**: Audit logging at scale
   - Solution: Async logging, batching, partitioning

4. **Memory**: Each session holds buffers
   - Solution: Limit buffer sizes, graceful degradation

5. **Redis**: Rate limit checks
   - Solution: Local caching, Redis Cluster

## Security Best Practices

1. **Never store plaintext credentials**
2. **Encrypt all private keys with KMS**
3. **Implement strict RBAC**
4. **Log all SSH commands (optional, compliance)**
5. **Regular key rotation**
6. **Host key pinning**
7. **No auto-accept of unknown hosts**
8. **Input sanitization for all user inputs**
9. **Rate limiting on all endpoints**
10. **TLS 1.3 only for WebSocket**

## Future Enhancements

### Ephemeral SSH Certificates

```
Architecture Changes:
1. Add Certificate Authority (CA) component
2. Integrate with HashiCorp Vault or custom CA
3. Flow:
   - User requests access to server
   - Backend generates short-lived cert (1 hour TTL)
   - Signed by CA
   - Used for SSH authentication
   - Auto-expires, no credential storage needed

Benefits:
- No credential storage
- Automatic expiration
- Centralized access control
- Audit trail built-in
- Revocation support

Implementation:
- Add CA service (Vault or golang.org/x/crypto/ssh)
- Modify SSH client to use certs
- Add cert issuance API endpoint
- Update host key management for CA trust
```

### Additional Features
- Session recording and playback
- Collaborative sessions (screen sharing)
- File transfer (SCP/SFTP)
- Bastion/jump host support
- MFA for sensitive operations
- Integration with Teleport/Boundary
- Kubernetes exec support
