package bot

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/s2gamb/music/internal/bot/albums"
	"github.com/s2gamb/music/internal/db"
)

type Handler struct {
	repo         *db.Repository
	albumHandler *albums.AlbumHandler
}

func NewHandler(repo *db.Repository) *Handler {
	return &Handler{
		repo:         repo,
		albumHandler: albums.NewAlbumHandler(repo),
	}
}

// Start command handler
func (h *Handler) Start(ctx context.Context, b *bot.Bot, update *models.Update) {
	// Parse command arguments: "/start album_123" -> ["album_123"]
	text := update.Message.Text
	parts := strings.Fields(text)
	var args []string
	if len(parts) > 1 {
		args = parts[1:]
	}

	if len(args) > 0 && strings.HasPrefix(args[0], "album_") {
		albumIDStr := strings.TrimPrefix(args[0], "album_")
		albumID, err := strconv.Atoi(albumIDStr)
		if err == nil {
			h.albumHandler.ShowAlbumDirectly(ctx, b, update.Message.Chat.ID, albumID)
			return
		}
	}
	h.albumHandler.ShowAlbums(ctx, b, update.Message.Chat.ID)
}

// Handle handles general text messages
func (h *Handler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Printf("Received message: %s", update.Message.Text)
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Эхо: " + update.Message.Text,
	})
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (h *Handler) HandleSelectCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.albumHandler.HandleCallbackSelect(ctx, b, update)
}

func (h *Handler) HandleAllCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.albumHandler.HandleCallbackAll(ctx, b, update)
}

func (h *Handler) HandlePlaylistCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.albumHandler.HandleCallbackPlaylist(ctx, b, update)
}

func (h *Handler) HandleMenuCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.albumHandler.HandleCallbackBackToMenu(ctx, b, update)
}

func (h *Handler) HandlePlayTrackCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.albumHandler.HandleCallbackPlayTrack(ctx, b, update)
}

func (h *Handler) HandleBackCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.albumHandler.HandleCallbackBack(ctx, b, update)
}

func (h *Handler) HandleShareCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.albumHandler.HandleCallbackShare(ctx, b, update)
}
