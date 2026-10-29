// Package sourceclosure inspects a bounded graph of checked declaration modules.
// It is not an executable loader, compiler entry point or PHP compatibility test.
package sourceclosure

import (
	"fmt"
	"path"
	"strings"

	"github.com/katsuakikureyama/GoHP/internal/corpus"
	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/phpversion"
)

type Module struct {
	SourceID     string                       `json:"source_id"`
	Bytes        int64                        `json:"bytes"`
	SHA256       string                       `json:"sha256"`
	Declarations []frontend.ModuleDeclaration `json:"declarations"`
}

type Edge struct {
	From   string `json:"from"`
	Line   int    `json:"line"`
	Suffix string `json:"suffix"`
	To     string `json:"to"`
	Reused bool   `json:"reused"`
}

// Report retains file identities and every edge in deterministic DFS order.
// graph-checked means syntax/identity discovery only, never executable loading.
// Unreached sources are byte-verified, but were not parsed as graph modules.
type Report struct {
	Schema         string   `json:"schema"`
	Status         string   `json:"status"`
	PHPTarget      string   `json:"php_target"`
	RegexProfile   string   `json:"regex_profile,omitempty"`
	ManifestSHA256 string   `json:"manifest_sha256"`
	Entry          string   `json:"entry"`
	Modules        []Module `json:"modules"`
	Edges          []Edge   `json:"edges"`
	Unreached      []string `json:"unreached"`
	Compatibility  string   `json:"compatibility"`
}

// Inspect verifies the entire source selection, then inspects only checked
// buffers reachable from entry. It never reopens source paths after verification.
// All errors return a zero report; no executable IR or generated artifact escapes.
func Inspect(manifestPath, sourceRoot, entry, version string, options frontend.Options) (Report, error) {
	if err := phpversion.ValidateTarget(version); err != nil {
		return Report{}, err
	}
	if options.ExperimentalPCRE2 && version != "8.3" {
		return Report{}, fmt.Errorf("experimental regex profile requires PHP 8.3")
	}
	verified, sources, err := corpus.LoadSourcesNoAliases(manifestPath, sourceRoot)
	if err != nil {
		return Report{}, err
	}
	selectedTarget := false
	for _, target := range verified.Manifest.PHPTargets {
		selectedTarget = selectedTarget || target == version
	}
	if !selectedTarget {
		return Report{}, fmt.Errorf("PHP target %q is not selected by the manifest", version)
	}
	byID := make(map[string]corpus.Source, len(sources))
	metadata := make(map[string]corpus.ManifestFile, len(sources))
	for _, source := range sources {
		byID[source.Path] = source
	}
	for _, file := range verified.Manifest.Files {
		if file.Role == "source" {
			metadata[file.Path] = file
		}
	}
	if _, exists := byID[entry]; !exists {
		return Report{}, fmt.Errorf("entry %q is not a selected source", entry)
	}
	report := Report{Schema: "gohp.source-closure/v0.1", Status: "graph-checked",
		PHPTarget: version, ManifestSHA256: verified.ManifestSHA256, Entry: entry,
		Modules: []Module{}, Edges: []Edge{}, Unreached: []string{}, Compatibility: "unverified"}
	// 0 = unseen, 1 = visiting, 2 = loaded. Cycles are excluded, not approximated.
	states := make(map[string]int, len(sources))
	type definition struct {
		file string
		line int
	}
	declarations := map[string]definition{}
	var visit func(string) error
	visit = func(id string) error {
		module, err := frontend.InspectModule(id, string(byID[id].Bytes), version, options)
		if err != nil {
			return err
		}
		states[id] = 1
		file := metadata[id]
		report.RegexProfile = module.RegexProfile
		report.Modules = append(report.Modules, Module{SourceID: id, Bytes: file.Bytes,
			SHA256: file.SHA256, Declarations: module.Declarations})
		for _, declaration := range module.Declarations {
			key := declaration.Kind + ":" + strings.ToLower(declaration.Name)
			if previous, exists := declarations[key]; exists {
				return fmt.Errorf("%s:%d: duplicate %s %q (first at %s:%d)", id,
					declaration.Line, declaration.Kind, declaration.Name, previous.file, previous.line)
			}
			declarations[key] = definition{file: id, line: declaration.Line}
		}
		for _, require := range module.Requires {
			// InspectModule validates the suffix before path.Join: normalization
			// must never turn traversal/aliases into accepted source identities.
			destination := path.Join(path.Dir(id), strings.TrimPrefix(require.Suffix, "/"))
			if _, exists := byID[destination]; !exists {
				return fmt.Errorf("%s:%d: require_once target %q is not a selected source", id, require.Line, destination)
			}
			if states[destination] == 1 {
				return fmt.Errorf("%s:%d: require_once cycle to %q is outside the static profile", id, require.Line, destination)
			}
			report.Edges = append(report.Edges, Edge{From: id, Line: require.Line,
				Suffix: require.Suffix, To: destination, Reused: states[destination] == 2})
			if states[destination] == 0 {
				if err := visit(destination); err != nil {
					return fmt.Errorf("%s:%d: require_once %q: %w", id, require.Line, require.Suffix, err)
				}
			}
		}
		states[id] = 2
		return nil
	}
	if err := visit(entry); err != nil {
		return Report{}, err
	}
	for _, source := range sources {
		if states[source.Path] == 0 {
			report.Unreached = append(report.Unreached, source.Path)
		}
	}
	return report, nil
}
