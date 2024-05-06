package main

import (
	"net/http"

	"github.com/tejasahluwalia/purple-check/handlers"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handlers.RenderHomepage)
	mux.HandleFunc("GET /connect-account", handlers.RenderConnectAccount)
	mux.HandleFunc("POST /search", handlers.HandleSearch)
	mux.HandleFunc("GET /profile/{username}", handlers.RenderProfile)
	mux.HandleFunc("GET /connect", handlers.HandleConnect)
	mux.HandleFunc("GET /refresh-token", handlers.RefreshAccessToken)
	mux.HandleFunc("GET /feedback", handlers.RenderFeedbackForm)
	mux.HandleFunc("POST /submit-feedback", handlers.HandleSubmitFeedback)
	fs := http.FileServer(http.Dir("static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))
	http.ListenAndServe(":8080", mux)
}
