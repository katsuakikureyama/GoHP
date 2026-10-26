package corpus

import (
	"os"
	"testing"
)

// The committed selection must remain readable without fetching upstream code.
// This checks metadata, not whether a WordPress checkout matches the pin.
func TestPinnedCorpusManifest(t *testing.T) {
	raw, err := os.ReadFile("../../corpus/wordpress-functions.json")
	if err != nil {
		t.Fatal(err)
	}
	m, err := parseManifest(raw)
	if err != nil {
		t.Fatal(err)
	}
	var subjects, provenance int
	for _, file := range m.Files {
		if file.Role == "source" {
			subjects++
			if file.Path != "wp-includes/functions.php" {
				t.Fatalf("unexpected selected source unit: %s", file.Path)
			}
		} else {
			provenance++
		}
	}
	if subjects != 1 || provenance != 2 {
		t.Fatalf("first inventory selects one whole source file and two provenance files, got %d/%d", subjects, provenance)
	}
}
