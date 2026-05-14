package webhook

import (
	"encoding/json"
	"log"
	"net/http"
	"purple-check/internal/models"
)

func Instagram(w http.ResponseWriter, r *http.Request) {
	var webhook models.InstagramWebhook

	err := json.NewDecoder(r.Body).Decode(&webhook)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid webhook payload", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	for _, entry := range webhook.Entry {
		for _, messageEvent := range entry.Messaging {
			if !shouldRouteMessageEvent(messageEvent) {
				continue
			}
			// messaging.RouteMessage(r.Context(), messageEvent)
		}
	}
}

func shouldRouteMessageEvent(messageEvent models.MessageEvent) bool {
	if messageEvent.Sender.Id == "" || messageEvent.Sender.Id == "954039343027729" {
		return false
	}
	if messageEvent.Message.Is_echo || messageEvent.Message.Is_deleted || messageEvent.Message.Is_unsupported {
		return false
	}
	if messageEvent.Message.Text != "" || messageEvent.Message.Quick_reply.Payload != "" {
		return true
	}
	if messageEvent.Postback != nil && messageEvent.Postback.Payload != "" {
		return true
	}
	if messageEvent.Referral != nil || messageEvent.Message.Referral != nil {
		return true
	}
	return messageEvent.Postback != nil && messageEvent.Postback.Referral != nil
}
