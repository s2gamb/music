package media

import (
	"os"
	"strings"
	"time"

	"github.com/hajimehoshi/go-mp3"
	"github.com/mikkyang/id3-go"
	"github.com/s2gamb/music/internal/models"
)

func sanitize(s string) string {
	return strings.ReplaceAll(s, "\x00", "")
}

// GetTrackInfo extracts metadata and duration from an MP3 file.
func GetTrackInfo(filePath, fileName string) (models.Track, error) {
	track := models.Track{
		Filename: fileName,
	}

	// 1. Read ID3 tags
	mp3File, err := id3.Open(filePath)
	if err == nil {
		defer mp3File.Close()
		track.Title = sanitize(strings.TrimSpace(mp3File.Title()))
		track.Artist = sanitize(strings.TrimSpace(mp3File.Artist()))
		track.Album = sanitize(strings.TrimSpace(mp3File.Album()))
	}

	// 2. Calculate duration
	file, err := os.Open(filePath)
	if err != nil {
		return track, err
	}
	defer file.Close()

	decoder, err := mp3.NewDecoder(file)
	if err != nil {
		return track, err
	}

	samples := decoder.Length() / 4
	track.Duration = time.Duration(samples) * time.Second / time.Duration(decoder.SampleRate())

	return track, nil
}
