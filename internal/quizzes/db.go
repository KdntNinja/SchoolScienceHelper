package quizzes

import (
	"context"
	"database/sql"
)

func GetQuiz(ctx context.Context, db *sql.DB, id string) (*Quiz, error) {
	row := db.QueryRowContext(ctx, `SELECT id, owner_id, title, created_at, updated_at FROM quizzes WHERE id=$1`, id)
	var q Quiz
	err := row.Scan(&q.ID, &q.OwnerID, &q.Title, &q.CreatedAt, &q.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &q, nil
}

func UpdateQuiz(ctx context.Context, db *sql.DB, q *Quiz) error {
	_, err := db.ExecContext(ctx, `UPDATE quizzes SET title=$1, updated_at=$2 WHERE id=$3 AND owner_id=$4`,
		q.Title, q.UpdatedAt, q.ID, q.OwnerID)
	return err
}

func DeleteQuiz(ctx context.Context, db *sql.DB, id, ownerID string) error {
	_, err := db.ExecContext(ctx, `DELETE FROM quizzes WHERE id=$1 AND owner_id=$2`, id, ownerID)
	return err
}
