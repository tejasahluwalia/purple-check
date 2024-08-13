package app

import (
	"net/http"
	"strings"
)

func Search(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusMethodNotAllowed)
		return
	}
	username := r.FormValue("search-term")
	username = strings.ToLower(username)
	if username == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	} else {
		http.Redirect(w, r, "/profile/"+username, http.StatusFound)
		return
	}
}