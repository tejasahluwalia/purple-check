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

type Sender interface {
	SendTextMessage(ctx context.Context, text string, userID string) error
	SendButtonMessage(ctx context.Context, buttons []ElementButton, text string, userID string) error
	SendSenderAction(ctx context.Context, action string, userID string) error
	UsernameFromUserID(ctx context.Context, userID string) (string, error)
	UsernameFromMedia(ctx context.Context, igMediaID string) (string, error)
}

type AccountTokenReader interface {
	GetAccountToken(ctx context.Context) (string, error)
}

type InstagramSender struct {
	Tokens AccountTokenReader
}

func (sender InstagramSender) UsernameFromUserID(ctx context.Context, userID string) (string, error) {
	return getUsernameFromUserID(ctx, sender.Tokens, userID)
}

func (sender InstagramSender) UsernameFromMedia(ctx context.Context, igMediaID string) (string, error) {
	igMedia, err := getIGMedia(ctx, sender.Tokens, igMediaID)
	if err != nil {
		return "", err
	}
	log.Print(igMedia)
	return igMedia.Username, nil
}

func (sender InstagramSender) SendButtonMessage(ctx context.Context, buttons []ElementButton, text string, userID string) error {
	body, err := json.Marshal(MessageRequestBody[MessageButtons]{
		MessageRecipient{
			ID: userID,
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

	return sendMessage(ctx, sender.Tokens, body)
}

func (sender InstagramSender) SendTextMessage(ctx context.Context, text string, userID string) error {
	body, err := json.Marshal(MessageRequestBody[MessageText]{
		MessageRecipient{
			ID: userID,
		},
		MessageText{
			Text: text,
		},
	})
	if err != nil {
		return fmt.Errorf("marshal text message: %w", err)
	}

	return sendMessage(ctx, sender.Tokens, body)
}

func (sender InstagramSender) SendSenderAction(ctx context.Context, action string, userID string) error {
	body, err := json.Marshal(SenderActionRequest{
		Recipient: MessageRecipient{
			ID: userID,
		},
		SenderAction: action,
	})
	if err != nil {
		return fmt.Errorf("marshal sender action: %w", err)
	}
	return sendMessage(ctx, sender.Tokens, body)
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

type IGOwner struct {
	ID string `json:"id"`
}

type IGMedia struct {
	// ID            string  `json:"id"`
	// MediaType     string  `json:"media_type"`
	// MediaURL      string  `json:"media_url"`
	// Owner         IGOwner `json:"owner"`
	// Timestamp     string  `json:"timestamp"`
	Username string `json:"username"`
	// Caption       string  `json:"caption"`
	// CommentsCount int     `json:"comments_count"`
	// LikeCount     int     `json:"like_count"`
	// Permalink     string  `json:"permalink"`
	// Shortcode     string  `json:"shortcode"`
	// ThumbnailURL  string  `json:"thumbnail_url"`
}

func getIGMedia(ctx context.Context, tokens AccountTokenReader, igMediaID string) (IGMedia, error) {
	token, err := accountToken(ctx, tokens)
	if err != nil {
		return IGMedia{}, err
	}

	url := API_URL + "/" + igMediaID
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return IGMedia{}, fmt.Errorf("create ig media request: %w", err)
	}

	q := req.URL.Query()
	q.Add("access_token", token)
	q.Add("fields", "username")
	req.URL.RawQuery = q.Encode()

	log.Print(req)
	resp, err := httpClient.Do(req)
	if err != nil {
		return IGMedia{}, fmt.Errorf("do ig media request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return IGMedia{}, fmt.Errorf("ig media API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var media IGMedia
	if err := json.NewDecoder(resp.Body).Decode(&media); err != nil {
		return IGMedia{}, fmt.Errorf("decode ig media response: %w", err)
	}

	return media, nil
}
