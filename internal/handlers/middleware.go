package handlers

import (
	"net/http"

	"SchoolScienceHelper/internal/auth"

	log "github.com/sirupsen/logrus"
)

// Auth middleware for all user related routes
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := auth.GetUserIDFromRequest(r)
		if err != nil || userID == "" {
			log.Warnf("[RequireAuth] Unauthorized access: %v, remote=%s, path=%s, cookie=%v", err, r.RemoteAddr, r.URL.Path, r.Header.Get("Cookie"))
			http.Redirect(w, r, "/auth", http.StatusFound)
			return
		}
		log.Infof("[RequireAuth] Authenticated user: %s, remote=%s, path=%s", userID, r.RemoteAddr, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
