package webhook

import (
	"encoding/json"
	"log"
	"net/http"

	"purple-check/internal/messaging"
)

func Instagram(w http.ResponseWriter, r *http.Request) {
	var webhook InstagramWebhook

	err := json.NewDecoder(r.Body).Decode(&webhook)
	if err != nil {
		log.Println(err)
	}

	if len(webhook.Entry) == 0 || len(webhook.Entry[0].Messaging) == 0 || webhook.Entry[0].Messaging[0].Message.Is_echo {
		w.WriteHeader(http.StatusOK)
		return
	}


	messageEvent := webhook.Entry[0].Messaging[0]
	userId := messageEvent.Sender.Id

	if userId == "954039343027729" {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusOK)

	if messageEvent.Postback != nil {
		messaging.RouteMessage(userId, messageEvent.Message.Text, messageEvent.Postback.Payload)
	} else {
		messaging.RouteMessage(userId, messageEvent.Message.Text, "")
	}
}
