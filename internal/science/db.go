package science

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

func GetPapersByBoardTierAndSubject(ctx context.Context, db *sql.DB, board, tier, subject string) ([]Paper, error) {
	query := `SELECT id, board, tier, year, subject, url FROM papers WHERE board = $1 AND tier = $2 AND subject = $3`
	rows, err := db.QueryContext(ctx, query, board, tier, subject)
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

// UpsertSpec inserts or updates a spec
func UpsertSpec(ctx context.Context, db *sql.DB, s Spec) error {
	_, err := db.ExecContext(ctx, `INSERT INTO specs (board, tier, subject, title, content) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (board, tier, subject, title) DO UPDATE SET content = EXCLUDED.content`,
		s.Board, s.Tier, s.Subject, s.Title, s.Content)
	return err
}

// UpsertPaper inserts or updates a paper
func UpsertPaper(ctx context.Context, db *sql.DB, p Paper) error {
	_, err := db.ExecContext(ctx, `INSERT INTO papers (board, tier, year, subject, url) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (board, tier, year, subject, url) DO NOTHING`,
		p.Board, p.Tier, p.Year, p.Subject, p.URL)
	return err
}

// UpsertQuestion inserts or updates a question
func UpsertQuestion(ctx context.Context, db *sql.DB, q Question) error {
	_, err := db.ExecContext(ctx, `INSERT INTO questions (board, tier, subject, topic, question, answer) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (board, tier, subject, topic, question) DO UPDATE SET answer = EXCLUDED.answer`,
		q.Board, q.Tier, q.Subject, q.Topic, q.Question, q.Answer)
	return err
}

// UpsertRevision inserts or updates a revision note
func UpsertRevision(ctx context.Context, db *sql.DB, r Revision) error {
	_, err := db.ExecContext(ctx, `INSERT INTO revision (board, tier, subject, topic, content) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (board, tier, subject, topic) DO UPDATE SET content = EXCLUDED.content`,
		r.Board, r.Tier, r.Subject, r.Topic, r.Content)
	return err
}
