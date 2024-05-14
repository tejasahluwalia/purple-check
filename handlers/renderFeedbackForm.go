package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"

	"github.com/tejasahluwalia/purple-check/models"
)

type FeedbackPageData struct {
	Receiver 		*models.Profile
	GiverExists 	bool
	Giver 			*models.Profile
}

func RenderFeedbackForm(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/layout.gohtml","templates/partials/header.gohtml", "templates/partials/search.gohtml", "templates/pages/feedback-form.gohtml")
	if err != nil {
		log.Println(err)
	}

	db, err := sql.Open("sqlite3", "db/purple-check.db")

	if err != nil {
		log.Println(err)
	}

	defer db.Close()
	
	var feedbackPageData FeedbackPageData

	receiver_id := r.URL.Query().Get("receiver_id")
	
	if receiver_id == "" {
		log.Println("Profile not found: ", err)
		http.NotFound(w, r)
		return
	} 
	
	stmt, err := db.Prepare("SELECT id, platform, platform_user_id, username, status, token FROM profiles WHERE id = ?")
	if err != nil {
		log.Println(err)
	}
	
	var profile models.Profile
	
	err = stmt.QueryRow(receiver_id).Scan(&profile.ID, &profile.Platform, &profile.PlatformUserID, &profile.Username, &profile.Status, &profile.Token)
	if err != nil {
		log.Println("Profile not found: ", err)
		http.NotFound(w, r)
		return
	}

	feedbackPageData.Receiver = &profile

	cookie, err := r.Cookie("platform_user_id")
	if err != nil {
		log.Println("Current user not found: ", err)
		feedbackPageData.GiverExists = false
		http.Redirect(w, r, "/connect-account?redirect_to_profile="+profile.Username, http.StatusSeeOther)
		return
	} else {
		stmt, err := db.Prepare("SELECT id, platform, platform_user_id, username, status, token FROM profiles WHERE platform = ? AND platform_user_id = ?")
		if err != nil {
			log.Println(err)
		}
		var currUser models.Profile
		err = stmt.QueryRow("instagram", cookie.Value).Scan(&currUser.ID, &currUser.Platform, &currUser.PlatformUserID, &currUser.Username, &currUser.Status, &currUser.Token)
		if err != nil {
			log.Println("Current user not found: ", err)
			feedbackPageData.GiverExists = false
			http.Redirect(w, r, "/connect-account?redirect_to_profile="+profile.Username, http.StatusSeeOther)
			return
		} else {
			feedbackPageData.GiverExists = true
			feedbackPageData.Giver = &currUser
		}
	}

	t.Execute(w, feedbackPageData)
}