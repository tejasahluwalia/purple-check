package webhook

import (
	"log"
	"net/http"
	"text/template"

	"purple-check/internal/config"
)

var API_HOST = "graph.instagram.com"

func VerifyInstagramHook(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("hub.mode")
	challenge := r.URL.Query().Get("hub.challenge")
	verify_token := r.URL.Query().Get("hub.verify_token")

	if mode == "subscribe" && verify_token == config.WEBHOOK_VERIFY_TOKEN {
		log.Println("WebhookInstagram Verified")
		tmpl := template.Must(template.New("challenge").Parse("{{.}}"))   /* import ( "html/template" ) */
		tmpl.Execute(w, challenge)
		return
	}
}

func subscribeAccountToWebhooks(userId string) {
	url := "https://" + API_HOST + "/v21.0/" + userId + "/subscribed_apps"
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		log.Println(err)
	}

	q := req.URL.Query()
	q.Add("access_token", config.ACCOUNT_TOKEN)
	q.Add("subscribed_fields", "messages,messaging_postbacks")
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}

	if resp.StatusCode == 200 {
		log.Println("Subscribed to webhooks.")
	}
	defer resp.Body.Close()
}

func SetupWebhooks(w http.ResponseWriter, r *http.Request) {
	subscribeAccountToWebhooks(config.ACCOUNT_ID)
}
