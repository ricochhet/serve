package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/ricochhet/serve/internal/config"
	"github.com/ricochhet/serve/internal/serverutil"
	"github.com/ricochhet/serve/pkg/embedx"
	"github.com/ricochhet/serve/pkg/errorx"
	"github.com/ricochhet/serve/pkg/fsx"
	"github.com/ricochhet/serve/pkg/logx"
)

type Context struct {
	ConfigFile string
	MapFile    string
	Hosts      bool
	TLS        *config.TLS
	FS         *embedx.EmbeddedFileSystem

	ServeFile bool
	servers   []*http.Server
}

var serverMap *serverutil.Map

// NewServer returns a new Server type with assets preloaded.
func NewServer(
	configFile string,
	mapFile string,
	hosts bool,
	tls *config.TLS,
	fs *embedx.EmbeddedFileSystem,
	serveFile bool,
) *Context {
	s := &Context{}
	if configFile != "" {
		s.ConfigFile = configFile
	}

	if mapFile != "" {
		s.MapFile = mapFile
	}

	s.Hosts = hosts
	s.TLS = tls
	s.FS = fs
	s.ServeFile = serveFile

	return s
}

// StartServer starts an HTTP server with the specified server configuration.
func (c *Context) StartServer() error {
	sm, err := config.Read[serverutil.Map](c.MapFile, true)
	if err != nil {
		return errorx.New("config.Read[Map]", err)
	}

	serverMap = sm

	config, err := config.Read[config.Config](c.ConfigFile, false)
	if err != nil {
		return errorx.New("config.Read[config.Config]", err)
	}

	h, err := serverutil.NewHosts()
	if err != nil {
		return errorx.New("serverutil.NewHosts", err)
	}

	if err := c.AddHosts(h, config); err != nil {
		return errorx.New("c.AddHosts", err)
	}

	c.maybeTLS(config)

	for _, cfg := range config.Servers {
		ctx := serverutil.New()

		ipLimit := 50
		if cfg.IPLimit != 0 {
			ipLimit = cfg.IPLimit
		}

		maxAge := 300
		if cfg.MaxAge != 0 {
			maxAge = cfg.MaxAge
		}

		r := chi.NewRouter()
		r.Use(middleware.Recoverer)
		r.Use(serverutil.WithLogging)
		r.Use(httprate.LimitByIP(ipLimit, time.Minute))
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{"https://*", "http://*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: cfg.AllowCredentials,
			MaxAge:           maxAge,
		}))

		r.NotFound(c.NotFoundHandler)

		ctx.New(r, c.TLS, &cfg.Timeouts)

		if err := c.startServer(ctx, &cfg); err != nil {
			return errorx.New("c.startServer", err)
		}
	}

	if err := c.RemoveHosts(h, config); err != nil {
		return errorx.New("c.RemoveHosts", err)
	}

	c.shutdown()

	return nil
}

// shutdown handles shutdown of all servers.
func (c *Context) shutdown() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, srv := range c.servers {
		if err := srv.Shutdown(ctx); err != nil {
			logx.Errorf("Error shutting down server: %v\n", err)
		}
	}
}

// maybeTLS sets TLS based on whether flags are based, or if relevant config options are used.
func (c *Context) maybeTLS(cfg *config.Config) {
	if c.TLS.CertFile == "" || c.TLS.KeyFile == "" { // default flags
		c.TLS.Enabled = false
	}

	if fsx.Exists(c.TLS.CertFile) && fsx.Exists(c.TLS.KeyFile) { // flags
		c.TLS.Enabled = true
		return
	}

	if fsx.Exists(cfg.TLS.CertFile) && fsx.Exists(cfg.TLS.KeyFile) { // config
		c.TLS = &cfg.TLS
	}
}

// startServer starts an HTTP server with the specified server configuration.
func (c *Context) startServer(srv *serverutil.HTTPServer, cfg *config.Server) error {
	if err := c.serveFileHandler(srv, cfg); err != nil {
		return errorx.New("serveFileHandler", err)
	}

	if err := c.registerFileBrowser(srv, cfg); err != nil {
		return errorx.New("registerFileBrowser", err)
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	c.servers = append(c.servers, srv.ListenAndServe(addr))

	return nil
}

// serveContentHandler handles the ServeFileHandler for each file entry.
func (c *Context) serveFileHandler(srv *serverutil.HTTPServer, cfg *config.Server) error {
	for _, f := range cfg.Files {
		if after, ok := strings.CutPrefix(f.Path, c.FS.Prefix); ok {
			f.Path = after
			if err := c.serveEmbeddedFile(f, srv, cfg); err != nil {
				return err
			}

			continue
		}

		if err := c.serveFile(f, srv, cfg); err != nil {
			return err
		}
	}

	return nil
}

// serveFile serves a file from the filesystem via ServeFileHandler.
func (c *Context) serveFile(f config.File, srv *serverutil.HTTPServer, cfg *config.Server) error {
	info, err := os.Stat(f.Path)
	if err != nil {
		return errorx.WithFramef("invalid path %s: %w", f.Path, err)
	}

	if info.IsDir() {
		if err := c.matchPattern(f, srv, cfg); err != nil {
			return errorx.New("matchPattern", err)
		}
	} else {
		if err := c.matchFile(f, srv, cfg); err != nil {
			return errorx.New("matchFile", err)
		}
	}

	return nil
}

// serveEmbeddedFile serves an embedded file via ServerContentHandler.
func (c *Context) serveEmbeddedFile(
	f config.File,
	srv *serverutil.HTTPServer,
	cfg *config.Server,
) error {
	b, err := c.FS.Read(f.Path)
	if err != nil {
		return errorx.New("c.FS.MaybeReadPrefixed", err)
	}

	route := filepath.ToSlash(f.Route)

	logx.Infof("Port %d: %s -> %s\n", cfg.Port, route, f.Route)
	srv.Handle(f.Route, serverMap.ServeContentHandler(f.Info, f.Path, b))

	return nil
}

// matchPattern handles file paths that contain glob information.
func (c *Context) matchPattern(
	f config.File,
	srv *serverutil.HTTPServer,
	cfg *config.Server,
) error {
	return filepath.Walk(f.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errorx.WithFrame(err)
		}

		if info.IsDir() {
			return nil
		}

		abs, err := filepath.Abs(path)
		if err != nil {
			return errorx.WithFramef("invalid path %s: %w", path, err)
		}

		rel, err := filepath.Rel(f.Path, path)
		if err != nil {
			return errorx.WithFramef("cannot get relative path for %s: %w", path, err)
		}

		route := filepath.ToSlash(filepath.Join(f.Route, rel))

		logx.Infof("Port %d: %s -> %s\n", cfg.Port, route, abs)
		srv.Handle(route, serverMap.ServeFileHandler(f.Info, abs, c.ServeFile))

		return nil
	})
}

// matchFile handles absolute file paths.
func (c *Context) matchFile(
	f config.File,
	srv *serverutil.HTTPServer,
	cfg *config.Server,
) error {
	abs, err := filepath.Abs(f.Path)
	if err != nil {
		return errorx.WithFramef("invalid path %s: %w", f.Path, err)
	}

	logx.Infof("Port %d: %s -> %s\n", cfg.Port, f.Route, abs)
	srv.Handle(f.Route, serverMap.ServeFileHandler(f.Info, abs, c.ServeFile))

	return nil
}
