package database

import (
	"database/sql"
	"fmt"
	"os"

	"purple-check/internal/config"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

var getDB = func() (*sql.DB, func()) {
	url := config.TURSO_DATABASE_URL + "?authToken=" + config.TURSO_AUTH_TOKEN

	db, err := sql.Open("libsql", url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db %s: %s", url, err)
		os.Exit(1)
	}

	return db, func() {
		db.Close()
	}
}

func GetDB() (*sql.DB, func()) {
	return getDB()
}
