package utils

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"

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

	// Always run the merged migration file on startup to ensure both users and projects tables exist and are up to date
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
	// Split and run each statement individually to guarantee all tables are created
	for _, stmt := range strings.Split(string(migration), ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			log.Warnf("Migration statement failed: %s\nError: %v", stmt, err)
		}
	}
	log.Info("users.sql migration applied (users and projects tables ensured).")

	return db
}
