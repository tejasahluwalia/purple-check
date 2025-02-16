package app

import (
	"net/http"
	"net/url"
	"purple-check/internal/helpers"
	"strings"
)

func Search(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	username := r.FormValue("search-term")
	username = strings.ToLower(username)
	username = strings.TrimPrefix(username, "@")

	isValid := helpers.ValidateUsername(username)

	if username == "" || !isValid {
		http.Error(w, "Invalid username", http.StatusBadRequest)
	} else {
		http.Redirect(w, r, "/profile/"+url.QueryEscape(username), http.StatusFound)
	}
}
