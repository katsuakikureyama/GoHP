package compat

import (
	"bytes"
	"fmt"
	"strings"
)

// Markdown renders the complete catalog target matrix. It is intentionally a
// compact index; machine consumers should use Report's JSON form.
func Markdown(r Report) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "# GoHP compatibility evidence\n\n")
	fmt.Fprintf(&b, "Overall: **%s**  \n", md(r.Overall))
	fmt.Fprintf(&b, "Run target: `%s`  \n", md(r.Metadata.PHPTarget))
	fmt.Fprintf(&b, "Catalog SHA-256: `%s`  \n", md(r.CatalogSHA256))
	fmt.Fprintf(&b, "Event SHA-256: `%s`\n\n", md(r.EventSHA256))
	commit := r.Metadata.Commit
	if commit == "" {
		commit = "unknown (run-scoped only)"
	}
	fmt.Fprintf(&b, "Revision: `%s` (tree: `%s`, dirty: `%s`, provenance: `%s`)\n\n", md(commit), md(orUnknown(r.Metadata.Tree)), md(orUnknown(r.Metadata.Dirty)), md(orUnknown(r.Metadata.Provenance)))
	if r.Metadata.Commit == "" || r.Metadata.Dirty != "clean" {
		b.WriteString("This evidence is run-scoped and must not be treated as a current-release pass.\n\n")
	}
	b.WriteString("`differential-pass (case-limited)` covers only its listed fixture cases. Stderr, exception and side-effect comparisons are not measured in schema v0.1.\n\n")
	b.WriteString("| Feature | PHP target | Implementation | Status | Contract / limitations | Evidence cases |\n")
	b.WriteString("|---|---|---|---|---|---|\n")
	for _, feature := range r.Features {
		limits := "none recorded"
		if len(feature.Limitations) != 0 {
			limits = strings.Join(feature.Limitations, "; ")
		}
		if feature.Status == "not-measured" {
			limits = "not measured in this run; " + limits
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s — %s | %s |\n", md(feature.ID), md(feature.Target), md(feature.Implementation), md(feature.Status), md(feature.Contract), md(limits), md(strings.Join(feature.Checks, ", ")))
	}
	b.WriteString("\n## Cases\n\n")
	b.WriteString("| Case | Kind | Status | Detail |\n|---|---|---|---|\n")
	for _, testCase := range r.Cases {
		fmt.Fprintf(&b, "| %s | %s | %s | %s |\n", md(testCase.ID), md(testCase.Kind), md(testCase.Status), md(testCase.Detail))
	}
	return b.Bytes()
}

func md(value string) string {
	value = strings.ReplaceAll(value, "|", "\\|")
	value = strings.ReplaceAll(value, "\r\n", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\r", " ")
	return value
}
func orUnknown(value string) string {
	if value == "" {
		return "unknown"
	}
	return value
}
