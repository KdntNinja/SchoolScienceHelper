package resources

import (
	"KdnSite/internal/auth"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// ListResources handles GET /api/resources
func ListResources(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := auth.GetUserIDFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		rows, err := db.QueryContext(r.Context(), `SELECT id, owner_id, type, title, content, created_at, updated_at FROM resources WHERE owner_id=$1`, userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var resources []*Resource
		for rows.Next() {
			var res Resource
			if err := rows.Scan(&res.ID, &res.OwnerID, &res.Type, &res.Title, &res.Content, &res.CreatedAt, &res.UpdatedAt); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			resources = append(resources, &res)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resources)
	}
}

// CreateResource handles POST /api/resources
func CreateResource(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := auth.GetUserIDFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var res Resource
		if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		res.ID = uuid.NewString()
		res.OwnerID = userID
		t := time.Now().Unix()
		res.CreatedAt = t
		res.UpdatedAt = t
		_, err = db.ExecContext(r.Context(), `INSERT INTO resources (id, owner_id, type, title, content, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			res.ID, res.OwnerID, res.Type, res.Title, res.Content, res.CreatedAt, res.UpdatedAt)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(res)
	}
}
