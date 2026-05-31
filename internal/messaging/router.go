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

type Router struct {
	Feedbacks   models.FeedbackRepository
	MessageLogs models.MessageLogRepository
	Sender      Sender
	Tokens      AccountTokenReader
}

func NewRouter(feedbacks models.FeedbackRepository, messageLogs models.MessageLogRepository, tokens AccountTokenReader) *Router {
	if conversations == nil {
		InitConversations()
	}
	return &Router{
		Feedbacks:   feedbacks,
		MessageLogs: messageLogs,
		Sender:      InstagramSender{Tokens: tokens},
		Tokens:      tokens,
	}
}

func (router *Router) RouteMessage(ctx context.Context, messageEvent models.MessageEvent) {
	userId := messageEvent.Sender.Id
	if userId == "" {
		return
	}

	// Send mark_seen and typing_on to acknowledge the user's message.
	logSend(router.sendSenderAction(ctx, "mark_seen", userId))
	logSend(router.sendSenderAction(ctx, "typing_on", userId))
	defer func() {
		logSend(router.sendSenderAction(ctx, "typing_off", userId))
	}()

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
				logSend(router.sendTextMessage(ctx, "Sorry, something went wrong. Please try again later.", userId))
			}
		} else {
			logSend(router.askForUsernameToSearch(ctx, userId))
		}
		return
	}

	if payload == payloadCancel {
		setUserConversationState(userId, ConversationState{Stage: stageStart, CurrentUser: state.CurrentUser})
		logSend(router.sendTextMessage(ctx, "Feedback cancelled.", userId))
		logSend(router.askForUsernameToSearch(ctx, userId))
		return
	}

	switch state.Stage {
	case stageStart:
		usernameToSearch, found := helpers.DetectUsername(message)
		if found {
			if err := router.searchForUserAndRespond(ctx, usernameToSearch, userId); err != nil {
				log.Printf("Failed to search for user %s: %v", usernameToSearch, err)
				logSend(router.sendTextMessage(ctx, "Sorry, something went wrong. Please try again later.", userId))
			}
			return
		}
		if usernameToRate, err := parseRatePayload(payload); err == nil {
			router.beginRatingFlow(ctx, userId, usernameToRate)
			return
		}
		logSend(router.askForUsernameToSearch(ctx, userId))
		return

	case stageAwaitingRole:
		if role, ok := parseRolePayload(payload); ok {
			newState := state
			newState.Stage = stageAwaitingDealStage
			newState.Role = role
			if err := router.askForDealStage(ctx, userId); err != nil {
				logSend(err)
				return
			}
			setUserConversationState(userId, newState)
			return
		}
		logSend(router.invalidResponseMessage(ctx, userId))
		return

	case stageAwaitingDealStage:
		if dealStage, ok := parseDealStagePayload(payload); ok {
			newState := state
			newState.Stage = stageAwaitingRating
			newState.DealStage = dealStage
			if err := router.askForRating(ctx, state.TargetUser, userId); err != nil {
				logSend(err)
				return
			}
			setUserConversationState(userId, newState)
			return
		}
		logSend(router.invalidResponseMessage(ctx, userId))
		return

	case stageAwaitingRating:
		rating, payloadTarget, err := parseRatingPayload(payload)
		if err != nil {
			logSend(router.invalidResponseMessage(ctx, userId))
			return
		}
		if !strings.EqualFold(payloadTarget, state.TargetUser) {
			logSend(router.sendTextMessage(ctx, "That rating option is stale. Please start again.", userId))
			setUserConversationState(userId, ConversationState{Stage: stageStart, CurrentUser: state.CurrentUser})
			logSend(router.askForUsernameToSearch(ctx, userId))
			return
		}
		receiverUsername := payloadTarget
		giverUsername := state.CurrentUser
		giverRole := state.Role
		if strings.EqualFold(giverUsername, receiverUsername) {
			logSend(router.sendTextMessage(ctx, "Sorry, you cannot leave feedback on your own profile.", userId))
			setUserConversationState(userId, ConversationState{Stage: stageStart, CurrentUser: giverUsername})
			logSend(router.askForUsernameToSearch(ctx, userId))
			return
		}
		receiverRole := receiverRoleFor(giverRole)

		if err := router.saveRating(ctx, models.FeedbackSentiment(rating), giverUsername, receiverUsername, giverRole, receiverRole, state.DealStage); err != nil {
			log.Printf("Failed to save rating: %v", err)
			logSend(router.sendTextMessage(ctx, "Sorry, something went wrong while saving your rating. Please try again later.", userId))
			return
		}
		logSend(router.sendTextMessage(ctx, "Thank you for submitting a rating.", userId))

		setUserConversationState(userId, ConversationState{Stage: stageStart, CurrentUser: giverUsername})
		logSend(router.askForUsernameToSearch(ctx, userId))
	}
}

func (router *Router) beginRatingFlow(ctx context.Context, userId string, usernameToRate string) {
	usernameToRate = helpers.NormalizeUsername(usernameToRate)
	err := helpers.ValidateUsername(usernameToRate)
	if err != nil {
		logSend(router.invalidResponseMessage(ctx, userId))
		return
	}

	username, err := router.sender().UsernameFromUserID(ctx, userId)
	if err != nil {
		log.Printf("Failed to get username for user %s: %v", userId, err)
		logSend(router.sendTextMessage(ctx, "Sorry, something went wrong. Please try again later.", userId))
		return
	}

	if strings.EqualFold(username, usernameToRate) {
		logSend(router.sendTextMessage(ctx, "Sorry, you cannot leave feedback on your own profile.", userId))
		setUserConversationState(userId, ConversationState{Stage: stageStart, CurrentUser: username})
		logSend(router.askForUsernameToSearch(ctx, userId))
		return
	}

	if err := router.askForRole(ctx, userId); err != nil {
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
	feedbackList, err := router.Feedbacks.GetAll(ctx, usernameToSearch, models.FeedbackRoleReceiver)
	if err != nil {
		return err
	}

	totalRatings := len(feedbackList)
	positiveRatings := 0
	for _, feedback := range feedbackList {
		if feedback.Rating == models.PositiveFeedback {
			positiveRatings++
		}
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
		return router.sendButtonMessage(ctx, buttons, "No ratings found for @"+usernameToSearch, userId)
	}
	positivePercentage := (float64(positiveRatings) / float64(totalRatings)) * 100
	ratingPlural := "ratings"
	if totalRatings == 1 {
		ratingPlural = "rating"
	}
	text := "@" + usernameToSearch + "\n\n" + strconv.FormatFloat(positivePercentage, 'f', 0, 64) + "% positive (" + strconv.Itoa(totalRatings) + " " + ratingPlural + ")"
	return router.sendButtonMessage(ctx, buttons, text, userId)
}

func (router *Router) saveRating(ctx context.Context, rating models.FeedbackSentiment, giverUsername string, receiverUsername string, giverRole models.TransactionRole, receiverRole models.TransactionRole, dealStage models.DealStage) error {
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
	return InstagramSender{Tokens: router.Tokens}
}

func (router *Router) sendTextMessage(ctx context.Context, text string, userId string) error {
	return router.sender().SendTextMessage(ctx, text, userId)
}

func (router *Router) sendButtonMessage(ctx context.Context, buttons []ElementButton, text string, userId string) error {
	return router.sender().SendButtonMessage(ctx, buttons, text, userId)
}

func (router *Router) sendSenderAction(ctx context.Context, action string, userId string) error {
	return router.sender().SendSenderAction(ctx, action, userId)
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

func parseRolePayload(payload string) (models.TransactionRole, bool) {
	command, value, ok := strings.Cut(payload, ":")
	if !ok || command != "ROLE" {
		return "", false
	}

	if value == string(models.TransactionRoleBuyer) || value == string(models.TransactionRoleSeller) {
		role := models.TransactionRole(value)
		return role, true
	}
	return "", false
}

func parseDealStagePayload(payload string) (models.DealStage, bool) {
	command, value, ok := strings.Cut(payload, ":")
	if !ok || command != "DEAL_STAGE" {
		return "", false
	}
	if value == string(models.DealStageComplete) || value == string(models.DealStageIncomplete) {
		stage := models.DealStage(value)
		return stage, true
	}
	return "", false
}

func parseRatingPayload(payload string) (models.FeedbackSentiment, string, error) {
	parts := strings.Split(payload, ":")
	if len(parts) != 3 || parts[0] != "RATING" {
		return "", "", errors.New("Invalid payload")
	}
	if parts[1] != string(models.PositiveFeedback) && parts[1] != string(models.NegativeFeedback) {
		return "", "", errors.New("Invalid payload")
	}
	username := helpers.NormalizeUsername(parts[2])
	err := helpers.ValidateUsername(username)
	if err != nil {
		return "", "", errors.New("Invalid payload")
	}
	sentiment := models.FeedbackSentiment(parts[1])
	return sentiment, username, nil
}

func receiverRoleFor(giverRole models.TransactionRole) models.TransactionRole {
	if giverRole == models.TransactionRoleBuyer {
		return models.TransactionRoleSeller
	}
	return models.TransactionRoleBuyer
}

func logSend(err error) {
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}
