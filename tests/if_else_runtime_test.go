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

func TestIfElseRuntimeGuards(t *testing.T) {
	for _, tt := range []struct{ name, source, message string }{
		{"condition", "<?php class Box {} function choose($value) { if ($value) { return 1; } else { return 0; } } echo choose(new Box());", "boolean conversion outside supported scalar/array subset"},
		{"else_body", "<?php function size($value) { if (false) { return 0; } else { return count($value); } } echo size(null);", "count requires array"},
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
				t.Fatalf("runtime guard was bypassed: %v\n%s", err, output)
			}
		})
	}
}
