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
	"github.com/katsuakikureyama/GoHP/internal/sourceclosure"
)

func main() {
	in := flag.String("in", "", "PHP input file")
	manifest := flag.String("manifest", "", "explicit corpus manifest for a static declaration closure (exclusive with -in)")
	sourceRoot := flag.String("source-root", "", "private immutable source root for -manifest")
	entry := flag.String("entry", "", "manifest-relative PHP entry for -manifest")
	out := flag.String("out", "", "Go output file (stdout if omitted)")
	version := flag.String("php-version", phpversion.DefaultTarget, "PHP language target: "+phpversion.TargetHelp())
	experimentalPCRE2 := flag.Bool("experimental-pcre2", false, "admit the experimental PHP 8.3 preg_match local-output syntax (does not enable a PCRE2 backend)")
	experimentalPCRE2Profile := flag.String("experimental-pcre2-profile", "", "explicit experimental regex profile: pcre2-php83-v1 (default) or pcre2-php83-match1-v1; requires -experimental-pcre2")
	report := flag.String("report", "", "optional provenance JSON path")
	flag.Parse()
	bundleMode := *manifest != "" || *sourceRoot != "" || *entry != ""
	if (bundleMode && (*in != "" || *manifest == "" || *sourceRoot == "" || *entry == "")) || (!bundleMode && *in == "") || flag.NArg() != 0 {
		flag.Usage()
		os.Exit(2)
	}
	if *report != "" && *out == "" {
		check(fmt.Errorf("-report requires -out; use fresh candidate paths"))
	}
	var src []byte
	var p ir.Program
	var graph sourceclosure.Report
	var e error
	options := frontend.Options{ExperimentalPCRE2: *experimentalPCRE2, PCRE2Profile: *experimentalPCRE2Profile}
	_, e = options.RegexProfile(*version)
	check(e)
	if bundleMode {
		p, graph, e = sourceclosure.Assemble(*manifest, *sourceRoot, *entry, *version, options)
	} else {
		src, e = os.ReadFile(*in)
		check(e)
		p, e = frontend.ParseWithOptions(string(src), *version, options)
	}
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
		provenance := map[string]any{
			"schema": "gohp.build/v0.1", "ir_version": ir.Version, "compiler_version": "0.1.0-dev",
			"php_target": *version, "regex_profile": p.RegexProfile, "source_sha256": fmt.Sprintf("%x", sha256.Sum256(src)),
			"generated_sha256": fmt.Sprintf("%x", sha256.Sum256(goSrc)), "input": *in,
			"output": *out, "source_map": sourceMap, "status": "generated",
			"compatibility": "not established by compilation",
		}
		if bundleMode {
			provenance["schema"] = "gohp.build/v0.2"
			delete(provenance, "source_sha256")
			delete(provenance, "input")
			provenance["manifest_sha256"] = graph.ManifestSHA256
			provenance["closure_profile"] = "static-declaration-closure/v1"
			provenance["entry"] = graph.Entry
			provenance["sources"] = p.Sources
			provenance["loading_graph"] = graph
		}
		b, e := json.MarshalIndent(provenance, "", "  ")
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
