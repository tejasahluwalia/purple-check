package profile

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"purple-check/internal/models"
)

type fakeFeedbackRepository struct {
	GetAllFunc func(ctx context.Context, username string, feedbackRole models.FeedbackRole) ([]models.Feedback, error)
}

func (f *fakeFeedbackRepository) GetAll(ctx context.Context, username string, feedbackRole models.FeedbackRole) ([]models.Feedback, error) {
	if f.GetAllFunc != nil {
		return f.GetAllFunc(ctx, username, feedbackRole)
	}
	return nil, nil
}

func (f *fakeFeedbackRepository) GetAllReceivers(ctx context.Context) ([]models.ReceiverStats, error) {
	return nil, nil
}

func (f *fakeFeedbackRepository) InsertOrUpdateOne(ctx context.Context, input models.FeedbackInput) error {
	return nil
}

func (f *fakeFeedbackRepository) DeleteOne(ctx context.Context, username string, feedbackID string) error {
	return nil
}

func (f *fakeFeedbackRepository) DeleteAll(ctx context.Context, username string) error {
	return nil
}

func TestHandlerGet_Valid(t *testing.T) {
	fakeRepo := &fakeFeedbackRepository{
		GetAllFunc: func(ctx context.Context, username string, feedbackRole models.FeedbackRole) ([]models.Feedback, error) {
			if username != "test_user" {
				t.Errorf("expected username 'test_user', got %q", username)
			}
			if feedbackRole != models.FeedbackRoleReceiver {
				t.Errorf("expected feedback role RECEIVER, got %v", feedbackRole)
			}
			return []models.Feedback{
				{
					Giver:    "giver1",
					Receiver: "test_user",
					Rating:   models.PositiveFeedback,
				},
				{
					Giver:    "giver2",
					Receiver: "test_user",
					Rating:   models.NegativeFeedback,
				},
				{
					Giver:    "giver3",
					Receiver: "test_user",
					Rating:   models.MixedFeedback,
				},
			}, nil
		},
	}

	handler := NewHandler(fakeRepo)
	req := httptest.NewRequest(http.MethodGet, "/profile/test_user", nil)
	req.SetPathValue("username", "test_user")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestHandlerGet_InvalidUsername(t *testing.T) {
	fakeRepo := &fakeFeedbackRepository{}
	handler := NewHandler(fakeRepo)

	// Usernames must be 3-30 chars. "a" is too short.
	req := httptest.NewRequest(http.MethodGet, "/profile/a", nil)
	req.SetPathValue("username", "a")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestHandlerGet_RepositoryError(t *testing.T) {
	fakeRepo := &fakeFeedbackRepository{
		GetAllFunc: func(ctx context.Context, username string, feedbackRole models.FeedbackRole) ([]models.Feedback, error) {
			return nil, errors.New("database connection lost")
		},
	}

	handler := NewHandler(fakeRepo)
	req := httptest.NewRequest(http.MethodGet, "/profile/test_user", nil)
	req.SetPathValue("username", "test_user")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
}
