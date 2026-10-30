package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

const match1CompilerSource = `<?php
function probe(string $subject) {
    $matches = null;
    $found = preg_match('/^(a+)+$/s', $subject, $matches, PREG_OFFSET_CAPTURE, 0);
    return preg_last_error();
}
echo probe('aaab');`

func TestPreg83Match1CompilerLowering(t *testing.T) {
	for _, tc := range []struct{ profile, method string }{
		{ir.Preg83Profile, "PregMatch83Local"},
		{ir.Preg83Match1Profile, "PregMatch83LocalMatch1"},
	} {
		t.Run(tc.profile, func(t *testing.T) {
			program, err := frontend.ParseWithOptions(match1CompilerSource, "8.3", frontend.Options{ExperimentalPCRE2: true, PCRE2Profile: tc.profile})
			if err != nil {
				t.Fatal(err)
			}
			output, mapping, err := GenerateWithMap(program)
			if err != nil {
				t.Fatal(err)
			}
			if program.RegexProfile != tc.profile || !strings.Contains(string(output), "Experimental regex profile: "+tc.profile) {
				t.Fatal("profile identity lost")
			}
			call := "." + tc.method + "("
			found := false
			for line, text := range strings.Split(string(output), "\n") {
				if strings.Contains(text, call) {
					phpLine, ok := mapping.PHPLine(line + 1)
					if !ok || phpLine != 4 {
						t.Fatalf("profile call source map = %d/%v", phpLine, ok)
					}
					found = true
				}
			}
			if !found || !strings.Contains(string(output), ".PregLastError83()") {
				t.Fatal("profile dispatch or shared error read missing")
			}
			if tc.profile == ir.Preg83Profile && strings.Contains(string(output), ".PregMatch83LocalMatch1(") {
				t.Fatal("default budget changed")
			}
		})
	}
}

func TestPreg83Match1ProfileGates(t *testing.T) {
	for _, tc := range []struct {
		name, target string
		options      frontend.Options
	}{
		{"missing_opt_in", "8.3", frontend.Options{PCRE2Profile: ir.Preg83Match1Profile}},
		{"unknown", "8.3", frontend.Options{ExperimentalPCRE2: true, PCRE2Profile: "unknown-budget"}},
		{"php84", "8.4", frontend.Options{ExperimentalPCRE2: true, PCRE2Profile: ir.Preg83Match1Profile}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := frontend.ParseWithOptions(match1CompilerSource, tc.target, tc.options); err == nil || !strings.Contains(err.Error(), "PHP003") {
				t.Fatalf("invalid options accepted: %v", err)
			}
		})
	}
	program, err := frontend.ParseWithOptions(match1CompilerSource, "8.3", frontend.Options{ExperimentalPCRE2: true, PCRE2Profile: ir.Preg83Match1Profile})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ profile, target string }{{ir.Preg83Match1Profile, "8.4"}, {"unknown-budget", "8.3"}, {"", "8.3"}} {
		program.RegexProfile, program.PHPVersion = tc.profile, tc.target
		if output, err := Generate(program); err == nil || len(output) != 0 {
			t.Fatalf("invalid manual profile %q/%q generated: %v", tc.profile, tc.target, err)
		}
	}
}
