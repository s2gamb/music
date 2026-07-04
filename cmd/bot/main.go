package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/go-telegram/bot"
	mybot "github.com/s2gamb/music/internal/bot" // Your internal logic
	"github.com/s2gamb/music/internal/db"        // Your db package
)

func main() {
	token := os.Getenv("BOT_TOKEN")
	dbURL := os.Getenv("DATABASE_URL")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if token == "" || dbURL == "" {
		log.Fatal("BOT_TOKEN and DATABASE_URL env variables are required")
	}

	// 1. Initialize DB Repository
	ctx := context.Background()
	repo, err := db.NewRepository(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer repo.Close(ctx)

	// 2. Initialize Handler with dependency
	handler := mybot.NewHandler(repo)

	// 3. Setup Bot
	opts := []bot.Option{
		bot.WithDefaultHandler(handler.Handle),
	}
	b, err := bot.New(token, opts...)
	if err != nil {
		log.Fatal(err)
	}

	// 4. Webhook Setup
	webhook := b.WebhookHandler()
	http.HandleFunc("/webhook/"+token, webhook)

	log.Printf("Starting bot server on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
