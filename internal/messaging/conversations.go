package messaging

type UserConversations map[string]int

var conversations UserConversations

func InitConversations() {
	conversations = make(UserConversations)
	return
}

func GetUserConversationStage(userId string) int {
	userConversationStage, exists := conversations[userId]
	if exists {
		return userConversationStage
	} else {
		stage := 1
		conversations[userId] = stage
		return stage
	}
}

func SetUserConversationStage(userId string, stage int) {
	conversations[userId] = stage
}
