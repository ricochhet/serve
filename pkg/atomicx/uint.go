package atomicx

import (
	"sync/atomic"

	"github.com/ricochhet/serve/pkg/jsonx"
)

type Uint32 struct {
	atomic.Uint32
}

type Uint64 struct {
	atomic.Uint64
}

type Uintptr struct {
	atomic.Uintptr
}

func (u *Uint32) MarshalJSON() ([]byte, error) {
	return jsonx.Marshal(u.Load())
}

func (u *Uint32) UnmarshalJSON(data []byte) error {
	return jsonx.Unmarshal(data, u.Store)
}

func (u *Uint64) MarshalJSON() ([]byte, error) {
	return jsonx.Marshal(u.Load())
}

func (u *Uint64) UnmarshalJSON(data []byte) error {
	return jsonx.Unmarshal(data, u.Store)
}

func (u *Uintptr) MarshalJSON() ([]byte, error) {
	return jsonx.Marshal(u.Load())
}

func (u *Uintptr) UnmarshalJSON(data []byte) error {
	return jsonx.Unmarshal(data, u.Store)
}
