package handlers

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"

	"github.com/tejasahluwalia/purple-check/models"
)

func HandleSubmitFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusMethodNotAllowed)
		return
	}

	db, err := sql.Open("sqlite3", "db/purple-check.db")

	if err != nil {
		log.Println(err)
	}

	defer db.Close()

	cookie, err := r.Cookie("platform_user_id")
	if err != nil {
		log.Println("Current user not found: ", err)
		http.Error(w, "Current user not found", http.StatusNotFound)
		return
	} else {
		stmt, err := db.Prepare("SELECT id, platform, platform_user_id, username, status, token FROM profiles WHERE platform = ? AND platform_user_id = ?")
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		var currUser models.Profile
		err = stmt.QueryRow("instagram", cookie.Value).Scan(&currUser.ID, &currUser.Platform, &currUser.PlatformUserID, &currUser.Username, &currUser.Status, &currUser.Token)
		if err != nil {
			log.Println("Current user not found: ", err)
			http.Error(w, "Current user not found", http.StatusNotFound)
			return
		} else {
			receiverID := r.FormValue("receiver_id")
			
			if receiverID == currUser.ID {
				log.Println("Cannot submit feedback to self")
				http.Error(w, "Cannot submit feedback to self", http.StatusBadRequest)
				return
			}

			stmt, err := db.Prepare("SELECT id, platform, platform_user_id, username, status, token FROM profiles WHERE id = ?")
			if err != nil {
				log.Println(err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			var receiver models.Profile
			err = stmt.QueryRow(receiverID).Scan(&receiver.ID, &receiver.Platform, &receiver.PlatformUserID, &receiver.Username, &receiver.Status, &receiver.Token)
			if err != nil {
				log.Println("Receiver profile not found: ", err)
				http.Error(w, "Receiver profile not found", http.StatusNotFound)
				return
			}
			comment := r.FormValue("comment")
			rating := r.FormValue("rating")

			if comment == "" || rating == "" {
				log.Println("Comment or rating is empty")
				http.Error(w, "Comment or rating is empty", http.StatusBadRequest)
				return
			}

			stmt, err = db.Prepare("INSERT INTO feedback(giver_id, receiver_id, rating, comment) VALUES(?, ?, ?, ?)")
			if err != nil {
				log.Println(err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			_, err = stmt.Exec(currUser.ID, receiver.ID, rating, comment)
			if err != nil {
				log.Println(err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/profile/"+receiver.Username, http.StatusFound)
			return
		}
	}
}