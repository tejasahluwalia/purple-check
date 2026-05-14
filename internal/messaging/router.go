package messaging

import (
	"context"
	"errors"
	"log"
	"strings"

	"purple-check/internal/helpers"
	"purple-check/internal/models"
)

// func RouteMessage(ctx context.Context, messageEvent models.MessageEvent) {
// 	userId := messageEvent.Sender.Id
// 	if userId == "" {
// 		return
// 	}
// 	message := strings.TrimSpace(messageEvent.Message.Text)
// 	payload := getPayload(messageEvent)
// 	ref := getReferral(messageEvent)

// 	state := getUserConversationState(userId)
// 	if ref != "" {
// 		username, ok := helpers.NormalizeUsername(ref)
// 		if !ok {
// 			log.Printf("Ignoring invalid referral %q for user %s", ref, userId)
// 		} else {
// 			state = ConversationState{Stage: stageStart, TargetUser: username, CurrentUser: state.CurrentUser}
// 			setUserConversationState(userId, state)
// 			ref = username
// 		}
// 	}

// 	// Add database logging
// 	db, closer, err := database.GetDB(ctx)
// 	if err != nil {
// 		log.Printf("Failed to open database for message log: %v", err)
// 	} else {
// 		defer closer()
// 		_, err = db.Exec(
// 			"INSERT INTO user_message_logs (user_id, message, stage, created_at) VALUES (?, ?, ?, ?)",
// 			userId,
// 			message+payload+ref,
// 			state.Stage,
// 			time.Now(),
// 		)
// 		if err != nil {
// 			log.Printf("Failed to log message: %v", err)
// 		} else if err := database.PushDB(ctx); err != nil {
// 			log.Printf("Failed to sync message log: %v", err)
// 		}
// 	}

// 	if ref != "" {
// 		beginRatingFlow(ctx, userId, ref)
// 		return
// 	}

// 	if payload == payloadSearch || payload == payloadLink {
// 		if state.TargetUser != "" {
// 			if err := searchForUserAndRespond(ctx, state.TargetUser, userId); err != nil {
// 				log.Printf("Failed to search for user %s: %v", state.TargetUser, err)
// 				logSend(sendTextMessage("Sorry, something went wrong. Please try again later.", userId))
// 			}
// 		} else {
// 			logSend(askForUsernameToSearch(userId))
// 		}
// 		return
// 	}

// 	if payload == payloadCancel {
// 		setUserConversationState(userId, ConversationState{Stage: stageStart, CurrentUser: state.CurrentUser})
// 		logSend(sendTextMessage("Feedback cancelled.", userId))
// 		logSend(askForUsernameToSearch(userId))
// 		return
// 	}

// 	switch state.Stage {
// 	case stageStart:
// 		usernameToSearch, found := helpers.DetectUsername(message)
// 		if found {
// 			if err := searchForUserAndRespond(ctx, usernameToSearch, userId); err != nil {
// 				log.Printf("Failed to search for user %s: %v", usernameToSearch, err)
// 				logSend(sendTextMessage("Sorry, something went wrong. Please try again later.", userId))
// 			}
// 			return
// 		}
// 		if usernameToRate, ok := parseRatePayload(payload); ok {
// 			beginRatingFlow(ctx, userId, usernameToRate)
// 			return
// 		}
// 		logSend(askForUsernameToSearch(userId))
// 		return

// 	case stageAwaitingRole:
// 		if role, ok := parseRolePayload(payload); ok {
// 			newState := state
// 			newState.Stage = stageAwaitingDealStage
// 			newState.Role = role
// 			if err := askForDealStage(userId); err != nil {
// 				logSend(err)
// 				return
// 			}
// 			setUserConversationState(userId, newState)
// 			return
// 		}
// 		logSend(invalidResponseMessage(userId))
// 		return

// 	case stageAwaitingDealStage:
// 		if dealStage, ok := parseDealStagePayload(payload); ok {
// 			newState := state
// 			newState.Stage = stageAwaitingRating
// 			newState.DealStage = dealStage
// 			if err := askForRating(state.TargetUser, userId); err != nil {
// 				logSend(err)
// 				return
// 			}
// 			setUserConversationState(userId, newState)
// 			return
// 		}
// 		logSend(invalidResponseMessage(userId))
// 		return

// 	case stageAwaitingRating:
// 		if rating, payloadTarget, ok := parseRatingPayload(payload); ok {
// 			if !strings.EqualFold(payloadTarget, state.TargetUser) {
// 				logSend(sendTextMessage("That rating option is stale. Please start again.", userId))
// 				setUserConversationState(userId, ConversationState{Stage: stageStart, CurrentUser: state.CurrentUser})
// 				logSend(askForUsernameToSearch(userId))
// 				return
// 			}
// 			receiverUsername := payloadTarget
// 			giverUsername := state.CurrentUser
// 			giverRole := state.Role
// 			if strings.EqualFold(giverUsername, receiverUsername) {
// 				logSend(sendTextMessage("Sorry, you cannot leave feedback on your own profile.", userId))
// 				setUserConversationState(userId, ConversationState{Stage: stageStart, CurrentUser: giverUsername})
// 				logSend(askForUsernameToSearch(userId))
// 				return
// 			}
// 			receiverRole, ok := receiverRoleFor(giverRole)
// 			if !ok {
// 				logSend(invalidResponseMessage(userId))
// 				return
// 			}
// 			if err := saveRating(ctx, rating, giverUsername, receiverUsername, giverRole, receiverRole, state.DealStage); err != nil {
// 				log.Printf("Failed to save rating: %v", err)
// 				logSend(sendTextMessage("Sorry, something went wrong while saving your rating. Please try again later.", userId))
// 				return
// 			}
// 			logSend(sendTextMessage("Thank you for submitting a rating.", userId))

// 			setUserConversationState(userId, ConversationState{Stage: stageStart, CurrentUser: giverUsername})
// 			logSend(askForUsernameToSearch(userId))
// 			return
// 		}
// 		logSend(invalidResponseMessage(userId))
// 		return
// 	}
// }

func beginRatingFlow(ctx context.Context, userId string, usernameToRate string) {
	usernameToRate = helpers.NormalizeUsername(usernameToRate)
	err := helpers.ValidateUsername(usernameToRate)
	if err != nil {
		logSend(invalidResponseMessage(userId))
		return
	}

	username, err := getUsernameFromUserID(userId)
	if err != nil {
		log.Printf("Failed to get username for user %s: %v", userId, err)
		logSend(sendTextMessage("Sorry, something went wrong. Please try again later.", userId))
		return
	}

	if strings.EqualFold(username, usernameToRate) {
		logSend(sendTextMessage("Sorry, you cannot leave feedback on your own profile.", userId))
		setUserConversationState(userId, ConversationState{Stage: stageStart, CurrentUser: username})
		logSend(askForUsernameToSearch(userId))
		return
	}

	if err := askForRole(userId); err != nil {
		logSend(err)
		return
	}
	setUserConversationState(userId, ConversationState{
		Stage:       stageAwaitingRole,
		TargetUser:  usernameToRate,
		CurrentUser: username,
	})
}

func getPayload(messageEvent models.MessageEvent) string {
	if messageEvent.Postback != nil {
		return messageEvent.Postback.Payload
	}
	return messageEvent.Message.Quick_reply.Payload
}

func getReferral(messageEvent models.MessageEvent) string {
	if messageEvent.Referral != nil {
		return messageEvent.Referral.Ref
	}
	if messageEvent.Message.Referral != nil {
		return messageEvent.Message.Referral.Ref
	}
	if messageEvent.Postback != nil && messageEvent.Postback.Referral != nil {
		return messageEvent.Postback.Referral.Ref
	}
	return ""
}

func parseRatePayload(payload string) (string, error) {
	command, value, ok := strings.Cut(payload, ":")
	if !ok || command != "RATE" {
		return "", errors.New("Invalid payload")
	}
	username := helpers.NormalizeUsername(value)
	err := helpers.ValidateUsername(username)
	if err != nil {
		return "", errors.New("Invalid payload")
	}
	return username, nil
}

func parseRolePayload(payload string) (string, bool) {
	command, value, ok := strings.Cut(payload, ":")
	if !ok || command != "ROLE" {
		return "", false
	}
	if value != roleBuyer && value != roleSeller {
		return "", false
	}
	return value, true
}

func parseDealStagePayload(payload string) (string, bool) {
	command, value, ok := strings.Cut(payload, ":")
	if !ok || command != "DEAL_STAGE" {
		return "", false
	}
	if value != dealStageComplete && value != dealStageIncomplete {
		return "", false
	}
	return value, true
}

func parseRatingPayload(payload string) (string, string, error) {
	parts := strings.Split(payload, ":")
	if len(parts) != 3 || parts[0] != "RATING" {
		return "", "", errors.New("Invalid payload")
	}
	if parts[1] != ratingPositive && parts[1] != ratingNegative && parts[1] != ratingMixed {
		return "", "", errors.New("Invalid payload")
	}
	username := helpers.NormalizeUsername(parts[2])
	err := helpers.ValidateUsername(username)
	if err != nil {
		return "", "", errors.New("Invalid payload")
	}
	return parts[1], username, nil
}

func receiverRoleFor(giverRole string) (string, bool) {
	switch giverRole {
	case roleBuyer:
		return roleSeller, true
	case roleSeller:
		return roleBuyer, true
	default:
		return "", false
	}
}

func logSend(err error) {
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}
