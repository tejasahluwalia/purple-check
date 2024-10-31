package messaging

import (
	"database/sql"
	"log"
	"strconv"
	"strings"

	"purple-check/internal/database"
)

func detectUsername(message string) (string, bool) {
	temp := strings.Split(message, "@")
	if len(temp) == 1 {
		return "", false
	} else {
		username, _ := strings.CutSuffix(temp[1], " ")
		return "@" + username, true
	}
}

func searchForUserAndRespond(usernameToSearch string, userId string) {
	sendTextMessage("Searching for "+usernameToSearch, userId)

	db, closer := database.GetDB()
	defer closer()

	var averageFeedbackRating sql.NullFloat64

	stmt, err := db.Prepare("SELECT AVG(rating) FROM feedback WHERE receiver = ? ORDER BY feedback.created_at DESC")
	if err != nil {
		log.Fatal(err)
	}

	err = stmt.QueryRow(strings.Split(usernameToSearch, "@")[1]).Scan(&averageFeedbackRating)
	if err != nil {
		log.Println(err)
	}

	sendTextMessage("Average rating for "+usernameToSearch+" is: "+strconv.FormatFloat(averageFeedbackRating.Float64, 'f', 2, 32), userId)

	buttons := []ElementButton{
		{
			Type:    "postback",
			Title:   "Leave feedback",
			Payload: "1:" + usernameToSearch,
		},
		{
			Type:    "postback",
			Title:   "Search for another user",
			Payload: "2",
		},
	}

	sendButtonMessage(buttons, "Would you like to leave feedback or search for another user?", userId)
}

func askForRating(username string, userId string) {
	buttons := []ElementButton{
		{
			Type:    "postback",
			Title:   "Positive",
			Payload: "1:" + username,
		},
		{
			Type:    "postback",
			Title:   "Neutral",
			Payload: "2:" + username,
		},
		{
			Type:    "postback",
			Title:   "Negative",
			Payload: "3:" + username,
		},
	}

	sendButtonMessage(buttons, "How was your interaction?", userId)
}
