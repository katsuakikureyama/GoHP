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
	"github.com/katsuakikureyama/GoHP/internal/phpversion"
)

func main() {
	in := flag.String("in", "", "PHP input file")
	out := flag.String("out", "", "Go output file (stdout if omitted)")
	version := flag.String("php-version", phpversion.DefaultTarget, "PHP language target: "+phpversion.TargetHelp())
	report := flag.String("report", "", "optional provenance JSON path")
	flag.Parse()
	if *in == "" {
		flag.Usage()
		os.Exit(2)
	}
	if *report != "" && *out == "" {
		check(fmt.Errorf("-report requires -out; use fresh candidate paths"))
	}
	src, e := os.ReadFile(*in)
	check(e)
	p, e := frontend.Parse(string(src), *version)
	check(e)
	goSrc, sourceMap, e := codegen.GenerateWithMap(p)
	check(e)
	if *out == "" {
		_, e = os.Stdout.Write(goSrc)
		check(e)
		return
	}
	files := []artifact{{path: *out, data: goSrc}}
	if *report != "" {
		b, e := json.MarshalIndent(map[string]any{
			"schema": "gohp.build/v0.1", "ir_version": ir.Version, "compiler_version": "0.1.0-dev",
			"php_target": *version, "source_sha256": fmt.Sprintf("%x", sha256.Sum256(src)),
			"generated_sha256": fmt.Sprintf("%x", sha256.Sum256(goSrc)), "input": *in,
			"output": *out, "source_map": sourceMap, "status": "generated",
			"compatibility": "not established by compilation",
		}, "", "  ")
		check(e)
		files = append(files, artifact{path: *report, data: append(b, '\n')})
	}
	check(publishArtifacts(files))
}
func check(e error) {
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
