package middleware

import (
	"log"
	"net/http"
	"strings"

	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/a-h/templ"
)

func ConfigureCSP(mux *http.ServeMux) http.Handler {
	cspConfig := CSPConfig{
		ScriptSrc: []string{"cdn.jsdelivr.net"}, // Add external script sources here
	}
	wrappedMux := withCSP(cspConfig)(mux)
	return wrappedMux
}

type CSPConfig struct {
	ScriptSrc []string // External script domains allowed
}

func withCSP(config CSPConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nonce, err := generateNonce()
			if err != nil {
				log.Printf("failed to generate nonce: %v", err)
				w.Header().Set("Content-Security-Policy", "script-src 'self'")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Combine all script sources
			scriptSources := append(
				[]string{"'self'", fmt.Sprintf("'nonce-%s'", nonce)},
				config.ScriptSrc...)

			csp := fmt.Sprintf("script-src %s", strings.Join(scriptSources, " "))
			w.Header().Set("Content-Security-Policy", csp)

			next.ServeHTTP(w, r.WithContext(templ.WithNonce(r.Context(), nonce)))
		})
	}
}

func generateNonce() (string, error) {
	nonceBytes := make([]byte, 16)
	_, err := rand.Read(nonceBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	return base64.StdEncoding.EncodeToString(nonceBytes), nil
}
