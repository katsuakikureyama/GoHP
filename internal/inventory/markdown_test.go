package inventory

import (
	"strings"
	"testing"
)

func TestMarkdownEscapesMetadataAndKeepsUnknownsVisible(t *testing.T) {
	report := Report{
		Status:        "complete",
		LockFreshness: "unknown",
		Notices:       []string{"<script>\nalert(1)</script>"},
		Inputs:        []Input{{Name: "composer|lock", Status: "present", SHA256: "hash"}},
		Root: Package{
			Name:     "root/[pkg]",
			Scope:    "root",
			Licenses: nil,
			Source:   &Reference{Type: "git", Reference: "a*b", URLPresent: true},
		},
		Links: []Link{{
			Owner: "root", Scope: "runtime", Relation: "suggest", Kind: "package",
			Name: "vendor/pkg", Constraint: "<description>", Compatibility: "",
		}},
		PlatformOverrides: []Override{{Source: "config", Name: "php", Value: []byte(`"8.3"`)}},
	}
	got := string(Markdown(report))
	for _, want := range []string{
		"Compatibility: **unverified**",
		"&lt;script&gt; alert(1)&lt;/script&gt;",
		"composer\\|lock",
		"root/\\[pkg\\]",
		"a\\*b",
		"URL present; intentionally omitted",
		"description: &lt;description&gt;",
		"&#34;8.3&#34;",
		"| unknown |",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Markdown missing %q:\n%s", want, got)
		}
	}
	for _, forbidden := range []string{"<script>", "<description>", "&\\#34;"} {
		if strings.Contains(got, forbidden) {
			t.Errorf("Markdown contains unescaped %q:\n%s", forbidden, got)
		}
	}
}

func TestMarkdownDistinguishesEmptyListsFromUnknown(t *testing.T) {
	report := Report{
		Root: Package{
			Name:             "root",
			Licenses:         []string{},
			ScriptEvents:     []string{},
			AutoloadModes:    []string{},
			AutoloadDevModes: []string{},
		},
	}
	got := string(Markdown(report))
	if !strings.Contains(got, "| root | unknown | unknown | unknown | none | unknown | unknown | none | none | none |") {
		t.Fatalf("empty list fields are not explicit:\n%s", got)
	}
}
