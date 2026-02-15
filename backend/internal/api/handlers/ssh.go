package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/matras34/UpworkNotifier/internal/api/middleware"
	"github.com/matras34/UpworkNotifier/internal/auth"
	"github.com/matras34/UpworkNotifier/internal/cache"
	"github.com/matras34/UpworkNotifier/internal/crypto"
	"github.com/matras34/UpworkNotifier/internal/db"
	"github.com/matras34/UpworkNotifier/internal/db/models"
	sshpkg "github.com/matras34/UpworkNotifier/internal/ssh"
	wspkg "github.com/matras34/UpworkNotifier/internal/websocket"
	"github.com/matras34/UpworkNotifier/pkg/logger"
	"golang.org/x/crypto/ssh"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
	CheckOrigin: func(r *http.Request) bool {
		// In production, validate origin properly
		return true
	},
}

type SSHHandler struct {
	db             *db.Database
	cache          *cache.Cache
	kms            *crypto.KMS
	logger         *logger.Logger
	sessionManager *sshpkg.SessionManager
	wsHub          *wspkg.Hub
}

func NewSSHHandler(
	database *db.Database,
	cache *cache.Cache,
	kms *crypto.KMS,
	logger *logger.Logger,
	sessionManager *sshpkg.SessionManager,
	wsHub *wspkg.Hub,
) *SSHHandler {
	return &SSHHandler{
		db:             database,
		cache:          cache,
		kms:            kms,
		logger:         logger,
		sessionManager: sessionManager,
		wsHub:          wsHub,
	}
}

func (h *SSHHandler) Connect(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade connection", map[string]interface{}{"error": err.Error()})
		return
	}

	// Create WebSocket client
	client := wspkg.NewClient(claims.UserID, conn, h.wsHub)
	h.wsHub.RegisterClient(client)

	// Start WebSocket pumps
	go client.WritePump()
	go client.ReadPump()

	// Wait for connection details from client
	go h.handleSSHConnection(client, claims)
}

func (h *SSHHandler) handleSSHConnection(client *wspkg.Client, claims *auth.Claims) {
	// This is a simplified version - in production, you'd wait for the client to send connection details
	// For now, we'll expect the connection details to be sent via the first WebSocket message
	
	// The actual SSH connection establishment happens after receiving connection details
	// from the WebSocket client
}

func (h *SSHHandler) EstablishSSH(client *wspkg.Client, userID uuid.UUID, connReq models.ConnectRequest, clientIP string) error {
	var host string
	var port int
	var username string
	var password *string
	var privateKey *string

	// If connection ID is provided, load from database
	if connReq.ConnectionID != nil {
		var conn models.SSHConnection
		err := h.db.QueryRow(
			`SELECT id, host, port, username, auth_type, encrypted_password, encrypted_private_key, encryption_salt, encryption_nonce
			 FROM ssh_connections WHERE id = $1 AND user_id = $2`,
			connReq.ConnectionID, userID,
		).Scan(&conn.ID, &conn.Host, &conn.Port, &conn.Username, &conn.AuthType, 
		       &conn.EncryptedPassword, &conn.EncryptedPrivateKey, &conn.EncryptionSalt, &conn.EncryptionNonce)

		if err != nil {
			return fmt.Errorf("connection not found: %w", err)
		}

		host = conn.Host
		port = conn.Port
		username = conn.Username

		// Decrypt credentials
		if conn.AuthType == "password" && conn.EncryptedPassword != nil {
			decrypted, err := h.kms.Decrypt(*conn.EncryptedPassword, *conn.EncryptionSalt, *conn.EncryptionNonce)
			if err != nil {
				return fmt.Errorf("failed to decrypt password: %w", err)
			}
			password = &decrypted
		} else if conn.AuthType == "privatekey" && conn.EncryptedPrivateKey != nil {
			decrypted, err := h.kms.Decrypt(*conn.EncryptedPrivateKey, *conn.EncryptionSalt, *conn.EncryptionNonce)
			if err != nil {
				return fmt.Errorf("failed to decrypt private key: %w", err)
			}
			privateKey = &decrypted
		}
	} else {
		// Use provided connection details
		host = connReq.Host
		port = connReq.Port
		username = connReq.Username
		password = connReq.Password
		privateKey = connReq.PrivateKey
	}

	// Check for known host keys
	var knownKeys []models.HostKey
	rows, err := h.db.Query(
		`SELECT algorithm, public_key FROM host_keys WHERE host = $1 AND port = $2`,
		host, port,
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var hk models.HostKey
			rows.Scan(&hk.Algorithm, &hk.PublicKey)
			knownKeys = append(knownKeys, hk)
		}
	}

	// Setup host key verification callback
	hostKeyChecker := sshpkg.NewHostKeyChecker(func(hostKeyInfo *sshpkg.HostKeyInfo) (bool, error) {
		// Send host key to client for verification
		msg := wspkg.Message{
			Type: wspkg.MessageTypeHostKey,
			Data: hostKeyInfo,
		}
		client.SendMessage(msg)

		// Wait for user confirmation (simplified - in production use a channel)
		time.Sleep(30 * time.Second)
		return false, fmt.Errorf("host key verification timeout")
	})

	// Add known keys
	for _, hk := range knownKeys {
		// Parse and add known key (simplified)
		_ = hk
	}

	// Create SSH client
	var sshPassword string
	if password != nil {
		sshPassword = *password
	}

	sshClient, err := sshpkg.NewSSHClient(
		host,
		port,
		username,
		sshPassword,
		privateKey,
		10*time.Second,
		hostKeyChecker.GetCallback(),
	)
	if err != nil {
		return fmt.Errorf("SSH connection failed: %w", err)
	}

	// Create session
	session, err := h.sessionManager.CreateSession(
		context.Background(),
		userID,
		sshClient,
		func(sessionID uuid.UUID, reason string) {
			h.handleSessionDisconnect(sessionID, reason)
		},
	)
	if err != nil {
		sshClient.Close()
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Link session to WebSocket client
	client.SetSSHSession(session)

	// Log session to database
	sessionToken := uuid.New().String()
	_, err = h.db.Exec(
		`INSERT INTO ssh_sessions (id, user_id, connection_id, session_token, client_ip, host, port, username, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		session.ID, userID, connReq.ConnectionID, sessionToken, clientIP, host, port, username, "active",
	)
	if err != nil {
		h.logger.Error("Failed to log session", map[string]interface{}{"error": err.Error()})
	}

	// Start SSH read pump
	go client.SSHReadPump()

	// Send connected message
	client.SendConnected()

	h.logger.Info("SSH session established", map[string]interface{}{
		"session_id": session.ID.String(),
		"user_id":    userID.String(),
		"host":       host,
	})

	return nil
}

func (h *SSHHandler) handleSessionDisconnect(sessionID uuid.UUID, reason string) {
	// Update session in database
	_, err := h.db.Exec(
		`UPDATE ssh_sessions SET status = $1, ended_at = $2, disconnect_reason = $3 WHERE id = $4`,
		"closed", time.Now(), reason, sessionID,
	)
	if err != nil {
		h.logger.Error("Failed to update session", map[string]interface{}{"error": err.Error()})
	}

	h.logger.Info("SSH session closed", map[string]interface{}{
		"session_id": sessionID.String(),
		"reason":     reason,
	})
}

func (h *SSHHandler) SaveConnection(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.CreateConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Name == "" || req.Host == "" || req.Username == "" {
		http.Error(w, "Name, host, and username are required", http.StatusBadRequest)
		return
	}

	if req.Port <= 0 || req.Port > 65535 {
		http.Error(w, "Invalid port", http.StatusBadRequest)
		return
	}

	if req.AuthType != "password" && req.AuthType != "privatekey" {
		http.Error(w, "Auth type must be 'password' or 'privatekey'", http.StatusBadRequest)
		return
	}

	// Encrypt credentials
	var encryptedPassword, encryptedKey, salt, nonce *string

	if req.AuthType == "password" && req.Password != nil {
		enc, s, n, err := h.kms.Encrypt(*req.Password)
		if err != nil {
			http.Error(w, "Failed to encrypt password", http.StatusInternalServerError)
			return
		}
		encryptedPassword = &enc
		salt = &s
		nonce = &n
	} else if req.AuthType == "privatekey" && req.PrivateKey != nil {
		// Validate private key format
		_, err := ssh.ParsePrivateKey([]byte(*req.PrivateKey))
		if err != nil {
			http.Error(w, "Invalid private key format", http.StatusBadRequest)
			return
		}

		enc, s, n, err := h.kms.Encrypt(*req.PrivateKey)
		if err != nil {
			http.Error(w, "Failed to encrypt private key", http.StatusInternalServerError)
			return
		}
		encryptedKey = &enc
		salt = &s
		nonce = &n
	}

	// Save to database
	connID := uuid.New()
	_, err := h.db.Exec(
		`INSERT INTO ssh_connections (id, user_id, name, host, port, username, auth_type, encrypted_password, encrypted_private_key, encryption_salt, encryption_nonce)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		connID, claims.UserID, req.Name, req.Host, req.Port, req.Username, req.AuthType,
		encryptedPassword, encryptedKey, salt, nonce,
	)

	if err != nil {
		h.logger.Error("Failed to save connection", map[string]interface{}{"error": err.Error()})
		http.Error(w, "Failed to save connection", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"id":      connID.String(),
		"message": "Connection saved successfully",
	})
}

func (h *SSHHandler) ListConnections(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := h.db.Query(
		`SELECT id, name, host, port, username, auth_type, created_at, updated_at
		 FROM ssh_connections WHERE user_id = $1 ORDER BY name`,
		claims.UserID,
	)
	if err != nil {
		http.Error(w, "Failed to fetch connections", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var connections []models.SSHConnection
	for rows.Next() {
		var conn models.SSHConnection
		err := rows.Scan(&conn.ID, &conn.Name, &conn.Host, &conn.Port, &conn.Username, &conn.AuthType, &conn.CreatedAt, &conn.UpdatedAt)
		if err != nil {
			continue
		}
		conn.UserID = claims.UserID
		connections = append(connections, conn)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(connections)
}

func (h *SSHHandler) DeleteConnection(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	connID := r.URL.Query().Get("id")
	if connID == "" {
		http.Error(w, "Connection ID required", http.StatusBadRequest)
		return
	}

	_, err := h.db.Exec(
		`DELETE FROM ssh_connections WHERE id = $1 AND user_id = $2`,
		connID, claims.UserID,
	)
	if err != nil {
		http.Error(w, "Failed to delete connection", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
