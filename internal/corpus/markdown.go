package corpus

import (
	"bytes"
	"fmt"
	"html"
	"strings"
)

// Markdown renders an inspection report without implying that the declared
// source identity or the selected bytes have been executed by either PHP or
// GoHP. It is deliberately a source-selection view, not compatibility proof.
func Markdown(report Report) []byte {
	var b bytes.Buffer
	b.WriteString("# GoHP pinned source corpus\n\n")
	fmt.Fprintf(&b, "Inspection status: **%s**  \n", markdownText(report.Status))
	fmt.Fprintf(&b, "Source identity: **%s** — file bytes are checked against this manifest; the declared repository and revision are not authenticated by inspection.  \n", markdownText(report.Status))
	b.WriteString("Compatibility: **unverified** — no GoHP compilation, generated-binary execution, native PHP execution, or output oracle comparison was observed.  \n")
	b.WriteString("Analysis: **not-measured** — this inspection records whole files and byte hashes only.\n\n")

	b.WriteString("## Declared source selection\n\n")
	b.WriteString("| Field | Declared value |\n|---|---|\n")
	fmt.Fprintf(&b, "| Corpus ID | %s |\n", markdownText(report.Manifest.ID))
	fmt.Fprintf(&b, "| Project | %s |\n", markdownText(report.Manifest.Project))
	fmt.Fprintf(&b, "| Version | %s |\n", markdownText(report.Manifest.Version))
	fmt.Fprintf(&b, "| Repository | %s |\n", markdownText(report.Manifest.Source.Repository))
	fmt.Fprintf(&b, "| Revision | %s |\n", markdownText(report.Manifest.Source.Revision))
	fmt.Fprintf(&b, "| PHP targets | %s |\n", list(report.Manifest.PHPTargets))
	fmt.Fprintf(&b, "| Selection unit | %s |\n", markdownText(report.Manifest.Selection.Unit))
	fmt.Fprintf(&b, "| Dependency closure | %s |\n", markdownText(report.Manifest.Selection.DependencyClosure))
	fmt.Fprintf(&b, "| Selection description | %s |\n", markdownText(report.Manifest.Selection.Description))
	fmt.Fprintf(&b, "| Manifest SHA-256 | %s |\n\n", markdownText(report.ManifestSHA256))

	b.WriteString("## Whole-file measurements\n\n")
	b.WriteString("These are whole-file selection totals, not function, syntax, call-site, compilation, execution, coverage, or compatibility measurements.\n\n")
	b.WriteString("| Role | Files | Bytes |\n|---|---:|---:|\n")
	fmt.Fprintf(&b, "| Subject | %d | %d |\n", report.Measurements.SubjectFiles, report.Measurements.SubjectBytes)
	fmt.Fprintf(&b, "| Provenance | %d | %d |\n\n", report.Measurements.ProvenanceFiles, report.Measurements.ProvenanceBytes)

	b.WriteString("| Unmeasured item | Value |\n|---|---|\n")
	b.WriteString("| Function units | n/a (not measured) |\n")
	b.WriteString("| Syntax nodes | n/a (not measured) |\n")
	b.WriteString("| Call sites | n/a (not measured) |\n")
	b.WriteString("| Compiled units | n/a (not measured) |\n")
	b.WriteString("| Executed cases | n/a (not measured) |\n\n")

	b.WriteString("## Pinned files\n\n")
	b.WriteString("| Path | Role | Bytes | SHA-256 | Inspection |\n|---|---|---:|---|---|\n")
	if len(report.Files) == 0 {
		b.WriteString("| unknown | unknown | n/a | unknown | unknown |\n")
	} else {
		for _, file := range report.Files {
			fmt.Fprintf(&b, "| %s | %s | %d | %s | %s |\n", markdownText(file.Path), markdownText(file.Role), file.Bytes, markdownText(file.SHA256), markdownText(file.Status))
		}
	}

	b.WriteString("\n## Notices\n\n")
	if len(report.Notices) == 0 {
		b.WriteString("- none\n")
	} else {
		for _, notice := range report.Notices {
			fmt.Fprintf(&b, "- %s\n", markdownText(notice))
		}
	}
	return b.Bytes()
}

func list(values []string) string {
	if len(values) == 0 {
		return "unknown"
	}
	items := make([]string, 0, len(values))
	for _, value := range values {
		items = append(items, markdownText(value))
	}
	return strings.Join(items, ", ")
}

// markdownText makes untrusted declared metadata safe to render in a single Markdown
// table cell or list item. Escaping HTML last preserves Markdown escapes.
func markdownText(value string) string {
	if value == "" {
		return "unknown"
	}
	value = strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ").Replace(value)
	value = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return ' '
		}
		return r
	}, value)
	value = strings.NewReplacer(
		"\\", "\\\\", "`", "\\`", "*", "\\*", "_", "\\_", "[", "\\[", "]", "\\]", "|", "\\|", "#", "\\#",
	).Replace(value)
	return html.EscapeString(value)
}
