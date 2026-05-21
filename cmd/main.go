package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"time"

	"purple-check/internal/config"
	"purple-check/internal/database"
	"purple-check/internal/instagram"
	"purple-check/internal/layout"
	"purple-check/internal/messaging"
	"purple-check/internal/middleware"
	"purple-check/internal/models"
	"purple-check/internal/routes"
	"purple-check/internal/routes/home"
	"purple-check/internal/routes/profile"
	"purple-check/internal/routes/search"
	"purple-check/internal/routes/webhook"
)

var origins = []string{
	"http://localhost:7331",
	"http://dev.purple-check.org",
}

func main() {
	messaging.InitConversations()

	appDB, err := database.InitDb(context.Background())
	if err != nil {
		log.Fatal("Failed to connect to database.\n", err)
	}
	defer func() {
		if err := database.Close(appDB); err != nil {
			slog.Error("error closing database", "error", err)
		}
	}()

	feedbacks := &models.FeedbackModel{DB: appDB}
	messageLogs := &models.MessageLogModel{DB: appDB}
	appSettings := &models.AppSettingModel{DB: appDB}
	tokenStore := instagram.NewTokenStore(appSettings, 24*time.Hour)
	if err := messaging.SetPersistentMenu(context.Background(), tokenStore); err != nil {
		slog.Error("error setting persistent menu", "error", err)
	}
	messageRouter := messaging.NewRouter(feedbacks, messageLogs, tokenStore)

	mux := http.NewServeMux()

	mux.Handle("/", home.NewHandler(feedbacks))
	mux.Handle("POST /search", search.NewHandler())

	mux.Handle("/profile/{username}", profile.NewHandler(feedbacks))
	mux.Handle("GET /privacy-policy", layout.Handler(routes.PrivacyPolicy(), layout.Head{
		Title: "Purple Check - Privacy Policy",
		URL:   "https://www.purple-check.org/privacy-policy",
	}))
	mux.Handle("GET /delete-my-data", layout.Handler(routes.DeleteMyData(), layout.Head{
		Title:       "Delete My Data",
		Description: "Delete your reviews from Purple Check.",
		URL:         "https://www.purple-check.org/delete-my-data",
	}))
	mux.Handle("GET /terms-of-service", layout.Handler(routes.TermsOfService(), layout.Head{
		Title: "Purple Check - Terms of Service",
		URL:   "https://www.purple-check.org/terms-of-service",
	}))

	mux.HandleFunc("GET /webhook/instagram", webhook.VerifyInstagramHook)
	mux.Handle("POST /webhook/instagram", webhook.NewInstagramHandler(messageRouter))
	mux.Handle("GET /webhook/instagram/setup", webhook.NewSetupHandler(tokenStore))
	mux.Handle("GET /instagram/refresh-access-token", instagram.NewRefreshAccessTokenHandler(tokenStore))
	mux.Handle("GET /static/", disableCacheInDevMode(http.StripPrefix("/static/", http.FileServer(http.Dir("static")))))

	handler := middleware.ConfigureCSP(mux)

	slog.Info("starting server", "port", config.PORT)
	err = http.ListenAndServe(":"+config.PORT, handler)
	log.Fatal(err)
}

func disableCacheInDevMode(next http.Handler) http.Handler {
	if !config.DEV {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
