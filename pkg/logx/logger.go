package logx

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/ricochhet/serve/pkg/atomicx"
	"github.com/ricochhet/serve/pkg/errorx"
	"github.com/sasha-s/go-deadlock"
)

type Logger struct {
	idx     int
	name    string
	writes  chan []byte
	done    chan struct{}
	timeout time.Duration // How long to wait before printing partial lines.
	buffers buffers       // Partial lines awaiting printing.
}

var Colors = []int{
	32, // Green.
	36, // Cyan.
	35, // Magenta.
	33, // Yellow.
	34, // Blue.
	31, // Red.
}
var mutex deadlock.Mutex

var out = colorable.NewColorableStdout()

type buffers [][]byte

var (
	LogTime           = &atomicx.Bool{}
	MaxProcNameLength = &atomicx.Int32{}
)

// NewLogger is a convenience function for CreateLogger that additionally assigns MaxProcNameLength.
func NewLogger(name string, colorIndex int) *Logger {
	MaxProcNameLength.Store(int32(len(name)))
	return newLogger(name, colorIndex)
}

// WriteTo writes buffer to w (io.Writer).
func (v *buffers) WriteTo(w io.Writer) (n int64, err error) {
	for _, b := range *v {
		nb, err := w.Write(b)
		n += int64(nb)

		if err != nil {
			v.consume(n)
			return n, errorx.WithFrame(err)
		}
	}

	v.consume(n)

	return n, nil
}

// consume consumes bytes in a byte slice.
func (v *buffers) consume(n int64) {
	for len(*v) > 0 {
		ln0 := int64(len((*v)[0]))
		if ln0 > n {
			(*v)[0] = (*v)[0][n:]
			return
		}

		n -= ln0
		*v = (*v)[1:]
	}
}

// Write writes p.
func (l *Logger) Write(p []byte) (int, error) {
	l.writes <- p

	<-l.done

	return len(p), nil
}

// writeBuffers writes any stored buffers, plus the given line, then empty out
// the buffers.
func (l *Logger) writeBuffers(line []byte) {
	mutex.Lock()
	fmt.Fprintf(out, "\x1b[%dm", Colors[l.idx])

	if LogTime.Load() {
		now := time.Now().Format("15:04:05")
		fmt.Fprintf(out, "%s %*s | ", now, MaxProcNameLength.Load(), l.name)
	} else {
		fmt.Fprintf(out, "%*s | ", MaxProcNameLength.Load(), l.name)
	}

	fmt.Fprintf(out, "\x1b[m")

	l.buffers = append(l.buffers, line)

	if _, err := l.buffers.WriteTo(out); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write to buffer :%v\n", err)
	}

	l.buffers = l.buffers[0:0]

	mutex.Unlock()
}

// writeLines bundle writes into lines, waiting briefly for completion of lines.
func (l *Logger) writeLines() {
	var tick <-chan time.Time

	for {
		select {
		case w, ok := <-l.writes:
			if !ok {
				if len(l.buffers) > 0 {
					l.writeBuffers([]byte("\n"))
				}

				return
			}

			buf := bytes.NewBuffer(w)

			for {
				line, err := buf.ReadBytes('\n')
				if len(line) > 0 {
					if line[len(line)-1] == '\n' {
						// Any text followed by a newline should flush
						// existing buffers, a bare newline should flush
						// existing buffers, but only if there are any.
						if len(line) != 1 || len(l.buffers) > 0 {
							l.writeBuffers(line)
						}

						tick = nil
					} else {
						l.buffers = append(l.buffers, line)
						tick = time.After(l.timeout)
					}
				}

				if err != nil {
					break
				}
			}

			l.done <- struct{}{}
		case <-tick:
			if len(l.buffers) > 0 {
				l.writeBuffers([]byte("\n"))
			}

			tick = nil
		}
	}
}

// newLogger creates a new logger with the given name and colorIndex.
func newLogger(name string, colorIndex int) *Logger {
	mutex.Lock()
	defer mutex.Unlock()

	l := &Logger{
		idx:     colorIndex,
		name:    name,
		writes:  make(chan []byte),
		done:    make(chan struct{}),
		timeout: 2 * time.Millisecond,
	}
	go l.writeLines()

	return l
}
