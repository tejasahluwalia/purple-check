package profile

import (
	"database/sql"
	"log/slog"
	"net/http"
	"purple-check/internal/helpers"
	"purple-check/internal/layout"
	"purple-check/internal/models"

	turso "turso.tech/database/tursogo"
)

func NewHandler(db *turso.TursoSyncDb, conn *sql.DB) http.Handler {
	feedbackModel := models.FeedbackModel{
		Conn: conn,
	}
	return &Handler{
		DB:            db,
		FeedbackModel: feedbackModel,
	}
}

type Handler struct {
	DB *turso.TursoSyncDb
	models.FeedbackModel
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
	feedbackList, err := h.FeedbackModel.GetAllForUser(r.Context(), username)
	if err != nil {
		slog.Error("Error retreiving user feedback", err)
	}
	viewModel := ViewModel{
		Username:     username,
		FeedbackList: feedbackList,
	}
	v := layout.Handler(View(viewModel), layout.Head{})
	v.ServeHTTP(w, r)
}
