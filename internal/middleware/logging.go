package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		if r.Method == http.MethodGet && r.URL.Path != "/webhook/instagram" && r.URL.Path != "/webhook/instagram/setup" {
			slog.InfoContext(r.Context(), "page_view",
				"path", r.URL.Path,
				"user_agent", r.UserAgent(),
				"referrer", r.Referer(),
				"duration", time.Since(start),
				"ip", r.RemoteAddr,
			)
		} else {
			slog.InfoContext(r.Context(),
				"request",
				"method", r.Method,
				"path", r.URL.Path,
				"duration", time.Since(start),
				"ip", r.RemoteAddr,
			)
		}
	})
}
