package messaging

type ConversationState struct {
	Stage       string
	TargetUser  string
	Role        string
	DealStage   string
	CurrentUser string
}

type UserConversations map[string]ConversationState

var conversations UserConversations

func InitConversations() {
	conversations = make(UserConversations)
}

func getUserConversationState(userId string) ConversationState {
	state, exists := conversations[userId]
	if !exists {
		newState := ConversationState{Stage: "START"}
		conversations[userId] = newState
		return newState
	}
	return state
}

func setUserConversationState(userId string, state ConversationState) {
	conversations[userId] = state
}

// GetUserConversationState exposes the conversation state for testing
func GetUserConversationState(userId string) ConversationState {
	return getUserConversationState(userId)
}

// SetUserConversationState exposes the state setter for testing
func SetUserConversationState(userId string, state ConversationState) {
	setUserConversationState(userId, state)
}
