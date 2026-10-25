package inventory

import (
	"bytes"
	"fmt"
	"html"
	"strings"
)

// Markdown renders Composer metadata as an inventory. It deliberately keeps
// the result separate from GoHP's bounded compatibility evidence: an inventory
// never certifies package execution or license suitability.
func Markdown(report Report) []byte {
	var b bytes.Buffer
	b.WriteString("# GoHP Composer inventory\n\n")
	fmt.Fprintf(&b, "Inventory status: **%s**  \n", text(report.Status))
	b.WriteString("Compatibility: **unverified** — this is metadata only, not a compatibility PASS.\n\n")

	b.WriteString("## Notices\n\n")
	if len(report.Notices) == 0 {
		b.WriteString("- unknown\n\n")
	} else {
		for _, notice := range report.Notices {
			fmt.Fprintf(&b, "- %s\n", text(notice))
		}
		b.WriteByte('\n')
	}

	b.WriteString("## Inputs and lock state\n\n")
	b.WriteString("| Input | Presence / status | SHA-256 |\n|---|---|---|\n")
	for _, input := range report.Inputs {
		fmt.Fprintf(&b, "| %s | %s | %s |\n", text(input.Name), text(input.Status), text(input.SHA256))
	}
	if len(report.Inputs) == 0 {
		b.WriteString("| unknown | unknown | unknown |\n")
	}
	fmt.Fprintf(&b, "\nLock freshness: **%s**  \n", text(report.LockFreshness))
	fmt.Fprintf(&b, "Lock content hash: %s\n\n", text(report.LockContentHash))

	b.WriteString("## Root package\n\n")
	packageTable(&b, []Package{report.Root})
	b.WriteString("\n## Locked packages\n\n")
	if len(report.Packages) == 0 {
		b.WriteString("No locked packages recorded (or unknown).\n\n")
	} else {
		packageTable(&b, report.Packages)
		b.WriteByte('\n')
	}

	b.WriteString("## Dependency links\n\n")
	b.WriteString("| Owner | Scope | Relation | Kind | Package | Constraint / description | Compatibility |\n|---|---|---|---|---|---|---|\n")
	if len(report.Links) == 0 {
		b.WriteString("| unknown | unknown | unknown | unknown | unknown | unknown | unverified |\n")
	} else {
		for _, link := range report.Links {
			fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s | %s |\n", text(link.Owner), text(link.Scope), text(link.Relation), text(link.Kind), text(link.Name), linkText(link), text(link.Compatibility))
		}
	}
	b.WriteString("\n## Platform overrides\n\n")
	b.WriteString("| Source | Name | Value |\n|---|---|---|\n")
	if len(report.PlatformOverrides) == 0 {
		b.WriteString("| unknown | unknown | unknown |\n")
	} else {
		for _, override := range report.PlatformOverrides {
			fmt.Fprintf(&b, "| %s | %s | %s |\n", text(override.Source), text(override.Name), rawText(override.Value))
		}
	}
	return b.Bytes()
}

func packageTable(b *bytes.Buffer, packages []Package) {
	b.WriteString("| Package | Scope | Version | Type | Licenses | Source reference | Dist reference | Script events | Autoload modes | Dev autoload modes |\n|---|---|---|---|---|---|---|---|---|---|\n")
	for _, pkg := range packages {
		fmt.Fprintf(b, "| %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
			text(pkg.Name), text(pkg.Scope), text(pkg.Version), text(pkg.Type), list(pkg.Licenses), referenceText(pkg.Source), referenceText(pkg.Dist), list(pkg.ScriptEvents), list(pkg.AutoloadModes), list(pkg.AutoloadDevModes))
	}
}

func linkText(link Link) string {
	if link.Relation == "suggest" {
		return "description: " + text(link.Constraint)
	}
	return text(link.Constraint)
}

// referenceText intentionally reports identity material and URL presence, but
// never the URL itself: package URLs can carry transient credentials or other
// sensitive query material and are not needed for this inventory.
func referenceText(ref *Reference) string {
	if ref == nil {
		return "unknown"
	}
	parts := []string{
		"type=" + text(ref.Type),
		"reference=" + text(ref.Reference),
		"shasum=" + text(ref.Shasum),
	}
	if ref.URLPresent {
		parts = append(parts, "URL present; intentionally omitted")
	} else {
		parts = append(parts, "URL not recorded")
	}
	return strings.Join(parts, "; ")
}

func list(values []string) string {
	if values == nil {
		return "unknown"
	}
	if len(values) == 0 {
		return "none"
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, text(value))
	}
	return strings.Join(result, ", ")
}

func rawText(value []byte) string {
	if len(value) == 0 {
		return "unknown"
	}
	return text(string(value))
}

func text(value string) string {
	if value == "" {
		return "unknown"
	}
	value = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return ' '
		}
		return r
	}, value)
	value = strings.NewReplacer(
		"\\", "\\\\", "`", "\\`", "*", "\\*", "_", "\\_", "[", "\\[", "]", "\\]", "|", "\\|", "#", "\\#",
	).Replace(value)
	// Escape HTML last so Markdown escaping cannot corrupt numeric entities
	// such as &#34; into visible &\#34; text.
	return html.EscapeString(value)
}
