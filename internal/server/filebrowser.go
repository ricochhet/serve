package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ricochhet/serve/internal/config"
	"github.com/ricochhet/serve/internal/serverutil"
	"github.com/ricochhet/serve/pkg/errorx"
	"github.com/ricochhet/serve/pkg/logx"
)

type FileBrowserEntry struct {
	Name    string    `json:"name"`
	IsDir   bool      `json:"isDir"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modTime"`
}

type FileBrowserResponse struct {
	Path    string             `json:"path"`
	Entries []FileBrowserEntry `json:"entries"`
}

func (c *Context) registerFileBrowser(srv *serverutil.HTTPServer, cfg *config.Server) error {
	f := cfg.FileBrowser
	if !f.Enabled {
		return nil
	}

	route := f.Route
	if route == "" {
		route = "/files"
	}

	root := f.Root
	if root == "" {
		root = "."
	}

	abs, err := filepath.Abs(root)
	if err != nil {
		return errorx.WithFramef("invalid root %s: %w", root, err)
	}

	if info, err := os.Stat(abs); err != nil || !info.IsDir() {
		return errorx.WithFramef("root does not exist or is not a directory: %s", abs)
	}

	logx.Infof("Port %d: file browser %s -> %s\n", cfg.Port, route, abs)

	srv.Handle(route, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(c.FS.Bytes("filebrowser.html"))
	}))

	srv.Handle(route+"/api", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.fileBrowserAPI(w, r, abs)
	}))

	return nil
}

func (c *Context) fileBrowserAPI(w http.ResponseWriter, r *http.Request, root string) {
	req := r.URL.Query().Get("path")
	if req == "" {
		req = "/"
	}

	path := filepath.ToSlash(filepath.Clean(req))
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	full := filepath.Join(root, filepath.FromSlash(path))
	if !strings.HasPrefix(full+string(filepath.Separator), root+string(filepath.Separator)) {
		http.Error(w, "StatusForbidden", http.StatusForbidden)
		return
	}

	dir, err := os.ReadDir(full)
	if err != nil {
		http.Error(w, "StatusNotFound", http.StatusNotFound)
		return
	}

	res := FileBrowserResponse{
		Path:    path,
		Entries: make([]FileBrowserEntry, 0, len(dir)),
	}

	var dirs, files []FileBrowserEntry

	for _, d := range dir {
		i, err := d.Info()
		if err != nil {
			continue
		}

		e := FileBrowserEntry{
			Name:    d.Name(),
			IsDir:   d.IsDir(),
			Size:    i.Size(),
			ModTime: i.ModTime().UTC(),
		}

		if d.IsDir() {
			dirs = append(dirs, e)
		} else {
			files = append(files, e)
		}
	}

	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].Name < dirs[j].Name
	})

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})

	res.Entries = append(dirs, files...)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}
