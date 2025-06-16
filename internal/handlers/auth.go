package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
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
	// Remove all user data from DB
	dbURL := os.Getenv("POSTGRES_DATABASE_URL")
	db, err := sql.Open("postgres", dbURL)
	if err == nil {
		defer db.Close()
		db.Exec(`DELETE FROM anki_cards WHERE owner_id = $1`, userID)
		db.Exec(`DELETE FROM anki_decks WHERE owner_id = $1`, userID)
		db.Exec(`DELETE FROM revision_resources WHERE owner_id = $1`, userID)
		db.Exec(`DELETE FROM resources WHERE owner_id = $1`, userID)
		db.Exec(`DELETE FROM leaderboard WHERE user_id = $1`, userID)
		db.Exec(`DELETE FROM projects WHERE owner_id = $1`, userID)
		db.Exec(`DELETE FROM users WHERE id = $1`, userID)
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
		"client_id":  clientID,
		"email":      req.Email,
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

// POST /api/auth/resend-verification - triggers Auth0 verification email
func ResendVerificationHandler(w http.ResponseWriter, r *http.Request) {
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
	// Get user email from Auth0
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", "https://"+domain+"/api/v2/users/"+userID, nil)
	req.Header.Set("Authorization", "Bearer "+apiToken)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode >= 400 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to get user info"))
		return
	}
	var user struct {
		Email string `json:"email"`
	}
	json.NewDecoder(resp.Body).Decode(&user)
	resp.Body.Close()
	if user.Email == "" {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("No email found"))
		return
	}
	// Trigger verification email
	payload := map[string]string{"user_id": userID, "client_id": os.Getenv("AUTH0_CLIENT_ID")}
	body, _ := json.Marshal(payload)
	req, _ = http.NewRequest("POST", "https://"+domain+"/api/v2/jobs/verification-email", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode >= 400 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to send verification email"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Verification email sent"))
}

// POST /api/auth/logout-all - logs out user from all sessions (Auth0 global logout)
func LogoutAllHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Just clear the cookie and redirect to Auth0 logout endpoint
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
	domain := os.Getenv("AUTH0_DOMAIN")
	clientID := os.Getenv("AUTH0_CLIENT_ID")
	if domain == "" || clientID == "" {
		w.WriteHeader(http.StatusOK)
		return
	}
	logoutURL := "https://" + domain + "/v2/logout?client_id=" + clientID + "&returnTo=" + os.Getenv("APP_URL")
	w.Header().Set("HX-Redirect", logoutURL)
	w.WriteHeader(http.StatusOK)
}

// GET /api/auth/sessions - (stub) returns active sessions (requires Auth0 Enterprise for full support)
func SessionsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Session listing not supported on free Auth0 tier."))
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

// Helper: Check email_verified and roles from JWT claims
func getUserClaimsFromJWT(r *http.Request) (userID string, emailVerified bool, roles []string, err error) {
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		return "", false, nil, err
	}
	claims, err := auth.ValidateAndParseJWT(cookie.Value)
	if err != nil {
		return "", false, nil, err
	}
	userID, _ = claims["sub"].(string)
	emailVerified, _ = claims["email_verified"].(bool)
	roles = nil
	if appMeta, ok := claims["https://app.kdnsite.site/app_metadata"].(map[string]interface{}); ok {
		if rs, ok := appMeta["roles"].([]interface{}); ok {
			for _, r := range rs {
				if s, ok := r.(string); ok {
					roles = append(roles, s)
				}
			}
		}
	}
	return
}

// Helper: Assign default role to user (call from callback or user provisioning)
func AssignDefaultRole(userID string) error {
	domain := os.Getenv("AUTH0_DOMAIN")
	apiToken := os.Getenv("AUTH0_MANAGEMENT_TOKEN")
	if domain == "" || apiToken == "" {
		return errors.New("Auth0 config missing")
	}
	roleID := os.Getenv("AUTH0_DEFAULT_ROLE_ID") // set this in env
	if roleID == "" {
		return errors.New("Default role ID not set")
	}
	client := &http.Client{Timeout: 10 * time.Second}
	payload := map[string][]string{"roles": {roleID}}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://"+domain+"/api/v2/users/"+userID+"/roles", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode >= 400 {
		return errors.New("Failed to assign role")
	}
	return nil
}
