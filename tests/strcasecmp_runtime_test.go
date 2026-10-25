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

func TestStrCaseCmpRuntimeTypeGuards(t *testing.T) {
	const prefix = "<?php class Box {} function compare($left, $right) { return strcasecmp($left, $right); } "
	for _, tt := range []struct{ name, call string }{
		{"int_left", "echo compare(1, '1');"},
		{"null_right", "echo compare('', null);"},
		{"array_left", "echo compare([], '');"},
		{"object_right", "echo compare('', new Box());"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p, err := frontend.Parse(prefix+tt.call, phpTarget(t))
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
			if err == nil || !bytes.Contains(output, []byte("strcasecmp requires two strings")) {
				t.Fatalf("unsupported value escaped strcasecmp guard: %v\n%s", err, output)
			}
		})
	}
}
