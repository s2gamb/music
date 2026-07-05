package db

import (
	"context"
	"errors"
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

func (r *Repository) GetAlbum(ctx context.Context, name string) (*models.Album, error) {
	var a models.Album
	err := r.conn.QueryRow(ctx, "SELECT id, name, artist, year, cover_file_id FROM albums WHERE name = $1", name).Scan(&a.ID, &a.Name, &a.Artist, &a.Year, &a.CoverFileID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("querying album by name: %w", err)
	}
	return &a, nil
}

func (r *Repository) GetAlbumByID(ctx context.Context, id int) (*models.Album, error) {
	var a models.Album
	err := r.conn.QueryRow(ctx, "SELECT id, name, artist, year, cover_file_id FROM albums WHERE id = $1", id).Scan(&a.ID, &a.Name, &a.Artist, &a.Year, &a.CoverFileID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("querying album by ID: %w", err)
	}
	return &a, nil
}

func (r *Repository) UpdateAlbum(ctx context.Context, album *models.Album) error {
	_, err := r.conn.Exec(ctx,
		"UPDATE albums SET year = $1, cover_file_id = $2 WHERE id = $3",
		album.Year, album.CoverFileID, album.ID)
	return err
}

func (r *Repository) GetAllAlbums(ctx context.Context) ([]*models.Album, error) {
	rows, err := r.conn.Query(ctx, "SELECT id, name, artist, year, cover_file_id FROM albums")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var albums []*models.Album
	for rows.Next() {
		var a models.Album
		if err := rows.Scan(&a.ID, &a.Name, &a.Artist, &a.Year, &a.CoverFileID); err != nil {
			return nil, err
		}
		albums = append(albums, &a)
	}
	return albums, nil
}

func (r *Repository) CreateTrack(ctx context.Context, track *models.Track) (int, error) {
	err := r.conn.QueryRow(ctx,
		"INSERT INTO tracks (album_id, filename, title, artist, album, duration, file_id) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		track.AlbumID, track.Filename, track.Title, track.Artist, track.Album, track.Duration.String(), track.FileID).Scan(&track.ID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert track: %w", err)
	}
	return track.ID, nil
}

func (r *Repository) GetTrackByID(ctx context.Context, id int) (*models.Track, error) {
	var track models.Track
	err := r.conn.QueryRow(ctx,
		"SELECT id, album_id, filename, title, artist, album, duration, file_id FROM tracks WHERE id = $1",
		id).Scan(&track.ID, &track.AlbumID, &track.Filename, &track.Title, &track.Artist, &track.Album, &track.Duration, &track.FileID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("querying track by ID: %w", err)
	}
	return &track, nil
}

func (r *Repository) GetTrackByFilename(ctx context.Context, filename string) (*models.Track, error) {
	var track models.Track
	err := r.conn.QueryRow(ctx,
		"SELECT id, album_id, filename, title, artist, album, duration, file_id FROM tracks WHERE filename = $1",
		filename).Scan(&track.ID, &track.AlbumID, &track.Filename, &track.Title, &track.Artist, &track.Album, &track.Duration, &track.FileID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("querying track by filename: %w", err)
	}
	return &track, nil
}

func (r *Repository) GetTracksByAlbumID(ctx context.Context, albumID int) ([]*models.Track, error) {
	var tracks []*models.Track
	rows, err := r.conn.Query(ctx,
		"SELECT id, album_id, filename, title, artist, album, duration, file_id FROM tracks WHERE album_id = $1",
		albumID)
	if err != nil {
		return nil, fmt.Errorf("querying tracks by album ID: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var track models.Track
		if err := rows.Scan(&track.ID, &track.AlbumID, &track.Filename, &track.Title, &track.Artist, &track.Album, &track.Duration, &track.FileID); err != nil {
			return nil, fmt.Errorf("scanning track: %w", err)
		}
		tracks = append(tracks, &track)
	}
	return tracks, nil
}
