package codegen

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/katsuakikureyama/GoHP/internal/ir"
)

// Mapping relates an inclusive range of physical Go lines to a PHP statement's
// starting line. Ranges are sorted and disjoint, including for nested statements.
type Mapping struct {
	GeneratedStartLine int    `json:"generated_start_line"`
	GeneratedEndLine   int    `json:"generated_end_line"`
	SourceLine         int    `json:"source_line"`
	SourceID           string `json:"source_id,omitempty"`
}

// SourceMap is a GoHP-specific line map, not the JavaScript source map format.
// File identities and hashes are supplied by the enclosing build report.
type SourceMap struct {
	Schema    string          `json:"schema"`
	Precision string          `json:"precision"`
	Mappings  []Mapping       `json:"mappings"`
	Sources   []ir.SourceFile `json:"sources,omitempty"`
}

// PHPLine returns no match for scaffolding, comments or unknown source positions.
// Never infer a location by carrying the preceding mapping into an unmapped gap.
func (m SourceMap) PHPLine(generatedLine int) (int, bool) {
	if m.Schema == "gohp.sourcemap/v0.2" {
		return 0, false
	}
	entry, ok := m.mapping(generatedLine)
	return entry.SourceLine, ok
}

// PHPLocation retains file identity. Line-only maps cannot supply a source ID.
func (m SourceMap) PHPLocation(generatedLine int) (ir.Location, bool) {
	if m.Schema != "gohp.sourcemap/v0.2" {
		return ir.Location{}, false
	}
	entry, ok := m.mapping(generatedLine)
	if !ok || entry.SourceID == "" {
		return ir.Location{}, false
	}
	return ir.Location{SourceID: entry.SourceID, Line: entry.SourceLine}, true
}

func (m SourceMap) mapping(generatedLine int) (Mapping, bool) {
	i := sort.Search(len(m.Mappings), func(i int) bool {
		return m.Mappings[i].GeneratedEndLine >= generatedLine
	})
	if i == len(m.Mappings) || m.Mappings[i].GeneratedStartLine > generatedLine {
		return Mapping{}, false
	}
	return m.Mappings[i], true
}

// Scope markers survive gofmt without relying on its layout decisions. They are
// ordinary comments, not Go //line directives: paths and PHP text are never
// interpolated into them. Quoted PHP strings cannot inject standalone markers.
func mapStatements(source []byte, sources ...ir.SourceFile) (SourceMap, error) {
	m := SourceMap{Schema: "gohp.sourcemap/v0.1", Precision: "statement-start-line", Mappings: []Mapping{}}
	if len(sources) > 0 {
		m.Schema = "gohp.sourcemap/v0.2"
		m.Precision = "statement-start-file-line"
		m.Sources = append([]ir.SourceFile(nil), sources...)
	}
	var stack []ir.Location
	for i, raw := range strings.Split(string(source), "\n") {
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, "// gohp:begin ") {
			fields := strings.Fields(strings.TrimPrefix(line, "// gohp:begin "))
			want := 1
			if len(sources) > 0 {
				want = 2
			}
			if len(fields) != want {
				return SourceMap{}, fmt.Errorf("invalid generated source marker arity")
			}
			n, err := strconv.Atoi(fields[0])
			if err != nil {
				return SourceMap{}, fmt.Errorf("invalid generated source marker: %w", err)
			}
			location := ir.Location{Line: n}
			if len(sources) > 0 {
				index, err := strconv.Atoi(fields[1])
				if err != nil || index < 1 || index > len(sources) {
					return SourceMap{}, fmt.Errorf("invalid generated source identity marker")
				}
				location.SourceID = sources[index-1].ID
			}
			stack = append(stack, location)
			continue
		}
		if line == "// gohp:end" {
			if len(stack) == 0 {
				return SourceMap{}, fmt.Errorf("unbalanced generated source markers")
			}
			stack = stack[:len(stack)-1]
			continue
		}
		if len(stack) == 0 || stack[len(stack)-1].Line <= 0 || line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		location, goLine := stack[len(stack)-1], i+1
		if n := len(m.Mappings); n > 0 && m.Mappings[n-1].SourceLine == location.Line && m.Mappings[n-1].SourceID == location.SourceID && m.Mappings[n-1].GeneratedEndLine == goLine-1 {
			m.Mappings[n-1].GeneratedEndLine = goLine
		} else {
			m.Mappings = append(m.Mappings, Mapping{GeneratedStartLine: goLine, GeneratedEndLine: goLine, SourceLine: location.Line, SourceID: location.SourceID})
		}
	}
	if len(stack) != 0 {
		return SourceMap{}, fmt.Errorf("unclosed generated source markers")
	}
	return m, nil
}
