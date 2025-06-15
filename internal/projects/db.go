package projects

import (
	"context"
	"database/sql"
)

func GetProject(ctx context.Context, db *sql.DB, id string) (*Project, error) {
	row := db.QueryRowContext(ctx, `SELECT id, owner_id, title, created_at, updated_at, data FROM projects WHERE id=$1`, id)
	var p Project
	err := row.Scan(&p.ID, &p.OwnerID, &p.Title, &p.CreatedAt, &p.UpdatedAt, &p.Data)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func UpdateProject(ctx context.Context, db *sql.DB, p *Project) error {
	_, err := db.ExecContext(ctx, `UPDATE projects SET title=$1, updated_at=$2, data=$3 WHERE id=$4 AND owner_id=$5`,
		p.Title, p.UpdatedAt, p.Data, p.ID, p.OwnerID)
	return err
}

func DeleteProject(ctx context.Context, db *sql.DB, id, ownerID string) error {
	_, err := db.ExecContext(ctx, `DELETE FROM projects WHERE id=$1 AND owner_id=$2`, id, ownerID)
	return err
}
