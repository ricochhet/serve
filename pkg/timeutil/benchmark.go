package timeutil

import (
	"time"

	"github.com/ricochhet/serve/pkg/errutil"
)

// Timer starts a timer with a function.
func Timer(fn func() error, name string, caller func(string, string)) error {
	start := time.Now()
	err := fn()
	elapsed := time.Since(start)
	caller(name, elapsed.String())

	return errutil.WithFrame(err)
}
