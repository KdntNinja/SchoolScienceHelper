package projects

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

var inMemoryProjects = []Project{}

// ListProjects handles GET /api/projects
func ListProjects(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inMemoryProjects)
}

// CreateProject handles POST /api/projects
func CreateProject(w http.ResponseWriter, r *http.Request) {
	var p Project
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	p.ID = strconv.FormatInt(time.Now().UnixNano()+int64(rand.Intn(1000)), 10)
	p.CreatedAt = time.Now().Unix()
	p.UpdatedAt = p.CreatedAt
	inMemoryProjects = append(inMemoryProjects, p)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}
