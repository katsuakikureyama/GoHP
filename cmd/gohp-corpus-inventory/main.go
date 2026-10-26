// Command gohp-corpus-inventory reports lexical observations for a verified
// source corpus. It uses the configured native PHP tokenizer but never runs
// selected PHP source as an application.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/katsuakikureyama/GoHP/internal/phpversion"
	"github.com/katsuakikureyama/GoHP/internal/sourceinventory"
)

type inventoryFunc func(context.Context, string, string, string) (sourceinventory.Report, error)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	return runWithInventory(args, stdout, stderr, sourceinventory.Inventory)
}

// runWithInventory permits contract tests to avoid a native PHP dependency.
// It builds the complete JSON report before a single attempt to write stdout.
func runWithInventory(args []string, stdout, stderr io.Writer, inventory inventoryFunc) int {
	flags := flag.NewFlagSet("gohp-corpus-inventory", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		fmt.Fprintln(stderr, "usage: gohp-corpus-inventory -manifest MANIFEST.json [-php-version 8.3|8.4] SOURCE_ROOT")
	}
	manifest := flags.String("manifest", "", "pinned corpus manifest")
	phpTarget := flags.String("php-version", phpversion.DefaultTarget, "PHP language target: "+phpversion.TargetHelp())
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if *manifest == "" || flags.NArg() != 1 || flags.Arg(0) == "" {
		flags.Usage()
		return 2
	}
	if err := phpversion.ValidateTarget(*phpTarget); err != nil {
		fmt.Fprintf(stderr, "invalid PHP target: %v\n", err)
		return 2
	}

	report, err := inventory(context.Background(), *manifest, flags.Arg(0), *phpTarget)
	if err != nil {
		// Inventory errors can contain source paths or native diagnostics. Keep
		// the command's public diagnostics stable and free of source content.
		fmt.Fprintln(stderr, "inventory corpus: failed")
		return 1
	}
	output, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintln(stderr, "encode corpus inventory: failed")
		return 1
	}
	output = append(output, '\n')
	n, err := stdout.Write(output)
	if err == nil && n != len(output) {
		err = io.ErrShortWrite
	}
	if err != nil {
		fmt.Fprintln(stderr, "write corpus inventory: failed")
		return 1
	}
	return 0
}
