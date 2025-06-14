package config

import (
	"os"
)

type Config map[string]string

func init() {
	config := make(Config)

	expected_keys := []string{"APP_ID", "WEBHOOK_VERIFY_TOKEN", "ACCOUNT_TOKEN", "ACCOUNT_ID", "TURSO_DATABASE_URL", "TURSO_AUTH_TOKEN", "LOCAL_DB_PATH", "PORT", "HOST", "DEV"}

	for _, key := range expected_keys {
		config[key] = os.Getenv(key)
		if config[key] == "" {
			if key == "DEV" {
				config[key] = "false"
			} else {
				panic("Missing key in .env file: " + key)
			}
		}
	}

	APP_ID = config["APP_ID"]
	WEBHOOK_VERIFY_TOKEN = config["WEBHOOK_VERIFY_TOKEN"]
	ACCOUNT_TOKEN = config["ACCOUNT_TOKEN"]
	ACCOUNT_ID = config["ACCOUNT_ID"]
	TURSO_DATABASE_URL = config["TURSO_DATABASE_URL"]
	TURSO_AUTH_TOKEN = config["TURSO_AUTH_TOKEN"]
	LOCAL_DB_PATH = config["LOCAL_DB_PATH"]
	HOST = config["HOST"]
	PORT = config["PORT"]
	DEV = config["DEV"] == "true"
}

var APP_ID string
var WEBHOOK_VERIFY_TOKEN string
var ACCOUNT_TOKEN string
var ACCOUNT_ID string
var TURSO_DATABASE_URL string
var TURSO_AUTH_TOKEN string
var HOST string
var PORT string
var LOCAL_DB_PATH string
var DEV bool

// UpdateAccountToken allows runtime updates to the Instagram access token
func UpdateAccountToken(newToken string) {
	ACCOUNT_TOKEN = newToken
}
