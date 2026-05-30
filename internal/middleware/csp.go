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
	cspConfig := CSPConfig{}
	wrappedMux := withCSP(cspConfig)(mux)
	return wrappedMux
}

type CSPConfig struct {
	ScriptSrc  []string
	StyleSrc   []string
	ImgSrc     []string
	FontSrc    []string
	ConnectSrc []string
}

func withCSP(config CSPConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nonce, err := generateNonce()
			if err != nil {
				log.Printf("failed to generate nonce: %v", err)
				w.Header().Set("Content-Security-Policy", buildCSP(config, ""))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Security-Policy", buildCSP(config, nonce))

			next.ServeHTTP(w, r.WithContext(templ.WithNonce(r.Context(), nonce)))
		})
	}
}

func buildCSP(config CSPConfig, nonce string) string {
	scriptDefaults := []string{"'self'"}
	if nonce != "" {
		scriptDefaults = append(scriptDefaults, fmt.Sprintf("'nonce-%s'", nonce))
	}

	directives := []string{
		"default-src 'self'",
		fmt.Sprintf("script-src %s", joinSources(config.ScriptSrc, scriptDefaults...)),
		fmt.Sprintf("style-src %s", joinSources(config.StyleSrc, "'self'")),
		fmt.Sprintf("img-src %s", joinSources(config.ImgSrc, "'self'", "data:")),
		fmt.Sprintf("font-src %s", joinSources(config.FontSrc, "'self'")),
		fmt.Sprintf("connect-src %s", joinSources(config.ConnectSrc, "'self'")),
		"base-uri 'self'",
		"form-action 'self'",
		"frame-ancestors 'none'",
		"object-src 'none'",
	}
	return strings.Join(directives, "; ")
}

func joinSources(configured []string, defaults ...string) string {
	return strings.Join(append(defaults, configured...), " ")
}

func generateNonce() (string, error) {
	nonceBytes := make([]byte, 16)
	_, err := rand.Read(nonceBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	return base64.StdEncoding.EncodeToString(nonceBytes), nil
}
