// Command gohp translates a closed PHP subset into Go source.
package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/katsuakikureyama/GoHP/internal/codegen"
	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func main() {
	in := flag.String("in", "", "PHP input file")
	out := flag.String("out", "", "Go output file (stdout if omitted)")
	version := flag.String("php-version", "8.3", "PHP target: 7.4 or 8.3")
	report := flag.String("report", "", "optional provenance JSON path")
	flag.Parse()
	if *in == "" {
		flag.Usage()
		os.Exit(2)
	}
	src, e := os.ReadFile(*in)
	check(e)
	p, e := frontend.Parse(string(src), *version)
	check(e)
	goSrc, e := codegen.Generate(p)
	check(e)
	if *out == "" {
		_, e = os.Stdout.Write(goSrc)
	} else {
		e = os.WriteFile(*out, goSrc, 0644)
	}
	check(e)
	if *report != "" {
		b, e := json.MarshalIndent(map[string]any{"schema": "gohp.build/v0.1", "ir_version": ir.Version, "compiler_version": "0.1.0-dev", "php_target": *version, "source_sha256": fmt.Sprintf("%x", sha256.Sum256(src)), "generated_sha256": fmt.Sprintf("%x", sha256.Sum256(goSrc)), "input": *in, "status": "generated", "compatibility": "not established by compilation"}, "", "  ")
		check(e)
		check(os.WriteFile(*report, append(b, '\n'), 0644))
	}
}
func check(e error) {
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
