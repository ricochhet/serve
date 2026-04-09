package cmdx

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
)

type Info []*info

type info struct {
	Usage string
	Desc  string
}

// Usage pretty prints CommandInfo and lists flag.PrintDefaults().
func (c *Info) Usage() {
	tw := tabwriter.NewWriter(os.Stderr, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "Usage:")

	for _, c := range *c {
		fmt.Fprintf(tw, "  %s\t# %s\n", c.Usage, c.Desc)
	}

	tw.Flush()
	flag.PrintDefaults()
}

// Expects checks if flag.Narg() < v+1 (expected arguments), calling Usage() if true.
func (c *Info) Expects(v int) {
	if flag.NArg() < v+1 {
		c.Usage()
	}
}
