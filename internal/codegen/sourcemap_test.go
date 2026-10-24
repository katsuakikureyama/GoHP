package codegen

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestStatementSourceMap(t *testing.T) {
	const php = `<?php
function identity($arg) {
  return $arg;
}
class Box {
  public function value() {
    echo "method";
    return 2;
  }
}
$items = [[1]];
foreach ($items as $key => $row) {
  foreach ($row as $value) {
    echo identity($value);
  }
  echo $key;
}
$box = new Box();
echo $box->value();
echo '// gohp:end';
echo 'line
// gohp:begin 999';
`
	for _, target := range []string{"8.3"} {
		t.Run(target, func(t *testing.T) {
			p, err := frontend.Parse(php, target)
			if err != nil {
				t.Fatal(err)
			}
			generated, m, err := GenerateWithMap(p)
			if err != nil {
				t.Fatal(err)
			}
			formatted, err := format.Source(generated)
			if err != nil || !bytes.Equal(formatted, generated) {
				t.Fatalf("not canonical formatted Go: %v", err)
			}
			lines := strings.Split(string(generated), "\n")
			for fragment, want := range map[string]int{
				`return rt.Clone(v["arg"])`:          3,
				`rt.Echo("method")`:                  7,
				`return rt.Clone(int64(2))`:          8,
				`v["items"] =`:                       11,
				`v["key"] =`:                         12,
				`v["row"] =`:                         12,
				`v["value"] =`:                       13,
				`rt.Echo(F_identity(`:                14,
				`rt.Echo(v["key"])`:                  16,
				`v["box"] =`:                         18,
				`rt.Echo((v["box"])`:                 19,
				`rt.Echo("// gohp:end")`:             20,
				`rt.Echo("line\n// gohp:begin 999")`: 21,
			} {
				found := false
				for i, line := range lines {
					if strings.Contains(line, fragment) {
						found = true
						if got, ok := m.PHPLine(i + 1); !ok || got != want {
							t.Errorf("Go line %d (%s): got PHP %d, %v; want %d", i+1, fragment, got, ok, want)
						}
					}
				}
				if !found {
					t.Errorf("missing generated fragment %s", fragment)
				}
			}
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "generated.go", generated, 0)
			if err != nil {
				t.Fatal(err)
			}
			loop := 0
			ast.Inspect(file, func(n ast.Node) bool {
				switch n := n.(type) {
				case *ast.RangeStmt:
					want := []int{12, 13}[loop]
					loop++
					for _, pos := range []token.Pos{n.Pos(), n.Body.Rbrace} {
						line := fset.Position(pos).Line
						if got, ok := m.PHPLine(line); !ok || got != want {
							t.Errorf("loop boundary %d mapped to %d, %v; want %d", line, got, ok, want)
						}
					}
				case *ast.FuncDecl:
					if _, ok := m.PHPLine(fset.Position(n.Pos()).Line); ok {
						t.Errorf("synthetic function declaration %s mapped", n.Name.Name)
					}
				}
				return true
			})
			if loop != 2 {
				t.Fatalf("expected two loops, got %d", loop)
			}
			last := 0
			for _, entry := range m.Mappings {
				if entry.GeneratedStartLine <= last || entry.GeneratedEndLine < entry.GeneratedStartLine || entry.GeneratedEndLine > len(lines) || entry.SourceLine < 1 || entry.SourceLine > 21 {
					t.Fatalf("invalid map entry: %+v", entry)
				}
				last = entry.GeneratedEndLine
			}
			for i, line := range lines {
				if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "//") || strings.Contains(line, "rt.Args(") {
					if _, ok := m.PHPLine(i + 1); ok {
						t.Errorf("scaffolding/comment/blank line %d mapped", i+1)
					}
				}
			}
			for _, line := range []int{-1, 0, len(lines) + 1} {
				if _, ok := m.PHPLine(line); ok {
					t.Errorf("out-of-bounds Go line %d mapped", line)
				}
			}
		})
	}
}

func TestSourceMapUnknownPositionsAndFailure(t *testing.T) {
	p, err := frontend.Parse("<?php echo 1; echo 2;", "8.3")
	if err != nil {
		t.Fatal(err)
	}
	p.Body[1].Line = 0
	generated, m, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	for i, line := range strings.Split(string(generated), "\n") {
		if strings.Contains(line, "rt.Echo(int64(2))") {
			if _, ok := m.PHPLine(i + 1); ok {
				t.Fatal("unknown position inherited previous statement's line")
			}
		}
	}
	_, empty, err := GenerateWithMap(ir.Program{Version: ir.Version, PHPVersion: "8.3"})
	if err != nil || empty.Mappings == nil || len(empty.Mappings) != 0 {
		t.Fatalf("empty map should serialize as [], got %+v, %v", empty, err)
	}
	p.Body[0].Expr = ir.Expr{Kind: "var", Text: "undefined", Line: 1}
	generated, m, err = GenerateWithMap(p)
	if err == nil || generated != nil || m.Schema != "" || m.Mappings != nil {
		t.Fatalf("failed generation leaked partial artifacts: %q, %+v, %v", generated, m, err)
	}
}
