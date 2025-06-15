package quizzes

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"KdnSite/internal/auth"

	"github.com/google/uuid"
)

// ListQuizzes handles GET /api/quizzes
func ListQuizzes(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := auth.GetUserIDFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		rows, err := db.QueryContext(r.Context(), `SELECT id, owner_id, title, created_at, updated_at FROM quizzes WHERE owner_id=$1`, userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var quizzes []*Quiz
		for rows.Next() {
			var q Quiz
			if err := rows.Scan(&q.ID, &q.OwnerID, &q.Title, &q.CreatedAt, &q.UpdatedAt); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			quizzes = append(quizzes, &q)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(quizzes)
	}
}

// CreateQuiz handles POST /api/quizzes
func CreateQuiz(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := auth.GetUserIDFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var q Quiz
		if err := json.NewDecoder(r.Body).Decode(&q); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		q.ID = uuid.NewString()
		q.OwnerID = userID
		t := time.Now().Unix()
		q.CreatedAt = t
		q.UpdatedAt = t
		_, err = db.ExecContext(r.Context(), `INSERT INTO quizzes (id, owner_id, title, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`,
			q.ID, q.OwnerID, q.Title, q.CreatedAt, q.UpdatedAt)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(q)
	}
}
