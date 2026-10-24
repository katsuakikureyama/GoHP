// Command gohp-inventory renders Composer project metadata without executing
// Composer, project scripts, or applications.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/katsuakikureyama/GoHP/internal/inventory"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run performs all reads and rendering before it writes stdout. A caller-owned
// Writer may still fail during its final write.
func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("gohp-inventory", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		fmt.Fprintln(stderr, "usage: gohp-inventory [-format markdown|json] PROJECT_DIR")
	}
	format := flags.String("format", "markdown", "output format: markdown or json")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() != 1 {
		flags.Usage()
		return 2
	}
	if *format != "markdown" && *format != "json" {
		fmt.Fprintf(stderr, "invalid -format %q: want markdown or json\n", *format)
		return 2
	}

	report, err := inventory.Read(flags.Arg(0))
	if err != nil {
		fmt.Fprintf(stderr, "read Composer metadata %q: %v\n", flags.Arg(0), err)
		return 1
	}
	var output []byte
	if *format == "json" {
		output, err = json.MarshalIndent(report, "", "  ")
		if err == nil {
			output = append(output, '\n')
		}
	} else {
		output = inventory.Markdown(report)
	}
	if err != nil {
		fmt.Fprintf(stderr, "encode Composer inventory: %v\n", err)
		return 1
	}
	if _, err := stdout.Write(output); err != nil {
		fmt.Fprintf(stderr, "write Composer inventory: %v\n", err)
		return 1
	}
	return 0
}
