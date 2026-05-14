package helpers

import (
	"errors"
	"strings"
	"unicode"
)

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

	username = NormalizeUsername(username)
	err := ValidateUsername(username)
	if err != nil {
		return "", false
	}

	return username, username != ""
}

func NormalizeUsername(username string) string {
	username = strings.TrimSpace(username)
	username = strings.ToLower(username)
	username = strings.TrimPrefix(username, "@")
	return username
}

func ValidateUsername(username string) error {
	if len(username) < 3 || len(username) > 30 {
		return errors.New("Invalid username.")
	}

	if strings.HasPrefix(username, ".") || strings.HasSuffix(username, ".") || strings.Contains(username, "..") {
		return errors.New("Invalid username.")
	}

	// Restrict to letters, numbers, periods and underscores
	for _, char := range username {
		if char > unicode.MaxASCII || !(unicode.IsLetter(char) || unicode.IsDigit(char) || char == '.' || char == '_') {
			return errors.New("Invalid username.")
		}
	}

	return nil
}
