package models

import (
	"context"
	"database/sql"
	"fmt"

	"purple-check/internal/database"
)

type Feedback struct {
	ID           string
	Giver        string
	Receiver     string
	Rating       string
	GiverRole    string `json:"giver_role"`
	ReceiverRole string `json:"receiver_role"`
	Comment      string
	CreatedAt    string `json:"created_at"`
}

type FeedbackInput struct {
	Giver        string
	Receiver     string
	Rating       string
	GiverRole    string
	ReceiverRole string
	DealStage    string
	Comment      string
}

// FeedbackModel expects the live Turso schema to provide feedback(id, giver,
// receiver, rating, giver_role, receiver_role, deal_stage, comment, created_at)
// with a unique constraint on (giver, receiver), and user_message_logs(user_id,
// message, stage, created_at). Migrations are intentionally out of scope here.
type FeedbackRepository interface {
	GetAllForUser(ctx context.Context, username string) ([]Feedback, error)
	CountForUser(ctx context.Context, username string) (int, error)
	CountPositiveForUser(ctx context.Context, username string) (int, error)
	InsertOrUpdateOne(ctx context.Context, input FeedbackInput) error
	DeleteOne(ctx context.Context, userID string, feedbackID string) error
	DeleteAllForUser(ctx context.Context, userID string) error
}

type FeedbackModel struct {
	DB *database.AppDB
}

func (m *FeedbackModel) conn() *sql.DB {
	if m.DB != nil {
		return m.DB.Conn
	}
	return nil
}

func (m *FeedbackModel) pull(ctx context.Context) error {
	if m.DB == nil {
		return nil
	}
	return m.DB.Pull(ctx)
}

func (m *FeedbackModel) push(ctx context.Context) error {
	if m.DB == nil {
		return nil
	}
	return m.DB.Push(ctx)
}

func (m *FeedbackModel) InsertOrUpdateOne(ctx context.Context, input FeedbackInput) error {
	conn := m.conn()
	if conn == nil {
		return fmt.Errorf("feedback model database connection is nil")
	}
	stmt, err := conn.PrepareContext(ctx, `INSERT INTO feedback
		(giver, receiver, rating, giver_role, receiver_role, deal_stage, comment)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(giver, receiver)
		DO UPDATE SET
			rating=excluded.rating,
			giver_role=excluded.giver_role,
			receiver_role=excluded.receiver_role,
			deal_stage=excluded.deal_stage,
			comment=excluded.comment`)
	if err != nil {
		return fmt.Errorf("prepare feedback upsert: %w", err)
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, input.Giver, input.Receiver, input.Rating, input.GiverRole, input.ReceiverRole, input.DealStage, input.Comment); err != nil {
		return fmt.Errorf("execute feedback upsert: %w", err)
	}
	if err := m.push(ctx); err != nil {
		return fmt.Errorf("push feedback upsert: %w", err)
	}
	return nil
}

func (m *FeedbackModel) GetAllForUser(ctx context.Context, username string) ([]Feedback, error) {
	if err := m.pull(ctx); err != nil {
		return nil, fmt.Errorf("pull feedback: %w", err)
	}
	conn := m.conn()
	if conn == nil {
		return nil, fmt.Errorf("feedback model database connection is nil")
	}
	var feedbackList []Feedback

	stmt, err := conn.PrepareContext(ctx, "SELECT id, giver, receiver, rating, giver_role, receiver_role, comment, created_at FROM feedback WHERE receiver = ? ORDER BY created_at DESC")
	if err != nil {
		return []Feedback{}, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, username)
	if err != nil {
		return []Feedback{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var feedback Feedback

		err = rows.Scan(&feedback.ID, &feedback.Giver, &feedback.Receiver, &feedback.Rating, &feedback.GiverRole, &feedback.ReceiverRole, &feedback.Comment, &feedback.CreatedAt)
		if err != nil {
			return []Feedback{}, err
		}

		feedbackList = append(feedbackList, feedback)
	}
	if err := rows.Err(); err != nil {
		return []Feedback{}, err
	}

	return feedbackList, nil
}

func (m *FeedbackModel) CountForUser(ctx context.Context, username string) (int, error) {
	if err := m.pull(ctx); err != nil {
		return 0, fmt.Errorf("pull feedback count: %w", err)
	}
	conn := m.conn()
	if conn == nil {
		return 0, fmt.Errorf("feedback model database connection is nil")
	}
	var count int
	if err := conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM feedback WHERE receiver = ?", username).Scan(&count); err != nil {
		return 0, fmt.Errorf("query feedback count: %w", err)
	}
	return count, nil
}

func (m *FeedbackModel) CountPositiveForUser(ctx context.Context, username string) (int, error) {
	if err := m.pull(ctx); err != nil {
		return 0, fmt.Errorf("pull positive feedback count: %w", err)
	}
	conn := m.conn()
	if conn == nil {
		return 0, fmt.Errorf("feedback model database connection is nil")
	}
	var count int
	if err := conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM feedback WHERE receiver = ? AND rating = 'POSITIVE'", username).Scan(&count); err != nil {
		return 0, fmt.Errorf("query positive feedback count: %w", err)
	}
	return count, nil
}

func (m *FeedbackModel) DeleteOne(ctx context.Context, userID string, feedbackID string) error {
	conn := m.conn()
	if conn == nil {
		return fmt.Errorf("feedback model database connection is nil")
	}
	if _, err := conn.ExecContext(ctx, "DELETE FROM feedback WHERE receiver = ? AND id = ?", userID, feedbackID); err != nil {
		return fmt.Errorf("delete feedback: %w", err)
	}
	if err := m.push(ctx); err != nil {
		return fmt.Errorf("push delete feedback: %w", err)
	}
	return nil
}

func (m *FeedbackModel) DeleteAllForUser(ctx context.Context, userID string) error {
	conn := m.conn()
	if conn == nil {
		return fmt.Errorf("feedback model database connection is nil")
	}
	if _, err := conn.ExecContext(ctx, "DELETE FROM feedback WHERE receiver = ?", userID); err != nil {
		return fmt.Errorf("delete all feedback: %w", err)
	}
	if err := m.push(ctx); err != nil {
		return fmt.Errorf("push delete all feedback: %w", err)
	}
	return nil
}

type MessageLogRepository interface {
	Insert(ctx context.Context, userID string, message string, stage string) error
}

type MessageLogModel struct {
	DB   *database.AppDB
	Conn *sql.DB
}

func (m *MessageLogModel) conn() *sql.DB {
	if m.Conn != nil {
		return m.Conn
	}
	if m.DB != nil {
		return m.DB.Conn
	}
	return nil
}

func (m *MessageLogModel) Insert(ctx context.Context, userID string, message string, stage string) error {
	conn := m.conn()
	if conn == nil {
		return fmt.Errorf("message log model database connection is nil")
	}
	if _, err := conn.ExecContext(ctx, "INSERT INTO user_message_logs (user_id, message, stage, created_at) VALUES (?, ?, ?, datetime('now'))", userID, message, stage); err != nil {
		return fmt.Errorf("insert message log: %w", err)
	}
	if m.DB != nil {
		if err := m.DB.Push(ctx); err != nil {
			return fmt.Errorf("push message log: %w", err)
		}
	}
	return nil
}
