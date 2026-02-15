package models

import (
	"time"

	"github.com/google/uuid"
)

type Organization struct {
	ID                     uuid.UUID `json:"id"`
	Name                   string    `json:"name"`
	MaxConcurrentSessions  int       `json:"max_concurrent_sessions"`
	MaxSavedConnections    int       `json:"max_saved_connections"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

type User struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	Email          string     `json:"email"`
	PasswordHash   string     `json:"-"`
	Role           string     `json:"role"`
	IsActive       bool       `json:"is_active"`
	LastLoginAt    *time.Time `json:"last_login_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type SSHConnection struct {
	ID                  uuid.UUID `json:"id"`
	UserID              uuid.UUID `json:"user_id"`
	Name                string    `json:"name"`
	Host                string    `json:"host"`
	Port                int       `json:"port"`
	Username            string    `json:"username"`
	AuthType            string    `json:"auth_type"`
	EncryptedPassword   *string   `json:"-"`
	EncryptedPrivateKey *string   `json:"-"`
	EncryptionSalt      *string   `json:"-"`
	EncryptionNonce     *string   `json:"-"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type HostKey struct {
	ID                uuid.UUID `json:"id"`
	ConnectionID      *uuid.UUID `json:"connection_id,omitempty"`
	Host              string    `json:"host"`
	Port              int       `json:"port"`
	Algorithm         string    `json:"algorithm"`
	PublicKey         string    `json:"public_key"`
	FingerprintSHA256 string    `json:"fingerprint_sha256"`
	FingerprintMD5    *string   `json:"fingerprint_md5,omitempty"`
	VerifiedAt        time.Time `json:"verified_at"`
	VerifiedByUserID  uuid.UUID `json:"verified_by_user_id"`
	CreatedAt         time.Time `json:"created_at"`
}

type SSHSession struct {
	ID               uuid.UUID  `json:"id"`
	UserID           uuid.UUID  `json:"user_id"`
	ConnectionID     *uuid.UUID `json:"connection_id,omitempty"`
	SessionToken     string     `json:"session_token"`
	ClientIP         string     `json:"client_ip"`
	Host             string     `json:"host"`
	Port             int        `json:"port"`
	Username         string     `json:"username"`
	Status           string     `json:"status"`
	StartedAt        time.Time  `json:"started_at"`
	LastActivityAt   time.Time  `json:"last_activity_at"`
	EndedAt          *time.Time `json:"ended_at,omitempty"`
	DisconnectReason *string    `json:"disconnect_reason,omitempty"`
	BytesSent        int64      `json:"bytes_sent"`
	BytesReceived    int64      `json:"bytes_received"`
	CreatedAt        time.Time  `json:"created_at"`
}

type AuditLog struct {
	ID             uuid.UUID              `json:"id"`
	UserID         *uuid.UUID             `json:"user_id,omitempty"`
	OrganizationID *uuid.UUID             `json:"organization_id,omitempty"`
	Action         string                 `json:"action"`
	ResourceType   string                 `json:"resource_type"`
	ResourceID     *uuid.UUID             `json:"resource_id,omitempty"`
	IPAddress      *string                `json:"ip_address,omitempty"`
	UserAgent      *string                `json:"user_agent,omitempty"`
	Details        map[string]interface{} `json:"details,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
}

type APIKey struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	Name       string     `json:"name"`
	KeyHash    string     `json:"-"`
	KeyPrefix  string     `json:"key_prefix"`
	Scopes     []string   `json:"scopes"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	IsActive   bool       `json:"is_active"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// Request/Response DTOs
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type CreateConnectionRequest struct {
	Name       string  `json:"name"`
	Host       string  `json:"host"`
	Port       int     `json:"port"`
	Username   string  `json:"username"`
	AuthType   string  `json:"auth_type"`
	Password   *string `json:"password,omitempty"`
	PrivateKey *string `json:"private_key,omitempty"`
}

type ConnectRequest struct {
	ConnectionID *uuid.UUID `json:"connection_id,omitempty"`
	Host         string     `json:"host"`
	Port         int        `json:"port"`
	Username     string     `json:"username"`
	AuthType     string     `json:"auth_type"`
	Password     *string    `json:"password,omitempty"`
	PrivateKey   *string    `json:"private_key,omitempty"`
}

type HostKeyVerificationRequest struct {
	SessionToken string `json:"session_token"`
	Accept       bool   `json:"accept"`
}

type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type TerminalResizeMessage struct {
	Rows int `json:"rows"`
	Cols int `json:"cols"`
}
