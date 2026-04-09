package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/ricochhet/serve/internal/serverutil"
	"github.com/ricochhet/serve/pkg/embedx"
	"github.com/ricochhet/serve/pkg/errorx"
	"github.com/ricochhet/serve/pkg/fsx"
	"github.com/ricochhet/serve/pkg/jsonx"
	"github.com/ricochhet/serve/pkg/logx"
)

type Context struct {
	ConfigFile string
	Hosts      bool
	TLS        *serverutil.TLS
	FS         *embedx.EmbeddedFileSystem

	servers []*http.Server
}

// NewServer returns a new Server type with assets preloaded.
func NewServer(
	configFile string,
	hosts bool,
	tls *serverutil.TLS,
	fs *embedx.EmbeddedFileSystem,
) *Context {
	s := &Context{}
	if configFile != "" {
		s.ConfigFile = configFile
	}

	s.Hosts = hosts
	s.TLS = tls
	s.FS = fs

	return s
}

// StartServer starts an HTTP server with the specified server configuration.
func (c *Context) StartServer() error {
	config, err := c.maybeReadConfig()
	if err != nil {
		return errorx.New("c.maybeReadConfig", err)
	}

	if err := c.addHosts(config); err != nil {
		return errorx.New("c.addHosts", err)
	}

	c.maybeTLS(config)

	for _, cfg := range config.Servers {
		ctx := serverutil.NewSafe()

		maxAge := 300
		if cfg.MaxAge != 0 {
			maxAge = cfg.MaxAge
		}

		r := chi.NewRouter()
		r.Use(middleware.Recoverer)
		r.Use(serverutil.WithLogging)
		r.Use(httprate.LimitByIP(50, time.Minute))
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{"https://*", "http://*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: cfg.AllowCredentials,
			MaxAge:           maxAge,
		}))

		r.NotFound(c.NotFoundHandler)

		ctx.SetLocked(&serverutil.HTTPServer{
			Router:   r,
			TLS:      c.TLS,
			Timeouts: &cfg.Timeouts,
		})

		if err := c.startServer(ctx, &cfg); err != nil {
			return errorx.New("c.startServer", err)
		}
	}

	if err := c.removeHosts(config); err != nil {
		return errorx.New("c.removeHosts", err)
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

// addHosts adds the specified hosts from the configuration.
func (c *Context) addHosts(cfg *serverutil.Config) error {
	if !c.isHostsValid(cfg) {
		return nil
	}

	h, err := serverutil.NewHosts()
	if err != nil {
		return errorx.New("serverutil.NewHosts", err)
	}

	return h.AddMap(cfg.Hosts)
}

// removeHosts removes the specified hosts from the configuration.
func (c *Context) removeHosts(cfg *serverutil.Config) error {
	if !c.isHostsValid(cfg) {
		return nil
	}

	h, err := serverutil.NewHosts()
	if err != nil {
		return errorx.New("serverutil.NewHosts", err)
	}

	return h.RemoveMap(cfg.Hosts)
}

// isHostsValid returns if the hosts state is valid.
func (c *Context) isHostsValid(cfg *serverutil.Config) bool {
	return c.Hosts && cfg.Hosts != nil && len(cfg.Hosts) != 0
}

// maybeTLS sets TLS based on whether flags are based, or if relevant config options are used.
func (c *Context) maybeTLS(cfg *serverutil.Config) {
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

// maybeReadConfig reads the file path if it exists, otherwise returning a default config.
func (c *Context) maybeReadConfig() (*serverutil.Config, error) {
	var (
		config *serverutil.Config
		err    error
	)

	exists := fsx.Exists(c.ConfigFile)
	switch {
	case exists:
		config, err = jsonx.ReadAndUnmarshal[serverutil.Config](c.ConfigFile)
		if err != nil {
			logx.Errorf("Error reading server config: %v\n", err)
		}

		return config, err
	case !exists && c.ConfigFile != "":
		return nil, fmt.Errorf("path specified but does not exist: %s", c.ConfigFile)
	default:
		logx.Infof("Starting with default server config\n")
		return c.newDefaultConfig(), nil
	}
}

// startServer starts an HTTP server with the specified server configuration.
func (c *Context) startServer(ctx *serverutil.Safe, cfg *serverutil.Server) error {
	if err := serveFileHandler(ctx, cfg); err != nil {
		return errorx.New("serveFileHandler", err)
	}

	if err := c.serveContentHandler(ctx, cfg); err != nil {
		return errorx.New("c.serveContentHandler", err)
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := ctx.ListenAndServe(addr)

	c.servers = append(c.servers, srv)

	return nil
}

// serveContentHandler handles the ServeFileHandler for each file entry.
func serveFileHandler(ctx *serverutil.Safe, cfg *serverutil.Server) error {
	for _, f := range cfg.FileEntries {
		info, err := os.Stat(f.Path)
		if err != nil {
			return errorx.WithFramef("invalid path %s: %w", f.Path, err)
		}

		if info.IsDir() {
			if err := matchPattern(f, ctx, cfg); err != nil {
				return errorx.New("matchPattern", err)
			}
		} else {
			if err := matchFile(f, ctx, cfg); err != nil {
				return errorx.New("matchFile", err)
			}
		}
	}

	return nil
}

// matchPattern handles file paths that contain glob information.
func matchPattern(
	f serverutil.FileEntry,
	ctx *serverutil.Safe,
	cfg *serverutil.Server,
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
		ctx.Handle(route, serverutil.ServeFileHandler(f.Info, abs))

		return nil
	})
}

// matchFile handles absolute file paths.
func matchFile(
	f serverutil.FileEntry,
	ctx *serverutil.Safe,
	cfg *serverutil.Server,
) error {
	abs, err := filepath.Abs(f.Path)
	if err != nil {
		return errorx.WithFramef("invalid path %s: %w", f.Path, err)
	}

	logx.Infof("Port %d: %s -> %s\n", cfg.Port, f.Route, abs)
	ctx.Handle(f.Route, serverutil.ServeFileHandler(f.Info, abs))

	return nil
}

// serveContentHandler handles the ServeContentHandler for each content entry.
func (c *Context) serveContentHandler(ctx *serverutil.Safe, cfg *serverutil.Server) error {
	for _, f := range cfg.ContentEntries {
		logx.Infof(
			"Port %d: %s -> %s (%d)\n",
			cfg.Port,
			f.Route,
			f.Name,
			len(f.Base64),
		)

		b, err := c.FS.ReadBase64(f.Base64)
		if err != nil {
			return errorx.WithFrame(err)
		}

		ctx.Handle(f.Route, serverutil.ServeContentHandler(f.Info, f.Name, b))
	}

	return nil
}
