package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config map[string]string

func init() {
	runningTests := strings.HasSuffix(os.Args[0], ".test")
	err := godotenv.Load()
	if err != nil && !runningTests {
		log.Fatal("Error loading .env file")
	}

	config := make(Config)

	expected_keys := []string{
		"APP_ID",
		"WEBHOOK_VERIFY_TOKEN",
		"ACCOUNT_ID",
		"ADMIN_TOKEN",
		"LOCAL_DB_PATH",
		"PORT",
		"HOST",
		"DEV",
		"INSTAGRAM_API_VERSION",
	}

	for _, key := range expected_keys {
		config[key] = os.Getenv(key)
		if config[key] == "" {
			if key == "DEV" {
				config[key] = "false"
			} else if key == "INSTAGRAM_API_VERSION" {
				config[key] = "v25.0"
			} else if key == "ADMIN_TOKEN" {
				config[key] = ""
			} else if runningTests {
				config[key] = "test"
			} else {
				panic("Missing key in .env file: " + key)
			}
		}
	}

	APP_ID = config["APP_ID"]
	WEBHOOK_VERIFY_TOKEN = config["WEBHOOK_VERIFY_TOKEN"]
	ADMIN_TOKEN = config["ADMIN_TOKEN"]
	ACCOUNT_ID = config["ACCOUNT_ID"]
	LOCAL_DB_PATH = config["LOCAL_DB_PATH"]
	HOST = config["HOST"]
	PORT = config["PORT"]
	DEV = config["DEV"] == "true"
	INSTAGRAM_API_VERSION = config["INSTAGRAM_API_VERSION"]
}

var APP_ID string
var WEBHOOK_VERIFY_TOKEN string
var ADMIN_TOKEN string
var ACCOUNT_ID string
var HOST string
var PORT string
var LOCAL_DB_PATH string
var DEV bool
var INSTAGRAM_API_VERSION string
