package config

import (
	"os"
	"strings"
)

func check(e error) {
    if e != nil {
        panic(e)
    }
}

type Config map[string]string

func init() {
	DOTENV_PATH := "./.env"
	dat, err := os.ReadFile(DOTENV_PATH)
	check(err)

	config := make(Config)

	expected_keys := []string{"CLIENT_ID", "CLIENT_SECRET", "APP_ID", "APP_SECRET"}
	
	lines := strings.Split(string(dat), "\n")
	for _, line := range lines {
		parts := strings.Split(line, "=")
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			config[key] = value
		}
	}

	for _, key := range expected_keys {
		if _, ok := config[key]; !ok {
			panic("Missing key in .env file: " + key)
		}
	}

	CLIENT_ID = config["CLIENT_ID"]
	CLIENT_SECRET = config["CLIENT_SECRET"]
	APP_ID = config["APP_ID"]
	APP_SECRET = config["APP_SECRET"]
}

var CLIENT_ID string
var CLIENT_SECRET string
var APP_ID string
var APP_SECRET string
var DB_PATH = "./internal/db/purple-check.db"