package database

import (
	"database/sql"
	"fmt"
	"os"

	"purple-check/internal/config"

	"github.com/tursodatabase/go-libsql"
)

func getConnector() *libsql.Connector {
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

	return connector
}

func GetDB() (*sql.DB, func()) {
	connector := getConnector()

	db := sql.OpenDB(connector)

	return db, func() {
		connector.Close()
		db.Close()
	}
}

func SyncDB() {
	connector := getConnector()
	connector.Close()
}
