package serverutil

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ricochhet/serve/cmd/serve/internal/configutil"
	"github.com/ricochhet/serve/pkg/contextutil"
)

type HTTPServer struct {
	Router chi.Router

	TLS      *configutil.TLS
	Timeouts *configutil.Timeouts
}

type Context struct {
	*contextutil.Context[HTTPServer]
}

func NewContext() *Context {
	return &Context{
		&contextutil.Context[HTTPServer]{},
	}
}

func (h *Context) Handle(pattern string, handler http.Handler) {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()

	h.Get().Router.Handle(pattern, handler)
}
