package tests

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/codegen"
	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

// Property inputs retain their deliberately unknown domain until the existing
// PregMatch83Local boundary. A rejected call must not publish either output
// local or execute the following statement. These cases intentionally build
// without CGO: each failure occurs before a regex backend could be reached.
func TestPreg83PropertyInputsRuntimeGuards(t *testing.T) {
	const prefix = `<?php
class Input {
    public $subject = null;
    public $offset = null;
    public $child = null;
}
function pass($value) { return $value; }
`
	for _, tt := range []struct {
		name   string
		source string
		want   string
	}{
		{
			name: "subject_nonstring",
			source: `$input = new Input();
$input->subject = pass(false);
$input->offset = 0;
$matches = null;
$result = preg_match('/a/s', $input->subject, $matches, PREG_OFFSET_CAPTURE, $input->offset);
echo 'published';`,
			want: "experimental preg83 requires actual string/string/int inputs",
		},
		{
			name: "offset_nonint",
			source: `$input = new Input();
$input->subject = 'a';
$input->offset = pass('0');
$matches = null;
$result = preg_match('/a/s', $input->subject, $matches, PREG_OFFSET_CAPTURE, $input->offset);
echo 'published';`,
			want: "experimental preg83 requires actual string/string/int inputs",
		},
		{
			name: "null_intermediate",
			source: `$input = new Input();
$matches = null;
$result = preg_match('/a/s', $input->child->subject, $matches, PREG_OFFSET_CAPTURE, 0);
echo 'published';`,
			want: "interface conversion",
		},
		{
			name: "scalar_intermediate",
			source: `$input = new Input();
$input->child = pass(false);
$matches = null;
$result = preg_match('/a/s', $input->child->subject, $matches, PREG_OFFSET_CAPTURE, 0);
echo 'published';`,
			want: "interface conversion",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			program, err := frontend.ParseWithOptions(prefix+tt.source, "8.3", frontend.Options{ExperimentalPCRE2: true})
			if err != nil {
				t.Fatal(err)
			}
			generated, err := codegen.Generate(program)
			if err != nil {
				t.Fatal(err)
			}
			pregCall := bytes.Index(generated, []byte("PregMatch83Local("))
			publication := bytes.Index(generated, []byte(`rt.Echo("published")`))
			if pregCall < 0 || publication < 0 || pregCall >= publication {
				t.Fatalf("guard call must precede publication:\n%s", generated)
			}

			dir := t.TempDir()
			file, binary := filepath.Join(dir, "main.go"), filepath.Join(dir, "app")
			if err := os.WriteFile(file, generated, 0600); err != nil {
				t.Fatal(err)
			}
			build := exec.Command("go", "build", "-o", binary, file)
			build.Env = append(os.Environ(), "CGO_ENABLED=0", "GOFLAGS=")
			if output, err := build.CombinedOutput(); err != nil {
				t.Fatalf("build: %v\n%s", err, output)
			}
			run := exec.Command(binary)
			run.Env = []string{"PATH="}
			output, err := run.CombinedOutput()
			if err == nil {
				t.Fatalf("expected property guard rejection, output=%s", output)
			}
			if tt.want != "" && !strings.Contains(string(output), tt.want) {
				t.Fatalf("guard output = %s, want %q", output, tt.want)
			}
			if bytes.Contains(output, []byte("published")) {
				t.Fatalf("guard failure published later output: %s", output)
			}
		})
	}
}
