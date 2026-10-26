package frontend

import (
	"os"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/phpversion"
)

func testTarget(t *testing.T) string {
	t.Helper()
	target := os.Getenv("GOHP_PHP_TARGET")
	if target == "" {
		target = phpversion.DefaultTarget
	}
	if err := phpversion.ValidateTarget(target); err != nil {
		t.Fatal(err)
	}
	return target
}
