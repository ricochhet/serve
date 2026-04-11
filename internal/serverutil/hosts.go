package serverutil

import (
	"path"
	"strings"

	"github.com/goodhosts/hostsfile"
	"github.com/ricochhet/serve/internal/config"
	"github.com/ricochhet/serve/pkg/errorx"
	"github.com/ricochhet/serve/pkg/fsx"
	"github.com/ricochhet/serve/pkg/logx"
	"github.com/ricochhet/serve/pkg/timex"
)

type Hosts struct {
	*hostsfile.Hosts
}

// NewHosts returns a new Hosts.
func NewHosts() (*Hosts, error) {
	h, err := hostsfile.NewHosts()
	if err != nil {
		return nil, errorx.New("hostsfile.NewHosts", err)
	}

	hosts := &Hosts{h}

	if err := hosts.backupHosts(); err != nil {
		return nil, errorx.New("hosts.backupHosts", err)
	}

	return hosts, nil
}

// AddHosts adds the specified hosts from the configuration.
func (h *Hosts) AddHosts(cfg *config.Config) error {
	if len(cfg.Hosts) == 0 {
		return nil
	}

	return h.addMap(cfg.Hosts)
}

// RemoveHosts removes the specified hosts from the configuration.
func (h *Hosts) RemoveHosts(cfg *config.Config) error {
	if len(cfg.Hosts) == 0 {
		return nil
	}

	return h.removeMap(cfg.Hosts)
}

// addMap adds an entry to the hosts file.
func (h *Hosts) addMap(hosts map[string]string) error {
	for key, value := range hosts {
		if err := h.add(key, value); err != nil {
			return err
		}
	}

	return h.Flush()
}

// removeMap removes an entry from the hosts file.
func (h *Hosts) removeMap(hosts map[string]string) error {
	for key, value := range hosts {
		if err := h.remove(key, value); err != nil {
			return err
		}
	}

	return h.Flush()
}

// add adds an entry to the hosts file.
func (h *Hosts) add(ip string, hosts ...string) error {
	logx.Infof("Adding hostsfile entry: %s %s\n", ip, strings.Join(hosts, " "))
	return h.Add(ip, hosts...)
}

// remove removes an entry from the hosts file.
func (h *Hosts) remove(ip string, hosts ...string) error {
	logx.Infof("Removing hostsfile entry: %s %s\n", ip, strings.Join(hosts, " "))
	return h.Remove(ip, hosts...)
}

// backupHosts writes the hostsfile to the specified directory with the current timestamp.
func (h *Hosts) backupHosts() error {
	return fsx.Write(path.Join("hosts", "hosts_"+timex.TimeStamp()), []byte(h.String()))
}
