package bot

import (
	"context"
	"log"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/s2gamb/music/internal/db"
)

type Handler struct {
	repo *db.Repository
}

func NewHandler(repo *db.Repository) *Handler {
	return &Handler{repo: repo}
}

// Start command handler
func (h *Handler) Start(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.showAlbums(ctx, b, update.Message.Chat.ID)
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
	trackID, err := strconv.Atoi(update.CallbackQuery.Data[len("select_"):])
	if err != nil {
		log.Printf("Error parsing track ID: %v", err)
		return
	}

	tracks, err := h.repo.GetTracksByAlbumID(ctx, trackID)
	if err != nil || len(tracks) == 0 {
		log.Printf("Error fetching tracks: %v", err)
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "No tracks found for this album.",
		})
		return
	}

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "Sending tracks...",
	})

	for _, track := range tracks {
		b.SendAudio(ctx, &bot.SendAudioParams{
			ChatID: update.CallbackQuery.Message.Message.Chat.ID,
			Audio:  &models.InputFileString{Data: track.FileID},
			Title:  track.Title,
		})
	}
}

func (h *Handler) showAlbums(ctx context.Context, b *bot.Bot, chatID int64) {
	albums, err := h.repo.GetAllAlbums(ctx)
	if err != nil {
		log.Printf("Error fetching albums: %v", err)
		return
	}

	if len(albums) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "No albums found.",
		})
		return
	}

	var keyboard [][]models.InlineKeyboardButton
	for _, album := range albums {
		keyboard = append(keyboard, []models.InlineKeyboardButton{
			{
				Text:         album.Name + " - " + album.Artist,
				CallbackData: "select_" + strconv.Itoa(album.ID),
			},
		})
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Choose an album:",
		ReplyMarkup: &models.InlineKeyboardMarkup{
			InlineKeyboard: keyboard,
		},
	})
}
