package main

import (
	"fmt"
	"log"
	"purple-check/internal/db"
)

func main() {
	db, closer := db.GetDB()
	defer closer()
	
	sts := `
		DROP TABLE IF EXISTS profiles;
        CREATE TABLE profiles(
            id INTEGER PRIMARY KEY,
            platform TEXT,
            platform_user_id TEXT,
            username TEXT,
            status TEXT DEFAULT 'not-connected',
            token TEXT, expires_in INTEGER,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
        CREATE UNIQUE INDEX idx_profiles_platform_user_id ON profiles(platform, platform_user_id);
        CREATE UNIQUE INDEX idx_profiles_platform_username ON profiles(platform, username);

		DROP TABLE IF EXISTS feedback;
		CREATE TABLE feedback(
            id INTEGER PRIMARY KEY,
            giver_id INTEGER,
            receiver_id INTEGER,
            rating INTEGER,
            comment TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );

        DROP TABLE IF EXISTS logs;
        CREATE TABLE logs(
            id INTEGER PRIMARY KEY,
            level INTEGER,
            message TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );

        DROP TABLE IF EXISTS tickets;
        CREATE TABLE tickets(
            id INTEGER PRIMARY KEY,
            user_id INTEGER,
            user_email TEXT,
            message TEXT,
            status TEXT DEFAULT 'open',
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
		`
	_, err := db.Exec(sts)

	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println("table profiles, feedback created")
}
