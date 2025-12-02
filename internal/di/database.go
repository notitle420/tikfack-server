package di

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

const defaultDatabaseURL = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"

// provideDatabase opens a PostgreSQL connection using the DATABASE_URL environment variable.
func provideDatabase() (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = defaultDatabaseURL
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return db, nil
}
