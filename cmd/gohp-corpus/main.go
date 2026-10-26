// Command gohp-corpus verifies a pinned, whole-file source corpus without
// parsing or executing the selected source files.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/katsuakikureyama/GoHP/internal/corpus"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run completes inspection and output rendering before writing stdout. A
// caller-owned Writer can still fail during the final single write.
func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("gohp-corpus", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		fmt.Fprintln(stderr, "usage: gohp-corpus -manifest MANIFEST [-format markdown|json] SOURCE_ROOT")
	}
	manifest := flags.String("manifest", "", "pinned corpus manifest")
	format := flags.String("format", "markdown", "output format: markdown or json")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if *manifest == "" || flags.NArg() != 1 {
		flags.Usage()
		return 2
	}
	if *format != "markdown" && *format != "json" {
		fmt.Fprintf(stderr, "invalid -format %q: want markdown or json\n", *format)
		return 2
	}

	report, err := corpus.Inspect(*manifest, flags.Arg(0))
	if err != nil {
		fmt.Fprintf(stderr, "inspect source corpus: %v\n", err)
		return 1
	}

	var output []byte
	if *format == "json" {
		output, err = json.MarshalIndent(report, "", "  ")
		if err == nil {
			output = append(output, '\n')
		}
	} else {
		output = corpus.Markdown(report)
	}
	if err != nil {
		fmt.Fprintf(stderr, "render source corpus: %v\n", err)
		return 1
	}
	if err := writeAll(stdout, output); err != nil {
		fmt.Fprintf(stderr, "write source corpus: %v\n", err)
		return 1
	}
	return 0
}

func writeAll(w io.Writer, output []byte) error {
	n, err := w.Write(output)
	if err != nil {
		return err
	}
	if n != len(output) {
		return io.ErrShortWrite
	}
	return nil
}
