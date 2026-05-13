package webhook

import (
	"log"
	"net/http"
	"time"

	"purple-check/internal/config"
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

func subscribeAccountToWebhooks(userId string) {
	url := "https://" + API_HOST + "/v21.0/" + userId + "/subscribed_apps"
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		log.Println(err)
		return
	}

	q := req.URL.Query()
	q.Add("access_token", config.ACCOUNT_TOKEN)
	q.Add("subscribed_fields", "messages,messaging_postbacks")
	req.URL.RawQuery = q.Encode()

	resp, err := webhookHTTPClient.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		log.Println("Subscribed to webhooks.")
	}
}

func SetupWebhooks(w http.ResponseWriter, r *http.Request) {
	subscribeAccountToWebhooks(config.ACCOUNT_ID)
}
