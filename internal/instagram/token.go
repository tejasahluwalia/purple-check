package instagram

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"purple-check/internal/config"
)

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func RefreshAccessToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentToken := config.ACCOUNT_TOKEN
	if currentToken == "" {
		slog.Error("No current access token found")
		http.Error(w, "No current access token configured", http.StatusInternalServerError)
		return
	}

	// Make request to Instagram API
	refreshURL := fmt.Sprintf("https://graph.instagram.com/refresh_access_token?grant_type=ig_refresh_token&access_token=%s", currentToken)
	
	resp, err := http.Get(refreshURL)
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

	// Update the global config variable
	config.UpdateAccountToken(refreshResponse.AccessToken)

	// Write new token to environment (optional - depends on your setup)
	// You might want to update a .env file or use another persistence method
	
	slog.Info("Access token refreshed successfully", 
		"expires_in", refreshResponse.ExpiresIn,
		"token_type", refreshResponse.TokenType)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"expires_in": refreshResponse.ExpiresIn,
		"message":    "Token refreshed successfully",
	})
}