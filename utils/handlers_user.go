package utils

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

type UserProfile struct {
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
}

type UserPreferences struct {
	Theme              string `json:"theme"`
	NotificationsEmail bool   `json:"notificationsEmail"`
	Language           string `json:"language"`
	FontSize           string `json:"fontSize"`
	HighContrast       bool   `json:"highContrast"`
}

type AuthCallbackRequest struct {
	Token string `json:"token"`
}

// GET /api/user/profile - returns the current user's profile info
func HandleUserProfile(w http.ResponseWriter, r *http.Request) {
	log.Infof("[HandleUserProfile] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	if r.Method != http.MethodGet {
		log.Warnf("[HandleUserProfile] Method not allowed: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID, err := GetUserIDFromRequest(r)
	if err != nil {
		log.Warnf("[HandleUserProfile] Unauthorized: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized: " + err.Error()))
		return
	}
	var profile UserProfile
	err = DB.QueryRow(`SELECT display_name, email FROM users WHERE id=$1`, userID).Scan(&profile.DisplayName, &profile.Email)
	if err == sql.ErrNoRows {
		log.Warnf("[HandleUserProfile] User not found: %s", userID)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("user not found"))
		return
	} else if err != nil {
		log.Errorf("[HandleUserProfile] DB error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("db error: " + err.Error()))
		return
	}
	log.Infof("[HandleUserProfile] Success for user %s", userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

// POST /api/user/profile - updates the current user's profile info
func HandleUserProfileUpdate(w http.ResponseWriter, r *http.Request) {
	log.Infof("[HandleUserProfileUpdate] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	if r.Method != http.MethodPost {
		log.Warnf("[HandleUserProfileUpdate] Method not allowed: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID, err := GetUserIDFromRequest(r)
	if err != nil {
		log.Warnf("[HandleUserProfileUpdate] Unauthorized: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized: " + err.Error()))
		return
	}
	var req UserProfile
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warnf("[HandleUserProfileUpdate] Invalid JSON: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid json"))
		return
	}
	if req.DisplayName == "" || req.Email == "" {
		log.Warnf("[HandleUserProfileUpdate] Missing displayName or email")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing displayName or email"))
		return
	}
	_, err = DB.Exec(`UPDATE users SET display_name=$1, email=$2 WHERE id=$3`, req.DisplayName, req.Email, userID)
	if err != nil {
		log.Errorf("[HandleUserProfileUpdate] DB error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("db error: " + err.Error()))
		return
	}
	log.Infof("[HandleUserProfileUpdate] Updated profile for user %s", userID)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// GET /api/user/preferences - returns the current user's preferences
func HandleUserPreferences(w http.ResponseWriter, r *http.Request) {
	log.Infof("[HandleUserPreferences] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	if r.Method != http.MethodGet {
		log.Warnf("[HandleUserPreferences] Method not allowed: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID, err := GetUserIDFromRequest(r)
	if err != nil {
		log.Warnf("[HandleUserPreferences] Unauthorized: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized: " + err.Error()))
		return
	}
	var prefs UserPreferences
	err = DB.QueryRow(`SELECT theme, notifications_email, language, font_size, high_contrast FROM users WHERE id=$1`, userID).Scan(&prefs.Theme, &prefs.NotificationsEmail, &prefs.Language, &prefs.FontSize, &prefs.HighContrast)
	if err != nil {
		log.Errorf("[HandleUserPreferences] DB error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("db error: " + err.Error()))
		return
	}
	log.Infof("[HandleUserPreferences] Success for user %s", userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prefs)
}

// POST /api/user/preferences - updates the current user's preferences
func HandleUserPreferencesUpdate(w http.ResponseWriter, r *http.Request) {
	log.Infof("[HandleUserPreferencesUpdate] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	if r.Method != http.MethodPost {
		log.Warnf("[HandleUserPreferencesUpdate] Method not allowed: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID, err := GetUserIDFromRequest(r)
	if err != nil {
		log.Warnf("[HandleUserPreferencesUpdate] Unauthorized: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized: " + err.Error()))
		return
	}
	var req UserPreferences
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warnf("[HandleUserPreferencesUpdate] Invalid JSON: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid json"))
		return
	}
	if req.Theme == "" {
		log.Warnf("[HandleUserPreferencesUpdate] Missing theme")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing theme"))
		return
	}
	_, err = DB.Exec(`UPDATE users SET theme=$1, notifications_email=$2, language=$3, font_size=$4, high_contrast=$5 WHERE id=$6`, req.Theme, req.NotificationsEmail, req.Language, req.FontSize, req.HighContrast, userID)
	if err != nil {
		log.Errorf("[HandleUserPreferencesUpdate] DB error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("db error: " + err.Error()))
		return
	}
	log.Infof("[HandleUserPreferencesUpdate] Updated preferences for user %s", userID)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
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

// POST /api/user/delete - delete the current user's account and all their projects
func HandleUserDelete(w http.ResponseWriter, r *http.Request) {
	log.Infof("[HandleUserDelete] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	if r.Method != http.MethodPost {
		log.Warnf("[HandleUserDelete] Method not allowed: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID, err := GetUserIDFromRequest(r)
	if err != nil {
		log.Warnf("[HandleUserDelete] Unauthorized: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized: " + err.Error()))
		return
	}
	// Delete all projects for this user
	_, err = DB.Exec(`DELETE FROM projects WHERE user_id=$1`, userID)
	if err != nil {
		log.Errorf("[HandleUserDelete] DB error (projects): %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("db error: " + err.Error()))
		return
	}
	// Delete the user
	_, err = DB.Exec(`DELETE FROM users WHERE id=$1`, userID)
	if err != nil {
		log.Errorf("[HandleUserDelete] DB error (user): %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("db error: " + err.Error()))
		return
	}
	log.Infof("[HandleUserDelete] Deleted user %s and all their projects", userID)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
