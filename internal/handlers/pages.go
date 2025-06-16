package handlers

import (
	"KdnSite/internal/auth"
	"KdnSite/internal/user"
	userpages "KdnSite/ui/pages/user"
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"
)

// getDisplayName returns the best display name for the user
func getDisplayName(ctx context.Context, db *sql.DB, claims map[string]interface{}) string {
	userID, _ := claims["sub"].(string)
	if userID != "" {
		if profile, err := user.GetUserProfile(ctx, db, userID); err == nil && profile != nil && profile.Username != "" {
			return profile.Username
		}
	}
	if name, ok := claims["name"].(string); ok && name != "" && !strings.Contains(name, "@") {
		return name
	}
	if email, ok := claims["email"].(string); ok && email != "" {
		if at := strings.Index(email, "@"); at > 0 {
			return email[:at]
		}
	}
	return "Student"
}

// DashPageHandler renders the dashboard with the user's display name
func DashPageHandler(w http.ResponseWriter, r *http.Request) {
	dbURL := os.Getenv("POSTGRES_DATABASE_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("[DashPageHandler] DB open error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	tokenStr := auth.GetJWTFromRequest(r)
	claims, err := auth.ValidateAndParseJWT(tokenStr)
	if err != nil {
		log.Printf("[DashPageHandler] JWT error: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	displayName := getDisplayName(r.Context(), db, claims)
	if err := userpages.Dash(displayName).Render(r.Context(), w); err != nil {
		log.Printf("[DashPageHandler] Render error: %v", err)
	}
}
