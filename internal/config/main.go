package config

import (
	"os"
)

type Config map[string]string

func init() {
	config := make(Config)

	expected_keys := []string{"CLIENT_ID", "CLIENT_SECRET", "APP_ID", "APP_SECRET", "WEBHOOK_VERIFY_TOKEN", "ACCOUNT_TOKEN", "ACCOUNT_ID", "PAGE_ACCESS_TOKEN", "TURSO_DATABASE_URL", "TURSO_AUTH_TOKEN"}

	for _, key := range expected_keys {
		config[key] = os.Getenv(key)
		if _, ok := config[key]; !ok {
			panic("Missing key in .env file: " + key)
		}
	}

	CLIENT_ID = config["CLIENT_ID"]
	CLIENT_SECRET = config["CLIENT_SECRET"]
	APP_ID = config["APP_ID"]
	APP_SECRET = config["APP_SECRET"]
	WEBHOOK_VERIFY_TOKEN = config["WEBHOOK_VERIFY_TOKEN"]
	ACCOUNT_TOKEN = config["ACCOUNT_TOKEN"]
	ACCOUNT_ID = config["ACCOUNT_ID"]
	PAGE_ACCESS_TOKEN = config["PAGE_ACCESS_TOKEN"]
	TURSO_DATABASE_URL = config["TURSO_DATABASE_URL"]
	TURSO_AUTH_TOKEN = config["TURSO_AUTH_TOKEN"]
	// DB_PATH = TURSO_DATABASE_URL + "?authToken=" + TURSO_AUTH_TOKEN
}

var CLIENT_ID string
var CLIENT_SECRET string
var APP_ID string
var APP_SECRET string
var WEBHOOK_VERIFY_TOKEN string
var ACCOUNT_TOKEN string
var ACCOUNT_ID string
var PAGE_ACCESS_TOKEN string
var TURSO_DATABASE_URL string
var TURSO_AUTH_TOKEN string
var DB_PATH = "./internal/db/purple-check.db"
