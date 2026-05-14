package models

import (
	"context"
	"database/sql"
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

type FeedbackModel struct {
	Conn *sql.DB
}

func (m *FeedbackModel) InsertOrUpdateOne(giver_role string, giver string, receiver string, rating string, comment string) (string, error) {
	return "", nil
}

func (m *FeedbackModel) GetAllForUser(ctx context.Context, username string) ([]Feedback, error) {
	var feedbackList []Feedback

	stmt, err := m.Conn.PrepareContext(ctx, "SELECT id, giver, receiver, rating, created_at FROM feedback WHERE receiver = ? ORDER BY created_at DESC")
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

		err = rows.Scan(&feedback.ID, &feedback.Giver, &feedback.Receiver, &feedback.Rating, &feedback.CreatedAt)
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

func (m *FeedbackModel) DeleteOne(user_id string, feedback_id string) error {
	return nil
}

func (m *FeedbackModel) DeleteAllForUser(user_id string) error {
	return nil
}
