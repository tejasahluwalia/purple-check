package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

type Profile struct {
	ID       string
	Username string
	Status   string
	Token	string
}

type Feedback struct {
	ID      string
	Giver   string
	Receiver  string
	UserID  string
	Comment string
}

type CurrentUser struct {
	ID       string
	Username string
}

type ProfilePageData struct {
	CurrentUser  *CurrentUser
	FeedbackList *[]Feedback
	Profile      *Profile
	ProfileExists bool
	CurrUserExists bool
	UsernameSearch string
}

var CLIENT_ID = "1845512079238812"
var CLIENT_SECRET = "b25cd8fd8234b6315478c29e23a001ce"

func renderHomepage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/layout.html", "templates/index.html")
	if err != nil {
		log.Fatal(err)
	}
	t.Execute(w, nil)
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusMethodNotAllowed)
		return
	}
	username := r.FormValue("username")
	if username == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	} else {
		http.Redirect(w, r, "/profile/"+username, http.StatusFound)
		return
	}
}

func renderProfile(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/layout.html", "templates/connect.html", "templates/profile.html")
	if err != nil {
		log.Fatal(err)
	}
	username := r.URL.Path[len("/profile/"):]
	// TODO:Get user from database
	t.Execute(w, ProfilePageData{ ProfileExists: false, CurrUserExists: false, UsernameSearch: username})
}

func handleConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Redirect(w, r, "/", http.StatusMethodNotAllowed)
		return
	}
	authCode := r.URL.Query().Get("code")
	if authCode == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	} else {
		formData := url.Values{}
		formData.Set("client_id", CLIENT_ID)
		formData.Set("client_secret", CLIENT_SECRET)
		formData.Set("grant_type", "authorization_code")
		formData.Set("redirect_uri", "https://www.purple-check.org/connect")
		formData.Set("code", authCode)

		resp, err := http.PostForm("https://api.instagram.com/oauth/access_token", formData)
		if err != nil {
			log.Fatal("Error while getting access token: ", err)
		}

		if resp.StatusCode != 200 {
			log.Fatal("Error while getting access token: ", resp.Body)
		}

		type AccessTokenResponse struct {
			AccessToken string `json:"access_token"`
			UserID	  uint `json:"user_id"`
		}

    	var res AccessTokenResponse

		json.NewDecoder(resp.Body).Decode(&res)

		resp, err = http.Get("https://graph.instagram.com/access_token?grant_type=ig_exchange_token&client_secret=" + CLIENT_SECRET + "&access_token=" + res.AccessToken)
		if err != nil {
			log.Fatal("Error while getting long lived access token: ", err)
		}

		if resp.StatusCode != 200 {
			log.Fatal("Error while getting long lived access token: ", resp.Body)
		}

		type LongLivedAccessTokenResponse struct {
			AccessToken string `json:"access_token"`
			TokenType	  uint `json:"token_type"`
			ExpiresIn	  uint `json:"expires_in"`
		}

    	var res2 LongLivedAccessTokenResponse

		json.NewDecoder(resp.Body).Decode(&res2)

		cookie := http.Cookie{Name: "access_token", Value: res2.AccessToken}
		http.SetCookie(w, &cookie)

		cookie = http.Cookie{Name: "user_id", Value: strconv.Itoa(int(res.UserID))}
		http.SetCookie(w, &cookie)

		cookie = http.Cookie{Name: "expires_in", Value: strconv.Itoa(int(res2.ExpiresIn))}
		http.SetCookie(w, &cookie)

		defer resp.Body.Close()
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
}

func refreshAccessToken(w http.ResponseWriter, r *http.Request) {
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

func main() {
	http.HandleFunc("/", renderHomepage)
	http.HandleFunc("/search", handleSearch)
	http.HandleFunc("/profile/", renderProfile)
	http.HandleFunc("/connect", handleConnect)
	http.HandleFunc("/refresh-token", refreshAccessToken)
	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
