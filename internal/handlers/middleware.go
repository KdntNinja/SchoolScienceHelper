package handlers

import (
	"context"
	"net/http"

	"KdnSite/internal/auth"

	log "github.com/sirupsen/logrus"
)

// Auth middleware for all user related routes
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := auth.GetJWTFromRequest(r)
		if tokenStr == "" {
			log.Warnf("[RequireAuth] No auth_token found, remote=%s, path=%s, cookie=%v", r.RemoteAddr, r.URL.Path, r.Header.Get("Cookie"))
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		claims, err := auth.ValidateAndParseJWT(tokenStr)
		if err != nil {
			log.Warnf("[RequireAuth] Invalid JWT: %v, remote=%s, path=%s", err, r.RemoteAddr, r.URL.Path)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		emailVerified, ok := claims["email_verified"].(bool)
		if !ok || !emailVerified {
			log.Warnf("[RequireAuth] Email not verified for user: %v, remote=%s, path=%s", claims["sub"], r.RemoteAddr, r.URL.Path)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		// Set user info in context for downstream handlers
		ctx := context.WithValue(r.Context(), "user", claims)
		log.Infof("[RequireAuth] Authenticated user: %v, remote=%s, path=%s", claims["sub"], r.RemoteAddr, r.URL.Path)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole middleware for admin endpoints
func RequireRole(role string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _, roles, err := getUserClaimsFromJWT(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		for _, userRole := range roles {
			if userRole == role {
				next.ServeHTTP(w, r)
				return
			}
		}
		w.WriteHeader(http.StatusForbidden)
	})
}

// AuthStatusHandler returns 200 if authenticated, 401 otherwise
func AuthStatusHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromRequest(r)
	if err != nil || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"authenticated": false}`))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"authenticated": true, "userID": "` + userID + `"}`))
}
