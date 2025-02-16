package helpers

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strings"
)

func GetRequestBody(r *http.Request) string {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	return bodyString
}

func GetResponseBody(r *http.Response) string {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	return bodyString
}

func DetectUsername(message string) (string, bool) {
	username := ""
	if i := strings.Index(message, "@"); i != -1 {
		endIndex := strings.Index(message[i+1:], " ")
		if endIndex == -1 {
			username = message[i+1:]
		} else {
			username = message[i+1 : i+1+endIndex]
		}
	}

	if username == "" {
		return "", false
	}

	username = strings.ToLower(username)
	username = strings.TrimPrefix(username, "@")

	if !ValidateUsername(username) {
		return "", false
	}

	return username, username != ""
}

func ValidateUsername(username string) bool {
	if len(username) < 3 || len(username) > 30 {
		return false
	}

	if strings.Contains(username, "..") {
		return false
	}

	// Restrict to letters, numbers, periods and underscores
	allowed := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789._"
	for _, char := range username {
		if !strings.Contains(allowed, string(char)) {
			return false
		}
	}

	return true
}
