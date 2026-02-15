package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/matras34/UpworkNotifier/internal/api/handlers"
	"github.com/matras34/UpworkNotifier/internal/api/middleware"
	"github.com/matras34/UpworkNotifier/internal/auth"
	"github.com/rs/cors"
)

type Router struct {
	*mux.Router
}

func NewRouter(
	authHandler *handlers.AuthHandler,
	sshHandler *handlers.SSHHandler,
	authMiddleware *middleware.AuthMiddleware,
	rateLimiter *middleware.RateLimiter,
	loggerMiddleware func(http.Handler) http.Handler,
	allowedOrigins []string,
) *Router {
	r := mux.NewRouter()

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	// Apply global middleware
	r.Use(loggerMiddleware)
	r.Use(c.Handler)

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// API v1
	api := r.PathPrefix("/api/v1").Subrouter()

	// Public routes (no auth required)
	public := api.PathPrefix("").Subrouter()
	public.Use(rateLimiter.Limit)
	public.HandleFunc("/auth/login", authHandler.Login).Methods("POST")
	public.HandleFunc("/auth/register", authHandler.Register).Methods("POST")

	// Protected routes (auth required)
	protected := api.PathPrefix("").Subrouter()
	protected.Use(authMiddleware.Authenticate)
	protected.Use(rateLimiter.Limit)

	// Auth routes
	protected.HandleFunc("/auth/me", authHandler.Me).Methods("GET")

	// SSH Connection routes
	protected.HandleFunc("/connections", sshHandler.ListConnections).Methods("GET")
	protected.Handle("/connections", 
		authMiddleware.RequirePermission(auth.PermissionWriteConnections)(
			http.HandlerFunc(sshHandler.SaveConnection),
		),
	).Methods("POST")
	protected.Handle("/connections", 
		authMiddleware.RequirePermission(auth.PermissionDeleteConnections)(
			http.HandlerFunc(sshHandler.DeleteConnection),
		),
	).Methods("DELETE")

	// WebSocket SSH connection
	protected.HandleFunc("/ssh/connect", sshHandler.Connect).Methods("GET")

	return &Router{r}
}
