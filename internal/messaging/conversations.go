package messaging

type UserConversations map[string]string

var conversations UserConversations

func InitConversations() {
	conversations = make(UserConversations)
	return
}

func GetUserConversationStage(userId string) string {
	userConversationStage, exists := conversations[userId]
	if exists {
		return userConversationStage
	} else {
		stage := "START"
		conversations[userId] = stage
		return stage
	}
}

func SetUserConversationStage(userId string, stage string) {
	conversations[userId] = stage
}
