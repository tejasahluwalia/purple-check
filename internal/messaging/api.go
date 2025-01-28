package messaging

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"purple-check/internal/config"
	"purple-check/internal/database"
)

var API_HOST = "graph.instagram.com"
var API_VERSION = "v22.0"
var API_URL = "https://" + API_HOST + "/" + API_VERSION

func sendButtonMessage(buttons []ElementButton, text string, userId string) {
	body, err := json.Marshal(MessageRequestBody[MessageButtons]{
		MessageRecipient{
			ID: userId,
		},
		MessageButtons{
			Attachment: MessageAttachment{
				Type: "template",
				Payload: AttachmentPayload{
					TemplateType: "button",
					Text:         text,
					Buttons:      buttons,
				},
			},
		},
	})
	if err != nil {
		log.Fatal("Unable to marshall message body.")
		return
	}

	sendMessage(body)
}

func sendTextMessage(text string, userId string) {
	body, err := json.Marshal(MessageRequestBody[MessageText]{
		MessageRecipient{
			ID: userId,
		},
		MessageText{
			Text: text,
		},
	})
	if err != nil {
		log.Fatal("Unable to marshall message body.")
		return
	}

	sendMessage(body)
}

func sendMessage(body []byte) {
	url := API_URL + "/me/messages"

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		log.Println(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.ACCOUNT_TOKEN)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		log.Println(bodyString)
	}

	defer resp.Body.Close()
}

func saveRating(rating string, giverUsername string, recieverUsername string, giverRole string, receiverRole string, dealStage string) {
	db, closer := database.GetDB()
	defer closer()

	stmt, err := db.Prepare(`INSERT INTO feedback
		(giver, receiver, rating, giver_role, receiver_role, deal_stage)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(giver, receiver)
		DO UPDATE SET
			rating=excluded.rating,
			giver_role=excluded.giver_role,
			deal_stage=excluded.deal_stage`)
	if err != nil {
		log.Fatal("Error preparing statement.")
	}

	_, err = stmt.Exec(giverUsername, recieverUsername, rating, giverRole, receiverRole, dealStage)
	if err != nil {
		log.Fatal("Error executing statement.", err)
	}
}

type UserProfileAPIResponse struct {
	Username string `json:"username"`
	ID       string `json:"id"`
}

func getUsernameFromUserID(userId string) (string, error) {
	userState := getUserConversationState(userId)
	if userState.CurrentUser != "" {
		return userState.CurrentUser, nil
	}

	url := API_URL + "/" + userId
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Println(err)
		return "", err
	}

	q := req.URL.Query()
	q.Add("access_token", config.ACCOUNT_TOKEN)
	q.Add("fields", "username")
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return "", err
	}

	var userProfileAPIResponse UserProfileAPIResponse

	if resp.StatusCode != 200 {
		log.Fatal("Error getting username.")
		return "", errors.New("IG_API_Error")
	}

	err = json.NewDecoder(resp.Body).Decode(&userProfileAPIResponse)
	if err != nil {
		log.Fatal("Error decoding response body.")
	}

	userState.CurrentUser = userProfileAPIResponse.Username

	defer resp.Body.Close()
	return userProfileAPIResponse.Username, nil
}

func SetPersistentMenu() {
	url := API_URL + "/" + config.ACCOUNT_ID + "/messenger_profile"

	body, err := json.Marshal(MessengerProfileRequestBody{
		Platform: "instagram",
		PersistentMenu: []PersistentMenu{
			{
				Locale:                "default",
				ComposerInputDisabled: false,
				CallToActions: []PersistentMenuCallToAction{
					{
						Title:   "Search for a user",
						Type:    "postback",
						Payload: "SEARCH",
					}, {
						Title:   "Rate a user",
						Type:    "postback",
						Payload: "RATE",
					},
				},
			},
		},
		IceBreakers: []IceBreakers{
			{
				Locale: "default",
				CallToActions: []IceBreakerCallToAction{
					{
						Question: "Search for a user",
						Payload:  "SEARCH",
					}, {
						Question: "Rate a user",
						Payload:  "RATE",
					},
				},
			},
		},
	})
	if err != nil {
		log.Fatal("Unable to marshall message body.")
		return
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		log.Println(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.ACCOUNT_TOKEN)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()
}
