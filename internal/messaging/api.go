package messaging

import (
	"bytes"
	"context"
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

func sendButtonMessage(ctx context.Context, tokens AccountTokenReader, buttons []ElementButton, text string, userId string) error {
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

	return sendMessage(ctx, tokens, body)
}

func sendTextMessage(ctx context.Context, tokens AccountTokenReader, text string, userId string) error {
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

	return sendMessage(ctx, tokens, body)
}

func sendMessage(ctx context.Context, tokens AccountTokenReader, body []byte) error {
	token, err := accountToken(ctx, tokens)
	if err != nil {
		return err
	}
	url := API_URL + "/me/messages"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create message request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

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

type UserProfileAPIResponse struct {
	Username string `json:"username"`
	ID       string `json:"id"`
}

func getUsernameFromUserID(ctx context.Context, tokens AccountTokenReader, userId string) (string, error) {
	userState := getUserConversationState(userId)
	if userState.CurrentUser != "" {
		return userState.CurrentUser, nil
	}

	token, err := accountToken(ctx, tokens)
	if err != nil {
		return "", err
	}

	url := API_URL + "/" + userId
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Println(err)
		return "", err
	}

	q := req.URL.Query()
	q.Add("access_token", token)
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

func SetPersistentMenu(ctx context.Context, tokens AccountTokenReader) error {
	token, err := accountToken(ctx, tokens)
	if err != nil {
		return err
	}
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
		return fmt.Errorf("marshal messenger profile body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create messenger profile request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("set persistent menu request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("set persistent menu returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	slog.Info("Persistent menu set.")
	return nil
}

func accountToken(ctx context.Context, tokens AccountTokenReader) (string, error) {
	if tokens == nil {
		return "", errors.New("account token store is nil")
	}
	token, err := tokens.GetAccountToken(ctx)
	if err != nil {
		return "", fmt.Errorf("get account token: %w", err)
	}
	if token == "" {
		return "", errors.New("account token is empty")
	}
	return token, nil
}
