package app

import (
	"net/http"
	"net/url"
	"strings"
)

func Search(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	username := r.FormValue("search-term")
	username = strings.ToLower(username)
	if username == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	} else {
		username = strings.TrimPrefix(username, "@")
		safeUsername := url.QueryEscape(username) // import "net/url"
		http.Redirect(w, r, "/profile/"+safeUsername, http.StatusFound)
		return
	}
}
