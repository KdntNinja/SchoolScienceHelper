package utils

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"

	log "github.com/sirupsen/logrus"
)

// SetupDB connects to Postgres and ensures the projects and users tables exist.
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

	// Ensure users table exists by running migration if needed
	var exists bool
	err = db.QueryRow(`SELECT EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'users')`).Scan(&exists)
	if err != nil {
		log.Fatalf("Failed to check users table existence: %v", err)
	}
	if !exists {
		log.Info("'users' table not found, running migration from assets/users.sql...")
		migrationPath := "assets/users.sql"
		if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
			altPath := filepath.Join("/assets", "users.sql")
			altPath2 := filepath.Join("/app/assets", "users.sql")
			if _, err := os.Stat(altPath); err == nil {
				migrationPath = altPath
			} else if _, err := os.Stat(altPath2); err == nil {
				migrationPath = altPath2
			} else {
				log.Fatalf("Failed to find users.sql migration at %s, %s, or %s", migrationPath, altPath, altPath2)
			}
		}
		migration, err := os.ReadFile(migrationPath)
		if err != nil {
			log.Fatalf("Failed to read users.sql migration: %v", err)
		}
		if _, err := db.Exec(string(migration)); err != nil {
			log.Fatalf("Failed to run users.sql migration: %v", err)
		}
		log.Info("'users' table migration applied.")
	}

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
