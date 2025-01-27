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
		message,  // Logs both regular messages and payloads via message parameter
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

		if payload != "" {
			switch strings.Split(payload, ":")[0] {
			case "RATE":
				usernameToRate := strings.Split(payload, ":")[1]
				setUserConversationState(userId, ConversationState{
					Stage:      "AWAITING_ROLE",
					TargetUser: usernameToRate,
				})
				askForRole(userId)
				return
			case "SEARCH":
				askForUsernameToSearch(userId)
				return
			}
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
		invalidRatingMessage(userId)
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
		invalidRatingMessage(userId)
		return

	case "AWAITING_RATING":
		if payload == "CANCEL" {
			setUserConversationState(userId, ConversationState{Stage: "START"})
			sendTextMessage("Rating cancelled.", userId)
			askForUsernameToSearch(userId)
			return
		} else if len(strings.Split(payload, ":")) == 2 {
			rating := strings.Split(payload, ":")[0]
			if rating == "POSITIVE" || rating == "NEUTRAL" || rating == "NEGATIVE" {
				setUserConversationState(userId, ConversationState{Stage: "START"})
				saveRating(rating, userId, state.TargetUser, state.Role, state.DealStage)
				sendTextMessage("Thank you for submitting a rating.", userId)
				askForUsernameToSearch(userId)
				return
			}
		}
		invalidRatingMessage(userId)
		return
	}
}
