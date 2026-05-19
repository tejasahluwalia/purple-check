package database

import (
	"context"
	"database/sql"

	"purple-check/internal/config"

	_ "modernc.org/sqlite"
)

func InitDb(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite", config.LOCAL_DB_PATH)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func Close(db *sql.DB) error {
	if db == nil {
		return nil
	}
	return db.Close()
}
