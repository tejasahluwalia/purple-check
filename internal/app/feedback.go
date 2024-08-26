package app

import (
	"database/sql"
	"log"
	"net/http"

	"purple-check/internal/config"
	"purple-check/internal/helpers"

	_ "modernc.org/sqlite"
)

func PutFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		log.Println("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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
	cookie_access_token, err := r.Cookie("access_token")
	if err != nil {
		log.Println("Current user not found: ", err)
		http.Error(w, "Current user not found", http.StatusNotFound)
		return
	}

	stmt, err := db.Prepare("SELECT id FROM profiles WHERE platform = ? AND platform_user_id = ? AND token = ?")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	var curr_user_id string
	err = stmt.QueryRow("instagram", cookie_platform_user_id.Value, cookie_access_token.Value).Scan(&curr_user_id)
	if err != nil {
		log.Println("Current user not found: ", err)
		http.Error(w, "Current user not found", http.StatusNotFound)
		return
	} 

	receiver_id := r.FormValue("receiver_id")
	
	if receiver_id == curr_user_id {
		log.Println("Cannot submit feedback to self")
		http.Error(w, "Cannot submit feedback to self", http.StatusBadRequest)
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
	_, err = stmt.Exec(curr_user_id, receiver_id, rating, comment)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	receiver_username := r.FormValue("receiver_username")
	w.Header().Set("HX-Location", "/profile/"+receiver_username)
	w.WriteHeader(http.StatusCreated)
}

func DeleteFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db, err := sql.Open("sqlite", config.LOCAL_DB_PATH)
	if err != nil {
		log.Println(err)
	}

	defer db.Close()

	curr_user := helpers.GetCurrUser(r, db)
	if curr_user == nil {
		log.Println("Current user not found")
		http.Error(w, "Current user not found", http.StatusNotFound)
		return
	}

	stmt, err := db.Prepare("SELECT feedback.id, feedback.giver_id, feedback.receiver_id, receiver.username FROM feedback JOIN profiles AS receiver ON feedback.receiver_id = receiver.id WHERE feedback.id = ?")
	if err != nil {
		log.Println(err)
	}

	id := r.PathValue("id")
	var feedback_id, giver_id, receiver_id, receiver_username string
	err = stmt.QueryRow(id).Scan(&feedback_id, &giver_id, &receiver_id, &receiver_username)
	if err != nil {
		log.Println(err)
		http.Error(w, "Feedback not found", http.StatusNotFound)
		return
	}

	if curr_user.ID != giver_id {
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

	w.Header().Set("HX-Location", "/profile/"+receiver_username)
	w.WriteHeader(http.StatusOK)
}