package messaging

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"purple-check/internal/config"
	"purple-check/internal/database"
	"purple-check/internal/helpers"
)

var API_HOST = "graph.instagram.com"

func sendButtonMessage(buttons []ElementButton, text string, userId string) {
	body, err := json.Marshal(MessageRequestBody[MessageButtons]{
		MessageRecipient{
			ID: userId,
		},
		MessageButtons{
			Attachment: MessageAttachment{
				Type: "template",
				Payload: AttachmentPayload{
					TemplateType: "generic",
					Elements: []PayloadElements{
						{
							Title:   text,
							Buttons: buttons,
						},
					},
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
	url := "https://" + API_HOST + "/v21.0/me/messages"

	reqBodyString := string(body)
	log.Println(reqBodyString)

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
	return
}

func saveRating(rating int, giverUserId string, recieverUsername string) {
	userProfileAPIResponse, err := getUsernameFromUserID(giverUserId)

	db, closer := database.GetDB()
	defer closer()

	stmt, err := db.Prepare("INSERT INTO feedback (giver, receiver, rating) VALUES (?, ?, ?)")
	if err != nil {
		log.Fatal("Error preparing statement.")
	}

	_, err = stmt.Exec(userProfileAPIResponse.Username, strings.Split(recieverUsername, "@")[1], rating)
	if err != nil {
		log.Fatal("Error executing statement.", err)
	}

	return
}

type UserProfileAPIResponse struct {
	Username string `json:"username"`
	ID       string `json:"id"`
}

func getUsernameFromUserID(userId string) (*UserProfileAPIResponse, error) {
	url := "https://" + API_HOST + "/v21.0/" + userId
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	q := req.URL.Query()
	q.Add("access_token", config.ACCOUNT_TOKEN)
	q.Add("fields", "username")
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var userProfileAPIResponse UserProfileAPIResponse

	if resp.StatusCode != 200 {
		log.Fatal("Error getting username.")
		return nil, errors.New("IG_API_Error")
	}

	err = json.NewDecoder(resp.Body).Decode(&userProfileAPIResponse)
	if err != nil {
		log.Fatal("Error decoding response body.")
	}

	defer resp.Body.Close()
	return &userProfileAPIResponse, nil
}

func SetPersistentMenu() {
	url := "https://" + API_HOST + "/v21.0/" + config.ACCOUNT_ID + "/messenger_profile"

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

	helpers.PrintReqBody(req)

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

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	log.Println(bodyString)

	defer resp.Body.Close()
	return
}
