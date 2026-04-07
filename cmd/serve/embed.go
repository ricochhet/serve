package main

import (
	"embed"
	"path/filepath"

	"github.com/ricochhet/serve/pkg/embedutil"
)

//go:embed wwwroot/serve/*
var webFS embed.FS

func Embed() *embedutil.EmbeddedFileSystem {
	return &embedutil.EmbeddedFileSystem{
		Initial: filepath.ToSlash(filepath.Join("wwwroot", "serve")),
		FS:      webFS,
	}
}
