package main

import (
	"path/filepath"

	"github.com/ricochhet/serve/internal/config"
	"github.com/ricochhet/serve/internal/server"
	"github.com/ricochhet/serve/internal/serverutil"
	"github.com/ricochhet/serve/pkg/errorx"
	"github.com/ricochhet/serve/pkg/fsx"
	"github.com/ricochhet/serve/pkg/logx"
	"github.com/ricochhet/serve/pkg/timex"
)

// dumpCmd command.
func dumpCmd(a ...string) error {
	return timex.Fn(func() error {
		err := Embed().Dump(a[0], "", func(f fsx.File, b []byte) error {
			logx.Infof("Writing: %s (%d bytes)\n", f.Path, f.Info.Size())
			return fsx.Write(filepath.Join("dump", f.Path), b)
		})
		if err != nil {
			logx.Errorf("Error dumping embedded files: %v\n", err)
		}

		return err
	}, func(elapsed string) {
		logx.Infof("Took %s\n", elapsed)
	})
}

// listCmd command.
func listCmd(a ...string) error {
	return timex.Fn(func() error {
		err := Embed().List(a[0], func(files []fsx.File) error {
			for _, file := range files {
				if file.Info.IsDir() {
					continue
				}

				logx.Infof("%s (%d bytes)\n", file.Path, file.Info.Size())
			}

			return nil
		})
		if err != nil {
			logx.Errorf("Error listing embedded files: %v\n", err)
		}

		return err
	}, func(elapsed string) {
		logx.Infof("Took %s\n", elapsed)
	})
}

// serverCmd command.
func serverCmd(s *server.Context) error {
	return timex.Fn(func() error {
		err := s.StartServer()
		if err != nil {
			logx.Errorf("Error starting server: %v\n", err)
		}

		return err
	}, func(elapsed string) {
		logx.Infof("Took %s\n", elapsed)
	})
}

// addHostsCmd command.
func addHostsCmd() error {
	return timex.Fn(func() error {
		h, err := serverutil.NewHosts()
		if err != nil {
			return errorx.WithFrame(err)
		}

		config, err := config.Read(flags.ConfigFile)
		if err != nil {
			return errorx.WithFrame(err)
		}

		return h.AddHosts(config)
	}, func(elapsed string) {
		logx.Infof("Took %s\n", elapsed)
	})
}

// removeHostsCmd command.
func removeHostsCmd() error {
	return timex.Fn(func() error {
		h, err := serverutil.NewHosts()
		if err != nil {
			return errorx.WithFrame(err)
		}

		config, err := config.Read(flags.ConfigFile)
		if err != nil {
			return errorx.WithFrame(err)
		}

		return h.RemoveHosts(config)
	}, func(elapsed string) {
		logx.Infof("Took %s\n", elapsed)
	})
}
