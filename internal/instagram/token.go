package instagram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"purple-check/internal/config"
	"purple-check/internal/models"
)

type TokenStore struct {
	ttl time.Duration

	mu        sync.Mutex
	token     string
	expiresAt time.Time

	GetUpstreamAccountToken func(ctx context.Context) (string, error)
	SetUpstreamAccountToken func(ctx context.Context, token string) error
}

func NewTokenStore(appSettings *models.AppSettingModel, ttl time.Duration) *TokenStore {
	return &TokenStore{
		ttl:                     ttl,
		GetUpstreamAccountToken: appSettings.GetAccountToken,
		SetUpstreamAccountToken: appSettings.SetAccountToken,
	}
}

func (s *TokenStore) GetAccountToken(ctx context.Context) (string, error) {
	if s == nil {
		return "", fmt.Errorf("cached token store is nil")
	}

	now := time.Now()
	s.mu.Lock()
	if s.token != "" && now.Before(s.expiresAt) {
		token := s.token
		s.mu.Unlock()
		return token, nil
	}
	s.mu.Unlock()

	if s.GetUpstreamAccountToken == nil {
		return "", fmt.Errorf("upstream token store is nil")
	}
	token, err := s.GetUpstreamAccountToken(ctx)
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	s.token = token
	s.expiresAt = now.Add(s.ttl)
	s.mu.Unlock()

	return token, nil
}

func (s *TokenStore) SetAccountToken(ctx context.Context, token string) error {
	if s == nil {
		return fmt.Errorf("cached token store is nil")
	}
	if s.SetUpstreamAccountToken == nil {
		return fmt.Errorf("upstream token store is nil")
	}
	if err := s.SetUpstreamAccountToken(ctx, token); err != nil {
		return err
	}

	s.mu.Lock()
	s.token = token
	s.expiresAt = time.Now().Add(s.ttl)
	s.mu.Unlock()

	return nil
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func NewRefreshAccessTokenHandler(tokens *TokenStore) http.Handler {
	return &RefreshAccessTokenHandler{Tokens: tokens}
}

type RefreshAccessTokenHandler struct {
	Tokens *TokenStore
}

func (h *RefreshAccessTokenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !authorized(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if h.Tokens == nil {
		slog.Error("Token store is not configured")
		http.Error(w, "Token store is not configured", http.StatusInternalServerError)
		return
	}

	currentToken, err := h.Tokens.GetAccountToken(r.Context())
	if err != nil {
		slog.Error("Failed to load current access token", "error", err)
		http.Error(w, "Failed to load current access token", http.StatusInternalServerError)
		return
	}
	if currentToken == "" {
		slog.Error("No current access token found")
		http.Error(w, "No current access token configured", http.StatusInternalServerError)
		return
	}

	refreshURL := fmt.Sprintf("https://graph.instagram.com/refresh_access_token?grant_type=ig_refresh_token&access_token=%s", currentToken)

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, refreshURL, nil)
	if err != nil {
		slog.Error("Failed to create refresh request", "error", err)
		http.Error(w, "Failed to create refresh request", http.StatusInternalServerError)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("Failed to refresh token", "error", err)
		http.Error(w, "Failed to refresh token", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		slog.Error("Instagram API error", "status", resp.StatusCode, "body", string(body))
		http.Error(w, fmt.Sprintf("Instagram API error: %d", resp.StatusCode), http.StatusBadRequest)
		return
	}

	var refreshResponse RefreshTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&refreshResponse); err != nil {
		slog.Error("Failed to decode refresh response", "error", err)
		http.Error(w, "Failed to decode response", http.StatusInternalServerError)
		return
	}

	if err := h.Tokens.SetAccountToken(r.Context(), refreshResponse.AccessToken); err != nil {
		slog.Error("Failed to persist refreshed token", "error", err)
		http.Error(w, "Failed to persist refreshed token", http.StatusInternalServerError)
		return
	}

	slog.Info("Access token refreshed successfully",
		"expires_in", refreshResponse.ExpiresIn,
		"token_type", refreshResponse.TokenType)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success":    true,
		"expires_in": refreshResponse.ExpiresIn,
		"message":    "Token refreshed successfully",
	})
}

func authorized(r *http.Request) bool {
	if config.ADMIN_TOKEN == "" {
		slog.Error("ADMIN_TOKEN is not configured")
		return false
	}
	return r.Header.Get("Authorization") == "Bearer "+config.ADMIN_TOKEN
}
