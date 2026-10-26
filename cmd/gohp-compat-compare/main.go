// Command gohp-compat-compare compares the recorded process-output evidence
// in exactly two saved GoHP compatibility archives. It never reruns tests or
// alters either archive.
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

// run loads both archives before it writes any comparison output. A reported
// mismatch or an unavailable comparison is a rendered result, not a command
// failure; command failures are reserved for bad invocation, unreadable
// evidence, encoding, and output I/O errors.
func run(args []string, stdout, stderr io.Writer) int {
	return runWith(args, stdout, stderr, compat.LoadArchive)
}

type archiveLoader func(string) (compat.Report, error)

func runWith(args []string, stdout, stderr io.Writer, load archiveLoader) int {
	flags := flag.NewFlagSet("gohp-compat-compare", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		fmt.Fprintln(stderr, "usage: gohp-compat-compare [-format markdown|json] LEFT_ARCHIVE_DIR RIGHT_ARCHIVE_DIR")
	}
	format := flags.String("format", "markdown", "output format: markdown or json")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() != 2 {
		flags.Usage()
		return 2
	}
	if *format != "markdown" && *format != "json" {
		fmt.Fprintf(stderr, "invalid -format %q: want markdown or json\n", *format)
		return 2
	}

	left, err := load(flags.Arg(0))
	if err != nil {
		fmt.Fprintf(stderr, "load left archive %q: %v\n", flags.Arg(0), err)
		return 1
	}
	right, err := load(flags.Arg(1))
	if err != nil {
		fmt.Fprintf(stderr, "load right archive %q: %v\n", flags.Arg(1), err)
		return 1
	}
	result := compat.CompareOutputs(left, right)

	var output []byte
	if *format == "json" {
		output, err = json.MarshalIndent(result, "", "  ")
		if err == nil {
			output = append(output, '\n')
		}
	} else {
		output = compat.OutputComparisonMarkdown(result)
	}
	if err != nil {
		fmt.Fprintf(stderr, "encode output comparison: %v\n", err)
		return 1
	}
	if _, err := stdout.Write(output); err != nil {
		fmt.Fprintf(stderr, "write output comparison: %v\n", err)
		return 1
	}
	return 0
}
