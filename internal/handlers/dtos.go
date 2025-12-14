package handlers

import "time"

type FolderDTO struct {
	ID             int64      `json:"id"`
	Path           string     `json:"path"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
	Available      bool       `json:"available"`
	LastSeenAt     *time.Time `json:"last_seen_at,omitempty"`
	LastScanAt     *time.Time `json:"last_scan_at,omitempty"`
	LastScanStatus *string    `json:"last_scan_status,omitempty"`
	LastScanError  *string    `json:"last_scan_error,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type TrackDTO struct {
	ID           int64      `json:"id"`
	FolderID     int64      `json:"folder_id"`
	ArtistID     *int64     `json:"artist_id,omitempty"`
	AlbumID      *int64     `json:"album_id,omitempty"`
	RelPath      string     `json:"rel_path"`
	Filename     string     `json:"filename"`
	Ext          string     `json:"ext"`
	Genre        *string    `json:"genre,omitempty"`
	Year         *int64     `json:"year,omitempty"`
	Rating       *int64     `json:"rating,omitempty"`
	SizeBytes    int64      `json:"size_bytes"`
	LastModified int64      `json:"last_modified"`
	LastSeenAt   time.Time  `json:"last_seen_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type ScanDTO struct {
	FolderID   int64      `json:"folder_id"`
	Status     string     `json:"status"` // "running" | "ok" | "error" | "skipped_unavailable"
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	Error      *string    `json:"error,omitempty"`
}

type ArtistDTO struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type AlbumDTO struct {
	ID        int64      `json:"id"`
	ArtistID  int64      `json:"artist_id"`
	Title     string     `json:"title"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type PlaylistDTO struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type PlaylistTrackDTO struct {
	ID         int64      `json:"id"`
	PlaylistID int64      `json:"playlist_id"`
	TrackID    int64      `json:"track_id"`
	Position   int64      `json:"position"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	Track      *TrackDTO  `json:"track,omitempty"`
}
