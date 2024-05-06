package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/tejasahluwalia/purple-check/models"
)

type ProfilePageData struct {
	CurrUser  		*models.Profile
	FeedbackList 	*[]models.Feedback
	Profile      	*models.Profile
	ProfileExists 	bool
	CurrUserExists 	bool
}

func RenderProfile(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/layout.gohtml", "templates/partials/connect.gohtml", "templates/pages/profile.gohtml", "templates/partials/search.gohtml", "templates/partials/feedbackList.gohtml")
	if err != nil {
		log.Fatal(err)
	}
	username := r.PathValue("username")
	username = strings.ToLower(username)
	db, err := sql.Open("sqlite3", "db/purple-check.db")

    if err != nil {
        log.Fatal(err)
    }

    defer db.Close()

	stmt, err := db.Prepare("SELECT id, platform, platform_user_id, username, status, token FROM profiles WHERE platform = ? AND username = ?")
	if err != nil {
		log.Fatal(err)
	}

	var profile models.Profile
	var feedbackList []models.Feedback
	
	err = stmt.QueryRow("instagram", username).Scan(&profile.ID, &profile.Platform, &profile.PlatformUserID, &profile.Username, &profile.Status, &profile.Token)
	if err != nil {
		stmt, err = db.Prepare("INSERT INTO profiles(platform, platform_user_id, username, status, token) VALUES(?, ?, ?, ?, ?)")
		if err != nil {
			log.Fatal(err)
		}
		_, err = stmt.Exec("instagram", nil, username, "not-connected", nil)
		if err != nil {
			log.Fatal(err)
		}
		stmt, err := db.Prepare("SELECT id, platform, platform_user_id, username, status, token FROM profiles WHERE platform = ? AND username = ?")
		if err != nil {
			log.Fatal(err)
		}
		err = stmt.QueryRow("instagram", username).Scan(&profile.ID, &profile.Platform, &profile.PlatformUserID, &profile.Username, &profile.Status, &profile.Token)
		if err != nil {
			log.Fatal(err)
		}
	}

	stmt, err = db.Prepare("SELECT id, giver_id, receiver_id, rating, comment, created_at FROM feedback WHERE receiver_id = ?")

	if err != nil {
		log.Fatal(err)
	}
	rows, err := stmt.Query(profile.ID)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var feedback models.Feedback
		err = rows.Scan(&feedback.ID, &feedback.GiverID, &feedback.ReceiverID, &feedback.Rating, &feedback.Comment, &feedback.CreatedAt)
		if err != nil {
			log.Fatal(err)
		}
		createdAt, err := time.Parse(time.RFC3339, feedback.CreatedAt)
		if err != nil {
			log.Fatal(err)
		}
		
		since := time.Since(createdAt)
		if since.Hours() < 24 {
			feedback.CreatedAt = "Today at " + createdAt.Format("3:15 PM")
		} else if since.Hours() < 48 {
			feedback.CreatedAt = "Yesterday at " + createdAt.Format("3:15 PM")
		} else {
			feedback.CreatedAt = createdAt.Format("Jan 2, 2006 at 3:04 PM")
		}

		feedbackList = append(feedbackList, feedback)
	}


	cookie, err := r.Cookie("platform_user_id")
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
		log.Fatal(err)
	}
	err = stmt.QueryRow("instagram", cookie.Value).Scan(&currUser.ID, &currUser.Platform, &currUser.PlatformUserID, &currUser.Username, &currUser.Status, &currUser.Token)
	if err != nil {
		currUserExists = false
	} else {
		currUserExists = true
	}

	var profilePageData ProfilePageData
	if profile.Status == "not-connected" {
		profilePageData.ProfileExists = false
	} else {
		profilePageData.ProfileExists = true
	}	
	profilePageData.Profile = &profile
	profilePageData.CurrUserExists = currUserExists
	profilePageData.CurrUser = &currUser
	profilePageData.FeedbackList = &feedbackList

	t.Execute(w, profilePageData)
}