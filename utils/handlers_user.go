package utils

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"
)

type UserProfile struct {
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
}

type UserPreferences struct {
	Theme string `json:"theme"`
}

type AuthCallbackRequest struct {
	Token string `json:"token"`
}

// GET /api/user/profile - returns the current user's profile info
func HandleUserProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID, err := GetUserIDFromRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized: " + err.Error()))
		return
	}
	var profile UserProfile
	err = DB.QueryRow(`SELECT display_name, email FROM users WHERE id=$1`, userID).Scan(&profile.DisplayName, &profile.Email)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("user not found"))
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("db error: " + err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

// POST /api/user/profile - updates the current user's profile info
func HandleUserProfileUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID, err := GetUserIDFromRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized: " + err.Error()))
		return
	}
	var req UserProfile
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid json"))
		return
	}
	if req.DisplayName == "" || req.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing displayName or email"))
		return
	}
	_, err = DB.Exec(`UPDATE users SET display_name=$1, email=$2 WHERE id=$3`, req.DisplayName, req.Email, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("db error: " + err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// GET /api/user/preferences - returns the current user's preferences
func HandleUserPreferences(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID, err := GetUserIDFromRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized: " + err.Error()))
		return
	}
	var prefs UserPreferences
	err = DB.QueryRow(`SELECT theme FROM users WHERE id=$1`, userID).Scan(&prefs.Theme)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("db error: " + err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prefs)
}

// POST /api/user/preferences - updates the current user's preferences
func HandleUserPreferencesUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID, err := GetUserIDFromRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized: " + err.Error()))
		return
	}
	var req UserPreferences
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid json"))
		return
	}
	if req.Theme == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing theme"))
		return
	}
	_, err = DB.Exec(`UPDATE users SET theme=$1 WHERE id=$2`, req.Theme, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("db error: " + err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// POST /api/auth/callback - sets the JWT as a secure, HttpOnly cookie
func HandleAuthCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req AuthCallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Token == "" {
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
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
