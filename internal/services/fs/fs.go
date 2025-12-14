package fs

import (
	"io/fs"
	"os"
	"path/filepath"
)

type FS interface {
	Stat(name string) (fs.FileInfo, error)
	WalkDir(root string, fn fs.WalkDirFunc) error
}

type OSFS struct{}

func (OSFS) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (OSFS) WalkDir(root string, fn fs.WalkDirFunc) error {
	return filepath.WalkDir(root, fn)
}
