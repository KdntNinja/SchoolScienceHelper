package projects

import (
	"net/http"
)

// ListProjects handles GET /api/projects
func ListProjects(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
	w.Write([]byte("[]"))
}

// CreateProject handles POST /api/projects
func CreateProject(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
	w.WriteHeader(http.StatusCreated)
}
