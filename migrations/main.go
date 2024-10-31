package main

import (
	"fmt"
	"log"
	"purple-check/internal/database"
)

func main() {
	db, closer := database.GetDB()
	defer closer()

	sts := `
		DROP TABLE IF EXISTS profiles;
  		DROP TABLE IF EXISTS logs;
        DROP TABLE IF EXISTS tickets;
		DROP TABLE IF EXISTS feedback;

		CREATE TABLE feedback(
            id INTEGER PRIMARY KEY,
            giver TEXT,
            receiver TEXT,
            rating INTEGER,
            comment TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
		`
	_, err := db.Exec(sts)

	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println("Database recreated.")
}
