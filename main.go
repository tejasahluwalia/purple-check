package main

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Profile struct {
	ID       	string
	Platform 	string
	PlatformUserID 	sql.NullString
	Username 	string
	Status   	string
	Token		sql.NullString
	CreatedAt	string
	UpdatedAt	string
}

type Feedback struct {
	ID      		string
	GiverID   		string
	ReceiverID  	string
	Comment 		string
	GiverRole   	string
	ReceiverRole  	string
	CreatedAt		string
}

type ProfilePageData struct {
	CurrUser  		*Profile
	FeedbackList 	*[]Feedback
	Profile      	*Profile
	ProfileExists 	bool
	CurrUserExists 	bool
	UsernameSearch 	string
}

type FeedbackPageData struct {
	Receiver 		*Profile
	GiverExists 	bool
	Giver 			*Profile
}

var CLIENT_ID = "1845512079238812"
var CLIENT_SECRET = "b25cd8fd8234b6315478c29e23a001ce"

func renderHomepage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/layout.gohtml", "templates/index.gohtml", "templates/search.gohtml")
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
	username = strings.ToLower(username)
	if username == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	} else {
		http.Redirect(w, r, "/profile/"+username, http.StatusFound)
		return
	}
}

func renderProfile(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/layout.gohtml", "templates/connect.gohtml", "templates/profile.gohtml", "templates/search.gohtml")
	if err != nil {
		log.Fatal(err)
	}
	username := r.URL.Path[len("/profile/"):]
	username = strings.ToLower(username)
	db, err := sql.Open("sqlite3", "data/purple-check.db")

    if err != nil {
        log.Fatal(err)
    }

    defer db.Close()

	stmt, err := db.Prepare("SELECT id, platform, platform_user_id, username, status, token FROM profiles WHERE platform = ? AND username = ?")
	if err != nil {
		log.Fatal(err)
	}

	var profile Profile
	var profileExists bool
	var feedbackList []Feedback
	err = stmt.QueryRow("instagram", username).Scan(&profile.ID, &profile.Platform, &profile.PlatformUserID, &profile.Username, &profile.Status, &profile.Token)
	if err != nil {
		log.Println("Profile not found: ", err)
		profileExists = false
	} else {
		if profile.PlatformUserID.Valid {
			profileExists = true
			stmt, err = db.Prepare("SELECT id, giver_id, receiver_id, comment, giver_role, receiver_role FROM feedback WHERE receiver_id = ?")
			if err != nil {
				log.Fatal(err)
			}
			rows, err := stmt.Query(profile.ID)
			if err != nil {
				log.Fatal(err)
			}
			for rows.Next() {
				var feedback Feedback
				err = rows.Scan(&feedback.ID, &feedback.GiverID, &feedback.ReceiverID, &feedback.Comment, &feedback.GiverRole, &feedback.ReceiverRole)
				if err != nil {
					log.Fatal(err)
				}
				feedbackList = append(feedbackList, feedback)
			}
		} else {
			profileExists = false
		}
	}

	cookie, err := r.Cookie("platform_user_id")
	var currUser Profile
	var currUserExists bool
	if err != nil {
		log.Println("Current user not found: ", err)
		currUserExists = false
		var profilePageData ProfilePageData
		profilePageData.ProfileExists = profileExists
		profilePageData.Profile = &profile
		profilePageData.CurrUserExists = currUserExists
		profilePageData.UsernameSearch = username
		profilePageData.FeedbackList = &feedbackList
		t.Execute(w, profilePageData)
		return
	} 
	stmt, err = db.Prepare("SELECT id, platform, platform_user_id, username, status, token FROM profiles WHERE platform = ? AND platform_user_id = ?")
	if err != nil {
		log.Fatal(err)
	}
	err = stmt.QueryRow("instagram", cookie.Value).Scan(&currUser.ID, &currUser.Platform, &currUser.PlatformUserID, &currUser.Username, &currUser.Status, &currUser.Token)
	if err != nil {
		currUserExists = false
		log.Println("Current user not found: ", err)
	} else {
		currUserExists = true
	}

	var profilePageData ProfilePageData
	profilePageData.ProfileExists = profileExists
	profilePageData.Profile = &profile
	profilePageData.CurrUserExists = currUserExists
	profilePageData.CurrUser = &currUser
	profilePageData.UsernameSearch = username
	profilePageData.FeedbackList = &feedbackList
	t.Execute(w, profilePageData)
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

		type UserNode struct {
			ID       string
			Username string
		}

		var userNode UserNode

		resp, err = http.Get("https://graph.instagram.com/me?fields=id,username&access_token=" + res2.AccessToken)
		if err != nil {
			log.Fatal("Error while getting user node: ", err)
		}

		if resp.StatusCode != 200 {
			log.Fatal("Error while getting user node: ", resp.Body)
		}

		json.NewDecoder(resp.Body).Decode(&userNode)

		db, err := sql.Open("sqlite3", "data/purple-check.db")

		if err != nil {
			log.Fatal(err)
		}

		defer db.Close()

		stmt, err := db.Prepare("SELECT id, platform, platform_user_id, username, status, token FROM profiles WHERE platform = ? AND platform_user_id = ?")
		if err != nil {
			log.Fatal(err)
		}

		var currUser Profile
		err = stmt.QueryRow("instagram", userNode.ID).Scan(&currUser.ID, &currUser.Platform, &currUser.PlatformUserID, &currUser.Username, &currUser.Status, &currUser.Token)

		if err != nil {
			log.Println("Current user not found: ", err)
			stmt, err = db.Prepare("INSERT INTO profiles(platform, platform_user_id, username, status, token) VALUES(?, ?, ?, ?, ?)")
			if err != nil {
				log.Fatal(err)
			}
			_, err = stmt.Exec("instagram", userNode.ID, userNode.Username, "connected", res2.AccessToken)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			stmt, err = db.Prepare("UPDATE profiles SET platform, token = ? WHERE platform_user_id = ?")
			if err != nil {
				log.Fatal(err)
			}
			_, err = stmt.Exec(res2.AccessToken, res.UserID)
			if err != nil {
				log.Fatal(err)
			}
		}

		cookie := http.Cookie{Name: "access_token", Value: res2.AccessToken}
		http.SetCookie(w, &cookie)

		cookie = http.Cookie{Name: "platform_user_id", Value: strconv.Itoa(int(res.UserID))}
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

func renderFeedbackForm(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/layout.gohtml", "templates/search.gohtml", "templates/feedbackForm.gohtml")
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("sqlite3", "data/purple-check.db")

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	var feedbackPageData FeedbackPageData

	cookie, err := r.Cookie("platform_user_id")
	if err != nil {
		log.Println("Current user not found: ", err)
		feedbackPageData.GiverExists = false
		http.Redirect(w, r, "/", http.StatusNotFound)
	} else {
		stmt, err := db.Prepare("SELECT id, platform, platform_user_id, username, status, token FROM profiles WHERE platform = ? AND platform_user_id = ?")
		if err != nil {
			log.Fatal(err)
		}
		var currUser Profile
		err = stmt.QueryRow("instagram", cookie.Value).Scan(&currUser.ID, &currUser.Platform, &currUser.PlatformUserID, &currUser.Username, &currUser.Status, &currUser.Token)
		if err != nil {
			log.Println("Current user not found: ", err)
			feedbackPageData.GiverExists = false
			http.Redirect(w, r, "/404", http.StatusNotFound)
		} else {
			feedbackPageData.GiverExists = true
			feedbackPageData.Giver = &currUser
		}
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		log.Println("Username not found: ", err)
		http.Redirect(w, r, "/", http.StatusNotFound)
	} else {
		stmt, err := db.Prepare("SELECT id, platform, platform_user_id, username, status, token FROM profiles WHERE platform = ? AND username = ?")
		if err != nil {
			log.Fatal(err)
		}
		var profile Profile
		err = stmt.QueryRow("instagram", username).Scan(&profile.ID, &profile.Platform, &profile.PlatformUserID, &profile.Username, &profile.Status, &profile.Token)
		if err != nil {
			log.Println("Profile not found: ", err)
			stmt, err = db.Prepare("INSERT INTO profiles(platform, platform_user_id, username, status, token) VALUES(?, ?, ?, ?, ?)")
			if err != nil {
				log.Fatal(err)
			}
			_, err = stmt.Exec("instagram", nil, username, "not-connected", nil)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			feedbackPageData.Receiver = &profile
		}
		t.Execute(w, feedbackPageData)
	}
}

func handleSubmitFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusMethodNotAllowed)
		return
	}

	db, err := sql.Open("sqlite3", "data/purple-check.db")

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	cookie, err := r.Cookie("platform_user_id")
	if err != nil {
		log.Println("Current user not found: ", err)
		http.Redirect(w, r, "/404", http.StatusNotFound)
	} else {
		stmt, err := db.Prepare("SELECT id, platform, platform_user_id, username, status, token FROM profiles WHERE platform = ? AND platform_user_id = ?")
		if err != nil {
			log.Fatal(err)
		}
		var currUser Profile
		err = stmt.QueryRow("instagram", cookie.Value).Scan(&currUser.ID, &currUser.Platform, &currUser.PlatformUserID, &currUser.Username, &currUser.Status, &currUser.Token)
		if err != nil {
			log.Println("Current user not found: ", err)
			http.Redirect(w, r, "/404", http.StatusNotFound)
		} else {
			receiverID := r.FormValue("receiver_id")
			comment := r.FormValue("comment")
			stmt, err = db.Prepare("INSERT INTO feedback(giver_id, receiver_id, comment, giver_role, receiver_role) VALUES(?, ?, ?, ?, ?)")
			if err != nil {
				log.Fatal(err)
			}
			_, err = stmt.Exec(currUser.ID, receiverID, comment, "instagram", "instagram")
			if err != nil {
				log.Fatal(err)
			}
			http.Redirect(w, r, "/profile/"+receiverID, http.StatusFound)
		}
	}
}

func main() {
	http.HandleFunc("/", renderHomepage)
	http.HandleFunc("/search", handleSearch)
	http.HandleFunc("/profile/", renderProfile)
	http.HandleFunc("/connect", handleConnect)
	http.HandleFunc("/refresh-token", refreshAccessToken)
	http.HandleFunc("/feedback", renderFeedbackForm)
	http.HandleFunc("/submit-feedback", handleSubmitFeedback)
	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
