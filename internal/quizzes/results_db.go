package quizzes

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

func CreateQuizResult(ctx context.Context, db *sql.DB, r *QuizResult) error {
	_, err := db.ExecContext(ctx, `INSERT INTO quiz_results (id, quiz_id, user_id, score, started_at, ended_at, answers) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		r.ID, r.QuizID, r.UserID, r.Score, r.StartedAt, r.EndedAt, pq.Array(r.Answers))
	return err
}

func ListQuizResults(ctx context.Context, db *sql.DB, userID string) ([]*QuizResult, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, quiz_id, user_id, score, started_at, ended_at, answers FROM quiz_results WHERE user_id=$1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []*QuizResult
	for rows.Next() {
		var r QuizResult
		if err := rows.Scan(&r.ID, &r.QuizID, &r.UserID, &r.Score, &r.StartedAt, &r.EndedAt, pq.Array(&r.Answers)); err != nil {
			return nil, err
		}
		results = append(results, &r)
	}
	return results, nil
}
