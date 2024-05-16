package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"

	_ "github.com/mattn/go-sqlite3"

	"github.com/tejasahluwalia/purple-check/constants"
	"github.com/tejasahluwalia/purple-check/models"
)

func HandleConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Redirect(w, r, "/", http.StatusMethodNotAllowed)
		return
	}
	authCode := r.URL.Query().Get("code")
	if authCode == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	} else {
		redirectToProfile := r.URL.Query().Get("state")

		formData := url.Values{}
		formData.Set("client_id", constants.CLIENT_ID)
		formData.Set("client_secret", constants.CLIENT_SECRET)
		formData.Set("grant_type", "authorization_code")
		formData.Set("redirect_uri", "https://www.purple-check.org/connect")
		formData.Set("code", authCode)

		type ErrorResponse struct {
			ErrorType string `json:"error_type"`
			Code	  uint `json:"code"`
			ErrorMsg  string `json:"error_message"`
		}

		resp, err := http.PostForm("https://api.instagram.com/oauth/access_token", formData)
		if err != nil {
			log.Println("Error while getting access token: ", err.Error())
		}

		if resp.StatusCode != 200 {
			var ErrorResponse ErrorResponse
			json.NewDecoder(resp.Body).Decode(&ErrorResponse)
			log.Println("Error while getting access token: ", ErrorResponse)
		}

		type AccessTokenResponse struct {
			AccessToken string `json:"access_token"`
			UserID	  uint `json:"user_id"`
		}

    	var res AccessTokenResponse

		json.NewDecoder(resp.Body).Decode(&res)

		resp, err = http.Get("https://graph.instagram.com/access_token?grant_type=ig_exchange_token&client_secret=" + constants.CLIENT_SECRET + "&access_token=" + res.AccessToken)
		if err != nil {
			log.Println("Error while getting long lived access token: ", resp.Body)
		}

		if resp.StatusCode != 200 {
			log.Println("Error while getting long lived access token: ", resp.Body)
		}

		type LongLivedAccessTokenResponse struct {
			AccessToken string `json:"access_token"`
			TokenType	  uint `json:"token_type"`
			ExpiresIn	  uint `json:"expires_in"`
		}

    	var res2 LongLivedAccessTokenResponse

		json.NewDecoder(resp.Body).Decode(&res2)

		type UserNode struct {
			ID       string
			Username string
		}

		var userNode UserNode

		resp, err = http.Get("https://graph.instagram.com/me?fields=id,username&access_token=" + res2.AccessToken)
		if err != nil {
			log.Println("Error while getting user node: ", err)
		}

		if resp.StatusCode != 200 {
			log.Println("Error while getting user node: ", resp.Body)
		}

		json.NewDecoder(resp.Body).Decode(&userNode)

		db, err := sql.Open("sqlite3", "db/purple-check.db")

		if err != nil {
			log.Println(err)
		}

		defer db.Close()

		stmt, err := db.Prepare("SELECT id, platform, platform_user_id, username, status, token FROM profiles WHERE platform = ? AND username = ? OR platform_user_id = ?")
		if err != nil {
			log.Println(err)
		}

		var currUser models.Profile
		err = stmt.QueryRow("instagram", userNode.Username, userNode.ID).Scan(&currUser.ID, &currUser.Platform, &currUser.PlatformUserID, &currUser.Username, &currUser.Status, &currUser.Token)

		if err != nil {
			log.Println("Current user not found: ", err)
			stmt, err = db.Prepare("INSERT INTO profiles(platform, platform_user_id, username, status, token, expires_in) VALUES(?, ?, ?, ?, ?, ?)")
			if err != nil {
				log.Println(err)
			}
			_, err = stmt.Exec("instagram", userNode.ID, userNode.Username, "connected", res2.AccessToken, res2.ExpiresIn)
			if err != nil {
				log.Println(err)
			}
		} else {
			log.Println("Current user found in db: ", currUser)
			stmt, err = db.Prepare("UPDATE profiles SET status = 'connected', token = ?, platform_user_id = ?, expires_in = ? WHERE platform = ? AND username = ?")
			if err != nil {
				log.Println(err)
			}
			_, err = stmt.Exec(res2.AccessToken, res.UserID, res2.ExpiresIn, "instagram", userNode.Username)
			if err != nil {
				log.Println(err)
			}
		}

		cookie := http.Cookie{Name: "access_token", Value: res2.AccessToken, HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode, MaxAge: int(res2.ExpiresIn)}
		http.SetCookie(w, &cookie)

		cookie = http.Cookie{Name: "platform_user_id", Value: strconv.Itoa(int(res.UserID)), HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode, MaxAge: int(res2.ExpiresIn)}
		http.SetCookie(w, &cookie)

		cookie = http.Cookie{Name: "expires_in", Value: strconv.Itoa(int(res2.ExpiresIn)), HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode, MaxAge: int(res2.ExpiresIn)}
		http.SetCookie(w, &cookie)

		defer resp.Body.Close()

		if redirectToProfile == "" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		} else {
			http.Redirect(w, r, "/profile/"+redirectToProfile, http.StatusFound)
			return
		}
	}
}