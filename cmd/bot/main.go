package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"

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

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// 1. Initialize DB Repository
	repo, err := db.NewRepository(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer repo.Close(ctx)

	// 2. Initialize Handler with dependency
	handler := mybot.NewHandler(repo)

	// 3. Setup Bot
	b, err := bot.New(token)
	if err != nil {
		log.Fatal(err)
	}

	// Register handlers explicitly
	b.RegisterHandler(bot.HandlerTypeMessageText, "start", bot.MatchTypeCommand, handler.Start)
	b.RegisterHandler(bot.HandlerTypeMessageText, "", bot.MatchTypePrefix, handler.Handle)

	go func() {
		log.Printf("Starting bot server on port %s...", port)
		log.Fatal(http.ListenAndServe(":"+port, b.WebhookHandler()))
	}()

	log.Printf("Bot listening...")

	b.StartWebhook(ctx)
}
