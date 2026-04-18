package serverutil

import (
	"maps"
	"strings"
)

type Map map[string]string

var mappings = map[string]string{
	".{{serve:mappings.go}}": "<serve:mappings.go>",
}

func (m *Map) Copy() {
	maps.Copy(mappings, *m)
}

func (m *Map) Parse(b []byte) []byte {
	maps.Copy(mappings, *m)

	s := string(b)
	for from, to := range mappings {
		if strings.Contains(s, from) {
			s = strings.ReplaceAll(s, from, to)
		}
	}

	return []byte(s)
}
