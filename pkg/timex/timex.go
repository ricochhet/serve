package timex

import (
	"time"

	"github.com/ricochhet/serve/pkg/errorx"
)

func TimeStamp() string {
	return time.Now().Format("2006-01-02_15-04-05")
}

// Fn starts a timer with a function.
func Fn(start func() error, end func(string)) error {
	now := time.Now()
	err := start()

	end(time.Since(now).String())

	return errorx.WithFrame(err)
}
