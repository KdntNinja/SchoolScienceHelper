package utils

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

var DB *sql.DB

func SetDB(db *sql.DB) {
	DB = db
}

func HandleProjectSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID, err := GetUserIDFromAuthHeader(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized: " + err.Error()))
		return
	}
	var req struct {
		Name string          `json:"name"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid json"))
		return
	}
	if req.Name == "" || len(req.Data) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing name or data"))
		return
	}
	ctx := context.Background()
	_, err = DB.ExecContext(ctx, `INSERT INTO projects (user_id, name, data, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id, name) DO UPDATE SET data = EXCLUDED.data, updated_at = NOW()`, userID, req.Name, req.Data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("db error: " + err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func HandleProjectLoad(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID, err := GetUserIDFromAuthHeader(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized: " + err.Error()))
		return
	}
	name := r.URL.Query().Get("name")
	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing name param"))
		return
	}
	var data json.RawMessage
	var updated time.Time
	err = DB.QueryRow(`SELECT data, updated_at FROM projects WHERE user_id=$1 AND name=$2`, userID, name).Scan(&data, &updated)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("db error: " + err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data))
}

func HandleProjectList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID, err := GetUserIDFromAuthHeader(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized: " + err.Error()))
		return
	}
	rows, err := DB.Query(`SELECT name, updated_at, is_public, public_id FROM projects WHERE user_id=$1`, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("db error: " + err.Error()))
		return
	}
	defer rows.Close()
	var projects []map[string]interface{}
	for rows.Next() {
		var name, publicID sql.NullString
		var updated time.Time
		var isPublic bool
		if err := rows.Scan(&name, &updated, &isPublic, &publicID); err != nil {
			continue
		}
		projects = append(projects, map[string]interface{}{
			"name":       name.String,
			"updated_at": updated,
			"is_public":  isPublic,
			"public_id":  publicID.String,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func HandleProjectDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID, err := GetUserIDFromAuthHeader(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized: " + err.Error()))
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid json or missing name"))
		return
	}
	_, err = DB.Exec(`DELETE FROM projects WHERE user_id=$1 AND name=$2`, userID, req.Name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("db error: " + err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func HandleProjectPublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID, err := GetUserIDFromAuthHeader(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized: " + err.Error()))
		return
	}
	var req struct {
		Name   string `json:"name"`
		Public bool   `json:"public"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid json or missing name"))
		return
	}
	var publicID sql.NullString
	if req.Public {
		publicID.String = GeneratePublicID()
		publicID.Valid = true
	} else {
		publicID.Valid = false
	}
	_, err = DB.Exec(`UPDATE projects SET is_public=$1, public_id=CASE WHEN $1 THEN COALESCE(public_id, $2) ELSE NULL END WHERE user_id=$3 AND name=$4`, req.Public, publicID.String, userID, req.Name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("db error: " + err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func HandleProjectLoadPublic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	publicID := r.URL.Query().Get("public_id")
	if publicID == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing public_id param"))
		return
	}
	var data json.RawMessage
	var updated time.Time
	err := DB.QueryRow(`SELECT data, updated_at FROM projects WHERE public_id=$1 AND is_public=TRUE`, publicID).Scan(&data, &updated)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("db error: " + err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data))
}
