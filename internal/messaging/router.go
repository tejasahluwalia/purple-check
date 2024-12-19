package messaging

import (
	"log"
	"strconv"
	"strings"
)

func RouteMessage(userId string, message string, payload string) {
	stage := GetUserConversationStage(userId)

	switch stage {
	case "START":
		if payload != "" {
			switch payload {
			case "SEARCH":
				sendTextMessage("Please send a username to search.", userId)
				SetUserConversationStage(userId, "SEARCH")
			case "RATE":
				sendTextMessage("Whom would you like to rate?", userId)
				SetUserConversationStage(userId, "RATE")
			}
		}

		sendTextMessage("Welcome to Purple Check. Select an option below.", userId)
		return

	case "SEARCH":
		usernameToSearch, found := detectUsername(message)
		if found {
			searchForUserAndRespond(usernameToSearch, userId)
			return
		} else {
			sendTextMessage("Please send a valid username.", userId)
			return
		}

	case "RATE":
		usernameToRate, found := detectUsername(message)
		if found {
			askForRating(usernameToRate, userId)
			SetUserConversationStage(userId, "AWAITING_RATING")
			return
		} else {
			sendTextMessage("Please send a valid username.", userId)
			return
		}

	case "AWAITING_RATING":
		if payload != "" {
			username := strings.Split(payload, ":")[1]
			rating, err := strconv.Atoi(strings.Split(payload, ":")[0])
			if err != nil {
				log.Fatal("Error parsing rating.")
			}
			SetUserConversationStage(userId, "START")
			saveRating(rating, userId, username)
			sendTextMessage("Thank you for your rating.", userId)
			return
		} else {
			sendTextMessage("Please rate the user.", userId)
			return
		}

	default:
		SetUserConversationStage(userId, "START")
		sendTextMessage("Restart", userId)
		return
	}
}
