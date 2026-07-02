package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/s2gamb/music/internal/models"
)

type Repository struct {
	conn *pgx.Conn
}

func NewRepository(ctx context.Context, dbURL string) (*Repository, error) {
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	return &Repository{conn: conn}, nil
}

func (r *Repository) Close(ctx context.Context) {
	r.conn.Close(ctx)
}

func (r *Repository) CreateAlbum(ctx context.Context, album *models.Album) (int, error) {
	err := r.conn.QueryRow(ctx,
		"INSERT INTO albums (name, artist, year, cover_file_id) VALUES ($1, $2, $3, $4) RETURNING id",
		album.Name, album.Artist, album.Year, album.CoverFileID).Scan(&album.ID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert album: %w", err)
	}
	return album.ID, nil
}

func (r *Repository) CreateTrack(ctx context.Context, track *models.Track) (int, error) {
	err := r.conn.QueryRow(ctx,
		"INSERT INTO tracks (album_id, filename, title, file_id) VALUES ($1, $2, $3, $4) RETURNING id",
		track.AlbumID, track.Filename, track.Title, track.FileID).Scan(&track.ID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert track: %w", err)
	}
	return track.ID, nil
}
