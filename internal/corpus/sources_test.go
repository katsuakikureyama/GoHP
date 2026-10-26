package corpus

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSourcesRetainsVerifiedBytes(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{"src/example.php": "<?php exit(99);", "NOTICE": "notice", "PROVENANCE": "version"}
	if err := os.Mkdir(filepath.Join(root, "src"), 0700); err != nil {
		t.Fatal(err)
	}
	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
	}
	manifest := writeManifest(t, root, validManifest(files))
	report, sources, err := LoadSources(manifest, root)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != "source-verified" || len(sources) != 1 || sources[0].Path != "src/example.php" || string(sources[0].Bytes) != files["src/example.php"] {
		t.Fatalf("unexpected source snapshot: %+v / %+v", report, sources)
	}
	if err := os.WriteFile(filepath.Join(root, "src/example.php"), []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if string(sources[0].Bytes) != files["src/example.php"] {
		t.Fatal("snapshot must not reopen mutated path")
	}
	if err := os.WriteFile(filepath.Join(root, "src/example.php"), []byte(files["src/example.php"]), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "PROVENANCE"), []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	report, sources, err = LoadSources(manifest, root)
	if err == nil || sources != nil {
		t.Fatal("late provenance failure leaked checked source buffers")
	}
	assertZeroReport(t, report)
}
