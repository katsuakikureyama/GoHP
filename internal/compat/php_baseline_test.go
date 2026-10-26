package compat

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func testPHPBaseline(version string) *PHPBaseline {
	empty, locale := "", "C"
	b := &PHPBaseline{Scope: "cli-startup-baseline", PHPVersion: version, SAPI: "cli", IntSize: 8, Timezone: "UTC", INI: map[string]*string{}, Locales: map[string]*string{}, Extensions: []PHPExtension{{Name: "Core", Version: &version}}, ZendExtensions: []PHPExtension{}}
	for _, key := range PHPINIKeys() {
		b.INI[key] = &empty
	}
	for _, key := range PHPLocaleKeys() {
		b.Locales[key] = &locale
	}
	return b
}

func TestPHPBaselineArchiveSchemas(t *testing.T) {
	legacyDir, legacy := savedRun(t, "8.3", "pass")
	currentDir, current := savedRun(t, "8.3", "pass", RunSchema)
	baselineDir, baseline := savedRun(t, "8.3", "pass", RunSchemaBaseline)
	for _, tt := range []struct {
		dir      string
		expected Report
	}{{legacyDir, legacy}, {baselineDir, baseline}, {currentDir, current}} {
		actual, err := LoadArchive(tt.dir)
		if err != nil || !reflect.DeepEqual(actual, tt.expected) {
			t.Fatalf("archive %s: %v", tt.expected.Schema, err)
		}
	}
	view, err := BuildView([]Report{legacy, current}, ViewFilter{Kind: "builtin"})
	if err != nil || view.Schema != "gohp.compatibility.view/v0.4" || view.Runs[0].Metadata.PHPBaseline != nil || view.Runs[1].Metadata.PHPBaseline == nil {
		t.Fatalf("mixed archive view lost baseline distinction: %+v, %v", view, err)
	}
	text := string(ViewMarkdown(view))
	for _, expected := range []string{"not recorded (legacy archive)", "Separate pre-suite CLI probe only", "ini.precision", "locale.LC", current.Metadata.PHPBaselineSHA256} {
		if !strings.Contains(text, expected) {
			t.Fatalf("view missing %q", expected)
		}
	}
	for _, tt := range []struct {
		name   string
		mutate func(*Report)
	}{
		{"missing", func(r *Report) { r.Metadata.PHPBaseline = nil }},
		{"hash", func(r *Report) { r.Metadata.PHPBaselineSHA256 = strings.Repeat("f", 64) }},
		{"changed_setting", func(r *Report) { v := "30"; r.Metadata.PHPBaseline.INI["precision"] = &v }},
		{"downgrade", func(r *Report) { r.Schema = RunSchemaLegacy }},
		{"identity", func(r *Report) {
			r.Metadata.PHPBaseline.PHPVersion = "8.4.30"
			r.Metadata.PHPBaselineSHA256 = PHPBaselineHash(r.Metadata.PHPBaseline)
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir, r := savedRun(t, "8.3", "pass", RunSchema)
			tt.mutate(&r)
			writeTestJSON(t, filepath.Join(dir, "run.json"), r)
			if _, err := LoadArchive(dir); err == nil {
				t.Fatal("invalid baseline archive accepted")
			}
		})
	}
	legacy.Schema = RunSchema
	writeTestJSON(t, filepath.Join(legacyDir, "run.json"), legacy)
	if _, err := LoadArchive(legacyDir); err == nil {
		t.Fatal("legacy archive acquired new guarantees")
	}
}

func TestPHPBaselineProjection(t *testing.T) {
	for _, tt := range []struct {
		name   string
		mutate func(*PHPBaseline)
		valid  bool
	}{
		{"complete", func(*PHPBaseline) {}, true},
		{"unknowns", func(b *PHPBaseline) {
			b.INI["assert.exception"] = nil
			b.Locales["LC_MESSAGES"] = nil
			b.Extensions[0].Version = nil
		}, true},
		{"missing_ini", func(b *PHPBaseline) { delete(b.INI, "precision") }, false},
		{"unknown_ini", func(b *PHPBaseline) { b.INI["secret"] = nil }, false},
		{"missing_locale", func(b *PHPBaseline) { delete(b.Locales, "LC_TIME") }, false},
		{"empty_locale", func(b *PHPBaseline) { v := ""; b.Locales["LC_TIME"] = &v }, false},
		{"duplicate_extension", func(b *PHPBaseline) { b.Extensions = append(b.Extensions, b.Extensions[0]) }, false},
		{"unsorted_extensions", func(b *PHPBaseline) { b.Extensions = append([]PHPExtension{{Name: "Z"}}, b.Extensions...) }, false},
		{"missing_extensions", func(b *PHPBaseline) { b.Extensions = nil }, false},
		{"missing_zend_extensions", func(b *PHPBaseline) { b.ZendExtensions = nil }, false},
		{"scope", func(b *PHPBaseline) { b.Scope = "per-fixture" }, false},
		{"sapi", func(b *PHPBaseline) { b.SAPI = "cgi" }, false},
		{"int_size", func(b *PHPBaseline) { b.IntSize = 0 }, false},
		{"timezone", func(b *PHPBaseline) { b.Timezone = "" }, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			b := testPHPBaseline("8.3.30")
			tt.mutate(b)
			if err := b.Validate("8.3.30"); (err == nil) != tt.valid {
				t.Fatalf("valid=%v: %v", tt.valid, err)
			}
		})
	}
	b := testPHPBaseline("8.3.30")
	before := PHPBaselineHash(b)
	b.INI["precision"] = nil
	if PHPBaselineHash(b) == before {
		t.Fatal("empty and unavailable configuration conflated")
	}
	b.Locales["LC_CTYPE"] = stringPointer("<script>|[x](bad)")
	m := testMetadata()
	m.PHPBaseline = b
	if text := string(Markdown(Report{Metadata: m})); strings.Contains(text, "<script>") || strings.Contains(text, "[x](bad)") {
		t.Fatal("baseline Markdown contains unescaped observation")
	}
}

func stringPointer(s string) *string { return &s }
