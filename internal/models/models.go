package models

import (
	"time"
)

type Album struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Artist      string `json:"artist"`
	Year        int16  `json:"year"`
	CoverFileID string `json:"cover_file_id"`
}

type Track struct {
	ID        int           `json:"id"`
	AlbumID   int           `json:"album_id"`
	Filename  string        `json:"filename"`
	Title     string        `json:"title"`
	Artist    string        `json:"artist"`
	Album     string        `json:"album"`
	Duration  time.Duration `json:"duration"`
	FileID    string        `json:"file_id"`
	CreatedAt time.Time     `json:"created_at"`
}
