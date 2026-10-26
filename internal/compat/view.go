package compat

import (
	"bytes"
	"fmt"
	"html"
	"strings"
)

type ViewFilter struct {
	Target string `json:"target,omitempty"`
	Kind   string `json:"kind,omitempty"`
	Status string `json:"status,omitempty"`
}

// View never combines different runs into one compatibility verdict. Runs stay
// in caller-provided order, not inferred temporal or revision order.
type View struct {
	Schema string     `json:"schema"`
	Notice string     `json:"notice"`
	Filter ViewFilter `json:"filter"`
	Runs   []ViewRun  `json:"runs"`
}

type ViewRun struct {
	RunSchema     string         `json:"run_schema"`
	Metadata      Metadata       `json:"metadata"`
	Overall       string         `json:"overall"`
	CatalogSHA256 string         `json:"catalog_sha256"`
	EventSHA256   string         `json:"event_sha256"`
	TotalFeatures int            `json:"total_features"`
	Features      []FeatureRow   `json:"features"`
	Cases         []CaseEvidence `json:"cases"`
}

const viewNotice = "Stored run evidence only; no current-release or combined-target support verdict. Retired PHP targets and below-policy oracle versions are historical records, not current support. Internal consistency is checked, not authenticity, source bytes or CI job success. No tests are executed. Differential results cover listed cases and stdout/successful exits. Run/v0.3 and later additionally require empty stderr on both sides under quiet-success/v1; v0.4 records stdout hashes and exit codes. Arbitrary stderr/warning equivalence, exceptions and side effects are not checked. Input order is preserved; no latest/best result is selected."

// Validate rejects unknown filter values before any archive is read.
func (filter ViewFilter) Validate() error {
	if filter.Target != "" && !knownEvidenceTarget(filter.Target) {
		return fmt.Errorf("invalid target filter %q", filter.Target)
	}
	if filter.Kind != "" && !oneOf(filter.Kind, "syntax", "operator", "type", "builtin", "extension", "tooling", "application") {
		return fmt.Errorf("invalid kind filter %q", filter.Kind)
	}
	if filter.Status != "" && !oneOf(filter.Status, "unsupported", "unverified", "compiler-check-pass", "differential-pass (case-limited)", "failed", "incomplete", "not-measured") {
		return fmt.Errorf("invalid status filter %q", filter.Status)
	}
	return nil
}

// BuildView selects rows from already loaded archives. Run metadata, original
// overall result and all case details survive filters, including empty matches.
func BuildView(reports []Report, filter ViewFilter) (View, error) {
	if len(reports) == 0 {
		return View{}, fmt.Errorf("at least one saved run is required")
	}
	if err := filter.Validate(); err != nil {
		return View{}, err
	}
	view := View{Schema: "gohp.compatibility.view/v0.4", Notice: viewNotice, Filter: filter, Runs: make([]ViewRun, 0, len(reports))}
	for _, report := range reports {
		run := ViewRun{RunSchema: report.Schema, Metadata: report.Metadata, Overall: report.Overall, CatalogSHA256: report.CatalogSHA256, EventSHA256: report.EventSHA256, TotalFeatures: len(report.Features), Features: []FeatureRow{}, Cases: report.Cases}
		for _, row := range report.Features {
			if (filter.Target == "" || row.Target == filter.Target) && (filter.Kind == "" || row.Kind == filter.Kind) && (filter.Status == "" || row.Status == filter.Status) {
				run.Features = append(run.Features, row)
			}
		}
		view.Runs = append(view.Runs, run)
	}
	return view, nil
}

func ViewMarkdown(view View) []byte {
	var b bytes.Buffer
	b.WriteString("# GoHP saved compatibility views\n\n")
	fmt.Fprintf(&b, "%s\n\n", viewText(view.Notice))
	fmt.Fprintf(&b, "Filters: target=%s; kind=%s; status=%s.\n\n", viewText(orAll(view.Filter.Target)), viewText(orAll(view.Filter.Kind)), viewText(orAll(view.Filter.Status)))
	for i, run := range view.Runs {
		m := run.Metadata
		fmt.Fprintf(&b, "## Input %d — PHP %s\n\n", i+1, viewText(m.PHPTarget))
		fmt.Fprintf(&b, "Stored overall: **%s**. Showing %d of %d feature/target rows; filters never change overall.\n\n", viewText(run.Overall), len(run.Features), run.TotalFeatures)
		// Keep configuration and revision differences visible even in filtered
		// views. No archive text is emitted as explicit Markdown or HTML links.
		for _, item := range [][2]string{
			{"Commit", m.Commit}, {"Tree", m.Tree}, {"Dirty", m.Dirty}, {"Provenance", m.Provenance},
			{"Started", m.StartedAt}, {"Run ID / attempt", m.RunID + " / " + m.Attempt}, {"Run URL", m.RunURL},
			{"Go", m.GoVersion}, {"Platform", m.GOOS + "/" + m.GOARCH}, {"PHP oracle", m.PHPVersion},
			{"PHP image", m.Image}, {"Go image", m.GoImage}, {"Command", m.Command}, {"Build flags", m.BuildFlags},
			{"Source SHA-256 (recorded, not rehashed)", m.SourceSHA256}, {"Catalog SHA-256", run.CatalogSHA256}, {"Events SHA-256", run.EventSHA256},
		} {
			fmt.Fprintf(&b, "%s: %s  \n", item[0], viewText(orUnknown(item[1])))
		}
		fmt.Fprintf(&b, "Test process exit: %d\n\n", m.TestExit)
		fmt.Fprintf(&b, "Recorded run schema: %s\n\n", viewText(run.RunSchema))
		writeBaselineMarkdown(&b, m)
		if len(run.Features) == 0 {
			b.WriteString("No matching feature rows. This is not a passing compatibility verdict.\n\n")
		} else {
			b.WriteString("| Feature / name | Kind | PHP target | Implementation | Stored status | Contract / limitations | Evidence cases |\n|---|---|---|---|---|---|---|\n")
			for _, row := range run.Features {
				fmt.Fprintf(&b, "| %s / %s | %s | %s | %s | %s | %s — %s | %s |\n", viewText(row.ID), viewText(row.Name), viewText(row.Kind), viewText(row.Target), viewText(row.Implementation), viewText(row.Status), viewText(row.Contract), viewText(strings.Join(row.Limitations, "; ")), viewText(strings.Join(row.Checks, ", ")))
			}
			b.WriteString("\n")
		}
		b.WriteString("All recorded cases (not filtered):\n\n| Case | Kind | Status | Quiet stderr | Detail |\n|---|---|---|---|---|\n")
		for _, c := range run.Cases {
			fmt.Fprintf(&b, "| %s | %s | %s | %s | %s |\n", viewText(c.ID), viewText(c.Kind), viewText(c.Status), viewText(quietStderrStatus(c)), viewText(c.Detail))
		}
		b.WriteString("\n")
	}
	return b.Bytes()
}

func orAll(s string) string {
	if s == "" {
		return "all"
	}
	return s
}

func viewText(s string) string {
	s = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return ' '
		}
		return r
	}, s)
	s = html.EscapeString(s)
	return strings.NewReplacer("\\", "\\\\", "`", "\\`", "*", "\\*", "_", "\\_", "[", "\\[", "]", "\\]", "|", "\\|", "#", "\\#").Replace(s)
}
