package main

import (
	"flag"
	"os"
	"strings"

	"github.com/ricochhet/serve/internal/server"
	"github.com/ricochhet/serve/internal/serverutil"
	"github.com/ricochhet/serve/pkg/cmdx"
	"github.com/ricochhet/serve/pkg/logx"
)

var (
	buildDate string
	gitHash   string
	buildOn   string
)

func version() {
	logx.Infof(logx.Get(), "serve-%s\n", gitHash)
	logx.Infof(logx.Get(), "Build date: %s\n", buildDate)
	logx.Infof(logx.Get(), "Build on: %s\n", buildOn)
	os.Exit(0)
}

func main() {
	logx.LogTime.Store(true)
	logx.MaxProcNameLength.Store(0)
	logx.Set(logx.NewLogger("serve", 0))
	logx.SetDebug(flags.Debug)
	_ = cmdx.QuickEdit(flags.QuickEdit)

	if cmd, err := commands(); cmd {
		if err != nil {
			logx.Errorf(logx.Get(), "Error running command: %v\n", err)
		}

		return
	}

	s := server.NewServer(flags.ConfigFile, flags.Hosts, &serverutil.TLS{
		Enabled:  true,
		CertFile: flags.CertFile,
		KeyFile:  flags.KeyFile,
	}, Embed())
	if err := serverCmd(s); err != nil {
		logx.Errorf(logx.Get(), "%v\n", err)
	}
}

// commands handles the specified command flags.
func commands() (bool, error) {
	args := flag.Args()
	if len(args) == 0 {
		return false, nil
	}

	cmd := strings.ToLower(args[0])
	rest := args[1:]

	switch cmd {
	case "dump", "d":
		cmds.Expects(1)
		return true, dumpCmd(rest...)
	case "list", "l":
		cmds.Expects(1)
		return true, listCmd(rest...)
	case "help", "h":
		cmds.Usage()
	case "version", "v":
		version()
	default:
		cmds.Usage()
	}

	return false, nil
}
