package webhook

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"purple-check/internal/messaging"
	"purple-check/internal/models"
)

type MessageRouter interface {
	RouteMessage(ctx context.Context, messageEvent models.MessageEvent)
}

func NewInstagramHandler(router MessageRouter) http.Handler {
	return &InstagramHandler{Router: router}
}

type InstagramHandler struct {
	Router MessageRouter
}

func Instagram(w http.ResponseWriter, r *http.Request) {
	NewInstagramHandler(nil).ServeHTTP(w, r)
}

func (h *InstagramHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var webhook models.InstagramWebhook

	err := json.NewDecoder(r.Body).Decode(&webhook)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid webhook payload", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	ctx := context.WithoutCancel(r.Context())
	for _, entry := range webhook.Entry {
		for _, messageEvent := range entry.Messaging {
			if !shouldRouteMessageEvent(messageEvent) {
				continue
			}
			if h.Router != nil {
				h.Router.RouteMessage(ctx, messageEvent)
			}
		}
	}
}

var _ MessageRouter = (*messaging.Router)(nil)

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
