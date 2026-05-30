package messaging

import (
	"purple-check/internal/cache"
	"purple-check/internal/models"
)

const (
	stageStart             = "START"
	stageAwaitingRole      = "AWAITING_ROLE"
	stageAwaitingDealStage = "AWAITING_DEAL_STAGE"
	stageAwaitingRating    = "AWAITING_RATING"

	payloadSearch = "SEARCH"
	payloadLink   = "LINK"
	payloadCancel = "CANCEL"

	roleBuyer  = "BUYER"
	roleSeller = "SELLER"

	dealStageComplete   = "COMPLETE"
	dealStageIncomplete = "INCOMPLETE"

	ratingPositive = "POSITIVE"
	ratingNegative = "NEGATIVE"
	ratingMixed    = "MIXED"
)

type ConversationState struct {
	Stage       string
	TargetUser  string
	Role        models.TransactionRole
	DealStage   models.DealStage
	CurrentUser string
}

var conversations *cache.Cache[string, ConversationState]

func InitConversations() {
	conversations = cache.New[string, ConversationState]()
}

type UserConversations map[string]ConversationState

func getUserConversationState(userId string) ConversationState {
	state, exists := conversations.Get(userId)
	if !exists {
		newState := ConversationState{Stage: stageStart}
		conversations.Set(userId, newState)
		return newState
	}
	return state
}

func setUserConversationState(userId string, state ConversationState) {
	conversations.Set(userId, state)
}
