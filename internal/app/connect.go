package app

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	_ "github.com/tursodatabase/libsql-client-go/libsql"

	"purple-check/internal/config"
)

type AccessTokenErrorResponse struct {
	ErrorType string `json:"error_type"`
	Code	  uint `json:"code"`
	ErrorMsg  string `json:"error_message"`
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	UserID	  uint `json:"user_id"`
}

type LongLivedAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType	  uint `json:"token_type"`
	ExpiresIn	  uint `json:"expires_in"`
}

type UserNode struct {
	ID       string
	Username string
}

type ErrorResponse struct {
	ErrorType string `json:"type"`
	Code	  uint `json:"code"`
	ErrorMsg  string `json:"message"`
	ErrorSubcode uint `json:"error_subcode"`
	FbtraceID string `json:"fbtrace_id"`
} 

func getAccessToken(code string) (userID uint, accessToken string, expiresIn uint) {
	formData := url.Values{}
	formData.Set("client_id", config.CLIENT_ID)
	formData.Set("client_secret", config.CLIENT_SECRET)
	formData.Set("grant_type", "authorization_code")
	formData.Set("redirect_uri", "https://www.purple-check.org/connect")
	formData.Set("code", code)

	resp, err := http.PostForm("https://api.instagram.com/oauth/access_token", formData)
	if err != nil {
		slog.Error("Error while getting access token")
	}
	if resp.StatusCode != 200 {
		var ErrorResponse AccessTokenErrorResponse
		json.NewDecoder(resp.Body).Decode(&ErrorResponse)
		slog.Error("Error while getting access token: ", "error_code",ErrorResponse.Code,
			"error_type", ErrorResponse.ErrorType,
			"error_message", ErrorResponse.ErrorMsg)
		return
	}
	var res AccessTokenResponse
	json.NewDecoder(resp.Body).Decode(&res)
	resp, err = http.Get("https://graph.instagram.com/access_token?grant_type=ig_exchange_token&client_secret=" + config.CLIENT_SECRET + "&access_token=" + res.AccessToken)
	if err != nil {
		slog.Error("Error while getting long lived access token")
		return
	}
	if resp.StatusCode != 200 {
		slog.Error("Error while getting long lived access token")
		return
	}
	var res2 LongLivedAccessTokenResponse
	json.NewDecoder(resp.Body).Decode(&res2)
	return res.UserID, res2.AccessToken, res2.ExpiresIn
}

func Connect(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed." ,http.StatusMethodNotAllowed)
		return
	}
	error := r.URL.Query().Get("error")
	if error != "" {
		slog.Error("Instagram connection error", 
			"error_description", r.URL.Query().Get("error_description"),
			"error_reason", r.URL.Query().Get("error_reason"),
			"error", r.URL.Query().Get("error"))
		http.Redirect(w, r, "/connect-account", http.StatusSeeOther)
		return
	}
	authCode := r.URL.Query().Get("code")
	if authCode == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	redirectToProfile := r.URL.Query().Get("state")
	userId, accessToken, expiresIn := getAccessToken(authCode)

	if userId == 0 || accessToken == "" || expiresIn == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var userNode UserNode
	resp, err := http.Get("https://graph.instagram.com/me?fields=id,username&access_token=" + accessToken)
	if err != nil {
		slog.Error("Error while getting user node: ", "error_type", "request_error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if resp.StatusCode != 200 {
		var errorResponse ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		slog.Error("Error while getting user node: ",  "error_type", errorResponse.ErrorType, "error_message", errorResponse.ErrorMsg,
			"error_code", errorResponse.Code, "error_subcode", errorResponse.ErrorSubcode, "fbtrace_id", error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	json.NewDecoder(resp.Body).Decode(&userNode)

	db, err := sql.Open("libsql", config.DB_PATH)
	if err != nil {
		slog.Error("Error while connecting to db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO profiles(platform, platform_user_id, username, status, token) VALUES(?, ?, ?, ?, ?) ON CONFLICT(platform, platform_user_id) DO UPDATE SET username=excluded.username, status=excluded.status, token=excluded.token")
	if err != nil {
		slog.Error("Error while preparing statement", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = stmt.Exec("instagram", userNode.ID, userNode.Username, "connected", accessToken)
	if err != nil {
		slog.Error("Error while executing statement", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	
	http.SetCookie(w, &http.Cookie{Name: "access_token", Value: accessToken, HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode, MaxAge: int(expiresIn)})
	http.SetCookie(w, &http.Cookie{Name: "platform_user_id", Value: strconv.Itoa(int(userId)), HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode, MaxAge: int(expiresIn)})
	
	if redirectToProfile == "" {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	} else {
		http.Redirect(w, r, "/profile/"+redirectToProfile, http.StatusTemporaryRedirect)
		return
	}
	
}