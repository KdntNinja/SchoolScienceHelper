package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"KdnSite/internal/auth"
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

// POST /api/auth/logout - clears the auth_token cookie
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("logged out"))
}

// POST /api/auth/delete - deletes the user from Auth0
func DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID, err := getUserIDFromJWT(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	domain := os.Getenv("AUTH0_DOMAIN")
	apiToken := os.Getenv("AUTH0_MANAGEMENT_TOKEN")
	if domain == "" || apiToken == "" {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Auth0 config missing"))
		return
	}
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("DELETE", "https://"+domain+"/api/v2/users/"+userID, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode >= 400 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to delete user in Auth0"))
		return
	}
	LogoutHandler(w, r)
}

// POST /api/auth/change-password - triggers Auth0 password change email
func ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing email"))
		return
	}
	domain := os.Getenv("AUTH0_DOMAIN")
	clientID := os.Getenv("AUTH0_CLIENT_ID")
	if domain == "" || clientID == "" {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Auth0 config missing"))
		return
	}
	payload := map[string]string{
		"client_id": clientID,
		"email":     req.Email,
		"connection": "Username-Password-Authentication",
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post("https://"+domain+"/dbconnections/change_password", "application/json", bytes.NewReader(body))
	if err != nil || resp.StatusCode >= 400 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to trigger password change"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Password change email sent"))
}

// Helper to extract user ID from JWT (sub claim)
func getUserIDFromJWT(r *http.Request) (string, error) {
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		return "", err
	}
	claims, err := auth.ValidateAndParseJWT(cookie.Value)
	if err != nil {
		return "", err
	}
	if sub, ok := claims["sub"].(string); ok {
		return sub, nil
	}
	return "", http.ErrNoCookie
}
