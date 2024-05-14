package handlers

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"github.com/tejasahluwalia/purple-check/models"
)

func HandleDeleteFeedback(w http.ResponseWriter, r *http.Request) { 
	if r.Method != "DELETE" {
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

	stmt, err = db.Prepare("SELECT id, giver_id FROM feedback WHERE id = ?")
	if err != nil {
		log.Println(err)
	}

	id := r.PathValue("id")
	var feedback_id, giver_id string
	err = stmt.QueryRow(id).Scan(&feedback_id, &giver_id)
	if err != nil {
		log.Println(err)
		http.Error(w, "Feedback not found", http.StatusNotFound)
		return
	}

	if currUser.ID != giver_id {
		log.Println("Unauthorized: Current user is not the giver of the feedback")
		http.Error(w, "Unauthorized: Current user is not the giver of the feedback", http.StatusUnauthorized)
		return
	}

	stmt, err = db.Prepare("DELETE FROM feedback WHERE id = ?")
	if err != nil {
		log.Println(err)
	}

	_, err = stmt.Exec(id)
	if err != nil {
		log.Println(err)
	}

	w.WriteHeader(http.StatusOK)
}