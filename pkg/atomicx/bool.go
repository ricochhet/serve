package atomicx

import (
	"sync/atomic"

	"github.com/ricochhet/serve/pkg/jsonx"
)

type Bool struct {
	atomic.Bool
}

func (b *Bool) MarshalJSON() ([]byte, error) {
	return jsonx.Marshal(b.Load())
}

func (b *Bool) UnmarshalJSON(data []byte) error {
	return jsonx.Unmarshal(data, b.Store)
}
