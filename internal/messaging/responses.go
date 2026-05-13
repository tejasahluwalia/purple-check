package messaging

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"

	"purple-check/internal/config"
	"purple-check/internal/database"
)

func searchForUserAndRespond(ctx context.Context, usernameToSearch string, userId string) error {
	if _, err := database.PullDB(ctx); err != nil {
		log.Println("Error pulling database changes.", err)
	}

	db, closer, err := database.GetDB(ctx)
	if err != nil {
		return err
	}
	defer closer()

	var positiveRatings int
	var totalRatings int

	err = db.QueryRow("SELECT COUNT(*) FROM feedback WHERE receiver = ? AND rating = 'POSITIVE'", usernameToSearch).Scan(&positiveRatings)
	if err != nil {
		return fmt.Errorf("query positive ratings: %w", err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM feedback WHERE receiver = ?", usernameToSearch).Scan(&totalRatings)
	if err != nil {
		return fmt.Errorf("query total ratings: %w", err)
	}

	buttons := []ElementButton{
		{
			Type:  "web_url",
			Title: "See all reviews",
			URL:   "https://" + config.HOST + "/profile/" + url.PathEscape(usernameToSearch),
		},
		{
			Type:    "postback",
			Title:   "Leave review",
			Payload: "RATE:" + usernameToSearch,
		},
		{
			Type:    "postback",
			Title:   "Search for another user",
			Payload: "SEARCH",
		},
	}

	if totalRatings == 0 {
		return sendButtonMessage(buttons, "No ratings found for @"+usernameToSearch, userId)
	} else {
		positivePercentage := (float64(positiveRatings) / float64(totalRatings)) * 100
		ratingPlural := "ratings"
		if totalRatings == 1 {
			ratingPlural = "rating"
		}
		return sendButtonMessage(buttons, "@"+usernameToSearch+"\n\n"+strconv.FormatFloat(positivePercentage, 'f', 0, 32)+"% positive ("+strconv.Itoa(totalRatings)+" "+ratingPlural+")", userId)
	}
}

func askForRating(usernameToRate string, userId string) error {
	buttons := []ElementButton{
		{
			Type:    "postback",
			Title:   "Positive",
			Payload: "RATING:POSITIVE:" + usernameToRate,
		},
		{
			Type:    "postback",
			Title:   "Negative",
			Payload: "RATING:NEGATIVE:" + usernameToRate,
		},
		{
			Type:    "postback",
			Title:   "Cancel",
			Payload: "CANCEL",
		},
	}

	return sendButtonMessage(buttons, "How was your interaction with @"+usernameToRate+"?", userId)
}

func askForUsernameToSearch(userId string) error {
	return sendTextMessage("Please enter the username (with '@' symbol) of the page you want to check. (e.g. @purplecheck_org)", userId)
}

func invalidResponseMessage(userId string) error {
	return sendTextMessage("Invalid response. Please select one of the options provided. Or click cancel.", userId)
}

func askForRole(userId string) error {
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
	return sendButtonMessage(buttons, "What was your role in this interaction?", userId)
}

func askForDealStage(userId string) error {
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
	return sendButtonMessage(buttons, "What was the stage of the deal?", userId)
}
