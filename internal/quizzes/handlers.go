package quizzes

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

var inMemoryQuizzes = []Quiz{}

// ListQuizzes handles GET /api/quizzes
func ListQuizzes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inMemoryQuizzes)
}

// CreateQuiz handles POST /api/quizzes
func CreateQuiz(w http.ResponseWriter, r *http.Request) {
	var q Quiz
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &q); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	q.ID = strconv.FormatInt(time.Now().UnixNano()+int64(rand.Intn(1000)), 10)
	q.CreatedAt = time.Now().Unix()
	q.UpdatedAt = q.CreatedAt
	inMemoryQuizzes = append(inMemoryQuizzes, q)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(q)
}
