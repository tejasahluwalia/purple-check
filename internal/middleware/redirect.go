package middleware

import (
	"net/http"
)

func RedirectNonWWW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host == "purple-check.org" {
			http.Redirect(w, r, "https://www.purple-check.org"+r.URL.Path, http.StatusPermanentRedirect)
			return
		}
		next.ServeHTTP(w, r)
	})
}
