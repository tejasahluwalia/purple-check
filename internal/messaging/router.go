package messaging

import (
	"strings"

	"purple-check/internal/helpers"
)

func RouteMessage(userId string, message string, payload string) {
	stage := getUserConversationStage(userId)

	switch stage {
	case "START":
		usernameToSearch, found := helpers.DetectUsername(message)
		if found {
			searchForUserAndRespond(usernameToSearch, userId)
			return
		}

		if payload != "" {
			switch strings.Split(payload, ":")[0] {
			case "RATE":
				usernameToRate := strings.Split(payload, ":")[1]
				askForRating(usernameToRate, userId)
				setUserConversationStage(userId, "AWAITING_RATING")
				return
			case "SEARCH":
				askForUsernameToSearch(userId)
				return
			}
		}

		askForUsernameToSearch(userId)
		return

	case "AWAITING_RATING":
		if payload == "CANCEL" {
			setUserConversationStage(userId, "START")
			sendTextMessage("Rating cancelled.", userId)
			askForUsernameToSearch(userId)
			return
		} else if len(strings.Split(payload, ":")) == 2 {
			usernameToRate := strings.Split(payload, ":")[1]
			rating := strings.Split(payload, ":")[0]
			if rating == "POSITIVE" || rating == "NEUTRAL" || rating == "NEGATIVE" {
				setUserConversationStage(userId, "START")
				saveRating(rating, userId, usernameToRate)
				sendTextMessage("Thank you for submitting a rating.", userId)
				askForUsernameToSearch(userId)
				return
			} else {
				invalidRatingMessage(userId)
				return
			}
		}
		if strings.ToLower(message) == "cancel" {
			setUserConversationStage(userId, "START")
			sendTextMessage("Rating cancelled.", userId)
			askForUsernameToSearch(userId)
			return
		}
		invalidRatingMessage(userId)
		return
	}
}
