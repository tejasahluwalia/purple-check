package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

func RefreshAccessToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("access_token")
	if err != nil {
		log.Fatal("Error while getting access token: ", err)
	}

	formData := url.Values{}
	formData.Set("grant_type", "ig_refresh_token")
	formData.Set("access_token", cookie.Value)

	resp, err := http.Get("https://graph.instagram.com/refresh_access_token?" + formData.Encode())
	if err != nil {
		log.Fatal("Error while refreshing access token: ", err)
	}

	if resp.StatusCode != 200 {
		log.Fatal("Error while refreshing access token: ", resp.Body)
	}

	type RefreshAccessTokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType	  uint `json:"token_type"`
		ExpiresIn	  uint `json:"expires_in"`
	}

	var res RefreshAccessTokenResponse

	json.NewDecoder(resp.Body).Decode(&res)

	cookieAccessToken := http.Cookie{Name: "access_token", Value: res.AccessToken}
	http.SetCookie(w, &cookieAccessToken)

	cookieExpiresIn := http.Cookie{Name: "expires_in", Value: strconv.Itoa(int(res.ExpiresIn))}
	http.SetCookie(w, &cookieExpiresIn)

	defer resp.Body.Close()
	http.Redirect(w, r, "/", http.StatusFound)
}