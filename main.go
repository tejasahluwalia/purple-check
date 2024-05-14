package main

import (
	"net/http"
	"os/exec"

	"github.com/tejasahluwalia/purple-check/handlers"
)

func main() {
	cmd := exec.Command("/bin/sh", "refresh.sh")

	mux := http.NewServeMux()
	
	mux.HandleFunc("GET /", handlers.RenderHomepage)
	mux.HandleFunc("GET /connect-account", handlers.RenderConnectAccount)
	mux.HandleFunc("POST /search", handlers.HandleSearch)
	mux.HandleFunc("GET /profile/{username}", handlers.RenderProfile)
	mux.HandleFunc("GET /connect", handlers.HandleConnect)
	mux.HandleFunc("GET /refresh-token", handlers.RefreshAccessToken)
	mux.HandleFunc("GET /feedback", handlers.RenderFeedbackForm)
	mux.HandleFunc("DELETE /feedback/{id}", handlers.HandleDeleteFeedback)
	mux.HandleFunc("POST /submit-feedback", handlers.HandleSubmitFeedback)
	fs := http.FileServer(http.Dir("static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

	go cmd.Run()
	http.ListenAndServe(":9990", mux)
}
