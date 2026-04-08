package timex

import (
	"time"

	"github.com/ricochhet/serve/pkg/errorx"
)

func TimeStamp() string {
	return time.Now().Format("2006-01-02_15-04-05")
}

// Fn starts a timer with a function.
func Fn(fn func() error, name string, caller func(string, string)) error {
	start := time.Now()
	err := fn()
	elapsed := time.Since(start)
	caller(name, elapsed.String())

	return errorx.WithFrame(err)
}
