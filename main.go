package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"

	"purple-check/internal/logger"
	"purple-check/internal/middleware"

	"purple-check/internal/app"
	"purple-check/internal/components"
	"purple-check/internal/config"
	"purple-check/internal/webhook"

	"github.com/a-h/templ"
)

type page struct {
	templ.Component
}

func (t page) handler(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), components.RequestContextKey, r)
	t.Render(ctx, w)
}

func main() {
	logger.Setup()
	// messaging.InitConversations()
	// messaging.SetPersistentMenu()

	mux := http.NewServeMux()
	handler := middleware.RedirectNonWWW(mux)

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), components.RequestContextKey, r)
		components.Layout(app.Index()).Render(ctx, w)
	})
	mux.HandleFunc("GET /privacy-policy", (page{components.Layout(app.PrivacyPolicy())}).handler)
	mux.HandleFunc("GET /delete-my-data", (page{components.Layout(app.DeleteMyData())}).handler)
	mux.HandleFunc("GET /terms-of-service", (page{components.Layout(app.TermsOfService())}).handler)
	mux.HandleFunc("POST /search", app.Search)
	mux.HandleFunc("GET /profile/{username}", (page{components.Layout(app.Profile())}).handler)

	mux.HandleFunc("GET /webhook/instagram", webhook.VerifyInstagramHook)
	mux.HandleFunc("POST /webhook/instagram", webhook.Instagram)
	mux.HandleFunc("GET /webhook/instagram/setup", webhook.SetupWebhooks)
	slog.Info("starting server", "port", config.PORT)
	err := http.ListenAndServe(":"+config.PORT, handler)
	log.Fatal(err)
}
