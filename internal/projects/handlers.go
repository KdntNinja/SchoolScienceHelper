package projects

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"KdnSite/internal/auth"

	"github.com/google/uuid"
)

// ListProjects handles GET /api/projects
func ListProjects(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := auth.GetUserIDFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		rows, err := db.QueryContext(r.Context(), `SELECT id, owner_id, title, created_at, updated_at, data FROM projects WHERE owner_id=$1`, userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var projects []*Project
		for rows.Next() {
			var p Project
			if err := rows.Scan(&p.ID, &p.OwnerID, &p.Title, &p.CreatedAt, &p.UpdatedAt, &p.Data); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			projects = append(projects, &p)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(projects)
	}
}

// CreateProject handles POST /api/projects
func CreateProject(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := auth.GetUserIDFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var p Project
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		p.ID = uuid.NewString()
		p.OwnerID = userID
		t := time.Now().Unix()
		p.CreatedAt = t
		p.UpdatedAt = t
		_, err = db.ExecContext(r.Context(), `INSERT INTO projects (id, owner_id, title, created_at, updated_at, data) VALUES ($1, $2, $3, $4, $5, $6)`,
			p.ID, p.OwnerID, p.Title, p.CreatedAt, p.UpdatedAt, p.Data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(p)
	}
}
