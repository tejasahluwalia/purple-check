package database

import (
	"context"
	"database/sql"
	"sync"

	"purple-check/internal/config"

	turso "turso.tech/database/tursogo"
)

var (
	dbOnce sync.Once
	// db      *turso.TursoSyncDb
	// dbErr   error
	// conn    *sql.DB
	// connErr error
)

func InitDb(ctx context.Context) (*turso.TursoSyncDb, *sql.DB, error) {
	db, dbErr := turso.NewTursoSyncDb(ctx, turso.TursoSyncDbConfig{
		Path:      config.LOCAL_DB_PATH,
		RemoteUrl: config.TURSO_DATABASE_URL,
		AuthToken: config.TURSO_AUTH_TOKEN,
	})
	if dbErr != nil {
		return nil, nil, dbErr
	}

	conn, connErr := db.Connect(ctx)
	if connErr != nil {
		return nil, nil, connErr
	}
	return db, conn, nil
}

// func GetDB(ctx context.Context) (*sql.DB, func(), error) {
// 	db, err := InitDb(ctx)
// 	if err != nil {
// 		return nil, func() {}, fmt.Errorf("open database: %w", err)
// 	}

// 	conn, err := db.Connect(ctx)
// 	if err != nil {
// 		return nil, func() {}, fmt.Errorf("connect database: %w", err)
// 	}

// 	return conn, func() {
// 		if err := conn.Close(); err != nil {
// 			log.Println("Error closing database connection.", err)
// 		}
// 	}, nil
// }

// func PullDB(ctx context.Context) (bool, error) {
// 	db, err := InitDb(ctx)
// 	if err != nil {
// 		return false, err
// 	}

// 	return db.Pull(ctx)
// }

// func PushDB(ctx context.Context) error {
// 	db, err := InitDb(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	return db.Push(ctx)
// }

// func StatsDB(ctx context.Context) (turso.TursoSyncDbStats, error) {
// 	db, err := InitDb(ctx)
// 	if err != nil {
// 		return turso.TursoSyncDbStats{}, err
// 	}

// 	return db.Stats(ctx)
// }

// func CheckpointDB(ctx context.Context) {
// 	db, err := InitDb(ctx)
// 	if err != nil {
// 		log.Println("Error opening database.", err)
// 		return
// 	}

// 	if err := db.Checkpoint(ctx); err != nil {
// 		log.Println("Error checkpointing database.", err)
// 	}
// }
