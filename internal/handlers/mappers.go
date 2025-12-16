package handlers

import (
	"time"

	"bottomley.ian/musicserver/internal/db"
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
	return trackDTOFromParts(tk, db.Artist{}, db.Album{}, db.Artist{})
}

func trackDTOFromJoinedRow(row db.GetTrackWithJoinsRow) TrackDTO {
	return trackDTOFromParts(row.Track, row.Artist, row.Album, row.Artist_2)
}

func trackDTOFromParts(tk db.Track, artist db.Artist, album db.Album, albumArtist db.Artist) TrackDTO {
	return TrackDTO{
		ID:           tk.ID,
		FolderID:     tk.FolderID,
		ArtistID:     int64PtrFromNullInt64(tk.ArtistID),
		AlbumID:      int64PtrFromNullInt64(tk.AlbumID),
		Artist:       artistSummaryFromArtist(artist),
		Album:        albumSummaryFromAlbum(album, albumArtist),
		RelPath:      tk.RelPath,
		Title:        tk.Title,
		Filename:     tk.Filename,
		Ext:          tk.Ext,
		Genre:        stringPtrFromNullString(tk.Genre),
		Year:         int64PtrFromNullInt64(tk.Year),
		Rating:       int64PtrFromNullInt64(tk.Rating),
		ImagePath:    stringPtrFromNullString(tk.ImagePath),
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

func tracksDTOFromPlayableRows(rows []db.ListPlayableTracksWithJoinsRow) []TrackDTO {
	out := make([]TrackDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, trackDTOFromParts(row.Track, row.Artist, row.Album, row.Artist_2))
	}
	return out
}

func tracksDTOFromAlbumRows(rows []db.ListPlayableTracksForAlbumRow) []TrackDTO {
	out := make([]TrackDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, trackDTOFromParts(row.Track, row.Artist, row.Album, row.Artist_2))
	}
	return out
}

func tracksDTOFromBase(rows []db.Track) []TrackDTO {
	out := make([]TrackDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, trackDTOFromDB(row))
	}
	return out
}

func artistDTOFromDB(a db.Artist) ArtistDTO {
	return ArtistDTO{
		ID:        a.ID,
		Name:      a.Name,
		DeletedAt: timePtrFromNullTime(a.DeletedAt),
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

func artistsDTOFromDB(rows []db.Artist) []ArtistDTO {
	out := make([]ArtistDTO, 0, len(rows))
	for _, a := range rows {
		out = append(out, artistDTOFromDB(a))
	}
	return out
}

func albumDTOFromDB(al db.Album) AlbumDTO {
	return albumDTOFromParts(al, db.Artist{})
}

func albumDTOFromParts(al db.Album, artist db.Artist) AlbumDTO {
	return AlbumDTO{
		ID:        al.ID,
		ArtistID:  al.ArtistID,
		Artist:    artistSummaryFromArtist(artist),
		Title:     al.Title,
		ImagePath: stringPtrFromNullString(al.ImagePath),
		DeletedAt: timePtrFromNullTime(al.DeletedAt),
		CreatedAt: al.CreatedAt,
		UpdatedAt: al.UpdatedAt,
	}
}

func albumsDTOFromDB(rows []db.Album) []AlbumDTO {
	out := make([]AlbumDTO, 0, len(rows))
	for _, al := range rows {
		out = append(out, albumDTOFromDB(al))
	}
	return out
}

func albumsDTOFromRows(rows []db.ListAlbumsWithArtistRow) []AlbumDTO {
	out := make([]AlbumDTO, 0, len(rows))
	for _, al := range rows {
		out = append(out, albumDTOFromParts(al.Album, al.Artist))
	}
	return out
}

func playlistDTOFromDB(p db.Playlist) PlaylistDTO {
	return PlaylistDTO{
		ID:        p.ID,
		Name:      p.Name,
		DeletedAt: timePtrFromNullTime(p.DeletedAt),
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func playlistsDTOFromDB(rows []db.Playlist) []PlaylistDTO {
	out := make([]PlaylistDTO, 0, len(rows))
	for _, p := range rows {
		out = append(out, playlistDTOFromDB(p))
	}
	return out
}

func playlistTrackDTOFromRow(pt db.ListPlaylistTracksRow) PlaylistTrackDTO {
	track := trackDTOFromParts(pt.Track, pt.Artist, pt.Album, pt.Artist_2)

	return PlaylistTrackDTO{
		ID:         pt.PlaylistTrack.ID,
		PlaylistID: pt.PlaylistTrack.PlaylistID,
		TrackID:    pt.PlaylistTrack.TrackID,
		Position:   pt.PlaylistTrack.Position,
		DeletedAt:  timePtrFromNullTime(pt.PlaylistTrack.DeletedAt),
		CreatedAt:  pt.PlaylistTrack.CreatedAt,
		UpdatedAt:  pt.PlaylistTrack.UpdatedAt,
		Track:      &track,
	}
}

func playlistTracksDTOFromRows(rows []db.ListPlaylistTracksRow) []PlaylistTrackDTO {
	out := make([]PlaylistTrackDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, playlistTrackDTOFromRow(row))
	}
	return out
}

func playlistTrackDTOFromPT(pt db.PlaylistTrack, track *TrackDTO) PlaylistTrackDTO {
	var trackDTO *TrackDTO
	if track != nil {
		trackDTO = track
	}
	return PlaylistTrackDTO{
		ID:         pt.ID,
		PlaylistID: pt.PlaylistID,
		TrackID:    pt.TrackID,
		Position:   pt.Position,
		DeletedAt:  timePtrFromNullTime(pt.DeletedAt),
		CreatedAt:  pt.CreatedAt,
		UpdatedAt:  pt.UpdatedAt,
		Track:      trackDTO,
	}
}

func artistSummaryFromArtist(ar db.Artist) *ArtistSummaryDTO {
	if ar.ID == 0 {
		return nil
	}
	return &ArtistSummaryDTO{
		ID:   ar.ID,
		Name: ar.Name,
	}
}

func albumSummaryFromAlbum(al db.Album, artist db.Artist) *AlbumSummaryDTO {
	if al.ID == 0 {
		return nil
	}
	return &AlbumSummaryDTO{
		ID:        al.ID,
		ArtistID:  al.ArtistID,
		Title:     al.Title,
		ImagePath: stringPtrFromNullString(al.ImagePath),
	}
}
