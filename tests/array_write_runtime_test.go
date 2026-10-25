package tests

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/codegen"
	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestArrayWriteRuntimeGuards(t *testing.T) {
	for _, tt := range []struct{ name, source, message string }{
		{"base", "<?php function update_value($a) { $a[0] = 1; return $a; } echo update_value(null);", "array assignment requires array"},
		{"key", "<?php function update_value($key) { $a = []; $a[$key] = 1; return $a; } $result = update_value(true);", "unsupported array key type"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p, err := frontend.Parse(tt.source, phpTarget(t))
			if err != nil {
				t.Fatal(err)
			}
			generated, err := codegen.Generate(p)
			if err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			file, binary := filepath.Join(dir, "main.go"), filepath.Join(dir, "app")
			if err := os.WriteFile(file, generated, 0600); err != nil {
				t.Fatal(err)
			}
			if output, err := exec.Command("go", "build", "-o", binary, file).CombinedOutput(); err != nil {
				t.Fatalf("build: %v\n%s", err, output)
			}
			run := exec.Command(binary)
			run.Env = []string{"PATH="}
			output, err := run.CombinedOutput()
			if err == nil || !bytes.Contains(output, []byte(tt.message)) {
				t.Fatalf("expected explicit runtime rejection, got %v\n%s", err, output)
			}
		})
	}
}
