package profile

import (
	"log/slog"
	"net/http"
	"purple-check/internal/helpers"
	"purple-check/internal/layout"
	"purple-check/internal/models"
)

func NewHandler(feedbacks models.FeedbackRepository) http.Handler {
	return &Handler{
		Feedbacks: feedbacks,
	}
}

type Handler struct {
	Feedbacks models.FeedbackRepository
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.Get(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	username := helpers.NormalizeUsername(r.PathValue("username"))
	err := helpers.ValidateUsername(username)
	if err != nil {
		http.Error(w, "Invalid username", http.StatusBadRequest)
		return
	}
	feedbackList, err := h.Feedbacks.GetAllForUser(r.Context(), username)
	if err != nil {
		slog.Error("Error retrieving user feedback", "error", err)
	}
	viewModel := ViewModel{
		Username:     username,
		FeedbackList: feedbackList,
	}
	v := layout.Handler(View(viewModel), layout.Head{})
	v.ServeHTTP(w, r)
}
