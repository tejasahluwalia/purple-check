package handlers

import (
	"net/http"
	"strings"
)

func HandleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusMethodNotAllowed)
		return
	}
	username := r.FormValue("username")
	username = strings.ToLower(username)
	if username == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	} else {
		http.Redirect(w, r, "/profile/"+username, http.StatusFound)
		return
	}
}