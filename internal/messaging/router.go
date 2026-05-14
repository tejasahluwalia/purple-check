package messaging

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	"purple-check/internal/config"
	"purple-check/internal/helpers"
	"purple-check/internal/models"
)

type Sender interface {
	SendTextMessage(text string, userID string) error
	SendButtonMessage(buttons []ElementButton, text string, userID string) error
	UsernameFromUserID(userID string) (string, error)
}

type InstagramSender struct{}

func (InstagramSender) SendTextMessage(text string, userID string) error {
	return sendTextMessage(text, userID)
}

func (InstagramSender) SendButtonMessage(buttons []ElementButton, text string, userID string) error {
	return sendButtonMessage(buttons, text, userID)
}

func (InstagramSender) UsernameFromUserID(userID string) (string, error) {
	return getUsernameFromUserID(userID)
}

type Router struct {
	Feedbacks   models.FeedbackRepository
	MessageLogs models.MessageLogRepository
	Sender      Sender
}

func NewRouter(feedbacks models.FeedbackRepository, messageLogs models.MessageLogRepository) *Router {
	if conversations == nil {
		InitConversations()
	}
	return &Router{
		Feedbacks:   feedbacks,
		MessageLogs: messageLogs,
		Sender:      InstagramSender{},
	}
}

func (router *Router) RouteMessage(ctx context.Context, messageEvent models.MessageEvent) {
	userId := messageEvent.Sender.Id
	if userId == "" {
		return
	}
	message := strings.TrimSpace(messageEvent.Message.Text)
	payload := getPayload(messageEvent)
	ref := getReferral(messageEvent)

	state := getUserConversationState(userId)
	if ref != "" {
		username := helpers.NormalizeUsername(ref)
		if err := helpers.ValidateUsername(username); err != nil {
			log.Printf("Ignoring invalid referral %q for user %s", ref, userId)
			ref = ""
		} else {
			state = ConversationState{Stage: stageStart, TargetUser: username, CurrentUser: state.CurrentUser}
			setUserConversationState(userId, state)
			ref = username
		}
	}

	if router.MessageLogs != nil {
		if err := router.MessageLogs.Insert(ctx, userId, message+payload+ref, state.Stage); err != nil {
			log.Printf("Failed to log message: %v", err)
		}
	}

	if ref != "" {
		router.beginRatingFlow(ctx, userId, ref)
		return
	}

	if payload == payloadSearch || payload == payloadLink {
		if state.TargetUser != "" {
			if err := router.searchForUserAndRespond(ctx, state.TargetUser, userId); err != nil {
				log.Printf("Failed to search for user %s: %v", state.TargetUser, err)
				logSend(router.sendTextMessage("Sorry, something went wrong. Please try again later.", userId))
			}
		} else {
			logSend(router.askForUsernameToSearch(userId))
		}
		return
	}

	if payload == payloadCancel {
		setUserConversationState(userId, ConversationState{Stage: stageStart, CurrentUser: state.CurrentUser})
		logSend(router.sendTextMessage("Feedback cancelled.", userId))
		logSend(router.askForUsernameToSearch(userId))
		return
	}

	switch state.Stage {
	case stageStart:
		usernameToSearch, found := helpers.DetectUsername(message)
		if found {
			if err := router.searchForUserAndRespond(ctx, usernameToSearch, userId); err != nil {
				log.Printf("Failed to search for user %s: %v", usernameToSearch, err)
				logSend(router.sendTextMessage("Sorry, something went wrong. Please try again later.", userId))
			}
			return
		}
		if usernameToRate, err := parseRatePayload(payload); err == nil {
			router.beginRatingFlow(ctx, userId, usernameToRate)
			return
		}
		logSend(router.askForUsernameToSearch(userId))
		return

	case stageAwaitingRole:
		if role, ok := parseRolePayload(payload); ok {
			newState := state
			newState.Stage = stageAwaitingDealStage
			newState.Role = role
			if err := router.askForDealStage(userId); err != nil {
				logSend(err)
				return
			}
			setUserConversationState(userId, newState)
			return
		}
		logSend(router.invalidResponseMessage(userId))
		return

	case stageAwaitingDealStage:
		if dealStage, ok := parseDealStagePayload(payload); ok {
			newState := state
			newState.Stage = stageAwaitingRating
			newState.DealStage = dealStage
			if err := router.askForRating(state.TargetUser, userId); err != nil {
				logSend(err)
				return
			}
			setUserConversationState(userId, newState)
			return
		}
		logSend(router.invalidResponseMessage(userId))
		return

	case stageAwaitingRating:
		rating, payloadTarget, err := parseRatingPayload(payload)
		if err != nil {
			logSend(router.invalidResponseMessage(userId))
			return
		}
		if !strings.EqualFold(payloadTarget, state.TargetUser) {
			logSend(router.sendTextMessage("That rating option is stale. Please start again.", userId))
			setUserConversationState(userId, ConversationState{Stage: stageStart, CurrentUser: state.CurrentUser})
			logSend(router.askForUsernameToSearch(userId))
			return
		}
		receiverUsername := payloadTarget
		giverUsername := state.CurrentUser
		giverRole := state.Role
		if strings.EqualFold(giverUsername, receiverUsername) {
			logSend(router.sendTextMessage("Sorry, you cannot leave feedback on your own profile.", userId))
			setUserConversationState(userId, ConversationState{Stage: stageStart, CurrentUser: giverUsername})
			logSend(router.askForUsernameToSearch(userId))
			return
		}
		receiverRole, ok := receiverRoleFor(giverRole)
		if !ok {
			logSend(router.invalidResponseMessage(userId))
			return
		}
		if err := router.saveRating(ctx, rating, giverUsername, receiverUsername, giverRole, receiverRole, state.DealStage); err != nil {
			log.Printf("Failed to save rating: %v", err)
			logSend(router.sendTextMessage("Sorry, something went wrong while saving your rating. Please try again later.", userId))
			return
		}
		logSend(router.sendTextMessage("Thank you for submitting a rating.", userId))

		setUserConversationState(userId, ConversationState{Stage: stageStart, CurrentUser: giverUsername})
		logSend(router.askForUsernameToSearch(userId))
	}
}

func (router *Router) beginRatingFlow(ctx context.Context, userId string, usernameToRate string) {
	usernameToRate = helpers.NormalizeUsername(usernameToRate)
	err := helpers.ValidateUsername(usernameToRate)
	if err != nil {
		logSend(router.invalidResponseMessage(userId))
		return
	}

	username, err := router.sender().UsernameFromUserID(userId)
	if err != nil {
		log.Printf("Failed to get username for user %s: %v", userId, err)
		logSend(router.sendTextMessage("Sorry, something went wrong. Please try again later.", userId))
		return
	}

	if strings.EqualFold(username, usernameToRate) {
		logSend(router.sendTextMessage("Sorry, you cannot leave feedback on your own profile.", userId))
		setUserConversationState(userId, ConversationState{Stage: stageStart, CurrentUser: username})
		logSend(router.askForUsernameToSearch(userId))
		return
	}

	if err := router.askForRole(userId); err != nil {
		logSend(err)
		return
	}
	setUserConversationState(userId, ConversationState{
		Stage:       stageAwaitingRole,
		TargetUser:  usernameToRate,
		CurrentUser: username,
	})
}

func (router *Router) searchForUserAndRespond(ctx context.Context, usernameToSearch string, userId string) error {
	if router.Feedbacks == nil {
		return fmt.Errorf("feedback repository is nil")
	}
	positiveRatings, err := router.Feedbacks.CountPositiveForUser(ctx, usernameToSearch)
	if err != nil {
		return err
	}
	totalRatings, err := router.Feedbacks.CountForUser(ctx, usernameToSearch)
	if err != nil {
		return err
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
			Payload: payloadSearch,
		},
	}

	if totalRatings == 0 {
		return router.sendButtonMessage(buttons, "No ratings found for @"+usernameToSearch, userId)
	}
	positivePercentage := (float64(positiveRatings) / float64(totalRatings)) * 100
	ratingPlural := "ratings"
	if totalRatings == 1 {
		ratingPlural = "rating"
	}
	text := "@" + usernameToSearch + "\n\n" + strconv.FormatFloat(positivePercentage, 'f', 0, 64) + "% positive (" + strconv.Itoa(totalRatings) + " " + ratingPlural + ")"
	return router.sendButtonMessage(buttons, text, userId)
}

func (router *Router) saveRating(ctx context.Context, rating string, giverUsername string, receiverUsername string, giverRole string, receiverRole string, dealStage string) error {
	if router.Feedbacks == nil {
		return fmt.Errorf("feedback repository is nil")
	}
	return router.Feedbacks.InsertOrUpdateOne(ctx, models.FeedbackInput{
		Giver:        giverUsername,
		Receiver:     receiverUsername,
		Rating:       rating,
		GiverRole:    giverRole,
		ReceiverRole: receiverRole,
		DealStage:    dealStage,
	})
}

func (router *Router) sender() Sender {
	if router.Sender != nil {
		return router.Sender
	}
	return InstagramSender{}
}

func (router *Router) sendTextMessage(text string, userId string) error {
	return router.sender().SendTextMessage(text, userId)
}

func (router *Router) sendButtonMessage(buttons []ElementButton, text string, userId string) error {
	return router.sender().SendButtonMessage(buttons, text, userId)
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
