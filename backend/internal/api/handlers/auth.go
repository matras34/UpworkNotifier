package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/matras34/UpworkNotifier/internal/auth"
	"github.com/matras34/UpworkNotifier/internal/cache"
	"github.com/matras34/UpworkNotifier/internal/db"
	"github.com/matras34/UpworkNotifier/internal/db/models"
	"github.com/matras34/UpworkNotifier/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db         *db.Database
	cache      *cache.Cache
	jwtManager *auth.JWTManager
	logger     *logger.Logger
	bcryptCost int
}

func NewAuthHandler(database *db.Database, cache *cache.Cache, jwtManager *auth.JWTManager, log *logger.Logger, bcryptCost int) *AuthHandler {
	return &AuthHandler{
		db:         database,
		cache:      cache,
		jwtManager: jwtManager,
		logger:     log,
		bcryptCost: bcryptCost,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password required", http.StatusBadRequest)
		return
	}

	// Check rate limiting for failed login attempts
	loginKey := "login_attempts:" + req.Email
	exists, _ := h.cache.Exists(r.Context(), loginKey)
	if exists {
		http.Error(w, "Too many login attempts. Please try again later.", http.StatusTooManyRequests)
		return
	}

	// Get user from database
	var user models.User
	err := h.db.QueryRow(
		`SELECT id, organization_id, email, password_hash, role, is_active 
		 FROM users WHERE email = $1`,
		req.Email,
	).Scan(&user.ID, &user.OrganizationID, &user.Email, &user.PasswordHash, &user.Role, &user.IsActive)

	if err == sql.ErrNoRows {
		h.trackFailedLogin(r.Context(), req.Email)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err != nil {
		h.logger.Error("Database error", map[string]interface{}{"error": err.Error()})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !user.IsActive {
		http.Error(w, "Account is disabled", http.StatusForbidden)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		h.trackFailedLogin(r.Context(), req.Email)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Update last login
	h.db.Exec("UPDATE users SET last_login_at = $1 WHERE id = $2", time.Now(), user.ID)

	// Generate JWT token
	token, err := h.jwtManager.GenerateToken(user.ID, user.OrganizationID, user.Email, user.Role)
	if err != nil {
		h.logger.Error("Failed to generate token", map[string]interface{}{"error": err.Error()})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	user.PasswordHash = "" // Don't send password hash

	response := models.LoginResponse{
		Token: token,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	h.logger.Info("User logged in", map[string]interface{}{
		"user_id": user.ID.String(),
		"email":   user.Email,
	})
}

func (h *AuthHandler) trackFailedLogin(ctx context.Context, email string) {
	loginKey := "login_attempts:" + email
	count, _ := h.cache.Increment(ctx, loginKey)
	if count >= 5 {
		h.cache.Expire(ctx, loginKey, 15*time.Minute)
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Simplified registration - in production, add email verification, etc.
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password required", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 8 {
		http.Error(w, "Password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), h.bcryptCost)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get default organization
	var orgID uuid.UUID
	err = h.db.QueryRow("SELECT id FROM organizations LIMIT 1").Scan(&orgID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create user
	userID := uuid.New()
	_, err = h.db.Exec(
		`INSERT INTO users (id, organization_id, email, password_hash, role, is_active)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		userID, orgID, req.Email, string(hashedPassword), "user", true,
	)

	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"users_email_key\"" {
			http.Error(w, "Email already exists", http.StatusConflict)
			return
		}
		h.logger.Error("Failed to create user", map[string]interface{}{"error": err.Error()})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User created successfully",
		"user_id": userID.String(),
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	// This endpoint returns the current user info based on JWT token
	// The user is already extracted by auth middleware
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Use auth middleware to get user from context",
	})
}
