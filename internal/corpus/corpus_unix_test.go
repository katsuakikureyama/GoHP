//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package corpus

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestVerifyRejectsFIFOBeforeOpen(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"src/example.php": "<?php\n",
		"NOTICE":          "notice\n",
		"PROVENANCE":      "provenance\n",
	}
	for name, contents := range files {
		filename := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filename, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	fifo := filepath.Join(root, "src", "selected.fifo")
	if err := syscall.Mkfifo(fifo, 0o600); err != nil {
		t.Fatal(err)
	}
	manifest := validManifest(files)
	manifest["files"].([]map[string]any)[0] = map[string]any{"path": "src/selected.fifo", "role": "source", "bytes": int64(1), "sha256": strings.Repeat("0", 64)}
	if report, err := Verify(writeManifest(t, root, manifest), root); err == nil || !strings.Contains(err.Error(), "not regular") || report.Schema != "" {
		t.Fatalf("FIFO source result = (%#v, %v)", report, err)
	}

	manifestFIFO := filepath.Join(root, "manifest.fifo")
	if err := syscall.Mkfifo(manifestFIFO, 0o600); err != nil {
		t.Fatal(err)
	}
	if report, err := Verify(manifestFIFO, root); err == nil || !strings.Contains(err.Error(), "regular file") || report.Schema != "" {
		t.Fatalf("FIFO manifest result = (%#v, %v)", report, err)
	}
}
