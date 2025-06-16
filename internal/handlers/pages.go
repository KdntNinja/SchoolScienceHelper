package handlers

import (
	"KdnSite/internal/auth"
	"KdnSite/internal/user"
	"database/sql"
	userpages "KdnSite/ui/pages/user"
	"net/http"
	"os"
	"strings"
)

// DashPageHandler renders the dashboard with the user's display name
func DashPageHandler(w http.ResponseWriter, r *http.Request) {
	dbURL := os.Getenv("POSTGRES_DATABASE_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	tokenStr := auth.GetJWTFromRequest(r)
	claims, err := auth.ValidateAndParseJWT(tokenStr)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	userID, _ := claims["sub"].(string)
	displayName := "Student"
	if userID != "" {
		profile, err := user.GetUserProfile(r.Context(), db, userID)
		if err == nil && profile != nil && profile.Username != "" {
			displayName = profile.Username
		} else if name, ok := claims["name"].(string); ok && name != "" && !strings.Contains(name, "@") {
			displayName = name
		} else if email, ok := claims["email"].(string); ok && email != "" {
			at := strings.Index(email, "@")
			if at > 0 {
				displayName = email[:at]
			}
		}
	}
	_ = userpages.Dash(displayName).Render(r.Context(), w)
}
