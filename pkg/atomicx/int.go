package atomicx

import (
	"sync/atomic"

	"github.com/ricochhet/serve/pkg/jsonx"
)

type Int32 struct {
	atomic.Int32
}

type Int64 struct {
	atomic.Int64
}

func (i *Int32) MarshalJSON() ([]byte, error) {
	return jsonx.Marshal(i.Load())
}

func (i *Int32) UnmarshalJSON(data []byte) error {
	return jsonx.Unmarshal(data, i.Store)
}

func (i *Int64) MarshalJSON() ([]byte, error) {
	return jsonx.Marshal(i.Load())
}

func (i *Int64) UnmarshalJSON(data []byte) error {
	return jsonx.Unmarshal(data, i.Store)
}
