package embedx

import (
	"embed"
	"path/filepath"
	"strings"

	"github.com/ricochhet/serve/pkg/errorx"
	"github.com/ricochhet/serve/pkg/fsx"
)

type EmbeddedFileSystem struct {
	Root   string
	FS     embed.FS
	Prefix string
}

// List return a list of files within the embedded fs, calling the list function in return.
func (e *EmbeddedFileSystem) List(pattern string, list func([]fsx.File) error) error {
	files, err := fsx.WalkDir(e.FS, pattern)
	if err != nil {
		return errorx.WithFrame(err)
	}

	return list(files)
}

// Dump return a []byte within the embedded fs, calling the dump function in return.
func (e *EmbeddedFileSystem) Dump(pattern, name string, dump func(fsx.File, []byte) error) error {
	files, err := fsx.WalkDir(e.FS, pattern)
	if err != nil {
		return errorx.New("fsx.WalkDir", err)
	}

	for _, file := range files {
		if file.Info.IsDir() {
			continue
		}

		if !strings.Contains(file.Path, name) {
			continue
		}

		data, err := e.Read(strings.TrimPrefix(file.Path, e.Root))
		if err != nil {
			return errorx.New("e.Read", err)
		}

		if err := dump(file, data); err != nil {
			return errorx.New("dump", err)
		}
	}

	return nil
}

// MaybeReadEmbedded reads the filename from a path, falling back to embed if it does not exist.
func (e *EmbeddedFileSystem) MaybeReadEmbedded(name string) ([]byte, error) {
	path := filepath.ToSlash(filepath.Join(e.Root, name))
	if fsx.Exists(path) {
		return fsx.Read(path)
	}

	return e.Read(name)
}

// Bytes reads a file from the embedded filesystem, returning a new byte slice if it cannot be read.
func (e *EmbeddedFileSystem) Bytes(name string) []byte {
	b, _ := e.Read(name)
	if b == nil {
		return []byte{}
	}

	return b
}

// Read reads a file from the embedded filesystem as an array of bytes.
func (e *EmbeddedFileSystem) Read(name string) ([]byte, error) {
	b, err := e.FS.ReadFile(filepath.ToSlash(filepath.Join(e.Root, name)))
	if err != nil {
		return nil, errorx.WithFrame(err)
	}

	return b, nil
}
