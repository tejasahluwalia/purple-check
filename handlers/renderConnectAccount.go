package handlers

import (
	"html/template"
	"log"
	"net/http"
)

func RenderConnectAccount(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./templates/layout.gohtml", "./templates/pages/connect-account.gohtml", "./templates/partials/search.gohtml", "./templates/partials/connect.gohtml")
	if err != nil {
		log.Fatal(err)
	}
	t.Execute(w, nil)
}
