package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func reset() {

    db, err := sql.Open("sqlite3", "purple-check.db")

    if err != nil {
        log.Fatal(err)
    }

    defer db.Close()

	sts := `
		DROP TABLE IF EXISTS profiles;
        CREATE TABLE profiles(id INTEGER PRIMARY KEY, platform TEXT, platform_user_id TEXT, username TEXT, status TEXT DEFAULT 'not-connected', token TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP);
        CREATE UNIQUE INDEX idx_profiles_platform_user_id ON profiles(platform, platform_user_id);

		DROP TABLE IF EXISTS feedback;
		CREATE TABLE feedback(id INTEGER PRIMARY KEY, giver_id INTEGER, receiver_id INTEGER, comment TEXT, giver_role TEXT, receiver_role TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP);
		`
    _, err = db.Exec(sts)

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("table profiles, feedback created")
}
