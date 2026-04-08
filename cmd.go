package main

import (
	"path/filepath"

	"github.com/ricochhet/serve/internal/server"
	"github.com/ricochhet/serve/pkg/fsx"
	"github.com/ricochhet/serve/pkg/logx"
	"github.com/ricochhet/serve/pkg/timex"
)

// dumpCmd command.
func dumpCmd(a ...string) error {
	return timex.Fn(func() error {
		err := Embed().Dump(a[0], "", func(f fsx.File, b []byte) error {
			logx.Infof(logx.Get(), "Writing: %s (%d bytes)\n", f.Path, f.Info.Size())
			return fsx.Write(filepath.Join("dump", f.Path), b)
		})
		if err != nil {
			logx.Errorf(logx.Get(), "Error dumping embedded files: %v\n", err)
		}

		return err
	}, "Dump", func(_, elapsed string) {
		logx.Infof(logx.Get(), "Took %s\n", elapsed)
	})
}

// listCmd command.
func listCmd(a ...string) error {
	return timex.Fn(func() error {
		err := Embed().List(a[0], func(files []fsx.File) error {
			for _, f := range files {
				logx.Infof(logx.Get(), "%s (%d bytes)\n", f.Path, f.Info.Size())
			}

			return nil
		})
		if err != nil {
			logx.Errorf(logx.Get(), "Error listing embedded files: %v\n", err)
		}

		return err
	}, "List", func(_, elapsed string) {
		logx.Infof(logx.Get(), "Took %s\n", elapsed)
	})
}

// serverCmd command.
func serverCmd(s *server.Context) error {
	return timex.Fn(func() error {
		err := s.StartServer()
		if err != nil {
			logx.Errorf(logx.Get(), "Error starting server: %v\n", err)
		}

		return err
	}, "Server", func(_, elapsed string) {
		logx.Infof(logx.Get(), "Took %s\n", elapsed)
	})
}
