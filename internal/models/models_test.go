package models

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}

	schema := `
	CREATE TABLE feedback (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		giver TEXT,
		receiver TEXT,
		rating TEXT,
		comment TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		giver_role TEXT,
		receiver_role TEXT,
		deal_stage TEXT,
		platform TEXT DEFAULT 'INSTAGRAM' NOT NULL,
		medium TEXT DEFAULT 'DIRECT' NOT NULL,
		source TEXT DEFAULT 'DM' NOT NULL
	);
	CREATE UNIQUE INDEX idx_feedback ON feedback (giver, receiver);
	CREATE INDEX idx_feedback_receiver_rating ON feedback (receiver, rating);
	`
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestFeedbackModelNormalization(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	model := &FeedbackModel{DB: db}
	ctx := context.Background()

	// 1. Test InsertOrUpdateOne normalizes mixed case usernames
	input := FeedbackInput{
		Giver:        "Giver_User_One",
		Receiver:     "Receiver_User_One",
		Rating:       PositiveFeedback,
		GiverRole:    TransactionRoleBuyer,
		ReceiverRole: TransactionRoleSeller,
		DealStage:    DealStageComplete,
		Comment:      "Great experience",
		Platform:     "INSTAGRAM",
		Medium:       "DIRECT",
		Source:       "DM",
	}

	if err := model.InsertOrUpdateOne(ctx, input); err != nil {
		t.Fatalf("InsertOrUpdateOne failed: %v", err)
	}

	// Verify database content directly
	var dbGiver, dbReceiver string
	err := db.QueryRow("SELECT giver, receiver FROM feedback LIMIT 1").Scan(&dbGiver, &dbReceiver)
	if err != nil {
		t.Fatalf("failed to query raw database values: %v", err)
	}

	if dbGiver != "giver_user_one" {
		t.Errorf("expected giver to be lowercased to 'giver_user_one', got %q", dbGiver)
	}
	if dbReceiver != "receiver_user_one" {
		t.Errorf("expected receiver to be lowercased to 'receiver_user_one', got %q", dbReceiver)
	}

	// 2. Test GetAll with mixed case fetches successfully
	feedbacks, err := model.GetAll(ctx, "RECEIVER_USER_ONE", FeedbackRoleReceiver)
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if len(feedbacks) != 1 {
		t.Fatalf("expected 1 feedback, got %d", len(feedbacks))
	}
	if feedbacks[0].Receiver != "receiver_user_one" {
		t.Errorf("expected fetched receiver username to be normalized, got %q", feedbacks[0].Receiver)
	}

	// 3. Test DeleteOne with mixed case removes successfully
	// Let's insert another one
	input2 := FeedbackInput{
		Giver:        "Giver_Two",
		Receiver:     "Receiver_Two",
		Rating:       NegativeFeedback,
		GiverRole:    TransactionRoleBuyer,
		ReceiverRole: TransactionRoleSeller,
		DealStage:    DealStageIncomplete,
		Comment:      "Bad experience",
	}
	if err := model.InsertOrUpdateOne(ctx, input2); err != nil {
		t.Fatalf("failed to insert second feedback: %v", err)
	}

	// Get feedback to find the ID
	feedbacks2, err := model.GetAll(ctx, "receiver_two", FeedbackRoleReceiver)
	if err != nil || len(feedbacks2) != 1 {
		t.Fatalf("failed to get second feedback: %v, len: %d", err, len(feedbacks2))
	}
	feedbackID := feedbacks2[0].ID

	// Delete it using mixed case receiver
	if err := model.DeleteOne(ctx, "RECEIVER_TWO", feedbackID); err != nil {
		t.Fatalf("DeleteOne failed: %v", err)
	}

	// Verify deletion
	feedbacksAfterDelete, err := model.GetAll(ctx, "receiver_two", FeedbackRoleReceiver)
	if err != nil {
		t.Fatalf("GetAll failed after delete: %v", err)
	}
	if len(feedbacksAfterDelete) != 0 {
		t.Errorf("expected feedback to be deleted, but still found %d items", len(feedbacksAfterDelete))
	}
}
