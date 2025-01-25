package messaging_test

import (
	"database/sql"
	"testing"

	"purple-check/internal/database"
	"purple-check/internal/messaging"
	
	"github.com/DATA-DOG/go-sqlmock"
)

func TestSaveRating(t *testing.T) {
	// Setup mock
	db, mock, _ := sqlmock.New()
	defer db.Close()
	database.SetMockDB(db, func() { db.Close() })

	// Mock username lookup
	origGetUsername := messaging.GetUsernameFromUserID
	messaging.GetUsernameFromUserID = func(userId string) (*messaging.UserProfileAPIResponse, error) {
		return &messaging.UserProfileAPIResponse{Username: "test_giver"}, nil
	}
	defer func() { messaging.GetUsernameFromUserID = origGetUsername }()

	t.Run("Successful insert", func(t *testing.T) {
		mock.ExpectPrepare("INSERT INTO feedback .+ VALUES .+ ON CONFLICT").
			ExpectExec().
			WithArgs("test_giver", "test_receiver", "POSITIVE").
			WillReturnResult(sqlmock.NewResult(1, 1))

		messaging.SaveRating("POSITIVE", "any_user_id", "test_receiver")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unmet expectations: %s", err)
		}
	})

	t.Run("Conflict update", func(t *testing.T) {
		mock.ExpectPrepare("INSERT INTO feedback").
			ExpectExec().
			WithArgs("test_giver", "test_receiver", "NEGATIVE").
			WillReturnResult(sqlmock.NewResult(0, 1)) // 0 new rows, 1 updated

		messaging.SaveRating("NEGATIVE", "any_user_id", "test_receiver")
	})

	t.Run("Database failure", func(t *testing.T) {
		mock.ExpectPrepare("INSERT INTO feedback").
			WillReturnError(sql.ErrConnDone)

		// Should handle error internally
		messaging.SaveRating("POSITIVE", "any_user_id", "test_receiver")
	})
}
