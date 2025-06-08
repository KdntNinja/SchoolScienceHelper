package utils

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

var DB *sql.DB

func SetDB(db *sql.DB) {
	DB = db
}

func HandleProjectSave(w http.ResponseWriter, r *http.Request) {
	log.Infof("[HandleProjectSave] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	if r.Method != http.MethodPost {
		log.Warnf("[HandleProjectSave] Method not allowed: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID, err := GetUserIDFromRequest(r)
	if err != nil {
		log.Warnf("[HandleProjectSave] Unauthorized: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized: " + err.Error()))
		return
	}
	var req struct {
		Name string          `json:"name"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warnf("[HandleProjectSave] Invalid JSON: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid json"))
		return
	}
	if req.Name == "" || len(req.Data) == 0 {
		log.Warnf("[HandleProjectSave] Missing name or data")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing name or data"))
		return
	}
	ctx := context.Background()
	_, err = DB.ExecContext(ctx, `INSERT INTO projects (user_id, name, data, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id, name) DO UPDATE SET data = EXCLUDED.data, updated_at = NOW()`, userID, req.Name, req.Data)
	if err != nil {
		log.Errorf("[HandleProjectSave] DB error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("db error: " + err.Error()))
		return
	}
	log.Infof("[HandleProjectSave] Saved project '%s' for user %s", req.Name, userID)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func HandleProjectLoad(w http.ResponseWriter, r *http.Request) {
	log.Infof("[HandleProjectLoad] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	if r.Method != http.MethodGet {
		log.Warnf("[HandleProjectLoad] Method not allowed: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID, err := GetUserIDFromRequest(r)
	if err != nil {
		log.Warnf("[HandleProjectLoad] Unauthorized: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized: " + err.Error()))
		return
	}
	name := r.URL.Query().Get("name")
	if name == "" {
		log.Warnf("[HandleProjectLoad] Missing name param")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing name param"))
		return
	}
	var data json.RawMessage
	var updated time.Time
	err = DB.QueryRow(`SELECT data, updated_at FROM projects WHERE user_id=$1 AND name=$2`, userID, name).Scan(&data, &updated)
	if err == sql.ErrNoRows {
		log.Warnf("[HandleProjectLoad] Not found: %s for user %s", name, userID)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
		return
	} else if err != nil {
		log.Errorf("[HandleProjectLoad] DB error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("db error: " + err.Error()))
		return
	}
	log.Infof("[HandleProjectLoad] Loaded project '%s' for user %s", name, userID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data))
}

func HandleProjectList(w http.ResponseWriter, r *http.Request) {
	log.Infof("[HandleProjectList] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	if r.Method != http.MethodGet {
		log.Warnf("[HandleProjectList] Method not allowed: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID, err := GetUserIDFromRequest(r)
	if err != nil {
		log.Warnf("[HandleProjectList] Unauthorized: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized: " + err.Error()))
		return
	}
	rows, err := DB.Query(`SELECT name, updated_at, is_public, public_id FROM projects WHERE user_id=$1`, userID)
	if err != nil {
		log.Errorf("[HandleProjectList] DB error: %v", err)
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
			log.Warnf("[HandleProjectList] Row scan error: %v", err)
			continue
		}
		projects = append(projects, map[string]interface{}{
			"name":       name.String,
			"updated_at": updated,
			"is_public":  isPublic,
			"public_id":  publicID.String,
		})
	}
	log.Infof("[HandleProjectList] Returned %d projects for user %s", len(projects), userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func HandleProjectDelete(w http.ResponseWriter, r *http.Request) {
	log.Infof("[HandleProjectDelete] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	if r.Method != http.MethodPost {
		log.Warnf("[HandleProjectDelete] Method not allowed: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID, err := GetUserIDFromRequest(r)
	if err != nil {
		log.Warnf("[HandleProjectDelete] Unauthorized: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized: " + err.Error()))
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		log.Warnf("[HandleProjectDelete] Invalid JSON or missing name: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid json or missing name"))
		return
	}
	res, err := DB.Exec(`DELETE FROM projects WHERE user_id=$1 AND name=$2`, userID, req.Name)
	if err != nil {
		log.Errorf("[HandleProjectDelete] DB error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("db error: " + err.Error()))
		return
	}
	n, _ := res.RowsAffected()
	log.Infof("[HandleProjectDelete] Deleted project '%s' for user %s (rows affected: %d)", req.Name, userID, n)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func HandleProjectPublish(w http.ResponseWriter, r *http.Request) {
	log.Infof("[HandleProjectPublish] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	if r.Method != http.MethodPost {
		log.Warnf("[HandleProjectPublish] Method not allowed: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID, err := GetUserIDFromRequest(r)
	if err != nil {
		log.Warnf("[HandleProjectPublish] Unauthorized: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized: " + err.Error()))
		return
	}
	var req struct {
		Name   string `json:"name"`
		Public bool   `json:"public"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		log.Warnf("[HandleProjectPublish] Invalid JSON or missing name: %v", err)
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
	_, err = DB.Exec(`UPDATE projects SET is_public=$1, public_id=$2 WHERE user_id=$3 AND name=$4`, req.Public, publicID, userID, req.Name)
	if err != nil {
		log.Errorf("[HandleProjectPublish] DB error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("db error: " + err.Error()))
		return
	}
	log.Infof("[HandleProjectPublish] Published project '%s' for user %s (public: %v)", req.Name, userID, req.Public)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func HandleProjectLoadPublic(w http.ResponseWriter, r *http.Request) {
	log.Infof("[HandleProjectLoadPublic] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	if r.Method != http.MethodGet {
		log.Warnf("[HandleProjectLoadPublic] Method not allowed: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	publicID := r.URL.Query().Get("id")
	if publicID == "" {
		log.Warnf("[HandleProjectLoadPublic] Missing id param")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing id param"))
		return
	}
	var data json.RawMessage
	var updated time.Time
	err := DB.QueryRow(`SELECT data, updated_at FROM projects WHERE public_id=$1 AND is_public=TRUE`, publicID).Scan(&data, &updated)
	if err == sql.ErrNoRows {
		log.Warnf("[HandleProjectLoadPublic] Not found: %s", publicID)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
		return
	} else if err != nil {
		log.Errorf("[HandleProjectLoadPublic] DB error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("db error: " + err.Error()))
		return
	}
	log.Infof("[HandleProjectLoadPublic] Loaded public project '%s'", publicID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data))
}
