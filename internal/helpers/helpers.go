package helpers

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"unicode"
)

func GetRequestBody(r *http.Request) (string, error) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	bodyString := string(bodyBytes)
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	return bodyString, nil
}

func GetResponseBody(r *http.Response) (string, error) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	bodyString := string(bodyBytes)
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	return bodyString, nil
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

	username, ok := NormalizeUsername(username)
	if !ok {
		return "", false
	}

	return username, username != ""
}

func NormalizeUsername(username string) (string, bool) {
	username = strings.TrimSpace(username)
	username = strings.ToLower(username)
	username = strings.TrimPrefix(username, "@")
	return username, ValidateUsername(username)
}

func ValidateUsername(username string) bool {
	if len(username) < 3 || len(username) > 30 {
		return false
	}

	if strings.HasPrefix(username, ".") || strings.HasSuffix(username, ".") || strings.Contains(username, "..") {
		return false
	}

	// Restrict to letters, numbers, periods and underscores
	for _, char := range username {
		if char > unicode.MaxASCII || !(unicode.IsLetter(char) || unicode.IsDigit(char) || char == '.' || char == '_') {
			return false
		}
	}

	return true
}
