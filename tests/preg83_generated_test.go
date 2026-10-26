//go:build gohp_pcre2 && cgo

package tests

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/katsuakikureyama/GoHP/internal/codegen"
	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

const preg83GeneratedPHPImage = "php:8.3.36-cli@sha256:98d555ff97b3865140ad31032508c3ca089fa97a8944b14ca9dc03ea05d788cd"

// TestPreg83GeneratedNativeComparison is deliberately opt-in: it proves the
// narrow B3 source -> IR -> generated Go path only when the caller supplies
// the pinned PHP image and an enabled PCRE2 Go build. It is not a claim that
// the default CGO-disabled compiler supports preg_match.
func TestPreg83GeneratedNativeComparison(t *testing.T) {
	image := os.Getenv("GOHP_PREG83_GENERATED_PHP_IMAGE")
	if image == "" {
		t.Skip("pinned native generated-source oracle not supplied")
	}
	if image != preg83GeneratedPHPImage {
		t.Fatalf("requires exact pinned PHP8.3.36 oracle, got %q", image)
	}

	fixture, err := filepath.Abs(filepath.Join("oracle", "preg83_generated.php"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	// A physical NUL is intentional for the bounded raw-byte test. Do this to
	// the common source before either implementation parses it.
	if bytes.Count(source, []byte("@NUL@")) != 1 {
		t.Fatal("expected exactly one binary source marker")
	}
	source = bytes.ReplaceAll(source, []byte("@NUL@"), []byte{0})
	t.Logf("native fixture sha256=%x; image=%s", sha256.Sum256(source), image)

	program, err := frontend.ParseWithOptions(string(source), "8.3", frontend.Options{ExperimentalPCRE2: true})
	if err != nil {
		t.Fatal(err)
	}
	generated, err := codegen.Generate(program)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("generated source sha256=%x Go observer sha256=%x PHP observer sha256=%x", sha256.Sum256(generated), sha256.Sum256([]byte(preg83GeneratedObserver)), sha256.Sum256([]byte(preg83GeneratedPHPObserver)))
	dir := t.TempDir()
	goFile := filepath.Join(dir, "main.go")
	observerFile := filepath.Join(dir, "observer.go")
	binary := filepath.Join(dir, "app")
	if err := os.WriteFile(goFile, generated, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(observerFile, []byte(preg83GeneratedObserver), 0600); err != nil {
		t.Fatal(err)
	}

	if out, err := exec.Command("go", "build", "-tags", "gohp_pcre2", "-o", binary, goFile, observerFile).CombinedOutput(); err != nil {
		t.Fatalf("generated build: %v\n%s\n%s", err, out, generated)
	}
	goContext, goCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer goCancel()
	goCommand := exec.CommandContext(goContext, binary)
	goCommand.Env = []string{"PATH="}
	var goStderr bytes.Buffer
	goCommand.Stderr = &goStderr
	goOutput, err := goCommand.Output()
	if err != nil || goStderr.Len() != 0 {
		t.Fatalf("generated execution: %v stderr=%s\n%s", err, goStderr.String(), goOutput)
	}

	phpFile := filepath.Join(dir, "oracle.php")
	if err := os.WriteFile(phpFile, append(append([]byte{}, source...), []byte(preg83GeneratedPHPObserver)...), 0644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", "run", "--rm", "--pull=never", "--network=none", "--read-only", "--cap-drop=ALL", "--security-opt=no-new-privileges", "--memory=128m", "--cpus=1", "--pids-limit=64", "--user=65534:65534", "--mount", "type=bind,src="+phpFile+",dst=/input.php,readonly", image, "php", "-n", "-d", "memory_limit=128M", "-d", "pcre.jit=0", "-d", "pcre.backtrack_limit=100000", "-d", "pcre.recursion_limit=10000", "/input.php")
	var phpStderr bytes.Buffer
	cmd.Stderr = &phpStderr
	phpOutput, err := cmd.Output()
	if err != nil || phpStderr.Len() != 0 {
		t.Fatalf("native oracle: error=%v stderr=%s", err, phpStderr.String())
	}

	var got, want any
	if err := json.Unmarshal(goOutput, &got); err != nil {
		t.Fatalf("generated JSON: %v\n%s", err, goOutput)
	}
	if err := json.Unmarshal(phpOutput, &want); err != nil {
		t.Fatalf("native JSON: %v\n%s", err, phpOutput)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("generated/native mismatch\ngenerated=%s\nnative=%s", goOutput, phpOutput)
	}
	t.Logf("generated sha256=%x native sha256=%x", sha256.Sum256(goOutput), sha256.Sum256(phpOutput))
	t.Logf("native JSON=%s", bytes.TrimSpace(phpOutput))
}

const preg83GeneratedObserver = `package main

import (
    "encoding/base64"
    "encoding/json"
    "fmt"
    rt "github.com/katsuakikureyama/GoHP/runtime"
)

type observation struct { Label string ` + "`json:\"label\"`" + `; Value any ` + "`json:\"value\"`" + ` }
type report struct { Schema string ` + "`json:\"schema\"`" + `; Observations []observation ` + "`json:\"observations\"`" + ` }

func typed(value any) any {
    switch value := value.(type) {
    case nil:
        return map[string]any{"type":"null", "value":nil}
    case bool:
        return map[string]any{"type":"bool", "value":value}
    case int64:
        return map[string]any{"type":"int", "value":value}
    case string:
        return map[string]any{"type":"string", "base64":base64.StdEncoding.EncodeToString([]byte(value))}
    case rt.Array:
        entries:=make([]any,0,len(value.Entries))
        for _, entry:=range value.Entries { entries=append(entries,map[string]any{"key":typed(entry.Key), "value":typed(entry.Value)}) }
        return map[string]any{"type":"array", "entries":entries}
    default:
        panic(fmt.Sprintf("unsupported generated observation type %T",value))
    }
}
func add(rows *[]observation,label string,value any) { *rows=append(*rows,observation{Label:label,Value:typed(value)}) }
func check(got,want any) { if got!=want { panic(fmt.Sprintf("got %v want %v",got,want)) } }
func init() {
    rows:=[]observation{}
    add(&rows,"basic",F_preg83_basic())
    add(&rows,"nomatch",F_preg83_nomatch())
    add(&rows,"named_unmatched",F_preg83_named_unmatched())
    add(&rows,"negative_offset",F_preg83_negative_offset())
    add(&rows,"middle_capture",F_preg83_middle_capture("zac",int64(-2)))
    add(&rows,"capture_copy",F_preg83_capture_copy())
    add(&rows,"binary_subject",F_preg83_binary_subject())
    add(&rows,"utf8_byte_offset",F_preg83_utf8_byte_offset())
    add(&rows,"lookahead_possessive",F_preg83_lookahead_possessive())
    add(&rows,"failure",F_preg83_failure())
    add(&rows,"shared_execution",F_preg83_shared_execution())

    // PHP's process-global oracle cannot model two host Executions. Check the
    // documented generated ABI explicitly: contexts are isolated, calls made
    // through an existing context share its state, and standalone wrappers are fresh.
    a,b:=&rt.Execution{},&rt.Execution{}
    FContext_preg83_failure(a)
    check(FContext_preg83_read_error(a),int64(2))
    check(FContext_preg83_read_error(b),int64(0))
    FContext_preg83_basic(b)
    check(FContext_preg83_read_error(b),int64(0))
    FContext_preg83_nomatch(a)
    check(FContext_preg83_read_error(a),int64(0))
    FContext_preg83_failure(a)
    FContext_preg83_basic(a)
    check(FContext_preg83_read_error(a),int64(0))
    check(F_preg83_read_error(),int64(0))

    out,err:=json.Marshal(report{Schema:"gohp.preg83-generated-reference/v1",Observations:rows})
    if err!=nil { panic(err) }
    fmt.Print(string(out))
}
`

const preg83GeneratedPHPObserver = `
function gohp_preg83_generated_typed($value) {
    if (is_string($value)) return ['type' => 'string', 'base64' => base64_encode($value)];
    if (is_int($value)) return ['type' => 'int', 'value' => $value];
    if (is_bool($value)) return ['type' => 'bool', 'value' => $value];
    if ($value === null) return ['type' => 'null', 'value' => null];
    if (!is_array($value)) throw new RuntimeException('unsupported observer type');
    $entries = [];
    foreach ($value as $key => $item) {
        $entries[] = ['key' => gohp_preg83_generated_typed($key), 'value' => gohp_preg83_generated_typed($item)];
    }
    return ['type' => 'array', 'entries' => $entries];
}
if (PHP_VERSION !== '8.3.36' || PHP_INT_SIZE !== 8 || setlocale(LC_CTYPE, 0) !== 'C' ||
    PCRE_VERSION !== '10.42 2022-12-12' ||
    ini_get('pcre.jit') !== '0' || ini_get('pcre.backtrack_limit') !== '100000' || ini_get('pcre.recursion_limit') !== '10000') {
    throw new RuntimeException('requires pinned PHP8.3.36 64-bit C locale CLI');
}
$observations = [];
foreach (['basic', 'nomatch', 'named_unmatched', 'negative_offset', 'middle_capture', 'capture_copy', 'binary_subject', 'utf8_byte_offset', 'lookahead_possessive', 'failure', 'shared_execution'] as $name) {
    $function = 'preg83_' . $name;
    $value = $name === 'middle_capture' ? $function('zac', -2) : $function();
    if ($name === 'failure' && ($value['result'] !== false || $value['error'] !== 2)) throw new RuntimeException('expected backtrack failure');
    $observations[] = ['label' => $name, 'value' => gohp_preg83_generated_typed($value)];
}
echo json_encode(['schema' => 'gohp.preg83-generated-reference/v1', 'observations' => $observations], JSON_THROW_ON_ERROR), "\n";
`
