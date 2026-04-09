package logx

import (
	"fmt"
	"sync/atomic"

	"github.com/ricochhet/serve/pkg/syncx"
)

var (
	debug atomic.Bool
	safe  = Safe{syncx.NewSafe[Logger]()}
)

type Safe struct {
	*syncx.Safe[Logger]
}

// New sets a new global Logger.
func New(name string, colorIndex int) {
	safe.SetLocked(NewLogger(name, colorIndex))
}

// Set sets the global Logger.
func Set(logger *Logger) {
	safe.SetLocked(logger)
}

// SetDebug sets the debug logging to the specified value.
func SetDebug(v bool) {
	debug.Store(v)
}

// IsDebug gets the debug state.
func IsDebug() bool {
	return debug.Load()
}

// Debug prints if IsDebug is true.
func Debug(a ...any) {
	if debug.Load() {
		fmt.Fprint(safe.GetLocked(), "[debug] ")
		fmt.Fprint(safe.GetLocked(), a...)
	}
}

// Debugf prints if IsDebug is true.
func Debugf(format string, a ...any) {
	if debug.Load() {
		fmt.Fprintf(safe.GetLocked(), "[debug] "+format, a...)
	}
}

// Warn prints a warn log.
func Warn(a ...any) {
	fmt.Fprint(safe.GetLocked(), "[warn] ")
	fmt.Fprint(safe.GetLocked(), a...)
}

// Warnf prints a warn log.
func Warnf(format string, a ...any) {
	fmt.Fprintf(safe.GetLocked(), "[warn] "+format, a...)
}

// Error prints an error log.
func Error(a ...any) {
	fmt.Fprint(safe.GetLocked(), "[error] ")
	fmt.Fprint(safe.GetLocked(), a...)
}

// Errorf prints a Error log.
func Errorf(format string, a ...any) {
	fmt.Fprintf(safe.GetLocked(), "[error] "+format, a...)
}

// Info prints an info log.
func Info(a ...any) {
	fmt.Fprint(safe.GetLocked(), "[info] ")
	fmt.Fprint(safe.GetLocked(), a...)
}

// Infof prints a Info log.
func Infof(format string, a ...any) {
	fmt.Fprintf(safe.GetLocked(), "[info] "+format, a...)
}
