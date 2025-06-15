package leaderboard

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// ListLeaderboard handles GET /api/leaderboard
func ListLeaderboard(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.QueryContext(r.Context(), `SELECT user_id, username, score, streak, rank FROM leaderboard ORDER BY score DESC, streak DESC LIMIT 50`)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var entries []*LeaderboardEntry
		for rows.Next() {
			var e LeaderboardEntry
			if err := rows.Scan(&e.UserID, &e.Username, &e.Score, &e.Streak, &e.Rank); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			entries = append(entries, &e)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries)
	}
}
