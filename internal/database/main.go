package database

import (
	"context"
	"database/sql"
	"fmt"

	"purple-check/internal/config"

	turso "turso.tech/database/tursogo"
)

type AppDB struct {
	Sync *turso.TursoSyncDb
	Conn *sql.DB
}

func InitDb(ctx context.Context) (*AppDB, error) {
	db, dbErr := turso.NewTursoSyncDb(ctx, turso.TursoSyncDbConfig{
		Path:      config.LOCAL_DB_PATH,
		RemoteUrl: config.TURSO_DATABASE_URL,
		AuthToken: config.TURSO_AUTH_TOKEN,
	})
	if dbErr != nil {
		return nil, dbErr
	}

	conn, connErr := db.Connect(ctx)
	if connErr != nil {
		return nil, connErr
	}
	return &AppDB{Sync: db, Conn: conn}, nil
}

func (db *AppDB) Close() error {
	if db == nil || db.Conn == nil {
		return nil
	}
	return db.Conn.Close()
}

func (db *AppDB) Pull(ctx context.Context) error {
	if db == nil || db.Sync == nil {
		return fmt.Errorf("database sync handle is nil")
	}
	_, err := db.Sync.Pull(ctx)
	return err
}

func (db *AppDB) Push(ctx context.Context) error {
	if db == nil || db.Sync == nil {
		return fmt.Errorf("database sync handle is nil")
	}
	return db.Sync.Push(ctx)
}

func (db *AppDB) Checkpoint(ctx context.Context) error {
	if db == nil || db.Sync == nil {
		return fmt.Errorf("database sync handle is nil")
	}
	return db.Sync.Checkpoint(ctx)
}
