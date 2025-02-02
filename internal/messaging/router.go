package messaging

import (
	"log"
	"strings"
	"time"

	"purple-check/internal/database"
	"purple-check/internal/helpers"
)

func RouteMessage(userId string, message string, payload string, ref string) {
	state := getUserConversationState(userId)

	// Add database logging
	db, closer := database.GetDB()
	defer closer()
	_, err := db.Exec(
		"INSERT INTO user_message_logs (user_id, message, stage, created_at) VALUES (?, ?, ?, ?)",
		userId,
		message+payload+ref,
		state.Stage,
		time.Now(),
	)
	if err != nil {
		log.Printf("Failed to log message: %v", err)
	}

	if ref != "" {
		// Handle referral
		setUserConversationState(userId, ConversationState{Stage: "START", TargetUser: ref})
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
		if payload == "SEARCH" || payload == "LINK" {
			if state.TargetUser != "" {
				searchForUserAndRespond(state.TargetUser, userId)
			} else {
				askForUsernameToSearch(userId)
			}
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
			if strings.EqualFold(giverRole, "BUYER") {
				receiverRole = "SELLER"
			} else if strings.EqualFold(giverRole, "SELLER") {
				receiverRole = "BUYER"
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
