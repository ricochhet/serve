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
func Read[T any](path string, passthru bool) (*T, error) {
	var (
		t   *T
		err error
	)

	exists := fsx.Exists(path)
	switch {
	case exists:
		t, err = jsonx.ReadAndUnmarshal[T](path)
		if err != nil {
			logx.Errorf("Error reading json: %v\n", err)
		}

		return t, err
	case !exists && path != "":
		if !passthru {
			return nil, fmt.Errorf("path specified but does not exist: %s", path)
		}

		return defaultT[T](), nil
	default:
		return defaultT[T](), nil
	}
}

func defaultT[T any]() *T {
	n := new(T)
	logx.Infof("Starting with default %T\n", *n)

	return n
}
