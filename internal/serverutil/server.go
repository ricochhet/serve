package serverutil

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ricochhet/serve/internal/config"
	"github.com/ricochhet/serve/pkg/logx"
	"github.com/ricochhet/serve/pkg/syncx"
)

type httpServer struct {
	Router chi.Router

	TLS      *config.TLS
	Timeouts *config.Timeouts
}

type HTTPServer struct {
	*syncx.Safe[httpServer]
}

func New() *HTTPServer {
	return &HTTPServer{
		&syncx.Safe[httpServer]{},
	}
}

func (s *HTTPServer) New(router chi.Router, tls *config.TLS, timeouts *config.Timeouts) {
	s.SetLocked(&httpServer{
		Router:   router,
		TLS:      tls,
		Timeouts: timeouts,
	})
}

func (s *HTTPServer) Handle(pattern string, handler http.Handler) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.Get().Router.Handle(pattern, handler)
}

// listenAndServe creates an HTTP server at the specified address.
func (s *HTTPServer) ListenAndServe(addr string) *http.Server {
	server := &http.Server{
		Addr:              addr,
		Handler:           s.Get().Router,
		ReadHeaderTimeout: time.Duration(s.Get().Timeouts.ReadHeader) * time.Second,
		ReadTimeout:       time.Duration(s.Get().Timeouts.Read) * time.Second,
		WriteTimeout:      time.Duration(s.Get().Timeouts.Write) * time.Second,
		IdleTimeout:       time.Duration(s.Get().Timeouts.Idle) * time.Second,
	}

	logx.Infof("Server listening on %s\n", addr)

	go func() {
		var err error

		if s.Get().TLS.Enabled {
			fmt.Fprintf(
				os.Stdout,
				"Server starting with tls: %s (cert) and %s (key)\n",
				s.Get().TLS.CertFile, s.Get().TLS.KeyFile,
			)
			err = server.ListenAndServeTLS(s.Get().TLS.CertFile, s.Get().TLS.KeyFile)
		} else {
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			logx.Infof(
				"Server %s failed: %v\n",
				strings.TrimPrefix(addr, ":"),
				err,
			)
		}
	}()

	return server
}
