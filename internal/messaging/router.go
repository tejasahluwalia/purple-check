package messaging

import (
	"log"
	"strings"
	"time"

	"purple-check/internal/database"
	"purple-check/internal/helpers"
)

func RouteMessage(userId string, message string, payload string) {
	state := getUserConversationState(userId)

	// Add database logging
	db, closer := database.GetDB()
	defer closer()
	_, err := db.Exec(
		"INSERT INTO user_message_logs (user_id, message, stage, created_at) VALUES (?, ?, ?, ?)",
		userId,
		message, // Logs both regular messages and payloads via message parameter
		state.Stage,
		time.Now(),
	)
	if err != nil {
		log.Printf("Failed to log message: %v", err)
	}

	switch state.Stage {
	case "START":
		usernameToSearch, found := helpers.DetectUsername(message)
		if found {
			searchForUserAndRespond(usernameToSearch, userId)
			return
		}
		if strings.HasPrefix(payload, "RATE:") {
			usernameToRate := strings.Split(payload, ":")[1]

			// Get the username of the person trying to leave feedback
			username, err := getUsernameFromUserID(userId)
			if err != nil {
				log.Printf("Failed to get username for user %s: %v", userId, err)
				sendTextMessage("Sorry, something went wrong. Please try again later.", userId)
				return
			}

			// Check if user is trying to rate themselves
			if strings.EqualFold(username, usernameToRate) {
				sendTextMessage("Sorry, you cannot leave feedback on your own profile.", userId)
				askForUsernameToSearch(userId)
				setUserConversationState(userId, ConversationState{
					Stage:       "START",
					CurrentUser: username,
				})
				return
			}

			setUserConversationState(userId, ConversationState{
				Stage:       "AWAITING_ROLE",
				TargetUser:  usernameToRate,
				CurrentUser: username,
			})
			askForRole(userId)
			return
		}
		if payload == "SEARCH" {
			askForUsernameToSearch(userId)
			return
		}
		askForUsernameToSearch(userId)
		return

	case "AWAITING_ROLE":
		if payload == "CANCEL" {
			setUserConversationState(userId, ConversationState{Stage: "START"})
			sendTextMessage("Feedback cancelled.", userId)
			askForUsernameToSearch(userId)
			return
		}
		if strings.HasPrefix(payload, "ROLE:") {
			role := strings.Split(payload, ":")[1]
			newState := state
			newState.Stage = "AWAITING_DEAL_STAGE"
			newState.Role = role
			setUserConversationState(userId, newState)
			askForDealStage(userId)
			return
		}
		invalidResponseMessage(userId)
		return

	case "AWAITING_DEAL_STAGE":
		if payload == "CANCEL" {
			setUserConversationState(userId, ConversationState{Stage: "START"})
			sendTextMessage("Feedback cancelled.", userId)
			askForUsernameToSearch(userId)
			return
		}
		if strings.HasPrefix(payload, "DEAL_STAGE:") {
			dealStage := strings.Split(payload, ":")[1]
			newState := state
			newState.Stage = "AWAITING_RATING"
			newState.DealStage = dealStage
			setUserConversationState(userId, newState)
			askForRating(state.TargetUser, userId)
			return
		}
		invalidResponseMessage(userId)
		return

	case "AWAITING_RATING":
		if payload == "CANCEL" {
			setUserConversationState(userId, ConversationState{Stage: "START"})
			sendTextMessage("Rating cancelled.", userId)
			askForUsernameToSearch(userId)
			return
		}
		if strings.HasPrefix(payload, "RATING:") {
			rating := strings.Split(payload, ":")[1]
			receiverUsername := state.TargetUser
			giverUsername := state.CurrentUser
			giverRole := state.Role
			var receiverRole string
			if giverRole == "buyer" {
				receiverRole = "seller"
			} else {
				receiverRole = "buyer"
			}
			saveRating(rating, giverUsername, receiverUsername, giverRole, receiverRole, state.DealStage)
			sendTextMessage("Thank you for submitting a rating.", userId)

			setUserConversationState(userId, ConversationState{Stage: "START"})
			askForUsernameToSearch(userId)
			return
		}
		invalidResponseMessage(userId)
		return
	}
}
