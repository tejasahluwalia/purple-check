package db

import (
	"database/sql"
	"fmt"
	"os"
	"purple-check/internal/config"

	"github.com/tursodatabase/go-libsql"
)

func init() {
	primaryUrl := config.TURSO_DATABASE_URL
	authToken := config.TURSO_AUTH_TOKEN
	dbPath := config.LOCAL_DB_PATH

	connector, err := libsql.NewEmbeddedReplicaConnector(dbPath, primaryUrl,
		libsql.WithAuthToken(authToken),
	)
	if err != nil {
		fmt.Println("Error creating connector:", err)
		os.Exit(1)
	}
	defer connector.Close()

	DB = sql.OpenDB(connector)

	defer DB.Close()
}

var DB *sql.DB

