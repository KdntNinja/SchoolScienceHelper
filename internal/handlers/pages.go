package handlers

import (
	"KdnSite/internal/auth"
	userpages "KdnSite/ui/pages/user"
	"net/http"
	"strings"
)

// DashPageHandler renders the dashboard with the user's display name
func DashPageHandler(w http.ResponseWriter, r *http.Request) {
	tokenStr := auth.GetJWTFromRequest(r)
	claims, err := auth.ValidateAndParseJWT(tokenStr)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	displayName := "Student"
	if name, ok := claims["name"].(string); ok && name != "" && !strings.Contains(name, "@") {
		displayName = name
	} else if email, ok := claims["email"].(string); ok && email != "" {
		at := strings.Index(email, "@")
		if at > 0 {
			displayName = email[:at]
		}
	}
	_ = userpages.Dash(displayName).Render(r.Context(), w)
}
