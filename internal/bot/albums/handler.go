package albums

import (
	"context"
	"log"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/s2gamb/music/internal/db"
)

type AlbumHandler struct {
	repo *db.Repository
}

func NewAlbumHandler(repo *db.Repository) *AlbumHandler {
	return &AlbumHandler{repo: repo}
}

func (h *AlbumHandler) HandleCallbackSelect(ctx context.Context, b *bot.Bot, update *models.Update) {
	albumID, err := strconv.Atoi(update.CallbackQuery.Data[len("select_"):])
	if err != nil {
		log.Printf("Error parsing album ID: %v", err)
		return
	}

	album, err := h.repo.GetAlbumByID(ctx, albumID)
	if err != nil || album == nil {
		log.Printf("Error fetching album: %v", err)
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Album not found.",
		})
		return
	}

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: update.CallbackQuery.ID})

	// Delete the album list message
	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
	})

	// Send the album photo message
	b.SendDocument(ctx, &bot.SendDocumentParams{
		ChatID:   update.CallbackQuery.Message.Message.Chat.ID,
		Document: &models.InputFileString{Data: album.CoverFileID},
		Caption:  album.Name + " - " + album.Artist + "\nYear: " + strconv.Itoa(int(album.Year)),
		ReplyMarkup: &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{{Text: "▶️ Play All", CallbackData: "all_" + strconv.Itoa(album.ID)}},
				{{Text: "📜 Playlist", CallbackData: "list_" + strconv.Itoa(album.ID)}},
				{{Text: "🔗 Share", CallbackData: "share_" + strconv.Itoa(album.ID)}},
				{{Text: "⬅️ Back to Albums", CallbackData: "menu"}},
			},
		},
	})
}

func (h *AlbumHandler) HandleCallbackBackToMenu(ctx context.Context, b *bot.Bot, update *models.Update) {
	// Delete the album photo message
	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
	})

	// Resend the main menu
	h.ShowAlbums(ctx, b, update.CallbackQuery.Message.Message.Chat.ID)
}

func (h *AlbumHandler) HandleCallbackAll(ctx context.Context, b *bot.Bot, update *models.Update) {
	albumID, err := strconv.Atoi(update.CallbackQuery.Data[len("all_"):])
	if err != nil {
		return
	}

	tracks, err := h.repo.GetTracksByAlbumID(ctx, albumID)
	if err != nil || len(tracks) == 0 {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "No tracks found.",
		})
		return
	}

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: update.CallbackQuery.ID, Text: "Sending tracks..."})

	for _, track := range tracks {
		b.SendAudio(ctx, &bot.SendAudioParams{
			ChatID: update.CallbackQuery.Message.Message.Chat.ID,
			Audio:  &models.InputFileString{Data: track.FileID},
			Title:  track.Title,
		})
	}
}

func (h *AlbumHandler) HandleCallbackPlayTrack(ctx context.Context, b *bot.Bot, update *models.Update) {
	trackID, err := strconv.Atoi(update.CallbackQuery.Data[len("play_"):])
	if err != nil {
		return
	}

	track, err := h.repo.GetTrackByID(ctx, trackID)
	if err != nil || track == nil {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: update.CallbackQuery.ID, Text: "Track not found."})
		return
	}

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: update.CallbackQuery.ID, Text: "Sending " + track.Title})

	b.SendAudio(ctx, &bot.SendAudioParams{
		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		Audio:  &models.InputFileString{Data: track.FileID},
		Title:  track.Title,
	})
}

func (h *AlbumHandler) HandleCallbackPlaylist(ctx context.Context, b *bot.Bot, update *models.Update) {
	albumID, err := strconv.Atoi(update.CallbackQuery.Data[len("list_"):])
	if err != nil {
		return
	}

	tracks, err := h.repo.GetTracksByAlbumID(ctx, albumID)
	if err != nil || len(tracks) == 0 {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: update.CallbackQuery.ID, Text: "No tracks found."})
		return
	}

	var keyboard [][]models.InlineKeyboardButton
	for _, track := range tracks {
		keyboard = append(keyboard, []models.InlineKeyboardButton{{
			Text:         track.Title,
			CallbackData: "play_" + strconv.Itoa(track.ID),
		}})
	}

	// Add Back button
	keyboard = append(keyboard, []models.InlineKeyboardButton{
		{Text: "🔙 Back", CallbackData: "back_" + strconv.Itoa(albumID)},
	})

	// Edit caption to indicate it's a playlist
	b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Caption:   "Select a track:",
	})

	b.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		ReplyMarkup: &models.InlineKeyboardMarkup{InlineKeyboard: keyboard},
	})
}

func (h *AlbumHandler) HandleCallbackBack(ctx context.Context, b *bot.Bot, update *models.Update) {
	albumID, err := strconv.Atoi(update.CallbackQuery.Data[len("back_"):])
	if err != nil {
		return
	}

	album, err := h.repo.GetAlbumByID(ctx, albumID)
	if err != nil || album == nil {
		return
	}

	b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Caption:   album.Name + " - " + album.Artist + "\nYear: " + strconv.Itoa(int(album.Year)),
	})

	b.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		ReplyMarkup: &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{{Text: "▶️ Play All", CallbackData: "all_" + strconv.Itoa(album.ID)}},
				{{Text: "📜 Playlist", CallbackData: "list_" + strconv.Itoa(album.ID)}},
				{{Text: "🔗 Share", CallbackData: "share_" + strconv.Itoa(album.ID)}},
				{{Text: "⬅️ Back", CallbackData: "menu"}},
			},
		},
	})
}

func (h *AlbumHandler) ShowAlbums(ctx context.Context, b *bot.Bot, chatID int64) {
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

func (h *AlbumHandler) HandleCallbackShare(ctx context.Context, b *bot.Bot, update *models.Update) {
	albumID, err := strconv.Atoi(update.CallbackQuery.Data[len("share_"):])
	if err != nil {
		return
	}

	link := "https://t.me/kinda_music_bot?start=album_" + strconv.Itoa(albumID)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		Text:      "Copy this link:\n`" + link + "`",
		ParseMode: "MarkdownV2",
	})
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
	})
}

func (h *AlbumHandler) ShowAlbumDirectly(ctx context.Context, b *bot.Bot, chatID int64, albumID int) {
	album, err := h.repo.GetAlbumByID(ctx, albumID)
	if err != nil || album == nil {
		return
	}

	b.SendDocument(ctx, &bot.SendDocumentParams{
		ChatID:   chatID,
		Document: &models.InputFileString{Data: album.CoverFileID},
		Caption:  album.Name + " - " + album.Artist + "\nYear: " + strconv.Itoa(int(album.Year)),
		ReplyMarkup: &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{{Text: "▶️ Play All", CallbackData: "all_" + strconv.Itoa(album.ID)}},
				{{Text: "📜 Playlist", CallbackData: "list_" + strconv.Itoa(album.ID)}},
				{{Text: "🔗 Share", CallbackData: "share_" + strconv.Itoa(album.ID)}},
				{{Text: "⬅️ Back", CallbackData: "menu"}},
			},
		},
	})
}
