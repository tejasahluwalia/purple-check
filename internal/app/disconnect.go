package app

import (
	"database/sql"
	"log"
	"net/http"

	"purple-check/internal/config"

	_ "modernc.org/sqlite"
)

func Disconnect(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite", config.LOCAL_DB_PATH)
	if err != nil {
		log.Println(err)
	}

	defer db.Close()
	
	cookie_platform_user_id, err := r.Cookie("platform_user_id")
	if err != nil {
		log.Println("Current user not found: ", err)
		http.Error(w, "Current user not found", http.StatusNotFound)
		return
	}

	stmt, err := db.Prepare("UPDATE profiles SET status = 'not-connected', token = NULL, expires_in = NULL WHERE platform = ? AND platform_user_id = ?")
	if err != nil {
		log.Println(err)
	}

	_, err = stmt.Exec("instagram", cookie_platform_user_id.Value)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to disconnect", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{Name: "platform_user_id", Value: "", HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode, MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: "access_token", Value: "", HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode, MaxAge: -1})
	http.Redirect(w, r, "/", http.StatusFound)
}