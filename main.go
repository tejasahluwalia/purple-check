package main

import (
	"context"
	"log"
	"net/http"

	"purple-check/internal/app"
	"purple-check/internal/components"
	"purple-check/internal/config"
	"purple-check/internal/db"
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

	db.SyncDB()

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("GET /", (&page{components.Layout(app.Index())}).handler)
	mux.HandleFunc("GET /privacy-policy", (&page{components.Layout(app.PrivacyPolicy())}).handler)
	mux.HandleFunc("GET /terms-of-service", (&page{components.Layout(app.TermsOfService())}).handler)
	mux.HandleFunc("GET /connect-account", (&page{components.Layout(app.ConnectAccount())}).handler)
	mux.HandleFunc("POST /search", app.Search)
	mux.HandleFunc("GET /profile/{username}", (&page{components.Layout(app.Profile())}).handler)
	mux.HandleFunc("GET /connect", app.Connect)
	mux.HandleFunc("GET /profile/{username}/add-feedback", (&page{components.Layout(app.AddFeedback())}).handler)
	mux.HandleFunc("PUT /feedback", app.PutFeedback)
	mux.HandleFunc("DELETE /feedback/{id}", app.DeleteFeedback)
	mux.HandleFunc("GET /disconnect-account", app.Disconnect)
	mux.HandleFunc("GET /delete-my-data", (&page{components.Layout(app.DeleteMyData())}).handler)
	mux.HandleFunc("DELETE /profile", app.DeleteAllUserFeedback)

	mux.HandleFunc("GET /webhook/instagram", webhook.VerifyInstagramHook)
	mux.HandleFunc("POST /webhook/instagram", webhook.Instagram)
	mux.HandleFunc("GET /webhook/instagram/setup", webhook.SetupWebhooks)
	log.Println("Starting server")
	http.ListenAndServe(":"+config.PORT, mux)
}
