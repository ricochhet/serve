package server

import (
	"encoding/base64"
	"os"
	"strings"

	"github.com/ricochhet/serve/internal/serverutil"
	"github.com/ricochhet/serve/pkg/embedx"
	"github.com/ricochhet/serve/pkg/errorx"
	"github.com/ricochhet/serve/pkg/logx"
)

// maybeBase64 checks if name contains a prefix associated with embedded files. If it does, return the embedded file as base64.
func maybeBase64(fs *embedx.EmbeddedFileSystem, name string) ([]byte, error) {
	if after, ok := strings.CutPrefix(name, "asset:"); ok {
		return maybeRead(fs, after), nil
	}

	b, err := base64.RawStdEncoding.DecodeString(name)
	if err != nil {
		return nil, errorx.WithFrame(err)
	}

	return b, nil
}

// maybeRead reads the specified name from the embedded filesystem. If it cannot be read, the program will exit.
func maybeRead(fs *embedx.EmbeddedFileSystem, name string) []byte {
	b, err := fs.Read(name)
	if err != nil {
		logx.Errorf(logx.Get(), "Error reading from embedded filesystem: %v\n", err)
		os.Exit(1)
	}

	return b
}

// newDefaultConfig creates a default Config with the embedded index bytes.
func (c *Context) newDefaultConfig() *serverutil.Config {
	return &serverutil.Config{
		Servers: []serverutil.Server{
			{
				Port: 8080,
				ContentEntries: []serverutil.ContentEntry{
					{
						Route:  "/",
						Name:   "index.html",
						Base64: base64.StdEncoding.EncodeToString(maybeRead(c.FS, "index.html")),
					},
					{
						Route:  "/404.html",
						Name:   "404.html",
						Base64: base64.StdEncoding.EncodeToString(maybeRead(c.FS, "404.html")),
					},
					{
						Route:  "/base.css",
						Name:   "base.css",
						Base64: base64.StdEncoding.EncodeToString(maybeRead(c.FS, "base.css")),
					},
				},
			},
		},
	}
}
