package webhook

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"purple-check/internal/config"
)

type Message struct {
	Sender struct {
		Id string `json:"id"`
	} `json:"sender"`
	Recipient struct {
		Id string `json:"id"`
	} `json:"recipient"`
	Timestamp int `json:"timestamp"`
	Message   struct {
		Mid  string `json:"mid"`
		Text string `json:"text"`
	} `json:"message"`
	Is_echo        bool `json:"is_echo"`
	Is_deleted     bool `json:"is_deleted"`
	Is_unsupported bool `json:"is_unsupported"`
	Attachments    []struct {
		Type    string `json:"type"`
		Payload struct {
			url string
		} `json:"payload"`
	} `json:"attachments"`
	Quick_reply struct {
		Payload string `json:"payload"`
	} `json:"quick_reply"`
	Referral struct {
		Ref    string `json:"ref"`
		Source string `json:"source"`
		Type   string `json:"type"`
	} `json:"referral"`
	Reply_to struct {
		Mid   string `json:"mid"`
		Story struct {
			Id  string `json:"id"`
			Url string `json:"url"`
		} `json:"story"`
	} `json:"reply_to"`
	Reaction struct {
		Mid      string `json:"mid"`
		Action   string `json:"action"`
		Reaction string `json:"reaction"`
		Emoji    string `json:"emoji"`
	} `json:"reaction"`
	Postback struct {
		Mid     string `json:"mid"`
		Title   string `json:"title"`
		Payload string `json:"payload"`
	} `json:"postback"`
	Read struct {
		Mid string `json:"mid"`
	} `json:"read"`
}

type InstagramWebhook struct {
	Object string `json:"object"`
	Entry  []struct {
		Id        string `json:"id"`
		Time      int    `json:"time"`
		Messaging []Message
	} `json:"entry"`
}

var api_host = "graph.instagram.com"

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

func Instagram(w http.ResponseWriter, r *http.Request) {

	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}

	log.Println(string(b))
	
	var webhook InstagramWebhook
	err = json.NewDecoder(r.Body).Decode(&webhook)
	if err != nil {
		log.Println(err)
	}

	log.Println(webhook)

	w.WriteHeader(http.StatusOK)
}

func subscribeAccountToWebhooks(user_id string) {
	url := "https://" + api_host + "/v20.0/" + user_id + "/subscribed_apps"
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Println(err)
	}

	q := req.URL.Query()
	q.Add("access_token", config.ACCOUNT_TOKEN)
	q.Add("subscribed_fields", "message_reactions,messages,messaging_optins,messaging_postbacks,messaging_referral,messaging_seen")
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

}

func SetupWebhooks(w http.ResponseWriter, r *http.Request) {
	subscribeAccountToWebhooks(config.ACCOUNT_ID)
}
