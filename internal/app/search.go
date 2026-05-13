package app

import (
	"net/http"
	"net/url"
	"purple-check/internal/helpers"
)

func Search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	username, ok := helpers.NormalizeUsername(r.FormValue("search-term"))
	if !ok {
		http.Error(w, "Invalid username", http.StatusBadRequest)
	} else {
		http.Redirect(w, r, "/profile/"+url.PathEscape(username), http.StatusFound)
	}
}
