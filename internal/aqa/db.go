package aqa

import (
	"context"
	"database/sql"
)

// These functions now filter by board and tier

func GetSpecsByBoardTier(ctx context.Context, db *sql.DB, board, tier string) ([]Spec, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, board, tier, subject, title, content FROM specs WHERE board = $1 AND tier = $2`, board, tier)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var specs []Spec
	for rows.Next() {
		var s Spec
		if err := rows.Scan(&s.ID, &s.Board, &s.Tier, &s.Subject, &s.Title, &s.Content); err != nil {
			return nil, err
		}
		specs = append(specs, s)
	}
	return specs, nil
}

func GetPapersByBoardTier(ctx context.Context, db *sql.DB, board, tier string) ([]Paper, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, board, tier, year, subject, url FROM papers WHERE board = $1 AND tier = $2`, board, tier)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var papers []Paper
	for rows.Next() {
		var p Paper
		if err := rows.Scan(&p.ID, &p.Board, &p.Tier, &p.Year, &p.Subject, &p.URL); err != nil {
			return nil, err
		}
		papers = append(papers, p)
	}
	return papers, nil
}

func GetQuestionsByBoardTier(ctx context.Context, db *sql.DB, board, tier string) ([]Question, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, board, tier, subject, topic, question, answer FROM questions WHERE board = $1 AND tier = $2`, board, tier)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var questions []Question
	for rows.Next() {
		var q Question
		if err := rows.Scan(&q.ID, &q.Board, &q.Tier, &q.Subject, &q.Topic, &q.Question, &q.Answer); err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}
	return questions, nil
}

func GetRevisionByBoardTier(ctx context.Context, db *sql.DB, board, tier string) ([]Revision, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, board, tier, subject, topic, content FROM revision WHERE board = $1 AND tier = $2`, board, tier)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var revs []Revision
	for rows.Next() {
		var r Revision
		if err := rows.Scan(&r.ID, &r.Board, &r.Tier, &r.Subject, &r.Topic, &r.Content); err != nil {
			return nil, err
		}
		revs = append(revs, r)
	}
	return revs, nil
}
