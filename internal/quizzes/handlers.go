package quizzes

import (
	"net/http"
)

// ListQuizzes handles GET /api/quizzes
func ListQuizzes(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
	w.Write([]byte("[]"))
}

// CreateQuiz handles POST /api/quizzes
func CreateQuiz(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
	w.WriteHeader(http.StatusCreated)
}
