package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/s2gamb/music/internal/db"
	"github.com/s2gamb/music/internal/media"
	"github.com/s2gamb/music/internal/models"
	"github.com/spf13/cobra"
)

var (
	coverPath  string
	year       string
	albumName  string
	artistFlag string
)

func init() {
	UploadCmd.Flags().StringVar(&coverPath, "cover", "", "Path to the cover image")
	UploadCmd.Flags().StringVar(&year, "year", "", "Year of the album")
	UploadCmd.Flags().StringVar(&albumName, "name", "", "Name of the album")
	UploadCmd.Flags().StringVar(&albumName, "album", "", "Name of the album (alias for --name)")
	UploadCmd.Flags().StringVar(&artistFlag, "artist", "", "Artist of the album/tracks")
}

var UploadCmd = &cobra.Command{
	Use:   "upload [directory]",
	Short: "Upload tracks from a directory to Telegram and save to DB",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dirPath := "./"
		if len(args) > 0 {
			dirPath = args[0]
		}

		dbURL := os.Getenv("DATABASE_URL")
		if dbURL == "" {
			return fmt.Errorf("DATABASE_URL environment variable is required")
		}

		repo, err := db.NewRepository(cmd.Context(), dbURL)
		if err != nil {
			return err
		}
		defer repo.Close(cmd.Context())

		entries, err := os.ReadDir(dirPath)
		if err != nil {
			return err
		}

		// Prepare album data (assuming all tracks in folder belong to same album)
		var album models.Album
		var albumCreated bool

		token := os.Getenv("TELEGRAM_BOT_TOKEN")
		chatID := os.Getenv("TELEGRAM_CHAT_ID")

		fmt.Printf("Scanning directory: %s\n", dirPath)

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".mp3") {
				continue
			}

			fullPath := filepath.Join(dirPath, entry.Name())
			fileName := entry.Name()

			existing, err := repo.GetTrackByFilename(cmd.Context(), fileName)
			if err != nil {
				return fmt.Errorf("database error: %w", err)
			}
			if existing != nil {
				fmt.Printf("Skipping %s: already exists\n", entry.Name())
				continue
			}

			track, err := media.GetTrackInfo(fullPath, fileName)
			if err != nil {
				fmt.Printf("Error processing %s: %v\n", entry.Name(), err)
				continue
			}

			// Apply fallback logic for track/album details
			if track.Album == "" {
				if albumName != "" {
					track.Album = albumName
				} else {
					absPath, err := filepath.Abs(dirPath)
					if err == nil {
						track.Album = filepath.Base(absPath)
					} else {
						track.Album = filepath.Base(dirPath)
					}
				}
			}

			if track.Artist == "" {
				if artistFlag != "" {
					track.Artist = artistFlag
				} else {
					track.Artist = "Unknown Artist"
				}
			}

			if track.Title == "" {
				track.Title = strings.TrimSuffix(fileName, filepath.Ext(fileName))
			}

			albumKey := track.Album
			var tmpAlbum *models.Album
			tmpAlbum, err = repo.GetAlbum(cmd.Context(), albumKey)
			if err != nil {
				return fmt.Errorf("database error: %w", err)
			}
			if tmpAlbum != nil {
				albumCreated = true
			}

			// Create Album record if not yet created
			if !albumCreated {
				y, _ := strconv.Atoi(year)
				album = models.Album{
					Name:   track.Album,
					Artist: track.Artist,
					Year:   int16(y),
				}

				// Upload cover if provided
				if coverPath != "" {
					cfID, err := media.UploadDocument(cmd.Context(), token, chatID, coverPath)
					if err != nil {
						fmt.Printf("Error uploading cover: %v\n", err)
					} else {
						album.CoverFileID = cfID
					}
				}

				_, err = repo.CreateAlbum(cmd.Context(), &album)
				if err != nil {
					return fmt.Errorf("failed to create album in DB: %w", err)
				}
				albumCreated = true
				fmt.Printf("Album '%s' created (ID: %d)\n", album.Name, album.ID)
			} else {
				album = *tmpAlbum
			}

			// Upload to Telegram
			fileID, err := media.UploadAudio(cmd.Context(), token, chatID, fullPath, track.Title, track.Artist)
			if err != nil {
				fmt.Printf("Error uploading %s: %v\n", entry.Name(), err)
				continue
			}
			track.FileID = fileID
			track.AlbumID = album.ID

			// Save track to DB
			_, err = repo.CreateTrack(cmd.Context(), &track)
			if err != nil {
				fmt.Printf("Error saving to DB: %v\n", err)
				continue
			}

			fmt.Printf("Successfully uploaded: %s - %s\n", track.Artist, track.Title)
		}

		return nil
	},
}
