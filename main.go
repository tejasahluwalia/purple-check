package main

import (
	"context"
	"log"
	"net/http"

	"purple-check/internal/app"
	"purple-check/internal/components"
	"purple-check/internal/config"
	"purple-check/internal/database"
	"purple-check/internal/messaging"
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

	database.SyncDB()

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("GET /", (&page{components.Layout(app.Index())}).handler)
	mux.HandleFunc("GET /privacy-policy", (&page{components.Layout(app.PrivacyPolicy())}).handler)
	mux.HandleFunc("GET /terms-of-service", (&page{components.Layout(app.TermsOfService())}).handler)
	mux.HandleFunc("POST /search", app.Search)
	mux.HandleFunc("GET /profile/{username}", (&page{components.Layout(app.Profile())}).handler)

	messaging.InitConversations()
	messaging.SetPersistentMenu()

	mux.HandleFunc("GET /webhook/instagram", webhook.VerifyInstagramHook)
	mux.HandleFunc("POST /webhook/instagram", webhook.Instagram)
	mux.HandleFunc("GET /webhook/instagram/setup", webhook.SetupWebhooks)
	log.Println("Starting server")
	http.ListenAndServe(":"+config.PORT, mux)
}
