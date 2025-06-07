package utils

import (
	"database/sql"
	"os"

	_ "github.com/lib/pq"

	log "github.com/sirupsen/logrus"
)

// SetupDB connects to Postgres and ensures the projects table exists.
func SetupDB() *sql.DB {
	pgURL := os.Getenv("POSTGRES_DSN")
	if pgURL == "" {
		log.Fatal("POSTGRES_DSN environment variable is not set!")
	}
	db, err := sql.Open("postgres", pgURL)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}
	log.Info("Connected to Postgres DB")
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS projects (
		id SERIAL PRIMARY KEY,
		user_id TEXT NOT NULL,
		name TEXT NOT NULL,
		data JSONB NOT NULL,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		is_public BOOLEAN NOT NULL DEFAULT FALSE,
		public_id TEXT UNIQUE
	)`)
	if err != nil {
		log.Fatalf("Failed to create projects table: %v", err)
	}
	return db
}
