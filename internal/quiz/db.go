package quiz

import (
	"database/sql"
	"encoding/json"
)

func GetAllQuizzes(db *sql.DB) ([]Quiz, []int, error) {
	rows, err := db.Query(`SELECT id, title, description, topic FROM quizzes`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var quizzes []Quiz
	var counts []int
	for rows.Next() {
		var q Quiz
		if err := rows.Scan(&q.ID, &q.Title, &q.Description, &q.Topic); err != nil {
			return nil, nil, err
		}
		var count int
		_ = db.QueryRow(`SELECT COUNT(*) FROM questions WHERE quiz_id = $1`, q.ID).Scan(&count)
		quizzes = append(quizzes, q)
		counts = append(counts, count)
	}
	return quizzes, counts, nil
}

func GetQuizByID(db *sql.DB, quizID string) (*Quiz, []Question, error) {
	var q Quiz
	if err := db.QueryRow(`SELECT id, title, description, topic FROM quizzes WHERE id = $1`, quizID).Scan(&q.ID, &q.Title, &q.Description, &q.Topic); err != nil {
		return nil, nil, err
	}
	rows, err := db.Query(`SELECT id, quiz_id, prompt, options, answer, explanation, difficulty FROM questions WHERE quiz_id = $1`, quizID)
	if err != nil {
		return &q, nil, err
	}
	defer rows.Close()
	var questions []Question
	for rows.Next() {
		var ques Question
		var optionsJSON string
		if err := rows.Scan(&ques.ID, &ques.QuizID, &ques.Prompt, &optionsJSON, &ques.Answer, &ques.Explanation, &ques.Difficulty); err != nil {
			return &q, nil, err
		}
		_ = json.Unmarshal([]byte(optionsJSON), &ques.Options)
		questions = append(questions, ques)
	}
	return &q, questions, nil
}

func SaveQuizAttempt(db *sql.DB, attempt UserQuizAttempt) error {
	answersJSON, _ := json.Marshal(attempt.Answers)
	_, err := db.Exec(`INSERT INTO user_quiz_attempts (id, user_id, quiz_id, answers, score, timestamp) VALUES ($1, $2, $3, $4, $5, NOW())`,
		attempt.ID, attempt.UserID, attempt.QuizID, string(answersJSON), attempt.Score)
	return err
}
