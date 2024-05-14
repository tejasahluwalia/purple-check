package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/tejasahluwalia/purple-check/models"
)

type ProfilePageData struct {
	CurrUser  		*models.Profile
	FeedbackList 	*[]models.Feedback
	Profile      	*models.Profile
	ProfileExists 	bool
	CurrUserExists 	bool
	Filter 			string
}

func stars(rating int) string {
	return strings.Repeat("★", rating) + strings.Repeat("☆", 5-rating)
}

func RenderProfile(w http.ResponseWriter, r *http.Request) {
	funcMap := template.FuncMap{
		"stars": stars,
	}
	t, err := template.New("layout.gohtml").Funcs(funcMap).ParseFiles("templates/layout.gohtml", "templates/partials/header.gohtml", "templates/partials/connect.gohtml", "templates/pages/profile.gohtml", "templates/partials/search.gohtml", "templates/partials/feedbackList.gohtml")
	if err != nil {
		log.Println(err)
	}
	username := r.PathValue("username")
	username = strings.ToLower(username)
	db, err := sql.Open("sqlite3", "db/purple-check.db")

    if err != nil {
        log.Println(err)
    }

    defer db.Close()

	stmt, err := db.Prepare("SELECT id, platform, platform_user_id, username, status, token FROM profiles WHERE platform = ? AND username = ?")
	if err != nil {
		log.Println(err)
	}

	var profile models.Profile
	var profileExists bool
	var feedbackList []models.Feedback
	
	err = stmt.QueryRow("instagram", username).Scan(&profile.ID, &profile.Platform, &profile.PlatformUserID, &profile.Username, &profile.Status, &profile.Token)
	if err != nil {
		stmt, err = db.Prepare("INSERT INTO profiles(platform, platform_user_id, username, status, token) VALUES(?, ?, ?, ?, ?)")
		if err != nil {
			log.Println(err)
		}
		_, err = stmt.Exec("instagram", nil, username, "not-connected", nil)
		if err != nil {
			log.Println(err)
		}
		stmt, err := db.Prepare("SELECT id, platform, platform_user_id, username, status, token FROM profiles WHERE platform = ? AND username = ?")
		if err != nil {
			log.Println(err)
		}
		err = stmt.QueryRow("instagram", username).Scan(&profile.ID, &profile.Platform, &profile.PlatformUserID, &profile.Username, &profile.Status, &profile.Token)
		if err != nil {
			log.Println(err)
		}
	} else {
		profileExists = true
	}

	stmt, err = db.Prepare("SELECT feedback.id, giver.id, giver.username, receiver.id, receiver.username, feedback.rating, feedback.comment, feedback.created_at FROM feedback JOIN profiles AS giver ON feedback.giver_id = giver.id JOIN profiles AS receiver ON feedback.receiver_id = receiver.id WHERE receiver_id = ? ORDER BY feedback.created_at DESC")

	if err != nil {
		log.Println(err)
	}

	rows, err := stmt.Query(profile.ID)
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		var feedback models.Feedback
		err = rows.Scan(&feedback.ID, &feedback.Giver.ID, &feedback.Giver.Username, &feedback.Receiver.ID, &feedback.Receiver.Username, &feedback.Rating, &feedback.Comment, &feedback.CreatedAt)
		if err != nil {
			log.Println(err)
		}
		feedbackList = append(feedbackList, feedback)
	}

	cookie_platform_user_id, err := r.Cookie("platform_user_id")
	var currUser models.Profile
	var currUserExists bool
	if err != nil {
		currUserExists = false
		var profilePageData ProfilePageData
		if profile.Status == "not-connected" {
			profilePageData.ProfileExists = false
		} else {
			profilePageData.ProfileExists = true
		}
		profilePageData.Profile = &profile
		profilePageData.CurrUserExists = currUserExists
		profilePageData.FeedbackList = &feedbackList
		t.Execute(w, profilePageData)
		return
	} 
	stmt, err = db.Prepare("SELECT id, platform, platform_user_id, username, status, token FROM profiles WHERE platform = ? AND platform_user_id = ?")
	if err != nil {
		log.Println(err)
	}
	err = stmt.QueryRow("instagram", cookie_platform_user_id.Value).Scan(&currUser.ID, &currUser.Platform, &currUser.PlatformUserID, &currUser.Username, &currUser.Status, &currUser.Token)
	if err != nil {
		currUserExists = false
	} else {
		currUserExists = true
	}

	var profilePageData ProfilePageData
	profilePageData.CurrUser = &currUser
	profilePageData.CurrUserExists = currUserExists
	profilePageData.Profile = &profile
	profilePageData.ProfileExists = profileExists
	profilePageData.FeedbackList = &feedbackList

	t.Execute(w, profilePageData)
}