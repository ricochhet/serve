package main

import (
	"embed"
	"path/filepath"

	"github.com/ricochhet/serve/pkg/embedx"
)

//go:embed wwwroot/serve/*
var webFS embed.FS

func Embed() *embedx.EmbeddedFileSystem {
	return &embedx.EmbeddedFileSystem{
		Root: filepath.ToSlash(filepath.Join("wwwroot", "serve")),
		FS:   webFS,
	}
}
