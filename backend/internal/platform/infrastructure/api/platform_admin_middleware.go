package api

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
)

const PlatformAdminKeyHeader = "X-Platform-Admin-Key"

func PlatformAdminMiddleware(configuredAPIKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if configuredAPIKey == "" {
				respondUnauthorized(w, "Platform admin API not configured")
				return
			}

			providedKey := r.Header.Get(PlatformAdminKeyHeader)
			if providedKey == "" {
				respondUnauthorized(w, "Missing API key")
				return
			}

			if !secureCompare(providedKey, configuredAPIKey) {
				respondUnauthorized(w, "Invalid API key")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func secureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func respondUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":   "Unauthorized",
		"message": message,
	})
}
