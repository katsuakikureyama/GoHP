package tests

import (
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/codegen"
)

// Validate the real committed differential fixture even without local PHP;
// independent native equivalence remains the oracle harness's responsibility.
func TestStringEmptyFixture(t *testing.T) {
	for _, target := range []string{"8.3", "8.4"} {
		t.Run(target, func(t *testing.T) {
			input, err := loadDifferentialInput("tests/fixtures/string_empty.php", target)
			if err != nil {
				t.Fatal(err)
			}
			if _, _, err := codegen.GenerateWithMap(input.program); err != nil {
				t.Fatal(err)
			}
		})
	}
}
