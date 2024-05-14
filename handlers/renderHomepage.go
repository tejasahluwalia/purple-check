package handlers

import (
	"html/template"
	"log"
	"net/http"
)

func RenderHomepage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./templates/layout.gohtml", "./templates/pages/index.gohtml", "./templates/partials/search.gohtml", "./templates/partials/header.gohtml")
	if err != nil {
		log.Println(err)
	}
	t.Execute(w, nil)
}
