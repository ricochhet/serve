package server

import (
	"github.com/ricochhet/serve/internal/config"
	"github.com/ricochhet/serve/internal/serverutil"
)

func (c *Context) AddHosts(h *serverutil.Hosts, cfg *config.Config) error {
	if !c.Hosts {
		return nil
	}

	return h.AddHosts(cfg)
}

func (c *Context) RemoveHosts(h *serverutil.Hosts, cfg *config.Config) error {
	if !c.Hosts {
		return nil
	}

	return h.RemoveHosts(cfg)
}
