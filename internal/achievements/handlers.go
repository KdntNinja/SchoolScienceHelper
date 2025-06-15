package achievements

import (
	"KdnSite/internal/auth"
	"database/sql"
	"encoding/json"
	"net/http"
)

// ListAchievements handles GET /api/achievements
func ListAchievements(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := auth.GetUserIDFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		rows, err := db.QueryContext(r.Context(), `SELECT id, user_id, name, desc, earned_at FROM achievements WHERE user_id=$1`, userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var achievements []*Achievement
		for rows.Next() {
			var a Achievement
			if err := rows.Scan(&a.ID, &a.UserID, &a.Name, &a.Desc, &a.EarnedAt); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			achievements = append(achievements, &a)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(achievements)
	}
}
