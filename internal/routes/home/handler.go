package home

import (
	"log/slog"
	"net/http"
	"purple-check/internal/layout"
	"purple-check/internal/models"
)

func NewHandler(repo models.FeedbackRepository) http.Handler {
	return &Handler{repo: repo}
}

type Handler struct {
	repo models.FeedbackRepository
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
	topStores, err := h.repo.GetTopStores(r.Context())
	if err != nil {
		slog.Error("failed to fetch top stores for home page", "error", err)
		// Render page without stats on error — non-critical.
		topStores = &models.TopStores{}
	}

	v := layout.Handler(View(topStores), layout.Head{
		Title: "Instagram Shop Reviews on Purple Check",
	})
	v.ServeHTTP(w, r)
}
