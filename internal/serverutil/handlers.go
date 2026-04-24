package serverutil

import (
	"bytes"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/ricochhet/serve/internal/config"
	"github.com/ricochhet/serve/pkg/logx"
)

type headerWriter struct {
	http.ResponseWriter

	statusCode int
	allowed    map[string]struct{}
}

// WriteHeader strips non-allowed headers when writing to the actual header.
func (h *headerWriter) WriteHeader(code int) {
	hdr := h.Header()
	for key := range hdr {
		if _, ok := h.allowed[http.CanonicalHeaderKey(key)]; !ok {
			hdr.Del(key)
		}
	}

	if h.statusCode != 0 {
		code = h.statusCode
	}

	h.ResponseWriter.WriteHeader(code)
}

// newHeaderWriter sets the allowed headers, returning a new headerWriter.
func newHeaderWriter(
	w http.ResponseWriter,
	name string,
	data []byte,
	info config.Info,
) *headerWriter {
	allowed := make(map[string]struct{})
	setContentType := false

	for key, value := range info.Headers {
		canonical := http.CanonicalHeaderKey(key)
		w.Header().Set(canonical, value)
		allowed[canonical] = struct{}{}

		if canonical == "Content-Type" {
			setContentType = true
		}
	}

	if !setContentType {
		ct := mime.TypeByExtension(filepath.Ext(name))
		if ct == "" {
			ct = http.DetectContentType(data)
		}

		if ct != "" {
			w.Header().Set("Content-Type", ct)
		}

		allowed["Content-Type"] = struct{}{}
	}

	return &headerWriter{
		ResponseWriter: w,
		statusCode:     info.StatusCode,
		allowed:        allowed,
	}
}

// WithLogging is a middleware that logs the method and URL path for the handler.
func WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logx.Infof("%s %s\n", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// ServeFileHandler creates a Handler for http.ServeFile.
func (m *Map) ServeFileHandler(info config.Info, name string, serveFile bool) http.Handler {
	if serveFile {
		return serveFileHandler(info, name)
	}

	return m.serveFileHandler(info, name)
}

// ServeContentHandler creates a Handler for http.ServeContent.
func (m *Map) ServeContentHandler(info config.Info, name string, data []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.serveContent(w, r, info, name, data)
	})
}

func serveFileHandler(info config.Info, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(newHeaderWriter(w, name, nil, info), r, name)
	})
}

func (m *Map) serveFileHandler(info config.Info, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile(name)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		m.serveContent(w, r, info, name, data)
	})
}

func (m *Map) serveContent(
	w http.ResponseWriter,
	r *http.Request,
	info config.Info,
	name string,
	data []byte,
) {
	data = m.Parse(data)
	http.ServeContent(
		newHeaderWriter(w, name, data, info),
		r,
		name,
		time.Now(),
		bytes.NewReader(data),
	)
}
