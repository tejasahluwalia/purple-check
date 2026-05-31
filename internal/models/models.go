package models

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"purple-check/internal/helpers"
)

type DealStage string

const (
	DealStageIncomplete DealStage = "INCOMPLETE"
	DealStageComplete   DealStage = "COMPLETE"
	DealStageNA         DealStage = ""
)

type FeedbackSentiment string

const (
	PositiveFeedback FeedbackSentiment = "POSITIVE"
	NegativeFeedback FeedbackSentiment = "NEGATIVE"
	MixedFeedback    FeedbackSentiment = "MIXED"
	NoFeedback       FeedbackSentiment = "NO_FEEDBACK"
)

type FeedbackRole string

const (
	FeedbackRoleGiver    FeedbackRole = "GIVER"
	FeedbackRoleReceiver FeedbackRole = "RECEIVER"
)

type TransactionRole string

const (
	TransactionRoleBuyer  TransactionRole = "BUYER"
	TransactionRoleSeller TransactionRole = "SELLER"
)

type Feedback struct {
	ID           string
	Giver        string
	Receiver     string
	GiverRole    TransactionRole `json:"giver_role"`
	ReceiverRole TransactionRole `json:"receiver_role"`
	Rating       FeedbackSentiment
	DealStage    DealStage
	Comment      sql.NullString
	Platform     string
	Medium       string
	Source       string
	CreatedAt    string `json:"created_at"`
}

type FeedbackInput struct {
	Giver        string
	Receiver     string
	GiverRole    TransactionRole `json:"giver_role"`
	ReceiverRole TransactionRole `json:"receiver_role"`
	Rating       FeedbackSentiment
	DealStage    DealStage
	Comment      string
	Platform     string
	Medium       string
	Source       string
}

type ReceiverStats struct {
	Username      string
	PositiveCount int
	NegativeCount int
	MixedCount    int
	TotalCount    int
	Score         float64
}

// WilsonLowerBound computes the lower bound of the Wilson score interval
// for a binomial proportion with 95% confidence (z=1.96).
// Returns 0 when total is 0.
func WilsonLowerBound(positive, total int) float64 {
	if total == 0 {
		return 0
	}
	n := float64(total)
	p := float64(positive) / n
	z := 1.96
	z2 := z * z

	numerator := p + z2/(2*n) - z*math.Sqrt((p*(1-p)+z2/(4*n))/n)
	denominator := 1 + z2/n
	return numerator / denominator
}

// FeedbackModel expects the live Turso schema to provide feedback(id, giver,
// receiver, rating, giver_role, receiver_role, deal_stage, comment, created_at)
// with a unique constraint on (giver, receiver), and user_message_logs(user_id,
// message, stage, created_at). Migrations are intentionally out of scope here.
type FeedbackRepository interface {
	GetAll(ctx context.Context, username string, feedbackRole FeedbackRole) ([]Feedback, error)
	GetAllReceivers(ctx context.Context) ([]ReceiverStats, error)
	InsertOrUpdateOne(ctx context.Context, input FeedbackInput) error
	DeleteOne(ctx context.Context, username string, feedbackID string) error
	DeleteAll(ctx context.Context, username string) error
}

type FeedbackModel struct {
	DB *sql.DB
}

func (m *FeedbackModel) InsertOrUpdateOne(ctx context.Context, input FeedbackInput) error {
	db := m.DB
	if db == nil {
		return fmt.Errorf("feedback model database connection is nil")
	}

	input.Giver = helpers.NormalizeUsername(input.Giver)
	input.Receiver = helpers.NormalizeUsername(input.Receiver)

	stmt, err := db.PrepareContext(ctx, `INSERT INTO feedback
		(giver, receiver, rating, giver_role, receiver_role, deal_stage, comment, platform, medium, source)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(giver, receiver)
		DO UPDATE SET
			rating=excluded.rating,
			giver_role=excluded.giver_role,
			receiver_role=excluded.receiver_role,
			deal_stage=excluded.deal_stage,
			comment=excluded.comment,
			platform=excluded.platform,
			medium=excluded.medium,
			source=excluded.source`)
	if err != nil {
		return fmt.Errorf("prepare feedback upsert: %w", err)
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, input.Giver, input.Receiver, input.Rating, input.GiverRole, input.ReceiverRole, input.DealStage, input.Comment, input.Platform, input.Medium, input.Source); err != nil {
		return fmt.Errorf("execute feedback upsert: %w", err)
	}
	return nil
}

func (m *FeedbackModel) GetAll(ctx context.Context, username string, feedbackRole FeedbackRole) ([]Feedback, error) {
	db := m.DB
	if db == nil {
		return nil, fmt.Errorf("feedback model database connection is nil")
	}
	var feedbackList []Feedback

	username = helpers.NormalizeUsername(username)

	var query string
	if feedbackRole == FeedbackRoleGiver {
		query = "SELECT id, giver, receiver, rating, giver_role, receiver_role, COALESCE(deal_stage, ''), COALESCE(comment, ''), platform, medium, source, created_at FROM feedback WHERE giver = ? ORDER BY created_at DESC"
	} else {
		query = "SELECT id, giver, receiver, rating, giver_role, receiver_role, COALESCE(deal_stage, ''), COALESCE(comment, ''), platform, medium, source, created_at FROM feedback WHERE receiver = ? ORDER BY created_at DESC"
	}

	stmt, err := db.PrepareContext(ctx, query)
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

		err = rows.Scan(&feedback.ID, &feedback.Giver, &feedback.Receiver, &feedback.Rating, &feedback.GiverRole, &feedback.ReceiverRole, &feedback.DealStage, &feedback.Comment, &feedback.Platform, &feedback.Medium, &feedback.Source, &feedback.CreatedAt)
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

func (m *FeedbackModel) GetAllReceivers(ctx context.Context) ([]ReceiverStats, error) {
	db := m.DB
	if db == nil {
		return nil, fmt.Errorf("feedback model database connection is nil")
	}

	var result []ReceiverStats

	rows, err := db.QueryContext(ctx,
		`SELECT receiver,
		        COUNT(*) as total_count,
		        SUM(CASE WHEN rating = 'POSITIVE' THEN 1 ELSE 0 END) as positive_count,
		        SUM(CASE WHEN rating = 'NEGATIVE' THEN 1 ELSE 0 END) as negative_count,
		        SUM(CASE WHEN rating = 'MIXED' THEN 1 ELSE 0 END) as mixed_count
		 FROM feedback
		 GROUP BY receiver`)
	if err != nil {
		return nil, fmt.Errorf("query all receivers: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var s ReceiverStats
		if err := rows.Scan(&s.Username, &s.TotalCount, &s.PositiveCount, &s.NegativeCount, &s.MixedCount); err != nil {
			return nil, fmt.Errorf("scan all receivers: %w", err)
		}
		s.Score = CalculateScore(s.PositiveCount, s.MixedCount, s.NegativeCount)
		result = append(result, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate all receivers: %w", err)
	}
	return result, nil
}

func (m *FeedbackModel) DeleteOne(ctx context.Context, userID string, feedbackID string) error {
	db := m.DB
	if db == nil {
		return fmt.Errorf("feedback model database connection is nil")
	}
	userID = helpers.NormalizeUsername(userID)
	if _, err := db.ExecContext(ctx, "DELETE FROM feedback WHERE receiver = ? AND id = ?", userID, feedbackID); err != nil {
		return fmt.Errorf("delete feedback: %w", err)
	}
	return nil
}

func (m *FeedbackModel) DeleteAll(ctx context.Context, userID string) error {
	db := m.DB
	if db == nil {
		return fmt.Errorf("feedback model database connection is nil")
	}
	userID = helpers.NormalizeUsername(userID)
	if _, err := db.ExecContext(ctx, "DELETE FROM feedback WHERE receiver = ?", userID); err != nil {
		return fmt.Errorf("delete all feedback: %w", err)
	}
	return nil
}

type MessageLogRepository interface {
	Insert(ctx context.Context, userID string, message string, stage string) error
}

type MessageLogModel struct {
	DB *sql.DB
}

func (m *MessageLogModel) Insert(ctx context.Context, userID string, message string, stage string) error {
	db := m.DB
	if db == nil {
		return fmt.Errorf("message log model database connection is nil")
	}
	if _, err := db.ExecContext(ctx, "INSERT INTO user_message_logs (user_id, message, stage, created_at) VALUES (?, ?, ?, datetime('now'))", userID, message, stage); err != nil {
		return fmt.Errorf("insert message log: %w", err)
	}
	return nil
}
