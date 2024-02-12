package handlers

import (
	"html/template"
	"log"
	"net/http"
)

func RenderHomepage(w http.ResponseWriter, r *http.Request) {
	log.Println("Rendering homepage")
	t, err := template.ParseFiles("./templates/layout.gohtml", "./templates/index.gohtml", "./templates/search.gohtml")
	if err != nil {
		log.Fatal(err)
	}
	t.Execute(w, nil)
}
