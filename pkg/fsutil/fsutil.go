package fsutil

import (
	"os"
	"path/filepath"

	"github.com/ricochhet/serve/pkg/errutil"
)

// Read reads a file from the specified path.
func Read(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errutil.WithFrame(err)
	}

	return data, nil
}

// Write writes to the specified path with the provided data.
func Write(path string, data []byte) error {
	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return errutil.New("os.MkdirAll", err)
	}

	err = os.WriteFile(path, data, 0o644)
	if err != nil {
		return errutil.New("os.WriteFile", err)
	}

	return nil
}

// Exists returns true if a file exists.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
