package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"

	"github.com/tejasahluwalia/purple-check/models"
)

type StaticPageData struct {
	CurrUserExists bool
}

func checkForCurrentUser(r *http.Request) bool {
	cookie_platform_user_id, err := r.Cookie("platform_user_id")
	if err != nil {
		return false
	}

	cookie_access_token, err := r.Cookie("access_token")
	if err != nil {
		return false
	}

	db, err := sql.Open("sqlite3", "db/purple-check.db")
	if err != nil {
		log.Println(err)
	}

	defer db.Close()

	stmt, err := db.Prepare("SELECT id, platform, platform_user_id, username, status, token FROM profiles WHERE platform = 'instagram' AND platform_user_id = ? AND token = ?")
	if err != nil {
		log.Println(err)
	}

	var profile models.Profile

	err = stmt.QueryRow(cookie_platform_user_id.Value, cookie_access_token.Value).Scan(&profile.ID, &profile.Platform, &profile.PlatformUserID, &profile.Username, &profile.Status, &profile.Token)
	if err != nil {
		log.Println("Profile not found: ", err)
		return false
	}

	return true
}

func RenderHomepage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./templates/layout.gohtml", "./templates/pages/index.gohtml", "./templates/partials/search.gohtml", "./templates/partials/header.gohtml")
	if err != nil {
		log.Println(err)
	}
	t.Execute(w, StaticPageData {
		CurrUserExists: checkForCurrentUser(r),
	})
}

func RenderPrivacyPolicy(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./templates/layout.gohtml", "./templates/pages/privacy-policy.gohtml", "./templates/partials/search.gohtml", "./templates/partials/header.gohtml")
	if err != nil {
		log.Println(err)
	}
	t.Execute(w, StaticPageData {
		CurrUserExists: checkForCurrentUser(r),
	})
}

func RenderTermsOfService(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./templates/layout.gohtml", "./templates/pages/terms-of-service.gohtml", "./templates/partials/search.gohtml", "./templates/partials/header.gohtml")
	if err != nil {
		log.Println(err)
	}
	t.Execute(w, StaticPageData {
		CurrUserExists: checkForCurrentUser(r),
	})
}

func RenderAbout(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./templates/layout.gohtml", "./templates/pages/about.gohtml", "./templates/partials/search.gohtml", "./templates/partials/header.gohtml")
	if err != nil {
		log.Println(err)
	}
	t.Execute(w, StaticPageData {
		CurrUserExists: checkForCurrentUser(r),
	})
}

func RenderContact(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./templates/layout.gohtml", "./templates/pages/contact.gohtml", "./templates/partials/search.gohtml", "./templates/partials/header.gohtml")
	if err != nil {
		log.Println(err)
	}
	t.Execute(w, StaticPageData {
		CurrUserExists: checkForCurrentUser(r),
	})
}

func Render404(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./templates/layout.gohtml", "./templates/pages/404.gohtml", "./templates/partials/search.gohtml", "./templates/partials/header.gohtml")
	if err != nil {
		log.Println(err)
	}
	w.WriteHeader(http.StatusNotFound)
	t.Execute(w, StaticPageData {
		CurrUserExists: checkForCurrentUser(r),
	})
}

