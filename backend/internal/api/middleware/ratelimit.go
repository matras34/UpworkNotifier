package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/matras34/UpworkNotifier/internal/cache"
)

type RateLimiter struct {
	cache          *cache.Cache
	requestsPerMin int
	window         time.Duration
}

func NewRateLimiter(cache *cache.Cache, requestsPerMin int) *RateLimiter {
	return &RateLimiter{
		cache:          cache,
		requestsPerMin: requestsPerMin,
		window:         time.Minute,
	}
}

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user from context
		claims := GetUserFromContext(r.Context())
		if claims == nil {
			// For unauthenticated requests, use IP
			key := fmt.Sprintf("ratelimit:ip:%s", r.RemoteAddr)
			allowed, err := rl.cache.CheckRateLimit(r.Context(), key, rl.requestsPerMin, rl.window)
			if err != nil || !allowed {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
		} else {
			// For authenticated requests, use user ID
			key := fmt.Sprintf("ratelimit:user:%s", claims.UserID.String())
			allowed, err := rl.cache.CheckRateLimit(r.Context(), key, rl.requestsPerMin, rl.window)
			if err != nil || !allowed {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
