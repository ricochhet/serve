package fsx

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ricochhet/serve/pkg/errorx"
)

type File struct {
	Path string
	Info fs.FileInfo
}

// Read reads a file from the specified path.
func Read(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errorx.WithFrame(err)
	}

	return data, nil
}

// Write writes to the specified path with the provided data.
func Write(path string, data []byte) error {
	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return errorx.New("os.MkdirAll", err)
	}

	err = os.WriteFile(path, data, 0o644)
	if err != nil {
		return errorx.New("os.WriteFile", err)
	}

	return nil
}

// Exists returns true if a file exists.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// WalkDir walks the directory starting at the specified root path.
func WalkDir(e fs.FS, root string) ([]File, error) {
	result := []File{}

	err := fs.WalkDir(e, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return errorx.New("fs.WalkDir", err)
		}

		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return errorx.New("d.Info", err)
			}

			result = append(result, File{Path: path, Info: info})
		}

		return nil
	})
	if err != nil {
		return nil, errorx.WithFrame(err)
	}

	return result, nil
}
