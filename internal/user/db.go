package user

import (
	"context"
	"database/sql"
)

func GetUserProfile(ctx context.Context, db *sql.DB, id string) (*UserProfile, error) {
	row := db.QueryRowContext(ctx, `SELECT id, email, username, created_at, avatar_url FROM users WHERE id=$1`, id)
	var u UserProfile
	err := row.Scan(&u.ID, &u.Email, &u.Username, &u.CreatedAt, &u.AvatarURL)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func UpdateUserProfile(ctx context.Context, db *sql.DB, u *UserProfile) error {
	_, err := db.ExecContext(ctx, `UPDATE users SET username=$1, avatar_url=$2 WHERE id=$3`, u.Username, u.AvatarURL, u.ID)
	return err
}
