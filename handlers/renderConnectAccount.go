package handlers

import (
	"html/template"
	"log"
	"net/http"
)

type ConnectAccountPageData struct {
	RedirectToProfile string 
	CurrUserExists    bool
}

func RenderConnectAccount(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./templates/layout.gohtml", "./templates/pages/connect-account.gohtml", "./templates/partials/header.gohtml", "./templates/partials/connect.gohtml")
	if err != nil {
		log.Println(err)
	}
	redirectToProfile := r.URL.Query().Get("redirect_to_profile")
	connectAccountPageData := ConnectAccountPageData{
		RedirectToProfile: redirectToProfile,
		CurrUserExists:    checkForCurrentUser(r),
	}
	t.Execute(w, connectAccountPageData)
}
