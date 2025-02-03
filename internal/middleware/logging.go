package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"
)

func isGetMethod(r *http.Request) bool {
	return r.Method == http.MethodGet
}

func isInstagramWebhookPath(path string) bool {
	return strings.HasPrefix(path, "/webhook/instagram")
}

func isStaticPath(path string) bool {
	return strings.HasPrefix(path, "/static")
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)

		if isGetMethod(r) && !isInstagramWebhookPath(r.URL.Path) && !isStaticPath(r.URL.Path) {
			slog.InfoContext(r.Context(), "page_view",
				"path", r.URL.Path,
				"user_agent", r.UserAgent(),
				"referrer", r.Referer(),
				"duration", time.Since(start),
				"ip", r.RemoteAddr,
			)
		} else if !isStaticPath(r.URL.Path) {
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
