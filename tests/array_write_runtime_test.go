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

type arrayRuntimeCase struct{ name, source, message string }

func TestArrayWriteRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"base", "<?php function update_value($a) { $a[0] = 1; return $a; } echo update_value(null);", "array assignment requires array"},
		{"key", "<?php function update_value($key) { $a = []; $a[$key] = 1; return $a; } $result = update_value(true);", "unsupported array key type"},
	})
}

func TestArrayAppendRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"null", "<?php function add_value($a) { $a[] = 1; return $a; } $result = add_value(null);", "array append requires array"},
		{"scalar", "<?php function add_value($a) { $a[] = 1; return $a; } $result = add_value(false);", "array append requires array"},
		{"overflow", "<?php $a = [9223372036854775807 => 'edge']; $a[] = 'overflow';", "array key overflow unsupported"},
	})
}

func checkArrayRuntimeGuards(t *testing.T, cases []arrayRuntimeCase) {
	t.Helper()
	for _, tt := range cases {
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
