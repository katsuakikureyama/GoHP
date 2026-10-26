package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/codegen"
	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestPreg83GeneratedDisabledBackend(t *testing.T) {
	p, err := frontend.ParseWithOptions(`<?php $matches = null;
$result = preg_match('/a/s', 'a', $matches, PREG_OFFSET_CAPTURE, 0);
echo 'must-not-publish';`, "8.3", frontend.Options{ExperimentalPCRE2: true})
	if err != nil {
		t.Fatal(err)
	}
	generated, err := codegen.Generate(p)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	source, binary := filepath.Join(dir, "main.go"), filepath.Join(dir, "app")
	if err := os.WriteFile(source, generated, 0600); err != nil {
		t.Fatal(err)
	}
	build := exec.Command("go", "build", "-o", binary, source)
	build.Env = append(os.Environ(), "CGO_ENABLED=0", "GOFLAGS=")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	run := exec.Command(binary)
	run.Env = []string{"PATH="}
	out, err := run.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "experimental preg83") ||
		!strings.Contains(string(out), "unavailable") || strings.Contains(string(out), "must-not-publish") {
		t.Fatalf("disabled backend did not stop explicitly: %v\n%s", err, out)
	}
}
