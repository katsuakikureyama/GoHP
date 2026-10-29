package corpus

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSourcesNoAliases(t *testing.T) {
	for _, kind := range []string{"file symlink", "directory symlink", "selected hardlink"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			files := map[string]string{"src/example.php": "<?php class Example {}", "NOTICE": "notice", "PROVENANCE": "source"}
			if err := os.Mkdir(filepath.Join(root, "src"), 0700); err != nil {
				t.Fatal(err)
			}
			for name, contents := range files {
				if err := os.WriteFile(filepath.Join(root, name), []byte(contents), 0600); err != nil {
					t.Fatal(err)
				}
			}
			manifest := validManifest(files)
			switch kind {
			case "file symlink":
				if err := os.Symlink("src/example.php", filepath.Join(root, "alias.php")); err != nil {
					t.Fatal(err)
				}
				manifest["files"].([]map[string]any)[0] = fileEntry("alias.php", "source", files["src/example.php"])
			case "directory symlink":
				if err := os.Symlink("src", filepath.Join(root, "alias")); err != nil {
					t.Fatal(err)
				}
				manifest["files"].([]map[string]any)[0] = fileEntry("alias/example.php", "source", files["src/example.php"])
			case "selected hardlink":
				if err := os.Link(filepath.Join(root, "src/example.php"), filepath.Join(root, "alias.php")); err != nil {
					t.Fatal(err)
				}
				manifest["files"] = append(manifest["files"].([]map[string]any), fileEntry("alias.php", "source", files["src/example.php"]))
			}
			manifestPath := writeManifest(t, root, manifest)
			// The old verifier contract is retained; only module-profile loading
			// imposes the new identity restriction.
			if _, _, err := LoadSources(manifestPath, root); err != nil {
				t.Fatalf("ordinary source verification changed: %v", err)
			}
			report, sources, err := LoadSourcesNoAliases(manifestPath, root)
			if err == nil || sources != nil {
				t.Fatalf("strict module selection accepted alias or leaked sources: %v", err)
			}
			assertZeroReport(t, report)
		})
	}
}
