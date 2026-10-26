package tests

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/codegen"
	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/phpversion"
)

// This checks a version-dependent declaration boundary, not PHP diagnostics
// equivalence or a generated-Go differential execution.
func TestNativeTargetDeclarationBoundary(t *testing.T) {
	target := phpTarget(t)
	version, err := exec.Command("php", "-r", "echo PHP_VERSION;").Output()
	if err != nil {
		t.Fatal(err)
	}
	if err := phpversion.ValidateOracle(target, string(version)); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"array_find", "array_find_key", "array_any", "array_all"} {
		source := "function " + name + "($x) { return $x; } echo " + name + "(7);"
		output, nativeErr := exec.Command("php", "-r", source).CombinedOutput()
		program, err := frontend.Parse("<?php "+source, target)
		if err != nil {
			t.Fatal(err)
		}
		_, compileErr := codegen.Generate(program)
		if target == "8.4" {
			if nativeErr == nil || !strings.Contains(string(output), "Cannot redeclare") || compileErr == nil {
				t.Fatalf("%s: PHP=%q %v compiler=%v", name, output, nativeErr, compileErr)
			}
		} else if nativeErr != nil || string(output) != "7" || compileErr != nil {
			t.Fatalf("%s: PHP=%q %v compiler=%v", name, output, nativeErr, compileErr)
		}
	}
	t.Logf("native declaration boundary checked for target %s, actual PHP %s; diagnostic equivalence not claimed", target, version)
}
