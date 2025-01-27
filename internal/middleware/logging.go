package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"purple-check/internal/helpers"
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

		body := helpers.GetRequestBody(r)
		headers := r.Header

		if isGetMethod(r) && !isInstagramWebhookPath(r.URL.Path) && !isStaticPath(r.URL.Path) {
			slog.InfoContext(r.Context(), "page_view",
				"path", r.URL.Path,
				"user_agent", r.UserAgent(),
				"referrer", r.Referer(),
				"duration", time.Since(start),
				"ip", r.RemoteAddr,
				"headers", headers,
			)
		} else if !isStaticPath(r.URL.Path) {
			slog.InfoContext(r.Context(),
				"request",
				"method", r.Method,
				"path", r.URL.Path,
				"duration", time.Since(start),
				"ip", r.RemoteAddr,
				"body", body,
				"headers", headers,
			)
		}
	})
}
