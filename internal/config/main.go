package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config map[string]string

func init() {
	config := make(Config)
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	expected_keys := []string{"CLIENT_ID", "CLIENT_SECRET", "APP_ID", "APP_SECRET", "WEBHOOK_VERIFY_TOKEN", "ACCOUNT_TOKEN", "ACCOUNT_ID", "TURSO_DATABASE_URL", "TURSO_AUTH_TOKEN", "LOCAL_DB_PATH", "PORT", "HOST"}

	for _, key := range expected_keys {
		config[key] = os.Getenv(key)
		if config[key] == "" {
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
	TURSO_DATABASE_URL = config["TURSO_DATABASE_URL"]
	TURSO_AUTH_TOKEN = config["TURSO_AUTH_TOKEN"]
	LOCAL_DB_PATH = config["LOCAL_DB_PATH"]
	HOST = config["HOST"]
	PORT = config["PORT"]
}

var CLIENT_ID string
var CLIENT_SECRET string
var APP_ID string
var APP_SECRET string
var WEBHOOK_VERIFY_TOKEN string
var ACCOUNT_TOKEN string
var ACCOUNT_ID string
var TURSO_DATABASE_URL string
var TURSO_AUTH_TOKEN string
var HOST string
var PORT string
var LOCAL_DB_PATH string
