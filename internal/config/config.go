package config

import (
	"fmt"

	"github.com/ricochhet/serve/pkg/fsx"
	"github.com/ricochhet/serve/pkg/jsonx"
	"github.com/ricochhet/serve/pkg/logx"
)

type Config struct {
	Hosts   map[string]string `json:"hosts"`
	TLS     TLS               `json:"tls"`
	Servers []Server          `json:"servers"`
}

type TLS struct {
	Enabled  bool   `json:"enabled"`
	CertFile string `json:"certFile"`
	KeyFile  string `json:"keyFile"`
}

type Timeouts struct {
	ReadHeader int `json:"readHeader"`
	Read       int `json:"read"`
	Write      int `json:"write"`
	Idle       int `json:"idle"`
}

type Server struct {
	Port             int      `json:"port"`
	AllowCredentials bool     `json:"allowCredentials"`
	IPLimit          int      `json:"ipLimit"`
	MaxAge           int      `json:"maxAge"`
	Timeouts         Timeouts `json:"timeouts"`

	Files []File `json:"files"`
}

type File struct {
	Route string `json:"route"`
	Path  string `json:"path"`

	Info Info `json:"info"`
}

type Info struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
}

// Read reads the path if it exists, otherwise returning a default config.
func Read(path string) (*Config, error) {
	var (
		config *Config
		err    error
	)

	exists := fsx.Exists(path)
	switch {
	case exists:
		config, err = jsonx.ReadAndUnmarshal[Config](path)
		if err != nil {
			logx.Errorf("Error reading server config: %v\n", err)
		}

		return config, err
	case !exists && path != "":
		return nil, fmt.Errorf("path specified but does not exist: %s", path)
	default:
		logx.Infof("Starting with default server config\n")
		return &Config{}, nil
	}
}
