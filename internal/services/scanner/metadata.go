package scanner

import (
	//"errors"
	"os"
	//"path/filepath"
	//"strings"

	"github.com/dhowden/tag"
	//"github.com/mewkiz/flac"
	//"github.com/mewkiz/flac/meta"
	//"github.com/bogem/id3v2"
)

type Picture struct {
	MIME        string
	Description string
	Data        []byte
}

type Metadata struct {
	Title  string
	Artist string
	Album  string
	Genre  string
	Year   int
	Track  int

	Picture *Picture
}

func (s *Scanner) ReadMetadata(path string) (Metadata, error) {
	var out Metadata

	// 1) Generic tag reader for common fields
	f, err := os.Open(path)
	if err != nil {
		return out, err
	}
	defer f.Close()

	m, err := tag.ReadFrom(f)
	if err == nil {
		out.Title = m.Title()
		out.Artist = m.Artist()
		out.Album = m.Album()
		out.Genre = m.Genre()
		out.Year = m.Year()
		out.Track, _ = m.Track()

		// sometimes works for both MP3/FLAC depending on tags
		if pic := m.Picture(); pic != nil && len(pic.Data) > 0 {
			out.Picture = &Picture{
				MIME:        pic.MIMEType,
				Description: pic.Description,
				Data:        pic.Data,
			}
		}
	}
	/*
		// 2) If picture missing, try format-specific fallback
		if out.Picture == nil {
			ext := strings.ToLower(filepath.Ext(path))
			switch ext {
			case ".mp3":
				if pic, e := readMP3Cover(path); e == nil && pic != nil {
					out.Picture = pic
				}
			case ".flac":
				if pic, e := readFLACCover(path); e == nil && pic != nil {
					out.Picture = pic
				}
			}
		}
	*/
	// If generic tag reading failed and we got nothing, return the error
	if err != nil && out.Title == "" && out.Artist == "" && out.Album == "" && out.Picture == nil {
		return out, err
	}
	return out, nil
}

/*
// --- MP3 cover art (ID3 APIC) ---

func readMP3Cover(path string) (*Picture, error) {
	tagFile, err := id3v2.Open(path, id3v2.Options{Parse: true})
	if err != nil {
		return nil, err
	}
	defer tagFile.Close()

	frames := tagFile.GetFrames(tag.CommonIDAttachedPicture) // "APIC"
	if len(frames) == 0 {
		return nil, errors.New("no APIC frames")
	}

	// Use first picture frame (often "Front Cover")
	for _, fr := range frames {
		apic, ok := fr.(id3v2.PictureFrame)
		if !ok || len(apic.Picture) == 0 {
			continue
		}
		return &Picture{
			MIME:        apic.MimeType,
			Description: apic.Description,
			Data:        apic.Picture,
		}, nil
	}
	return nil, errors.New("no usable APIC frame")
}

// --- FLAC cover art (PICTURE metadata block) ---

func readFLACCover(path string) (*Picture, error) {
	stream, err := flac.ParseFile(path)
	if err != nil {
		return nil, err
	}

	// Prefer front cover if present; otherwise take first picture.
	var first *meta.Picture
	var front *meta.Picture

	for _, block := range stream.Blocks {
		pic, ok := block.Body.(*meta.Picture)
		if !ok || pic == nil || len(pic.Data) == 0 {
			continue
		}
		if first == nil {
			first = pic
		}
		// 3 is "Cover (front)" in FLAC picture type enum
		if pic.Type == meta.PictureTypeCoverFront {
			front = pic
			break
		}
	}

	p := front
	if p == nil {
		p = first
	}
	if p == nil {
		return nil, errors.New("no FLAC picture blocks")
	}

	return &Picture{
		MIME:        p.MIME,
		Description: p.Description,
		Data:        p.Data,
	}, nil
}
*/
