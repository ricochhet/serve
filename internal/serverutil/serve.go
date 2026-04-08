package serverutil

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ricochhet/serve/pkg/logx"
)

// listenAndServe creates an HTTP server at the specified address.
func (s *Safe) ListenAndServe(addr string) *http.Server {
	server := &http.Server{
		Addr:              addr,
		Handler:           s.Get().Router,
		ReadHeaderTimeout: time.Duration(s.Get().Timeouts.ReadHeader) * time.Second,
		ReadTimeout:       time.Duration(s.Get().Timeouts.Read) * time.Second,
		WriteTimeout:      time.Duration(s.Get().Timeouts.Write) * time.Second,
		IdleTimeout:       time.Duration(s.Get().Timeouts.Idle) * time.Second,
	}

	logx.Infof(logx.Get(), "Server listening on %s\n", addr)

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
				logx.Get(),
				"Server %s failed: %v\n",
				strings.TrimPrefix(addr, ":"),
				err,
			)
		}
	}()

	return server
}
