// Command gohp-compat-view renders read-only views of archived GoHP
// compatibility evidence. It never reruns tests or alters an archive.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/katsuakikureyama/GoHP/internal/compat"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run writes no stdout until every archive has loaded and the complete view has
// been built. A caller-supplied Writer can still fail part-way through a write.
func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("gohp-compat-view", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		fmt.Fprintln(stderr, "usage: gohp-compat-view [-target 7.4|8.3] [-kind KIND] [-status STATUS] [-format markdown|json] ARCHIVE_DIR [ARCHIVE_DIR...]")
	}
	target := flags.String("target", "", "PHP target: 7.4 or 8.3")
	kind := flags.String("kind", "", "catalog feature kind")
	status := flags.String("status", "", "feature status")
	format := flags.String("format", "markdown", "output format: markdown or json")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() == 0 {
		flags.Usage()
		return 2
	}
	filter := compat.ViewFilter{Target: *target, Kind: *kind, Status: *status}
	if err := filter.Validate(); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if *format != "markdown" && *format != "json" {
		fmt.Fprintf(stderr, "invalid -format %q: want markdown or json\n", *format)
		return 2
	}

	reports := make([]compat.Report, 0, flags.NArg())
	for _, archive := range flags.Args() {
		report, err := compat.LoadArchive(archive)
		if err != nil {
			fmt.Fprintf(stderr, "load archive %q: %v\n", archive, err)
			return 1
		}
		reports = append(reports, report)
	}
	view, err := compat.BuildView(reports, filter)
	if err != nil {
		fmt.Fprintf(stderr, "build compatibility view: %v\n", err)
		return 1
	}

	var output []byte
	if *format == "json" {
		output, err = json.MarshalIndent(view, "", "  ")
		if err == nil {
			output = append(output, '\n')
		}
	} else {
		output = compat.ViewMarkdown(view)
	}
	if err != nil {
		fmt.Fprintf(stderr, "encode compatibility view: %v\n", err)
		return 1
	}
	if _, err := stdout.Write(output); err != nil {
		fmt.Fprintf(stderr, "write compatibility view: %v\n", err)
		return 1
	}
	return 0
}
