package server

import (
	"encoding/base64"

	"github.com/ricochhet/serve/internal/serverutil"
)

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
						Base64: base64.StdEncoding.EncodeToString(c.FS.Bytes("index.html")),
					},
					{
						Route:  "/404.html",
						Name:   "404.html",
						Base64: base64.StdEncoding.EncodeToString(c.FS.Bytes("404.html")),
					},
					{
						Route:  "/base.css",
						Name:   "base.css",
						Base64: base64.StdEncoding.EncodeToString(c.FS.Bytes("base.css")),
					},
				},
			},
		},
	}
}
