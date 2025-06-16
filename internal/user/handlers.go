package user

import (
	"KdnSite/internal/auth"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// GetProfile handles GET /api/user/profile
func GetProfile(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := auth.GetUserIDFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		profile, err := GetUserProfile(r.Context(), db, userID)
		if err != nil {
			// Auto-provision user if not found
			claims, _ := auth.ValidateAndParseJWT(auth.GetJWTFromRequest(r))
			email, _ := claims["email"].(string)
			name, _ := claims["name"].(string)
			username := name
			if username == "" && email != "" {
				at := strings.Index(email, "@")
				if at > 0 {
					username = email[:at]
				}
			}
			if username == "" {
				username = "Student"
			}
			_, err := db.ExecContext(r.Context(), `INSERT INTO users (id, email, username, created_at) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`, userID, email, username, time.Now().Unix())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			profile, err = GetUserProfile(r.Context(), db, userID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(profile)
	}
}

// UpdateProfile handles POST /api/user/profile
func UpdateProfile(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := auth.GetUserIDFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var u UserProfile
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		u.ID = userID
		if err := UpdateUserProfile(r.Context(), db, &u); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
