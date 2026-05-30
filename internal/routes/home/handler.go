package home

import (
	"cmp"
	"log/slog"
	"net/http"
	"purple-check/internal/layout"
	"purple-check/internal/models"
	"slices"
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
	receiverStats, err := h.repo.GetAllReceivers(r.Context())
	if err != nil {
		slog.Error("failed to fetch top stores for home page", "error", err)
		// Render page without stats on error — non-critical.
		receiverStats = []models.ReceiverStats{}
	}

	var vm ViewModel

	slices.SortFunc(receiverStats, func(a, b models.ReceiverStats) int {
		return cmp.Compare(b.Score, a.Score)
	})
	vm.receivers.MostPositive = slices.Clone(receiverStats[:20])

	slices.SortFunc(receiverStats, func(a, b models.ReceiverStats) int {
		return cmp.Compare(1-b.Score, 1-a.Score)
	})
	vm.receivers.MostNegative = slices.Clone(receiverStats[:20])
	v := layout.Handler(View(vm), layout.Head{
		Title: "Instagram Shop Reviews on Purple Check",
	})
	v.ServeHTTP(w, r)
}
