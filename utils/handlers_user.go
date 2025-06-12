package utils

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

type AuthCallbackRequest struct {
	Token string `json:"token"`
}

// POST /api/auth/callback - sets the JWT as a secure, HttpOnly cookie
func HandleAuthCallback(w http.ResponseWriter, r *http.Request) {
	log.Infof("[HandleAuthCallback] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	if r.Method != http.MethodPost {
		log.Warnf("[HandleAuthCallback] Method not allowed: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req AuthCallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Token == "" {
		log.Warnf("[HandleAuthCallback] Invalid token in request: err=%v, body=%v", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid token"))
		return
	}
	secure := true
	if os.Getenv("GO_ENV") != "production" {
		secure = false
	}
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    req.Token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(30 * 24 * time.Hour), // 30 days
	}
	http.SetCookie(w, cookie)
	log.Infof("[HandleAuthCallback] Set auth_token cookie for remote=%s, secure=%v, path=%s, expires=%v", r.RemoteAddr, secure, cookie.Path, cookie.Expires)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
