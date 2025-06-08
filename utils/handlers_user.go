package utils

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type UserProfile struct {
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
}

// GET /api/user/profile - returns the current user's profile info
func HandleUserProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID, err := GetUserIDFromAuthHeader(r)
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
	userID, err := GetUserIDFromAuthHeader(r)
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
