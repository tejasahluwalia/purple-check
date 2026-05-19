package profile

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"purple-check/internal/models"
)

type feedbackRepo struct {
	username string
	err      error
}

func (repo *feedbackRepo) GetAllForUser(ctx context.Context, username string) ([]models.Feedback, error) {
	repo.username = username
	return []models.Feedback{{ID: "1", Giver: "bob", Receiver: username, Rating: "POSITIVE"}}, repo.err
}

func (repo *feedbackRepo) CountForUser(ctx context.Context, username string) (int, error) {
	return 0, nil
}

func (repo *feedbackRepo) CountPositiveForUser(ctx context.Context, username string) (int, error) {
	return 0, nil
}

func (repo *feedbackRepo) InsertOrUpdateOne(ctx context.Context, input models.FeedbackInput) error {
	return nil
}

func (repo *feedbackRepo) DeleteOne(ctx context.Context, userID string, feedbackID string) error {
	return nil
}

func (repo *feedbackRepo) DeleteAllForUser(ctx context.Context, userID string) error {
	return nil
}

func TestGetUsesFeedbackRepository(t *testing.T) {
	repo := &feedbackRepo{}
	handler := NewHandler(repo)
	req := httptest.NewRequest(http.MethodGet, "/profile/@Alice", nil)
	req.SetPathValue("username", "@Alice")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if repo.username != "alice" {
		t.Fatalf("repository username = %q, want %q", repo.username, "alice")
	}
}

func TestGetRejectsInvalidUsername(t *testing.T) {
	repo := &feedbackRepo{}
	handler := NewHandler(repo)
	req := httptest.NewRequest(http.MethodGet, "/profile/..bad", nil)
	req.SetPathValue("username", "..bad")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if repo.username != "" {
		t.Fatalf("repository should not have been called, got %q", repo.username)
	}
}

func TestGetReturnsServerErrorOnRepositoryError(t *testing.T) {
	repo := &feedbackRepo{err: errors.New("database unavailable")}
	handler := NewHandler(repo)
	req := httptest.NewRequest(http.MethodGet, "/profile/alice", nil)
	req.SetPathValue("username", "alice")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
