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
		log.Fatal(err)
	}

	defer db.Close()

	cookie, err := r.Cookie("platform_user_id")
	if err != nil {
		log.Println("Current user not found: ", err)
		http.Redirect(w, r, "/404", http.StatusNotFound)
	} else {
		stmt, err := db.Prepare("SELECT id, platform, platform_user_id, username, status, token FROM profiles WHERE platform = ? AND platform_user_id = ?")
		if err != nil {
			log.Fatal(err)
		}
		var currUser models.Profile
		err = stmt.QueryRow("instagram", cookie.Value).Scan(&currUser.ID, &currUser.Platform, &currUser.PlatformUserID, &currUser.Username, &currUser.Status, &currUser.Token)
		if err != nil {
			log.Println("Current user not found: ", err)
			http.Redirect(w, r, "/404", http.StatusNotFound)
		} else {
			receiverID := r.FormValue("receiver_id")
			stmt, err := db.Prepare("SELECT id, platform, platform_user_id, username, status, token FROM profiles WHERE id = ?")
			if err != nil {
				log.Fatal(err)
			}
			var receiver models.Profile
			err = stmt.QueryRow(receiverID).Scan(&receiver.ID, &receiver.Platform, &receiver.PlatformUserID, &receiver.Username, &receiver.Status, &receiver.Token)
			if err != nil {
				log.Println("Receiver profile not found: ", err)
				http.Redirect(w, r, "/404", http.StatusNotFound)
			}
			comment := r.FormValue("comment")
			giverRole := r.FormValue("giver_role")
			receiverRole := r.FormValue("receiver_role")
			stmt, err = db.Prepare("INSERT INTO feedback(giver_id, receiver_id, comment, giver_role, receiver_role) VALUES(?, ?, ?, ?, ?)")
			if err != nil {
				log.Fatal(err)
			}
			_, err = stmt.Exec(currUser.ID, receiver.ID, comment, giverRole, receiverRole)
			if err != nil {
				log.Fatal(err)
			}
			http.Redirect(w, r, "/profile/"+receiver.Username, http.StatusFound)
		}
	}
}