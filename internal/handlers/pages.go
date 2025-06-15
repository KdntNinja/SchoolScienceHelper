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
	if name, ok := claims["name"].(string); ok && name != "" {
		displayName = name
	} else if email, ok := claims["email"].(string); ok && email != "" {
		at := strings.Index(email, "@")
		if at > 0 {
			local := email[:at]
			if len(local) > 0 {
				displayName = strings.Title(local)
			}
		}
	}
	_ = userpages.Dash(displayName).Render(r.Context(), w)
}

// ...move or wrap page rendering handlers here if needed...
