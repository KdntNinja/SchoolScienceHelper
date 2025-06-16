package handlers

import (
	"encoding/json"
	"net/http"
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
	// Optionally include more user info if needed
	resp := map[string]interface{}{
		"authenticated": true,
		"user":          user,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
