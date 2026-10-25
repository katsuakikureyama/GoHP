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

func TestElseifRuntimeGuardAndSkip(t *testing.T) {
	for _, tt := range []struct {
		name, spelling, first string
		fails                 bool
	}{
		{"reached_elseif", "elseif", "false", true},
		{"reached_else_if", "else if", "false", true},
		{"skipped_elseif", "elseif", "true", false},
		{"skipped_else_if", "else if", "true", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := "<?php class Box {} function choose($first, $later) { if ($first) { return 1; } " + tt.spelling + " ($later) { return 2; } else { return 3; } } echo choose(" + tt.first + ", new Box());"
			p, err := frontend.Parse(source, phpTarget(t))
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
			if tt.fails {
				if err == nil || !bytes.Contains(output, []byte("boolean conversion outside supported scalar/array subset")) {
					t.Fatalf("later condition bypassed the Bool guard: %v\n%s", err, output)
				}
			} else if err != nil || string(output) != "1" {
				t.Fatalf("unreachable later condition was evaluated: %v\n%s", err, output)
			}
		})
	}
}
