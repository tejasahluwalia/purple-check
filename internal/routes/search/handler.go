package search

import (
	"net/http"
	"net/url"
	"purple-check/internal/helpers"
)

func NewHandler() http.Handler {
	return &Handler{}
}

type Handler struct {
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	username := helpers.NormalizeUsername(r.FormValue("search-term"))
	err := helpers.ValidateUsername(username)
	if err != nil {
		http.Error(w, "Invalid username", http.StatusBadRequest)
	} else {
		http.Redirect(w, r, "/profile/"+url.PathEscape(username), http.StatusFound)
	}
}
