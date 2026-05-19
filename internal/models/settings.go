package models

import (
	"context"
	"database/sql"
	"fmt"

	"purple-check/internal/database"
)

const InstagramAccountTokenKey = "instagram_account_token"

type AppSettingModel struct {
	DB   *database.AppDB
	Conn *sql.DB
}

func (m *AppSettingModel) conn() *sql.DB {
	if m.Conn != nil {
		return m.Conn
	}
	if m.DB != nil {
		return m.DB.Conn
	}
	return nil
}

func (m *AppSettingModel) pull(ctx context.Context) error {
	if m.DB == nil {
		return nil
	}
	return m.DB.Pull(ctx)
}

func (m *AppSettingModel) push(ctx context.Context) error {
	if m.DB == nil {
		return nil
	}
	return m.DB.Push(ctx)
}

func (m *AppSettingModel) Get(ctx context.Context, key string) (string, error) {
	if err := m.pull(ctx); err != nil {
		return "", fmt.Errorf("pull app setting: %w", err)
	}
	conn := m.conn()
	if conn == nil {
		return "", fmt.Errorf("app setting database connection is nil")
	}

	var value string
	if err := conn.QueryRowContext(ctx, "SELECT value FROM app_settings WHERE key = ?", key).Scan(&value); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("app setting %q not found", key)
		}
		return "", fmt.Errorf("query app setting %q: %w", key, err)
	}
	return value, nil
}

func (m *AppSettingModel) Set(ctx context.Context, key string, value string) error {
	conn := m.conn()
	if conn == nil {
		return fmt.Errorf("app setting database connection is nil")
	}
	if _, err := conn.ExecContext(ctx, `
		INSERT INTO app_settings (key, value, updated_at)
		VALUES (?, ?, datetime('now'))
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			updated_at = datetime('now')`, key, value); err != nil {
		return fmt.Errorf("upsert app setting %q: %w", key, err)
	}
	if err := m.push(ctx); err != nil {
		return fmt.Errorf("push app setting %q: %w", key, err)
	}
	return nil
}

func (m *AppSettingModel) GetAccountToken(ctx context.Context) (string, error) {
	return m.Get(ctx, InstagramAccountTokenKey)
}

func (m *AppSettingModel) SetAccountToken(ctx context.Context, token string) error {
	return m.Set(ctx, InstagramAccountTokenKey, token)
}
