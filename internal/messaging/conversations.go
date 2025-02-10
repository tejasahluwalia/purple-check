package messaging

import "purple-check/internal/cache"

type ConversationState struct {
	Stage       string
	TargetUser  string
	Role        string
	DealStage   string
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
		newState := ConversationState{Stage: "START"}
		conversations.Set(userId, newState)
		return newState
	}
	return state
}

func setUserConversationState(userId string, state ConversationState) {
	conversations.Set(userId, state)
}
