package handlers

import (
	"bottomley.ian/musicserver/internal/db"
	"time"
)

func folderDTOFromDB(f db.Folder) FolderDTO {
	var deletedAt *time.Time
	if f.DeletedAt.Valid {
		t := f.DeletedAt.Time
		deletedAt = &t
	}

	var lastSeenAt *time.Time
	if f.LastSeenAt.Valid {
		t := f.LastSeenAt.Time
		lastSeenAt = &t
	}

	var lastScanAt *time.Time
	if f.LastScanAt.Valid {
		t := f.LastScanAt.Time
		lastScanAt = &t
	}

	var lastScanStatus *string
	if f.LastScanStatus.Valid {
		s := f.LastScanStatus.String
		lastScanStatus = &s
	}

	var lastScanError *string
	if f.LastScanError.Valid {
		s := f.LastScanError.String
		lastScanError = &s
	}

	return FolderDTO{
		ID:             f.ID,
		Path:           f.Path,
		DeletedAt:      deletedAt,
		Available:      f.Available == 1,
		LastSeenAt:     lastSeenAt,
		LastScanAt:     lastScanAt,
		LastScanStatus: lastScanStatus,
		LastScanError:  lastScanError,
		CreatedAt:      f.CreatedAt,
		UpdatedAt:      f.UpdatedAt,
	}
}

func foldersDTOFromDB(rows []db.Folder) []FolderDTO {
	out := make([]FolderDTO, 0, len(rows))
	for _, f := range rows {
		out = append(out, folderDTOFromDB(f))
	}
	return out
}

func trackDTOFromDB(tk db.Track) TrackDTO {

	return TrackDTO{
		ID:           tk.ID,
		FolderID:     tk.FolderID,
		RelPath:      tk.RelPath,
		Filename:     tk.Filename,
		Ext:          tk.Ext,
		SizeBytes:    tk.SizeBytes,
		LastModified: tk.LastModified,
		LastSeenAt:   tk.LastSeenAt,
		DeletedAt:    timePtrFromNullTime(tk.DeletedAt),
		CreatedAt:    tk.CreatedAt,
		UpdatedAt:    tk.UpdatedAt,
	}
}

func tracksDTOFromDB(rows []db.Track) []TrackDTO {
	out := make([]TrackDTO, 0, len(rows))
	for _, f := range rows {
		out = append(out, trackDTOFromDB(f))
	}
	return out
}
