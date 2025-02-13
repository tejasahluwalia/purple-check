package app

import (
	"log/slog"
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
	username = strings.TrimPrefix(username, "@")
	// Instagram usernames can only be 30 chars or less
	if len(username) > 30 {
		username = username[:30]
	}
	// Restrict to letters, numbers, periods and underscores
	allowed := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789._"
	filtered := ""
	for _, char := range username {
		if strings.Contains(allowed, string(char)) {
			filtered += string(char)
		}
	}
	username = filtered
	slog.Info("searching for username", "username", username)
	if username == "" {
		http.Error(w, "Invalid username", http.StatusBadRequest)
	} else {
		http.Redirect(w, r, "/profile/"+url.QueryEscape(username), http.StatusFound)
	}
}
