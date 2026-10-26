package codegen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestTargetScopedArrayAnyDeclaration(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "..", "tests", "fixtures", "php83_array_any_declaration.php"))
	if err != nil {
		t.Fatal(err)
	}

	php83, err := frontend.Parse(string(source), "8.3")
	if err != nil {
		t.Fatal(err)
	}
	output, err := Generate(php83)
	if err != nil || len(output) == 0 {
		t.Fatalf("PHP 8.3 user declaration did not generate: %v\n%s", err, output)
	}
	for _, want := range []string{
		"func F_array_any(",
		`F_array_any("owned")`,
		"rt.GetType(",
	} {
		if !strings.Contains(string(output), want) {
			t.Fatalf("PHP 8.3 user declaration lowering missing %q:\n%s", want, output)
		}
	}

	php84, err := frontend.Parse(string(source), "8.4")
	if err != nil {
		t.Fatal(err)
	}
	output, err = Generate(php84)
	if err == nil || len(output) != 0 || !strings.Contains(err.Error(), "function name reserved in PHP 8.4: ArRaY_AnY") {
		t.Fatalf("PHP 8.4 reserved declaration generated %q: %v", output, err)
	}
}
