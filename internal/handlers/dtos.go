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
	ID           int64             `json:"id"`
	FolderID     int64             `json:"folder_id"`
	ArtistID     *int64            `json:"artist_id,omitempty"`
	AlbumID      *int64            `json:"album_id,omitempty"`
	Artist       *ArtistSummaryDTO `json:"artist,omitempty"`
	Album        *AlbumSummaryDTO  `json:"album,omitempty"`
	RelPath      string            `json:"rel_path"`
	Title        string            `json:"title"`
	Filename     string            `json:"filename"`
	Ext          string            `json:"ext"`
	Genre        *string           `json:"genre,omitempty"`
	Year         *int64            `json:"year,omitempty"`
	Rating       *int64            `json:"rating,omitempty"`
	DurationSec  *int64            `json:"duration_seconds,omitempty"`
	ImagePath    *string           `json:"image_path,omitempty"`
	SizeBytes    int64             `json:"size_bytes"`
	LastModified int64             `json:"last_modified"`
	LastSeenAt   time.Time         `json:"last_seen_at"`
	DeletedAt    *time.Time        `json:"deleted_at,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
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
	ID        int64             `json:"id"`
	ArtistID  int64             `json:"artist_id"`
	Artist    *ArtistSummaryDTO `json:"artist,omitempty"`
	Title     string            `json:"title"`
	ImagePath *string           `json:"image_path,omitempty"`
	DeletedAt *time.Time        `json:"deleted_at,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type ArtistSummaryDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type AlbumSummaryDTO struct {
	ID        int64   `json:"id"`
	ArtistID  int64   `json:"artist_id"`
	Title     string  `json:"title"`
	ImagePath *string `json:"image_path,omitempty"`
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

type DayViewDTO struct {
	Year    int64             `json:"year"`
	Month   int64             `json:"month"`
	Day     int64             `json:"day"`
	Entries []JournalEntryDTO `json:"entries"`
}

type JournalAssetDTO struct {
	Path     string `json:"path"`
	Filename string `json:"filename"`
	Size     int64  `json:"size_bytes"`
}

type updateJournalEntryRequest struct {
	Raw string `json:"raw"`
}

type SettingDTO struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SettingKeyDTO struct {
	Key         string `json:"key"`
	Description string `json:"description"`
}

type MetalPriceDTO struct {
	Name              string  `json:"name"`
	GBP               float64 `json:"gbp"`
	USD               float64 `json:"usd"`
	Symbol            string  `json:"symbol"`
	UpdatedAt         string  `json:"updatedAt"`
	UpdatedAtReadable string  `json:"updatedAtReadable"`
}

type MetalsPricesDTO struct {
	Gold   MetalPriceDTO `json:"gold"`
	Silver MetalPriceDTO `json:"silver"`
}

type JournalDTO struct {
	Year          int64     `json:"year"`
	Month         int64     `json:"month"`
	Day           int64     `json:"day"`
	SizeBytes     int64     `json:"size_bytes"`
	Hash          string    `json:"hash"`
	Tags          []string  `json:"tags,omitempty"`
	LastCheckedAt time.Time `json:"last_checked_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type JournalDayDTO struct {
	Year  int64  `json:"year"`
	Month int64  `json:"month"`
	Day   int64  `json:"day"`
	Raw   string `json:"raw"`
}

type JournalGitSyncDTO struct {
	Status        string `json:"status"`
	CommitSkipped bool   `json:"commit_skipped"`
	CommitOutput  string `json:"commit_output,omitempty"`
	PullOutput    string `json:"pull_output,omitempty"`
	PushOutput    string `json:"push_output,omitempty"`
}

type JournalEntryDTO struct {
	ID           int64     `json:"id"`
	Year         int64     `json:"year"`
	Month        int64     `json:"month"`
	Day          int64     `json:"day"`
	Position     int64     `json:"position"`
	Title        string    `json:"title"`
	RawLine      string    `json:"raw_line"`
	Hash         string    `json:"hash"`
	Body         *string   `json:"body,omitempty"`
	Status       *string   `json:"status"`
	Tags         []string  `json:"tags,omitempty"`
	PropertyKeys []string  `json:"property_keys,omitempty"`
	Type         string    `json:"type"`
	ScheduledAt  *string   `json:"scheduled_at"`
	DeadlineAt   *string   `json:"deadline_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
