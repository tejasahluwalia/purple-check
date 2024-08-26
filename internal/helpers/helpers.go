package helpers

import (
	"log"
	"net/http"
	"strings"

	"purple-check/internal/db"
	"purple-check/internal/models"
)

func IsLoggedIn(r *http.Request) bool {
	cookie_platform_user_id, err := r.Cookie("platform_user_id")
	if err != nil {
		return false
	}
	return cookie_platform_user_id.Value != ""
}

func GetCurrUser(r *http.Request) *models.Profile {
	cookie_platform_user_id, err := r.Cookie("platform_user_id")
	if err != nil {
		return nil
	}

	cookie_access_token, err := r.Cookie("access_token")
	if err != nil {
		return nil
	}

	db, closer := db.GetDB()
	defer closer()
	

	stmt, err := db.Prepare("SELECT id, platform, platform_user_id, username, status, token FROM profiles WHERE platform = 'instagram' AND platform_user_id = ? AND token = ?")
	if err != nil {
		log.Println(err)
	}

	var profile models.Profile
	err = stmt.QueryRow(cookie_platform_user_id.Value, cookie_access_token.Value).Scan(&profile.ID, &profile.Platform, &profile.PlatformUserID, &profile.Username, &profile.Status, &profile.Token)
	if err != nil {
		log.Println(err)
	}

	return &profile
}

func GetProfile(r *http.Request) *models.Profile {
	username := r.PathValue("username")
	if username == "" {
		return nil
	}
	username = strings.ToLower(username)
	
	db, closer := db.GetDB()
	defer closer()

	stmt, err := db.Prepare("SELECT id, platform, platform_user_id, username, status, token FROM profiles WHERE platform = 'instagram' AND username = ?")
	if err != nil {
		log.Println(err)
	} 

	var profile models.Profile

	err = stmt.QueryRow(username).Scan(&profile.ID, &profile.Platform, &profile.PlatformUserID, &profile.Username, &profile.Status, &profile.Token)
	if err != nil {
		stmt, err := db.Prepare("INSERT INTO profiles(platform, platform_user_id, username, status, token) VALUES(?, ?, ?, ?, ?);")
		if err != nil {
			log.Println(err)
		}

		_, err = stmt.Exec("instagram", nil, username, "not-connected", nil)
		if err != nil {
			log.Println(err)
		}
		
		err = stmt.QueryRow(username).Scan(&profile.ID, &profile.Platform, &profile.PlatformUserID, &profile.Username, &profile.Status, &profile.Token)
		if err != nil {
			log.Println(err)
		}
	}

	return &profile
}

func IsProfileCurrUser(r *http.Request, profile_platform_user models.Profile) bool {
	cookie_platform_user_id, err := r.Cookie("platform_user_id")
	if err != nil {
		return false
	}
	cookie_platform_username, err := r.Cookie("platform_username")
	if err != nil {
		return false
	}

	return cookie_platform_user_id.Value == profile_platform_user.PlatformUserID.String || cookie_platform_username.Value == profile_platform_user.Username
}