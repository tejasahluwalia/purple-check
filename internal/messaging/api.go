package messaging

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"time"

	"purple-check/internal/config"
)

var API_HOST = "graph.instagram.com"
var API_VERSION = config.INSTAGRAM_API_VERSION
var API_URL = "https://" + API_HOST + "/" + API_VERSION

var httpClient = &http.Client{Timeout: 10 * time.Second}

func sendButtonMessage(buttons []ElementButton, text string, userId string) error {
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
		return fmt.Errorf("marshal button message: %w", err)
	}

	return sendMessage(body)
}

func sendTextMessage(text string, userId string) error {
	body, err := json.Marshal(MessageRequestBody[MessageText]{
		MessageRecipient{
			ID: userId,
		},
		MessageText{
			Text: text,
		},
	})
	if err != nil {
		return fmt.Errorf("marshal text message: %w", err)
	}

	return sendMessage(body)
}

func sendMessage(body []byte) error {
	url := API_URL + "/me/messages"

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create message request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.ACCOUNT_TOKEN)

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send message request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("send message returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// func saveRating(ctx context.Context, rating string, giverUsername string, recieverUsername string, giverRole string, receiverRole string, dealStage string) error {
// 	db, closer, err := database.GetDB(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	defer closer()

// 	stmt, err := db.Prepare(`INSERT INTO feedback
// 		(giver, receiver, rating, giver_role, receiver_role, deal_stage)
// 		VALUES (?, ?, ?, ?, ?, ?)
// 		ON CONFLICT(giver, receiver)
// 		DO UPDATE SET
// 			rating=excluded.rating,
// 			giver_role=excluded.giver_role,
// 			receiver_role=excluded.receiver_role,
// 			deal_stage=excluded.deal_stage`)
// 	if err != nil {
// 		return fmt.Errorf("prepare save rating: %w", err)
// 	}
// 	defer stmt.Close()

// 	_, err = stmt.Exec(giverUsername, recieverUsername, rating, giverRole, receiverRole, dealStage)
// 	if err != nil {
// 		return fmt.Errorf("execute save rating: %w", err)
// 	}

// 	if err := database.PushDB(ctx); err != nil {
// 		return fmt.Errorf("push save rating: %w", err)
// 	}

// 	return nil
// }

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

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer resp.Body.Close()

	var userProfileAPIResponse UserProfileAPIResponse

	if resp.StatusCode != 200 {
		return "", errors.New("IG_API_Error: Unable to retrieve username")
	}

	err = json.NewDecoder(resp.Body).Decode(&userProfileAPIResponse)
	if err != nil {
		return "", fmt.Errorf("decode user profile response: %w", err)
	}

	userState.CurrentUser = userProfileAPIResponse.Username
	setUserConversationState(userId, userState)

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
					},
				},
			},
		},
		IceBreakers: []IceBreakers{
			{
				Locale: "default",
				CallToActions: []IceBreakerCallToAction{
					{
						Question: "Start here",
						Payload:  "SEARCH",
					},
				},
			},
		},
	})
	if err != nil {
		log.Println("Unable to marshal message body.", err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		log.Println(err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.ACCOUNT_TOKEN)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		slog.Error("Error setting persistent menu")
	} else {
		slog.Info("Persistent menu set.")
	}

}
