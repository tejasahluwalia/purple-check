package messaging

import (
	"database/sql"
	"log"
	"strconv"

	"purple-check/internal/database"
)

func searchForUserAndRespond(usernameToSearch string, userId string) {
	db, closer := database.GetDB()
	defer closer()

	var rating sql.NullFloat64
	var totalRatings int

	err := db.QueryRow("SELECT COUNT(*) FROM feedback WHERE receiver = ? AND rating = 'POSITIVE'", usernameToSearch).Scan(&rating)
	if err != nil {
		log.Fatal("Error querying database.", err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM feedback WHERE receiver = ?", usernameToSearch).Scan(&totalRatings)
	if err != nil {
		log.Fatal("Error querying database.", err)
	}

	buttons := []ElementButton{
		{
			Type:  "web_url",
			Title: "See all feedback",
			URL:   "https://dev.purple-check.org/profile/" + usernameToSearch,
		},
		{
			Type:    "postback",
			Title:   "Leave feedback",
			Payload: "RATE:" + usernameToSearch,
		},
		{
			Type:    "postback",
			Title:   "Search for another user",
			Payload: "SEARCH",
		},
	}

	if totalRatings == 0 {
		sendButtonMessage(buttons, "No ratings found for @"+usernameToSearch, userId)
		return
	} else {
		positivePercentage := (rating.Float64 / float64(totalRatings)) * 100
		sendButtonMessage(buttons, "@"+usernameToSearch+"\nRating: "+strconv.FormatFloat(positivePercentage, 'f', 2, 32)+"% Positive\nTotal Ratings: "+strconv.Itoa(totalRatings)+"\n\nTo search for another user, enter the username.\n\nClick below to leave feedback for @"+usernameToSearch+".", userId)
		return
	}
}

func askForRating(usernameToRate string, userId string) {
	buttons := []ElementButton{
		{
			Type:    "postback",
			Title:   "Positive",
			Payload: "POSITIVE:" + usernameToRate,
		},
		{
			Type:    "postback",
			Title:   "Negative",
			Payload: "NEGATIVE:" + usernameToRate,
		},
		{
			Type:    "postback",
			Title:   "Cancel feedback submission",
			Payload: "CANCEL",
		},
	}

	sendButtonMessage(buttons, "How was your interaction with @"+usernameToRate+"?", userId)
}

func askForUsernameToSearch(userId string) {
	sendTextMessage("Please enter the username (with '@' symbol) of the page you want to check. (e.g. @purplecheck_org)", userId)
}

func invalidRatingMessage(userId string) {
	sendTextMessage("Invalid rating. Please select one of the options provided. Or click cancel.", userId)
}

func askForRole(userId string) {
	buttons := []ElementButton{
		{
			Type:    "postback",
			Title:   "Buyer",
			Payload: "ROLE:BUYER",
		},
		{
			Type:    "postback",
			Title:   "Seller",
			Payload: "ROLE:SELLER",
		},
		{
			Type:    "postback",
			Title:   "Cancel",
			Payload: "CANCEL",
		},
	}
	sendButtonMessage(buttons, "What was your role in this interaction?", userId)
}

func askForDealStage(userId string) {
	buttons := []ElementButton{
		{
			Type:    "postback",
			Title:   "Completed Deal",
			Payload: "DEAL_STAGE:COMPLETE",
		},
		{
			Type:    "postback",
			Title:   "Incomplete Deal",
			Payload: "DEAL_STAGE:INCOMPLETE",
		},
		{
			Type:    "postback",
			Title:   "Cancel",
			Payload: "CANCEL",
		},
	}
	sendButtonMessage(buttons, "What was the stage of the deal?", userId)
}
