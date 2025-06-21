package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
)

// AuthCheckHandler returns the user's authentication status and basic info (if logged in)
func AuthCheckHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value("user").(map[string]interface{})
	if !ok || user == nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"authenticated": false,
		})
		return
	}
	// Always use Auth0 nickname (or name) for username
	if user["nickname"] != nil {
		user["username"] = user["nickname"]
	} else if user["name"] != nil {
		user["username"] = user["name"]
	}
	// Fetch avatar_url from DB and inject as user.picture if present
	userID, _ := user["sub"].(string)
	if userID != "" {
		dbURL := os.Getenv("POSTGRES_DATABASE_URL")
		db, err := sql.Open("postgres", dbURL)
		if err == nil {
			defer db.Close()
			var avatarURL string
			db.QueryRow(`SELECT avatar_url FROM users WHERE id = $1`, userID).Scan(&avatarURL)
			if avatarURL != "" {
				user["picture"] = avatarURL
			}
		}
	}
	resp := map[string]interface{}{
		"authenticated": true,
		"user":          user,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
