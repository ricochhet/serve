package serverutil

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
