package messaging

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
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

type Referral struct {
	Receiver  string
	GiverRole models.TransactionRole
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
		slog.Error("Sender ID does not exist, this should never happen.")
		return
	}

	// Send mark_seen and typing_on to acknowledge the user's message.
	router.sendSenderAction(ctx, "mark_seen", userId)
	router.sendSenderAction(ctx, "typing_on", userId)
	defer func() {
		router.sendSenderAction(ctx, "typing_off", userId)
	}()

	message := strings.TrimSpace(messageEvent.Message.Text)
	payload := getPayload(messageEvent)
	ref := getReferral(messageEvent)
	attachments := messageEvent.Message.Attachments

	state := getUserConversationState(userId)
	if ref != nil {
		username := helpers.NormalizeUsername(ref.Receiver)
		if err := helpers.ValidateUsername(username); err != nil {
			log.Printf("Ignoring invalid referral %q for user %s", ref, userId)
		} else {
			state = ConversationState{Stage: stageAwaitingRating, TargetUser: username, CurrentUser: state.CurrentUser, Role: ref.GiverRole}
			setUserConversationState(userId, state)
			router.beginRatingFlow(ctx, userId, ref.Receiver)
			return
		}
	}

	if len(attachments) > 0 {
		getUsernameFromAttachment(router.sender(), ctx, attachments)
	}

	if router.MessageLogs != nil {
		if err := router.MessageLogs.Insert(ctx, userId, fmt.Sprintf(message, payload, ref), state.Stage); err != nil {
			log.Printf("Failed to log message: %v", err)
		}
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
	var positiveCount, negativeCount, mixedCount int
	for _, feedback := range feedbackList {
		if feedback.Rating == models.PositiveFeedback {
			positiveCount += 1
		}
		if feedback.Rating == models.NegativeFeedback {
			negativeCount += 1
		}
		if feedback.Rating == models.MixedFeedback {
			mixedCount += 1
		}
	}

	score := models.CalculateScore(positiveCount, mixedCount, negativeCount)
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
	ratingPlural := "ratings"
	if totalRatings == 1 {
		ratingPlural = "rating"
	}
	text := fmt.Sprintf("@%s\n\nScore: %s/100 (%d %s)", usernameToSearch, strconv.FormatFloat(score*100, 'f', 0, 64), totalRatings, ratingPlural)
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

func getReferral(messageEvent models.MessageEvent) *Referral {
	if messageEvent.Referral != nil {
		return parseReferralParameter(messageEvent.Referral.Ref)
	}
	if messageEvent.Message.Referral != nil {
		return parseReferralParameter(messageEvent.Message.Referral.Ref)
	}
	if messageEvent.Postback != nil && messageEvent.Postback.Referral != nil {
		return parseReferralParameter(messageEvent.Postback.Referral.Ref)
	}
	return nil
}

func getUsernameFromAttachment(sender Sender, ctx context.Context, attachments []models.Attachment) string {
	for _, attachment := range attachments {
		username, err := sender.UsernameFromMedia(ctx, attachment.Payload.IGPostMediaID)
		log.Print(username, err)
	}
	return ""
}

// Referral parameter is a string in the form receiver=<receiver_username>-giver_role=<giver_role>
func parseReferralParameter(ref string) *Referral {
	log.Print(ref)
	return nil
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
