package messaging_test

import (
	"purple-check/internal/messaging"
	"testing"
)

// Mock message sending
type testMessenger struct {
	lastMessage  string
	lastPayload  string
	searchCalled bool
}

func (t *testMessenger) sendText(text string, userId string) {
	t.lastMessage = text
}

func (t *testMessenger) askForRating(user string, userId string) {
	t.lastPayload = "RATE:" + user
}

func (t *testMessenger) searchForUser(user string, userId string) {
	t.searchCalled = true
}

func TestConversationFlow(t *testing.T) {
	t.Run("START stage with username detection", func(t *testing.T) {
		messaging.InitConversations()
		mock := &testMessenger{}
		
		// Override dependencies
		origSearch := messaging.SearchForUserAndRespond
		messaging.SearchForUserAndRespond = mock.searchForUser
		defer func() { messaging.SearchForUserAndRespond = origSearch }()

		messaging.RouteMessage("user1", "Check @testuser", "")
		
		if !mock.searchCalled {
			t.Error("Should trigger user search")
		}
		if messaging.GetUserConversationStage("user1") != "START" {
			t.Error("Should remain in START stage after search")
		}
	})

	t.Run("START stage with RATE payload", func(t *testing.T) {
		messaging.InitConversations()
		mock := &testMessenger{}
		
		origAsk := messaging.AskForRating
		messaging.AskForRating = mock.askForRating
		defer func() { messaging.AskForRating = origAsk }()

		messaging.RouteMessage("user1", "", "RATE:testuser")
		
		if mock.lastPayload != "RATE:testuser" {
			t.Error("Should ask for rating")
		}
		if messaging.GetUserConversationStage("user1") != "AWAITING_RATING" {
			t.Error("Should transition to AWAITING_RATING stage")
		}
	})

	t.Run("AWAITING_RATING with valid payload", func(t *testing.T) {
		messaging.InitConversations()
		mock := &testMessenger{}
		messaging.SetUserConversationStage("user1", "AWAITING_RATING")

		origSave := messaging.SaveRating
		saved := false
		messaging.SaveRating = func(rating, userId, username string) {
			saved = true
		}
		defer func() { messaging.SaveRating = origSave }()

		messaging.RouteMessage("user1", "", "POSITIVE:testuser")
		
		if !saved {
			t.Error("Should save rating")
		}
		if messaging.GetUserConversationStage("user1") != "START" {
			t.Error("Should return to START stage")
		}
	})

	t.Run("AWAITING_RATING with cancel payload", func(t *testing.T) {
		messaging.InitConversations()
		mock := &testMessenger{}
		messaging.SetUserConversationStage("user1", "AWAITING_RATING")

		messaging.RouteMessage("user1", "", "CANCEL")
		
		if messaging.GetUserConversationStage("user1") != "START" {
			t.Error("Should reset to START stage")
		}
	})

	t.Run("AWAITING_RATING with text cancel", func(t *testing.T) {
		messaging.InitConversations()
		mock := &testMessenger{}
		messaging.SetUserConversationStage("user1", "AWAITING_RATING")

		messaging.RouteMessage("user1", "cancel", "")
		
		if messaging.GetUserConversationStage("user1") != "START" {
			t.Error("Should handle text cancellation")
		}
	})

	t.Run("AWAITING_RATING with invalid payload", func(t *testing.T) {
		messaging.InitConversations()
		mock := &testMessenger{}
		messaging.SetUserConversationStage("user1", "AWAITING_RATING")

		origInvalid := messaging.InvalidRatingMessage
		called := false
		messaging.InvalidRatingMessage = func(userId string) {
			called = true
		}
		defer func() { messaging.InvalidRatingMessage = origInvalid }()

		messaging.RouteMessage("user1", "", "INVALID:testuser")
		
		if !called {
			t.Error("Should trigger invalid rating message")
		}
	})
}
