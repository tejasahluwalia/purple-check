package webhook

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"purple-check/internal/config"
	"purple-check/internal/messaging"
)

var API_HOST = "graph.instagram.com"

var webhookHTTPClient = &http.Client{Timeout: 10 * time.Second}

func VerifyInstagramHook(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("hub.mode")
	challenge := r.URL.Query().Get("hub.challenge")
	verify_token := r.URL.Query().Get("hub.verify_token")

	if mode == "subscribe" && verify_token == config.WEBHOOK_VERIFY_TOKEN {
		log.Println("WebhookInstagram Verified")
		w.Write([]byte(challenge))
		return
	}
}

func subscribeAccountToWebhooks(ctx context.Context, userId string, tokens messaging.AccountTokenReader) error {
	if tokens == nil {
		return fmt.Errorf("token store is nil")
	}
	token, err := tokens.GetAccountToken(ctx)
	if err != nil {
		return fmt.Errorf("get account token: %w", err)
	}
	if token == "" {
		return fmt.Errorf("account token is empty")
	}

	url := "https://" + API_HOST + "/" + config.INSTAGRAM_API_VERSION + "/" + userId + "/subscribed_apps"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("create subscribe request: %w", err)
	}

	q := req.URL.Query()
	q.Add("access_token", token)
	q.Add("subscribed_fields", "messages,messaging_postbacks")
	req.URL.RawQuery = q.Encode()

	resp, err := webhookHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("subscribe request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("subscribe request returned status %d", resp.StatusCode)
	}

	log.Println("Subscribed to webhooks.")
	return nil
}

func NewSetupHandler(tokens messaging.AccountTokenReader) http.Handler {
	return &SetupHandler{Tokens: tokens}
}

type SetupHandler struct {
	Tokens messaging.AccountTokenReader
}

func (h *SetupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := subscribeAccountToWebhooks(r.Context(), config.ACCOUNT_ID, h.Tokens); err != nil {
		log.Println(err)
		http.Error(w, "Failed to subscribe account to webhooks", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
