package filebrowser

import (
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/ricochhet/serve/internal/config"
	"github.com/ricochhet/serve/internal/serverutil"
	"github.com/ricochhet/serve/pkg/embedx"
	"github.com/ricochhet/serve/pkg/errorx"
	"github.com/ricochhet/serve/pkg/logx"
	"github.com/ricochhet/serve/pkg/templatex"
)

type Entry struct {
	Name    string
	IsDir   bool
	Size    int64
	ModTime time.Time
	NavPath string
}

type Crumb struct {
	Label string
	Path  string
}

type Page struct {
	Path    string
	Crumbs  []Crumb
	Entries []Entry
	Error   string
}

var (
	tmplx   = templatex.New()
	tmpl, _ = tmplx.GetTemplate("filebrowser")
)

func Register(fs *embedx.EmbeddedFileSystem, srv *serverutil.HTTPServer, cfg *config.Server) error {
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

	tmplx.Register(&templatex.Template{
		FS:   fs,
		Name: "filebrowser",
		Path: "filebrowser.html",
		FuncMap: &template.FuncMap{
			"formatSize": formatSize,
			"formatDate": formatDate,
			"parentPath": parentOf,
		},
	})

	abs, err := filepath.Abs(root)
	if err != nil {
		return errorx.WithFramef("invalid root %s: %w", root, err)
	}

	if info, err := os.Stat(abs); err != nil || !info.IsDir() {
		return errorx.WithFramef("root does not exist or is not a directory: %s", abs)
	}

	logx.Infof("Port %d: file browser %s -> %s\n", cfg.Port, route, abs)

	srv.Handle(route, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page(w, r, abs)
	}))

	return nil
}

func page(w http.ResponseWriter, r *http.Request, root string) {
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
		templatex.Render(tmpl, w, http.StatusForbidden, Page{
			Path:   path,
			Crumbs: breadCrumbs(path),
			Error:  "Forbidden",
		})

		return
	}

	dir, err := os.ReadDir(full)
	if err != nil {
		templatex.Render(tmpl, w, http.StatusNotFound, Page{
			Path:   path,
			Crumbs: breadCrumbs(path),
			Error:  "Directory not found",
		})

		return
	}

	var dirs, files []Entry

	for _, d := range dir {
		i, err := d.Info()
		if err != nil {
			continue
		}

		var navPath string
		if path == "/" {
			navPath = "/" + d.Name()
		} else {
			navPath = path + "/" + d.Name()
		}

		e := Entry{
			Name:    d.Name(),
			IsDir:   d.IsDir(),
			Size:    i.Size(),
			ModTime: i.ModTime().UTC(),
			NavPath: navPath,
		}

		if d.IsDir() {
			dirs = append(dirs, e)
		} else {
			files = append(files, e)
		}
	}

	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Name < dirs[j].Name })
	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })

	templatex.Render(tmpl, w, http.StatusOK, Page{
		Path:    path,
		Crumbs:  breadCrumbs(path),
		Entries: slices.Concat(dirs, files),
	})
}
