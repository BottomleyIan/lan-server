package fs

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type FS interface {
	Stat(name string) (fs.FileInfo, error)
	WalkDir(root string, fn fs.WalkDirFunc) error
	Open(name string) (fs.File, error)
	ReadDir(name string) ([]fs.DirEntry, error)
	MkdirAll(path string, perm fs.FileMode) error
	Create(name string) (io.WriteCloser, error)
	WriteFile(name string, data []byte, perm fs.FileMode) error
	ReadFile(name string) ([]byte, error)
}

type OSFS struct{}

func (OSFS) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (OSFS) WalkDir(root string, fn fs.WalkDirFunc) error {
	return filepath.WalkDir(root, fn)
}

func (OSFS) Open(name string) (fs.File, error) {
	return os.Open(name)
}

func (OSFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(name)
}

func (OSFS) MkdirAll(path string, perm fs.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (OSFS) Create(name string) (io.WriteCloser, error) {
	return os.Create(name)
}

func (OSFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}

func (OSFS) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func ExpandUserPath(path string) (string, error) {
	if path == "" {
		return path, nil
	}
	if path == "~" || strings.HasPrefix(path, "~"+string(os.PathSeparator)) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if path == "~" {
			return home, nil
		}
		return filepath.Join(home, strings.TrimPrefix(path, "~"+string(os.PathSeparator))), nil
	}
	return path, nil
}

func ExpandPath(path string) (string, error) {
	p := strings.TrimSpace(path)
	expanded, err := ExpandUserPath(p)
	if err != nil {
		return "", err
	}
	p = filepath.Clean(expanded)
	if !filepath.IsAbs(p) {
		abs, err := filepath.Abs(p)
		if err != nil {
			return "", err
		}
		p = abs
	}
	return p, nil
}
