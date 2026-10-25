package codegen

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Mapping relates an inclusive range of physical Go lines to a PHP statement's
// starting line. Ranges are sorted and disjoint, including for nested statements.
type Mapping struct {
	GeneratedStartLine int `json:"generated_start_line"`
	GeneratedEndLine   int `json:"generated_end_line"`
	SourceLine         int `json:"source_line"`
}

// SourceMap is a GoHP-specific line map, not the JavaScript source map format.
// File identities and hashes are supplied by the enclosing build report.
type SourceMap struct {
	Schema    string    `json:"schema"`
	Precision string    `json:"precision"`
	Mappings  []Mapping `json:"mappings"`
}

// PHPLine returns no match for scaffolding, comments or unknown source positions.
// Never infer a location by carrying the preceding mapping into an unmapped gap.
func (m SourceMap) PHPLine(generatedLine int) (int, bool) {
	i := sort.Search(len(m.Mappings), func(i int) bool {
		return m.Mappings[i].GeneratedEndLine >= generatedLine
	})
	if i == len(m.Mappings) || m.Mappings[i].GeneratedStartLine > generatedLine {
		return 0, false
	}
	return m.Mappings[i].SourceLine, true
}

// Scope markers survive gofmt without relying on its layout decisions. They are
// ordinary comments, not Go //line directives: paths and PHP text are never
// interpolated into them. Quoted PHP strings cannot inject standalone markers.
func mapStatements(source []byte) (SourceMap, error) {
	m := SourceMap{Schema: "gohp.sourcemap/v0.1", Precision: "statement-start-line", Mappings: []Mapping{}}
	var stack []int
	for i, raw := range strings.Split(string(source), "\n") {
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, "// gohp:begin ") {
			n, err := strconv.Atoi(strings.TrimPrefix(line, "// gohp:begin "))
			if err != nil {
				return SourceMap{}, fmt.Errorf("invalid generated source marker: %w", err)
			}
			stack = append(stack, n)
			continue
		}
		if line == "// gohp:end" {
			if len(stack) == 0 {
				return SourceMap{}, fmt.Errorf("unbalanced generated source markers")
			}
			stack = stack[:len(stack)-1]
			continue
		}
		if len(stack) == 0 || stack[len(stack)-1] <= 0 || line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		phpLine, goLine := stack[len(stack)-1], i+1
		if n := len(m.Mappings); n > 0 && m.Mappings[n-1].SourceLine == phpLine && m.Mappings[n-1].GeneratedEndLine == goLine-1 {
			m.Mappings[n-1].GeneratedEndLine = goLine
		} else {
			m.Mappings = append(m.Mappings, Mapping{goLine, goLine, phpLine})
		}
	}
	if len(stack) != 0 {
		return SourceMap{}, fmt.Errorf("unclosed generated source markers")
	}
	return m, nil
}
