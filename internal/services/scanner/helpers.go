package scanner

import (
	"os"
	"path/filepath"
	"strings"
)

func expandPath(p string) (string, error) {
	p = strings.TrimSpace(p)

	// Expand "~" and "~/" to the user's home dir
	if p == "~" || strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if p == "~" {
			p = home
		} else {
			p = filepath.Join(home, p[2:])
		}
	}

	// Clean and make absolute (optional but recommended)
	p = filepath.Clean(p)
	if !filepath.IsAbs(p) {
		abs, err := filepath.Abs(p)
		if err != nil {
			return "", err
		}
		p = abs
	}
	return p, nil
}
