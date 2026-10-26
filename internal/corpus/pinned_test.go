package corpus

import "testing"

// The committed selection must remain readable without fetching upstream code.
// This checks metadata, not whether a WordPress checkout matches the pin.
func TestPinnedCorpusManifest(t *testing.T) {
	m, _, err := readManifest("../../corpora/wordpress-functions/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var subjects, provenance int
	for _, file := range m.Files {
		if file.Role == "subject" {
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
