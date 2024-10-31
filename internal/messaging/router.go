package messaging

import (
	"log"
	"strconv"
	"strings"
)

func RouteMessage(userId string, message string, payload string) {
	stage := GetUserConversationStage(userId)

	switch stage {
	case 1:
		usernameToSearch, found := detectUsername(message)
		if !found {
			sendTextMessage("Please send a username to search.", userId)
			return
		} else {
			SetUserConversationStage(userId, 2)
			searchForUserAndRespond(usernameToSearch, userId)
			return
		}
	case 2:
		if payload == "" {
			SetUserConversationStage(userId, 1)
			sendTextMessage("Restart", userId)
			return
		} else {
			username := strings.Split(payload, ":")[1]
			askForRating(username, userId)
			SetUserConversationStage(userId, 3)
			return
		}
	case 3:
		if payload == "" {
			SetUserConversationStage(userId, 1)
			sendTextMessage("Restart", userId)
			return
		} else {
			username := strings.Split(payload, ":")[1]
			rating, err := strconv.Atoi(strings.Split(payload, ":")[0])
			if err != nil {
				log.Fatal("Error parsing rating.")
			}
			SetUserConversationStage(userId, 1)
			saveRating(rating, userId, username)
			sendTextMessage("Thank you for your rating.", userId)
			return
		}
	}
	return
}
