package logx

import (
	"fmt"
	"sync/atomic"

	"github.com/ricochhet/serve/pkg/syncx"
)

var (
	safe = Safe{syncx.NewSafe[Logger]()}
	mode atomic.Uint32
)

type Mode uint32

const (
	ModeDebug Mode = 1 << iota
	ModeInfo
	ModeWarn
	ModeError

	ModeAllDebug   = ModeDebug | ModeInfo | ModeWarn | ModeError
	ModeAllRelease = ModeInfo | ModeWarn | ModeError

	ModeNone Mode = 0
)

type Safe struct {
	*syncx.Safe[Logger]
}

// New sets a new global Logger.
func New(name string, colorIndex int, m Mode) {
	SetMode(m)
	safe.SetLocked(NewLogger(name, colorIndex))
}

// Set sets the global Logger.
func Set(logger *Logger) {
	safe.SetLocked(logger)
}

// SetMode sets the logging mode to the specified value.
func SetMode(m Mode) {
	mode.Store(uint32(m))
}

// HasMode checks if the mode is active.
func HasMode(m Mode) bool {
	return Mode(mode.Load())&m != 0
}

// Flag enables/disables the mode based on the condition provided.
func Flag(m Mode, flag bool) {
	if flag {
		AddMode(m)
	} else {
		RemoveMode(m)
	}
}

// AddMode adds the mode.
func AddMode(m Mode) {
	mode.Store(uint32(Mode(mode.Load()) | m))
}

// RemoveMode removes the mode.
func RemoveMode(m Mode) {
	mode.Store(uint32(Mode(mode.Load()) &^ m))
}

// Debug prints if IsDebug is true.
func Debug(a ...any) {
	if HasMode(ModeDebug) {
		fmt.Fprint(safe.GetLocked(), "[debug] ")
		fmt.Fprint(safe.GetLocked(), a...)
	}
}

// Debugf prints if IsDebug is true.
func Debugf(format string, a ...any) {
	if HasMode(ModeDebug) {
		fmt.Fprintf(safe.GetLocked(), "[debug] "+format, a...)
	}
}

// Warn prints a warn log.
func Warn(a ...any) {
	if HasMode(ModeWarn) {
		fmt.Fprint(safe.GetLocked(), "[warn] ")
		fmt.Fprint(safe.GetLocked(), a...)
	}
}

// Warnf prints a warn log.
func Warnf(format string, a ...any) {
	if HasMode(ModeWarn) {
		fmt.Fprintf(safe.GetLocked(), "[warn] "+format, a...)
	}
}

// Error prints an error log.
func Error(a ...any) {
	if HasMode(ModeError) {
		fmt.Fprint(safe.GetLocked(), "[error] ")
		fmt.Fprint(safe.GetLocked(), a...)
	}
}

// Errorf prints a Error log.
func Errorf(format string, a ...any) {
	if HasMode(ModeError) {
		fmt.Fprintf(safe.GetLocked(), "[error] "+format, a...)
	}
}

// Info prints an info log.
func Info(a ...any) {
	if HasMode(ModeInfo) {
		fmt.Fprint(safe.GetLocked(), "[info] ")
		fmt.Fprint(safe.GetLocked(), a...)
	}
}

// Infof prints a Info log.
func Infof(format string, a ...any) {
	if HasMode(ModeInfo) {
		fmt.Fprintf(safe.GetLocked(), "[info] "+format, a...)
	}
}
