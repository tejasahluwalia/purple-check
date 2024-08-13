package main

import (
	"context"
	"net/http"

	"purple-check/internal/app"
	"purple-check/internal/components"

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
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("GET /", (&page{app.Layout(app.Homepage())}).handler)
	mux.HandleFunc("GET /privacy-policy", (&page{app.Layout(app.PrivacyPolicy())}).handler)
	mux.HandleFunc("GET /terms-of-service", (&page{app.Layout(app.TermsOfService())}).handler)
	mux.HandleFunc("GET /connect-account", (&page{app.Layout(app.ConnectAccount())}).handler)
	mux.HandleFunc("POST /search", app.Search)
	mux.HandleFunc("GET /profile/{username}", (&page{app.Layout(app.Profile())}).handler)
	mux.HandleFunc("GET /connect", app.Connect)
	mux.HandleFunc("GET /profile/{username}/add-feedback", (&page{app.Layout(app.AddFeedback())}).handler)
	mux.HandleFunc("PUT /feedback", app.PutFeedback)
	mux.HandleFunc("DELETE /feedback/{id}", app.DeleteFeedback)
	mux.HandleFunc("GET /disconnect-account", app.Disconnect)
	mux.HandleFunc("GET /delete-my-data", (&page{app.Layout(app.DeleteMyData())}).handler)
	mux.HandleFunc("DELETE /profile", app.DeleteAllUserFeedback)

	http.ListenAndServe(":9980", mux)
}

