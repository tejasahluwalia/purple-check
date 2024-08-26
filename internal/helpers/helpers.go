package helpers

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"purple-check/internal/config"
	"purple-check/internal/models"

	_ "modernc.org/sqlite"
)

func IsLoggedIn(r *http.Request) bool {
	cookie_platform_user_id, err := r.Cookie("platform_user_id")
	if err != nil {
		return false
	}
	return cookie_platform_user_id.Value != ""
}

func GetCurrUser(r *http.Request, db *sql.DB) *models.Profile {
	cookie_platform_user_id, err := r.Cookie("platform_user_id")
	if err != nil {
		return nil
	}

	cookie_access_token, err := r.Cookie("access_token")
	if err != nil {
		return nil
	}

	if db == nil {
		db, err = sql.Open("sqlite", config.LOCAL_DB_PATH)

		if err != nil {
			log.Println(err)
		}

		defer db.Close()
	}

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
	db, err := sql.Open("sqlite", config.LOCAL_DB_PATH)

    if err != nil {
        log.Println(err)
    }

    defer db.Close()

	stmt, err := db.Prepare("INSERT OR IGNORE INTO profiles(platform, platform_user_id, username, status, token) VALUES(@platform, @platform_user_id, @username, @status, @token);")
	if err != nil {
		log.Println(err)
	}

	var profile models.Profile
	
	_, err = stmt.Exec(sql.Named("platform", "instagram"), sql.Named("platform_user_id", nil), sql.Named("username", username), sql.Named("status", "not-connected"), sql.Named("token", nil))
	if err != nil {
		log.Println(err)
	}

	stmt, err = db.Prepare("SELECT id, platform, platform_user_id, username, status, token FROM profiles WHERE platform = 'instagram' AND username = ?")
	if err != nil {
		log.Println(err)
	}

	err = stmt.QueryRow(username).Scan(&profile.ID, &profile.Platform, &profile.PlatformUserID, &profile.Username, &profile.Status, &profile.Token)
	if err != nil {
		log.Println(err)
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