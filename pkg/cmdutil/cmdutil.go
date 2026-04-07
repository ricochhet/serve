package cmdutil

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
)

type Commands []*Command

type Command struct {
	Usage string
	Desc  string
}

// Usage runs flag.PrintDefaults() and exits with code 0.
func (c *Commands) Usage() {
	tw := tabwriter.NewWriter(os.Stderr, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "Usage:")

	for _, c := range *c {
		fmt.Fprintf(tw, "  %s\t# %s\n", c.Usage, c.Desc)
	}

	tw.Flush()
	flag.PrintDefaults()
	os.Exit(0)
}

// Check checks if flag.Narg() < v+1, calling Usage() if true.
func (c *Commands) Check(v int) {
	if flag.NArg() < v+1 {
		c.Usage()
	}
}
