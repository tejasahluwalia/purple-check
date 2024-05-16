package handlers

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"github.com/tejasahluwalia/purple-check/models"
)

func HandleDisconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db, err := sql.Open("sqlite3", "db/purple-check.db")
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

	stmt, err := db.Prepare("SELECT id, platform_user_id, platform FROM profiles WHERE platform = ? AND platform_user_id = ?")
	if err != nil {
		log.Println(err)
	}

	var currUser models.Profile
	err = stmt.QueryRow("instagram", cookie_platform_user_id.Value).Scan(&currUser.ID, &currUser.PlatformUserID, &currUser.Platform)
	if err != nil {
		log.Println(err)
		http.Error(w, "Current user not found", http.StatusNotFound)
		return
	}

	stmt, err = db.Prepare("UPDATE profiles SET status = ?, token = NULL, expires_in = NULL WHERE id = ?")
	if err != nil {
		log.Println(err)
	}

	_, err = stmt.Exec("not-connected", currUser.ID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to disconnect", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)	
}