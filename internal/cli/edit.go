package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/s2gamb/music/internal/db"
	"github.com/s2gamb/music/internal/media"
	"github.com/spf13/cobra"
)

var (
	editAlbumName string
	editCoverPath string
	editYear      string
)

func init() {
	EditCmd.Flags().StringVar(&editAlbumName, "name", "", "Name of the album to edit")
	EditCmd.Flags().StringVar(&editCoverPath, "cover", "", "New path to the cover image")
	EditCmd.Flags().StringVar(&editYear, "year", "", "New year of the album")
	// Removed MarkFlagRequired("name") to allow directory inference
}

var EditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit album metadata",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbURL := os.Getenv("DATABASE_URL")
		if dbURL == "" {
			return fmt.Errorf("DATABASE_URL environment variable is required")
		}

		repo, err := db.NewRepository(cmd.Context(), dbURL)
		if err != nil {
			return err
		}
		defer repo.Close(cmd.Context())

		albumName := editAlbumName
		if albumName == "" {
			entries, err := os.ReadDir("./")
			if err != nil {
				return fmt.Errorf("failed to scan directory: %w", err)
			}
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".mp3") {
					info, err := media.GetTrackInfo(entry.Name(), entry.Name())
					if err == nil && info.Album != "" {
						albumName = info.Album
						fmt.Printf("Detected album from files: %s\n", albumName)
						break
					}
				}
			}
		}

		if albumName == "" {
			return fmt.Errorf("could not determine album name; use --name flag")
		}

		// Fetch existing album
		album, err := repo.GetAlbum(cmd.Context(), albumName)
		if err != nil {
			return fmt.Errorf("album not found: %w", err)
		}

		// Update fields
		if editYear != "" {
			y, err := strconv.Atoi(editYear)
			if err != nil {
				return fmt.Errorf("invalid year format: %w", err)
			}
			album.Year = int16(y)
		}

		if editCoverPath != "" {
			token := os.Getenv("TELEGRAM_BOT_TOKEN")
			chatID := os.Getenv("TELEGRAM_CHAT_ID")
			cfID, err := media.UploadDocument(cmd.Context(), token, chatID, editCoverPath)
			if err != nil {
				return fmt.Errorf("failed to upload new cover: %w", err)
			}
			album.CoverFileID = cfID
		}

		// TODO: Add UpdateAlbum method to Repository
		err = repo.UpdateAlbum(cmd.Context(), album)
		if err != nil {
			return fmt.Errorf("failed to update album in DB: %w", err)
		}

		fmt.Printf("Album '%s' updated successfully\n", album.Name)
		return nil
	},
}
