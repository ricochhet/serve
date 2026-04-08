package serverutil

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ricochhet/serve/pkg/syncx"
)

type HTTPServer struct {
	Router chi.Router

	TLS      *TLS
	Timeouts *Timeouts
}

type Safe struct {
	*syncx.Safe[HTTPServer]
}

func NewSafe() *Safe {
	return &Safe{
		&syncx.Safe[HTTPServer]{},
	}
}

func (s *Safe) Handle(pattern string, handler http.Handler) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.Get().Router.Handle(pattern, handler)
}
