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
		slog.InfoContext(r.Context(),
			"request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start),
			"remote_ip", r.RemoteAddr,
		)
	})
}
