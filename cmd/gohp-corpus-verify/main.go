// Command gohp-corpus-verify checks pinned source bytes without executing PHP.
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

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("gohp-corpus-verify", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		fmt.Fprintln(stderr, "usage: gohp-corpus-verify -manifest MANIFEST.json SOURCE_ROOT")
	}
	manifest := flags.String("manifest", "", "pinned corpus manifest")
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
	report, err := corpus.Verify(*manifest, flags.Arg(0))
	if err != nil {
		fmt.Fprintf(stderr, "verify corpus: %v\n", err)
		return 1
	}
	output, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "encode corpus verification: %v\n", err)
		return 1
	}
	output = append(output, '\n')
	n, err := stdout.Write(output)
	if err == nil && n != len(output) {
		err = io.ErrShortWrite
	}
	if err != nil {
		fmt.Fprintf(stderr, "write corpus verification: %v\n", err)
		return 1
	}
	return 0
}
