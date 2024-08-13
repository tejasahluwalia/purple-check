package config

import (
	"os"
)

var CLIENT_ID = os.Getenv("CLIENT_ID")
var CLIENT_SECRET = os.Getenv("CLIENT_SECRET")
var DB_PATH = "./internal/db/purple-check.db"