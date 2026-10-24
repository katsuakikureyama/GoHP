package inventory

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadKeepsDeclarationsSeparateWithoutExecuting(t *testing.T) {
	root := t.TempDir()
	before := map[string][]byte{}
	for _, name := range []string{"composer.json", "composer.lock"} {
		b, err := os.ReadFile(filepath.Join("testdata", "project", name))
		if err != nil {
			t.Fatal(err)
		}
		before[name] = b
		writeInput(t, root, name, string(b))
	}
	// No PHP/Composer program is available to this operation.
	t.Setenv("PATH", "")
	r, err := Read(root)
	if err != nil {
		t.Fatal(err)
	}
	if r.Status != "metadata-only" || r.Compatibility != "unverified" || r.LockFreshness != "unverified" {
		t.Fatalf("metadata promoted to compatibility: %+v", r)
	}
	if len(r.Packages) != 2 || r.Root.Name != "gohp-fixture/project" || len(r.Root.ScriptEvents) != 1 {
		t.Fatalf("missing inventory: %+v", r)
	}
	var library Package
	for _, p := range r.Packages {
		if p.Name == "gohp-fixture/library" {
			library = p
		}
	}
	if library.Scope != "production" || library.Source.Reference != "fixture-commit-a" || library.Version != "1.2.3" || len(library.Licenses) != 2 {
		t.Fatalf("locked metadata lost: %+v", library)
	}
	if len(r.PlatformOverrides) != 4 {
		t.Fatalf("platform overrides lost: %+v", r.PlatformOverrides)
	}
	for _, override := range r.PlatformOverrides {
		if override.Name == "ext-missing" && string(override.Value) != "false" {
			t.Fatal("false override lost")
		}
	}
	found := map[string]Link{}
	for _, l := range r.Links {
		if l.Compatibility != "unverified" {
			t.Fatal("dependency marked supported")
		}
		found[l.Owner+"|"+l.Relation+"|"+l.Name] = l
	}
	for _, key := range []string{"composer.json|require|php", "composer.json|require-dev|gohp-fixture/dev-tool", "gohp-fixture/library|require-dev|gohp-fixture/package-tests-only", "gohp-fixture/library|require|ext-mbstring", "gohp-fixture/library|provide|virtual/logger", "gohp-fixture/library|replace|gohp-fixture/replaced", "gohp-fixture/library|conflict|gohp-fixture/old", "gohp-fixture/library|suggest|ext-curl", "composer.lock|platform|php"} {
		if _, ok := found[key]; !ok {
			t.Errorf("missing distinct declaration %s", key)
		}
	}
	if found["gohp-fixture/library|require|ext-mbstring"].Kind != "extension" || found["composer.json|require|php"].Kind != "php" {
		t.Fatal("platform classification lost")
	}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"secret-script-value", "secret-source-value", "secret-dist-value", "secret-repository-value", "must-not-execute.php"} {
		if bytes.Contains(b, []byte(secret)) {
			t.Fatalf("omitted metadata appeared: %s", secret)
		}
	}
	files, err := os.ReadDir(root)
	if err != nil || len(files) != 2 {
		t.Fatalf("inventory wrote files: %v, %v", files, err)
	}
	for name, old := range before {
		now, err := os.ReadFile(filepath.Join(root, name))
		if err != nil || !bytes.Equal(old, now) {
			t.Fatalf("changed input %s: %v", name, err)
		}
	}
}

func TestReadMissingLockIsNotCompatibilityEvidence(t *testing.T) {
	root := t.TempDir()
	writeInput(t, root, "composer.json", `{"require":{"vendor/missing":"^1","ext-unknown":"*","custom-platform":"*"}}`)
	r, err := Read(root)
	if err != nil {
		t.Fatal(err)
	}
	if r.Inputs[1].Status != "absent" || r.LockFreshness != "not-available" || r.Compatibility != "unverified" || len(r.Packages) != 0 {
		t.Fatalf("missing lock accepted as complete: %+v", r)
	}
	if r.Links[0].Kind != "unknown" {
		t.Fatalf("unknown platform treated as package: %+v", r.Links)
	}
}

func TestReadRejectsAmbiguousAndUnsafeMetadata(t *testing.T) {
	for _, input := range []string{
		`{"require":{"php":"^8"},"require":{}}`,
		`{"require":{"php":"^8","php":"*"}}`,
		`{"require":null}`, `{"require":{"php":83}}`, `{"license":123}`,
		`{"config":{"platform":{"php":true}}}`, `{"scripts":[]}`,
		`null`, `[]`, `{} {}`, `{"name":`,
	} {
		t.Run(input, func(t *testing.T) {
			root := t.TempDir()
			writeInput(t, root, "composer.json", input)
			if _, err := Read(root); err == nil {
				t.Fatal("invalid/ambiguous manifest accepted")
			}
		})
	}
	for _, lock := range []string{
		`{}`, `{"content-hash":"x","packages":[],"packages-dev":null}`,
		`{"content-hash":"x","packages":[{"name":"a/b"}],"packages-dev":[]}`,
		`{"content-hash":"x","packages":[{"name":"a/b","version":"1"}],"packages-dev":[{"name":"a/b","version":"2"}]}`,
	} {
		root := t.TempDir()
		writeInput(t, root, "composer.json", `{}`)
		writeInput(t, root, "composer.lock", lock)
		if _, err := Read(root); err == nil {
			t.Fatal("invalid lock treated as absence or valid metadata")
		}
	}
	t.Run("size and depth bounds", func(t *testing.T) {
		root := t.TempDir()
		writeInput(t, root, "composer.json", strings.Repeat(" ", maxInputBytes+1))
		if _, err := Read(root); err == nil {
			t.Fatal("oversized metadata accepted")
		}
		writeInput(t, root, "composer.json", `{"extra":`+strings.Repeat("[", 70)+"0"+strings.Repeat("]", 70)+"}")
		if _, err := Read(root); err == nil {
			t.Fatal("deep metadata accepted")
		}
	})
	for _, mode := range []string{"escape", "dangling", "directory"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			writeInput(t, root, "composer.json", `{}`)
			path := filepath.Join(root, "composer.lock")
			if mode == "directory" {
				if err := os.Mkdir(path, 0700); err != nil {
					t.Fatal(err)
				}
			} else {
				target := filepath.Join(t.TempDir(), "lock.json")
				if mode == "escape" {
					if err := os.WriteFile(target, []byte(`{}`), 0600); err != nil {
						t.Fatal(err)
					}
				}
				if err := os.Symlink(target, path); err != nil {
					t.Skipf("symlinks unavailable: %v", err)
				}
			}
			if _, err := Read(root); err == nil {
				t.Fatal("unsafe lock accepted as valid or absent")
			}
		})
	}
}

func writeInput(t *testing.T, root, name, data string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, name), []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
}
