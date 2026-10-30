package sourceclosure

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestAssembleMatch1Profile(t *testing.T) {
	files := map[string]string{
		"entry.php":  "<?php\nrequire_once __DIR__ . '/helper.php';\necho probe('aaab');\n",
		"helper.php": "<?php\nfunction probe(string $subject) { $matches = null; $found = preg_match('/^(a+)+$/s', $subject, $matches, PREG_OFFSET_CAPTURE, 0); return preg_last_error(); }\n",
	}
	root, manifest := writeFixture(t, files, []string{"8.3"})
	program, report, err := Assemble(manifest, root, "entry.php", "8.3", frontend.Options{ExperimentalPCRE2: true, PCRE2Profile: ir.Preg83Match1Profile})
	if err != nil {
		t.Fatal(err)
	}
	if program.RegexProfile != ir.Preg83Match1Profile || report.RegexProfile != ir.Preg83Match1Profile || len(report.Modules) != 2 || len(program.Sources) != 2 {
		t.Fatal("closure profile/source identities lost")
	}
	if len(program.Functions) != 1 || program.Functions[0].SourceID != "helper.php" {
		t.Fatal("original dependency declaration identity lost")
	}
}

func TestRejectMatch1OptionsBeforeLoading(t *testing.T) {
	for _, tc := range []struct {
		target  string
		options frontend.Options
	}{
		{"8.3", frontend.Options{PCRE2Profile: ir.Preg83Match1Profile}},
		{"8.3", frontend.Options{ExperimentalPCRE2: true, PCRE2Profile: "unknown-budget"}},
		{"8.4", frontend.Options{ExperimentalPCRE2: true, PCRE2Profile: ir.Preg83Match1Profile}},
	} {
		_, report, err := Assemble("/missing-manifest", "/missing-source", "entry.php", tc.target, tc.options)
		if err == nil || !strings.Contains(err.Error(), "PHP003") {
			t.Fatalf("profile not rejected before loading: %v", err)
		}
		assertZeroReport(t, report)
	}
}
