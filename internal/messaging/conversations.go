package messaging

type UserConversations map[string]string

var conversations UserConversations

func InitConversations() {
	conversations = make(UserConversations)
	return
}

func getUserConversationStage(userId string) string {
	userConversationStage, exists := conversations[userId]
	if exists {
		return userConversationStage
	} else {
		stage := "START"
		conversations[userId] = stage
		return stage
	}
}

func setUserConversationStage(userId string, stage string) {
	conversations[userId] = stage
}

// GetUserConversationStage exposes the conversation stage for testing
func GetUserConversationStage(userId string) string {
	return getUserConversationStage(userId)
}

// SetUserConversationStage exposes the stage setter for testing
func SetUserConversationStage(userId string, stage string) {
	setUserConversationStage(userId, stage)
}
