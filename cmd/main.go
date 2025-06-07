package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"purple-check/internal/logger"
	"purple-check/internal/middleware"

	"purple-check/internal/app"
	"purple-check/internal/components"
	"purple-check/internal/config"
	"purple-check/internal/instagram"
	"purple-check/internal/messaging"
	"purple-check/internal/webhook"

	"github.com/a-h/templ"
)

func disableCacheInDevMode(next http.Handler) http.Handler {
	if !config.DEV {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

type page struct {
	templ.Component
}

func (t page) handler(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), components.RequestContextKey, r)
	t.Render(ctx, w)
}

func main() {
	logger.Setup()
	messaging.InitConversations()
	messaging.SetPersistentMenu()

	mux := http.NewServeMux()
	handler := middleware.RedirectNonWWW(mux)

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("GET /static/", disableCacheInDevMode(http.StripPrefix("/static/", fs)))

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), components.RequestContextKey, r)
		components.Layout(app.Index(), components.Head{
			Title:       "Purple Check",
			Description: "Purple Check is a review platform for buyers and sellers on Instagram.",
			URL:         "https://www.purple-check.org",
		}).Render(ctx, w)
	})
	mux.HandleFunc("GET /privacy-policy", (page{components.Layout(app.PrivacyPolicy(), components.Head{
		Title: "Purple Check - Privacy Policy",
		URL:   "https://www.purple-check.org/privacy-policy",
	})}).handler)
	mux.HandleFunc("GET /delete-my-data", (page{components.Layout(app.DeleteMyData(), components.Head{
		Title:       "Delete My Data",
		Description: "Delete your reviews from Purple Check.",
		URL:         "https://www.purple-check.org/delete-my-data",
	})}).handler)
	mux.HandleFunc("GET /terms-of-service", (page{components.Layout(app.TermsOfService(), components.Head{
		Title: "Purple Check - Terms of Service",
		URL:   "https://www.purple-check.org/terms-of-service",
	})}).handler)
	mux.HandleFunc("POST /search", app.Search)
	mux.HandleFunc("GET /profile/{username}", func(w http.ResponseWriter, r *http.Request) {
		username := r.PathValue("username")
		if username == "" {
			http.Error(w, "Invalid username", http.StatusBadRequest)
		}
		ctx := context.WithValue(r.Context(), components.RequestContextKey, r)
		components.Layout(app.Profile(), components.Head{
			Title:       fmt.Sprintf("Reviews for @%s on Instagram", username),
			Description: fmt.Sprintf("Read and write reviews for @%s on Instagram. See recent feedback left by buyers and sellers.", username),
			URL:         fmt.Sprintf("https://www.purple-check.org/profile/%s", username),
		}).Render(ctx, w)
	})

	mux.HandleFunc("GET /webhook/instagram", webhook.VerifyInstagramHook)
	mux.HandleFunc("POST /webhook/instagram", webhook.Instagram)
	mux.HandleFunc("GET /webhook/instagram/setup", webhook.SetupWebhooks)
	mux.HandleFunc("GET /instagram/refresh-access-token", instagram.RefreshAccessToken)
	slog.Info("starting server", "port", config.PORT)
	err := http.ListenAndServe(":"+config.PORT, handler)
	log.Fatal(err)
}
