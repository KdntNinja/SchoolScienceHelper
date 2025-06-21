package quiz

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

// List all quizzes
func ListQuizzes(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		quizzes, counts, err := GetAllQuizzes(db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var resp []map[string]interface{}
		for i, q := range quizzes {
			resp = append(resp, map[string]interface{}{
				"id":             q.ID,
				"title":          q.Title,
				"description":    q.Description,
				"topic":          q.Topic,
				"question_count": counts[i],
			})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// Get a quiz by ID (with questions)
func GetQuiz(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		quizID := r.URL.Query().Get("id")
		if quizID == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		quiz, questions, err := GetQuizByID(db, quizID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		resp := map[string]interface{}{
			"id":          quiz.ID,
			"title":       quiz.Title,
			"description": quiz.Description,
			"topic":       quiz.Topic,
			"questions":   questions,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// Submit quiz attempt
func SubmitQuizAttempt(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDRaw := r.Context().Value("user_id")
		userID, ok := userIDRaw.(string)
		if !ok || userID == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var req struct {
			QuizID  string `json:"quiz_id"`
			Answers []int  `json:"answers"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, questions, err := GetQuizByID(db, req.QuizID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		score := 0
		results := make([]map[string]interface{}, len(questions))
		for i, q := range questions {
			correct := i < len(req.Answers) && req.Answers[i] == q.Answer
			if correct {
				score++
			}
			results[i] = map[string]interface{}{
				"question_id":    q.ID,
				"correct":        correct,
				"user_answer":    req.Answers[i],
				"correct_answer": q.Answer,
				"explanation":    q.Explanation,
			}
		}
		attempt := UserQuizAttempt{
			ID:        uuid.NewString(),
			UserID:    userID,
			QuizID:    req.QuizID,
			Answers:   req.Answers,
			Score:     score,
			Timestamp: "",
		}
		_ = SaveQuizAttempt(db, attempt)
		resp := map[string]interface{}{
			"score":   score,
			"total":   len(questions),
			"results": results,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
