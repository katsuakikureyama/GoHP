package corpus

import (
	"os"
	"reflect"
	"testing"
)

// This is a metadata-boundary contract only. It does not read WordPress
// sources, verify their bytes, invoke PHP, or assert dependency resolution.
func TestPinnedDependencyTriageManifest(t *testing.T) {
	base := readCommittedManifest(t, "../../corpus/wordpress-functions.json")
	triage := readCommittedManifest(t, "../../corpus/wordpress-dependency-triage.json")

	if base.ID == triage.ID || triage.ID != "wordpress-dependency-triage-54106f6" {
		t.Fatalf("manifest identities must be distinct and triage identity stable: %q / %q", base.ID, triage.ID)
	}
	if !reflect.DeepEqual(base.Upstream, triage.Upstream) {
		t.Fatalf("triage must retain the pinned upstream: %#v / %#v", base.Upstream, triage.Upstream)
	}
	if !reflect.DeepEqual(base.PHPTargets, triage.PHPTargets) {
		t.Fatalf("triage must retain the PHP target policy: %#v / %#v", base.PHPTargets, triage.PHPTargets)
	}
	for _, manifest := range []Manifest{base, triage} {
		if manifest.Selection.Unit != "whole-file" || manifest.Selection.Dependencies != "unresolved" {
			t.Fatalf("%s must remain a whole-file, unresolved-dependency selection: %#v", manifest.ID, manifest.Selection)
		}
	}

	baseSources := sourceFiles(base)
	if len(baseSources) != 1 || baseSources[0].Path != "wp-includes/functions.php" {
		t.Fatalf("original manifest must remain the one-source selection: %#v", baseSources)
	}
	wantPaths := []string{"wp-includes/functions.php", "wp-includes/plugin.php", "wp-includes/class-wp-hook.php", "wp-includes/class-wp-filter-sentinel.php", "wp-includes/l10n.php", "wp-includes/option.php", "wp-includes/formatting.php"}
	selected := sourceFiles(triage)
	var paths []string
	var sourceBytes int64
	for _, file := range selected {
		paths = append(paths, file.Path)
		sourceBytes += file.Bytes
	}
	if !reflect.DeepEqual(paths, wantPaths) {
		t.Fatalf("unexpected triage source selection: %v", paths)
	}
	if selected[0] != baseSources[0] {
		t.Fatal("shared original source identity changed between selections")
	}
	if sourceBytes != 908791 {
		t.Fatalf("unexpected triage source byte denominator: %d", sourceBytes)
	}

	if got, want := nonsourceFiles(triage), nonsourceFiles(base); !reflect.DeepEqual(got, want) {
		t.Fatalf("triage must retain the exact notice/provenance files: %#v / %#v", got, want)
	}
	if got := nonsourceFiles(triage); len(got) != 2 || got[0].Role != "notice" || got[1].Role != "provenance" {
		t.Fatalf("triage must contain exactly notice and provenance entries: %#v", got)
	}
}

func readCommittedManifest(t *testing.T, name string) Manifest {
	t.Helper()
	raw, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := parseManifest(raw)
	if err != nil {
		t.Fatal(err)
	}
	return manifest
}

func sourceFiles(manifest Manifest) []ManifestFile {
	var result []ManifestFile
	for _, file := range manifest.Files {
		if file.Role == "source" {
			result = append(result, file)
		}
	}
	return result
}

func nonsourceFiles(manifest Manifest) []ManifestFile {
	var result []ManifestFile
	for _, file := range manifest.Files {
		if file.Role != "source" {
			result = append(result, file)
		}
	}
	return result
}
