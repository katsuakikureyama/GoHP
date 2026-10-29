package tests

import (
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/codegen"
)

// Check the committed real-file fixture's compiler contract without requiring
// a local PHP oracle. Native equivalence remains TestDifferential's job.
func TestStaticClosureFixture(t *testing.T) {
	for _, target := range []string{"8.3", "8.4"} {
		t.Run(target, func(t *testing.T) {
			input, err := loadDifferentialInput("tests/fixtures/static_declaration_closure/manifest.json", target)
			if err != nil {
				t.Fatal(err)
			}
			generated, sourceMap, err := codegen.GenerateWithMap(input.program)
			if err != nil {
				t.Fatal(err)
			}
			if len(generated) == 0 || len(input.program.Sources) != 4 {
				t.Fatal("fixture lost its generated closure or four original sources")
			}
			if err := validateClosureSourceMap(input, sourceMap); err != nil {
				t.Fatal(err)
			}
		})
	}
}
